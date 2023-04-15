package fe

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"
)

func CreateBootstrapLexer() ictiobus.Lexer {
	bootLx := ictiobus.NewLazyLexer()

	// default state, shared by all
	bootLx.RegisterClass(TCEscseq, "")
	bootLx.RegisterClass(TCHeaderTokens, "")
	bootLx.RegisterClass(TCHeaderGrammar, "")
	bootLx.RegisterClass(TCHeaderActions, "")
	//bootLx.RegisterClass(tcDirStart, "")

	// default patterns and defs
	bootLx.AddPattern(`%!.`, lex.LexAs(TCEscseq.ID()), "", 0)
	bootLx.AddPattern(`%%[Tt][Oo][Kk][Ee][Nn][Ss]`, lex.LexAndSwapState(TCHeaderTokens.ID(), "tokens"), "", 0)
	bootLx.AddPattern(`%%[Gg][Rr][Aa][Mm][Mm][Aa][Rr]`, lex.LexAndSwapState(TCHeaderGrammar.ID(), "grammar"), "", 0)
	bootLx.AddPattern(`%%[Aa][Cc][Tt][Ii][Oo][Nn][Ss]`, lex.LexAndSwapState(TCHeaderActions.ID(), "actions"), "", 0)
	//bootLx.AddPattern(`%[Ss][Tt][Aa][Rr][Tt]`, lex.LexAs(tcDirStart.ID()), "")

	// t-a-g, you're it, glub glub glubglubglub glub 38D

	// Sure would 8e gr8 if we had a st8 stack ::::/

	bootLx.RegisterClass(TCId, "state-t")
	bootLx.RegisterClass(TCId, "state-a")
	bootLx.RegisterClass(TCId, "state-g")
	bootLx.AddPattern(`\s+`, lex.Discard(), "state-t", 0)
	bootLx.AddPattern(`\s+`, lex.Discard(), "state-a", 0)
	bootLx.AddPattern(`\s+`, lex.Discard(), "state-g", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAndSwapState(TCId.ID(), "tokens"), "state-t", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAndSwapState(TCId.ID(), "actions"), "state-a", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAndSwapState(TCId.ID(), "grammar"), "state-g", 0)

	// tokens classes
	bootLx.RegisterClass(TCFreeformText, "tokens")
	bootLx.RegisterClass(TCDirShift, "tokens")
	bootLx.RegisterClass(TCDirHuman, "tokens")
	bootLx.RegisterClass(TCDirToken, "tokens")
	//bootLx.RegisterClass(tcDirDefault, "tokens")
	bootLx.RegisterClass(TCDirDiscard, "tokens")
	bootLx.RegisterClass(TCDirPriority, "tokens")
	bootLx.RegisterClass(TCDirState, "tokens")
	bootLx.RegisterClass(TCLineStartFreeformText, "tokens")
	bootLx.RegisterClass(TCLineStartEscseq, "tokens")

	// tokens patterns
	bootLx.AddPattern(`\n\s*%!.`, lex.LexAs(TCLineStartEscseq.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee]`, lex.LexAndSwapState(TCDirState.ID(), "state-t"), "tokens", 0)
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee][Ss][Hh][Ii][Ff][Tt]`, lex.LexAs(TCDirShift.ID()), "tokens", 1)
	bootLx.AddPattern(`%[Hh][Uu][Mm][Aa][Nn]`, lex.LexAs(TCDirHuman.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Tt][Oo][Kk][Ee][Nn]`, lex.LexAs(TCDirToken.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Dd][Ii][Ss][Cc][Aa][Rr][Dd]`, lex.LexAs(TCDirDiscard.ID()), "tokens", 0)
	bootLx.AddPattern(`%[Pp][Rr][Ii][Oo][Rr][Ii][Tt][Yy]`, lex.LexAs(TCDirPriority.ID()), "tokens", 0)
	bootLx.AddPattern(`[^\S\n]+`, lex.Discard(), "tokens", 0)
	bootLx.AddPattern(`\n\s*[^%\s]+[^%\n]*`, lex.LexAs(TCLineStartFreeformText.ID()), "tokens", 0)
	bootLx.AddPattern(`\n`, lex.Discard(), "tokens", 0)
	bootLx.AddPattern(`[^%\s]+[^%\n]*`, lex.LexAs(TCFreeformText.ID()), "tokens", 0)
	//bootLx.AddPattern(`%[Dd][Ee][Ff][Aa][Uu][Ll][Tt]`, lex.LexAs(tcDirDefault.ID()), "tokens")

	// grammar classes
	bootLx.RegisterClass(TCEq, "grammar")
	bootLx.RegisterClass(TCAlt, "grammar")
	bootLx.RegisterClass(TCNonterminal, "grammar")
	bootLx.RegisterClass(TCTerminal, "grammar")
	bootLx.RegisterClass(TCEpsilon, "grammar")
	bootLx.RegisterClass(TCDirState, "grammar")
	bootLx.RegisterClass(TCLineStartNonterminal, "grammar")

	// grammar patterns
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee]`, lex.LexAndSwapState(TCDirState.ID(), "state-g"), "grammar", 0)
	bootLx.AddPattern(`[^\S\n]+`, lex.Discard(), "grammar", 0)
	bootLx.AddPattern(`\n\s*{[A-Za-z][^}]*}`, lex.LexAs(TCLineStartNonterminal.ID()), "grammar", 1)
	bootLx.AddPattern(`\n`, lex.Discard(), "grammar", 0)
	bootLx.AddPattern(`\|`, lex.LexAs(TCAlt.ID()), "grammar", 0)
	bootLx.AddPattern(`{}`, lex.LexAs(TCEpsilon.ID()), "grammar", 0)
	bootLx.AddPattern(`{[A-Za-z][^}]*}`, lex.LexAs(TCNonterminal.ID()), "grammar", 0)
	bootLx.AddPattern(`[^=\s]\S*|\S\S+`, lex.LexAs(TCTerminal.ID()), "grammar", 0)
	bootLx.AddPattern(`=`, lex.LexAs(TCEq.ID()), "grammar", 0)

	// actions classes
	bootLx.RegisterClass(TCAttrRef, "actions")
	bootLx.RegisterClass(TCInt, "actions")
	bootLx.RegisterClass(TCNonterminal, "actions")
	bootLx.RegisterClass(TCDirSymbol, "actions")
	bootLx.RegisterClass(TCDirProd, "actions")
	bootLx.RegisterClass(TCDirWith, "actions")
	bootLx.RegisterClass(TCDirHook, "actions")
	bootLx.RegisterClass(TCDirSet, "actions")
	bootLx.RegisterClass(TCDirIndex, "actions")
	bootLx.RegisterClass(TCId, "actions")
	bootLx.RegisterClass(TCTerminal, "actions")
	bootLx.RegisterClass(TCEpsilon, "actions")
	bootLx.RegisterClass(TCDirState, "actions")

	// actions patterns
	bootLx.AddPattern(`\s+`, lex.Discard(), "actions", 0)
	bootLx.AddPattern(`(?:{(?:&|\.)(?:[0-9]+)?}|{[0-9]+}|{\^}|{[A-Za-z][^{}]*}|[^\s{}]+)\.[\$A-Za-z][\$A-Za-z0-9_]*`, lex.LexAs(TCAttrRef.ID()), "actions", 0)
	bootLx.AddPattern(`,`, lex.Discard(), "actions", 0)
	bootLx.AddPattern(`[0-9]+`, lex.LexAs(TCInt.ID()), "actions", 0)
	bootLx.AddPattern(`{[A-Za-z][^}]*}`, lex.LexAs(TCNonterminal.ID()), "actions", 0)
	bootLx.AddPattern(`%[Ss][Tt][Aa][Tt][Ee]`, lex.LexAndSwapState(TCDirState.ID(), "state-a"), "actions", 0)
	bootLx.AddPattern(`%[Ss][Yy][Mm][Bb][Oo][Ll]`, lex.LexAs(TCDirSymbol.ID()), "actions", 0)
	bootLx.AddPattern(`(?:->|%[Pp][Rr][Oo][Dd])`, lex.LexAs(TCDirProd.ID()), "actions", 0)
	bootLx.AddPattern(`\(\)`, lex.Discard(), "actions", 0)
	bootLx.AddPattern(`(?:\(|%[Ww][Ii][Tt][Hh])`, lex.LexAs(TCDirWith.ID()), "actions", 0)
	bootLx.AddPattern(`(?:=|%[Hh][Oo][Oo][Kk])`, lex.LexAs(TCDirHook.ID()), "actions", 0)
	bootLx.AddPattern(`\)`, lex.Discard(), "actions", 0)
	bootLx.AddPattern(`(?::|%[Ss][Ee][Tt])`, lex.LexAs(TCDirSet.ID()), "actions", 0)
	bootLx.AddPattern(`%[Ii][Nn][Dd][Ee][Xx]`, lex.LexAs(TCDirIndex.ID()), "actions", 0)
	bootLx.AddPattern(`[A-Za-z][A-Za-z0-9_-]*`, lex.LexAs(TCId.ID()), "actions", 0)
	bootLx.AddPattern(`{}`, lex.LexAs(TCEpsilon.ID()), "actions", 0)
	bootLx.AddPattern(`\S+`, lex.LexAs(TCTerminal.ID()), "actions", 0)

	return bootLx
}
