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
	bootstrapSDDGrammarRulesValue(sdd)
	bootstrapSDDGrammarRuleValue(sdd)
	bootstrapSDDStateInstructionState(sdd)
	bootstrapSDDIDExprValue(sdd)
	bootstrapSDDTextValue(sdd)
	bootstrapSDDTextElementValue(sdd)
	bootstrapSDDAlternationsValue(sdd)
	bootstrapSDDProductionValue(sdd)
	bootstrapSDDSymbolSequenceValue(sdd)
	bootstrapSDDSymbolValue(sdd)

	// NEXT STEPS:
	//
	// TOKENS-BLOCK:
	// - AST struct for it
	// - Mock all 4 TOKENS-CONTENT rules
	// - create function bootstrapSDDTokensBlockAST
	// - update AST string() to print out the tokens AST block
	// - remove TOKENS-BLOCK mock for BLOCKS
	// - uncomment BLOCK -> TOKENS-BLOCK rule in bootstrapSDDBlockAST
	//
	// TOKENS-CONTENT:
	// - AST struct for it
	// - Mock TOKENS-STATE-BLOCK
	// - Mock both TOKENS-ENTIRIES
	// - create function bootstrapSDDTokensContentAST
	// - update AST string() to print out the tokens AST content block
	// - remove TOKENS-CONTENT mock
	//
	// TOKENS-STATE-BLOCK:
	// - (TOKENS-ENTRIES should already be mocked)
	// - create function bootstrapSDDTokensStateBlockAST
	// - update AST string() to print out the tokens state block
	// - remove NoFlow STATE-INSTRUCTION -> TOKENS-STATE-BLOCK
	// - remove TOKENS-STATE-BLOCK mock
	//
	// TOKENS-ENTRIES:
	// - Mock all four TOKENS-ENTRY rules
	// - create function bootstrapSDDTokensEntriesValue
	// - remove TOKENS-ENTRIES mock
	//
	// TOKENS-ENTRY:
	// - Mock PATTERN rule
	// - Mock all three TOKEN-OPTS rules
	// - create function bootstrapSDDTokensEntryValue
	// - remove TOKENS-ENTRY mock
	//
	// PATTERN:
	// - create function bootstrapSDDPatternValue
	// - remove PATTERN mock
	// - remove NoFlow TEXT -> PATTERN
	//
	// TOKEN-OPTS:
	// - Mock all five TOKEN-OPTION rules
	// - create function bootstrapSDDTokenOptsValue
	// - remove TOKEN-OPTS mock
	//
	// TOKEN-OPTION:
	// - Mock DISCARD rule
	// - Mock STATESHIFT rule
	// - Mock TOKEN rule
	// - Mock HUMAN rule
	// - Mock PRIORITY rule
	// - create function bootstrapSDDTokenOptionValue
	// - remove TOKEN-OPTION mock
	//
	// DISCARD:
	// - create function bootstrapSDDDiscardValue
	// - remove DISCARD mock
	//
	// STATESHIFT:
	// - create function bootstrapSDDStateShiftValue
	// - remove STATESHIFT mock
	// - remove NoFlow TEXT -> STATESHIFT
	//
	// TOKEN:
	// - create function bootstrapSDDTokenValue
	// - remove TOKEN mock
	// - remove NoFlow TEXT -> TOKEN
	//
	// HUMAN:
	// - create function bootstrapSDDHumanValue
	// - remove HUMAN mock
	// - remove NoFlow TEXT -> HUMAN
	//
	// PRIORITY:
	// - create function bootstrapSDDPriorityValue
	// - remove PRIORITY mock
	// - remove NoFlow TEXT -> PRIORITY
	//
	// BREAK HERE (done with token branch)
	//
	// ACTIONS-BLOCK:
	// - AST struct for it
	// - Mock all 4 ACTIONS-CONTENT rules
	// - create function bootstrapSDDActionsBlockAST
	// - update AST string() to print out the actions AST block
	// - remove ACTIONS-BLOCK mock for BLOCKS
	// - uncomment BLOCK -> ACTIONS-BLOCK rule in bootstrapSDDBlockAST
	//
	// ACTIONS-CONTENT:
	// - AST struct for it
	// - Mock ACTIONS-STATE-BLOCK
	// - Mock both SYMBOL-ACTIONS-LIST rules
	// - create function bootstrapSDDActionsContentAST
	// - update AST string() to print out the actions AST content block
	// - remove ACTIONS-CONTENT mock
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

	// permanently in place until tokens and actions branches are started.

	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "ACTIONS-STATE-BLOCK")
	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "NEWLINES", "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "TOKENS-STATE-BLOCK")

	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "ACTIONS-STATE-BLOCK")
	sdd.SetNoFlow(true, "STATE-INSTRUCTION", []string{tcDirState.ID(), "ID-EXPR"}, "state", translation.NodeRelation{}, -1, "TOKENS-STATE-BLOCK")

	sdd.SetNoFlow(true, "TEXT", []string{"TEXT", "TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "PATTERN")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "PATTERN")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT", "TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "STATESHIFT")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "STATESHIFT")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT", "TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "TOKEN")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "TOKEN")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT", "TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "PRIORITY")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "PRIORITY")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT", "TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "HUMAN")
	sdd.SetNoFlow(true, "TEXT", []string{"TEXT-ELEMENT"}, "value", translation.NodeRelation{}, -1, "HUMAN")

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
		sddFnEpsilonStringList,
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
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
		},
	)

	sdd.BindSynthesizedAttribute(
		"SYMBOL", []string{tcTerminal.ID()},
		"value",
		sddFnGetTerminal,
		[]translation.AttrRef{
			{Relation: translation.NodeRelation{Type: translation.RelSymbol, Index: 0}, Name: "$text"},
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
