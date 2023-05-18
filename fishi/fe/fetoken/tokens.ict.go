// Package fetoken contains the token classes used by the frontend
// of FISHI. It is in a separate package so that it can be imported and
// used by external packages while still allowing those external packages to be
// imported by the rest of the frontend.
package fetoken

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --lalr --ir github.com/dekarrin/ictiobus/fishi/syntax.AST --dest fishi/fe -l FISHI -v 1.0 --hooks fishi/syntax docs/fishi.md --dev
*/

import (
	"github.com/dekarrin/ictiobus/lex"
)

var (
	// TCAlt is the token class representing an alternations bar '|' in FISHI.
	TCAlt = lex.NewTokenClass("alt", "alternations bar '|'")

	// TCAttrRef is the token class representing an attribute reference literal in FISHI.
	TCAttrRef = lex.NewTokenClass("attr-ref", "attribute reference literal")

	// TCDirDiscard is the token class representing a %discard directive in FISHI.
	TCDirDiscard = lex.NewTokenClass("dir-discard", "%discard directive")

	// TCDirHook is the token class representing a %hook directive '=' in FISHI.
	TCDirHook = lex.NewTokenClass("dir-hook", "%hook directive '='")

	// TCDirHuman is the token class representing a %human directive in FISHI.
	TCDirHuman = lex.NewTokenClass("dir-human", "%human directive")

	// TCDirIndex is the token class representing a %index directive in FISHI.
	TCDirIndex = lex.NewTokenClass("dir-index", "%index directive")

	// TCDirPriority is the token class representing a %priority directive in FISHI.
	TCDirPriority = lex.NewTokenClass("dir-priority", "%priority directive")

	// TCDirProd is the token class representing a %prod directive '->' in FISHI.
	TCDirProd = lex.NewTokenClass("dir-prod", "%prod directive '->'")

	// TCDirSet is the token class representing a %set directive ':' in FISHI.
	TCDirSet = lex.NewTokenClass("dir-set", "%set directive ':'")

	// TCDirShift is the token class representing a %stateshift directive in FISHI.
	TCDirShift = lex.NewTokenClass("dir-shift", "%stateshift directive")

	// TCDirState is the token class representing a %state directive in FISHI.
	TCDirState = lex.NewTokenClass("dir-state", "%state directive")

	// TCDirSymbol is the token class representing a %symbol directive in FISHI.
	TCDirSymbol = lex.NewTokenClass("dir-symbol", "%symbol directive")

	// TCDirToken is the token class representing a %token directive in FISHI.
	TCDirToken = lex.NewTokenClass("dir-token", "%token directive")

	// TCDirWith is the token class representing a %with directive '(' in FISHI.
	TCDirWith = lex.NewTokenClass("dir-with", "%with directive '('")

	// TCEpsilon is the token class representing an epsilon production '{}' in FISHI.
	TCEpsilon = lex.NewTokenClass("epsilon", "epsilon production '{}'")

	// TCEq is the token class representing a rule production operator '=' in FISHI.
	TCEq = lex.NewTokenClass("eq", "rule production operator '='")

	// TCEscseq is the token class representing An Escape Sequence in FISHI.
	TCEscseq = lex.NewTokenClass("escseq", "Escape Sequence")

	// TCFreeformText is the token class representing a freeform text in FISHI.
	TCFreeformText = lex.NewTokenClass("freeform-text", "freeform text")

	// TCHdrActions is the token class representing a %%actions header in FISHI.
	TCHdrActions = lex.NewTokenClass("hdr-actions", "%%actions header")

	// TCHdrGrammar is the token class representing a %%grammar header in FISHI.
	TCHdrGrammar = lex.NewTokenClass("hdr-grammar", "%%grammar header")

	// TCHdrTokens is the token class representing a %%tokens header in FISHI.
	TCHdrTokens = lex.NewTokenClass("hdr-tokens", "%%tokens header")

	// TCId is the token class representing an identifier in FISHI.
	TCId = lex.NewTokenClass("id", "identifier")

	// TCInt is the token class representing an integer literal in FISHI.
	TCInt = lex.NewTokenClass("int", "integer literal")

	// TCNlEscseq is the token class representing an escape sequence in FISHI.
	TCNlEscseq = lex.NewTokenClass("nl-escseq", "escape sequence")

	// TCNlFreeformText is the token class representing a freeform text in FISHI.
	TCNlFreeformText = lex.NewTokenClass("nl-freeform-text", "freeform text")

	// TCNlNonterm is the token class representing a non-terminal symbol literal after this line in FISHI.
	TCNlNonterm = lex.NewTokenClass("nl-nonterm", "non-terminal symbol literal after this line")

	// TCNonterm is the token class representing a non-terminal symbol literal in FISHI.
	TCNonterm = lex.NewTokenClass("nonterm", "non-terminal symbol literal")

	// TCTerm is the token class representing a terminal symbol literal in FISHI.
	TCTerm = lex.NewTokenClass("term", "terminal symbol literal")
)

var all = map[string]lex.TokenClass{
	"alt":              TCAlt,
	"attr-ref":         TCAttrRef,
	"dir-discard":      TCDirDiscard,
	"dir-hook":         TCDirHook,
	"dir-human":        TCDirHuman,
	"dir-index":        TCDirIndex,
	"dir-priority":     TCDirPriority,
	"dir-prod":         TCDirProd,
	"dir-set":          TCDirSet,
	"dir-shift":        TCDirShift,
	"dir-state":        TCDirState,
	"dir-symbol":       TCDirSymbol,
	"dir-token":        TCDirToken,
	"dir-with":         TCDirWith,
	"epsilon":          TCEpsilon,
	"eq":               TCEq,
	"escseq":           TCEscseq,
	"freeform-text":    TCFreeformText,
	"hdr-actions":      TCHdrActions,
	"hdr-grammar":      TCHdrGrammar,
	"hdr-tokens":       TCHdrTokens,
	"id":               TCId,
	"int":              TCInt,
	"nl-escseq":        TCNlEscseq,
	"nl-freeform-text": TCNlFreeformText,
	"nl-nonterm":       TCNlNonterm,
	"nonterm":          TCNonterm,
	"term":             TCTerm,
}

// ByID returns the TokenClass in FISHI that has the given ID. If no token
// class with that ID exists, nil is returned.
func ByID(id string) lex.TokenClass {
	return all[id]
}
