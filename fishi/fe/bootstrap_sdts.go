package fe

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/trans"
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
		"make_fishispec",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSBlocksValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"BLOCKS", []string{"BLOCKS", "BLOCK"},
		"value",
		"block_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"BLOCKS", []string{"BLOCK"},
		"value",
		"block_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
}

func bootstrapSDTSBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"BLOCK", []string{"GRAMMAR-BLOCK"},
		"ast",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"BLOCK", []string{"TOKENS-BLOCK"},
		"ast",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"BLOCK", []string{"ACTIONS-BLOCK"},
		"ast",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
}

func bootstrapSDTSActionsBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-BLOCK", []string{TCHeaderActions.ID(), "ACTIONS-CONTENT"},
		"ast",
		"make_actions_block",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

func bootstrapSDTSTokensBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-BLOCK", []string{TCHeaderTokens.ID(), "TOKENS-CONTENT"},
		"ast",
		"make_tokens_block",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

// TODO: finish converting
func bootstrapSDTSGrammarBlockAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-BLOCK", []string{TCHeaderGrammar.ID(), "GRAMMAR-CONTENT"},
		"ast",
		"make_grammar_block",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

func bootstrapSDTSTokensContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-ENTRIES"},
		"ast",
		"tokens_content_blocks_start_entry_list",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-STATE-BLOCK-LIST"},
		"ast",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-CONTENT", []string{"TOKENS-ENTRIES", "TOKENS-STATE-BLOCK-LIST"},
		"ast",
		"tokens_content_blocks_prepend",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionsContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST"},
		"ast",
		"actions_content_blocks_start_symbol_actions_list",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"ACTIONS-STATE-BLOCK-LIST"},
		"ast",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-CONTENT", []string{"SYMBOL-ACTIONS-LIST", "ACTIONS-STATE-BLOCK-LIST"},
		"ast",
		"actions_content_blocks_prepend",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarContentAST(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-RULES"},
		"ast",
		"grammar_content_blocks_start_rule_list",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-STATE-BLOCK-LIST"},
		"ast",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "list"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-RULES", "GRAMMAR-STATE-BLOCK-LIST"},
		"ast",
		"grammar_content_blocks_prepend",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "list"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK", []string{"STATE-INSTRUCTION", "GRAMMAR-RULES"},
		"value",
		"make_grammar_content_node",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "state"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionsStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "SYMBOL-ACTIONS-LIST"},
		"value",
		"make_actions_content_node",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "state"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensStateBlockValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-STATE-BLOCK", []string{"STATE-INSTRUCTION", "TOKENS-ENTRIES"},
		"value",
		"make_tokens_content_node",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "state"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSProdActionsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-ACTIONS", []string{"PROD-ACTIONS", "PROD-ACTION"},
		"value",
		"prod_action_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PROD-ACTIONS", []string{"PROD-ACTION"},
		"value",
		"prod_action_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSAttrRefsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ATTR-REFS", []string{"ATTR-REFS", TCAttrRef.ID()},
		"value",
		"attr_ref_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$ft"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"ATTR-REFS", []string{TCAttrRef.ID()},
		"value",
		"attr_ref_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$ft"},
		},
	)
}

func bootstrapSDTSWithClauseValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"WITH-CLAUSE", []string{TCDirWith.ID(), "ATTR-REFS"},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSSemanticActionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTION", []string{TCDirSet.ID(), TCAttrRef.ID(), TCDirHook.ID(), TCId.ID()},
		"value",
		"make_semantic_action",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$ft"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 3}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 3}, Name: "$ft"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTION", []string{TCDirSet.ID(), TCAttrRef.ID(), TCDirHook.ID(), TCId.ID(), "WITH-CLAUSE"},
		"value",
		"make_semantic_action",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$ft"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 3}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 3}, Name: "$ft"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 4}, Name: "value"},
		},
	)
}

func bootstrapSDTSSemanticActionsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTIONS", "SEMANTIC-ACTION"},
		"value",
		"semantic_action_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"SEMANTIC-ACTIONS", []string{"SEMANTIC-ACTION"},
		"value",
		"semantic_action_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionSymbolValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{TCNonterminal.ID()},
		"value",
		"get_nonterminal",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{TCTerminal.ID()},
		"value",
		"get_terminal",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{TCInt.ID()},
		"value",
		"get_int",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL", []string{TCId.ID()},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}

func bootstrapSDTSActionSymbolSequenceValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL-SEQUENCE", "ACTION-SYMBOL"},
		"value",
		"string_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-SYMBOL-SEQUENCE", []string{"ACTION-SYMBOL"},
		"value",
		"string_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionProductionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTION-PRODUCTION", []string{"ACTION-SYMBOL-SEQUENCE"},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTION-PRODUCTION", []string{TCEpsilon.ID()},
		"value",
		"epsilon_string_list",
		nil,
	)
}

func bootstrapSDTSProdAddrValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-ADDR", []string{TCDirIndex.ID(), TCInt.ID()},
		"value",
		"make_prod_specifier_index",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PROD-ADDR", []string{"ACTION-PRODUCTION"},
		"value",
		"make_prod_specifier_literal",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSProdSpecifierValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-SPECIFIER", []string{TCDirProd.ID(), "PROD-ADDR"},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PROD-SPECIFIER", []string{TCDirProd.ID()},
		"value",
		"make_prod_specifier_next",
		nil,
	)
}

func bootstrapSDTSProdActionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PROD-ACTION", []string{"PROD-SPECIFIER", "SEMANTIC-ACTIONS"},
		"value",
		"make_prod_action",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSSymbolActionsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL-ACTIONS", []string{TCDirSymbol.ID(), TCNonterminal.ID(), "PROD-ACTIONS"},
		"value",
		"make_symbol_actions",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$ft"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDTSSymbolActionsListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS-LIST", "SYMBOL-ACTIONS"},
		"value",
		"symbol_actions_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"SYMBOL-ACTIONS-LIST", []string{"SYMBOL-ACTIONS"},
		"value",
		"symbol_actions_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSActionsStateBlockListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK-LIST", "ACTIONS-STATE-BLOCK"},
		"value",
		"actions_state_block_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ACTIONS-STATE-BLOCK-LIST", []string{"ACTIONS-STATE-BLOCK"},
		"value",
		"actions_state_block_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensStateBlockListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-STATE-BLOCK-LIST", []string{"TOKENS-STATE-BLOCK-LIST", "TOKENS-STATE-BLOCK"},
		"value",
		"tokens_state_block_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-STATE-BLOCK-LIST", []string{"TOKENS-STATE-BLOCK"},
		"value",
		"tokens_state_block_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarStateBlockListValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK-LIST", "GRAMMAR-STATE-BLOCK"},
		"list",
		"grammar_state_block_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "list"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK-LIST", []string{"GRAMMAR-STATE-BLOCK"},
		"list",
		"grammar_state_block_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarRulesValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULES", "GRAMMAR-RULE"},
		"value",
		"rule_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULE"},
		"value",
		"rule_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensEntriesValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRIES", []string{"TOKENS-ENTRIES", "TOKENS-ENTRY"},
		"value",
		"entry_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRIES", []string{"TOKENS-ENTRY"},
		"value",
		"entry_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokensEntryValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKENS-ENTRY", []string{"PATTERN", "TOKEN-OPTS"},
		"value",
		"make_token_entry",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSGrammarRuleValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"GRAMMAR-RULE", []string{TCLineStartNonterminal.ID(), TCEq.ID(), "ALTERNATIONS"},
		"value",
		"make_rule",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDTSAlternationsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"PRODUCTION"},
		"value",
		"string_list_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"ALTERNATIONS", TCAlt.ID(), "PRODUCTION"},
		"value",
		"string_list_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDTSProductionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PRODUCTION", []string{"SYMBOL-SEQUENCE"},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"PRODUCTION", []string{TCEpsilon.ID()},
		"value",
		"epsilon_string_list",
		nil,
	)
}

func bootstrapSDTSSymbolSequenceValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL-SEQUENCE", []string{"SYMBOL-SEQUENCE", "SYMBOL"},
		"value",
		"string_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"SYMBOL-SEQUENCE", []string{"SYMBOL"},
		"value",
		"string_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSPriorityValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PRIORITY", []string{TCDirPriority.ID(), "TEXT"},
		"value",
		"trim_string",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSHumanValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"HUMAN", []string{TCDirHuman.ID(), "TEXT"},
		"value",
		"trim_string",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokenValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKEN", []string{TCDirToken.ID(), "TEXT"},
		"value",
		"trim_string",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSStateshiftValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"STATESHIFT", []string{TCDirShift.ID(), "TEXT"},
		"value",
		"trim_string",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokenOptionValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"DISCARD"},
		"value",
		"make_discard_option",
		nil,
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"STATESHIFT"},
		"value",
		"make_stateshift_option",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"TOKEN"},
		"value",
		"make_token_option",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"HUMAN"},
		"value",
		"make_human_option",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTION", []string{"PRIORITY"},
		"value",
		"make_priority_option",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSTokenOptsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTS", []string{"TOKEN-OPTS", "TOKEN-OPTION"},
		"value",
		"token_opt_list_append",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TOKEN-OPTS", []string{"TOKEN-OPTION"},
		"value",
		"token_opt_list_start",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSPattern(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"PATTERN", []string{"TEXT"},
		"value",
		"trim_string",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSSymbolValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"SYMBOL", []string{TCNonterminal.ID()},
		"value",
		"get_nonterminal",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"SYMBOL", []string{TCTerminal.ID()},
		"value",
		"get_terminal",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}

func bootstrapSDTSStateInstructionState(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"STATE-INSTRUCTION", []string{TCDirState.ID(), "ID-EXPR"},
		"state",
		"make_state_ins",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "$ft"},
		},
	)
}

func bootstrapSDTSIDExprValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"ID-EXPR", []string{TCId.ID()},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"ID-EXPR", []string{TCTerminal.ID()},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}

func bootstrapSDTSTextValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TEXT", []string{"LINE-START-TEXT-ELEMENT", "TEXT-ELEMENTS"},
		"value",
		"append_strings_trimmed",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TEXT", []string{"TEXT-ELEMENTS"},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TEXT", []string{"LINE-START-TEXT-ELEMENT"},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}
func bootstrapSDTSTextElementsValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENTS", []string{"TEXT-ELEMENTS", "TEXT-ELEMENT"},
		"value",
		"append_strings",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENTS", []string{"TEXT-ELEMENT"},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDTSLineStartTextElementValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"LINE-START-TEXT-ELEMENT", []string{TCLineStartFreeformText.ID()},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	sdts.BindSynthesizedAttribute(
		"LINE-START-TEXT-ELEMENT", []string{TCLineStartEscseq.ID()},
		"value",
		"interpret_escape",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}

func bootstrapSDTSTextElementValue(sdts ictiobus.SDTS) {
	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENT", []string{TCFreeformText.ID()},
		"value",
		"identity",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdts.BindSynthesizedAttribute(
		"TEXT-ELEMENT", []string{TCEscseq.ID()},
		"value",
		"interpret_escape",
		[]trans.AttrRef{
			{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}
