package fishi

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"
)

func CreateBootstrapLexer() ictiobus.Lexer {
	bootLx := ictiobus.NewLexer()

	// default state, shared by all
	bootLx.RegisterClass(tcEscseq, "")
	bootLx.RegisterClass(tcHeaderTokens, "")
	bootLx.RegisterClass(tcHeaderGrammar, "")
	bootLx.RegisterClass(tcHeaderActions, "")
	//bootLx.RegisterClass(tcDirStart, "")
	bootLx.RegisterClass(tcDirState, "")

	// default patterns and defs
	bootLx.AddPattern(`%!.`, lex.LexAs(tcEscseq.ID()), "", 0)
	bootLx.AddPattern(`%%[Tt][Oo][Kk][Ee][Nn][Ss]`, lex.LexAndSwapState(tcHeaderTokens.ID(), "tokens"), "", 0)
	bootLx.AddPattern(`%%[Gg][Rr][Aa][Mm][Mm][Aa][Rr]`, lex.LexAndSwapState(tcHeaderGrammar.ID(), "grammar"), "", 0)
	bootLx.AddPattern(`%%[Aa][Cc][Tt][Ii][Oo][Nn][Ss]`, lex.LexAndSwapState(tcHeaderActions.ID(), "actions"), "", 0)
	//bootLx.AddPattern(`%[Ss][Tt][Aa][Rr][Tt]`, lex.LexAs(tcDirStart.ID()), "")
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee]`, lex.LexAs(tcDirState.ID()), "", 0)

	// tokens classes
	bootLx.RegisterClass(tcFreeformText, "tokens")
	bootLx.RegisterClass(tcDirShift, "tokens")
	bootLx.RegisterClass(tcDirHuman, "tokens")
	bootLx.RegisterClass(tcDirToken, "tokens")
	//bootLx.RegisterClass(tcDirDefault, "tokens")
	bootLx.RegisterClass(tcNewline, "tokens")
	bootLx.RegisterClass(tcDirDiscard, "tokens")
	bootLx.RegisterClass(tcDirPriority, "tokens")
	bootLx.RegisterClass(tcInt, "tokens")

	// tokens patterns
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee][Ss][Hh][Ii][Ff][Tt]`, lex.LexAs(tcDirShift.ID()), "tokens", 1)
	bootLx.AddPattern(`%[Hh][Uu][Mm][Aa][Nn]`, lex.LexAs(tcDirHuman.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Tt][Oo][Kk][Ee][Nn]`, lex.LexAs(tcDirToken.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Dd][Ii][Ss][Cc][Aa][Rr][Dd]`, lex.LexAs(tcDirDiscard.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Pp][Rr][Ii][Oo][Rr][Ii][Tt][Yy]`, lex.LexAs(tcDirPriority.ID()), "tokens", 0)
	//bootLx.AddPattern(`%[Dd][Ee][Ff][Aa][Uu][Ll][Tt]`, lex.LexAs(tcDirDefault.ID()), "tokens")
	bootLx.AddPattern(`\n`, lex.LexAs(tcNewline.ID()), "tokens", 0)
	bootLx.AddPattern(`[^%\s]+[^%\n]*`, lex.LexAs(tcFreeformText.ID()), "tokens", 0)
	bootLx.AddPattern(`[^\S\n]+`, lex.Discard(), "tokens", 0)

	// grammar classes
	bootLx.RegisterClass(tcNewline, "grammar")
	bootLx.RegisterClass(tcEq, "grammar")
	bootLx.RegisterClass(tcAlt, "grammar")
	bootLx.RegisterClass(tcNonterminal, "grammar")
	bootLx.RegisterClass(tcTerminal, "grammar")
	bootLx.RegisterClass(tcEpsilon, "grammar")

	// gramamr patterns
	bootLx.AddPattern(`\n`, lex.LexAs(tcNewline.ID()), "grammar", 0)
	bootLx.AddPattern(`[^\S\n]+`, lex.Discard(), "grammar", 0)
	bootLx.AddPattern(`\|`, lex.LexAs(tcAlt.ID()), "grammar", 0)
	bootLx.AddPattern(`{}`, lex.LexAs(tcEpsilon.ID()), "grammar", 0)
	bootLx.AddPattern(`{[A-Za-z][^}]*}`, lex.LexAs(tcNonterminal.ID()), "grammar", 0)
	bootLx.AddPattern(`[^=\s]\S*|\S\S+`, lex.LexAs(tcTerminal.ID()), "grammar", 0)
	bootLx.AddPattern(`=`, lex.LexAs(tcEq.ID()), "grammar", 0)

	// actions classes
	bootLx.RegisterClass(tcAttrRef, "actions")
	bootLx.RegisterClass(tcInt, "actions")
	bootLx.RegisterClass(tcNonterminal, "actions")
	bootLx.RegisterClass(tcDirSymbol, "actions")
	bootLx.RegisterClass(tcDirProd, "actions")
	bootLx.RegisterClass(tcDirWith, "actions")
	bootLx.RegisterClass(tcDirHook, "actions")
	bootLx.RegisterClass(tcDirAction, "actions")
	bootLx.RegisterClass(tcDirIndex, "actions")
	bootLx.RegisterClass(tcId, "actions")
	bootLx.RegisterClass(tcTerminal, "actions")
	bootLx.RegisterClass(tcEpsilon, "actions")

	// actions patterns
	bootLx.AddPattern(`\s+`, lex.Discard(), "actions", 0)
	bootLx.AddPattern(`(?:{[A-Za-z][^}]*}|\S+)(?:\$\d+)?\.[\$A-Za-z][$A-Za-z0-9_-]*`, lex.LexAs(tcAttrRef.ID()), "actions", 0)
	bootLx.AddPattern(`[0-9]+`, lex.LexAs(tcInt.ID()), "actions", 0)
	bootLx.AddPattern(`{[A-Za-z][^}]*}`, lex.LexAs(tcNonterminal.ID()), "actions", 0)
	bootLx.AddPattern(`%[Ss][Yy][Mm][Bb][Oo][Ll]`, lex.LexAs(tcDirSymbol.ID()), "actions", 0)
	bootLx.AddPattern(`%[Pp][Rr][Oo][Dd]`, lex.LexAs(tcDirProd.ID()), "actions", 0)
	bootLx.AddPattern(`%[Ww][Ii][Tt][Hh]`, lex.LexAs(tcDirWith.ID()), "actions", 0)
	bootLx.AddPattern(`%[Hh][Oo][Oo][Kk]`, lex.LexAs(tcDirHook.ID()), "actions", 0)
	bootLx.AddPattern(`%[Aa][Cc][Tt][Ii][Oo][Nn]`, lex.LexAs(tcDirAction.ID()), "actions", 0)
	bootLx.AddPattern(`%[Ii][Nn][Dd][Ee][Xx]`, lex.LexAs(tcDirIndex.ID()), "actions", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAs(tcId.ID()), "actions", 0)
	bootLx.AddPattern(`{}`, lex.LexAs(tcEpsilon.ID()), "actions", 0)
	bootLx.AddPattern(`\S+`, lex.LexAs(tcTerminal.ID()), "actions", 0)

	return bootLx
}
