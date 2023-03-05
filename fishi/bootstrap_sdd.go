package fishi

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/translation"
)

func CreateBootstrapSDD() ictiobus.SDD {
	sdd := ictiobus.NewSDD()

	bootstrapSDDFishispecAST(sdd)
	bootstrapSDDBlocksValue(sdd)
	bootstrapSDDBlockAST(sdd)
	bootstrapSDDGrammarBlockAST(sdd)
	bootstrapSDDGrammarContentAST(sdd)
	bootstrapSDDGrammarStateBlockValue(sdd)
	bootstrapSDDGrammarRulesValue(sdd)
	bootstrapSDDGrammarRuleValue(sdd)
	bootstrapSDDAlternationsValue(sdd)
	bootstrapSDDProductionValue(sdd)
	bootstrapSDDSymbolSequenceValue(sdd)
	bootstrapSDDSymbolValue(sdd)
	bootstrapSDDStateInstructionState(sdd)
	bootstrapSDDIDExprValue(sdd)
	bootstrapSDDTextValue(sdd)
	bootstrapSDDTextElementValue(sdd)

	return sdd
}

func bootstrapSDDFishispecAST(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"FISHISPEC", []string{"BLOCKS"},
		"ast",
		sddFnMakeFishispec,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDBlocksValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"BLOCKS", []string{"BLOCKS", "BLOCK"},
		"value",
		sddFnBlockListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"BLOCKS", []string{"BLOCK"},
		"value",
		sddFnBlockListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
}

func bootstrapSDDBlockAST(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"BLOCK", []string{"GRAMMAR-BLOCK"},
		"ast",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"BLOCK", []string{"TOKENS-BLOCK"},
		"ast",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"BLOCK", []string{"ACTIONS-BLOCK"},
		"ast",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

func bootstrapSDDGrammarBlockAST(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "GRAMMAR-CONTENT"},
		"ast",
		sddFnMakeGrammarBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "NEWLINES", "GRAMMAR-CONTENT"},
		"ast",
		sddFnGrammarContentBlocksAppendStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
}

func bootstrapSDDGrammarContentAST(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-CONTENT", "GRAMMAR-STATE-BLOCK"},
		"ast",
		sddFnGrammarContentBlocksAppendStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-CONTENT", "GRAMMAR-RULES"},
		"ast",
		sddFnGrammarContentBlocksAppendRuleList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-STATE-BLOCK"},
		"ast",
		sddFnGrammarContentBlocksStartStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-RULES"},
		"ast",
		sddFnGrammarContentBlocksStartRuleList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDGrammarStateBlockValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-STATE-BLOCK", []string{"STATE-INSTRUCTION", "NEWLINES", "GRAMMAR-RULES"},
		"value",
		sddFnMakeGrammarContentNode,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "state"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDDGrammarRulesValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULES", "NEWLINES", "GRAMMAR-RULE"},
		"value",
		sddFnRuleListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULE"},
		"value",
		sddFnRuleListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDGrammarRuleValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULE", []string{tcNonterminal.ID(), tcEq.ID(), "ALTERNATIONS"},
		"value",
		sddFnMakeRule,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "$text"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULE", []string{tcNonterminal.ID(), tcEq.ID(), "ALTERNATIONS", "NEWLINES"},
		"value",
		sddFnMakeRule,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "$text"},
		},
	)
}

func bootstrapSDDAlternationsValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"PRODUCTION"},
		"value",
		sddFnStringListListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"ALTERNATIONS", tcAlt.ID(), "PRODUCTION"},
		"value",
		sddFnStringListListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"ALTERNATIONS", []string{"ALTERNATIONS", "NEWLINES", tcAlt.ID(), "PRODUCTION"},
		"value",
		sddFnStringListListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 3}, Name: "value"},
		},
	)
}

func bootstrapSDDProductionValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"PRODUCTION", []string{"SYMBOL-SEQUENCE"},
		"value",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"PRODUCTION", []string{tcEpsilon.ID()},
		"value",
		sddFnNilStringList,
		[]translation.AttrRef{},
	)
}

func bootstrapSDDSymbolSequenceValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"SYMBOL-SEQUENCE", []string{"SYMBOL-SEQUENCE", "SYMBOL"},
		"value",
		sddFnStringListAppend,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"SYMBOL-SEQUENCE", []string{"SYMBOL"},
		"value",
		sddFnStringListStart,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDSymbolValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"SYMBOL", []string{tcNonterminal.ID()},
		"value",
		sddFnGetNonterminal,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"SYMBOL", []string{tcTerminal.ID()},
		"value",
		sddFnGetTerminal,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDStateInstructionState(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"},
		"state",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"},
		"state",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDDIDExprValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"ID-EXPR", []string{tcId.ID()},
		"value",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"ID-EXPR", []string{tcTerminal.ID()},
		"value",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"ID-EXPR", []string{"TEXT"},
		"value",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDTextValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"TEXT", []string{"TEXT-ELEMENT"},
		"value",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"TEXT", []string{"TEXT", "TEXT-ELEMENT"},
		"value",
		sddFnAppendStrings,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
}

func bootstrapSDDTextElementValue(sdd ictiobus.SDD) {
	sdd.BindSynthesizedAttribute(
		"TEXT-ELEMENT", []string{tcFreeformText.ID()},
		"value",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"TEXT-ELEMENT", []string{tcEscseq.ID()},
		"value",
		sddFnInterpretEscape,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
}
