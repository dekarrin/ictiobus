package fishi

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/translation"
)

func CreateBootstrapSDTS() ictiobus.SDTS {
	sdts := ictiobus.NewSDTS()

	bootstrapSDTSFishispecAST(sdts)
	bootstrapSDTSBlocksValue(sdts)
	bootstrapSDTSBlockAST(sdts)
	bootstrapSDTSGrammarBlockAST(sdts)
	bootstrapSDTSGrammarContentAST(sdts)
	bootstrapSDTSGrammarStateBlockListValue(sdts)
	bootstrapSDTSGrammarStateBlockValue(sdts)
	bootstrapSDTSGrammarRulesValue(sdts)
	bootstrapSDTSGrammarRuleValue(sdts)
	bootstrapSDTSStateInstructionState(sdts)
	bootstrapSDTSIDExprValue(sdts)
	bootstrapSDTSTextValue(sdts)
	bootstrapSDTSTextElementsValue(sdts)
	bootstrapSDTSTextElementValue(sdts)
	bootstrapSDTSLineStartTextElementValue(sdts)
	bootstrapSDTSAlternationsValue(sdts)
	bootstrapSDTSProductionValue(sdts)
	bootstrapSDTSSymbolSequenceValue(sdts)
	bootstrapSDTSSymbolValue(sdts)
	bootstrapSDTSTokensBlockAST(sdts)
	bootstrapSDTSTokensContentAST(sdts)
	bootstrapSDTSTokensStateBlockListValue(sdts)
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
	bootstrapSDTSActionsStateBlockListValue(sdts)
	bootstrapSDTSActionsStateBlockValue(sdts)
	bootstrapSDTSSymbolActionsListValue(sdts)
	bootstrapSDTSSymbolActionsValue(sdts)
	bootstrapSDTSProdActionsValue(sdts)
	bootstrapSDTSProdActionValue(sdts)
	bootstrapSDTSProdSpecifierValue(sdts)
	bootstrapSDTSProdAddrValue(sdts)
	bootstrapSDTSActionProductionValue(sdts)
	bootstrapSDTSActionSymbolSequenceValue(sdts)
	bootstrapSDTSActionSymbolValue(sdts)
	bootstrapSDTSSemanticActionsValue(sdts)
	bootstrapSDTSSemanticActionValue(sdts)
	bootstrapSDTSWithClauseValue(sdts)
	bootstrapSDTSAttrRefsValue(sdts)

	return sdts
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
		"TOKENS-BLOCK", []string{tcHeaderTokens.ID(), "TOKENS-CONTENT"},
		"ast",
		sdtsFnMakeTokensBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
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
		"GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "GRAMMAR-CONTENT"},
		"ast",
		sdtsFnMakeGrammarBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

func bootstrapSDTSTokensContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-ENTRIES"},
		"ast",
		sdtsFnTokensContentBlocksStartEntryList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-STATE-BLOCK-LIST"},
		"ast",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-ENTRIES", "TOKENS-STATE-BLOCK-LIST"},
		"ast",
		sdtsFnTokensContentBlocksAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionsContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST"},
		"ast",
		sdtsFnActionsContentBlocksStartSymbolActionsList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"ACTIONS-STATE-BLOCK-LIST"},
		"ast",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST", "ACTIONS-STATE-BLOCK-LIST"},
		"ast",
		sdtsFnActionsContentBlocksAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-RULES"},
		"ast",
		sdtsFnGrammarContentBlocksStartRuleList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-STATE-BLOCK-LIST"},
		"ast",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "list"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-RULES", "GRAMMAR-STATE-BLOCK-LIST"},
		"ast",
		sdtsFnGrammarContentBlocksAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "list"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK", []string{"STATE-INSTRUCTION", "GRAMMAR-RULES"},
		"value",
		sdtsFnMakeGrammarContentNode,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "state"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
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

func bootstrapSDTSAttrRefsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ATTR-REFS", []string{"ATTR-REFS", tcAttrRef.ID()},
		"value",
		sdtsFnAttrRefListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"ATTR-REFS", []string{tcAttrRef.ID()},
		"value",
		sdtsFnAttrRefListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}

func bootstrapSDTSWithClauseValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"WITH-CLAUSE", []string{tcDirWith.ID(), "ATTR-REFS"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSSemanticActionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTION", []string{tcDirAction.ID(), tcAttrRef.ID(), tcDirHook.ID(), tcId.ID()},
		"value",
		sdtsFnMakeSemanticAction,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 3}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTION", []string{tcDirAction.ID(), tcAttrRef.ID(), tcDirHook.ID(), tcId.ID(), "WITH-CLAUSE"},
		"value",
		sdtsFnMakeSemanticAction,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 3}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 4}, Name: "value"},
		},
	)
}

func bootstrapSDTSSemanticActionsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTIONS", "SEMANTIC-ACTION"},
		"value",
		sdtsFnSemanticActionListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTION"},
		"value",
		sdtsFnSemanticActionListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionSymbolValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{tcNonterminal.ID()},
		"value",
		sdtsFnGetNonterminal,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{tcTerminal.ID()},
		"value",
		sdtsFnGetTerminal,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{tcInt.ID()},
		"value",
		sdtsFnGetInt,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{tcId.ID()},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}

func bootstrapSDTSActionSymbolSequenceValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL-SEQUENCE", "ACTION-SYMBOL"},
		"value",
		sdtsFnStringListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL"},
		"value",
		sdtsFnStringListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionProductionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTION-PRODUCTION", []string{"ACTION-SYMBOL-SEQUENCE"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-PRODUCTION", []string{tcEpsilon.ID()},
		"value",
		sdtsFnEpsilonStringList,
		nil,
	)
}

func bootstrapSDTSProdAddrValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-ADDR", []string{tcDirIndex.ID(), tcInt.ID()},
		"value",
		sdtsFnMakeProdSpecifierIndex,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PROD-ADDR", []string{"ACTION-PRODUCTION"},
		"value",
		sdtsFnMakeProdSpecifierLiteral,
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

func bootstrapSDTSActionsStateBlockListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK-LIST", "ACTIONS-STATE-BLOCK"},
		"value",
		sdtsFnActionsStateBlockListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK"},
		"value",
		sdtsFnActionsStateBlockListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensStateBlockListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-STATE-BLOCK-LIST", []string{"TOKENS-STATE-BLOCK-LIST", "TOKENS-STATE-BLOCK"},
		"value",
		sdtsFnTokensStateBlockListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-STATE-BLOCK-LIST", []string{"TOKENS-STATE-BLOCK"},
		"value",
		sdtsFnTokensStateBlockListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarStateBlockListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK-LIST", "GRAMMAR-STATE-BLOCK"},
		"list",
		sdtsFnGrammarStateBlockListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "list"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK"},
		"list",
		sdtsFnGrammarStateBlockListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "TOKENS-ENTRIES"},
		"value",
		sdtsFnMakeTokensContentNode,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "state"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarRulesValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULES", "GRAMMAR-RULE"},
		"value",
		sdtsFnRuleListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
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
		"TOKENS-ENTRIES", []string{"TOKENS-ENTRIES", "TOKENS-ENTRY"},
		"value",
		sdtsFnEntryListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
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
		"TOKENS-ENTRY", []string{"PATTERN", "TOKEN-OPTS"},
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
		nil,
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
		"STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"},
		"state",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
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
		"TEXT", []string{"LINE-START-TEXT-ELEMENT", "TEXT-ELEMENTS"},
		"value",
		sdtsFnAppendStringsTrimmed,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TEXT", []string{"TEXT-ELEMENTS"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TEXT", []string{"LINE-START-TEXT-ELEMENT"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}
func bootstrapSDTSTextElementsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENTS", []string{"TEXT-ELEMENTS", "TEXT-ELEMENT"},
		"value",
		sdtsFnAppendStrings,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENTS", []string{"TEXT-ELEMENT"},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSLineStartTextElementValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"LINE-START-TEXT-ELEMENT", []string{tcLineStartFreeformText.ID()},
		"value",
		sdtsFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"LINE-START-TEXT-ELEMENT", []string{tcLineStartEscseq.ID()},
		"value",
		sdtsFnInterpretEscape,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
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
