package syntax

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/types"
)

type BlockType int

const (
	BlockTypeError BlockType = iota
	BlockTypeGrammar
	BlockTypeTokens
	BlockTypeActions
)

type Block interface {
	Type() BlockType
	Grammar() GrammarBlock
	Tokens() TokensBlock
	Actions() ActionsBlock
}

type ErrorBlock bool

func (errBlock ErrorBlock) Type() BlockType {
	return BlockTypeError
}

func (errBlock ErrorBlock) Grammar() GrammarBlock {
	panic("not grammar-type block")
}

func (errBlock ErrorBlock) Tokens() TokensBlock {
	panic("not tokens-type block")
}

func (errBlock ErrorBlock) Actions() ActionsBlock {
	panic("not actions-type block")
}

func (errBlock ErrorBlock) String() string {
	return "<Block: ERR>"
}

type GrammarBlock struct {
	Content []GrammarContent
}

func (agb GrammarBlock) String() string {
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

func (agb GrammarBlock) Type() BlockType {
	return BlockTypeGrammar
}

func (agb GrammarBlock) Grammar() GrammarBlock {
	return agb
}

func (agb GrammarBlock) Tokens() TokensBlock {
	panic("not tokens-type block")
}

func (agb GrammarBlock) Actions() ActionsBlock {
	panic("not actions-type block")
}

type ActionsBlock struct {
	Content []ActionsContent
}

func (aab ActionsBlock) String() string {
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

func (aab ActionsBlock) Type() BlockType {
	return BlockTypeActions
}

func (aab ActionsBlock) Grammar() GrammarBlock {
	panic("not grammar-type block")
}

func (aab ActionsBlock) Tokens() TokensBlock {
	panic("not tokens-type block")
}

func (aab ActionsBlock) Actions() ActionsBlock {
	return aab
}

type TokensBlock struct {
	Content []TokensContent
}

func (atb TokensBlock) String() string {
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

func (atb TokensBlock) Type() BlockType {
	return BlockTypeTokens
}

func (atb TokensBlock) Grammar() GrammarBlock {
	panic("not grammar-type block")
}

func (atb TokensBlock) Tokens() TokensBlock {
	return atb
}

func (atb TokensBlock) Actions() ActionsBlock {
	panic("not actions-type block")
}

type TokenOptionType int

const (
	TokenOptDiscard TokenOptionType = iota
	TokenOptStateshift
	TokenOptToken
	TokenOptHuman
	TokenOptPriority
)

type TokenOption struct {
	Type  TokenOptionType
	Value string

	Src types.Token
}

type TokenEntry struct {
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

func (entry TokenEntry) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s -> ", entry.Pattern))
	sb.WriteString(fmt.Sprintf("Discard: %v, ", entry.Discard))
	sb.WriteString(fmt.Sprintf("Shift: %q, ", entry.Shift))
	sb.WriteString(fmt.Sprintf("Token: %q, ", entry.Token))
	sb.WriteString(fmt.Sprintf("Human: %q, ", entry.Human))
	sb.WriteString(fmt.Sprintf("Priority: %d", entry.Priority))

	return sb.String()
}

type GrammarRule struct {
	Rule grammar.Rule

	Src types.Token
}

func (agr GrammarRule) String() string {
	return agr.Rule.String()
}

type TokensContent struct {
	Entries []TokenEntry
	State   string

	Src      types.Token
	SrcState types.Token
}

func (content TokensContent) String() string {
	if len(content.Entries) > 0 {
		return fmt.Sprintf("(State: %q, Entries: %v)", content.State, content.Entries)
	} else {
		return fmt.Sprintf("(State: %q, Entries: (empty))", content.State)
	}
}

type GrammarContent struct {
	Rules []GrammarRule
	State string

	Src      types.Token
	SrcState types.Token
}

func (content GrammarContent) String() string {
	if len(content.Rules) > 0 {
		return fmt.Sprintf("(State: %q, Rules: %v)", content.State, content.Rules)
	} else {
		return fmt.Sprintf("(State: %q, Rules: (empty))", content.State)
	}
}

type ActionsContent struct {
	Actions []SymbolActions
	State   string

	Src      types.Token
	SrcState types.Token
}

func (content ActionsContent) String() string {
	if len(content.Actions) > 0 {
		return fmt.Sprintf("(State: %q, Actions: %v)", content.State, content.Actions)
	} else {
		return fmt.Sprintf("(State: %q, Actions: (empty))", content.State)
	}
}

type AttrRef struct {
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
func ParseAttrRef(s string) (AttrRef, error) {
	dotSpl := strings.Split(s, ".")
	if len(dotSpl) < 2 {
		return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
	}

	attrName := dotSpl[len(dotSpl)-1]
	nodeRefStr := strings.Join(dotSpl[:len(dotSpl)-1], ".")

	ar := AttrRef{Attribute: attrName}

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
					return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
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
					return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
				}
				ar.Occurance = num
			}
			return ar, nil
		} else {
			// then it has to be a parsable number
			num, err := strconv.Atoi(str)
			if err != nil {
				return AttrRef{}, fmt.Errorf("invalid attribute reference: %q", s)
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

func (ar AttrRef) String() string {
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

type SemanticAction struct {
	LHS  AttrRef
	Hook string
	With []AttrRef

	SrcHook types.Token
	Src     types.Token
}

func (sa SemanticAction) String() string {
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

type ProductionAction struct {
	ProdNext    bool
	ProdIndex   int
	ProdLiteral []string

	Actions []SemanticAction

	Src types.Token

	// SrcVal is where the production action "value" is set; that is, the index
	// or production. It will be nil if it is simply a prodNext.
	SrcVal types.Token
}

func (pa ProductionAction) String() string {
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

type SymbolActions struct {
	Symbol  string
	Actions []ProductionAction

	Src    types.Token
	SrcSym types.Token
}

func (sa SymbolActions) String() string {
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
