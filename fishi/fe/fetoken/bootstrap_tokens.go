// Package fetok contains the token classes for the frontend. It is in a
// separate package so that it can be imported and used by external packages
// while still allowing those external packages to be imported by the rest of
// the frontend, such as the fishi syntax package.
package fetoken

import (
	"github.com/dekarrin/ictiobus/lex"
)

var (

	// %default is not in this version, not needed to self-describe
	//TCDirDefault    = lex.NewTokenClass("default_dir", "'default' directive")

	// %start is not in this version, not needed to self-describe
	//TCDirStart      = lex.NewTokenClass("start_dir", "'start' directive")

	TCHeaderTokens          = lex.NewTokenClass("tokens_header", "'tokens' header")
	TCHeaderGrammar         = lex.NewTokenClass("grammar_header", "'grammar' header")
	TCHeaderActions         = lex.NewTokenClass("actions_header", "'actions' header")
	TCDirSet                = lex.NewTokenClass("set_dir", "'set' directive ':'")
	TCDirDiscard            = lex.NewTokenClass("discard", "%discard")
	TCDirHook               = lex.NewTokenClass("hook_dir", "'hook' directive")
	TCDirHuman              = lex.NewTokenClass("human_dir", "'human' directive")
	TCDirIndex              = lex.NewTokenClass("index_dir", "'index' directive")
	TCDirProd               = lex.NewTokenClass("prod_dir", "'prod' directive '->'")
	TCDirShift              = lex.NewTokenClass("shift_dir", "'stateshift' directive")
	TCDirPriority           = lex.NewTokenClass("priority_dir", "'priority' directive")
	TCDirState              = lex.NewTokenClass("state_dir", "'state' directive")
	TCDirSymbol             = lex.NewTokenClass("symbol_dir", "'symbol' directive")
	TCDirToken              = lex.NewTokenClass("token_dir", "'token' directive")
	TCDirWith               = lex.NewTokenClass("with_dir", "'with' directive")
	TCFreeformText          = lex.NewTokenClass("freeform_text", "freeform text value")
	TCLineStartFreeformText = lex.NewTokenClass("line_start_freeform_text", "freeform text value at line start")
	//TCNewline       = lex.NewTokenClass("newline", "'\\n'")
	TCTerminal             = lex.NewTokenClass("terminal", "terminal symbol")
	TCNonterminal          = lex.NewTokenClass("nonterminal", "non-terminal symbol")
	TCEq                   = lex.NewTokenClass("eq", "'='")
	TCAlt                  = lex.NewTokenClass("alt", "'|'")
	TCAttrRef              = lex.NewTokenClass("attr_ref", "attribute reference")
	TCInt                  = lex.NewTokenClass("int", "integer value")
	TCId                   = lex.NewTokenClass("id", "identifier")
	TCEscseq               = lex.NewTokenClass("escseq", "escape sequence")
	TCLineStartEscseq      = lex.NewTokenClass("line_start_escseq", "escape sequence at line start")
	TCEpsilon              = lex.NewTokenClass("epsilon", "epsilon production")
	TCLineStartNonterminal = lex.NewTokenClass("line_start_nonterminal", "non-terminal symbol at line start")
)
