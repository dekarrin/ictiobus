// Package fetoken contains the token classes for the frontend. It is in a
// separate package so that it can be imported and used by external packages
// while still allowing those external packages to be imported by the rest of
// the frontend, such as the fishi syntax package.
package fetoken

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --lalr --ir github.com/dekarrin/ictiobus/fishi/syntax.AST --dest fishi/fe -l FISHI -v 1.0.0 --hooks fishi/syntax fishi.md --dev
*/

import "github.com/dekarrin/ictiobus/lex"

var (
	TCAlt            = lex.NewTokenClass("alt", "alternations bar '|'")
	TCAttrRef        = lex.NewTokenClass("attr-ref", "attribute reference literal")
	TCDirDiscard     = lex.NewTokenClass("dir-discard", "%discard directive")
	TCDirHook        = lex.NewTokenClass("dir-hook", "%hook directive '='")
	TCDirHuman       = lex.NewTokenClass("dir-human", "%human directive")
	TCDirIndex       = lex.NewTokenClass("dir-index", "%index directive")
	TCDirPriority    = lex.NewTokenClass("dir-priority", "%priority directive")
	TCDirProd        = lex.NewTokenClass("dir-prod", "%prod directive '->'")
	TCDirSet         = lex.NewTokenClass("dir-set", "%set directive ':'")
	TCDirShift       = lex.NewTokenClass("dir-shift", "%stateshift directive")
	TCDirState       = lex.NewTokenClass("dir-state", "%state directive")
	TCDirSymbol      = lex.NewTokenClass("dir-symbol", "%symbol directive")
	TCDirToken       = lex.NewTokenClass("dir-token", "%token directive")
	TCDirWith        = lex.NewTokenClass("dir-with", "%with directive '('")
	TCEpsilon        = lex.NewTokenClass("epsilon", "epsilon production '{}'")
	TCEq             = lex.NewTokenClass("eq", "rule production operator '='")
	TCEscseq         = lex.NewTokenClass("escseq", "Escape Sequence")
	TCFreeformText   = lex.NewTokenClass("freeform-text", "freeform text")
	TCHdrActions     = lex.NewTokenClass("hdr-actions", "%%actions header")
	TCHdrGrammar     = lex.NewTokenClass("hdr-grammar", "%%grammar header")
	TCHdrTokens      = lex.NewTokenClass("hdr-tokens", "%%tokens header")
	TCId             = lex.NewTokenClass("id", "identifier")
	TCInt            = lex.NewTokenClass("int", "integer literal")
	TCNlEscseq       = lex.NewTokenClass("nl-escseq", "escape sequence after this line")
	TCNlFreeformText = lex.NewTokenClass("nl-freeform-text", "freeform text after this line")
	TCNlNonterm      = lex.NewTokenClass("nl-nonterm", "non-terminal symbol literal after this line")
	TCNonterm        = lex.NewTokenClass("nonterm", "non-terminal symbol literal")
	TCTerm           = lex.NewTokenClass("term", "terminal symbol literal")
)