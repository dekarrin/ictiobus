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

func sddFnMakeFishispec(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astBlock)
	if !ok {
		return AST{}
	}

	return AST{nodes: list}
}

func sddFnBlockListAppend(_, _ string, args []interface{}) interface{} {
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

func sddFnBlockListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astBlock)
	if !ok {
		var errBl astErrorBlock
		toAppend = errBl
	}

	return []astBlock{toAppend}
}

func sddFnMakeGrammarBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}

	return astGrammarBlock{content: list}
}

func sddFnMakeTokensBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	return astTokensBlock{content: list}
}

func sddFnGrammarContentBlocksAppendStateBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		list = []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
	}
	toAppend, ok := args[1].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{state: SDDErrMsg("producing this grammar content: first argument is not a grammar content")}
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

func sddFnTokensContentBlocksAppendStateBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}
	toAppend, ok := args[1].(astTokensContent)
	if !ok {
		toAppend = astTokensContent{state: SDDErrMsg("producing this tokens content: first argument is not a tokens content")}
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

func sddFnGrammarContentBlocksStartStateBlock(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{state: SDDErrMsg("producing this grammar content: first argument is not a grammar content")}
	}

	return []astGrammarContent{toAppend}
}

func sddFnTokensContentBlocksStartStateBlock(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astTokensContent)
	if !ok {
		toAppend = astTokensContent{state: SDDErrMsg("producing this tokens content: first argument is not a tokens content")}
	}

	return []astTokensContent{toAppend}
}

func sddFnGrammarContentBlocksAppendRuleList(_, _ string, args []interface{}) interface{} {
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

func sddFnTokensContentBlocksAppendRuleList(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokensContent)
	if !ok {
		list = []astTokensContent{{state: SDDErrMsg("producing this tokens content list: first argument is not a tokens content list")}}
	}

	entries, ok := args[1].([]astTokenEntry)
	if !ok {
		entries = []astTokenEntry{{pattern: SDDErrMsg("producing this token entry list: second argument is not a token entry list")}}
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

func sddFnGrammarContentBlocksStartRuleList(_, _ string, args []interface{}) interface{} {
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

func sddFnTokensContentBlocksStartRuleList(_, _ string, args []interface{}) interface{} {
	entries, ok := args[0].([]astTokenEntry)
	if !ok {
		entries = []astTokenEntry{{pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}
	toAppend := astTokensContent{
		entries: entries,
		state:   "",
	}

	return []astTokensContent{toAppend}
}

func sddFnMakeGrammarContentNode(_, _ string, args []interface{}) interface{} {
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

func sddFnMakeTokensContentNode(_, _ string, args []interface{}) interface{} {
	state, ok := args[0].(string)
	if !ok {
		state = SDDErrMsg("STATE value is not a string")
	}
	entries, ok := args[1].([]astTokenEntry)
	if !ok {
		entries = []astTokenEntry{{pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list")}}
	}
	return astTokensContent{entries: entries, state: state}
}

func sddFnTrimString(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return SDDErrMsg("argument is not a string")
	}
	return strings.TrimSpace(str)
}

func sddFnIdentity(_, _ string, args []interface{}) interface{} { return args[0] }

func sddFnInterpretEscape(_, _ string, args []interface{}) interface{} {
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

func sddFnAppendStrings(_, _ string, args []interface{}) interface{} {
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

func sddFnGetNonterminal(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return strings.ToUpper(str[1 : len(str)-1])
}

func sddFnGetTerminal(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	return strings.ToLower(str)
}

func sddFnRuleListAppend(_, _ string, args []interface{}) interface{} {
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

func sddFnEntryListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astTokenEntry)
	if !ok {
		list = []astTokenEntry{{pattern: SDDErrMsg("producing this token entry list: first argument is not a token entry list list")}}
	}

	toAppend, ok := args[1].(astTokenEntry)
	if !ok {
		toAppend = astTokenEntry{pattern: SDDErrMsg("producing this token entry: second argument is not a token entry")}
	}

	list = append(list, toAppend)
	return list
}

func sddFnRuleListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(grammar.Rule)
	if !ok {
		toAppend = grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: second argument is not a rule")}
	}

	return []grammar.Rule{toAppend}
}

func sddFnEntryListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astTokenEntry)
	if !ok {
		toAppend = astTokenEntry{pattern: SDDErrMsg("producing this token entry: second argument is not a token entry")}
	}

	return []astTokenEntry{toAppend}
}

func sddFnStringListAppend(_, _ string, args []interface{}) interface{} {
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

func sddFnStringListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(string)
	if !ok {
		toAppend = ErrString
	}

	return []string{toAppend}
}

func sddFnStringListListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].([]string)
	if !ok {
		toAppend = []string{SDDErrMsg("producing this string list: first argument is not a string list")}
	}

	return [][]string{toAppend}
}

func sddFnStringListListAppend(_, _ string, args []interface{}) interface{} {
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

func sddFnEpsilonStringList(_, _ string, args []interface{}) interface{} {
	strList := grammar.Epsilon.Copy()
	return []string(strList)
}

func sddFnMakeRule(_, _ string, args []interface{}) interface{} {
	ntInterface := sddFnGetNonterminal("", "", args[0:1])

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

func sddFnMakeTokenEntry(_, _ string, args []interface{}) interface{} {
	pattern, ok := args[0].(string)
	if !ok {
		pattern = SDDErrMsg("first argument (pattern) is not a string")
	}

	tokenOpts, ok := args[1].([]astTokenOption)
	if !ok {
		tokenOpts = []astTokenOption{{value: SDDErrMsg("producing this token option list: second argument (tokenOpts) is not a token option list")}}
	}

	t := astTokenEntry{pattern: pattern}

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
