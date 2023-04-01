package ir

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/types"
)

type AST struct {
	Nodes []ASTBlock
}

func (ast AST) String() string {
	var sb strings.Builder

	sb.WriteRune('<')
	if len(ast.Nodes) > 0 {
		sb.WriteRune('\n')
		for i := range ast.Nodes {
			n := ast.Nodes[i]
			switch n.Type() {
			case BlockTypeError:
				sb.WriteString("  <ERR>\n")
			case BlockTypeGrammar:
				gram := n.Grammar()
				sb.WriteString("  <GRAMMAR:\n")
				for j := range gram.Content {
					cont := gram.Content[j]
					if cont.State != "" {
						sb.WriteString("    <RULE-SET FOR STATE " + fmt.Sprintf("%q\n", cont.State))
					} else {
						sb.WriteString("    <RULE-SET FOR ALL STATES\n")
					}
					for k := range cont.Rules {
						r := cont.Rules[k]
						sb.WriteString("      * " + r.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case BlockTypeTokens:
				toks := n.Tokens()
				sb.WriteString("  <TOKENS:\n")
				for j := range toks.Content {
					cont := toks.Content[j]
					if cont.State != "" {
						sb.WriteString("    <ENTRY-SET FOR STATE " + fmt.Sprintf("%q\n", cont.State))
					} else {
						sb.WriteString("    <ENTRY-SET FOR ALL STATES\n")
					}
					for k := range cont.Entries {
						entry := cont.Entries[k]
						sb.WriteString("      * " + entry.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case BlockTypeActions:
				acts := n.Actions()
				sb.WriteString("  <ACTIONS:\n")
				for j := range acts.Content {
					cont := acts.Content[j]
					if cont.State != "" {
						sb.WriteString("    <ACTION-SET FOR STATE " + fmt.Sprintf("%q\n", cont.State))
					} else {
						sb.WriteString("    <ACTION-SET FOR ALL STATES\n")
					}
					for k := range cont.Actions {
						action := cont.Actions[k]
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

type ASTBlockType int

const (
	BlockTypeError ASTBlockType = iota
	BlockTypeGrammar
	BlockTypeTokens
	BlockTypeActions
)

type ASTBlock interface {
	Type() ASTBlockType
	Grammar() ASTGrammarBlock
	Tokens() ASTTokensBlock
	Actions() ASTActionsBlock
}

type ASTErrorBlock bool

func (errBlock ASTErrorBlock) Type() ASTBlockType {
	return BlockTypeError
}

func (errBlock ASTErrorBlock) Grammar() ASTGrammarBlock {
	panic("not grammar-type block")
}

func (errBlock ASTErrorBlock) Tokens() ASTTokensBlock {
	panic("not tokens-type block")
}

func (errBlock ASTErrorBlock) Actions() ASTActionsBlock {
	panic("not actions-type block")
}

func (errBlock ASTErrorBlock) String() string {
	return "<Block: ERR>"
}

type ASTGrammarBlock struct {
	Content []ASTGrammarContent
}

func (agb ASTGrammarBlock) String() string {
	var sb strings.Builder

	sb.WriteString("<Block: GRAMMAR, Content: {")
	for i := range agb.Content {
		sb.WriteString(agb.Content[i].String())
		if i+1 < len(agb.Content) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune('}')
	return sb.String()
}

func (agb ASTGrammarBlock) Type() ASTBlockType {
	return BlockTypeGrammar
}

func (agb ASTGrammarBlock) Grammar() ASTGrammarBlock {
	return agb
}

func (agb ASTGrammarBlock) Tokens() ASTTokensBlock {
	panic("not tokens-type block")
}

func (agb ASTGrammarBlock) Actions() ASTActionsBlock {
	panic("not actions-type block")
}

type ASTActionsBlock struct {
	Content []ASTActionsContent
}

func (aab ASTActionsBlock) String() string {
	var sb strings.Builder

	sb.WriteString("<Block: GRAMMAR, Content: {")
	for i := range aab.Content {
		sb.WriteString(aab.Content[i].String())
		if i+1 < len(aab.Content) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune('}')
	return sb.String()
}

func (aab ASTActionsBlock) Type() ASTBlockType {
	return BlockTypeActions
}

func (aab ASTActionsBlock) Grammar() ASTGrammarBlock {
	panic("not grammar-type block")
}

func (aab ASTActionsBlock) Tokens() ASTTokensBlock {
	panic("not tokens-type block")
}

func (aab ASTActionsBlock) Actions() ASTActionsBlock {
	return aab
}

type ASTTokensBlock struct {
	Content []ASTTokensContent
}

func (atb ASTTokensBlock) String() string {
	var sb strings.Builder

	sb.WriteString("<Block: TOKENS, Content: {")
	for i := range atb.Content {
		sb.WriteString(atb.Content[i].String())
		if i+1 < len(atb.Content) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune('}')
	return sb.String()
}

func (atb ASTTokensBlock) Type() ASTBlockType {
	return BlockTypeTokens
}

func (atb ASTTokensBlock) Grammar() ASTGrammarBlock {
	panic("not grammar-type block")
}

func (atb ASTTokensBlock) Tokens() ASTTokensBlock {
	return atb
}

func (atb ASTTokensBlock) Actions() ASTActionsBlock {
	panic("not actions-type block")
}

type ASTTokenOptionType int

const (
	TokenOptDiscard ASTTokenOptionType = iota
	TokenOptStateshift
	TokenOptToken
	TokenOptHuman
	TokenOptPriority
)

type ASTTokenOption struct {
	Type  ASTTokenOptionType
	Value string

	Src types.Token
}

type ASTTokenEntry struct {
	Pattern  string
	Discard  bool
	Shift    string
	Token    string
	Human    string
	Priority int

	Src types.Token

	// in theory may be multiple options for the same option type; while it
	// is not allowed semantically, it is allowed syntactically, so track it so
	// we can do proper error reporting later.
	SrcDiscard  []types.Token
	SrcShift    []types.Token
	SrcToken    []types.Token
	SrcHuman    []types.Token
	SrcPriority []types.Token

	// (don't need a patternTok because that pattern is the first symbol and
	// there can only be one; tok will be the same as patternTok)
}

func (entry ASTTokenEntry) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s -> ", entry.Pattern))
	sb.WriteString(fmt.Sprintf("Discard: %v, ", entry.Discard))
	sb.WriteString(fmt.Sprintf("Shift: %q, ", entry.Shift))
	sb.WriteString(fmt.Sprintf("Token: %q, ", entry.Token))
	sb.WriteString(fmt.Sprintf("Human: %q, ", entry.Human))
	sb.WriteString(fmt.Sprintf("Priority: %d", entry.Priority))

	return sb.String()
}

type ASTGrammarRule struct {
	Rule grammar.Rule

	Src types.Token
}

func (agr ASTGrammarRule) String() string {
	return agr.Rule.String()
}

type ASTTokensContent struct {
	Entries []ASTTokenEntry
	State   string

	Src      types.Token
	SrcState types.Token
}

func (content ASTTokensContent) String() string {
	if len(content.Entries) > 0 {
		return fmt.Sprintf("(State: %q, Entries: %v)", content.State, content.Entries)
	} else {
		return fmt.Sprintf("(State: %q, Entries: (empty))", content.State)
	}
}

type ASTGrammarContent struct {
	Rules []ASTGrammarRule
	State string

	Src      types.Token
	SrcState types.Token
}

func (content ASTGrammarContent) String() string {
	if len(content.Rules) > 0 {
		return fmt.Sprintf("(State: %q, Rules: %v)", content.State, content.Rules)
	} else {
		return fmt.Sprintf("(State: %q, Rules: (empty))", content.State)
	}
}

type ASTActionsContent struct {
	Actions []ASTSymbolActions
	State   string

	Src      types.Token
	SrcState types.Token
}

func (content ASTActionsContent) String() string {
	if len(content.Actions) > 0 {
		return fmt.Sprintf("(State: %q, Actions: %v)", content.State, content.Actions)
	} else {
		return fmt.Sprintf("(State: %q, Actions: (empty))", content.State)
	}
}

type ASTAttrRef struct {
	Symbol   string
	Terminal bool

	Head          bool
	TermInProd    bool
	NontermInProd bool
	SymInProd     bool

	Occurance int
	Attribute string

	Src types.Token
}

// ParseAttrRef does a simple parse on an attribute reference from a string that
// makes it up. Does not set tok; caller must do so if needed.
func ParseAttrRef(s string) (ASTAttrRef, error) {
	dotSpl := strings.Split(s, ".")
	if len(dotSpl) < 2 {
		return ASTAttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
	}

	attrName := dotSpl[len(dotSpl)-1]
	nodeRefStr := strings.Join(dotSpl[:len(dotSpl)-1], ".")

	ar := ASTAttrRef{Attribute: attrName}

	if nodeRefStr[0] == '{' && nodeRefStr[len(nodeRefStr)-1] == '}' {
		str := nodeRefStr[1 : len(nodeRefStr)-1]
		if (str[0] >= 'A' && str[0] <= 'Z') || (str[0] >= 'a' && str[0] <= 'z') {
			// nonterminal-by-name reference
			ar.Symbol = str

			// get index from $num... sequence at end of ref str
			allSplits := strings.Split(nodeRefStr, "$")
			if len(allSplits) > 1 {
				lastSplit := allSplits[len(allSplits)-1]
				firstSplits := strings.Join(allSplits[:len(allSplits)-1], "$")

				var err error
				ar.Occurance, err = strconv.Atoi(lastSplit)
				if err != nil {
					// not an error, it's optional
					ar.Occurance = 0
				} else {
					ar.Symbol = firstSplits
				}
			}
			return ar, nil
		} else if str == "^" {
			ar.Head = true
			return ar, nil
		} else if strings.HasPrefix(str, ".") {
			ar.TermInProd = true
			if len(str) > 1 {
				str = str[1:]
				num, err := strconv.Atoi(str)
				if err != nil {
					return ASTAttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
				}
				ar.Occurance = num
			}
			return ar, nil
		} else if strings.HasPrefix(str, "&") {
			ar.NontermInProd = true
			if len(str) > 1 {
				str = str[1:]
				num, err := strconv.Atoi(str)
				if err != nil {
					return ASTAttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
				}
				ar.Occurance = num
			}
			return ar, nil
		} else {
			// then it has to be a parsable number
			num, err := strconv.Atoi(str)
			if err != nil {
				return ASTAttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
			}
			ar.Occurance = num
			ar.SymInProd = true
			return ar, nil
		}
	} else {
		// terminal-by-name reference
		ar.Terminal = true
		ar.Symbol = nodeRefStr

		// get index from $num... sequence at end of ref str
		allSplits := strings.Split(nodeRefStr, "$")
		if len(allSplits) > 1 {
			lastSplit := allSplits[len(allSplits)-1]
			firstSplits := strings.Join(allSplits[:len(allSplits)-1], "$")

			var err error
			ar.Occurance, err = strconv.Atoi(lastSplit)
			if err != nil {
				// not an error, it's optional
				ar.Occurance = 0
			} else {
				ar.Symbol = firstSplits
			}
		}

		return ar, nil
	}
}

func (ar ASTAttrRef) String() string {
	var sb strings.Builder

	if ar.Head {
		sb.WriteString("{^}")
	} else if ar.TermInProd {
		sb.WriteString("{.")
		if ar.Occurance > 0 {
			sb.WriteString(fmt.Sprintf("%d", ar.Occurance))
		}
		sb.WriteString("}")
	} else if ar.NontermInProd {
		sb.WriteString("{&")
		if ar.Occurance > 0 {
			sb.WriteString(fmt.Sprintf("%d", ar.Occurance))
		}
		sb.WriteString("}")
	} else if ar.SymInProd {
		sb.WriteString("{")
		sb.WriteString(fmt.Sprintf("%d", ar.Occurance))
		sb.WriteString("}")
	} else if ar.Terminal {
		sb.WriteString(ar.Symbol)
		if ar.Occurance > 0 {
			sb.WriteString(fmt.Sprintf("$%d", ar.Occurance))
		}
	} else {
		sb.WriteString("{")
		sb.WriteString(ar.Symbol)
		if ar.Occurance > 0 {
			sb.WriteString(fmt.Sprintf("$%d", ar.Occurance))
		}
		sb.WriteString("}")
	}

	sb.WriteRune('.')
	sb.WriteString(ar.Attribute)

	return sb.String()
}

type ASTSemanticAction struct {
	LHS  ASTAttrRef
	Hook string
	With []ASTAttrRef

	SrcHook types.Token
	Src     types.Token
}

func (sa ASTSemanticAction) String() string {
	var sb strings.Builder

	sb.WriteString(sa.LHS.String())
	sb.WriteString(" = ")
	sb.WriteString(sa.Hook)

	sb.WriteRune('(')
	for i := range sa.With {
		sb.WriteString(sa.With[i].String())
		if i+1 < len(sa.With) {
			sb.WriteString(", ")
		}
	}
	sb.WriteRune(')')

	return sb.String()
}

type ASTProductionAction struct {
	ProdNext    bool
	ProdIndex   int
	ProdLiteral []string

	Actions []ASTSemanticAction

	Src types.Token

	// SrcVal is where the production action "value" is set; that is, the index
	// or production. It will be nil if it is simply a prodNext.
	SrcVal types.Token
}

func (pa ASTProductionAction) String() string {
	var sb strings.Builder

	sb.WriteString("prod ")
	if pa.ProdNext {
		sb.WriteString("(next)")
	} else if pa.ProdLiteral != nil {
		sb.WriteRune('[')
		if len(pa.ProdLiteral) == 1 && pa.ProdLiteral[0] == grammar.Epsilon[0] {
			sb.WriteString("ε")
		} else {
			sb.WriteString(strings.Join(pa.ProdLiteral, " "))
		}
		sb.WriteRune(']')
	} else {
		sb.WriteString(fmt.Sprintf("(index %d)", pa.ProdIndex))
	}

	sb.WriteString(": ")
	for i := range pa.Actions {
		sb.WriteString(pa.Actions[i].String())
		if i+1 < len(pa.Actions) {
			sb.WriteString("; ")
		}
	}
	return sb.String()
}

type ASTSymbolActions struct {
	Symbol  string
	Actions []ASTProductionAction

	Src    types.Token
	SrcSym types.Token
}

func (sa ASTSymbolActions) String() string {
	var sb strings.Builder

	sb.WriteString(sa.Symbol)
	sb.WriteString(": [")

	for i := range sa.Actions {
		sb.WriteString(sa.Actions[i].String())
		if i+1 < len(sa.Actions) {
			sb.WriteString(", ")
		}
	}

	sb.WriteRune(']')

	return sb.String()
}
