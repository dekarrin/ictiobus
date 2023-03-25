package fishi

import (
	"fmt"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/grammar"
)

func CreateBootstrapParser() (ictiobus.Parser, []string) {
	g := CreateBootstrapGrammar()
	if err := g.Validate(); err != nil {
		panic(fmt.Sprintf("bootstrap grammar failed: %s", err.Error()))
	}

	// now, can we make a parser from this?

	var parser ictiobus.Parser
	parser, ambigWarns, err := ictiobus.NewCLRParser(g, true)
	if err != nil {
		panic(fmt.Sprintf("bootstrap parser failed: %s", err.Error()))
	}
	return parser, ambigWarns
}

func CreateBootstrapGrammar() grammar.Grammar {
	bootCfg := grammar.Grammar{}

	bootCfg.AddTerm(tcHeaderTokens.ID(), tcHeaderTokens)
	bootCfg.AddTerm(tcHeaderGrammar.ID(), tcHeaderGrammar)
	bootCfg.AddTerm(tcHeaderActions.ID(), tcHeaderActions)
	bootCfg.AddTerm(tcDirAction.ID(), tcDirAction)
	//bootCfg.AddTerm(tcDirDefault.ID(), tcDirDefault)
	bootCfg.AddTerm(tcDirHook.ID(), tcDirHook)
	bootCfg.AddTerm(tcDirHuman.ID(), tcDirHuman)
	bootCfg.AddTerm(tcDirIndex.ID(), tcDirIndex)
	bootCfg.AddTerm(tcDirProd.ID(), tcDirProd)
	bootCfg.AddTerm(tcDirShift.ID(), tcDirShift)
	//bootCfg.AddTerm(tcDirStart.ID(), tcDirStart)
	bootCfg.AddTerm(tcDirState.ID(), tcDirState)
	bootCfg.AddTerm(tcDirSymbol.ID(), tcDirSymbol)
	bootCfg.AddTerm(tcDirToken.ID(), tcDirToken)
	bootCfg.AddTerm(tcDirWith.ID(), tcDirWith)
	bootCfg.AddTerm(tcFreeformText.ID(), tcFreeformText)
	//bootCfg.AddTerm(tcNewline.ID(), tcNewline)
	bootCfg.AddTerm(tcTerminal.ID(), tcTerminal)
	bootCfg.AddTerm(tcNonterminal.ID(), tcNonterminal)
	bootCfg.AddTerm(tcEq.ID(), tcEq)
	bootCfg.AddTerm(tcAlt.ID(), tcAlt)
	bootCfg.AddTerm(tcAttrRef.ID(), tcAttrRef)
	bootCfg.AddTerm(tcInt.ID(), tcInt)
	bootCfg.AddTerm(tcId.ID(), tcId)
	bootCfg.AddTerm(tcEscseq.ID(), tcEscseq)
	bootCfg.AddTerm(tcEpsilon.ID(), tcEpsilon)
	bootCfg.AddTerm(tcDirDiscard.ID(), tcDirDiscard)
	bootCfg.AddTerm(tcDirPriority.ID(), tcDirPriority)

	bootCfg.AddRule("FISHISPEC", []string{"BLOCKS"})

	bootCfg.AddRule("BLOCKS", []string{"BLOCKS", "BLOCK"})
	bootCfg.AddRule("BLOCKS", []string{"BLOCK"})

	bootCfg.AddRule("BLOCK", []string{"GRAMMAR-BLOCK"})
	bootCfg.AddRule("BLOCK", []string{"TOKENS-BLOCK"})
	bootCfg.AddRule("BLOCK", []string{"ACTIONS-BLOCK"})

	bootCfg.AddRule("ACTIONS-BLOCK", []string{tcHeaderActions.ID(), "ACTIONS-CONTENT"})

	bootCfg.AddRule("ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST", "ACTIONS-STATE-BLOCK-LIST"})
	bootCfg.AddRule("ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST"})
	bootCfg.AddRule("ACTIONS-CONTENT", []string{"ACTIONS-STATE-BLOCK-LIST"})

	bootCfg.AddRule("ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK-LIST", "ACTIONS-STATE-BLOCK"})
	bootCfg.AddRule("ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK"})

	bootCfg.AddRule("ACTIONS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "SYMBOL-ACTIONS-LIST"})

	bootCfg.AddRule("SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS-LIST", "SYMBOL-ACTIONS"})
	bootCfg.AddRule("SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS"})

	bootCfg.AddRule("SYMBOL-ACTIONS", []string{tcDirSymbol.ID(), tcNonterminal.ID(), "PROD-ACTIONS"})

	bootCfg.AddRule("PROD-ACTIONS", []string{"PROD-ACTIONS", "PROD-ACTION"})
	bootCfg.AddRule("PROD-ACTIONS", []string{"PROD-ACTION"})

	bootCfg.AddRule("PROD-ACTION", []string{"PROD-SPECIFIER", "SEMANTIC-ACTIONS"})

	bootCfg.AddRule("SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTIONS", "SEMANTIC-ACTION"})
	bootCfg.AddRule("SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTION"})

	bootCfg.AddRule("SEMANTIC-ACTION", []string{tcDirAction.ID(), tcAttrRef.ID(), tcDirHook.ID(), tcId.ID()})
	bootCfg.AddRule("SEMANTIC-ACTION", []string{tcDirAction.ID(), tcAttrRef.ID(), tcDirHook.ID(), tcId.ID(), "WITH-CLAUSE"})

	bootCfg.AddRule("WITH-CLAUSE", []string{tcDirWith.ID(), "ATTR-REFS"})

	bootCfg.AddRule("ATTR-REFS", []string{"ATTR-REFS", tcAttrRef.ID()})
	bootCfg.AddRule("ATTR-REFS", []string{tcAttrRef.ID()})

	bootCfg.AddRule("PROD-SPECIFIER", []string{tcDirProd.ID(), "PROD-ADDR"})
	bootCfg.AddRule("PROD-SPECIFIER", []string{tcDirProd.ID()})

	bootCfg.AddRule("PROD-ADDR", []string{tcDirIndex.ID(), tcInt.ID()})
	bootCfg.AddRule("PROD-ADDR", []string{"ACTION-PRODUCTION"})

	bootCfg.AddRule("ACTION-PRODUCTION", []string{"ACTION-SYMBOL-SEQUENCE"})
	bootCfg.AddRule("ACTION-PRODUCTION", []string{tcEpsilon.ID()})

	bootCfg.AddRule("ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL-SEQUENCE", "ACTION-SYMBOL"})
	bootCfg.AddRule("ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL"})

	bootCfg.AddRule("ACTION-SYMBOL", []string{tcNonterminal.ID()})
	bootCfg.AddRule("ACTION-SYMBOL", []string{tcTerminal.ID()})
	bootCfg.AddRule("ACTION-SYMBOL", []string{tcInt.ID()})
	bootCfg.AddRule("ACTION-SYMBOL", []string{tcId.ID()})

	// tokens
	bootCfg.AddRule("TOKENS-BLOCK", []string{tcHeaderTokens.ID(), "TOKENS-CONTENT"})

	bootCfg.AddRule("TOKENS-CONTENT", []string{"TOKENS-ENTRIES", "TOKENS-STATE-BLOCK-LIST"})
	bootCfg.AddRule("TOKENS-CONTENT", []string{"TOKENS-ENTRIES"})
	bootCfg.AddRule("TOKENS-CONTENT", []string{"TOKENS-STATE-BLOCK-LIST"})

	bootCfg.AddRule("TOKENS-STATE-BLOCK-LIST", []string{"TOKENS-STATE-BLOCK-LIST", "TOKENS-STATE-BLOCK"})
	bootCfg.AddRule("TOKENS-STATE-BLOCK-LIST", []string{"TOKENS-STATE-BLOCK"})

	bootCfg.AddRule("TOKENS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "TOKENS-ENTRIES"})

	bootCfg.AddRule("TOKENS-ENTRIES", []string{"TOKENS-ENTRIES", "TOKENS-ENTRY"})
	bootCfg.AddRule("TOKENS-ENTRIES", []string{"TOKENS-ENTRY"})

	bootCfg.AddRule("TOKENS-ENTRY", []string{"PATTERN", "TOKEN-OPTS"})

	bootCfg.AddRule("TOKEN-OPTS", []string{"TOKEN-OPTS", "TOKEN-OPTION"})
	bootCfg.AddRule("TOKEN-OPTS", []string{"TOKEN-OPTION"})

	bootCfg.AddRule("TOKEN-OPTION", []string{"DISCARD"})
	bootCfg.AddRule("TOKEN-OPTION", []string{"STATESHIFT"})
	bootCfg.AddRule("TOKEN-OPTION", []string{"TOKEN"})
	bootCfg.AddRule("TOKEN-OPTION", []string{"HUMAN"})
	bootCfg.AddRule("TOKEN-OPTION", []string{"PRIORITY"})

	bootCfg.AddRule("DISCARD", []string{tcDirDiscard.ID()})
	bootCfg.AddRule("STATESHIFT", []string{tcDirShift.ID(), "TEXT"})
	bootCfg.AddRule("TOKEN", []string{tcDirToken.ID(), "TEXT"})
	bootCfg.AddRule("HUMAN", []string{tcDirHuman.ID(), "TEXT"})
	bootCfg.AddRule("PRIORITY", []string{tcDirPriority.ID(), "TEXT"})

	bootCfg.AddRule("PATTERN", []string{"TEXT"})

	bootCfg.AddRule("GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "GRAMMAR-CONTENT"})

	bootCfg.AddRule("GRAMMAR-CONTENT", []string{"GRAMMAR-RULES", "GRAMMAR-STATE-BLOCK-LIST"})
	bootCfg.AddRule("GRAMMAR-CONTENT", []string{"GRAMMAR-RULES"})
	bootCfg.AddRule("GRAMMAR-CONTENT", []string{"GRAMMAR-STATE-BLOCK-LIST"})

	bootCfg.AddRule("GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK-LIST", "GRAMMAR-STATE-BLOCK"})
	bootCfg.AddRule("GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK"})

	bootCfg.AddRule("GRAMMAR-STATE-BLOCK", []string{"STATE-INSTRUCTION", "GRAMMAR-RULES"})

	bootCfg.AddRule("GRAMMAR-RULES", []string{"GRAMMAR-RULES", "GRAMMAR-RULE"})
	bootCfg.AddRule("GRAMMAR-RULES", []string{"GRAMMAR-RULE"})

	bootCfg.AddRule("GRAMMAR-RULE", []string{tcNonterminal.ID(), tcEq.ID(), "ALTERNATIONS"})

	bootCfg.AddRule("ALTERNATIONS", []string{"PRODUCTION"})
	bootCfg.AddRule("ALTERNATIONS", []string{"ALTERNATIONS", tcAlt.ID(), "PRODUCTION"})

	bootCfg.AddRule("PRODUCTION", []string{"SYMBOL-SEQUENCE"})
	bootCfg.AddRule("PRODUCTION", []string{tcEpsilon.ID()})

	bootCfg.AddRule("SYMBOL-SEQUENCE", []string{"SYMBOL-SEQUENCE", "SYMBOL"})
	bootCfg.AddRule("SYMBOL-SEQUENCE", []string{"SYMBOL"})

	bootCfg.AddRule("SYMBOL", []string{tcNonterminal.ID()})
	bootCfg.AddRule("SYMBOL", []string{tcTerminal.ID()})

	bootCfg.AddRule("STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"})

	bootCfg.AddRule("ID-EXPR", []string{tcId.ID()})
	bootCfg.AddRule("ID-EXPR", []string{tcTerminal.ID()})

	bootCfg.AddRule("TEXT", []string{"TEXT", "TEXT-ELEMENT"})
	bootCfg.AddRule("TEXT", []string{"TEXT-ELEMENT"})

	bootCfg.AddRule("TEXT-ELEMENT", []string{tcFreeformText.ID()})
	bootCfg.AddRule("TEXT-ELEMENT", []string{tcEscseq.ID()})

	bootCfg.Start = "FISHISPEC"
	bootCfg.RemoveUnusedTerminals()

	return bootCfg
}
