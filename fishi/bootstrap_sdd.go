package fishi

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/translation"
)

func CreateBootstrapSDD() ictiobus.SDTS {
	sdd := ictiobus.NewSDTS()

	// fill in the gaps until this part is fully written out
	bootstrapSDDFakeSynth(sdd, "BLOCK", []string{"TOKENS-BLOCK"}, "ast", astGrammarBlock{content: []astGrammarContent{
		{
			state: "COULD BE TOKENS, grammar block until done",
			rules: []grammar.Rule{
				{
					NonTerminal: "TOKEN",
					Productions: []grammar.Production{
						{"TOKEN", "TOKEN", "TOKEN"},
					},
				},
			},
		},
	}})

	bootstrapSDDFakeSynth(sdd, "BLOCK", []string{"ACTIONS-BLOCK"}, "ast", astGrammarBlock{content: []astGrammarContent{
		{
			state: "COULD BE ACTIONS, grammar block until done",
			rules: []grammar.Rule{
				{
					NonTerminal: "ACTION",
					Productions: []grammar.Production{
						{"ACTION", "ACTION", "ACTION"},
					},
				},
			},
		},
	}})

	// need these until we fill in the ACTIONs-BLOCK and TOKENS-BLOCK rules

	bootstrapSDDFishispecAST(sdd)
	bootstrapSDDBlocksValue(sdd)
	bootstrapSDDBlockAST(sdd)
	bootstrapSDDGrammarBlockAST(sdd)
	bootstrapSDDGrammarContentAST(sdd)
	bootstrapSDDGrammarStateBlockValue(sdd)

	// GRAMMAR branch mock, return later.
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULE"},
		"value",
		func(_, _ string, args []interface{}) interface{} {
			return []grammar.Rule{
				{
					NonTerminal: "FAKE-NT",
					Productions: []grammar.Production{
						{"FAKE-NT-1", "FAKE-NT-2", "FAKE-NT-3"},
					},
				},
			}
		},
		[]translation.AttrRef{},
	)

	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULES", []string{"GRAMMAR-RULES", "NEWLINES", "GRAMMAR-RULE"},
		"value",
		func(_, _ string, args []interface{}) interface{} {
			list, ok := args[0].([]grammar.Rule)
			if !ok {
				return []grammar.Rule{}
			}

			toAppend := grammar.Rule{
				NonTerminal: "FAKE-NT",
				Productions: []grammar.Production{
					{"FAKE-NT-1", "FAKE-NT-2", "FAKE-NT-3"},
				},
			}

			list = append(list, toAppend)
			return list
		},
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)

	// state-inst branch mocks.

	bootstrapSDDFakeSynth(sdd, "STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"}, "state", "someFakeState")
	bootstrapSDDFakeSynth(sdd, "STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"}, "state", "someFakeStateNoNL")

	/* steps for next part in SDD:
	// state-inst BRANCH:
	- mock out ID-EXPR for all three sets
	- uncomment bootstrapSDDStateInstructionState
	- remove STATE-INSTRUCTIONS mocks
	- ID-EXPR will likely NOT need noFlows, so be suspicious if hit that cond

	ONCE WORKING:
	- mock out TEXT for both sets
	- uncomment bootstrapSDDIdExprValue
	- remove ID-EXPR mocks
	- TEXT will surely need noFlows as it also rolls up to token blocks. add as needed.

	ONCE WORKING:
	- mock out TEXT-ELEMENT for both sets
	- uncomment bootstrapSDDTextValue
	- remove TEXT mocks
	- TEXT-ELEMENT will likely NOT need noFlows as it also rolls up to token blocks. be suspicious if hit that cond

	ONCE WORKING:
	- uncomment bootstrapSDDTextElementValue
	- remove TEXT-ELEMENT mocks
	- VALIDATE. if works, done with branch of CFG.



	// AFTER:
	- mock out GRAMMAR-RULE for both sets
	- uncommentGrammarRulesValue
	- remove GRAMMAR-RULES mocks
	*/

	/*

		bootstrapSDDGrammarRulesValue(sdd)
		bootstrapSDDGrammarRuleValue(sdd)
		bootstrapSDDAlternationsValue(sdd)
		bootstrapSDDProductionValue(sdd)
		bootstrapSDDSymbolSequenceValue(sdd)
		bootstrapSDDSymbolValue(sdd)
		bootstrapSDDStateInstructionState(sdd)
		bootstrapSDDIDExprValue(sdd)
		bootstrapSDDTextValue(sdd)
		bootstrapSDDTextElementValue(sdd)*/

	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "ACTIONS-STATE-BLOCK")
	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "TOKENS-STATE-BLOCK")

	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "ACTIONS-STATE-BLOCK")
	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "TOKENS-STATE-BLOCK")

	return sdd
}

func bootstrapSDDFakeSynth(sdd ictiobus.SDTS, head string, prod []string, name string, value interface{}) {
	sdd.BindSynthesizedAttribute(
		head, prod,
		name,
		func(_, _ string, args []interface{}) interface{} { return value },
		nil,
	)
}

func bootstrapSDDFishispecAST(sdd ictiobus.SDTS) {
	sdd.BindSynthesizedAttribute(
		"FISHISPEC", []string{"BLOCKS"},
		"ast",
		sddFnMakeFishispec,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "value"},
		},
	)
}

func bootstrapSDDBlocksValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDBlockAST(sdd ictiobus.SDTS) {
	sdd.BindSynthesizedAttribute(
		"BLOCK", []string{"GRAMMAR-BLOCK"},
		"ast",
		sddFnIdentity,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
		},
	)
	/*
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
		)*/
}

func bootstrapSDDGrammarBlockAST(sdd ictiobus.SDTS) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "GRAMMAR-CONTENT"},
		"ast",
		sddFnMakeGrammarBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "ast"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-BLOCK", []string{tcHeaderGrammar.ID(), "NEWLINES", "GRAMMAR-CONTENT"},
		"ast",
		sddFnMakeGrammarBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "ast"},
		},
	)
}

func bootstrapSDDGrammarContentAST(sdd ictiobus.SDTS) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-CONTENT", "GRAMMAR-STATE-BLOCK"},
		"ast",
		sddFnGrammarContentBlocksAppendStateBlock,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 1}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-CONTENT", []string{"GRAMMAR-CONTENT", "GRAMMAR-RULES"},
		"ast",
		sddFnGrammarContentBlocksAppendRuleList,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "ast"},
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

func bootstrapSDDGrammarStateBlockValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDGrammarRulesValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDGrammarRuleValue(sdd ictiobus.SDTS) {
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULE", []string{tcNonterminal.ID(), tcEq.ID(), "ALTERNATIONS"},
		"value",
		sddFnMakeRule,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
	sdd.BindSynthesizedAttribute(
		"GRAMMAR-RULE", []string{tcNonterminal.ID(), tcEq.ID(), "ALTERNATIONS", "NEWLINES"},
		"value",
		sddFnMakeRule,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 2}, Name: "value"},
		},
	)
}

func bootstrapSDDAlternationsValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDProductionValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDSymbolSequenceValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDSymbolValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDStateInstructionState(sdd ictiobus.SDTS) {
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

func bootstrapSDDIDExprValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDTextValue(sdd ictiobus.SDTS) {
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

func bootstrapSDDTextElementValue(sdd ictiobus.SDTS) {
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
