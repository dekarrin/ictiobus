package translation

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/stack"
	"github.com/dekarrin/ictiobus/types"
)

// TODO: merge this with the ParseTree version
const (
	treeLevelEmpty               = "        "
	treeLevelOngoing             = "  |     "
	treeLevelPrefix              = "  |%s: "
	treeLevelPrefixLast          = `  \%s: `
	treeLevelPrefixNamePadChar   = '-'
	treeLevelPrefixNamePadAmount = 3
)

func makeTreeLevelPrefix(msg string) string {
	for len([]rune(msg)) < treeLevelPrefixNamePadAmount {
		msg = string(treeLevelPrefixNamePadChar) + msg
	}
	return fmt.Sprintf(treeLevelPrefix, msg)
}

func makeTreeLevelPrefixLast(msg string) string {
	for len([]rune(msg)) < treeLevelPrefixNamePadAmount {
		msg = string(treeLevelPrefixNamePadChar) + msg
	}
	return fmt.Sprintf(treeLevelPrefixLast, msg)
}

type AnnotatedParseTree struct {
	// Terminal is whether this node is for a terminal symbol.
	Terminal bool

	// Symbol is the symbol at this node's head.
	Symbol string

	// Source is only available when Terminal is true.
	Source types.Token

	// Children is all children of the parse tree.
	Children []*AnnotatedParseTree

	// Attributes is the data for attributes at the given position in the parse
	// tree.
	Attributes NodeAttrs
}

// AddAttributes adds annotation fields to the given parse tree. Returns an
// AnnotatedParseTree with only auto fields set ('$text' for terminals, '$id'
// for all nodes).
func AddAttributes(root types.ParseTree) AnnotatedParseTree {
	treeStack := stack.Stack[*types.ParseTree]{Of: []*types.ParseTree{&root}}
	annoRoot := AnnotatedParseTree{}
	annotatedStack := stack.Stack[*AnnotatedParseTree]{Of: []*AnnotatedParseTree{&annoRoot}}

	idGen := NewIDGenerator(0)

	for treeStack.Len() > 0 {
		curTreeNode := treeStack.Pop()
		curAnnoNode := annotatedStack.Pop()

		curAnnoNode.Terminal = curTreeNode.Terminal
		curAnnoNode.Symbol = curTreeNode.Value
		curAnnoNode.Source = curTreeNode.Source
		curAnnoNode.Children = make([]*AnnotatedParseTree, len(curTreeNode.Children))
		curAnnoNode.Attributes = NodeAttrs{
			string("$id"): idGen.Next(),
		}

		if curTreeNode.Terminal {
			curAnnoNode.Attributes[string("$text")] = curAnnoNode.Source.Lexeme()
		}

		// put child nodes on stack in reverse order to get left-first
		for i := len(curTreeNode.Children) - 1; i >= 0; i-- {
			newAnnoNode := &AnnotatedParseTree{}
			curAnnoNode.Children[i] = newAnnoNode
			treeStack.Push(curTreeNode.Children[i])
			annotatedStack.Push(newAnnoNode)
		}
	}

	return annoRoot
}

func (apt AnnotatedParseTree) String() string {
	return apt.leveledStr("", "")
}

func (apt AnnotatedParseTree) leveledStr(firstPrefix, contPrefix string) string {
	var sb strings.Builder

	sb.WriteString(firstPrefix)
	if apt.Terminal {
		sb.WriteString(fmt.Sprintf("(%s: %s = %q)", apt.ID().String(), apt.Symbol, apt.Source.Lexeme()))
	} else {
		sb.WriteString(fmt.Sprintf("(%s: %s )", apt.ID().String(), apt.Symbol))
	}

	for i := range apt.Children {
		sb.WriteRune('\n')
		var leveledFirstPrefix string
		var leveledContPrefix string
		if i+1 < len(apt.Children) {
			leveledFirstPrefix = contPrefix + makeTreeLevelPrefix("")
			leveledContPrefix = contPrefix + treeLevelOngoing
		} else {
			leveledFirstPrefix = contPrefix + makeTreeLevelPrefixLast("")
			leveledContPrefix = contPrefix + treeLevelEmpty
		}
		itemOut := apt.Children[i].leveledStr(leveledFirstPrefix, leveledContPrefix)
		sb.WriteString(itemOut)
	}

	return sb.String()
}

// Returns the ID of this node in the parse tree. All nodes have an ID
// accessible via the special predefined attribute '$id'; this function serves
// as a shortcut to getting the value from the node attributes with casting and
// sanity checking handled.
//
// If for whatever reason the ID has not been set on this node, IDZero is
// returned.
func (apt AnnotatedParseTree) ID() APTNodeID {
	var id APTNodeID
	untyped, ok := apt.Attributes["$id"]
	if !ok {
		return id
	}

	id, ok = untyped.(APTNodeID)
	if !ok {
		panic(fmt.Sprintf("$id attribute set to non-APTNodeID typed value: %v", untyped))
	}

	return id
}

// Rule returns the head and production of the grammar rule associated with the
// creation of this node in the parse tree. If apt is for a terminal, prod will
// be empty.
func (apt AnnotatedParseTree) Rule() (head string, prod []string) {
	if apt.Terminal {
		return apt.Symbol, nil
	}

	// need to gather symbol names from created nodes
	prod = []string{}
	for i := range apt.Children {
		prod = append(prod, apt.Children[i].Symbol)
	}

	return apt.Symbol, prod
}

// SymbolOf returns the symbol of the node referred to by rel. Additionally, a
// second 'ok' value is returned that specifies whether a node matches rel. Iff
// the second value is false, the first value should not be relied on.
func (apt AnnotatedParseTree) SymbolOf(rel NodeRelation) (symbol string, ok bool) {
	node, ok := apt.RelativeNode(rel)
	if !ok {
		return "", false
	}
	return node.Symbol, true
}

// AttributeValueOf returns the value of the named attribute in the node
// referred to by ref. Additionally, a second 'ok' value is returned that
// specifies whether ref refers to an existing attribute in the node whose
// relation to apt matches that specified in ref; if the returned 'ok' value is
// false, val should be considered a nil value and unsafe to use.
func (apt AnnotatedParseTree) AttributeValueOf(ref AttrRef) (val interface{}, ok bool) {
	// first get the attributes
	attributes, ok := apt.AttributesOf(ref.Relation)
	if !ok {
		return nil, false
	}

	attrVal, ok := attributes[ref.Name]
	return attrVal, ok
}

// RelativeNode returns the node pointed to by rel. Specifically, it returns the
// node that is related to apt in the way specified by rel, which can be at most
// one node as per the definition of rel's type.
//
// RelHead will cause apt itself to be returned; all others select a child node.
//
// A second 'ok' value is returned. This value is true if rel is a relation that
// exists in apt. If rel specifies a node that does not exist, the ok value will
// be false and the returned related node should not be used.
func (apt AnnotatedParseTree) RelativeNode(rel NodeRelation) (related *AnnotatedParseTree, ok bool) {
	if rel.Type == RelHead {
		return &apt, true
	} else if rel.Type == RelSymbol {
		symIdx := rel.Index
		if symIdx >= len(apt.Children) {
			return nil, false
		}
		return apt.Children[symIdx], true
	} else if rel.Type == RelNonTerminal {
		searchNonTermIdx := rel.Index

		// find the nth non-terminal
		curNonTermIdx := -1
		foundIdx := -1
		for i := range apt.Children {
			childNode := apt.Children[i]

			if childNode.Terminal {
				continue
			} else {
				curNonTermIdx++
				if curNonTermIdx == searchNonTermIdx {
					foundIdx = i
					break
				}
			}
		}
		if foundIdx == -1 {
			return nil, false
		}
		return apt.Children[foundIdx], true
	} else if rel.Type == RelTerminal {
		searchTermIdx := rel.Index

		// find the nth non-terminal
		curTermIdx := -1
		foundIdx := -1
		for i := range apt.Children {
			childNode := apt.Children[i]

			if !childNode.Terminal {
				continue
			} else {
				curTermIdx++
				if curTermIdx == searchTermIdx {
					foundIdx = i
					break
				}
			}
		}
		if foundIdx == -1 {
			return nil, false
		}
		return apt.Children[foundIdx], true
	} else {
		// not a valid AttrRelNode, can't handle it
		return nil, false
	}
}

// AttributesOf gets the Attributes of the node referred to by the given
// AttrRelNode value. For valid relations (those for which apt has a match for
// among itself (for the head symbol) and children (for the produced symbols)),
// the Attributes of the specified node are returned, as well as a second 'ok'
// value which will be true. If rel specifies a node that doesn't exist relative
// to apt, then the second value will be false and the returned node attributes
// will be nil.
func (apt AnnotatedParseTree) AttributesOf(rel NodeRelation) (attributes NodeAttrs, ok bool) {
	node, ok := apt.RelativeNode(rel)
	if !ok {
		return nil, false
	}
	return node.Attributes, true
}
