package fishi

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/translation"
)

func CreateBootstrapSDTS() ictiobus.SDTS {
	sdts := ictiobus.NewSDTS()

	bootstrapSDTSFishispecAST(sdts)
	bootstrapSDTSBlocksValue(sdts)
	bootstrapSDTSBlockAST(sdts)
	bootstrapSDTSGrammarBlockAST(sdts)
	bootstrapSDTSGrammarContentAST(sdts)
	bootstrapSDTSGrammarStateBlockValue(sdts)
	bootstrapSDTSGrammarRulesValue(sdts)
	bootstrapSDTSGrammarRuleValue(sdts)
	bootstrapSDTSStateInstructionState(sdts)
	bootstrapSDTSIDExprValue(sdts)
	bootstrapSDTSTextValue(sdts)
	bootstrapSDTSTextElementValue(sdts)
	bootstrapSDTSAlternationsValue(sdts)
	bootstrapSDTSProductionValue(sdts)
	bootstrapSDTSSymbolSequenceValue(sdts)
	bootstrapSDTSSymbolValue(sdts)
	bootstrapSDTSTokensBlockAST(sdts)
	bootstrapSDTSTokensContentAST(sdts)
	bootstrapSDTSTokensStateBlockValue(sdts)
	bootstrapSDTSTokensEntriesValue(sdts)
	bootstrapSDTSTokensEntryValue(sdts)
	bootstrapSDTSPattern(sdts)
	bootstrapSDTSTokenOptsValue(sdts)
	bootstrapSDTSTokenOptionValue(sdts)
	bootstrapSDTSStateshiftValue(sdts)
	bootstrapSDTSTokenValue(sdts)
	bootstrapSDTSHumanValue(sdts)
	bootstrapSDTSPriorityValue(sdts)

	bootstrapSDTSActionsBlockAST(sdts)
	bootstrapSDTSActionsContentAST(sdts)
	bootstrapSDTSActionsStateBlockValue(sdts)
	bootstrapSDTSSymbolActionsListValue(sdts)
	bootstrapSDTSSymbolActionsValue(sdts)
	bootstrapSDTSProdActionsValue(sdts)
	bootstrapSDTSProdActionValue(sdts)
	bootstrapSDTSProdSpecifierValue(sdts)

	bootstrapSDTSFakeSynth(sdts, "PROD-ADDR", []string{tcDirIndex.ID(), tcInt.ID()}, "value", box.Pair[string, interface{}]{"INDEX", 2})
	bootstrapSDTSFakeSynth(sdts, "PROD-ADDR", []string{"ACTION-PRODUCTION"}, "value", box.Pair[string, interface{}]{"LITERAL", []string{"1", "+", "seven"}})

	bootstrapSDTSFakeSynth(sdts, "SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTIONS", "SEMANTIC-ACTION"}, "value", []semanticAction{{hook: "FAKE"}})
	bootstrapSDTSFakeSynth(sdts, "SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTION"}, "value", []semanticAction{{hook: "FAKE-2.0"}})

	sdts.SetNoFlow(true, "SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTIONS", "SEMANTIC-ACTION"}, "value", translation.NodeRelation{}, -1, "SEMANTIC-ACTIONS")
	sdts.SetNoFlow(true, "SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTION"}, "value", translation.NodeRelation{}, -1, "SEMANTIC-ACTIONS")

	// NEXT STEPS:
	//
	// PROD-SPECIFIER:
	// - Mock both PROD-ADDR rules DONE
	// - create function bootstrapSDTSProdSpecifierValue DONE
	// - remove PROD-SPECIFIER mock DONE
	//
	// PROD-ADDR:
	// - Mock both ACTION-PRODUCTION rules
	// - create function bootstrapSDTSProdAddrValue
	// - remove PROD-ADDR mock
	//
	// ACTION-PRODUCTION:
	// - Mock both ACTION-SYMBOL-SEQUENCE rules
	// - create function bootstrapSDTSActionProductionValue
	// - remove ACTION-PRODUCTION mock
	//
	// ACTION-SYMBOL-SEQUENCE:
	// - Mock all four ACTION-SYMBOL rules
	// - create function bootstrapSDTSActionSymbolSequenceValue
	// - remove ACTION-SYMBOL-SEQUENCE mock
	//
	// ACTION-SYMBOL:
	// - create function bootstrapSDTSActionSymbolValue
	// - remove ACTION-SYMBOL mock
	//
	// SEMANTIC-ACTIONS:
	// - Mock both SEMANTIC-ACTION rules
	// - create function bootstrapSDTSSemanticActionsValue
	// - remove SEMANTIC-ACTIONS mock
	// - remove NoFlows for SEMANTIC-ACTIONS
	//
	// SEMANTIC-ACTION:
	// - Mock WITH-CLAUSE rule
	// - create function bootstrapSDTSSemanticActionValue
	// - remove SEMANTIC-ACTION mock
	//
	// WITH-CLAUSE:
	// - Mock both ATTR-REFS rules
	// - create function bootstrapSDTSWithClauseValue
	// - remove WITH-CLAUSE mock
	//
	// ATTR-REFS:
	// - create function bootstrapSDTSAttrRefsValue
	// - remove ATTR-REFS mock
	//

	return sdts
}

func bootstrapSDTSFakeSynth(sdts ictiobus.SDTS, head string, prod []string, name string, value interface{}) {
	sdts.BindSynthesizedAttribute(
		head, prod,
		name,
		func(_, _ string, args []interface{}) interface{} { return value },
		nil,
	)
}

func bootstrapSDTSFishispecAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"FISHISPEC", []string{"BLOCKS"},
		"ast",
		sdtsFnMakeFishispec,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSBlocksValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"BLOCKS", []string{"BLOCKS", "BLOCK"},
		"value",
		sdtsFnBlockListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"BLOCKS", []string{"BLOCK"},
		"value",
		sdtsFnBlockListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
}

func bootstrapSDTSBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"BLOCK", []string{"GRAMMAR-BLOCK"},
		"ast",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"BLOCK", []string{"TOKENS-BLOCK"},
		"ast",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"BLOCK", []string{"ACTIONS-BLOCK"},
		"ast",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
}

func bootstrapSDTSActionsBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-BLOCK", []string{tcHeaderActions.ID(), "ACTIONS-CONTENT"},
		"ast",
		sdtsFnMakeActionsBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

func bootstrapSDTSTokensBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-BLOCK", []string{tcHeaderTokens.ID(), "TOKENS-CONTENT"},
		"ast",
		sdtsFnMakeTokensBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-BLOCK", []string{tcHeaderTokens.ID(), "NEWLINES", "TOKENS-CONTENT"},
		"ast",
		sdtsFnMakeTokensBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "ast"},
		},
	)
}

func bootstrapSDTSGrammarBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "GRAMMAR-CONTENT"},
		"ast",
		sdtsFnMakeGrammarBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "NEWLINES", "GRAMMAR-CONTENT"},
		"ast",
		sdtsFnMakeGrammarBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "ast"},
		},
	)
}

func bootstrapSDTSTokensContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-CONTENT", "TOKENS-STATE-BLOCK"},
		"ast",
		sdtsFnTokensContentBlocksAppendStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-CONTENT", "TOKENS-ENTRIES"},
		"ast",
		sdtsFnTokensContentBlocksAppendEntryList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-STATE-BLOCK"},
		"ast",
		sdtsFnTokensContentBlocksStartStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-ENTRIES"},
		"ast",
		sdtsFnTokensContentBlocksStartEntryList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionsContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"ACTIONS-CONTENT", "ACTIONS-STATE-BLOCK"},
		"ast",
		sdtsFnActionsContentBlocksAppendStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"ACTIONS-CONTENT", "SYMBOL-ACTIONS-LIST"},
		"ast",
		sdtsFnActionsContentBlocksAppendSymbolActionsList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"ACTIONS-STATE-BLOCK"},
		"ast",
		sdtsFnActionsContentBlocksStartStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST"},
		"ast",
		sdtsFnActionsContentBlocksStartSymbolActionsList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-CONTENT", "GRAMMAR-STATE-BLOCK"},
		"ast",
		sdtsFnGrammarContentBlocksAppendStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-CONTENT", "GRAMMAR-RULES"},
		"ast",
		sdtsFnGrammarContentBlocksAppendRuleList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-STATE-BLOCK"},
		"ast",
		sdtsFnGrammarContentBlocksStartStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-RULES"},
		"ast",
		sdtsFnGrammarContentBlocksStartRuleList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK", []string{"STATE-INSTRUCTION", "NEWLINES", "GRAMMAR-RULES"},
		"value",
		sdtsFnMakeGrammarContentNode,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "state"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionsStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "SYMBOL-ACTIONS-LIST"},
		"value",
		sdtsFnMakeActionsContentNode,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "state"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSProdActionsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-ACTIONS", []string{"PROD-ACTIONS", "PROD-ACTION"},
		"value",
		sdtsFnProdActionListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PROD-ACTIONS", []string{"PROD-ACTION"},
		"value",
		sdtsFnProdActionListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSProdSpecifierValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-SPECIFIER", []string{tcDirProd.ID(), "PROD-ADDR"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PROD-SPECIFIER", []string{tcDirProd.ID()},
		"value",
		sdtsFnMakeProdSpecifierNext,
		nil,
	)
}

func bootstrapSDTSProdActionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-ACTION", []string{"PROD-SPECIFIER", "SEMANTIC-ACTIONS"},
		"value",
		sdtsFnMakeProdAction,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSSymbolActionsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL-ACTIONS", []string{tcDirSymbol.ID(), tcNonterminal.ID(), "PROD-ACTIONS"},
		"value",
		sdtsFnMakeSymbolActions,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDTSSymbolActionsListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS-LIST", "SYMBOL-ACTIONS"},
		"value",
		sdtsFnSymbolActionsListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS"},
		"value",
		sdtsFnSymbolActionsListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "NEWLINES", "TOKENS-ENTRIES"},
		"value",
		sdtsFnMakeTokensContentNode,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "state"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarRulesValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULES", "NEWLINES", "GRAMMAR-RULE"},
		"value",
		sdtsFnRuleListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULE"},
		"value",
		sdtsFnRuleListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensEntriesValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRIES", []string{"TOKENS-ENTRIES", "NEWLINES", "TOKENS-ENTRY"},
		"value",
		sdtsFnEntryListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRIES", []string{"TOKENS-ENTRY"},
		"value",
		sdtsFnEntryListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensEntryValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRY", []string{"PATTERN", "NEWLINES", "TOKEN-OPTS"},
		"value",
		sdtsFnMakeTokenEntry,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRY", []string{"PATTERN", "NEWLINES", "TOKEN-OPTS", "NEWLINES"},
		"value",
		sdtsFnMakeTokenEntry,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRY", []string{"PATTERN", "TOKEN-OPTS"},
		"value",
		sdtsFnMakeTokenEntry,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRY", []string{"PATTERN", "TOKEN-OPTS", "NEWLINES"},
		"value",
		sdtsFnMakeTokenEntry,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarRuleValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULE", []string{tcNonterminal.ID(), tcEq.ID(), "ALTERNATIONS"},
		"value",
		sdtsFnMakeRule,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULE", []string{tcNonterminal.ID(), tcEq.ID(), "ALTERNATIONS", "NEWLINES"},
		"value",
		sdtsFnMakeRule,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDTSAlternationsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"PRODUCTION"},
		"value",
		sdtsFnStringListListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"ALTERNATIONS", tcAlt.ID(), "PRODUCTION"},
		"value",
		sdtsFnStringListListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"ALTERNATIONS", "NEWLINES", tcAlt.ID(), "PRODUCTION"},
		"value",
		sdtsFnStringListListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 3}, Name: "value"},
		},
	)
}

func bootstrapSDTSProductionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PRODUCTION", []string{"SYMBOL-SEQUENCE"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PRODUCTION", []string{tcEpsilon.ID()},
		"value",
		sdtsFnEpsilonStringList,
		[]translation.AttrRef{},
	)
}

func bootstrapSDTSSymbolSequenceValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL-SEQUENCE", []string{"SYMBOL-SEQUENCE", "SYMBOL"},
		"value",
		sdtsFnStringListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"SYMBOL-SEQUENCE", []string{"SYMBOL"},
		"value",
		sdtsFnStringListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSPriorityValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PRIORITY", []string{tcDirPriority.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSHumanValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"HUMAN", []string{tcDirHuman.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokenValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKEN", []string{tcDirToken.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSStateshiftValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"STATESHIFT", []string{tcDirShift.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokenOptionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"DISCARD"},
		"value",
		sdtsFnMakeDiscardOption,
		nil,
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"STATESHIFT"},
		"value",
		sdtsFnMakeStateshiftOption,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"TOKEN"},
		"value",
		sdtsFnMakeTokenOption,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"HUMAN"},
		"value",
		sdtsFnMakeHumanOption,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"PRIORITY"},
		"value",
		sdtsFnMakePriorityOption,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokenOptsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTS", []string{"TOKEN-OPTS", "NEWLINES", "TOKEN-OPTION"},
		"value",
		sdtsFnTokenOptListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTS", []string{"TOKEN-OPTS", "TOKEN-OPTION"},
		"value",
		sdtsFnTokenOptListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTS", []string{"TOKEN-OPTION"},
		"value",
		sdtsFnTokenOptListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSPattern(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PATTERN", []string{"TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSSymbolValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL", []string{tcNonterminal.ID()},
		"value",
		sdtsFnGetNonterminal,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"SYMBOL", []string{tcTerminal.ID()},
		"value",
		sdtsFnGetTerminal,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}

func bootstrapSDTSStateInstructionState(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"},
		"state",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"},
		"state",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSIDExprValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ID-EXPR", []string{tcId.ID()},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"ID-EXPR", []string{tcTerminal.ID()},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"ID-EXPR", []string{"TEXT"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTextValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TEXT", []string{"TEXT-ELEMENT"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"TEXT", []string{"TEXT", "TEXT-ELEMENT"},
		"value",
		sdtsFnAppendStrings,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSTextElementValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENT", []string{tcFreeformText.ID()},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENT", []string{tcEscseq.ID()},
		"value",
		sdtsFnInterpretEscape,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}
