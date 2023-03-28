package fishi

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/translation"
)

const (
	ErrString            = "<ERR>"
	ErrWithMessageString = "<ERR: %s>"
)

var (
	HooksTable = map[string]translation.AttributeSetter{
		"make_fishispec":                                   sdtsFnMakeFishispec,
		"block_list_append":                                sdtsFnBlockListAppend,
		"block_list_start":                                 sdtsFnBlockListStart,
		"make_grammar_block":                               sdtsFnMakeGrammarBlock,
		"make_tokens_block":                                sdtsFnMakeTokensBlock,
		"make_actions_block":                               sdtsFnMakeActionsBlock,
		"grammar_content_blocks_start_rule_list":           sdtsFnGrammarContentBlocksStartRuleList,
		"tokens_content_blocks_start_entry_list":           sdtsFnTokensContentBlocksStartEntryList,
		"actions_content_blocks_start_symbol_actions_list": sdtsFnActionsContentBlocksStartSymbolActionsList,
		"actions_content_blocks_prepend":                   sdtsFnActionsContentBlocksPrepend,
		"tokens_content_blocks_prepend":                    sdtsFnTokensContentBlocksPrepend,
		"grammar_content_blocks_prepend":                   sdtsFnGrammarContentBlocksPrepend,
		"make_prod_action":                                 sdtsFnMakeProdAction,
		"make_symbol_actions":                              sdtsFnMakeSymbolActions,
		"make_grammar_content_node":                        sdtsFnMakeGrammarContentNode,
		"make_actions_content_node":                        sdtsFnMakeActionsContentNode,
		"make_tokens_content_node":                         sdtsFnMakeTokensContentNode,
		"trim_string":                                      sdtsFnTrimString,
		"make_discard_option":                              sdtsFnMakeDiscardOption,
		"make_stateshift_option":                           sdtsFnMakeStateshiftOption,
		"make_human_option":                                sdtsFnMakeHumanOption,
		"make_token_option":                                sdtsFnMakeTokenOption,
		"make_priority_option":                             sdtsFnMakePriorityOption,
		"identity":                                         sdtsFnIdentity,
		"interpret_escape":                                 sdtsFnInterpretEscape,
		"append_strings":                                   sdtsFnAppendStrings,
		"append_strings_trimmed":                           sdtsFnAppendStringsTrimmed,
		"get_nonterminal":                                  sdtsFnGetNonterminal,
		"get_int":                                          sdtsFnGetInt,
		"get_terminal":                                     sdtsFnGetTerminal,
		"rule_list_append":                                 sdtsFnRuleListAppend,
		"entry_list_append":                                sdtsFnEntryListAppend,
		"actions_state_block_list_append":                  sdtsFnActionsStateBlockListAppend,
		"tokens_state_block_list_append":                   sdtsFnTokensStateBlockListAppend,
		"grammar_state_block_list_append":                  sdtsFnGrammarStateBlockListAppend,
		"symbol_actions_list_append":                       sdtsFnSymbolActionsListAppend,
		"prod_action_list_append":                          sdtsFnProdActionListAppend,
		"semantic_action_list_append":                      sdtsFnSemanticActionListAppend,
		"attr_ref_list_append":                             sdtsFnAttrRefListAppend,
		"attr_ref_list_start":                              sdtsFnAttrRefListStart,
		"get_attr_ref":                                     sdtsFnGetAttrRef,
		"make_semantic_action":                             sdtsFnMakeSemanticAction,
		"make_prod_specifier_next":                         sdtsFnMakeProdSpecifierNext,
		"make_prod_specifier_index":                        sdtsFnMakeProdSpecifierIndex,
		"make_prod_specifier_literal":                      sdtsFnMakeProdSpecifierLiteral,
		"prod_action_list_start":                           sdtsFnProdActionListStart,
		"semantic_action_list_start":                       sdtsFnSemanticActionListStart,
		"rule_list_start":                                  sdtsFnRuleListStart,
		"grammar_state_block_list_start":                   sdtsFnGrammarStateBlockListStart,
		"tokens_state_block_list_start":                    sdtsFnTokensStateBlockListStart,
		"actions_state_block_list_start":                   sdtsFnActionsStateBlockListStart,
		"symbol_actions_list_start":                        sdtsFnSymbolActionsListStart,
		"entry_list_start":                                 sdtsFnEntryListStart,
		"string_list_append":                               sdtsFnStringListAppend,
		"token_opt_list_start":                             sdtsFnTokenOptListStart,
		"token_opt_list_append":                            sdtsFnTokenOptListAppend,
		"string_list_start":                                sdtsFnStringListStart,
		"string_list_list_start":                           sdtsFnStringListListStart,
		"string_list_list_append":                          sdtsFnStringListListAppend,
		"epsilon_string_list":                              sdtsFnEpsilonStringList,
		"make_rule":                                        sdtsFnMakeRule,
		"make_token_entry":                                 sdtsFnMakeTokenEntry,
	}
)

func SDDErrMsg(msg string, a ...interface{}) string {
	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	return fmt.Sprintf(ErrWithMessageString, msg)
}

func sdtsFnMakeFishispec(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astBlock)
	if !ok {
		return AST{}
	}

	return AST{nodes: list}
}

func sdtsFnBlockListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnBlockListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(astBlock)
	if !ok {
		var errBl astErrorBlock
		toAppend = errBl
	}

	return []astBlock{toAppend}
}

func sdtsFnMakeGrammarBlock(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	return astGrammarBlock{content: list}
}

func sdtsFnMakeTokensBlock(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	return astTokensBlock{content: list}
}

func sdtsFnMakeActionsBlock(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astActionsContent)
	if !ok {
		list = []astActionsContent{{state: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	return astActionsBlock{content: list}
}

func sdtsFnGrammarContentBlocksStartRuleList(_ translation.SetterInfo, args []interface{}) interface{} {
	rules, ok := args[0].([]astGrammarRule)
	if !ok {
		rules = []astGrammarRule{{rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule list: first argument is not a rule list")}}}
	}
	toAppend := astGrammarContent{
		rules: rules,
		state: "",
	}

	return []astGrammarContent{toAppend}
}

func sdtsFnTokensContentBlocksStartEntryList(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnActionsContentBlocksStartSymbolActionsList(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnActionsContentBlocksPrepend(_ translation.SetterInfo, args []interface{}) interface{} {
	// state blocks
	list, ok := args[0].([]astActionsContent)
	if !ok {
		list = []astActionsContent{{state: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	// stateless block
	actions, ok := args[1].([]symbolActions)
	if !ok {
		actions = []symbolActions{{symbol: SDDErrMsg("producing this actions content list: second argument is not a symbol actions list")}}
	}
	toAppend := astActionsContent{
		actions: actions,
		state:   "",
	}

	list = append([]astActionsContent{toAppend}, list...)

	return list
}

func sdtsFnTokensContentBlocksPrepend(_ translation.SetterInfo, args []interface{}) interface{} {
	// state blocks
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	// stateless block
	tokens, ok := args[1].([]tokenEntry)
	if !ok {
		tokens = []tokenEntry{{pattern: SDDErrMsg("producing this tokens content list: second argument is not a token entry list")}}
	}
	toAppend := astTokensContent{
		entries: tokens,
		state:   "",
	}

	list = append([]astTokensContent{toAppend}, list...)

	return list
}

func sdtsFnGrammarContentBlocksPrepend(_ translation.SetterInfo, args []interface{}) interface{} {
	// state blocks
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	// stateless block
	rules, ok := args[1].([]astGrammarRule)
	if !ok {
		rules = []astGrammarRule{{rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this grammar content list: second argument is not a grammar rule list")}}}
	}
	toAppend := astGrammarContent{
		rules: rules,
		state: "",
	}

	list = append([]astGrammarContent{toAppend}, list...)

	return list
}

func sdtsFnMakeProdAction(_ translation.SetterInfo, args []interface{}) interface{} {
	prodSpec, ok := args[0].(box.Pair[string, interface{}])
	if !ok {
		prodSpec = box.Pair[string, interface{}]{First: "LITERAL", Second: []string{SDDErrMsg("producing this production action: first argument is not a pair of string, any")}}
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

func sdtsFnMakeSymbolActions(_ translation.SetterInfo, args []interface{}) interface{} {
	nonTermUntyped := sdtsFnGetNonterminal(translation.SetterInfo{}, args[0:1])
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

func sdtsFnMakeGrammarContentNode(_ translation.SetterInfo, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		state = SDDErrMsg("STATE value is not a string")
	}
	rules, ok := args[1].([]astGrammarRule)
	if !ok {
		rules = []astGrammarRule{{rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule list: second argument is not a rule list")}}}
	}
	return astGrammarContent{rules: rules, state: state}
}

func sdtsFnMakeActionsContentNode(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnMakeTokensContentNode(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnTrimString(_ translation.SetterInfo, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}
	return strings.TrimSpace(str)
}

func sdtsFnMakeDiscardOption(info translation.SetterInfo, args []interface{}) interface{} {
	return astTokenOption{optType: tokenOptDiscard, tok: info.FirstToken}
}

func sdtsFnMakeStateshiftOption(info translation.SetterInfo, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptStateshift, value: state, tok: info.FirstToken}
}

func sdtsFnMakeHumanOption(info translation.SetterInfo, args []interface{}) interface{} {
	human, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptHuman, value: human, tok: info.FirstToken}
}

func sdtsFnMakeTokenOption(info translation.SetterInfo, args []interface{}) interface{} {
	t, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptToken, value: t, tok: info.FirstToken}
}

func sdtsFnMakePriorityOption(info translation.SetterInfo, args []interface{}) interface{} {
	priority, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return astTokenOption{optType: tokenOptPriority, value: priority, tok: info.FirstToken}
}

func sdtsFnIdentity(_ translation.SetterInfo, args []interface{}) interface{} { return args[0] }

func sdtsFnInterpretEscape(_ translation.SetterInfo, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("escape sequence $text is not a string")
	}

	str = strings.TrimLeftFunc(str, unicode.IsSpace) // lets us handle startLineEscseq as well, glub!
	if len(str) < len("%!") {
		return SDDErrMsg("escape sequence $text does not appear to have enough characters: %q", str)
	}

	// escape sequence is %!, so just take the chars after that
	return str[len("%!"):]
}

func sdtsFnAppendStrings(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnAppendStringsTrimmed(_ translation.SetterInfo, args []interface{}) interface{} {
	str1, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("first argument is not a string")
	}
	str2, ok := args[1].(string)
	if !ok {
		return SDDErrMsg("second argument is not a string")
	}

	return strings.TrimSpace(str1 + str2)
}

func sdtsFnGetNonterminal(_ translation.SetterInfo, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return strings.TrimSpace(str)
}

func sdtsFnGetInt(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnGetTerminal(_ translation.SetterInfo, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return strings.ToLower(str)
}

func sdtsFnRuleListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarRule)
	if !ok {
		list = []astGrammarRule{{rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule list: first argument is not a rule list")}}}
	}

	toAppend, ok := args[1].(astGrammarRule)
	if !ok {
		toAppend = astGrammarRule{rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: second argument is not a rule")}}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnEntryListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnActionsStateBlockListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astActionsContent)
	if !ok {
		list = []astActionsContent{{state: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	toAppend, ok := args[1].(astActionsContent)
	if !ok {
		toAppend = astActionsContent{state: SDDErrMsg("producing this actions content list: second argument is not an actions content")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnTokensStateBlockListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	toAppend, ok := args[1].(astTokensContent)
	if !ok {
		toAppend = astTokensContent{state: SDDErrMsg("producing this tokens content list: second argument is not a tokens content")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnGrammarStateBlockListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	toAppend, ok := args[1].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{state: SDDErrMsg("producing this grammar content list: second argument is not a grammar content")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnSymbolActionsListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnProdActionListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnSemanticActionListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnAttrRefListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]AttrRef)
	if !ok {
		list = []AttrRef{{symbol: SDDErrMsg("producing this AttrRef list: first argument is not an AttrRef list")}}
	}

	toAppend := sdtsFnGetAttrRef(translation.SetterInfo{}, args[1:]).(AttrRef)

	list = append(list, toAppend)
	return list
}

func sdtsFnAttrRefListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend := sdtsFnGetAttrRef(translation.SetterInfo{}, args[0:]).(AttrRef)

	return []AttrRef{toAppend}
}

func sdtsFnGetAttrRef(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnMakeSemanticAction(_ translation.SetterInfo, args []interface{}) interface{} {
	attrRef := sdtsFnGetAttrRef(translation.SetterInfo{}, args[0:1]).(AttrRef)

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

func sdtsFnMakeProdSpecifierNext(_ translation.SetterInfo, args []interface{}) interface{} {
	// need exact generic-filled type to match later expectations.
	spec := box.Pair[string, interface{}]{First: "NEXT", Second: ""}
	return spec
}

func sdtsFnMakeProdSpecifierIndex(_ translation.SetterInfo, args []interface{}) interface{} {
	index := sdtsFnGetInt(translation.SetterInfo{}, args)
	// need exact generic-filled type to match later expectations.
	spec := box.Pair[string, interface{}]{First: "INDEX", Second: index}
	return spec
}

func sdtsFnMakeProdSpecifierLiteral(_ translation.SetterInfo, args []interface{}) interface{} {
	prod, ok := args[0].([]string)

	if !ok {
		prod = []string{SDDErrMsg("producing this production specifier: first argument is not an action production")}
	}

	// need exact generic-filled type to match later expectations.
	spec := box.Pair[string, interface{}]{First: "LITERAL", Second: prod}
	return spec
}

func sdtsFnProdActionListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(productionAction)
	if !ok {
		toAppend = productionAction{prodLiteral: []string{SDDErrMsg("producing this production action list: first argument is not a production action")}}
	}

	return []productionAction{toAppend}
}

func sdtsFnSemanticActionListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(semanticAction)
	if !ok {
		toAppend = semanticAction{hook: SDDErrMsg("producing this semantic action list: first argument is not a semantic actions")}
	}

	return []semanticAction{toAppend}
}

func sdtsFnRuleListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(astGrammarRule)
	if !ok {
		toAppend = astGrammarRule{rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: first argument is not a rule")}}
	}

	return []astGrammarRule{toAppend}
}

func sdtsFnGrammarStateBlockListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content")}
	}

	return []astGrammarContent{toAppend}
}

func sdtsFnTokensStateBlockListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(astTokensContent)
	if !ok {
		toAppend = astTokensContent{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content")}
	}

	return []astTokensContent{toAppend}
}

func sdtsFnActionsStateBlockListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(astActionsContent)
	if !ok {
		toAppend = astActionsContent{state: SDDErrMsg("producing this actions content list: first argument is not an actions content")}
	}

	return []astActionsContent{toAppend}
}

func sdtsFnSymbolActionsListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(symbolActions)
	if !ok {
		toAppend = symbolActions{symbol: SDDErrMsg("producing this symbol action: first argument is not a rule")}
	}

	return []symbolActions{toAppend}
}

func sdtsFnEntryListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(tokenEntry)
	if !ok {
		toAppend = tokenEntry{pattern: SDDErrMsg("producing this token entry: first argument is not a token entry")}
	}

	return []tokenEntry{toAppend}
}

func sdtsFnStringListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnTokenOptListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(astTokenOption)
	if !ok {
		toAppend = astTokenOption{value: SDDErrMsg("first argument is not a token option")}
	}

	return []astTokenOption{toAppend}
}

func sdtsFnTokenOptListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnStringListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(string)
	if !ok {
		toAppend = ErrString
	}

	return []string{toAppend}
}

func sdtsFnStringListListStart(_ translation.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].([]string)
	if !ok {
		toAppend = []string{SDDErrMsg("producing this string list: first argument is not a string list")}
	}

	return [][]string{toAppend}
}

func sdtsFnStringListListAppend(_ translation.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnEpsilonStringList(_ translation.SetterInfo, args []interface{}) interface{} {
	strList := grammar.Epsilon.Copy()
	return []string(strList)
}

func sdtsFnMakeRule(info translation.SetterInfo, args []interface{}) interface{} {
	ntInterface := sdtsFnGetNonterminal(translation.SetterInfo{}, args[0:1])

	nt, ok := ntInterface.(string)
	if !ok {
		nt = SDDErrMsg("first argument is not a string")
	}

	productions, ok := args[1].([][]string)
	if !ok {
		productions = [][]string{{SDDErrMsg("producing this list of lists of strings: second argument is not a [][]string")}}
	}

	gr := grammar.Rule{NonTerminal: nt, Productions: []grammar.Production{}}

	for _, p := range productions {
		gr.Productions = append(gr.Productions, p)
	}

	r := astGrammarRule{
		rule: gr,
		tok:  info.FirstToken,
	}

	return r
}

func sdtsFnMakeTokenEntry(info translation.SetterInfo, args []interface{}) interface{} {
	pattern, ok := args[0].(string)
	if !ok {
		pattern = SDDErrMsg("first argument (pattern) is not a string")
	}

	tokenOpts, ok := args[1].([]astTokenOption)
	if !ok {
		tokenOpts = []astTokenOption{{value: SDDErrMsg("producing this token option list: second argument (tokenOpts) is not a token option list")}}
	}

	t := tokenEntry{pattern: pattern, tok: info.FirstToken}

	for _, opt := range tokenOpts {
		switch opt.optType {
		case tokenOptDiscard:
			t.discard = true
			t.discardTok = append(t.discardTok, opt.tok)
		case tokenOptHuman:
			t.human = opt.value
			t.humanTok = append(t.humanTok, opt.tok)
		case tokenOptPriority:
			prior, err := strconv.Atoi(opt.value)
			if err == nil {
				t.priority = prior
			}
			t.priorityTok = append(t.priorityTok, opt.tok)
		case tokenOptStateshift:
			t.shift = opt.value
			t.shiftTok = append(t.shiftTok, opt.tok)
		case tokenOptToken:
			t.token = opt.value
			t.tokenTok = append(t.tokenTok, opt.tok)
		}
	}
	return t
}
