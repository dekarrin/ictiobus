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

const ErrString = "<ERR>"

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
		return []astGrammarContent{}
	}
	toAppend, ok := args[1].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{}
	}

	list = append(list, toAppend)
	return list
}

func sddFnGrammarContentBlocksStartStateBlock(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(astGrammarContent)
	if !ok {
		toAppend = astGrammarContent{}
	}

	return []astGrammarContent{toAppend}
}

func sddFnGrammarContentBlocksAppendRuleList(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([]astGrammarContent)
	if !ok {
		return []astGrammarContent{}
	}

	rules, ok := args[1].([]grammar.Rule)
	if !ok {
		rules = []grammar.Rule{}
	}
	toAppend := astGrammarContent{
		rules: rules,
		state: "",
	}

	list = append(list, toAppend)
	return list
}

func sddFnGrammarContentBlocksStartRuleList(_, _ string, args []interface{}) interface{} {
	rules, ok := args[0].([]grammar.Rule)
	if !ok {
		rules = []grammar.Rule{}
	}
	toAppend := astGrammarContent{
		rules: rules,
		state: "",
	}

	return []astGrammarContent{toAppend}
}

func sddFnMakeGrammarContentNode(_, _ string, args []interface{}) interface{} {
	rules, ok := args[0].([]grammar.Rule)
	if !ok {
		rules = []grammar.Rule{}
	}
	state, ok := args[1].(string)
	if !ok {
		state = ErrString
	}
	return astGrammarContent{rules: rules, state: state}
}

func sddFnIdentity(_, _ string, args []interface{}) interface{} { return args[0] }

func sddFnInterpretEscape(_, _ string, args []interface{}) interface{} {
	str, ok := args[0].(string)
	if !ok {
		return ErrString
	}

	// escape sequence is %!, so just take the chars after that
	return str[len("%!"):]
}

func sddFnAppendStrings(_, _ string, args []interface{}) interface{} {
	str1, ok := args[0].(string)
	if !ok {
		return ErrString
	}
	str2, ok := args[1].(string)
	if !ok {
		return ErrString
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
		return []grammar.Rule{}
	}

	toAppend, ok := args[1].(grammar.Rule)
	if !ok {
		toAppend = grammar.Rule{
			NonTerminal: ErrString,
		}
	}

	list = append(list, toAppend)
	return list
}

func sddFnRuleListStart(_, _ string, args []interface{}) interface{} {
	toAppend, ok := args[0].(grammar.Rule)
	if !ok {
		toAppend = grammar.Rule{}
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
		toAppend = []string{}
	}

	return [][]string{toAppend}
}

func sddFnStringListListAppend(_, _ string, args []interface{}) interface{} {
	list, ok := args[0].([][]string)
	if !ok {
		return [][]string{}
	}

	toAppend, ok := args[1].([]string)
	if !ok {
		toAppend = []string{ErrString}
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
		nt = ErrString
	}

	productions, ok := args[1].([][]string)
	if !ok {
		productions = [][]string{}
	}

	r := grammar.Rule{NonTerminal: nt, Productions: []grammar.Production{}}

	for _, p := range productions {
		r.Productions = append(r.Productions, p)
	}

	return r
}
