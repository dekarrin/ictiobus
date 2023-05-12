// Package syntax provides functions for building up an abstract syntax tree
// from a FISHI markdown file. It is the interface between the generated
// ictiobus compiler frontend for FISHI and the rest of the fishi package.
package syntax

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/types"
)

// AST is the a8stract syntax tree of a fishi spec.
type AST struct {

	// Nodes is the nodes that make up the AST. There will be one per top-level
	// FISHI section (%%grammar, %%tokens, %%actions) encountered in the
	// specification the AST represents.
	Nodes []Block
}

// String returns the string representation of an AST.
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

// BlockType is the type of a FISHI [Block].
type BlockType int

const (
	// BlockTypeError is an unrecognized type of FISHI block.
	BlockTypeError BlockType = iota

	// BlockTypeGrammar denotes a %%grammar section from a spec written in
	// FISHI.
	BlockTypeGrammar

	// BlockTypeTokens denotes a %%tokens section from a spec written in FISHI.
	BlockTypeTokens

	// BlockTypeActions denotes an %%actions sectoin from a spec written in
	// FISHI.
	BlockTypeActions
)

// Block is a main dividing section of a FISHI spec. It contains either grammar
// rules, token definitions, or syntax-directed translation rules for the
// language described by the spec it is associated with.
type Block interface {

	// Type returns the type of the Block.
	Type() BlockType

	// Grammar converts the Block into a GrammarBlock. Panics if the Block's
	// type is not BlockTypeGrammar.
	Grammar() GrammarBlock

	// Tokens converts the Block into a TokensBlock. Panics if the Block's type
	// is not BlockTypeTokens.
	Tokens() TokensBlock

	// Actions converts the Block into an ActionsBlock. Panics if the Block's
	// type is not BlockTypeActions.
	Actions() ActionsBlock
}

// ErrorBlock is a Block representing an unrecognized kind of FISHI section.
type ErrorBlock struct{}

// Type returns BlockTypeError.
func (errBlock ErrorBlock) Type() BlockType {
	return BlockTypeError
}

// Grammar panics immediately. It is included to implement Block.
func (errBlock ErrorBlock) Grammar() GrammarBlock {
	panic("not grammar-type block")
}

// Tokens panics immediately. It is included to implement Block.
func (errBlock ErrorBlock) Tokens() TokensBlock {
	panic("not tokens-type block")
}

// Actions panics immediately. It is included to implement Block.
func (errBlock ErrorBlock) Actions() ActionsBlock {
	panic("not actions-type block")
}

// String returns a string representation of the ErrorBlock.
func (errBlock ErrorBlock) String() string {
	return "<Block: ERR>"
}

// GrammarBlock contains the contents of a single block of grammar instructions
// from a FISHI spec. It is represented in FISHI as a %%grammar section.
type GrammarBlock struct {

	// Content is the content blocks that make up this section. There will be
	// one per state declared in the grammar section this GrammarBlock was
	// created from.
	Content []GrammarContent
}

// String returns a string representation of the GrammarBlock.
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

// Type returns BlockTypeGrammar.
func (agb GrammarBlock) Type() BlockType {
	return BlockTypeGrammar
}

// Grammar returns this GrammarBlock. It is included to implement Block.
func (agb GrammarBlock) Grammar() GrammarBlock {
	return agb
}

// Tokens panics immediately. It is included to implement Block.
func (agb GrammarBlock) Tokens() TokensBlock {
	panic("not tokens-type block")
}

// Actions panics immediately. It is included to implement Block.
func (agb GrammarBlock) Actions() ActionsBlock {
	panic("not actions-type block")
}

// ActionsBlock contains the contents of a single block of SDTS definition rules
// from a FISHI spec. It is represented in FISHI as an %%actions section.
type ActionsBlock struct {

	// Content is the content blocks that make up this section. There will be
	// one per state declared in the actions section this ActionsBlock was
	// created from.
	Content []ActionsContent
}

// String returns a string representation of the ActionsBlock.
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

// Type returns BlockTypeActions.
func (aab ActionsBlock) Type() BlockType {
	return BlockTypeActions
}

// Grammar panics immediately. It is included to implement Block.
func (aab ActionsBlock) Grammar() GrammarBlock {
	panic("not grammar-type block")
}

// Tokens panics immediately. It is included to implement Block.
func (aab ActionsBlock) Tokens() TokensBlock {
	panic("not tokens-type block")
}

// Actions returns this ActionsBlock. It is included to implement Block.
func (aab ActionsBlock) Actions() ActionsBlock {
	return aab
}

// TokensBlock contains the contents of a single block of token declarations
// from a FISHI spec. It is represented in FISHI as a %%tokens section.
type TokensBlock struct {

	// Content is the content blocks that make up this section. There will be
	// one per state declared in the tokens section this TokensBlock was
	// created from.
	Content []TokensContent
}

// String returns a string representation of the TokensBlock.
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

// Type returns BlockTypeTokens.
func (atb TokensBlock) Type() BlockType {
	return BlockTypeTokens
}

// Grammar panics immediately. It is included to implement Block.
func (atb TokensBlock) Grammar() GrammarBlock {
	panic("not grammar-type block")
}

// Tokens returns this TokensBlock. It is included to implement Block.
func (atb TokensBlock) Tokens() TokensBlock {
	return atb
}

// Actions panics immediately. It is included to implement Block.
func (atb TokensBlock) Actions() ActionsBlock {
	panic("not actions-type block")
}

// TokenOptionsType is the type of option that a TokenOption represents.
type TokenOptionType int

const (
	// TokenOptDiscard is a token option type indicating that a pattern found by
	// the lexer should be discarded. It is represented by the %discard
	// directive in FISHI source code.
	TokenOptDiscard TokenOptionType = iota

	// TokenOptStateshift is a token option type indicating that a pattern found
	// by the lexer should make it change to a new state. It is represented by
	// the %stateshift directive in FISHI source code.
	TokenOptStateshift

	// TokenOptToken is a token option type indicating that a pattern found by
	// the lexer should be lexed as a new token and passed to the parser. It is
	// represented by the %token directive in FISHI source code.
	TokenOptToken

	// TokenOptHuman is a token option type that gives the human readable name
	// for a lexed token. It is represented by the %human directive in FISHI
	// source code.
	TokenOptHuman

	// TokenOptPriority is a token option type indicating that a pattern should
	// be treated as a certain priority by the lexer. It is represented by the
	// %priority directive in FISHI source code.
	TokenOptPriority
)

// TokenOption is a directive associated with a pattern in a %%tokens block of a
// FISHI spec.
type TokenOption struct {
	// Type is the type of the TokenOption.
	Type TokenOptionType

	// Value is the string value of the option as lexed from a FISHI spec. Only
	// certain types of TokenOptions will have a value; for types that do not
	// accept a value, Value will be the empty string.
	Value string

	// Src is the token that represents this TokenOption as lexed from a FISHI
	// spec.
	Src types.Token
}

// TokenEntry is a single full entry from a %%tokens block of a FISHI spec. It
// includes the pattern for the lexer to recognize as well as options indicating
// what the lexer should do once that pattern is matched.
type TokenEntry struct {

	// Pattern is the pattern that the lexer must recognize before performing
	// the actions indicated by the options associated with that pattern.
	Pattern string

	// Discard is true if the entry contains a %discard directive.
	Discard bool

	// Shift is set to the value of the %stateshift directive in the entry. If
	// the entry does not contain one, Shift will be an empty string.
	Shift string

	// Token is set to the value of the %token directive in the entry. If the
	// entry does not contain one, Token will be an empty string.
	Token string

	// Human is set to the value of the %human directive in the entry. If the
	// entry does not contain one, Human will be an empty string.
	Human string

	// Priority is set to the value of the %priority directive in the entry. If
	// the entry does not contain one, Priority will be 0, although note that
	// this cannot be distinguished from a %priority directive set to 0 without
	// also consulting SrcPriority.
	Priority int

	// Src is the first token that represents a part of this TokenEntry as lexed
	// from a FISHI spec.
	Src types.Token

	// SrcDiscard is all first tokens of any %discard directives that are a part
	// of this TokenEntry as lexed from a FISHI spec.
	SrcDiscard []types.Token

	// SrcShift is all first tokens of any %stateshift directives that are a
	// part of this TokenEntry as lexed from a FISHI spec.
	SrcShift []types.Token

	// SrcToken is all first tokens of any %token directives that are a part of
	// this TokenEntry as lexed from a FISHI spec.
	SrcToken []types.Token

	// SrcHuman is all first tokens of any %human directives that are a part of
	// this TokenEntry as lexed from a FISHI spec.
	SrcHuman []types.Token

	// SrcPriority is all first tokens of any %priority directives that are a
	// part of this TokenEntry as lexed from a FISHI spec.
	SrcPriority []types.Token

	// (don't need a patternTok because that pattern is the first symbol and
	// there can only be one; tok will be the same as patternTok)
}

// String returns a string representation of the TokenEntry.
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

// GrammarRule is a single complete grammar rule from a %%grammar block of a
// FISHI spec. It includes the non-terminal symbol at the head of the rule, and
// one or more productions that can be derived from that non-terminal.
type GrammarRule struct {

	// Rule holds the non-terminal and all productions parsed for this
	// GrammarRule.
	Rule grammar.Rule

	// Src is the first token that represents a part of this GrammarRule as
	// lexed from a FISHI spec.
	Src types.Token
}

// String returns a string representation of the GrammarRule.
func (agr GrammarRule) String() string {
	return agr.Rule.String()
}

// TokensContent is a series of token entries grouped with the lexer state they
// are used in from a %%tokens section of a FISHI spec.
type TokensContent struct {
	// Entries is the token entries for the lexer state.
	Entries []TokenEntry

	// State is the lexer state that the Entries are defined for.
	State string

	// Src is the first token that represents a part of this TokensContent as
	// lexed from a FISHI spec.
	Src types.Token

	// SrcState is the first token that represents a part of the %state
	// directive that defines the state that this TokensContent is for. If it is
	// for the default state, this will be nil.
	SrcState types.Token
}

// String returns a string representation of the TokensContent.
func (content TokensContent) String() string {
	if len(content.Entries) > 0 {
		return fmt.Sprintf("(State: %q, Entries: %v)", content.State, content.Entries)
	} else {
		return fmt.Sprintf("(State: %q, Entries: (empty))", content.State)
	}
}

// GrammarContent is a series of grammar rules grouped with the state they are
// used in from a %%grammar section of a FISHI spec. Note that multiple states
// for a grammar are not supported, so State will always be the empty string.
type GrammarContent struct {

	// Rules is the rules in the GrammarContent.
	Rules []GrammarRule

	// State is the state that the rules apply to. It will always be the empty
	// string.
	State string

	// Src is the first token that represents a part of this GrammarContent as
	// lexed from a FISHI spec.
	Src types.Token

	// SrcState is the first token that represents a part of the %state
	// directive that defines the state that this GrammarContent is for. As
	// states for grammar sections other than the default are not supported,
	// this will always be nil.
	SrcState types.Token
}

// String returns a string representation of the GrammarContent.
func (content GrammarContent) String() string {
	if len(content.Rules) > 0 {
		return fmt.Sprintf("(State: %q, Rules: %v)", content.State, content.Rules)
	} else {
		return fmt.Sprintf("(State: %q, Rules: (empty))", content.State)
	}
}

// ActionsContent is a series of syntax-directed translation actions grouped
// with the state they are used in from an %%actions section of a FISHI spec.
// Note that multiple for a syntax-directed translation scheme are not
// supported, so State will always be the empty string.
type ActionsContent struct {
	// Actions is a series of SDTS actions that each apply to a given head
	// symbol of a grammar rule.
	Actions []SymbolActions

	// State is the state that the actions apply to. It will always be the empty
	// string.
	State string

	// Src is the first token that represents a part of this ActionsContent as
	// lexed from a FISHI spec.
	Src types.Token

	// SrcState is the first token that represents a part of the %state
	// directive that defines the state that this ActionsContent is for. As
	// states for actions sections other than the default are not supported,
	// this will always be nil.
	SrcState types.Token
}

// String returns a string representation of the ActionsContent.
func (content ActionsContent) String() string {
	if len(content.Actions) > 0 {
		return fmt.Sprintf("(State: %q, Actions: %v)", content.State, content.Actions)
	} else {
		return fmt.Sprintf("(State: %q, Actions: (empty))", content.State)
	}
}

// AttrRef is a reference to an attribute of a particular symbol in a grammar
// rule production. It consists of two parts; the symbol it refers to, and the
// name of the attribute. An AttrRef has five different ways it may refer to a
// symbol: The head symbol, the nth symbol in the production, the nth
// non-terminal symbol in the production, the nth terminal symbol in the
// production, or the nth instance of a symbol with a particular name in the
// production (with whether or not the symbol name refers to a terminal
// explicitly denoted).
type AttrRef struct {
	// Symbol is the symbol name included in the AttrRef in a FISHI spec. This
	// will only be set if the AttrRef refers to a particular symbol by name;
	// otherwise, Symbol will be set to the empty string.
	Symbol string

	// Terminal is whether Symbol refers to a terminal symbol.
	Terminal bool

	// Head is whether the AttrRef refers to the Head symbol.
	Head bool

	// TermInProd is whether the AttrRef refers to the nth terminal symbol in
	// the production. If true, Occurance is n.
	TermInProd bool

	// TermInProd is whether the AttrRef refers to the nth non-terminal symbol
	// in the production. If true, Occurance is n.
	NontermInProd bool

	// TermInProd is whether the AttrRef refers to the nth symbol in the
	// production. If true, Occurance is n.
	SymInProd bool

	// Occurance is the index of the reference, and represents n when the
	// AttrRef refers to the nth occurance of some criteria. It is not valid if
	// Head is true.
	Occurance int

	// Attribute is the name of the attribute being referred to.
	Attribute string

	// Src is the first token that represents a part of this AttrRef as lexed
	// from a FISHI spec.
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

// String returns a string representation of the AttrRef.
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

// SemanticAction is a single syntax-directed action to perform. It takes some
// arguments from symbols in the grammar rule it is defined on, passes those to
// a hook function, and assigns the result to the attribute of another symbol
// in the node in the parse tree it is called on.
type SemanticAction struct {
	// LHS is the left-hand side of the action. It is a reference to the
	// attribute and symbol node it should assign the result of the action to.
	LHS AttrRef

	// Hook is the name of the hook function to call.
	Hook string

	// With is references to the attributes whose values should be used as
	// arguments to the hook function.
	With []AttrRef

	// Src is the first token that represents the name of the hook as
	// lexed from a FISHI spec.
	SrcHook types.Token

	// Src is the first token that represents a part of this SemanticAction as
	// lexed from a FISHI spec.
	Src types.Token
}

// String returns a string representation of the SemanticAction.
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

// ProductionAction is a series of syntax-directed definitions defined for a
// production of a non-terminal symbol.
type ProductionAction struct {
	// ProdNext is whether the production referred to is left unspecified, ergo
	// is the 'next' production after the last one (or the first production, if
	// this is the first ProductionAction for the symbol).
	ProdNext bool

	// ProdIndex is the index of the production within all productions of the
	// symbol that this action is for.
	ProdIndex int

	// ProdLiteral is the literal symbols in the production of the symbol that
	// this action is for.
	ProdLiteral []string

	// Actions is the actions to perform when the production specified by this
	// ProductionAction is encountered during syntax-directed translation.
	Actions []SemanticAction

	// Src is the first token that represents a part of this ProductionAction as
	// lexed from a FISHI spec.
	Src types.Token

	// SrcVal is where the production action "value" is set; that is, the index
	// or production. It will be nil if it is simply a prodNext.
	SrcVal types.Token
}

// String returns a string representation of the ProductionAction.
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

// SymbolActions is a series of SDTS actions defined for productions of a
// non-terminal symbol.
type SymbolActions struct {
	// Symbol is the non-terminal that the Actions are defined for.
	Symbol string

	// Actions is the actions for the productions of Symbol.
	Actions []ProductionAction

	// Src is the first token that represents a part of this SymbolActions as
	// lexed from a FISHI spec.
	Src types.Token

	// SrcSym is the first token that represents a part of the symbol as lexed
	// from a FISHI spec.
	SrcSym types.Token
}

// String returns a string representation of the SymbolActions.
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
