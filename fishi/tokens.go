package fishi

import "github.com/dekarrin/ictiobus/lex"

var (

	// %default is not in this version, not needed to self-describe
	//tcDirDefault    = lex.NewTokenClass("default_dir", "'default' directive")

	// %start is not in this version, not needed to self-describe
	//tcDirStart      = lex.NewTokenClass("start_dir", "'start' directive")

	tcHeaderTokens  = lex.NewTokenClass("tokens_header", "'tokens' header")
	tcHeaderGrammar = lex.NewTokenClass("grammar_header", "'grammar' header")
	tcHeaderActions = lex.NewTokenClass("actions_header", "'actions' header")
	tcDirAction     = lex.NewTokenClass("action_dir", "'action' directive")
	tcDirDiscard    = lex.NewTokenClass("discard", "%discard")
	tcDirHook       = lex.NewTokenClass("hook_dir", "'hook' directive")
	tcDirHuman      = lex.NewTokenClass("human_dir", "'human' directive")
	tcDirIndex      = lex.NewTokenClass("index_dir", "'index' directive")
	tcDirProd       = lex.NewTokenClass("prod_dir", "'prod' directive")
	tcDirShift      = lex.NewTokenClass("shift_dir", "'stateshift' directive")
	tcDirPriority   = lex.NewTokenClass("priority_dir", "'priority' directive")
	tcDirState      = lex.NewTokenClass("state_dir", "'state' directive")
	tcDirSymbol     = lex.NewTokenClass("symbol_dir", "'symbol' directive")
	tcDirToken      = lex.NewTokenClass("token_dir", "'token' directive")
	tcDirWith       = lex.NewTokenClass("with_dir", "'with' directive")
	tcFreeformText  = lex.NewTokenClass("freeform_text", "freeform text value")
	tcNewline       = lex.NewTokenClass("newline", "'\\n'")
	tcTerminal      = lex.NewTokenClass("terminal", "terminal symbol")
	tcNonterminal   = lex.NewTokenClass("nonterminal", "non-terminal symbol")
	tcEq            = lex.NewTokenClass("eq", "'='")
	tcAlt           = lex.NewTokenClass("alt", "'|'")
	tcAttrRef       = lex.NewTokenClass("attr_ref", "attribute reference")
	tcInt           = lex.NewTokenClass("int", "integer value")
	tcId            = lex.NewTokenClass("id", "identifier")
	tcEscseq        = lex.NewTokenClass("escseq", "escape sequence")
	tcEpsilon       = lex.NewTokenClass("epsilon", "epsilon production")
)
