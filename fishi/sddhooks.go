package fishi

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
)

type AST struct {
	nodes []astBlock
}

func (ast AST) String() string {
	var sb strings.Builder

	sb.WriteRune('<')
	if len(ast.nodes) > 0 {
		sb.WriteRune('\n')
		for i := range ast.nodes {
			n := ast.nodes[i]
			switch n.Type() {
			case blockTypeError:
				sb.WriteString("  <ERR>\n")
			case blockTypeGrammar:
				gram := n.Grammar()
				sb.WriteString("  <GRAMMAR:\n")
				for j := range gram.content {
					cont := gram.content[j]
					sb.WriteString("    <CONTENT:\n")
					sb.WriteString("      STATE: " + fmt.Sprintf("%q\n", cont.state))
					for k := range cont.rules {
						r := cont.rules[k]
						sb.WriteString("      R: " + r.String() + "\n")
					}
					sb.WriteString("    >,\n")
				}
				sb.WriteString("  >\n")
			}
		}
	}
	sb.WriteRune('>')

	return sb.String()
}

type blockType int

const (
	blockTypeError blockType = iota
	blockTypeGrammar
)

type astBlock interface {
	Type() blockType
	Grammar() astGrammarBlock
}

type astErrorBlock bool

func (errBlock astErrorBlock) Type() blockType {
	return blockTypeError
}

func (errBlock astErrorBlock) Grammar() astGrammarBlock {
	panic("not grammar type block")
}

func (errBlock astErrorBlock) String() string {
	return "<Block: ERR>"
}

type astGrammarBlock struct {
	content []astGrammarContent
}

func (bl astGrammarBlock) String() string {
	var sb strings.Builder

	sb.WriteString("<Block: GRAMMAR, Content: {")
	for i := range bl.content {
		sb.WriteString(bl.content[i].String())
		if i+1 < len(bl.content) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune('}')
	return sb.String()
}

func (agb astGrammarBlock) Type() blockType {
	return blockTypeGrammar
}

func (agb astGrammarBlock) Grammar() astGrammarBlock {
	return agb
}

type astGrammarContent struct {
	rules []grammar.Rule
	state string
}

func (content astGrammarContent) String() string {
	if len(content.rules) > 0 {
		return fmt.Sprintf("(State: %q, Rules: %v)", content.state, content.rules)
	} else {
		return fmt.Sprintf("(State: %q, Rules: (empty))", content.state)
	}
}

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
		list = []astGrammarContent{}
	}

	return astGrammarBlock{content: list}
}

func sddFnGrammarContentBlocksAppendStateBlock(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		return []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
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

func sddFnGrammarContentBlocksStartStateBlock(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{state: SDDErrMsg("producing this grammar content: first argument is not a grammar content")}
	}

	return []astGrammarContent{toAppend}
}

func sddFnGrammarContentBlocksAppendRuleList(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		return []astGrammarContent{{state: SDDErrMsg("producing this grammar content list: first argument is not a grammar content list")}}
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

func sddFnRuleListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(grammar.Rule)
	if !ok {
		toAppend = grammar.Rule{NonTerminal: SDDErrMsg("producing this rule: second argument is not a rule")}
	}

	return []grammar.Rule{toAppend}
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

func sddFnNilStringList(_, _ string, args []interface{}) interface{} {
	var strList []string
	return strList
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
