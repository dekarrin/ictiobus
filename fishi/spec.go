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

	// LexerPatterns is a map of state names to the lexer patterns that are used
	// while in that state.
	LexerMatches map[string]Pattern

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
