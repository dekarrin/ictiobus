package fishi

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/translation"
)

func CreateBootstrapSDTS() ictiobus.SDTS {
	sdts := ictiobus.NewSDTS()

	bootstrapSDDFishispecAST(sdts)
	bootstrapSDDBlocksValue(sdts)
	bootstrapSDDBlockAST(sdts)
	bootstrapSDDGrammarBlockAST(sdts)
	bootstrapSDDGrammarContentAST(sdts)
	bootstrapSDDGrammarStateBlockValue(sdts)
	bootstrapSDDGrammarRulesValue(sdts)
	bootstrapSDDGrammarRuleValue(sdts)
	bootstrapSDDStateInstructionState(sdts)
	bootstrapSDDIDExprValue(sdts)
	bootstrapSDDTextValue(sdts)
	bootstrapSDDTextElementValue(sdts)
	bootstrapSDDAlternationsValue(sdts)
	bootstrapSDDProductionValue(sdts)
	bootstrapSDDSymbolSequenceValue(sdts)
	bootstrapSDDSymbolValue(sdts)
	bootstrapSDDTokensBlockAST(sdts)
	bootstrapSDDTokensContentAST(sdts)
	bootstrapSDDTokensStateBlockValue(sdts)
	bootstrapSDDTokensEntriesValue(sdts)
	bootstrapSDDTokensEntryValue(sdts)
	bootstrapSDDPattern(sdts)
	bootstrapSDDTokenOptsValue(sdts)
	bootstrapSDDTokenOptionValue(sdts)
	bootstrapSDDStateshiftValue(sdts)
	bootstrapSDDTokenValue(sdts)
	bootstrapSDDHumanValue(sdts)
	bootstrapSDDPriorityValue(sdts)

	bootstrapSDDActionsBlockAST(sdts)
	bootstrapSDDActionsContentAST(sdts)

	bootstrapSDDFakeSynth(sdts, "ACTIONS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "SYMBOL-ACTIONS-LIST"}, "value", astActionsContent{state: "fakeFromSTATEBLOCK"})

	bootstrapSDDFakeSynth(sdts, "SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS-LIST", "SYMBOL-ACTIONS"}, "value", []symbolActions{{symbol: "symACTfake"}})
	bootstrapSDDFakeSynth(sdts, "SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS"}, "value", []symbolActions{{symbol: "symACTfake2"}})

	// NEXT STEPS:
	//
	// ACTIONS-CONTENT:
	// - AST struct for it (DONE)
	// - Mock ACTIONS-STATE-BLOCK (DONE)
	// - Mock both SYMBOL-ACTIONS-LIST rules (DONE)
	// - create function bootstrapSDDActionsContentAST (DONE)
	// - update AST string() to print out the actions AST content block (DONE)
	// - remove ACTIONS-CONTENT mock (DONE)
	// - need several NoFlows from Symbol-ACtions-List... or proceed
	//
	// ACTIONS-STATE-BLOCK:
	// - (SYMBOL-ACTIONS-LIST should already be mocked)
	// - create function bootstrapSDDActionsStateBlockAST
	// - update AST string() to print out the actions state block
	// - remove NoFlow STATE-INSTRUCTION -> ACTIONS-STATE-BLOCK
	// - remove ACTIONS-STATE-BLOCK mock
	//
	// SYMBOL-ACTIONS-LIST:
	// - Mock SYMBOL-ACTIONS rule
	// - create function bootstrapSDDSymbolActionsListValue
	// - remove SYMBOL-ACTIONS-LIST mock
	//
	// SYMBOL-ACTIONS:
	// - Mock both PROD-ACTIONS rules
	// - create function bootstrapSDDSymbolActionsValue
	// - remove SYMBOL-ACTIONS mock
	//
	// PROD-ACTIONS:
	// - Mock PROD-ACTION rule
	// - create function bootstrapSDDProdActionsValue
	// - remove PROD-ACTIONS mock
	//
	// PROD-ACTION:
	// - Mock both PROD-SPECIFIER rules
	// - Mock both SEMANTIC-ACTIONS rules
	// - create function bootstrapSDDProdActionValue
	// - remove PROD-ACTION mock
	//
	// PROD-SPECIFIER:
	// - Mock bot PROD-ADDR rules
	// - create function bootstrapSDDProdSpecifierValue
	// - remove PROD-SPECIFIER mock
	//
	// PROD-ADDR:
	// - Mock both ACTION-PRODUCTION rules
	// - create function bootstrapSDDProdAddrValue
	// - remove PROD-ADDR mock
	//
	// ACTION-PRODUCTION:
	// - Mock both ACTION-SYMBOL-SEQUENCE rules
	// - create function bootstrapSDDActionProductionValue
	// - remove ACTION-PRODUCTION mock
	//
	// ACTION-SYMBOL-SEQUENCE:
	// - Mock all four ACTION-SYMBOL rules
	// - create function bootstrapSDDActionSymbolSequenceValue
	// - remove ACTION-SYMBOL-SEQUENCE mock
	//
	// ACTION-SYMBOL:
	// - create function bootstrapSDDActionSymbolValue
	// - remove ACTION-SYMBOL mock
	//
	// SEMANTIC-ACTIONS:
	// - Mock both SEMANTIC-ACTION rules
	// - create function bootstrapSDDSemanticActionsValue
	// - remove SEMANTIC-ACTIONS mock
	//
	// SEMANTIC-ACTION:
	// - Mock WITH-CLAUSE rule
	// - create function bootstrapSDDSemanticActionValue
	// - remove SEMANTIC-ACTION mock
	//
	// WITH-CLAUSE:
	// - Mock both ATTR-REFS rules
	// - create function bootstrapSDDWithClauseValue
	// - remove WITH-CLAUSE mock
	//
	// ATTR-REFS:
	// - create function bootstrapSDDAttrRefsValue
	// - remove ATTR-REFS mock
	//

	sdts.SetNoFlow(true, "ACTIONS-CONTENT", []string{"ACTIONS-CONTENT", "ACTIONS-STATE-BLOCK"}, "ast", translation.NodeRelation{}, -1, "ACTIONS-CONTENT")
	sdts.SetNoFlow(true, "ACTIONS-CONTENT", []string{"ACTIONS-CONTENT", "SYMBOL-ACTIONS-LIST"}, "ast", translation.NodeRelation{}, -1, "ACTIONS-CONTENT")
	sdts.SetNoFlow(true, "ACTIONS-CONTENT", []string{"ACTIONS-STATE-BLOCK"}, "ast", translation.NodeRelation{}, -1, "ACTIONS-CONTENT")
	sdts.SetNoFlow(true, "ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST"}, "ast", translation.NodeRelation{}, -1, "ACTIONS-CONTENT")

	sdts.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "ACTIONS-STATE-BLOCK")
	sdts.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "ACTIONS-STATE-BLOCK")

	return sdts
}

func bootstrapSDDFakeSynth(sdts ictiobus.SDTS, head string, prod []string, name string, value interface{}) {
	sdts.BindSynthesizedAttribute(
		head, prod,
		name,
		func(_, _ string, args []interface{}) interface{} { return value },
		nil,
	)
}

func bootstrapSDDFishispecAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"FISHISPEC", []string{"BLOCKS"},
		"ast",
		sdtsFnMakeFishispec,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDBlocksValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDBlockAST(sdts ictiobus.SDTS) {
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

func bootstrapSDDActionsBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-BLOCK", []string{tcHeaderActions.ID(), "ACTIONS-CONTENT"},
		"ast",
		sdtsFnMakeActionsBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

func bootstrapSDDTokensBlockAST(sdts ictiobus.SDTS) {
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

func bootstrapSDDGrammarBlockAST(sdts ictiobus.SDTS) {
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

func bootstrapSDDTokensContentAST(sdts ictiobus.SDTS) {
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

func bootstrapSDDActionsContentAST(sdts ictiobus.SDTS) {
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
		sdtsFnTokensContentBlocksStartStateBlock,
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

func bootstrapSDDGrammarContentAST(sdts ictiobus.SDTS) {
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

func bootstrapSDDGrammarStateBlockValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDTokensStateBlockValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDGrammarRulesValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDTokensEntriesValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDTokensEntryValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDGrammarRuleValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDAlternationsValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDProductionValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDSymbolSequenceValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDPriorityValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PRIORITY", []string{tcDirPriority.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDDHumanValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"HUMAN", []string{tcDirHuman.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDDTokenValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKEN", []string{tcDirToken.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDDStateshiftValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"STATESHIFT", []string{tcDirShift.ID(), "TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDDTokenOptionValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDTokenOptsValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDPattern(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PATTERN", []string{"TEXT"},
		"value",
		sdtsFnTrimString,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDSymbolValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDStateInstructionState(sdts ictiobus.SDTS) {
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

func bootstrapSDDIDExprValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDTextValue(sdts ictiobus.SDTS) {
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

func bootstrapSDDTextElementValue(sdts ictiobus.SDTS) {
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
