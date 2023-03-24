package fishi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
)

const (
	ErrString            = "<ERR>"
	ErrWithMessageString = "<ERR: %s>"
)

func SDDErrMsg(msg string, a ...interface{}) string {
	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	return fmt.Sprintf(ErrWithMessageString, msg)
}

func sdtsFnMakeFishispec(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astBlock)
	if !ok {
		return AST{}
	}

	return AST{nodes: list}
}

func sdtsFnBlockListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astBlock)
	if !ok {
		return []astBlock{}
	}

	toAppend, ok := args[1].(astBlock)
	if !ok {
		var errBl astErrorBlock
		toAppend = errBl
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnBlockListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astBlock)
	if !ok {
		var errBl astErrorBlock
		toAppend = errBl
	}

	return []astBlock{toAppend}
}

func sdtsFnMakeGrammarBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	return astGrammarBlock{content: list}
}

func sdtsFnMakeTokensBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	return astTokensBlock{content: list}
}

func sdtsFnMakeActionsBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astActionsContent)
	if !ok {
		list = []astActionsContent{{state: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	return astActionsBlock{content: list}
}

func sdtsFnGrammarContentBlocksAppendStateBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}
	toAppend, ok := args[1].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{state: SDDErrMsg("producing this grammar content: second argument is not a grammar content")}
	}

	// TODO: the following nonsense is needed because GRAMMAR-RULES -> GRAMMAR-RULES GRAMMAR-RULE may never get invoked,
	// in favor ofsimply having a list of GRAMMAR-CONTENTs. in future, check this, and try to force grammar-content to
	// be list. for now, the parse tree works fine. So we'll just glue it together HERE too.

	// if toAppend's state is the same as an existing one, we simply add to the existing list instead of appending.
	if ok {
		for i := range list {
			if list[i].state == toAppend.state {
				list[i].rules = append(list[i].rules, toAppend.rules...)
				return list
			}
		}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnActionsContentBlocksAppendStateBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astActionsContent)
	if !ok {
		list = []astActionsContent{{state: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}
	toAppend, ok := args[1].(astActionsContent)
	if !ok {
		toAppend = astActionsContent{state: SDDErrMsg("producing this actions content: second argument is not an actions content")}
	}

	// TODO: the following nonsense is needed because GRAMMAR-RULES -> GRAMMAR-RULES GRAMMAR-RULE may never get invoked,
	// in favor ofsimply having a list of GRAMMAR-CONTENTs. in future, check this, and try to force grammar-content to
	// be list. for now, the parse tree works fine. So we'll just glue it together HERE too.
	// same for SYMBOL-ACTIONS-LIST

	// if toAppend's state is the same as an existing one, we simply add to the existing list instead of appending.
	if ok {
		for i := range list {
			if list[i].state == toAppend.state {
				list[i].actions = append(list[i].actions, toAppend.actions...)
				return list
			}
		}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnTokensContentBlocksAppendStateBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}
	toAppend, ok := args[1].(astTokensContent)
	if !ok {
		toAppend = astTokensContent{state: SDDErrMsg("producing this tokens content: second argument is not a tokens content")}
	}

	// TODO: the following nonsense is needed because GRAMMAR-RULES -> GRAMMAR-RULES GRAMMAR-RULE may never get invoked,
	// in favor ofsimply having a list of GRAMMAR-CONTENTs. in future, check this, and try to force grammar-content to
	// be list. for now, the parse tree works fine. So we'll just glue it together HERE too.
	// (this same thing applies to TOKENS, which has the same pattern)

	// if toAppend's state is the same as an existing one, we simply add to the existing list instead of appending.
	if ok {
		for i := range list {
			if list[i].state == toAppend.state {
				list[i].entries = append(list[i].entries, toAppend.entries...)
				return list
			}
		}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnGrammarContentBlocksStartStateBlock(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{state: SDDErrMsg("producing this grammar content: first argument is not a grammar content")}
	}

	return []astGrammarContent{toAppend}
}

func sdtsFnTokensContentBlocksStartStateBlock(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astTokensContent)
	if !ok {
		toAppend = astTokensContent{state: SDDErrMsg("producing this tokens content: first argument is not a tokens content")}
	}

	return []astTokensContent{toAppend}
}

func sdtsFnActionsContentBlocksStartStateBlock(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astActionsContent)
	if !ok {
		toAppend = astActionsContent{state: SDDErrMsg("producing this actions content: first argument is not an actions content")}
	}

	return []astActionsContent{toAppend}
}

func sdtsFnGrammarContentBlocksAppendRuleList(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	rules, ok := args[1].([]grammar.Rule)
	if !ok {
		rules = []grammar.Rule{}
	}
	toAppend := astGrammarContent{
		rules: rules,
		state: "",
	}
	if !ok {
		toAppend.state = SDDErrMsg("producing the rule list for this content block: second argument is not a rule list")
	}

	// TODO: the following nonsense is needed because GRAMMAR-RULES -> GRAMMAR-RULES GRAMMAR-RULE may never get invoked,
	// in favor ofsimply having a list of GRAMMAR-CONTENTs. in future, check this, and try to force grammar-content to
	// be list. for now, the parse tree works fine. So we'll just glue it together HERE too.

	// if toAppend's state is the same as an existing one, we simply add to the existing list instead of appending.
	if ok {
		for i := range list {
			if list[i].state == toAppend.state {
				list[i].rules = append(list[i].rules, toAppend.rules...)
				return list
			}
		}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnTokensContentBlocksAppendEntryList(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	entries, ok := args[1].([]tokenEntry)
	if !ok {
		entries = []tokenEntry{{pattern: SDDErrMsg("producing this token entry list: second argument is not a token entry list")}}
	}
	toAppend := astTokensContent{
		entries: entries,
		state:   "",
	}
	if !ok {
		toAppend.state = SDDErrMsg("producing this token entry list for this content block: second argument is not a token entry list list")
	}

	// TODO: the following nonsense is needed because GRAMMAR-RULES -> GRAMMAR-RULES GRAMMAR-RULE may never get invoked,
	// in favor ofsimply having a list of GRAMMAR-CONTENTs. in future, check this, and try to force grammar-content to
	// be list. for now, the parse tree works fine. So we'll just glue it together HERE too.
	// also applies to TOKEN

	// if toAppend's state is the same as an existing one, we simply add to the existing list instead of appending.
	if ok {
		for i := range list {
			if list[i].state == toAppend.state {
				list[i].entries = append(list[i].entries, toAppend.entries...)
				return list
			}
		}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnActionsContentBlocksAppendSymbolActionsList(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astActionsContent)
	if !ok {
		list = []astActionsContent{{state: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	actions, ok := args[1].([]symbolActions)
	if !ok {
		actions = []symbolActions{{symbol: SDDErrMsg("producing this symbol actions list: second argument is not a symbol actions list")}}
	}
	toAppend := astActionsContent{
		actions: actions,
		state:   "",
	}
	if !ok {
		toAppend.state = SDDErrMsg("producing this symbol actions list for this content block: second argument is not a symbol actions list")
	}

	// TODO: the following nonsense is needed because GRAMMAR-RULES -> GRAMMAR-RULES GRAMMAR-RULE may never get invoked,
	// in favor ofsimply having a list of GRAMMAR-CONTENTs. in future, check this, and try to force grammar-content to
	// be list. for now, the parse tree works fine. So we'll just glue it together HERE too.
	// also applies to ACTIONS

	// if toAppend's state is the same as an existing one, we simply add to the existing list instead of appending.
	if ok {
		for i := range list {
			if list[i].state == toAppend.state {
				list[i].actions = append(list[i].actions, toAppend.actions...)
				return list
			}
		}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnGrammarContentBlocksStartRuleList(_, _ string, args []interface{}) interface{} {
	rules, ok := args[0].([]grammar.Rule)
	if !ok {
		rules = []grammar.Rule{{NonTerminal: SDDErrMsg("producing this rule list: first argument is not a rule list")}}
	}
	toAppend := astGrammarContent{
		rules: rules,
		state: "",
	}

	return []astGrammarContent{toAppend}
}

func sdtsFnTokensContentBlocksStartEntryList(_, _ string, args []interface{}) interface{} {
	entries, ok := args[0].([]tokenEntry)
	if !ok {
		entries = []tokenEntry{{pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}
	toAppend := astTokensContent{
		entries: entries,
		state:   "",
	}

	return []astTokensContent{toAppend}
}

func sdtsFnActionsContentBlocksStartSymbolActionsList(_, _ string, args []interface{}) interface{} {
	actions, ok := args[0].([]symbolActions)
	if !ok {
		actions = []symbolActions{{symbol: SDDErrMsg("producing this symbol actions list: first argument is not a symbol actions list")}}
	}
	toAppend := astActionsContent{
		actions: actions,
		state:   "",
	}

	return []astActionsContent{toAppend}
}

func sdtsFnMakeProdAction(_, _ string, args []interface{}) interface{} {
	prodSpec, ok := args[0].(box.Pair[string, interface{}])
	if !ok {
		prodSpec = box.Pair[string, interface{}]{"LITERAL", []string{SDDErrMsg("producing this production action: first argument is not a pair of string, any")}}
	}

	semActions, ok := args[1].([]semanticAction)
	if !ok {
		semActions = []semanticAction{{hook: SDDErrMsg("producing this production action: second argument is not a semantic action list")}}
	}

	pa := productionAction{
		actions: semActions,
	}

	if prodSpec.First == "LITERAL" {
		pa.prodLiteral = prodSpec.Second.([]string)
	} else if prodSpec.First == "INDEX" {
		pa.prodIndex = prodSpec.Second.(int)
	} else if prodSpec.First == "NEXT" {
		pa.prodNext = true
	} else {
		pa.prodLiteral = []string{SDDErrMsg("producing this production action: first argument is not a pair of string/interface{}")}
	}

	return pa
}

func sdtsFnMakeSymbolActions(_, _ string, args []interface{}) interface{} {
	nonTermUntyped := sdtsFnGetNonterminal("", "", args[0:1])
	nonTerm := nonTermUntyped.(string)

	prodActions, ok := args[1].([]productionAction)
	if !ok {
		prodActions = []productionAction{{prodLiteral: []string{SDDErrMsg("producing this production action list: second argument is not a production action list")}}}
	}

	sa := symbolActions{
		symbol:  nonTerm,
		actions: prodActions,
	}

	return sa
}

func sdtsFnMakeGrammarContentNode(_, _ string, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		state = SDDErrMsg("STATE value is not a string")
	}
	rules, ok := args[1].([]grammar.Rule)
	if !ok {
		rules = []grammar.Rule{{NonTerminal: SDDErrMsg("producing this rule list: second argument is not a rule list")}}
	}
	return astGrammarContent{rules: rules, state: state}
}

func sdtsFnMakeActionsContentNode(_, _ string, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		state = SDDErrMsg("STATE value is not a string")
	}
	actions, ok := args[1].([]symbolActions)
	if !ok {
		actions = []symbolActions{{symbol: SDDErrMsg("producing this symbol actions list: second argument is not a symbol actions list")}}
	}
	return astActionsContent{actions: actions, state: state}
}

func sdtsFnMakeTokensContentNode(_, _ string, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		state = SDDErrMsg("STATE value is not a string")
	}
	entries, ok := args[1].([]tokenEntry)
	if !ok {
		entries = []tokenEntry{{pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}
	return astTokensContent{entries: entries, state: state}
}

func sdtsFnTrimString(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}
	return strings.TrimSpace(str)
}

func sdtsFnMakeDiscardOption(_, _ string, args []interface{}) interface{} {
	return astTokenOption{optType: tokenOptDiscard}
}

func sdtsFnMakeStateshiftOption(_, _ string, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptStateshift, value: state}
}

func sdtsFnMakeHumanOption(_, _ string, args []interface{}) interface{} {
	human, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptHuman, value: human}
}

func sdtsFnMakeTokenOption(_, _ string, args []interface{}) interface{} {
	tok, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptToken, value: tok}
}

func sdtsFnMakePriorityOption(_, _ string, args []interface{}) interface{} {
	priority, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptPriority, value: priority}
}

func sdtsFnIdentity(_, _ string, args []interface{}) interface{} { return args[0] }

func sdtsFnInterpretEscape(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("escape sequence $text is not a string")
	}

	if len(str) < len("%!") {
		return SDDErrMsg("escape sequence $text does not appear to have enough characters: %q", str)
	}

	// escape sequence is %!, so just take the chars after that
	return str[len("%!"):]
}

func sdtsFnAppendStrings(_, _ string, args []interface{}) interface{} {
	str1, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("first argument is not a string")
	}
	str2, ok := args[1].(string)
	if !ok {
		return SDDErrMsg("second argument is not a string")
	}

	return str1 + str2
}

func sdtsFnGetNonterminal(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return str
}

func sdtsFnGetInt(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return -1
	}

	iVal, err := strconv.Atoi(str)
	if err != nil {
		iVal = -1
	}
	return iVal
}

func sdtsFnGetTerminal(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return strings.ToLower(str)
}

func sdtsFnRuleListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]grammar.Rule)
	if !ok {
		list = []grammar.Rule{{NonTerminal: SDDErrMsg("producing this rule list: first argument is not a rule list")}}
	}

	toAppend, ok := args[1].(grammar.Rule)
	if !ok {
		toAppend = grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: second argument is not a rule")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnEntryListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]tokenEntry)
	if !ok {
		list = []tokenEntry{{pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}

	toAppend, ok := args[1].(tokenEntry)
	if !ok {
		toAppend = tokenEntry{pattern: SDDErrMsg("producing this token entry: second argument is not a token entry")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnSymbolActionsListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]symbolActions)
	if !ok {
		list = []symbolActions{{symbol: SDDErrMsg("producing this symbol actions list: first argument is not a symbol actions list")}}
	}

	toAppend, ok := args[1].(symbolActions)
	if !ok {
		toAppend = symbolActions{symbol: SDDErrMsg("producing this symbol actions: second argument is not a symbol actions")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnProdActionListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]productionAction)
	if !ok {
		list = []productionAction{{prodLiteral: []string{SDDErrMsg("producing this production action list: first argument is not a production actions list")}}}
	}

	toAppend, ok := args[1].(productionAction)
	if !ok {
		toAppend = productionAction{prodLiteral: []string{SDDErrMsg("producing this production action list: first argument is not a production action")}}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnSemanticActionListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]semanticAction)
	if !ok {
		list = []semanticAction{{hook: SDDErrMsg("producing this semantic action list: first argument is not a semantic actions list")}}
	}

	toAppend, ok := args[1].(semanticAction)
	if !ok {
		toAppend = semanticAction{hook: SDDErrMsg("producing this semantic action list: first argument is not a semantic action")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnAttrRefListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]AttrRef)
	if !ok {
		list = []AttrRef{{symbol: SDDErrMsg("producing this AttrRef list: first argument is not an AttrRef list")}}
	}

	toAppend := sdtsFnGetAttrRef("", "", args[1:]).(AttrRef)

	list = append(list, toAppend)
	return list
}

func sdtsFnAttrRefListStart(_, _ string, args []interface{}) interface{} {
	toAppend := sdtsFnGetAttrRef("", "", args[0:]).(AttrRef)

	return []AttrRef{toAppend}
}

func sdtsFnGetAttrRef(_, _ string, args []interface{}) interface{} {
	var attrRef AttrRef

	str, ok := args[0].(string)
	if !ok {
		attrRef = AttrRef{symbol: SDDErrMsg("producing this semantic action: first argument is not a string to be parsed into an AttrRef")}
	} else {
		var err error
		attrRef, err = ParseAttrRef(str)
		if err != nil {
			attrRef = AttrRef{symbol: SDDErrMsg("producing this semantic action: first argument is not a valid AttrRef: %v", err.Error())}
		}
	}

	return attrRef
}

func sdtsFnMakeSemanticAction(_, _ string, args []interface{}) interface{} {
	attrRef := sdtsFnGetAttrRef("", "", args[0:1]).(AttrRef)

	hookId, ok := args[1].(string)
	if !ok {
		hookId = SDDErrMsg("producing this semantic action: second argument is not a string")
	}

	var argRefs []AttrRef
	if len(args) > 2 {
		argRefs, ok = args[2].([]AttrRef)
		if !ok {
			argRefs = []AttrRef{{symbol: SDDErrMsg("producing this semantic action: third argument is not an attrRef list")}}
		}
	}

	sa := semanticAction{
		lhs:  attrRef,
		hook: hookId,
		with: argRefs,
	}

	return sa
}

func sdtsFnMakeProdSpecifierNext(_, _ string, args []interface{}) interface{} {
	// need exact generic-filled type to match later expectations.
	spec := box.Pair[string, interface{}]{"NEXT", ""}
	return spec
}

func sdtsFnMakeProdSpecifierIndex(_, _ string, args []interface{}) interface{} {
	index := sdtsFnGetInt("", "", args)
	// need exact generic-filled type to match later expectations.
	spec := box.Pair[string, interface{}]{"INDEX", index}
	return spec
}

func sdtsFnMakeProdSpecifierLiteral(_, _ string, args []interface{}) interface{} {
	prod, ok := args[0].([]string)

	if !ok {
		prod = []string{SDDErrMsg("producing this production specifier: first argument is not an action production")}
	}

	// need exact generic-filled type to match later expectations.
	spec := box.Pair[string, interface{}]{"LITERAL", prod}
	return spec
}

func sdtsFnProdActionListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(productionAction)
	if !ok {
		toAppend = productionAction{prodLiteral: []string{SDDErrMsg("producing this production action list: first argument is not a production action")}}
	}

	return []productionAction{toAppend}
}

func sdtsFnSemanticActionListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(semanticAction)
	if !ok {
		toAppend = semanticAction{hook: SDDErrMsg("producing this semantic action list: first argument is not a semantic actions")}
	}

	return []semanticAction{toAppend}
}

func sdtsFnRuleListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(grammar.Rule)
	if !ok {
		toAppend = grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: first argument is not a rule")}
	}

	return []grammar.Rule{toAppend}
}

func sdtsFnSymbolActionsListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(symbolActions)
	if !ok {
		toAppend = symbolActions{symbol: SDDErrMsg("producing this symbol action: first argument is not a rule")}
	}

	return []symbolActions{toAppend}
}

func sdtsFnEntryListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(tokenEntry)
	if !ok {
		toAppend = tokenEntry{pattern: SDDErrMsg("producing this token entry: first argument is not a token entry")}
	}

	return []tokenEntry{toAppend}
}

func sdtsFnStringListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]string)
	if !ok {
		return []string{}
	}

	toAppend, ok := args[1].(string)
	if !ok {
		toAppend = ErrString
	}

	list = append(list, toAppend)

	return list
}

func sdtsFnTokenOptListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astTokenOption)
	if !ok {
		toAppend = astTokenOption{value: SDDErrMsg("first argument is not a token option")}
	}

	return []astTokenOption{toAppend}
}

func sdtsFnTokenOptListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokenOption)
	if !ok {
		return []astTokenOption{{value: SDDErrMsg("producing this token option list: first argument is not a token option list")}}
	}

	toAppend, ok := args[1].(astTokenOption)
	if !ok {
		toAppend = astTokenOption{value: SDDErrMsg("producing this token option: second argument is not a token option")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnStringListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(string)
	if !ok {
		toAppend = ErrString
	}

	return []string{toAppend}
}

func sdtsFnStringListListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].([]string)
	if !ok {
		toAppend = []string{SDDErrMsg("producing this string list: first argument is not a string list")}
	}

	return [][]string{toAppend}
}

func sdtsFnStringListListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([][]string)
	if !ok {
		return [][]string{{SDDErrMsg("producing this string list list: first argument is not a [][]string")}}
	}

	toAppend, ok := args[1].([]string)
	if !ok {
		toAppend = []string{SDDErrMsg("producing this string list: second argument is not a string list")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnEpsilonStringList(_, _ string, args []interface{}) interface{} {
	strList := grammar.Epsilon.Copy()
	return []string(strList)
}

func sdtsFnMakeRule(_, _ string, args []interface{}) interface{} {
	ntInterface := sdtsFnGetNonterminal("", "", args[0:1])

	nt, ok := ntInterface.(string)
	if !ok {
		nt = SDDErrMsg("first argument is not a string")
	}

	productions, ok := args[1].([][]string)
	if !ok {
		productions = [][]string{{SDDErrMsg("producing this list of lists of strings: second argument is not a [][]string")}}
	}

	r := grammar.Rule{NonTerminal: nt, Productions: []grammar.Production{}}

	for _, p := range productions {
		r.Productions = append(r.Productions, p)
	}

	return r
}

func sdtsFnMakeTokenEntry(_, _ string, args []interface{}) interface{} {
	pattern, ok := args[0].(string)
	if !ok {
		pattern = SDDErrMsg("first argument (pattern) is not a string")
	}

	tokenOpts, ok := args[1].([]astTokenOption)
	if !ok {
		tokenOpts = []astTokenOption{{value: SDDErrMsg("producing this token option list: second argument (tokenOpts) is not a token option list")}}
	}

	t := tokenEntry{pattern: pattern}

	for _, opt := range tokenOpts {
		switch opt.optType {
		case tokenOptDiscard:
			t.discard = true
		case tokenOptHuman:
			t.human = opt.value
		case tokenOptPriority:
			prior, err := strconv.Atoi(opt.value)
			if err == nil {
				t.priority = prior
			}
		case tokenOptStateshift:
			t.shift = opt.value
		case tokenOptToken:
			t.token = opt.value
		}
	}
	return t
}
