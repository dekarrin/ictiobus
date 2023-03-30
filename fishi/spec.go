package fishi

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/translation"
	"github.com/dekarrin/ictiobus/types"
	"github.com/dekarrin/rosed"
)

// Spec is a series of statements that together give the specification for a
// compiler frontend of a language. It is created by processing an AST and
// checking it for errors with NewSpec.
type Spec struct {
	// Tokens is all of the tokens that are used in the language.
	Tokens []types.TokenClass

	// Patterns is a map of state names to the lexer patterns that are used
	// while in that state.
	Patterns map[string][]Pattern

	// Grammar is the syntactical specification of the language.
	Grammar grammar.Grammar

	// TranslationScheme outlines the Syntax-Directed Translation Scheme for the
	// language by giving the instructions for each attribute.
	TranslationScheme []SDD
}

// Pattern is a lexer pattern that is used to match a token, along with the
// action that the lexer should take when it matches.
type Pattern struct {
	// Regex is the regular expression that is used to match the token.
	Regex *regexp.Regexp

	// Action is the action that the lexer should take when it matches the
	// token.
	Action lex.Action

	// Priority is the priority of the pattern. 0 is the lowest priority, and
	// higher numbers will take precedence over lower numbers.
	Priority int
}

// SDD is a Syntax-Directed Definition, that is, a single instruction for a
// single attribute.
type SDD struct {
	// Attribute is the reference to the attribute that this SDD will set. If
	// this is RelHead, then the attribute will be set on the head of the
	// relation and it is a synthesized attribute; otherwise, this is an
	// inherited attribute.
	Attribute translation.AttrRef

	// Rule is the grammar haed and production that this SDD is for. This will
	// have exactly one production in it, as opposed to Rule structs stored in
	// the grammar, which contain *all* productions for a given head symbol.
	Rule grammar.Rule

	// Hook is the name of the hook that is called to get the value to set on
	// the attribute.
	Hook string

	// Args is the list of arguments to the hook.
	Args []translation.AttrRef
}

// NewSpec reads an AST and creates a LanguageSpec from it. If the AST has
// any errors, then an error is returned. If the AST has non-error warnings,
// they will be returned in the warnings slice. Warnings will be present and
// valid even if err is non-nil; spec will only be valid if err is nil.
func NewSpec(ast AST) (spec Spec, warnings []Warning, err error) {
	ls := Spec{
		Patterns: make(map[string][]Pattern),
	}

	// all tokens blocks must be processed before any grammar blocks, and all
	// grammar blocks must be processed before any actions blocks.

	// first, gather each type of AST block into a single listing
	tokensBlocks := []ASTTokensContent{}
	grammarBlocks := []ASTGrammarContent{}
	actionsBlocks := []ASTActionsContent{}

	for _, bl := range ast.Nodes {
		switch bl := bl.(type) {
		case ASTTokensBlock:
			tokBl := bl.Tokens()
			tokensBlocks = append(tokensBlocks, tokBl.Content...)
		case ASTGrammarBlock:
			gramBl := bl.Grammar()
			grammarBlocks = append(grammarBlocks, gramBl.Content...)
		case ASTActionsBlock:
			actBl := bl.Actions()
			actionsBlocks = append(actionsBlocks, actBl.Content...)
		}
	}

	warnings = []Warning{}

	// var to store warnings returned from functions before adding to warnings
	var subWarns []Warning

	// go over tokensBlocks to get token classes

	// pre-scan to prep for tokenBlocks
	tcDefsTable, states := scanTokenClasses(tokensBlocks)
	subWarns = checkForDuplicateHumanDefs(tcDefsTable)
	if len(subWarns) > 0 {
		warnings = append(warnings, subWarns...)
	}

	classes, warns := buildTokenClasses(tcDefsTable, states)
	if len(warns) > 0 {
		warnings = append(warnings, warns...)
	}

	// put classes into spec, ordered alphabetically
	tokClassNamesAlpha := textfmt.OrderedKeys(classes)
	for _, tok := range tokClassNamesAlpha {
		ls.Tokens = append(ls.Tokens, classes[tok])
	}

	// go over tokensBlocks to get lexer patterns
	ls.Patterns, subWarns, err = analzyeASTTokensContentSlice(tokensBlocks, states, classes)
	if len(subWarns) > 0 {
		warnings = append(warnings, subWarns...)
	}
	if err != nil {
		return ls, warnings, err
	}

	// go over grammarBlocks to get grammar
	ls.Grammar, subWarns, err = analyzeASTGrammarContentSlice(grammarBlocks, classes)
	if len(subWarns) > 0 {
		warnings = append(warnings, subWarns...)
	}
	if err != nil {
		return ls, warnings, err
	}

	// go over actionsBlocks to get translation scheme
	// TODO: when done, put this in a function
	ls.TranslationScheme, subWarns, err = analyzeASTActionsContentSlice(actionsBlocks, ls.Grammar)
	if len(subWarns) > 0 {
		warnings = append(warnings, subWarns...)
	}
	if err != nil {
		return ls, warnings, err
	}

	return ls, warnings, nil
}

func analyzeASTActionsContentSlice(
	actionsBlocks []ASTActionsContent,
	g grammar.Grammar,
) ([]SDD, []Warning, error) {
	var warnings []Warning
	var scheme []SDD

	for _, actBl := range actionsBlocks {
		if actBl.State != "" {
			return nil, warnings, fmt.Errorf("actions blocks in non-default state not supported yet")
		}

		for _, symAct := range actBl.Actions {
			// remove brace and make upper-case
			ruleHead := strings.ToUpper(symAct.Symbol[1 : len(symAct.Symbol)-1])

			// we will need the grammar rule to check the production action(s)

			gRule := g.Rule(ruleHead)
			if gRule.NonTerminal == "" {
				synErr := types.NewSyntaxErrorFromToken(fmt.Sprintf("'%s' is not a non-terminal symbol in the grammar", ruleHead), symAct.symTok)
				return nil, warnings, synErr
			}

			forProdIdx := -1
			for _, prodAct := range symAct.Actions {
				sddRule := grammar.Rule{NonTerminal: ruleHead}

				if prodAct.ProdNext {
					// next production specified

					forProdIdx++
					if forProdIdx >= len(gRule.Productions) {
						prodsStr := textfmt.Pluralize(len(gRule.Productions), "production", "-s")
						errFmt := "'->' by itself specifies production #%d, but grammar for %s only defines %s"
						errMsg := fmt.Sprintf(errFmt, forProdIdx+1, ruleHead, prodsStr)
						synErr := types.NewSyntaxErrorFromToken(errMsg, prodAct.valTok)
						return nil, warnings, synErr
					}
				} else if len(prodAct.ProdLiteral) > 0 {
					// specific production specified

					// find the production within the grammar rule

					// we need to go through each symbol and convert it to a
					// full grammar.Production. Check each symbol; if it is
					// wrapped in braces, it's a non-terminal, so we remove the
					// braces and make it upper-case. If it's not wrapped in
					// braces, it's a terminal, so we make it lower-case.
					convertedProd := make(grammar.Production, len(prodAct.ProdLiteral))

					for i, sym := range prodAct.ProdLiteral {
						if sym[0] == '{' && sym[len(sym)-1] == '}' {
							convertedProd[i] = strings.ToUpper(sym[1 : len(sym)-1])
						} else {
							convertedProd[i] = strings.ToLower(sym)
						}
					}

					// now find the produciton that matches it
					found := false
					for i, prod := range gRule.Productions {
						if prod.Equal(convertedProd) {
							found = true
							forProdIdx = i
							break
						}
					}

					// if there's no match, we have an error
					if !found {
						errFmt := "no grammar rule specifies %s -> '%s'"
						errMsg := fmt.Sprintf(errFmt, convertedProd.String(), ruleHead)
						synErr := types.NewSyntaxErrorFromToken(errMsg, prodAct.valTok)
						return nil, warnings, synErr
					}
				} else {
					// production specified by index, by far the easiest case
					forProdIdx = prodAct.ProdIndex
				}

				// get the production specified from the grammar rule
				sddRule.Productions = []grammar.Production{gRule.Productions[forProdIdx]}

				// go through and create an SDD for each semantic action listed
				for _, semAct := range prodAct.Actions {
					sdd := SDD{Rule: sddRule.Copy()}

					var err error

					// convert LHS ASTAttrRef to valid translation.AttrRef
					sdd.Attribute, err = attrRefFromASTAttrRef(semAct.LHS, g, sddRule)
					if err != nil {
						synErr := types.NewSyntaxErrorFromToken("invalid attrRef: "+err.Error(), semAct.LHS.tok)
						return nil, warnings, synErr
					}

					// do the same for each arg to the hook
					if len(semAct.With) > 0 {
						sdd.Args = make([]translation.AttrRef, len(semAct.With))
						for i, arg := range semAct.With {
							sdd.Args[i], err = attrRefFromASTAttrRef(arg, g, sddRule)
							if err != nil {
								synErr := types.NewSyntaxErrorFromToken("invalid attrRef: "+err.Error(), arg.tok)
								return nil, warnings, synErr
							}
						}
					}

					// finally, get the hook name
					sdd.Hook = semAct.Hook

					scheme = append(scheme, sdd)
				}
			}
		}
	}

	return scheme, warnings, nil
}

func analyzeASTGrammarContentSlice(
	grammarBlocks []ASTGrammarContent,
	classes map[string]types.TokenClass,
) (grammar.Grammar, []Warning, error) {
	var warnings []Warning

	g := grammar.Grammar{}
	hitFirst := false

	// track terminals in the grammar to make sure they're all used
	seenTerminals := make(map[string]bool)
	for _, gBl := range grammarBlocks {
		if gBl.State != "" {
			return g, warnings, fmt.Errorf("grammar blocks in non-default state not supported yet")
		}

		for _, rule := range gBl.Rules {
			head := rule.Rule.NonTerminal
			// remove braces and make upper-case
			head = strings.ToUpper(head[1 : len(head)-1])

			for _, prod := range rule.Rule.Productions {
				newProd := grammar.Production{}
				for _, sym := range prod {
					// if it starts with a brace, it's a non-terminal, drop braces and make upper-case
					if sym[0] == '{' && sym[len(sym)-1] == '}' {
						sym = strings.ToUpper(sym[1 : len(sym)-1])
					} else {
						// else, it's a terminal, make lower-case...
						sym = strings.ToLower(sym)

						// ...and make sure it's in the lexer's terminals
						if _, ok := classes[sym]; !ok {
							synErr := types.NewSyntaxErrorFromToken(fmt.Sprintf("terminal '%s' is not a defined token class in any tokens block", sym), rule.tok)
							return g, warnings, synErr
						}

						// mark it as seen
						seenTerminals[sym] = true
					}

					newProd = append(newProd, sym)
				}
				g.AddRule(head, newProd)
			}

			if !hitFirst {
				g.Start = head
				hitFirst = true
			}
		}
	}

	// make sure all terminals are used (deterministically to aid debugging)
	orderedTokenClassNames := textfmt.OrderedKeys(classes)
	for _, tcName := range orderedTokenClassNames {
		tc := classes[tcName]
		if _, ok := seenTerminals[tc.ID()]; !ok {
			w := Warning{
				Type:    WarnUnusedTerminal,
				Message: fmt.Sprintf("token class '%s' is defined in a tokens block but not used as terminal in any grammar rule", tc.ID()),
			}
			warnings = append(warnings, w)
		}
	}

	// validate the grammar
	if err := g.Validate(); err != nil {
		return g, warnings, err
	}

	return g, warnings, nil
}

func analzyeASTTokensContentSlice(
	tokensBlocks []ASTTokensContent,
	existingStates box.StringSet,
	classes map[string]types.TokenClass,
) (map[string][]Pattern, []Warning, error) {
	var warnings []Warning

	pats := make(map[string][]Pattern)

	for _, tokBl := range tokensBlocks {
		var p Pattern
		for _, entry := range tokBl.Entries {
			// either an entry specifies discard, OR it specifies up to one each
			// of stateshift, token, human. priority may be in either.
			var err error

			// get the pattern
			p.Regex, err = regexp.Compile(entry.Pattern)
			if err != nil {
				synErr := types.NewSyntaxErrorFromToken(fmt.Sprintf("invalid regular expression: %s", err.Error()), entry.tok)
				return nil, warnings, synErr
			}

			// make sure we only have one maximum of each option
			if len(entry.discardTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate discard directive for entry", entry.discardTok[1])
				return nil, warnings, synErr
			}
			if len(entry.humanTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate human directive for entry", entry.humanTok[1])
				return nil, warnings, synErr
			}
			if len(entry.priorityTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate priority directive for entry", entry.priorityTok[1])
				return nil, warnings, synErr
			}
			if len(entry.shiftTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate state shift directive for entry", entry.shiftTok[1])
				return nil, warnings, synErr
			}

			// make sure mutually exclusive options are not used
			if entry.Discard {
				// then there'd betta not be a human directive, a token
				// directive, or a shift directive.
				// TODO: grammar should rly enforce this

				// error report on the *2nd* token to break things

				firstTok := entry.discardTok[0]
				firstIsDiscard := true
				var secondTok types.Token

				if len(entry.humanTok) > 0 {
					humanTok := entry.humanTok[0]
					putEntryTokenInCorrectPosForDiscardCheck(&firstTok, &secondTok, &firstIsDiscard, humanTok)
				}
				if len(entry.tokenTok) > 0 {
					tokenTok := entry.tokenTok[0]
					putEntryTokenInCorrectPosForDiscardCheck(&firstTok, &secondTok, &firstIsDiscard, tokenTok)
				}
				if len(entry.shiftTok) > 0 {
					shiftTok := entry.shiftTok[0]
					putEntryTokenInCorrectPosForDiscardCheck(&firstTok, &secondTok, &firstIsDiscard, shiftTok)
				}

				if secondTok != nil {
					var fullErrMsg string
					if firstIsDiscard {
						errMsg := "human/token/stateshift directive cannot be added to discarded entry:"
						synErr1 := types.NewSyntaxErrorFromToken("initial discard defined here", firstTok)
						synErr2 := types.NewSyntaxErrorFromToken("directive not allowed", secondTok)

						fullErrMsg = errMsg + "\n" + synErr1.FullMessage() + "\n" + synErr2.FullMessage()
					} else {
						errMsg := "can't discard an entry that will be used for stateshift or token lexing:"
						synErr1 := types.NewSyntaxErrorFromToken("initial directive defined here", firstTok)
						synErr2 := types.NewSyntaxErrorFromToken("discard directive not allowed", secondTok)

						fullErrMsg = errMsg + "\n" + synErr1.FullMessage() + "\n" + synErr2.FullMessage()
					}

					return nil, warnings, fmt.Errorf(fullErrMsg)
				}

				// if here, the only options that could be present are discard and priority. take the discard.

				p.Action = lex.Discard()
			} else {
				// from here, it could be a stateshift, a token, or both. human
				// is allowed if token is present.

				if entry.Human != "" {
					// then there'd 8etta be a token directive

					if entry.Token == "" {
						synErr := types.NewSyntaxErrorFromToken("human directive given without token directive", entry.humanTok[0])
						return nil, warnings, synErr
					}
				}

				if entry.Token == "" && entry.Shift == "" {
					synErr := types.NewSyntaxErrorFromToken("entry must have a discard, token, or stateshift directive", entry.tok)
					return nil, warnings, synErr
				}

				// don't try to shift to non-existent state
				if entry.Shift != "" {
					if !existingStates.Has(entry.Shift) {
						synErr := types.NewSyntaxErrorFromToken("bad stateshift; shifted-to-state does not exist", entry.shiftTok[0])
						return nil, warnings, synErr
					}

					if entry.Shift == tokBl.State {
						synErr := types.NewSyntaxErrorFromToken("bad stateshift; already in that state", entry.shiftTok[0])
						return nil, warnings, synErr
					}
				}

				// all checks complete, now build the action

				if entry.Token != "" {
					class := classes[entry.Token]

					if entry.Shift != "" {
						// stateshift and token
						p.Action = lex.LexAndSwapState(class.ID(), entry.Shift)
					} else {
						// just token
						p.Action = lex.LexAs(class.ID())
					}
				} else {
					// just stateshift
					p.Action = lex.SwapState(entry.Shift)
				}
			}

			// finally, check for priority
			if len(entry.priorityTok) > 0 {
				if p.Priority == 0 {
					warn := types.NewSyntaxErrorFromToken("setting priority to 0 has no effect", entry.priorityTok[0])
					warnings = append(warnings, Warning{
						Type:    WarnPriorityZero,
						Message: warn.FullMessage(),
					})
				} else if p.Priority < 0 {
					synErr := types.NewSyntaxErrorFromToken("priority cannot be negative", entry.priorityTok[0])
					return nil, warnings, synErr
				}

				p.Priority = entry.Priority
			}
		}

		// add the pattern to the lexer
		statePats, ok := pats[tokBl.State]
		if !ok {
			statePats = make([]Pattern, 0)
		}
		statePats = append(statePats, p)
		pats[tokBl.State] = statePats
	}

	return pats, warnings, nil
}

// r is rule to check against, only first production is checked.
func attrRefFromASTAttrRef(astRef ASTAttrRef, g grammar.Grammar, r grammar.Rule) (translation.AttrRef, error) {
	if astRef.Head {
		return translation.AttrRef{
			Relation: translation.NodeRelation{
				Type: translation.RelHead,
			},
			Name: astRef.Attribute,
		}, nil
	} else if astRef.SymInProd {
		// make sure the rule has the right number of symbols
		if astRef.Occurance >= len(r.Productions[0]) {
			return translation.AttrRef{}, fmt.Errorf("symbol index out of range; production only has %d symbols", len(r.Productions[0]))
		}
		return translation.AttrRef{
			Relation: translation.NodeRelation{
				Type:  translation.RelSymbol,
				Index: astRef.Occurance,
			},
			Name: astRef.Attribute,
		}, nil
	} else if astRef.NontermInProd {
		// make sure that the rule has that number of non-terminals
		nontermCount := 0
		for _, sym := range r.Productions[0] {
			if g.IsNonTerminal(sym) {
				nontermCount++
			}
		}
		if astRef.Occurance >= nontermCount {
			return translation.AttrRef{}, fmt.Errorf("non-terminal index out of range; production only has %d non-terminals", nontermCount)
		}
		return translation.AttrRef{
			Relation: translation.NodeRelation{
				Type:  translation.RelNonTerminal,
				Index: astRef.Occurance,
			},
			Name: astRef.Attribute,
		}, nil
	} else if astRef.TermInProd {
		// make sure that the rule has that number of terminals
		termCount := 0
		for _, sym := range r.Productions[0] {
			if g.IsTerminal(sym) {
				termCount++
			}
		}
		if astRef.Occurance >= termCount {
			return translation.AttrRef{}, fmt.Errorf("terminal index out of range; production only has %d terminals", termCount)
		}
		return translation.AttrRef{
			Relation: translation.NodeRelation{
				Type:  translation.RelTerminal,
				Index: astRef.Occurance,
			},
			Name: astRef.Attribute,
		}, nil
	} else {
		// it's an instance of a particular symbol. find out the symbol index.
		symIndexes := []int{}
		for i, sym := range r.Productions[0] {
			if sym == astRef.Symbol {
				symIndexes = append(symIndexes, i)
			}
		}
		if len(symIndexes) == 0 {
			return translation.AttrRef{}, fmt.Errorf("no symbol %s in production", astRef.Symbol)
		}
		if astRef.Occurance >= len(symIndexes) {
			return translation.AttrRef{}, fmt.Errorf("symbol index out of range; production only has %d instances of %s", len(symIndexes), astRef.Symbol)
		}
		return translation.AttrRef{
			Relation: translation.NodeRelation{
				Type:  translation.RelSymbol,
				Index: symIndexes[astRef.Occurance],
			},
			Name: astRef.Attribute,
		}, nil
	}
}

func putEntryTokenInCorrectPosForDiscardCheck(first, second *types.Token, discardIsFirst *bool, tok types.Token) {
	if *discardIsFirst {
		if tokenIsBefore(tok, *first) {
			*second = *first
			*first = tok
			*discardIsFirst = false
		} else {
			if *second == nil || tokenIsBefore(tok, *second) {
				*second = tok
			}
		}
	} else {
		// discard is in second place. leave it there.
		if tokenIsBefore(tok, *first) {
			*first = tok
		}
	}
}

func tokenIsBefore(t1, t2 types.Token) bool {
	return t1.Line() < t2.Line() || (t1.Line() == t2.Line() && t1.LinePos() < t2.LinePos())
}

func buildTokenClasses(tcDefsTable map[string][]box.Pair[string, types.Token], states box.StringSet) (map[string]types.TokenClass, []Warning) {
	var warnings []Warning

	classes := make(map[string]types.TokenClass)
	// build token classes
	for tok, humanDefs := range tcDefsTable {
		// if there is no human definition, then use the token name
		human := tok
		if len(humanDefs) > 0 {
			human = humanDefs[len(humanDefs)-1].First
		} else {
			newWarn := Warning{
				Type:    WarnMissingHumanDef,
				Message: fmt.Sprintf("no human-readable name given for token %q; defaulting to %q", tok, tok),
			}
			warnings = append(warnings, newWarn)
		}

		tokenClass := lex.NewTokenClass(tok, human)
		classes[tok] = tokenClass
	}

	return classes, warnings
}

func checkForDuplicateHumanDefs(tcSymTable map[string][]box.Pair[string, types.Token]) []Warning {
	var warnings []Warning
	// warn for duplicate human definitions (but not missing; we will handle
	// that during reading of tokenBlocks)
	for tok, humanDefs := range tcSymTable {
		if len(humanDefs) > 1 {
			msgStart := fmt.Sprintf("multiple distinct human-readable names given for token %q:", tok)
			var msg string
			for _, hd := range humanDefs {
				synErr := types.NewSyntaxErrorFromToken("human name last defined here", hd.Second)
				msg += fmt.Sprintf("%s\n", synErr.FullMessage())
			}
			msg = rosed.Edit(msg).IndentOpts(1, rosed.Options{IndentStr: "  "}).String()
			fullWarn := Warning{
				Type:    WarnDuplicateHumanDefs,
				Message: msgStart + "\n" + msg,
			}
			warnings = append(warnings, fullWarn)
		}
	}

	return warnings
}

// scanTokenClasses scans the blocks for distinct tokens and their human
// definitions, as well as the states that are defined.
//
// blocks is scanned for all token classes and lexer state names.
// do not error check (but do track for multiple human definition text)
// until the scan is complete even if we could; that way, all errors are
// reported at once.
func scanTokenClasses(blocks []ASTTokensContent) (map[string][]box.Pair[string, types.Token], box.StringSet) {

	// tcSymTable is tok-name -> pairs of human-name and token where that human
	// name is first defined. Uses slice of pairs instead of map to preserve
	// order.
	tcSymTable := make(map[string][]box.Pair[string, types.Token])

	states := box.NewStringSet()

	for _, bl := range blocks {
		// don't check for at least least one entry; an empty tokens or state
		// block is not an error

		// track states we have entries for
		if bl.State != "" {
			states.Add(bl.State)
		}

		for _, entry := range bl.Entries {
			if entry.Token != "" {
				humanDefs, ok := tcSymTable[entry.Token]
				if !ok {
					humanDefs = []box.Pair[string, types.Token]{}
					tcSymTable[entry.Token] = humanDefs
				}
				if entry.Human != "" {
					keepHuman := true

					// if we already have human definition(s) for this token,
					// only add this one if it is a distinct string.
					if len(humanDefs) != 0 {
						_, alreadyExists := slices.Any(humanDefs, func(hd box.Pair[string, types.Token]) bool {
							return hd.First == entry.Human
						})
						// only need to save a new one if it doesn't already exist
						keepHuman = !alreadyExists
					}

					if keepHuman {
						humanDefs = append(humanDefs, box.PairOf(entry.Human, entry.humanTok[len(entry.humanTok)-1]))
						tcSymTable[entry.Token] = humanDefs
					}
				}
			}
		}
	}

	return tcSymTable, states
}
