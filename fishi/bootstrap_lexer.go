package fishi

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"
)

func CreateBootstrapLexer() ictiobus.Lexer {
	bootLx := ictiobus.NewLazyLexer()

	// default state, shared by all
	bootLx.RegisterClass(tcEscseq, "")
	bootLx.RegisterClass(tcHeaderTokens, "")
	bootLx.RegisterClass(tcHeaderGrammar, "")
	bootLx.RegisterClass(tcHeaderActions, "")
	//bootLx.RegisterClass(tcDirStart, "")

	// default patterns and defs
	bootLx.AddPattern(`%!.`, lex.LexAs(tcEscseq.ID()), "", 0)
	bootLx.AddPattern(`%%[Tt][Oo][Kk][Ee][Nn][Ss]`, lex.LexAndSwapState(tcHeaderTokens.ID(), "tokens"), "", 0)
	bootLx.AddPattern(`%%[Gg][Rr][Aa][Mm][Mm][Aa][Rr]`, lex.LexAndSwapState(tcHeaderGrammar.ID(), "grammar"), "", 0)
	bootLx.AddPattern(`%%[Aa][Cc][Tt][Ii][Oo][Nn][Ss]`, lex.LexAndSwapState(tcHeaderActions.ID(), "actions"), "", 0)
	//bootLx.AddPattern(`%[Ss][Tt][Aa][Rr][Tt]`, lex.LexAs(tcDirStart.ID()), "")

	// t-a-g, you're it, glub glub glubglubglub glub 38D

	// Sure would 8e gr8 if we had a st8 stack ::::/

	bootLx.RegisterClass(tcId, "state-t")
	bootLx.RegisterClass(tcId, "state-a")
	bootLx.RegisterClass(tcId, "state-g")
	bootLx.AddPattern(`\s+`, lex.Discard(), "state-t", 0)
	bootLx.AddPattern(`\s+`, lex.Discard(), "state-a", 0)
	bootLx.AddPattern(`\s+`, lex.Discard(), "state-g", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAndSwapState(tcId.ID(), "tokens"), "state-t", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAndSwapState(tcId.ID(), "actions"), "state-a", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAndSwapState(tcId.ID(), "tokens"), "state-g", 0)

	// tokens classes
	bootLx.RegisterClass(tcFreeformText, "tokens")
	bootLx.RegisterClass(tcDirShift, "tokens")
	bootLx.RegisterClass(tcDirHuman, "tokens")
	bootLx.RegisterClass(tcDirToken, "tokens")
	//bootLx.RegisterClass(tcDirDefault, "tokens")
	bootLx.RegisterClass(tcDirDiscard, "tokens")
	bootLx.RegisterClass(tcDirPriority, "tokens")
	bootLx.RegisterClass(tcInt, "tokens")
	bootLx.RegisterClass(tcDirState, "tokens")
	bootLx.RegisterClass(tcLineStartFreeformText, "tokens")
	bootLx.RegisterClass(tcLineStartEscseq, "tokens")

	// tokens patterns
	bootLx.AddPattern(`\n\s*%!.`, lex.LexAs(tcLineStartEscseq.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee]`, lex.LexAndSwapState(tcDirState.ID(), "state-t"), "tokens", 0)
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee][Ss][Hh][Ii][Ff][Tt]`, lex.LexAs(tcDirShift.ID()), "tokens", 1)
	bootLx.AddPattern(`%[Hh][Uu][Mm][Aa][Nn]`, lex.LexAs(tcDirHuman.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Tt][Oo][Kk][Ee][Nn]`, lex.LexAs(tcDirToken.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Dd][Ii][Ss][Cc][Aa][Rr][Dd]`, lex.LexAs(tcDirDiscard.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Pp][Rr][Ii][Oo][Rr][Ii][Tt][Yy]`, lex.LexAs(tcDirPriority.ID()), "tokens", 0)
	bootLx.AddPattern(`[^\S\n]+`, lex.Discard(), "tokens", 0)
	bootLx.AddPattern(`\n\s*[^%\s]+[^%\n]*`, lex.LexAs(tcLineStartFreeformText.ID()), "tokens", 0)
	bootLx.AddPattern(`\n`, lex.Discard(), "tokens", 0)
	bootLx.AddPattern(`[^%\s]+[^%\n]*`, lex.LexAs(tcFreeformText.ID()), "tokens", 0)
	//bootLx.AddPattern(`%[Dd][Ee][Ff][Aa][Uu][Ll][Tt]`, lex.LexAs(tcDirDefault.ID()), "tokens")

	// grammar classes
	bootLx.RegisterClass(tcEq, "grammar")
	bootLx.RegisterClass(tcAlt, "grammar")
	bootLx.RegisterClass(tcNonterminal, "grammar")
	bootLx.RegisterClass(tcTerminal, "grammar")
	bootLx.RegisterClass(tcEpsilon, "grammar")
	bootLx.RegisterClass(tcDirState, "grammar")
	bootLx.RegisterClass(tcLineStartNonterminal, "grammar")

	// grammar patterns
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee]`, lex.LexAndSwapState(tcDirState.ID(), "state-g"), "grammar", 0)
	bootLx.AddPattern(`[^\S\n]+`, lex.Discard(), "grammar", 0)
	bootLx.AddPattern(`\n\s*{[A-Za-z][^}]*}`, lex.LexAs(tcLineStartNonterminal.ID()), "grammar", 1)
	bootLx.AddPattern(`\n`, lex.Discard(), "grammar", 0)
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
	bootLx.RegisterClass(tcDirState, "actions")

	// actions patterns
	bootLx.AddPattern(`\s+`, lex.Discard(), "actions", 0)
	bootLx.AddPattern(`(?:{\*}|{[A-Za-z][^}]*}|\S+)(?:\$\d+)?\.[\$A-Za-z][$A-Za-z0-9_-]*`, lex.LexAs(tcAttrRef.ID()), "actions", 0)
	bootLx.AddPattern(`[0-9]+`, lex.LexAs(tcInt.ID()), "actions", 0)
	bootLx.AddPattern(`{[A-Za-z][^}]*}`, lex.LexAs(tcNonterminal.ID()), "actions", 0)
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee]`, lex.LexAndSwapState(tcDirState.ID(), "state-a"), "grammar", 0)
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
