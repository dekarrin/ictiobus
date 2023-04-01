package syntax

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dekarrin/ictiobus/fishi/fe"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/ictiobus/types"
)

const (
	ErrString            = "<ERR>"
	ErrWithMessageString = "<ERR: %s>"
)

var (
	HooksTable = map[string]trans.AttributeSetter{
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
		"make_state_ins":                                   sdtsFnMakeStateIns,
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

func sdtsFnMakeFishispec(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTBlock)
	if !ok {
		// can't directly return nil because we'd lose the type information
		var ast []ASTBlock
		return ast
	}

	return list
}

func sdtsFnBlockListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTBlock)
	if !ok {
		return []ASTBlock{}
	}

	toAppend, ok := args[1].(ASTBlock)
	if !ok {
		var errBl ASTErrorBlock
		toAppend = errBl
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnBlockListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTBlock)
	if !ok {
		var errBl ASTErrorBlock
		toAppend = errBl
	}

	return []ASTBlock{toAppend}
}

func sdtsFnMakeGrammarBlock(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTGrammarContent)
	if !ok {
		list = []ASTGrammarContent{{State: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	return ASTGrammarBlock{Content: list}
}

func sdtsFnMakeTokensBlock(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTTokensContent)
	if !ok {
		list = []ASTTokensContent{{State: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	return ASTTokensBlock{Content: list}
}

func sdtsFnMakeActionsBlock(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTActionsContent)
	if !ok {
		list = []ASTActionsContent{{State: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	return ASTActionsBlock{Content: list}
}

func sdtsFnGrammarContentBlocksStartRuleList(_ trans.SetterInfo, args []interface{}) interface{} {
	rules, ok := args[0].([]ASTGrammarRule)
	if !ok {
		rules = []ASTGrammarRule{{Rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule list: first argument is not a rule list")}}}
	}
	toAppend := ASTGrammarContent{
		Rules: rules,
		State: "",
	}

	return []ASTGrammarContent{toAppend}
}

func sdtsFnTokensContentBlocksStartEntryList(_ trans.SetterInfo, args []interface{}) interface{} {
	entries, ok := args[0].([]ASTTokenEntry)
	if !ok {
		entries = []ASTTokenEntry{{Pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}
	toAppend := ASTTokensContent{
		Entries: entries,
		State:   "",
	}

	return []ASTTokensContent{toAppend}
}

func sdtsFnActionsContentBlocksStartSymbolActionsList(_ trans.SetterInfo, args []interface{}) interface{} {
	actions, ok := args[0].([]ASTSymbolActions)
	if !ok {
		actions = []ASTSymbolActions{{Symbol: SDDErrMsg("producing this symbol actions list: first argument is not a symbol actions list")}}
	}
	toAppend := ASTActionsContent{
		Actions: actions,
		State:   "",
	}

	return []ASTActionsContent{toAppend}
}

func sdtsFnActionsContentBlocksPrepend(_ trans.SetterInfo, args []interface{}) interface{} {
	// state blocks
	list, ok := args[0].([]ASTActionsContent)
	if !ok {
		list = []ASTActionsContent{{State: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	// stateless block
	actions, ok := args[1].([]ASTSymbolActions)
	if !ok {
		actions = []ASTSymbolActions{{Symbol: SDDErrMsg("producing this actions content list: second argument is not a symbol actions list")}}
	}
	toAppend := ASTActionsContent{
		Actions: actions,
		State:   "",
	}

	list = append([]ASTActionsContent{toAppend}, list...)

	return list
}

func sdtsFnTokensContentBlocksPrepend(_ trans.SetterInfo, args []interface{}) interface{} {
	// state blocks
	list, ok := args[0].([]ASTTokensContent)
	if !ok {
		list = []ASTTokensContent{{State: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	// stateless block
	tokens, ok := args[1].([]ASTTokenEntry)
	if !ok {
		tokens = []ASTTokenEntry{{Pattern: SDDErrMsg("producing this tokens content list: second argument is not a token entry list")}}
	}
	toAppend := ASTTokensContent{
		Entries: tokens,
		State:   "",
	}

	list = append([]ASTTokensContent{toAppend}, list...)

	return list
}

func sdtsFnGrammarContentBlocksPrepend(_ trans.SetterInfo, args []interface{}) interface{} {
	// state blocks
	list, ok := args[0].([]ASTGrammarContent)
	if !ok {
		list = []ASTGrammarContent{{State: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	// stateless block
	rules, ok := args[1].([]ASTGrammarRule)
	if !ok {
		rules = []ASTGrammarRule{{Rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this grammar content list: second argument is not a grammar rule list")}}}
	}
	toAppend := ASTGrammarContent{
		Rules: rules,
		State: "",
	}

	list = append([]ASTGrammarContent{toAppend}, list...)

	return list
}

func sdtsFnMakeProdAction(info trans.SetterInfo, args []interface{}) interface{} {
	prodSpec, ok := args[0].(box.Triple[string, interface{}, types.Token])
	if !ok {
		prodSpec = box.Triple[string, interface{}, types.Token]{
			First:  "LITERAL",
			Second: []string{SDDErrMsg("producing this production action: first argument is not a pair of string, any")},
		}
	}

	semActions, ok := args[1].([]ASTSemanticAction)
	if !ok {
		semActions = []ASTSemanticAction{{Hook: SDDErrMsg("producing this production action: second argument is not a semantic action list")}}
	}

	pa := ASTProductionAction{
		Actions: semActions,
		Src:     info.FirstToken,
		SrcVal:  prodSpec.Third,
	}

	if prodSpec.First == "LITERAL" {
		pa.ProdLiteral = prodSpec.Second.([]string)
	} else if prodSpec.First == "INDEX" {
		pa.ProdIndex = prodSpec.Second.(int)
	} else if prodSpec.First == "NEXT" {
		pa.ProdNext = true
	} else {
		pa.ProdLiteral = []string{SDDErrMsg("producing this production action: first argument is not a pair of string/interface{}")}
	}

	return pa
}

func sdtsFnMakeSymbolActions(info trans.SetterInfo, args []interface{}) interface{} {
	nonTermUntyped := sdtsFnGetNonterminal(trans.SetterInfo{}, args[0:1])
	nonTerm := nonTermUntyped.(string)

	// also grab the nonTerm's token from args
	ntTok, ok := args[1].(types.Token)
	if !ok {
		ntTok = lex.NewToken(
			types.TokenError,
			SDDErrMsg("producing this symbol actions: second argument is not a token"),
			0, 0, "",
		)
	}

	prodActions, ok := args[2].([]ASTProductionAction)
	if !ok {
		prodActions = []ASTProductionAction{{ProdLiteral: []string{SDDErrMsg("producing this production action list: third argument is not a production action list")}}}
	}

	sa := ASTSymbolActions{
		Symbol:  nonTerm,
		Actions: prodActions,

		Src:    info.FirstToken,
		SrcSym: ntTok,
	}

	return sa
}

func sdtsFnMakeStateIns(info trans.SetterInfo, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		state = SDDErrMsg("state ID is not a string")
	}

	// also grab the state ID's token from args
	stateTok, ok := args[1].(types.Token)
	if !ok {
		stateTok = lex.NewToken(
			types.TokenError,
			SDDErrMsg("producing this state ID: second argument is not a token"),
			0, 0, "",
		)
	}

	return box.Pair[string, types.Token]{First: state, Second: stateTok}
}

func sdtsFnMakeGrammarContentNode(info trans.SetterInfo, args []interface{}) interface{} {
	state, ok := args[0].(box.Pair[string, types.Token])
	if !ok {
		state = box.Pair[string, types.Token]{First: SDDErrMsg("STATE value is not a string/token pair")}
	}

	rules, ok := args[1].([]ASTGrammarRule)
	if !ok {
		rules = []ASTGrammarRule{{Rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule list: second argument is not a rule list")}}}
	}
	return ASTGrammarContent{Rules: rules, State: state.First, SrcState: state.Second, Src: info.FirstToken}
}

func sdtsFnMakeActionsContentNode(info trans.SetterInfo, args []interface{}) interface{} {
	state, ok := args[0].(box.Pair[string, types.Token])
	if !ok {
		state = box.Pair[string, types.Token]{First: SDDErrMsg("STATE value is not a string")}
	}
	actions, ok := args[1].([]ASTSymbolActions)
	if !ok {
		actions = []ASTSymbolActions{{Symbol: SDDErrMsg("producing this symbol actions list: second argument is not a symbol actions list")}}
	}
	return ASTActionsContent{Actions: actions, State: state.First, SrcState: state.Second, Src: info.FirstToken}
}

func sdtsFnMakeTokensContentNode(info trans.SetterInfo, args []interface{}) interface{} {
	state, ok := args[0].(box.Pair[string, types.Token])
	if !ok {
		state = box.Pair[string, types.Token]{First: SDDErrMsg("STATE value is not a string")}
	}

	entries, ok := args[1].([]ASTTokenEntry)
	if !ok {
		entries = []ASTTokenEntry{{Pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}

	return ASTTokensContent{Entries: entries, State: state.First, SrcState: state.Second, Src: info.FirstToken}
}

func sdtsFnTrimString(_ trans.SetterInfo, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}
	return strings.TrimSpace(str)
}

func sdtsFnMakeDiscardOption(info trans.SetterInfo, args []interface{}) interface{} {
	return ASTTokenOption{Type: TokenOptDiscard, Src: info.FirstToken}
}

func sdtsFnMakeStateshiftOption(info trans.SetterInfo, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return ASTTokenOption{Type: TokenOptStateshift, Value: state, Src: info.FirstToken}
}

func sdtsFnMakeHumanOption(info trans.SetterInfo, args []interface{}) interface{} {
	human, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return ASTTokenOption{Type: TokenOptHuman, Value: human, Src: info.FirstToken}
}

func sdtsFnMakeTokenOption(info trans.SetterInfo, args []interface{}) interface{} {
	t, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return ASTTokenOption{Type: TokenOptToken, Value: t, Src: info.FirstToken}
}

func sdtsFnMakePriorityOption(info trans.SetterInfo, args []interface{}) interface{} {
	priority, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}

	return ASTTokenOption{Type: TokenOptPriority, Value: priority, Src: info.FirstToken}
}

func sdtsFnIdentity(_ trans.SetterInfo, args []interface{}) interface{} { return args[0] }

func sdtsFnInterpretEscape(_ trans.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnAppendStrings(_ trans.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnAppendStringsTrimmed(_ trans.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnGetNonterminal(_ trans.SetterInfo, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return strings.TrimSpace(str)
}

func sdtsFnGetInt(_ trans.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnGetTerminal(_ trans.SetterInfo, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return strings.ToLower(str)
}

func sdtsFnRuleListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTGrammarRule)
	if !ok {
		list = []ASTGrammarRule{{Rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule list: first argument is not a rule list")}}}
	}

	toAppend, ok := args[1].(ASTGrammarRule)
	if !ok {
		toAppend = ASTGrammarRule{Rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: second argument is not a rule")}}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnEntryListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTTokenEntry)
	if !ok {
		list = []ASTTokenEntry{{Pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}

	toAppend, ok := args[1].(ASTTokenEntry)
	if !ok {
		toAppend = ASTTokenEntry{Pattern: SDDErrMsg("producing this token entry: second argument is not a token entry")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnActionsStateBlockListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTActionsContent)
	if !ok {
		list = []ASTActionsContent{{State: SDDErrMsg("producing this actions content list: first argument is not an actions content list")}}
	}

	toAppend, ok := args[1].(ASTActionsContent)
	if !ok {
		toAppend = ASTActionsContent{State: SDDErrMsg("producing this actions content list: second argument is not an actions content")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnTokensStateBlockListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTTokensContent)
	if !ok {
		list = []ASTTokensContent{{State: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	toAppend, ok := args[1].(ASTTokensContent)
	if !ok {
		toAppend = ASTTokensContent{State: SDDErrMsg("producing this tokens content list: second argument is not a tokens content")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnGrammarStateBlockListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTGrammarContent)
	if !ok {
		list = []ASTGrammarContent{{State: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	toAppend, ok := args[1].(ASTGrammarContent)
	if !ok {
		toAppend = ASTGrammarContent{State: SDDErrMsg("producing this grammar content list: second argument is not a grammar content")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnSymbolActionsListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTSymbolActions)
	if !ok {
		list = []ASTSymbolActions{{Symbol: SDDErrMsg("producing this symbol actions list: first argument is not a symbol actions list")}}
	}

	toAppend, ok := args[1].(ASTSymbolActions)
	if !ok {
		toAppend = ASTSymbolActions{Symbol: SDDErrMsg("producing this symbol actions: second argument is not a symbol actions")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnProdActionListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTProductionAction)
	if !ok {
		list = []ASTProductionAction{{ProdLiteral: []string{SDDErrMsg("producing this production action list: first argument is not a production actions list")}}}
	}

	toAppend, ok := args[1].(ASTProductionAction)
	if !ok {
		toAppend = ASTProductionAction{ProdLiteral: []string{SDDErrMsg("producing this production action list: first argument is not a production action")}}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnSemanticActionListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTSemanticAction)
	if !ok {
		list = []ASTSemanticAction{{Hook: SDDErrMsg("producing this semantic action list: first argument is not a semantic actions list")}}
	}

	toAppend, ok := args[1].(ASTSemanticAction)
	if !ok {
		toAppend = ASTSemanticAction{Hook: SDDErrMsg("producing this semantic action list: first argument is not a semantic action")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnAttrRefListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTAttrRef)
	if !ok {
		list = []ASTAttrRef{{Symbol: SDDErrMsg("producing this AttrRef list: first argument is not an AttrRef list")}}
	}

	// get token of attr ref to build fake info object to pass to sdtsFnGetAttrRef's info.
	fakeInfo := makeFakeInfo(args[2], fe.TCAttrRef.ID(), "value")
	toAppend := sdtsFnGetAttrRef(fakeInfo, args[1:]).(ASTAttrRef)

	list = append(list, toAppend)
	return list
}

func sdtsFnAttrRefListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	// get token of attr ref to build fake info object to pass to sdtsFnGetAttrRef's info.
	fakeInfo := makeFakeInfo(args[1], fe.TCAttrRef.ID(), "value")
	toAppend := sdtsFnGetAttrRef(fakeInfo, args[0:]).(ASTAttrRef)

	return []ASTAttrRef{toAppend}
}

func sdtsFnGetAttrRef(info trans.SetterInfo, args []interface{}) interface{} {
	var attrRef ASTAttrRef

	str, ok := args[0].(string)
	if !ok {
		attrRef = ASTAttrRef{Symbol: SDDErrMsg("producing this semantic action: first argument is not a string to be parsed into an AttrRef")}
	} else {
		var err error
		attrRef, err = ParseAttrRef(str)
		if err != nil {
			attrRef = ASTAttrRef{Symbol: SDDErrMsg("producing this semantic action: first argument is not a valid AttrRef: %v", err.Error())}
		}
	}

	attrRef.Src = info.FirstToken

	return attrRef
}

func sdtsFnMakeSemanticAction(info trans.SetterInfo, args []interface{}) interface{} {
	fakeInfo := makeFakeInfo(args[1], fe.TCAttrRef.ID(), "value")
	attrRef := sdtsFnGetAttrRef(fakeInfo, args[0:1]).(ASTAttrRef)

	hookId, ok := args[2].(string)
	if !ok {
		hookId = SDDErrMsg("producing this semantic action: third argument is not a string")
	}

	hookTok, ok := args[3].(types.Token)
	if !ok {
		hookTok = lex.NewToken(
			types.TokenError,
			SDDErrMsg("producing this semantic action: argument is not a token"),
			0, 0, "",
		)
	}

	var argRefs []ASTAttrRef
	if len(args) > 4 {
		argRefs, ok = args[4].([]ASTAttrRef)
		if !ok {
			argRefs = []ASTAttrRef{{Symbol: SDDErrMsg("producing this semantic action: fifth argument is not an attrRef list")}}
		}
	}

	sa := ASTSemanticAction{
		LHS:     attrRef,
		Hook:    hookId,
		With:    argRefs,
		SrcHook: hookTok,
		Src:     info.FirstToken,
	}

	return sa
}

func sdtsFnMakeProdSpecifierNext(info trans.SetterInfo, args []interface{}) interface{} {
	// need exact generic-filled type to match later expectations.
	spec := box.Triple[string, interface{}, types.Token]{First: "NEXT", Second: "", Third: info.FirstToken}
	return spec
}

func sdtsFnMakeProdSpecifierIndex(info trans.SetterInfo, args []interface{}) interface{} {
	index := sdtsFnGetInt(trans.SetterInfo{}, args)
	// need exact generic-filled type to match later expectations.
	spec := box.Triple[string, interface{}, types.Token]{First: "INDEX", Second: index, Third: info.FirstToken}
	return spec
}

func sdtsFnMakeProdSpecifierLiteral(info trans.SetterInfo, args []interface{}) interface{} {
	prod, ok := args[0].([]string)

	if !ok {
		prod = []string{SDDErrMsg("producing this production specifier: first argument is not an action production")}
	}

	// need exact generic-filled type to match later expectations.
	spec := box.Triple[string, interface{}, types.Token]{First: "LITERAL", Second: prod, Third: info.FirstToken}
	return spec
}

func sdtsFnProdActionListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTProductionAction)
	if !ok {
		toAppend = ASTProductionAction{ProdLiteral: []string{SDDErrMsg("producing this production action list: first argument is not a production action")}}
	}

	return []ASTProductionAction{toAppend}
}

func sdtsFnSemanticActionListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTSemanticAction)
	if !ok {
		toAppend = ASTSemanticAction{Hook: SDDErrMsg("producing this semantic action list: first argument is not a semantic actions")}
	}

	return []ASTSemanticAction{toAppend}
}

func sdtsFnRuleListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTGrammarRule)
	if !ok {
		toAppend = ASTGrammarRule{Rule: grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: first argument is not a rule")}}
	}

	return []ASTGrammarRule{toAppend}
}

func sdtsFnGrammarStateBlockListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTGrammarContent)
	if !ok {
		toAppend = ASTGrammarContent{State: SDDErrMsg("producing this grammar content list: first argument is not a grammar content")}
	}

	return []ASTGrammarContent{toAppend}
}

func sdtsFnTokensStateBlockListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTTokensContent)
	if !ok {
		toAppend = ASTTokensContent{State: SDDErrMsg("producing this tokens content list: first argument is not a tokens content")}
	}

	return []ASTTokensContent{toAppend}
}

func sdtsFnActionsStateBlockListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTActionsContent)
	if !ok {
		toAppend = ASTActionsContent{State: SDDErrMsg("producing this actions content list: first argument is not an actions content")}
	}

	return []ASTActionsContent{toAppend}
}

func sdtsFnSymbolActionsListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTSymbolActions)
	if !ok {
		toAppend = ASTSymbolActions{Symbol: SDDErrMsg("producing this symbol action: first argument is not a rule")}
	}

	return []ASTSymbolActions{toAppend}
}

func sdtsFnEntryListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTTokenEntry)
	if !ok {
		toAppend = ASTTokenEntry{Pattern: SDDErrMsg("producing this token entry: first argument is not a token entry")}
	}

	return []ASTTokenEntry{toAppend}
}

func sdtsFnStringListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnTokenOptListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(ASTTokenOption)
	if !ok {
		toAppend = ASTTokenOption{Value: SDDErrMsg("first argument is not a token option")}
	}

	return []ASTTokenOption{toAppend}
}

func sdtsFnTokenOptListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
	list, ok := args[0].([]ASTTokenOption)
	if !ok {
		return []ASTTokenOption{{Value: SDDErrMsg("producing this token option list: first argument is not a token option list")}}
	}

	toAppend, ok := args[1].(ASTTokenOption)
	if !ok {
		toAppend = ASTTokenOption{Value: SDDErrMsg("producing this token option: second argument is not a token option")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnStringListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].(string)
	if !ok {
		toAppend = ErrString
	}

	return []string{toAppend}
}

func sdtsFnStringListListStart(_ trans.SetterInfo, args []interface{}) interface{} {
	toAppend, ok := args[0].([]string)
	if !ok {
		toAppend = []string{SDDErrMsg("producing this string list: first argument is not a string list")}
	}

	return [][]string{toAppend}
}

func sdtsFnStringListListAppend(_ trans.SetterInfo, args []interface{}) interface{} {
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

func sdtsFnEpsilonStringList(_ trans.SetterInfo, args []interface{}) interface{} {
	strList := grammar.Epsilon.Copy()
	return []string(strList)
}

func sdtsFnMakeRule(info trans.SetterInfo, args []interface{}) interface{} {
	ntInterface := sdtsFnGetNonterminal(trans.SetterInfo{}, args[0:1])

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

	r := ASTGrammarRule{
		Rule: gr,
		Src:  info.FirstToken,
	}

	return r
}

func sdtsFnMakeTokenEntry(info trans.SetterInfo, args []interface{}) interface{} {
	pattern, ok := args[0].(string)
	if !ok {
		pattern = SDDErrMsg("first argument (pattern) is not a string")
	}

	tokenOpts, ok := args[1].([]ASTTokenOption)
	if !ok {
		tokenOpts = []ASTTokenOption{{Value: SDDErrMsg("producing this token option list: second argument (tokenOpts) is not a token option list")}}
	}

	t := ASTTokenEntry{Pattern: pattern, Src: info.FirstToken}

	for _, opt := range tokenOpts {
		switch opt.Type {
		case TokenOptDiscard:
			t.Discard = true
			t.SrcDiscard = append(t.SrcDiscard, opt.Src)
		case TokenOptHuman:
			t.Human = opt.Value
			t.SrcHuman = append(t.SrcHuman, opt.Src)
		case TokenOptPriority:
			prior, err := strconv.Atoi(opt.Value)
			if err == nil {
				t.Priority = prior
			}
			t.SrcPriority = append(t.SrcPriority, opt.Src)
		case TokenOptStateshift:
			t.Shift = opt.Value
			t.SrcShift = append(t.SrcShift, opt.Src)
		case TokenOptToken:
			t.Token = opt.Value
			t.SrcToken = append(t.SrcToken, opt.Src)
		}
	}
	return t
}

// for hooks to generate fake info object when needed. Sym and name can be blank
// if desired. Returned SetterInfo will always have synthetic set to true.
func makeFakeInfo(from interface{}, sym, name string) trans.SetterInfo {
	tok, ok := from.(types.Token)
	if !ok {
		tok = lex.NewToken(
			types.TokenError,
			SDDErrMsg("argument is not a token"),
			0, 0, "",
		)
	}

	info := trans.SetterInfo{
		GrammarSymbol: sym,
		Synthetic:     true,
		Name:          name,
		FirstToken:    tok,
	}

	return info
}
