package fishi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/types"
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
					if cont.state != "" {
						sb.WriteString("    <RULE-SET FOR STATE " + fmt.Sprintf("%q\n", cont.state))
					} else {
						sb.WriteString("    <RULE-SET FOR ALL STATES\n")
					}
					for k := range cont.rules {
						r := cont.rules[k]
						sb.WriteString("      * " + r.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case blockTypeTokens:
				toks := n.Tokens()
				sb.WriteString("  <TOKENS:\n")
				for j := range toks.content {
					cont := toks.content[j]
					if cont.state != "" {
						sb.WriteString("    <ENTRY-SET FOR STATE " + fmt.Sprintf("%q\n", cont.state))
					} else {
						sb.WriteString("    <ENTRY-SET FOR ALL STATES\n")
					}
					for k := range cont.entries {
						entry := cont.entries[k]
						sb.WriteString("      * " + entry.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case blockTypeActions:
				acts := n.Actions()
				sb.WriteString("  <ACTIONS:\n")
				for j := range acts.content {
					cont := acts.content[j]
					if cont.state != "" {
						sb.WriteString("    <ACTION-SET FOR STATE " + fmt.Sprintf("%q\n", cont.state))
					} else {
						sb.WriteString("    <ACTION-SET FOR ALL STATES\n")
					}
					for k := range cont.actions {
						action := cont.actions[k]
						sb.WriteString("      * " + action.String() + "\n")
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
	tok     types.Token
}

type tokenEntry struct {
	pattern  string
	discard  bool
	shift    string
	token    string
	human    string
	priority int

	tok types.Token

	// in theory may be multiple options for the same option type; while it
	// is not allowed semantically, it is allowed syntactically, so track it so
	// we can do proper error reporting later.
	discardTok  []types.Token
	shiftTok    []types.Token
	tokenTok    []types.Token
	humanTok    []types.Token
	priorityTok []types.Token

	// (don't need a patternTok because that pattern is the first symbol and
	// there can only be one; tok will be the same as patternTok)
}

func (entry tokenEntry) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s -> ", entry.pattern))
	sb.WriteString(fmt.Sprintf("Discard: %v, ", entry.discard))
	sb.WriteString(fmt.Sprintf("Shift: %q, ", entry.shift))
	sb.WriteString(fmt.Sprintf("Token: %q, ", entry.token))
	sb.WriteString(fmt.Sprintf("Human: %q, ", entry.human))
	sb.WriteString(fmt.Sprintf("Priority: %d", entry.priority))

	return sb.String()
}

type astGrammarRule struct {
	rule grammar.Rule
	tok  types.Token
}

func (agr astGrammarRule) String() string {
	return agr.rule.String()
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
	rules []astGrammarRule
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

type AttrRef struct {
	symbol   string
	terminal bool

	head          bool
	termInProd    bool
	nontermInProd bool
	symInProd     bool

	occurance int
	attribute string

	tok types.Token
}

// ParseAttrRef does a simple parse on an attribute reference from a string that
// makes it up. Does not set tok; caller must do so if needed.
func ParseAttrRef(s string) (AttrRef, error) {
	dotSpl := strings.Split(s, ".")
	if len(dotSpl) < 2 {
		return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
	}

	attrName := dotSpl[len(dotSpl)-1]
	nodeRefStr := strings.Join(dotSpl[:len(dotSpl)-1], ".")

	ar := AttrRef{attribute: attrName}

	if nodeRefStr[0] == '{' && nodeRefStr[len(nodeRefStr)-1] == '}' {
		str := nodeRefStr[1 : len(nodeRefStr)-1]
		if (str[0] >= 'A' && str[0] <= 'Z') || (str[0] >= 'a' && str[0] <= 'z') {
			// nonterminal-by-name reference
			ar.symbol = str

			// get index from $num... sequence at end of ref str
			allSplits := strings.Split(nodeRefStr, "$")
			if len(allSplits) > 1 {
				lastSplit := allSplits[len(allSplits)-1]
				firstSplits := strings.Join(allSplits[:len(allSplits)-1], "$")

				var err error
				ar.occurance, err = strconv.Atoi(lastSplit)
				if err != nil {
					// not an error, it's optional
					ar.occurance = 0
				} else {
					ar.symbol = firstSplits
				}
			}
			return ar, nil
		} else if str == "^" {
			ar.head = true
			return ar, nil
		} else if strings.HasPrefix(str, ".") {
			ar.termInProd = true
			if len(str) > 1 {
				str = str[1:]
				num, err := strconv.Atoi(str)
				if err != nil {
					return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
				}
				ar.occurance = num
			}
			return ar, nil
		} else if strings.HasPrefix(str, "&") {
			ar.nontermInProd = true
			if len(str) > 1 {
				str = str[1:]
				num, err := strconv.Atoi(str)
				if err != nil {
					return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
				}
				ar.occurance = num
			}
			return ar, nil
		} else {
			// then it has to be a parsable number
			num, err := strconv.Atoi(str)
			if err != nil {
				return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
			}
			ar.occurance = num
			ar.symInProd = true
			return ar, nil
		}
	} else {
		// terminal-by-name reference
		ar.terminal = true
		ar.symbol = nodeRefStr

		// get index from $num... sequence at end of ref str
		allSplits := strings.Split(nodeRefStr, "$")
		if len(allSplits) > 1 {
			lastSplit := allSplits[len(allSplits)-1]
			firstSplits := strings.Join(allSplits[:len(allSplits)-1], "$")

			var err error
			ar.occurance, err = strconv.Atoi(lastSplit)
			if err != nil {
				// not an error, it's optional
				ar.occurance = 0
			} else {
				ar.symbol = firstSplits
			}
		}

		return ar, nil
	}
}

func (ar AttrRef) String() string {
	var sb strings.Builder

	if ar.head {
		sb.WriteString("{^}")
	} else if ar.termInProd {
		sb.WriteString("{.")
		if ar.occurance > 0 {
			sb.WriteString(fmt.Sprintf("%d", ar.occurance))
		}
		sb.WriteString("}")
	} else if ar.nontermInProd {
		sb.WriteString("{&")
		if ar.occurance > 0 {
			sb.WriteString(fmt.Sprintf("%d", ar.occurance))
		}
		sb.WriteString("}")
	} else if ar.symInProd {
		sb.WriteString("{")
		sb.WriteString(fmt.Sprintf("%d", ar.occurance))
		sb.WriteString("}")
	} else if ar.terminal {
		sb.WriteString(ar.symbol)
		if ar.occurance > 0 {
			sb.WriteString(fmt.Sprintf("$%d", ar.occurance))
		}
	} else {
		sb.WriteString("{")
		sb.WriteString(ar.symbol)
		if ar.occurance > 0 {
			sb.WriteString(fmt.Sprintf("$%d", ar.occurance))
		}
		sb.WriteString("}")
	}

	sb.WriteRune('.')
	sb.WriteString(ar.attribute)

	return sb.String()
}

type semanticAction struct {
	lhs  AttrRef
	hook string
	with []AttrRef

	hookTok types.Token
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

	sb.WriteString("prod ")
	if pa.prodNext {
		sb.WriteString("(next)")
	} else if pa.prodLiteral != nil {
		sb.WriteRune('[')
		if len(pa.prodLiteral) == 1 && pa.prodLiteral[0] == grammar.Epsilon[0] {
			sb.WriteString("ε")
		} else {
			sb.WriteString(strings.Join(pa.prodLiteral, " "))
		}
		sb.WriteRune(']')
	} else {
		sb.WriteString(fmt.Sprintf("(index %d)", pa.prodIndex))
	}

	sb.WriteString(": ")
	for i := range pa.actions {
		sb.WriteString(pa.actions[i].String())
		if i+1 < len(pa.actions) {
			sb.WriteString("; ")
		}
	}
	return sb.String()
}

type symbolActions struct {
	symbol  string
	actions []productionAction
}

func (sa symbolActions) String() string {
	var sb strings.Builder

	sb.WriteString(sa.symbol)
	sb.WriteString(": [")

	for i := range sa.actions {
		sb.WriteString(sa.actions[i].String())
		if i+1 < len(sa.actions) {
			sb.WriteString(", ")
		}
	}

	sb.WriteRune(']')

	return sb.String()
}
