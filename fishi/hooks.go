package fishi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
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

func sdtsFnMakeGrammarContentNode(_, _ string, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		state = SDDErrMsg("STATE value is not a string")
	}
	rules, ok := args[1].([]grammar.Rule)
	if !ok {
		rules = []grammar.Rule{}
	}
	return astGrammarContent{rules: rules, state: state}
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

	return strings.ToUpper(str[1 : len(str)-1])
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
		list = []tokenEntry{{pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list list")}}
	}

	toAppend, ok := args[1].(tokenEntry)
	if !ok {
		toAppend = tokenEntry{pattern: SDDErrMsg("producing this token entry: second argument is not a token entry")}
	}

	list = append(list, toAppend)
	return list
}

func sdtsFnRuleListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(grammar.Rule)
	if !ok {
		toAppend = grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: second argument is not a rule")}
	}

	return []grammar.Rule{toAppend}
}

func sdtsFnEntryListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(tokenEntry)
	if !ok {
		toAppend = tokenEntry{pattern: SDDErrMsg("producing this token entry: second argument is not a token entry")}
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
