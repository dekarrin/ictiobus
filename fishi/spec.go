package fishi

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/fishi/syntax"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/rosed"
)

// Spec is a series of statements that together give the specification for a
// compiler frontend of a language. It is created by processing an AST and
// checking it for errors with NewSpec.
type Spec struct {
	// Tokens is all of the tokens that are used in the language.
	Tokens []lex.TokenClass

	// Patterns is a map of state names to the lexer patterns that are used
	// while in that state.
	Patterns map[string][]Pattern

	// Grammar is the syntactical specification of the language.
	Grammar grammar.CFG

	// TranslationScheme outlines the Syntax-Directed Translation Scheme for the
	// language by giving the instructions for each attribute.
	TranslationScheme []SDD
}

// SpecMetadata is data that is not strictly part of the spec but tells info
// about the language it was generated for and how it was generated.
type SpecMetadata struct {
	// Language is name of the language.
	Language string

	// Version is the version of the language.
	Version string

	// InvocationArgs are the arguments
	InvocationArgs string
}

// ValidateSDTS builds the Lexer and SDTS and runs validation on several
// simulated parse trees to ensure that the SDTS is valid and works.
func (spec Spec) ValidateSDTS(opts trans.ValidationOptions, hooks trans.HookMap) ([]Warning, error) {
	lx, err := spec.CreateLexer(true)
	if err != nil {
		return nil, fmt.Errorf("lexer creation error: %w", err)
	}

	valProd := lx.FakeLexemeProducer(true, "")

	sdts, err := spec.CreateSDTS()
	if err != nil {
		return nil, fmt.Errorf("SDTS creation error: %w", err)
	}

	sdts.SetHooks(hooks)

	// validate the SDTS. the first defined attribute will be the IR attribute.
	irAttrName := spec.TranslationScheme[0].Attribute.Name

	allWarnings := []Warning{}

	warns, sdtsErr := trans.Validate(sdts, spec.Grammar, irAttrName, opts, valProd)
	for _, w := range warns {
		allWarnings = append(allWarnings, Warning{Type: WarnValidation, Message: w})
	}

	if sdtsErr != nil {
		return allWarnings, fmt.Errorf("SDTS validation error: %w", sdtsErr)
	}

	return allWarnings, nil
}

// ClassMap is a map of string to token class with that ID.
func (spec Spec) ClassMap() map[string]lex.TokenClass {
	classes := map[string]lex.TokenClass{}
	for _, class := range spec.Tokens {
		classes[class.ID()] = class
	}
	return classes
}

// CreateLexer uses the Tokens and Patterns in the spec to create a new Lexer.
func (spec Spec) CreateLexer(lazy bool) (lex.Lexer, error) {
	var lx lex.Lexer

	if lazy {
		lx = ictiobus.NewLazyLexer()
	} else {
		lx = ictiobus.NewLexer()
	}

	// find the tokens we need to register classes for
	toBeRegisteredOrd := map[string][]string{}
	toBeRegisteredSet := map[string]box.StringSet{}

	for state := range spec.Patterns {
		statePats := spec.Patterns[state]
		stateToRegOrd, ok := toBeRegisteredOrd[state]
		stateToRegSet := toBeRegisteredSet[state]
		if !ok {
			stateToRegOrd = []string{}
			stateToRegSet = box.NewStringSet()
		}

		for _, pat := range statePats {
			if pat.Action.Type == lex.ActionScan || pat.Action.Type == lex.ActionScanAndState {
				// we need to register the classes for this pattern
				if !stateToRegSet.Has(pat.Action.ClassID) {
					stateToRegOrd = append(stateToRegOrd, pat.Action.ClassID)
					stateToRegSet.Add(pat.Action.ClassID)
				}
			}
		}

		toBeRegisteredOrd[state] = stateToRegOrd
		toBeRegisteredSet[state] = stateToRegSet
	}

	classes := spec.ClassMap()

	// register the classes
	for state, classIDs := range toBeRegisteredOrd {
		for _, id := range classIDs {
			lx.RegisterClass(classes[id], state)
		}
	}

	// add the patterns
	for state, pats := range spec.Patterns {
		for _, pat := range pats {
			err := lx.AddPattern(pat.Regex.String(), pat.Action, state, pat.Priority)
			if err != nil {
				// all error conditions should be handled; should never happen
				return nil, err
			}
		}
	}

	// done!
	return lx, nil
}

// CreateMostRestrictiveParser creates the most restrictive parser possible for
// the language it represents. They will be tried in this order: LL(1), SLR(1),
// LALR(1), CLR(1).
//
// AllowAmbig only applies for parser types that can auto-resolve ambiguity,
// e.g. it does not apply to an LL(k) parser.
func (spec Spec) CreateMostRestrictiveParser(allowAmbig bool) (parse.Parser, []Warning, error) {
	p, warns, err := spec.CreateParser(parse.LL1, false)
	if err != nil {
		p, warns, err = spec.CreateParser(parse.SLR1, allowAmbig)
		if err != nil {
			p, warns, err = spec.CreateParser(parse.LALR1, allowAmbig)
			if err != nil {
				p, warns, err = spec.CreateParser(parse.CLR1, allowAmbig)
				if err != nil {
					return p, warns, fmt.Errorf("no parser can be generated for grammar; for CLR(1) parser, got: %w", err)
				}
			}
		}
	}

	return p, warns, err
}

// CreateParser uses the Grammar in the spec to create a new Parser of the
// given type. Returns an error if the type is not supported.
func (spec Spec) CreateParser(t parse.Algorithm, allowAmbig bool) (parse.Parser, []Warning, error) {
	var warns []Warning
	var p parse.Parser
	var err error

	var ambigWarns []string
	switch t {
	case parse.LALR1:
		p, ambigWarns, err = ictiobus.NewLALRParser(spec.Grammar, allowAmbig)
	case parse.CLR1:
		p, ambigWarns, err = ictiobus.NewCLRParser(spec.Grammar, allowAmbig)
	case parse.SLR1:
		p, ambigWarns, err = ictiobus.NewSLRParser(spec.Grammar, allowAmbig)
	case parse.LL1:
		if allowAmbig {
			return nil, nil, fmt.Errorf("LL(k) parsers do not support ambiguous grammars")
		}

		p, err = ictiobus.NewLLParser(spec.Grammar)
	default:
		return nil, nil, fmt.Errorf("unsupported parser type: %s", t)
	}

	if err != nil {
		return nil, nil, err
	}

	for _, warn := range ambigWarns {
		warns = append(warns, Warning{
			Type:    WarnAmbiguousGrammar,
			Message: warn,
		})
	}

	return p, warns, nil
}

// CreateSDTS uses the TranslationScheme in the spec to create a new SDTS.
func (spec Spec) CreateSDTS() (trans.SDTS, error) {
	sdts := ictiobus.NewSDTS()

	for _, sdd := range spec.TranslationScheme {
		if sdd.Attribute.Relation.Type == trans.RelHead {
			err := sdts.Bind(sdd.Rule.NonTerminal, sdd.Rule.Productions[0], sdd.Attribute.Name, sdd.Hook, sdd.Args)
			if err != nil {
				return nil, fmt.Errorf("cannot bind synthesized attribute for %s: %w", sdd.Attribute, err)
			}
		} else {
			err := sdts.BindI(sdd.Rule.NonTerminal, sdd.Rule.Productions[0], sdd.Attribute.Name, sdd.Hook, sdd.Args, sdd.Attribute.Relation)
			if err != nil {
				return nil, fmt.Errorf("cannot bind inherited attribute for %s: %w", sdd.Attribute, err)
			}
		}
	}

	return sdts, nil
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
	Attribute trans.AttrRef

	// Rule is the grammar haed and production that this SDD is for. This will
	// have exactly one production in it, as opposed to Rule structs stored in
	// the grammar, which contain *all* productions for a given head symbol.
	Rule grammar.Rule

	// Hook is the name of the hook that is called to get the value to set on
	// the attribute.
	Hook string

	// Args is the list of arguments to the hook.
	Args []trans.AttrRef
}

// NewSpec reads an AST and creates a LanguageSpec from it. If the AST has
// any errors, then an error is returned. If the AST has non-error warnings,
// they will be returned in the warnings slice. Warnings will be present and
// valid even if err is non-nil; spec will only be valid if err is nil.
//
// Uses the Options to determine what to validate. Only the SDTS options are
// recognized at this time.
func NewSpec(ast AST) (spec Spec, warnings []Warning, err error) {
	ls := Spec{
		Patterns: make(map[string][]Pattern),
	}

	// all tokens blocks must be processed before any grammar blocks, and all
	// grammar blocks must be processed before any actions blocks.

	// first, gather each type of AST block into a single listing
	tokensBlocks := []syntax.TokensContent{}
	grammarBlocks := []syntax.GrammarContent{}
	actionsBlocks := []syntax.ActionsContent{}

	for _, bl := range ast.Nodes {
		switch bl := bl.(type) {
		case syntax.TokensBlock:
			tokBl := bl.Tokens()
			tokensBlocks = append(tokensBlocks, tokBl.Content...)
		case syntax.GrammarBlock:
			gramBl := bl.Grammar()
			grammarBlocks = append(grammarBlocks, gramBl.Content...)
		case syntax.ActionsBlock:
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
	actionsBlocks []syntax.ActionsContent,
	g grammar.CFG,
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
				synErr := lex.NewSyntaxErrorFromToken(fmt.Sprintf("'%s' is not a non-terminal symbol in the grammar", ruleHead), symAct.SrcSym)
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
						synErr := lex.NewSyntaxErrorFromToken(errMsg, prodAct.SrcVal)
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
						if sym != "" && sym[0] == '{' && sym[len(sym)-1] == '}' {
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
						synErr := lex.NewSyntaxErrorFromToken(errMsg, prodAct.SrcVal)
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
						synErr := lex.NewSyntaxErrorFromToken("invalid attrRef: "+err.Error(), semAct.LHS.Src)
						return nil, warnings, synErr
					}
					// if the lhs is not the head node, it is by definition an inherited attribute, not supported
					if sdd.Attribute.Relation.Type != trans.RelHead {
						w := Warning{
							Type:    WarnEFInheritedAttributes,
							Message: fmt.Sprintf("SDTS rule for %s: %s is an inherited attribute (not an attribute of {^})", sdd.Rule.String(), sdd.Attribute),
						}
						warnings = append(warnings, w)
					}

					// make shore we aren't trying to set somefin starting with a '$'; those are reserved
					if strings.HasPrefix(sdd.Attribute.Name, "$") {
						synErr := lex.NewSyntaxErrorFromToken("cannot create attribute starting with reserved marker '$'", semAct.LHS.Src)
						return nil, warnings, synErr
					}

					// do the same for each arg to the hook
					if len(semAct.With) > 0 {
						sdd.Args = make([]trans.AttrRef, len(semAct.With))
						for i, arg := range semAct.With {
							sdd.Args[i], err = attrRefFromASTAttrRef(arg, g, sddRule)
							if err != nil {
								synErr := lex.NewSyntaxErrorFromToken("invalid attrRef: "+err.Error(), arg.Src)
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
	grammarBlocks []syntax.GrammarContent,
	classes map[string]lex.TokenClass,
) (grammar.CFG, []Warning, error) {
	var warnings []Warning

	g := grammar.CFG{}
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
					// epsilons should be left alone
					if sym != "" {
						if sym[0] == '{' && sym[len(sym)-1] == '}' {
							// if it's wrapped in braces, it's a non-terminal; drop braces and make upper-case
							sym = strings.ToUpper(sym[1 : len(sym)-1])
						} else {
							// else, it's a terminal, make lower-case...
							sym = strings.ToLower(sym)

							// ...and make sure it's in the lexer's terminals
							if _, ok := classes[sym]; !ok {
								synErr := lex.NewSyntaxErrorFromToken(fmt.Sprintf("terminal '%s' is not a defined token class in any tokens block", sym), rule.Src)
								return g, warnings, synErr
							}

							// get token class
							tc := classes[sym]
							g.AddTerm(sym, tc)

							// mark it as seen
							seenTerminals[sym] = true
						}
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
		return g, warnings, fmt.Errorf("invalid grammar: %w", err)
	}

	return g, warnings, nil
}

func analzyeASTTokensContentSlice(
	tokensBlocks []syntax.TokensContent,
	existingStates box.StringSet,
	classes map[string]lex.TokenClass,
) (map[string][]Pattern, []Warning, error) {
	var warnings []Warning

	pats := make(map[string][]Pattern)

	for _, tokBl := range tokensBlocks {
		for _, entry := range tokBl.Entries {
			var p Pattern

			// either an entry specifies discard, OR it specifies up to one each
			// of stateshift, token, human. priority may be in either.
			var err error

			// get the pattern
			p.Regex, err = regexp.Compile(entry.Pattern)
			if err != nil {
				synErr := lex.NewSyntaxErrorFromToken(fmt.Sprintf("invalid regular expression: %s", err.Error()), entry.Src)
				return nil, warnings, synErr
			}

			// make sure we only have one maximum of each option
			if len(entry.SrcDiscard) > 1 {
				synErr := lex.NewSyntaxErrorFromToken("duplicate discard directive for entry", entry.SrcDiscard[1])
				return nil, warnings, synErr
			}
			if len(entry.SrcHuman) > 1 {
				synErr := lex.NewSyntaxErrorFromToken("duplicate human directive for entry", entry.SrcHuman[1])
				return nil, warnings, synErr
			}
			if len(entry.SrcPriority) > 1 {
				synErr := lex.NewSyntaxErrorFromToken("duplicate priority directive for entry", entry.SrcPriority[1])
				return nil, warnings, synErr
			}
			if len(entry.SrcShift) > 1 {
				synErr := lex.NewSyntaxErrorFromToken("duplicate state shift directive for entry", entry.SrcShift[1])
				return nil, warnings, synErr
			}

			// make sure mutually exclusive options are not used
			if entry.Discard {
				// then there'd betta not be a human directive, a token
				// directive, or a shift directive.

				// error report on the *2nd* token to break things

				firstTok := entry.SrcDiscard[0]
				firstIsDiscard := true
				var secondTok lex.Token

				if len(entry.SrcHuman) > 0 {
					humanTok := entry.SrcHuman[0]
					putEntryTokenInCorrectPosForDiscardCheck(&firstTok, &secondTok, &firstIsDiscard, humanTok)
				}
				if len(entry.SrcToken) > 0 {
					tokenTok := entry.SrcToken[0]
					putEntryTokenInCorrectPosForDiscardCheck(&firstTok, &secondTok, &firstIsDiscard, tokenTok)
				}
				if len(entry.SrcShift) > 0 {
					srcShift := entry.SrcShift[0]
					putEntryTokenInCorrectPosForDiscardCheck(&firstTok, &secondTok, &firstIsDiscard, srcShift)
				}

				if secondTok != nil {
					var fullErrMsg string
					if firstIsDiscard {
						errMsg := "human/token/stateshift directive cannot be added to discarded entry:"
						synErr1 := lex.NewSyntaxErrorFromToken("initial discard defined here", firstTok)
						synErr2 := lex.NewSyntaxErrorFromToken("directive not allowed", secondTok)

						fullErrMsg = errMsg + "\n" + synErr1.FullMessage() + "\n" + synErr2.FullMessage()
					} else {
						errMsg := "can't discard an entry that will be used for stateshift or token lexing:"
						synErr1 := lex.NewSyntaxErrorFromToken("initial directive defined here", firstTok)
						synErr2 := lex.NewSyntaxErrorFromToken("discard directive not allowed", secondTok)

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
						synErr := lex.NewSyntaxErrorFromToken("human directive given without token directive", entry.SrcHuman[0])
						return nil, warnings, synErr
					}
				}

				if entry.Token == "" && entry.Shift == "" {
					synErr := lex.NewSyntaxErrorFromToken("entry must have a discard, token, or stateshift directive", entry.Src)
					return nil, warnings, synErr
				}

				// don't try to shift to non-existent state
				if entry.Shift != "" {
					if !existingStates.Has(entry.Shift) {
						synErr := lex.NewSyntaxErrorFromToken("bad stateshift; shifted-to-state does not exist", entry.SrcShift[0])
						return nil, warnings, synErr
					}

					if entry.Shift == tokBl.State {
						synErr := lex.NewSyntaxErrorFromToken("bad stateshift; already in that state", entry.SrcShift[0])
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
			if len(entry.SrcPriority) > 0 {
				if entry.Priority == 0 {
					warn := lex.NewSyntaxErrorFromToken("setting priority to 0 has no effect", entry.SrcPriority[0])
					warnings = append(warnings, Warning{
						Type:    WarnPriorityZero,
						Message: warn.FullMessage(),
					})
				} else if entry.Priority < 0 {
					synErr := lex.NewSyntaxErrorFromToken("priority cannot be negative", entry.SrcPriority[0])
					return nil, warnings, synErr
				}

				p.Priority = entry.Priority
			}

			// add the pattern to the lexer
			statePats, ok := pats[tokBl.State]
			if !ok {
				statePats = make([]Pattern, 0)
			}
			statePats = append(statePats, p)
			pats[tokBl.State] = statePats
		}
	}

	return pats, warnings, nil
}

// r is rule to check against, only first production is checked.
func attrRefFromASTAttrRef(astRef syntax.AttrRef, g grammar.CFG, r grammar.Rule) (trans.AttrRef, error) {
	var ar trans.AttrRef
	if astRef.Head {
		ar = trans.AttrRef{
			Relation: trans.NodeRelation{
				Type: trans.RelHead,
			},
			Name: astRef.Attribute,
		}
	} else if astRef.SymInProd {
		// make sure the rule has the right number of symbols
		if astRef.Occurance >= len(r.Productions[0]) {
			symCount := textfmt.Pluralize(len(r.Productions[0]), "symbol", "-s")
			return trans.AttrRef{}, fmt.Errorf("symbol index out of range; production only has %s (%s)", symCount, r.Productions[0])
		}
		ar = trans.AttrRef{
			Relation: trans.NodeRelation{
				Type:  trans.RelSymbol,
				Index: astRef.Occurance,
			},
			Name: astRef.Attribute,
		}
	} else if astRef.NontermInProd {
		// make sure that the rule has that number of non-terminals
		nontermCount := 0
		for _, sym := range r.Productions[0] {
			if g.IsNonTerminal(sym) {
				nontermCount++
			}
		}
		if astRef.Occurance >= nontermCount {
			return trans.AttrRef{}, fmt.Errorf("non-terminal index out of range; production only has %d non-terminals", nontermCount)
		}
		ar = trans.AttrRef{
			Relation: trans.NodeRelation{
				Type:  trans.RelNonTerminal,
				Index: astRef.Occurance,
			},
			Name: astRef.Attribute,
		}
	} else if astRef.TermInProd {
		// make sure that the rule has that number of terminals
		termCount := 0
		for _, sym := range r.Productions[0] {
			if g.IsTerminal(sym) {
				termCount++
			}
		}
		if astRef.Occurance >= termCount {
			return trans.AttrRef{}, fmt.Errorf("terminal index out of range; production only has %d terminals", termCount)
		}
		ar = trans.AttrRef{
			Relation: trans.NodeRelation{
				Type:  trans.RelTerminal,
				Index: astRef.Occurance,
			},
			Name: astRef.Attribute,
		}
	} else {
		// it's an instance of a particular symbol. find out the symbol index.
		symIndexes := []int{}
		for i, sym := range r.Productions[0] {
			if sym == astRef.Symbol {
				symIndexes = append(symIndexes, i)
			}
		}
		if len(symIndexes) == 0 {
			return trans.AttrRef{}, fmt.Errorf("no symbol %s in production", astRef.Symbol)
		}
		if astRef.Occurance >= len(symIndexes) {
			return trans.AttrRef{}, fmt.Errorf("symbol index out of range; production only has %d instances of %s", len(symIndexes), astRef.Symbol)
		}
		ar = trans.AttrRef{
			Relation: trans.NodeRelation{
				Type:  trans.RelSymbol,
				Index: symIndexes[astRef.Occurance],
			},
			Name: astRef.Attribute,
		}
	}

	valErr := validateParsedAttrRef(ar, g, r)
	if valErr != nil {
		return ar, valErr
	}
	return ar, nil
}

// it is assumed that the parsed ref refers to an existing symbol; not checking
// that here.
func validateParsedAttrRef(ar trans.AttrRef, g grammar.CFG, r grammar.Rule) error {
	// need to know if we are dealing with a terminal or not
	isTerm := false

	// no need to check RelHead; by definition this cannot be a terminal
	if ar.Relation.Type == trans.RelSymbol {
		isTerm = g.IsTerminal(r.Productions[0][ar.Relation.Index])
	} else {
		isTerm = ar.Relation.Type == trans.RelTerminal
	}

	symName, err := ar.ResolveSymbol(g, r.NonTerminal, r.Productions[0])
	if err != nil {
		panic("unresolvable attrRef passed to validateParsedAttrRef")
	}

	if isTerm {
		// then anyfin we take from it, glub, must start with '$'
		if ar.Name != "$text" && ar.Name != "$ft" {
			return fmt.Errorf("referred-to terminal %q only has '$text' and '$ft' attributes", symName)
		}
	} else {
		// then we cannot take '$text' from it and it's an error.
		if ar.Name == "$text" {
			return fmt.Errorf("referred-to non-terminal %q does not have lexed text attribute \"$text\"", symName)
		}
	}

	return nil
}

func putEntryTokenInCorrectPosForDiscardCheck(first, second *lex.Token, discardIsFirst *bool, tok lex.Token) {
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

func tokenIsBefore(t1, t2 lex.Token) bool {
	return t1.Line() < t2.Line() || (t1.Line() == t2.Line() && t1.LinePos() < t2.LinePos())
}

func buildTokenClasses(tcDefsTable map[string][]box.Pair[string, lex.Token], states box.StringSet) (map[string]lex.TokenClass, []Warning) {
	var warnings []Warning

	classes := make(map[string]lex.TokenClass)
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

func checkForDuplicateHumanDefs(tcSymTable map[string][]box.Pair[string, lex.Token]) []Warning {
	var warnings []Warning
	// warn for duplicate human definitions (but not missing; we will handle
	// that during reading of tokenBlocks)
	for tok, humanDefs := range tcSymTable {
		if len(humanDefs) > 1 {
			msgStart := fmt.Sprintf("multiple distinct human-readable names given for token %q:", tok)
			var msg string
			for _, hd := range humanDefs {
				synErr := lex.NewSyntaxErrorFromToken("human name last defined here", hd.Second)
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
func scanTokenClasses(blocks []syntax.TokensContent) (map[string][]box.Pair[string, lex.Token], box.StringSet) {

	// tcSymTable is tok-name -> pairs of human-name and token where that human
	// name is first defined. Uses slice of pairs instead of map to preserve
	// order.
	tcSymTable := make(map[string][]box.Pair[string, lex.Token])

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
					humanDefs = []box.Pair[string, lex.Token]{}
					tcSymTable[entry.Token] = humanDefs
				}
				if entry.Human != "" {
					keepHuman := true

					// if we already have human definition(s) for this token,
					// only add this one if it is a distinct string.
					if len(humanDefs) != 0 {
						_, alreadyExists := slices.Any(humanDefs, func(hd box.Pair[string, lex.Token]) bool {
							return hd.First == entry.Human
						})
						// only need to save a new one if it doesn't already exist
						keepHuman = !alreadyExists
					}

					if keepHuman {
						humanDefs = append(humanDefs, box.PairOf(entry.Human, entry.SrcHuman[len(entry.SrcHuman)-1]))
						tcSymTable[entry.Token] = humanDefs
					}
				}
			}
		}
	}

	return tcSymTable, states
}
