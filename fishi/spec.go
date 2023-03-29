package fishi

import (
	"fmt"
	"regexp"

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
	Patterns map[string]Pattern

	// Grammar is the syntactical specification of the language.
	Grammar grammar.Grammar

	// TranslationScheme outlines the Syntax-Directed Translation Scheme for the
	// language by giving the instructions for each attribute.
	TranslationScheme []SDD
}

// Pattern is a lexer pattern that is used to match a token, along with the
// action that the lexer should take when it matches.
type Pattern struct {
	// Pattern is the regular expression that is used to match the token.
	Pattern *regexp.Regexp

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

	// Rule is the grammar rule that this SDD is for.
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
		Patterns: make(map[string]Pattern),
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

	// go over tokensBlocks to get token classes

	// pre-scan to prep for tokenBlocks
	tcDefsTable, states := scanTokenClasses(tokensBlocks)
	dupes := checkForDuplicateHumanDefs(tcDefsTable)
	if len(dupes) > 0 {
		warnings = append(warnings, dupes...)
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
	for _, tokBl := range tokensBlocks {
		for _, entry := range tokBl.Entries {
			// either an entry specifies discard, OR it specifies up to one each
			// of stateshift, token, human. priority may be in either.

			// make sure we only have one maximum of each option
			if len(entry.discardTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate discard directive for entry", entry.discardTok[1])
				return ls, warnings, synErr
			}
			if len(entry.humanTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate human directive for entry", entry.humanTok[1])
				return ls, warnings, synErr
			}
			if len(entry.priorityTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate priority directive for entry", entry.priorityTok[1])
				return ls, warnings, synErr
			}
			if len(entry.shiftTok) > 1 {
				synErr := types.NewSyntaxErrorFromToken("duplicate state shift directive for entry", entry.shiftTok[1])
				return ls, warnings, synErr
			}

			// make sure mutually exclusive options are not used
			if entry.Discard {
				// then there'd betta not be a human directive, a token
				// directive, or a shift directive. TODO: grammar should rly enforce this

			}
		}
	}

	// go over grammarBlocks to get grammar

	// go over actionsBlocks to get translation scheme

	return ls, warnings, nil
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
				msg += fmt.Sprintf("%s\n", synErr.Error())
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
