package syntax

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dekarrin/ictiobus/fishi/fe/fetoken"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
)

var (
	// HooksTable contains all bindings of STDS hook names to their
	// implementation functions. It is passed to the compiler frontend
	// automatically on creation and is used for translating parse trees
	// returned by the FISHI parser into an [AST].
	HooksTable = trans.HookMap{
		"make_fishispec":                           sdtsFnMakeFishispec,
		"block_list_append":                        sdtsFnBlockListAppend,
		"block_list_start":                         sdtsFnBlockListStart,
		"make_gblock":                              sdtsFnMakeGrammarBlock,
		"make_tblock":                              sdtsFnMakeTokensBlock,
		"make_ablock":                              sdtsFnMakeActionsBlock,
		"grammar_content_blocks_start_rule_list":   sdtsFnGrammarContentBlocksStartRuleList,
		"tokens_content_blocks_start_entry_list":   sdtsFnTokensContentBlocksStartEntryList,
		"actions_content_blocks_start_sym_actions": sdtsFnActionsContentBlocksStartSymbolActionsList,
		"actions_content_blocks_prepend":           sdtsFnActionsContentBlocksPrepend,
		"tokens_content_blocks_prepend":            sdtsFnTokensContentBlocksPrepend,
		"grammar_content_blocks_prepend":           sdtsFnGrammarContentBlocksPrepend,
		"make_prod_action":                         sdtsFnMakeProdAction,
		"make_symbol_actions":                      sdtsFnMakeSymbolActions,
		"make_state_ins":                           sdtsFnMakeStateIns,
		"make_grammar_content_node":                sdtsFnMakeGrammarContentNode,
		"make_actions_content_node":                sdtsFnMakeActionsContentNode,
		"make_tokens_content_node":                 sdtsFnMakeTokensContentNode,
		"trim_string":                              sdtsFnTrimString,
		"make_discard_option":                      sdtsFnMakeDiscardOption,
		"make_stateshift_option":                   sdtsFnMakeStateshiftOption,
		"make_human_option":                        sdtsFnMakeHumanOption,
		"make_token_option":                        sdtsFnMakeTokenOption,
		"make_priority_option":                     sdtsFnMakePriorityOption,
		"ident":                                    sdtsFnIdentity,
		"interpret_escape":                         sdtsFnInterpretEscape,
		"append_strings":                           sdtsFnAppendStrings,
		"append_strings_trimmed":                   sdtsFnAppendStringsTrimmed,
		"get_nonterminal":                          sdtsFnGetNonterminal,
		"get_int":                                  sdtsFnGetInt,
		"get_terminal":                             sdtsFnGetTerminal,
		"rule_list_append":                         sdtsFnRuleListAppend,
		"entry_list_append":                        sdtsFnEntryListAppend,
		"actions_state_block_list_append":          sdtsFnActionsStateBlockListAppend,
		"tokens_state_block_list_append":           sdtsFnTokensStateBlockListAppend,
		"grammar_state_block_list_append":          sdtsFnGrammarStateBlockListAppend,
		"symbol_actions_list_append":               sdtsFnSymbolActionsListAppend,
		"prod_action_list_append":                  sdtsFnProdActionListAppend,
		"semantic_action_list_append":              sdtsFnSemanticActionListAppend,
		"attr_ref_list_append":                     sdtsFnAttrRefListAppend,
		"attr_ref_list_start":                      sdtsFnAttrRefListStart,
		"get_attr_ref":                             sdtsFnGetAttrRef,
		"make_semantic_action":                     sdtsFnMakeSemanticAction,
		"make_prod_specifier_next":                 sdtsFnMakeProdSpecifierNext,
		"make_prod_specifier_index":                sdtsFnMakeProdSpecifierIndex,
		"make_prod_specifier_literal":              sdtsFnMakeProdSpecifierLiteral,
		"prod_action_list_start":                   sdtsFnProdActionListStart,
		"semantic_action_list_start":               sdtsFnSemanticActionListStart,
		"rule_list_start":                          sdtsFnRuleListStart,
		"grammar_state_block_list_start":           sdtsFnGrammarStateBlockListStart,
		"tokens_state_block_list_start":            sdtsFnTokensStateBlockListStart,
		"actions_state_block_list_start":           sdtsFnActionsStateBlockListStart,
		"symbol_actions_list_start":                sdtsFnSymbolActionsListStart,
		"entry_list_start":                         sdtsFnEntryListStart,
		"string_list_append":                       sdtsFnStringListAppend,
		"token_opt_list_start":                     sdtsFnTokenOptListStart,
		"token_opt_list_append":                    sdtsFnTokenOptListAppend,
		"string_list_start":                        sdtsFnStringListStart,
		"string_list_list_start":                   sdtsFnStringListListStart,
		"string_list_list_append":                  sdtsFnStringListListAppend,
		"epsilon_string_list":                      sdtsFnEpsilonStringList,
		"make_rule":                                sdtsFnMakeRule,
		"make_token_entry":                         sdtsFnMakeTokenEntry,
	}
)

func sdtsFnMakeFishispec(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]Block)
	if !ok {
		// can't directly return nil because we'd lose the type information
		return nil, newArgTypeError(args, 0, "[]Block")
	}

	return AST{Nodes: list}, nil
}

func sdtsFnBlockListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]Block)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]Block")
	}

	toAppend, ok := args[1].(Block)
	if !ok {
		return nil, newArgTypeError(args, 1, "Block")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnBlockListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(Block)
	if !ok {
		return nil, newArgTypeError(args, 0, "Block")
	}

	return []Block{toAppend}, nil
}

func sdtsFnMakeGrammarBlock(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]GrammarContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]GrammarContent")
	}

	return GrammarBlock{Content: list}, nil
}

func sdtsFnMakeTokensBlock(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]TokensContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]TokensContent")
	}

	return TokensBlock{Content: list}, nil
}

func sdtsFnMakeActionsBlock(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]ActionsContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]ActionsContent")
	}

	return ActionsBlock{Content: list}, nil
}

func sdtsFnGrammarContentBlocksStartRuleList(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	rules, ok := args[0].([]GrammarRule)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]GrammarRule")
	}
	toAppend := GrammarContent{
		Rules: rules,
		State: "",
	}

	return []GrammarContent{toAppend}, nil
}

func sdtsFnTokensContentBlocksStartEntryList(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	entries, ok := args[0].([]TokenEntry)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]TokenEntry")
	}
	toAppend := TokensContent{
		Entries: entries,
		State:   "",
	}

	return []TokensContent{toAppend}, nil
}

func sdtsFnActionsContentBlocksStartSymbolActionsList(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	actions, ok := args[0].([]SymbolActions)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]SymbolActions")
	}
	toAppend := ActionsContent{
		Actions: actions,
		State:   "",
	}

	return []ActionsContent{toAppend}, nil
}

func sdtsFnActionsContentBlocksPrepend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	// state blocks
	list, ok := args[0].([]ActionsContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]ActionsContent")
	}

	// stateless block
	actions, ok := args[1].([]SymbolActions)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]SymbolActions")
	}
	toAppend := ActionsContent{
		Actions: actions,
		State:   "",
	}

	list = append([]ActionsContent{toAppend}, list...)

	return list, nil
}

func sdtsFnTokensContentBlocksPrepend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	// state blocks
	list, ok := args[0].([]TokensContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]TokensContent")
	}

	// stateless block
	tokens, ok := args[1].([]TokenEntry)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]TokenEntry")
	}
	toAppend := TokensContent{
		Entries: tokens,
		State:   "",
	}

	list = append([]TokensContent{toAppend}, list...)

	return list, nil
}

func sdtsFnGrammarContentBlocksPrepend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	// state blocks
	list, ok := args[0].([]GrammarContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]GrammarContent")
	}

	// stateless block
	rules, ok := args[1].([]GrammarRule)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]GrammarRule")
	}
	toAppend := GrammarContent{
		Rules: rules,
		State: "",
	}

	list = append([]GrammarContent{toAppend}, list...)

	return list, nil
}

func sdtsFnMakeProdAction(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	prodSpec, ok := args[0].(box.Triple[string, interface{}, lex.Token])
	if !ok {
		return nil, newArgTypeError(args, 0, "box.Triple[string, interface{}, lex.Token]")
	}

	semActions, ok := args[1].([]SemanticAction)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]SemanticAction")
	}

	pa := ProductionAction{
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
		return nil, newArgError(args, 0, "unknown spec type %q; must be one of \"LITERAL\", \"INDEX\", or \"NEXT\"", prodSpec.First)
	}

	return pa, nil
}

func sdtsFnMakeSymbolActions(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	nonTermUntyped, err := sdtsFnGetNonterminal(trans.SetterInfo{}, args[0:1])
	if err != nil {
		return nil, err
	}
	nonTerm := nonTermUntyped.(string)

	// also grab the nonTerm's token from args
	ntTok, ok := args[1].(lex.Token)
	if !ok {
		return nil, newArgTypeError(args, 1, "lex.Token")
	}

	prodActions, ok := args[2].([]ProductionAction)
	if !ok {
		return nil, newArgTypeError(args, 2, "[]ProductionAction")
	}

	sa := SymbolActions{
		Symbol:  nonTerm,
		Actions: prodActions,

		Src:    info.FirstToken,
		SrcSym: ntTok,
	}

	return sa, nil
}

func sdtsFnMakeStateIns(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	state, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	// also grab the state ID's token from args
	stateTok, ok := args[1].(lex.Token)
	if !ok {
		return nil, newArgTypeError(args, 1, "lex.Token")
	}

	return box.Pair[string, lex.Token]{First: state, Second: stateTok}, nil
}

func sdtsFnMakeGrammarContentNode(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	state, ok := args[0].(box.Pair[string, lex.Token])
	if !ok {
		return nil, newArgTypeError(args, 0, "box.Pair[string, lex.Token]")
	}

	rules, ok := args[1].([]GrammarRule)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]GrammarRule")
	}
	return GrammarContent{Rules: rules, State: state.First, SrcState: state.Second, Src: info.FirstToken}, nil
}

func sdtsFnMakeActionsContentNode(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	state, ok := args[0].(box.Pair[string, lex.Token])
	if !ok {
		return nil, newArgTypeError(args, 0, "box.Pair[string, lex.Token]")
	}

	actions, ok := args[1].([]SymbolActions)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]SymbolActions")
	}
	return ActionsContent{Actions: actions, State: state.First, SrcState: state.Second, Src: info.FirstToken}, nil
}

func sdtsFnMakeTokensContentNode(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	state, ok := args[0].(box.Pair[string, lex.Token])
	if !ok {
		return nil, newArgTypeError(args, 0, "box.Pair[string, lex.Token]")
	}

	entries, ok := args[1].([]TokenEntry)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]TokenEntry")
	}

	return TokensContent{Entries: entries, State: state.First, SrcState: state.Second, Src: info.FirstToken}, nil
}

func sdtsFnTrimString(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	str, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}
	return strings.TrimSpace(str), nil
}

func sdtsFnMakeDiscardOption(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	return TokenOption{Type: TokenOptDiscard, Src: info.FirstToken}, nil
}

func sdtsFnMakeStateshiftOption(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	state, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	return TokenOption{Type: TokenOptStateshift, Value: state, Src: info.FirstToken}, nil
}

func sdtsFnMakeHumanOption(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	human, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	return TokenOption{Type: TokenOptHuman, Value: human, Src: info.FirstToken}, nil
}

func sdtsFnMakeTokenOption(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	t, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	return TokenOption{Type: TokenOptToken, Value: t, Src: info.FirstToken}, nil
}

func sdtsFnMakePriorityOption(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	priority, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	return TokenOption{Type: TokenOptPriority, Value: priority, Src: info.FirstToken}, nil
}

func sdtsFnIdentity(_ trans.SetterInfo, args []interface{}) (interface{}, error) { return args[0], nil }

func sdtsFnInterpretEscape(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	str, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	str = strings.TrimLeftFunc(str, unicode.IsSpace) // lets us handle startLineEscseq as well, glub!
	if len(str) < len("%!") {
		return nil, newArgError(args, 0, "string too short to be an escape sequence: %q", str)
	}

	// escape sequence is %!, so just take the chars after that
	return str[len("%!"):], nil
}

func sdtsFnAppendStrings(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	str1, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}
	str2, ok := args[1].(string)
	if !ok {
		return nil, newArgTypeError(args, 1, "string")
	}

	return str1 + str2, nil
}

func sdtsFnAppendStringsTrimmed(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	str1, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}
	str2, ok := args[1].(string)
	if !ok {
		return nil, newArgTypeError(args, 1, "string")
	}

	return strings.TrimSpace(str1 + str2), nil
}

func sdtsFnGetNonterminal(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	str, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	return strings.TrimSpace(str), nil
}

func sdtsFnGetInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	str, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	iVal, err := strconv.Atoi(str)
	if err != nil {
		return nil, newArgError(args, 0, err.Error())
	}
	return iVal, nil
}

func sdtsFnGetTerminal(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	str, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	return strings.ToLower(str), nil
}

func sdtsFnRuleListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]GrammarRule)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]GrammarRule")
	}

	toAppend, ok := args[1].(GrammarRule)
	if !ok {
		return nil, newArgTypeError(args, 1, "GrammarRule")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnEntryListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]TokenEntry)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]TokenEntry")
	}

	toAppend, ok := args[1].(TokenEntry)
	if !ok {
		return nil, newArgTypeError(args, 1, "TokenEntry")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnActionsStateBlockListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]ActionsContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]ActionsContent")
	}

	toAppend, ok := args[1].(ActionsContent)
	if !ok {
		return nil, newArgTypeError(args, 1, "ActionsContent")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnTokensStateBlockListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]TokensContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]TokensContent")
	}

	toAppend, ok := args[1].(TokensContent)
	if !ok {
		return nil, newArgTypeError(args, 1, "TokensContent")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnGrammarStateBlockListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]GrammarContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]GrammarContent")
	}

	toAppend, ok := args[1].(GrammarContent)
	if !ok {
		return nil, newArgTypeError(args, 1, "GrammarContent")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnSymbolActionsListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]SymbolActions)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]SymbolActions")
	}

	toAppend, ok := args[1].(SymbolActions)
	if !ok {
		return nil, newArgTypeError(args, 1, "SymbolActions")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnProdActionListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]ProductionAction)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]ProductionAction")
	}

	toAppend, ok := args[1].(ProductionAction)
	if !ok {
		return nil, newArgTypeError(args, 1, "ProductionAction")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnSemanticActionListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]SemanticAction)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]SemanticAction")
	}

	toAppend, ok := args[1].(SemanticAction)
	if !ok {
		return nil, newArgTypeError(args, 1, "SemanticAction")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnAttrRefListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]AttrRef)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]AttrRef")
	}

	// get token of attr ref to build fake info object to pass to sdtsFnGetAttrRef's info.
	fakeInfo, err := makeFakeInfo(args[2], fetoken.TCAttrRef.ID(), "value")
	if err != nil {
		return nil, newArgError(args, 2, err.Error())
	}

	toAppendUncast, err := sdtsFnGetAttrRef(fakeInfo, args[1:])
	if err != nil {
		if hookErr, ok := err.(*hookArgError); ok {
			hookErr.Args = args
			hookErr.ArgNum = 1
			return nil, hookErr
		}
		return nil, err
	}
	toAppend := toAppendUncast.(AttrRef)

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnAttrRefListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	// get token of attr ref to build fake info object to pass to sdtsFnGetAttrRef's info.
	fakeInfo, err := makeFakeInfo(args[1], fetoken.TCAttrRef.ID(), "value")
	if err != nil {
		return nil, newArgError(args, 1, err.Error())
	}

	toAppendUncast, err := sdtsFnGetAttrRef(fakeInfo, args[0:])
	if err != nil {
		return nil, err
	}
	toAppend := toAppendUncast.(AttrRef)

	return []AttrRef{toAppend}, nil
}

func sdtsFnGetAttrRef(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	var attrRef AttrRef

	str, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	} else {
		var err error
		attrRef, err = ParseAttrRef(str)
		if err != nil {
			return nil, newArgError(args, 0, err.Error())
		}
	}

	attrRef.Src = info.FirstToken

	return attrRef, nil
}

func sdtsFnMakeSemanticAction(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	fakeInfo, err := makeFakeInfo(args[1], fetoken.TCAttrRef.ID(), "value")
	if err != nil {
		return nil, newArgError(args, 1, err.Error())
	}

	attrRefUncast, err := sdtsFnGetAttrRef(fakeInfo, args[0:])
	if err != nil {
		return nil, err
	}
	attrRef := attrRefUncast.(AttrRef)

	hookId, ok := args[2].(string)
	if !ok {
		return nil, newArgTypeError(args, 2, "string")
	}

	hookTok, ok := args[3].(lex.Token)
	if !ok {
		return nil, newArgTypeError(args, 3, "lex.Token")
	}

	var argRefs []AttrRef
	if len(args) > 4 {
		argRefs, ok = args[4].([]AttrRef)
		if !ok {
			return nil, newArgTypeError(args, 4, "[]AttrRef")
		}
	}

	sa := SemanticAction{
		LHS:     attrRef,
		Hook:    hookId,
		With:    argRefs,
		SrcHook: hookTok,
		Src:     info.FirstToken,
	}

	return sa, nil
}

func sdtsFnMakeProdSpecifierNext(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	// need exact generic-filled type to match later expectations.
	spec := box.Triple[string, interface{}, lex.Token]{First: "NEXT", Second: "", Third: info.FirstToken}
	return spec, nil
}

func sdtsFnMakeProdSpecifierIndex(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	index, err := sdtsFnGetInt(trans.SetterInfo{}, args)
	if err != nil {
		return nil, err
	}

	// need exact generic-filled type to match later expectations.
	spec := box.Triple[string, interface{}, lex.Token]{First: "INDEX", Second: index, Third: info.FirstToken}
	return spec, nil
}

func sdtsFnMakeProdSpecifierLiteral(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	prod, ok := args[0].([]string)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]string")
	}

	// need exact generic-filled type to match later expectations.
	spec := box.Triple[string, interface{}, lex.Token]{First: "LITERAL", Second: prod, Third: info.FirstToken}
	return spec, nil
}

func sdtsFnProdActionListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(ProductionAction)
	if !ok {
		return nil, newArgTypeError(args, 0, "ProductionAction")
	}

	return []ProductionAction{toAppend}, nil
}

func sdtsFnSemanticActionListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(SemanticAction)
	if !ok {
		return nil, newArgTypeError(args, 0, "SemanticAction")
	}

	return []SemanticAction{toAppend}, nil
}

func sdtsFnRuleListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(GrammarRule)
	if !ok {
		return nil, newArgTypeError(args, 0, "GrammarRule")
	}

	return []GrammarRule{toAppend}, nil
}

func sdtsFnGrammarStateBlockListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(GrammarContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "GrammarContent")
	}

	return []GrammarContent{toAppend}, nil
}

func sdtsFnTokensStateBlockListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(TokensContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "TokensContent")
	}

	return []TokensContent{toAppend}, nil
}

func sdtsFnActionsStateBlockListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(ActionsContent)
	if !ok {
		return nil, newArgTypeError(args, 0, "ActionsContent")
	}

	return []ActionsContent{toAppend}, nil
}

func sdtsFnSymbolActionsListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(SymbolActions)
	if !ok {
		return nil, newArgTypeError(args, 0, "SymbolActions")
	}

	return []SymbolActions{toAppend}, nil
}

func sdtsFnEntryListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(TokenEntry)
	if !ok {
		return nil, newArgTypeError(args, 0, "TokenEntry")
	}

	return []TokenEntry{toAppend}, nil
}

func sdtsFnStringListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]string)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]string")
	}

	toAppend, ok := args[1].(string)
	if !ok {
		return nil, newArgTypeError(args, 1, "string")
	}

	list = append(list, toAppend)

	return list, nil
}

func sdtsFnTokenOptListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].(TokenOption)
	if !ok {
		return nil, newArgTypeError(args, 0, "TokenOption")
	}

	return []TokenOption{toAppend}, nil
}

func sdtsFnTokenOptListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]TokenOption)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]TokenOption")
	}

	toAppend, ok := args[1].(TokenOption)
	if !ok {
		return nil, newArgTypeError(args, 1, "TokenOption")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnStringListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend := coerceToString(args[0])
	return []string{toAppend}, nil
}

func sdtsFnStringListListStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	toAppend, ok := args[0].([]string)
	if !ok {
		return nil, newArgTypeError(args, 0, "[]string")
	}

	return [][]string{toAppend}, nil
}

func sdtsFnStringListListAppend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	list, ok := args[0].([][]string)
	if !ok {
		return nil, newArgTypeError(args, 0, "[][]string")
	}

	toAppend, ok := args[1].([]string)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]string")
	}

	list = append(list, toAppend)
	return list, nil
}

func sdtsFnEpsilonStringList(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	strList := grammar.Epsilon.Copy()
	return []string(strList), nil
}

func sdtsFnMakeRule(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	ntInterface, err := sdtsFnGetNonterminal(trans.SetterInfo{}, args)
	if err != nil {
		return nil, err
	}

	nt, ok := ntInterface.(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	productions, ok := args[1].([][]string)
	if !ok {
		return nil, newArgTypeError(args, 1, "[][]string")
	}

	gr := grammar.Rule{NonTerminal: nt, Productions: []grammar.Production{}}

	for _, p := range productions {
		gr.Productions = append(gr.Productions, p)
	}

	r := GrammarRule{
		Rule: gr,
		Src:  info.FirstToken,
	}

	return r, nil
}

func sdtsFnMakeTokenEntry(info trans.SetterInfo, args []interface{}) (interface{}, error) {
	pattern, ok := args[0].(string)
	if !ok {
		return nil, newArgTypeError(args, 0, "string")
	}

	tokenOpts, ok := args[1].([]TokenOption)
	if !ok {
		return nil, newArgTypeError(args, 1, "[]TokenOption")
	}

	t := TokenEntry{Pattern: pattern, Src: info.FirstToken}

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
	return t, nil
}

// for hooks to generate fake info object when needed. Sym and name can be blank
// if desired. Returned SetterInfo will always have synthetic set to true.
func makeFakeInfo(from interface{}, sym, name string) (trans.SetterInfo, error) {
	tok, ok := from.(lex.Token)
	if !ok {
		return trans.SetterInfo{}, fmt.Errorf("not a token")
	}

	info := trans.SetterInfo{
		GrammarSymbol: sym,
		Synthetic:     true,
		Name:          name,
		FirstToken:    tok,
	}

	return info, nil
}

func coerceToString(a interface{}) string {
	// is it just a string? return if so
	if str, ok := a.(string); ok {
		return str
	}

	// otherwise, is it a stringer? call String() and return if so
	if str, ok := a.(fmt.Stringer); ok {
		return str.String()
	}

	// otherwise, is it an error? call Error() and return if so
	if err, ok := a.(error); ok {
		return err.Error()
	}

	// finally, if none of those, get the default formatting and return that
	return fmt.Sprintf("%v", a)
}
