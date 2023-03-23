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
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case blockTypeTokens:
				toks := n.Tokens()
				sb.WriteString("  <TOKENS:\n")
				for j := range toks.content {
					cont := toks.content[j]
					sb.WriteString("    <CONTENT:\n")
					sb.WriteString("      STATE: " + fmt.Sprintf("%q\n", cont.state))
					for k := range cont.entries {
						entry := cont.entries[k]
						sb.WriteString("      E: " + entry.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case blockTypeActions:
				acts := n.Actions()
				sb.WriteString("  <ACTIONS:\n")
				for j := range acts.content {
					cont := acts.content[j]
					sb.WriteString("    <CONTENT:\n")
					sb.WriteString("      STATE: " + fmt.Sprintf("%q\n", cont.state))
					for k := range cont.actions {
						action := cont.actions[k]
						sb.WriteString("      A: " + action.String() + "\n")
					}
					sb.WriteString("    >\n")
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
	blockTypeTokens
	blockTypeActions
)

type astBlock interface {
	Type() blockType
	Grammar() astGrammarBlock
	Tokens() astTokensBlock
	Actions() astActionsBlock
}

type astErrorBlock bool

func (errBlock astErrorBlock) Type() blockType {
	return blockTypeError
}

func (errBlock astErrorBlock) Grammar() astGrammarBlock {
	panic("not grammar-type block")
}

func (errBlock astErrorBlock) Tokens() astTokensBlock {
	panic("not tokens-type block")
}

func (errBlock astErrorBlock) Actions() astActionsBlock {
	panic("not actions-type block")
}

func (errBlock astErrorBlock) String() string {
	return "<Block: ERR>"
}

type astGrammarBlock struct {
	content []astGrammarContent
}

func (agb astGrammarBlock) String() string {
	var sb strings.Builder

	sb.WriteString("<Block: GRAMMAR, Content: {")
	for i := range agb.content {
		sb.WriteString(agb.content[i].String())
		if i+1 < len(agb.content) {
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

func (agb astGrammarBlock) Tokens() astTokensBlock {
	panic("not tokens-type block")
}

func (agb astGrammarBlock) Actions() astActionsBlock {
	panic("not actions-type block")
}

type astActionsBlock struct {
	content []astActionsContent
}

func (aab astActionsBlock) String() string {
	var sb strings.Builder

	sb.WriteString("<Block: GRAMMAR, Content: {")
	for i := range aab.content {
		sb.WriteString(aab.content[i].String())
		if i+1 < len(aab.content) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune('}')
	return sb.String()
}

func (aab astActionsBlock) Type() blockType {
	return blockTypeActions
}

func (aab astActionsBlock) Grammar() astGrammarBlock {
	panic("not grammar-type block")
}

func (aab astActionsBlock) Tokens() astTokensBlock {
	panic("not tokens-type block")
}

func (aab astActionsBlock) Actions() astActionsBlock {
	return aab
}

type astTokensBlock struct {
	content []astTokensContent
}

func (atb astTokensBlock) String() string {
	var sb strings.Builder

	sb.WriteString("<Block: TOKENS, Content: {")
	for i := range atb.content {
		sb.WriteString(atb.content[i].String())
		if i+1 < len(atb.content) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune('}')
	return sb.String()
}

func (atb astTokensBlock) Type() blockType {
	return blockTypeTokens
}

func (atb astTokensBlock) Grammar() astGrammarBlock {
	panic("not grammar-type block")
}

func (atb astTokensBlock) Tokens() astTokensBlock {
	return atb
}

func (atb astTokensBlock) Actions() astActionsBlock {
	panic("not actions-type block")
}

type astTokenOptionType int

const (
	tokenOptDiscard astTokenOptionType = iota
	tokenOptStateshift
	tokenOptToken
	tokenOptHuman
	tokenOptPriority
)

type astTokenOption struct {
	optType astTokenOptionType
	value   string
}

type tokenEntry struct {
	pattern  string
	discard  bool
	shift    string
	token    string
	human    string
	priority int
}

func (entry tokenEntry) String() string {
	var sb strings.Builder

	sb.WriteRune('{')
	sb.WriteString(fmt.Sprintf("%q -> ", entry.pattern))
	sb.WriteString(fmt.Sprintf("Discard: %v, ", entry.discard))
	sb.WriteString(fmt.Sprintf("Shift: %q, ", entry.shift))
	sb.WriteString(fmt.Sprintf("Token: %q, ", entry.token))
	sb.WriteString(fmt.Sprintf("Human: %q, ", entry.human))
	sb.WriteString(fmt.Sprintf("Priority: %d", entry.priority))
	sb.WriteRune('}')

	return sb.String()
}

type astTokensContent struct {
	entries []tokenEntry
	state   string
}

func (content astTokensContent) String() string {
	if len(content.entries) > 0 {
		return fmt.Sprintf("(State: %q, Entries: %v)", content.state, content.entries)
	} else {
		return fmt.Sprintf("(State: %q, Entries: (empty))", content.state)
	}
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

type astActionsContent struct {
	actions []symbolActions
	state   string
}

func (content astActionsContent) String() string {
	if len(content.actions) > 0 {
		return fmt.Sprintf("(State: %q, Actions: %v)", content.state, content.actions)
	} else {
		return fmt.Sprintf("(State: %q, Actions: (empty))", content.state)
	}
}

type attrRef struct {
	symbol    string
	terminal  bool
	occurance int
	attribute string
}

func (ar attrRef) String() string {
	var sb strings.Builder

	if ar.terminal {
		sb.WriteString(ar.symbol)
	} else {
		sb.WriteRune('{')
		sb.WriteString(ar.symbol)
		sb.WriteRune('}')
	}

	sb.WriteRune('$')
	sb.WriteString(fmt.Sprintf("%d", ar.occurance))
	sb.WriteRune('.')
	sb.WriteString(ar.attribute)

	return sb.String()
}

type semanticAction struct {
	lhs  attrRef
	hook string
	with []attrRef
}

func (sa semanticAction) String() string {
	var sb strings.Builder

	sb.WriteString(sa.lhs.String())
	sb.WriteString(" = ")
	sb.WriteString(sa.hook)

	sb.WriteRune('(')
	for i := range sa.with {
		sb.WriteString(sa.with[i].String())
		if i+1 < len(sa.with) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune(')')

	return sb.String()
}

type productionAction struct {
	prodNext    bool
	prodIndex   int
	prodLiteral []string

	actions []semanticAction
}

func (pa productionAction) String() string {
	var sb strings.Builder

	sb.WriteString("(prod ")
	if pa.prodNext {
		sb.WriteString("(next)")
	} else if pa.prodLiteral != nil {
		sb.WriteRune('{')
		if len(pa.prodLiteral) == 1 && pa.prodLiteral[0] == grammar.Epsilon[0] {
			sb.WriteString("ε")
		} else {
			sb.WriteString(strings.Join(pa.prodLiteral, " "))
		}
		sb.WriteRune('}')
	} else {
		sb.WriteString(fmt.Sprintf("(index %d)", pa.prodIndex))
	}

	sb.WriteString(": {")
	for i := range pa.actions {
		sb.WriteString(pa.actions[i].String())
		if i+1 < len(pa.actions) {
			sb.WriteString("; ")
		}
	}

	sb.WriteRune(')')

	return sb.String()
}

type symbolActions struct {
	symbol  string
	actions []productionAction
}

func (sa symbolActions) String() string {
	var sb strings.Builder

	sb.WriteRune('{')
	sb.WriteString(sa.symbol)
	sb.WriteString("}: [")

	for i := range sa.actions {
		sb.WriteString(sa.actions[i].String())
		if i+1 < len(sa.actions) {
			sb.WriteString(", ")
		}
	}

	sb.WriteRune(']')

	return sb.String()
}
