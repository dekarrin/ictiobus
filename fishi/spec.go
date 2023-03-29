package fishi

import (
	"regexp"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/translation"
	"github.com/dekarrin/ictiobus/types"
)

// LanguageSpec is a series of statements that together give the specification
// for the complete frontend of a language. It is created by processing an AST
// and checking it for errors.
type LanguageSpec struct {
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

// CreateSpec reads an AST and creates a LanguageSpec from it. If the AST has
// any errors, then an error is returned.
func CreateSpec(ast AST) (LanguageSpec, error) {
	ls := LanguageSpec{
		Patterns: make(map[string]Pattern),
	}

	// first, gather each type of AST block into a single listing (per state)
	tokensBlocks := map[string][]ASTTokensContent{}
	grammarBlocks := map[string][]ASTGrammarContent{}
	actionsBlocks := map[string][]ASTActionsContent{}

	for _, bl := range ast.Nodes {
		switch bl := bl.(type) {
		case ASTTokensBlock:
			tokBl := bl.Tokens()

			for i := range tokBl.Content {
				tokCont := tokBl.Content[i]
				tokSlice, ok := tokensBlocks[tokCont.State]
				if !ok {
					tokSlice = []ASTTokensContent{}
				}
				tokSlice = append(tokSlice, tokCont)
				tokensBlocks[tokCont.State] = tokSlice
			}
		case ASTGrammarBlock:
			gramBl := bl.Grammar()

			for i := range gramBl.Content {
				gramCont := gramBl.Content[i]
				gramSlice, ok := grammarBlocks[gramCont.State]
				if !ok {
					gramSlice = []ASTGrammarContent{}
				}
				gramSlice = append(gramSlice, gramCont)
				grammarBlocks[gramCont.State] = gramSlice
			}
		case ASTActionsBlock:
			actBl := bl.Actions()

			for i := range actBl.Content {
				actCont := actBl.Content[i]
				actSlice, ok := actionsBlocks[actCont.State]
				if !ok {
					actSlice = []ASTActionsContent{}
				}
				actSlice = append(actSlice, actCont)
				actionsBlocks[actCont.State] = actSlice
			}
		}
	}

	// go over tokensBlocks
	for i := range tokensBlocks {
		bl := tokensBlocks[i]
	}

	// go over grammarBlocks

	// go over actionsBlocks

	return ls, nil
}
