package fe

import (
	"fmt"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/grammar"

	. "github.com/dekarrin/ictiobus/fishi/fe/fetoken"
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

	bootCfg.AddTerm(TCHeaderTokens.ID(), TCHeaderTokens)
	bootCfg.AddTerm(TCHeaderGrammar.ID(), TCHeaderGrammar)
	bootCfg.AddTerm(TCHeaderActions.ID(), TCHeaderActions)
	bootCfg.AddTerm(TCDirSet.ID(), TCDirSet)
	//bootCfg.AddTerm(tcDirDefault.ID(), tcDirDefault)
	bootCfg.AddTerm(TCDirHook.ID(), TCDirHook)
	bootCfg.AddTerm(TCDirHuman.ID(), TCDirHuman)
	bootCfg.AddTerm(TCDirIndex.ID(), TCDirIndex)
	bootCfg.AddTerm(TCDirProd.ID(), TCDirProd)
	bootCfg.AddTerm(TCDirShift.ID(), TCDirShift)
	//bootCfg.AddTerm(tcDirStart.ID(), tcDirStart)
	bootCfg.AddTerm(TCDirState.ID(), TCDirState)
	bootCfg.AddTerm(TCDirSymbol.ID(), TCDirSymbol)
	bootCfg.AddTerm(TCDirToken.ID(), TCDirToken)
	bootCfg.AddTerm(TCDirWith.ID(), TCDirWith)
	bootCfg.AddTerm(TCFreeformText.ID(), TCFreeformText)
	//bootCfg.AddTerm(tcNewline.ID(), tcNewline)
	bootCfg.AddTerm(TCTerminal.ID(), TCTerminal)
	bootCfg.AddTerm(TCNonterminal.ID(), TCNonterminal)
	bootCfg.AddTerm(TCEq.ID(), TCEq)
	bootCfg.AddTerm(TCAlt.ID(), TCAlt)
	bootCfg.AddTerm(TCAttrRef.ID(), TCAttrRef)
	bootCfg.AddTerm(TCInt.ID(), TCInt)
	bootCfg.AddTerm(TCId.ID(), TCId)
	bootCfg.AddTerm(TCEscseq.ID(), TCEscseq)
	bootCfg.AddTerm(TCEpsilon.ID(), TCEpsilon)
	bootCfg.AddTerm(TCDirDiscard.ID(), TCDirDiscard)
	bootCfg.AddTerm(TCDirPriority.ID(), TCDirPriority)
	bootCfg.AddTerm(TCLineStartFreeformText.ID(), TCLineStartFreeformText)
	bootCfg.AddTerm(TCLineStartEscseq.ID(), TCLineStartEscseq)
	bootCfg.AddTerm(TCLineStartNonterminal.ID(), TCLineStartNonterminal)

	bootCfg.AddRule("FISHISPEC", []string{"BLOCKS"})

	bootCfg.AddRule("BLOCKS", []string{"BLOCKS", "BLOCK"})
	bootCfg.AddRule("BLOCKS", []string{"BLOCK"})

	bootCfg.AddRule("BLOCK", []string{"GRAMMAR-BLOCK"})
	bootCfg.AddRule("BLOCK", []string{"TOKENS-BLOCK"})
	bootCfg.AddRule("BLOCK", []string{"ACTIONS-BLOCK"})

	bootCfg.AddRule("ACTIONS-BLOCK", []string{TCHeaderActions.ID(), "ACTIONS-CONTENT"})

	bootCfg.AddRule("ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST", "ACTIONS-STATE-BLOCK-LIST"})
	bootCfg.AddRule("ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST"})
	bootCfg.AddRule("ACTIONS-CONTENT", []string{"ACTIONS-STATE-BLOCK-LIST"})

	bootCfg.AddRule("ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK-LIST", "ACTIONS-STATE-BLOCK"})
	bootCfg.AddRule("ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK"})

	bootCfg.AddRule("ACTIONS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "SYMBOL-ACTIONS-LIST"})

	bootCfg.AddRule("SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS-LIST", "SYMBOL-ACTIONS"})
	bootCfg.AddRule("SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS"})

	bootCfg.AddRule("SYMBOL-ACTIONS", []string{TCDirSymbol.ID(), TCNonterminal.ID(), "PROD-ACTIONS"})

	bootCfg.AddRule("PROD-ACTIONS", []string{"PROD-ACTIONS", "PROD-ACTION"})
	bootCfg.AddRule("PROD-ACTIONS", []string{"PROD-ACTION"})

	bootCfg.AddRule("PROD-ACTION", []string{"PROD-SPECIFIER", "SEMANTIC-ACTIONS"})

	bootCfg.AddRule("SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTIONS", "SEMANTIC-ACTION"})
	bootCfg.AddRule("SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTION"})

	bootCfg.AddRule("SEMANTIC-ACTION", []string{TCDirSet.ID(), TCAttrRef.ID(), TCDirHook.ID(), TCId.ID()})
	bootCfg.AddRule("SEMANTIC-ACTION", []string{TCDirSet.ID(), TCAttrRef.ID(), TCDirHook.ID(), TCId.ID(), "WITH-CLAUSE"})

	bootCfg.AddRule("WITH-CLAUSE", []string{TCDirWith.ID(), "ATTR-REFS"})

	bootCfg.AddRule("ATTR-REFS", []string{"ATTR-REFS", TCAttrRef.ID()})
	bootCfg.AddRule("ATTR-REFS", []string{TCAttrRef.ID()})

	bootCfg.AddRule("PROD-SPECIFIER", []string{TCDirProd.ID(), "PROD-ADDR"})
	bootCfg.AddRule("PROD-SPECIFIER", []string{TCDirProd.ID()})

	bootCfg.AddRule("PROD-ADDR", []string{TCDirIndex.ID(), TCInt.ID()})
	bootCfg.AddRule("PROD-ADDR", []string{"ACTION-PRODUCTION"})

	bootCfg.AddRule("ACTION-PRODUCTION", []string{"ACTION-SYMBOL-SEQUENCE"})
	bootCfg.AddRule("ACTION-PRODUCTION", []string{TCEpsilon.ID()})

	bootCfg.AddRule("ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL-SEQUENCE", "ACTION-SYMBOL"})
	bootCfg.AddRule("ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL"})

	bootCfg.AddRule("ACTION-SYMBOL", []string{TCNonterminal.ID()})
	bootCfg.AddRule("ACTION-SYMBOL", []string{TCTerminal.ID()})
	bootCfg.AddRule("ACTION-SYMBOL", []string{TCInt.ID()})
	bootCfg.AddRule("ACTION-SYMBOL", []string{TCId.ID()})

	// tokens
	bootCfg.AddRule("TOKENS-BLOCK", []string{TCHeaderTokens.ID(), "TOKENS-CONTENT"})

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

	bootCfg.AddRule("DISCARD", []string{TCDirDiscard.ID()})
	bootCfg.AddRule("STATESHIFT", []string{TCDirShift.ID(), "TEXT"})
	bootCfg.AddRule("TOKEN", []string{TCDirToken.ID(), "TEXT"})
	bootCfg.AddRule("HUMAN", []string{TCDirHuman.ID(), "TEXT"})
	bootCfg.AddRule("PRIORITY", []string{TCDirPriority.ID(), "TEXT"})

	bootCfg.AddRule("PATTERN", []string{"TEXT"})

	bootCfg.AddRule("GRAMMAR-BLOCK", []string{TCHeaderGrammar.ID(), "GRAMMAR-CONTENT"})

	bootCfg.AddRule("GRAMMAR-CONTENT", []string{"GRAMMAR-RULES", "GRAMMAR-STATE-BLOCK-LIST"})
	bootCfg.AddRule("GRAMMAR-CONTENT", []string{"GRAMMAR-RULES"})
	bootCfg.AddRule("GRAMMAR-CONTENT", []string{"GRAMMAR-STATE-BLOCK-LIST"})

	bootCfg.AddRule("GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK-LIST", "GRAMMAR-STATE-BLOCK"})
	bootCfg.AddRule("GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK"})

	bootCfg.AddRule("GRAMMAR-STATE-BLOCK", []string{"STATE-INSTRUCTION", "GRAMMAR-RULES"})

	bootCfg.AddRule("GRAMMAR-RULES", []string{"GRAMMAR-RULES", "GRAMMAR-RULE"})
	bootCfg.AddRule("GRAMMAR-RULES", []string{"GRAMMAR-RULE"})

	bootCfg.AddRule("GRAMMAR-RULE", []string{TCLineStartNonterminal.ID(), TCEq.ID(), "ALTERNATIONS"})

	bootCfg.AddRule("ALTERNATIONS", []string{"PRODUCTION"})
	bootCfg.AddRule("ALTERNATIONS", []string{"ALTERNATIONS", TCAlt.ID(), "PRODUCTION"})

	bootCfg.AddRule("PRODUCTION", []string{"SYMBOL-SEQUENCE"})
	bootCfg.AddRule("PRODUCTION", []string{TCEpsilon.ID()})

	bootCfg.AddRule("SYMBOL-SEQUENCE", []string{"SYMBOL-SEQUENCE", "SYMBOL"})
	bootCfg.AddRule("SYMBOL-SEQUENCE", []string{"SYMBOL"})

	bootCfg.AddRule("SYMBOL", []string{TCNonterminal.ID()})
	bootCfg.AddRule("SYMBOL", []string{TCTerminal.ID()})

	bootCfg.AddRule("STATE-INSTRUCTION", []string{TCDirState.ID(), "ID-EXPR"})

	bootCfg.AddRule("ID-EXPR", []string{TCId.ID()})
	bootCfg.AddRule("ID-EXPR", []string{TCTerminal.ID()})

	// Needed SDTS updates:
	// - update TEXT it's completely changed *done*
	// - add LINE-START-TEXT-ELEMENT *done*
	// - add TEXT-ELEMENTS
	bootCfg.AddRule("TEXT", []string{"LINE-START-TEXT-ELEMENT", "TEXT-ELEMENTS"})
	bootCfg.AddRule("TEXT", []string{"TEXT-ELEMENTS"})
	bootCfg.AddRule("TEXT", []string{"LINE-START-TEXT-ELEMENT"})

	bootCfg.AddRule("TEXT-ELEMENTS", []string{"TEXT-ELEMENTS", "TEXT-ELEMENT"})
	bootCfg.AddRule("TEXT-ELEMENTS", []string{"TEXT-ELEMENT"})

	bootCfg.AddRule("LINE-START-TEXT-ELEMENT", []string{TCLineStartEscseq.ID()})
	bootCfg.AddRule("LINE-START-TEXT-ELEMENT", []string{TCLineStartFreeformText.ID()})

	bootCfg.AddRule("TEXT-ELEMENT", []string{TCFreeformText.ID()})
	bootCfg.AddRule("TEXT-ELEMENT", []string{TCEscseq.ID()})

	bootCfg.Start = "FISHISPEC"
	bootCfg.RemoveUnusedTerminals()

	return bootCfg
}
