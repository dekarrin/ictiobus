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
	blockTypeTokens
)

type astBlock interface {
	Type() blockType
	Grammar() astGrammarBlock
	Tokens() astTokensBlock
}

type astErrorBlock bool

func (errBlock astErrorBlock) Type() blockType {
	return blockTypeError
}

func (errBlock astErrorBlock) Grammar() astGrammarBlock {
	panic("not grammar type block")
}

func (errBlock astErrorBlock) Tokens() astTokensBlock {
	panic("not tokens type block")
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
	panic("not tokens type block")
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
	panic("not grammar type block")
}

func (atb astTokensBlock) Tokens() astTokensBlock {
	return atb
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

type astTokenEntry struct {
	pattern  string
	discard  bool
	shift    string
	token    string
	human    string
	priority int
}

func (entry astTokenEntry) String() string {
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
	entries []astTokenEntry
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
