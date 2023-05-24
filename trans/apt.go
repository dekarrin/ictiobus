package trans

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
)

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

// AnnotatedTree is a parse tree annotated with attributes at every node.
// These attributes are set by calling hook functions on other attributes, and
// doing so succesively is what eventually implements a complete syntax-directed
// translation.
//
// Some attributes are built in; these begin with a $. This includes $id, which
// is the ID of a node and is defined on all nodes; $ft, which is the first
// token lexed for a node and is defined for all nodes except for terminal
// nodes produced by an epsilon production; $text, which is the lexed text and
// is defined only for terminal symbol nodes.
//
// An AnnotatedTree can be created from a parse.Tree by calling [Annotate].
type AnnotatedTree struct {
	// Terminal is whether this node is for a terminal symbol.
	Terminal bool

	// Symbol is the symbol at this node's head.
	Symbol string

	// Source is only available when Terminal is true.
	Source lex.Token

	// Children is all children of the parse tree.
	Children []*AnnotatedTree

	// Attributes is the data for attributes at the given position in the parse
	// tree.
	Attributes nodeAttrs
}

// Annotate adds attribute fields to the given parse tree to convert it to an
// AnnotatedTree. Returns an AnnotatedTree with only auto fields set ('$text'
// for terminals, '$id' for all nodes, and '$ft' for all nodes except epsilon
// terminal nodes, representing the first Token of the expression).
func Annotate(root parse.Tree) AnnotatedTree {
	treeStack := box.NewStack([]*parse.Tree{&root})
	annoRoot := AnnotatedTree{}
	annotatedStack := box.NewStack([]*AnnotatedTree{&annoRoot})

	idGen := newIDGenerator(0)

	for treeStack.Len() > 0 {
		curTreeNode := treeStack.Pop()
		curAnnoNode := annotatedStack.Pop()

		curAnnoNode.Terminal = curTreeNode.Terminal
		curAnnoNode.Symbol = curTreeNode.Value
		curAnnoNode.Source = curTreeNode.Source
		curAnnoNode.Children = make([]*AnnotatedTree, len(curTreeNode.Children))
		curAnnoNode.Attributes = nodeAttrs{
			string("$id"): idGen.Next(),
		}

		if curTreeNode.Terminal {
			if curTreeNode.Value == "" {
				// epsilon productions are a special terminal that doesn't
				// *have* a lexed value, so leave as an empty string if epsilon
				curAnnoNode.Attributes[string("$text")] = ""
			} else {
				curAnnoNode.Attributes[string("$text")] = curAnnoNode.Source.Lexeme()
			}
		}

		// put child nodes on stack in reverse order to get left-first
		for i := len(curTreeNode.Children) - 1; i >= 0; i-- {
			newAnnoNode := &AnnotatedTree{}
			curAnnoNode.Children[i] = newAnnoNode
			treeStack.Push(curTreeNode.Children[i])
			annotatedStack.Push(newAnnoNode)
		}
	}

	// now that we have the tree, traverse it again to set $ft
	annotatedStack = box.NewStack([]*AnnotatedTree{&annoRoot})
	for annotatedStack.Len() > 0 {
		curAnnoNode := annotatedStack.Pop()

		// enshore $first is set by calling First()
		curAnnoNode.First()

		// put child nodes on stack in reverse order to get left-first
		for i := len(curAnnoNode.Children) - 1; i >= 0; i-- {
			annotatedStack.Push(curAnnoNode.Children[i])
		}
	}

	return annoRoot
}

// ATLeaf is a convenience function for creating a new AnnotatedTree that
// represents a terminal symbol. The Source token is set to t if one is
// provided. Note that t's type being ...Token is simply to make it optional;
// only the first such provided t is examined. The given id must be unique for
// the entire tree.
func ATLeaf(id uint64, term string, t ...lex.Token) *AnnotatedTree {
	pt := &AnnotatedTree{Terminal: true, Symbol: term, Attributes: nodeAttrs{"$id": aptNodeID(id), "$text": ""}}
	if len(t) > 0 {
		pt.Source = t[0]
		pt.Attributes["$text"] = t[0].Lexeme()
		pt.Attributes["$ft"] = nil

		// only set $ft to non-nil if it's not the epsilon terminal
		if pt.Attributes["$text"] != "" {
			pt.Attributes["$ft"] = t[0]
		}
	} else {
		pt.Attributes["$text"] = "dummy2"
		pt.Attributes["$ft"] = lex.NewToken(lex.MakeDefaultClass("dummy"), "dummy2", 7, 1, "dummy1 dummy2 dummy3")
	}
	return pt
}

// ATNode is a convenience function for creating a new AnnotatedTree that
// represents a non-terminal symbol with minimal properties. The given id must
// be unique for the entire tree.
func ATNode(id uint64, nt string, children ...*AnnotatedTree) *AnnotatedTree {
	pt := &AnnotatedTree{
		Terminal:   false,
		Symbol:     nt,
		Children:   children,
		Attributes: nodeAttrs{"$id": aptNodeID(id)},
	}
	// and also calculate $ft
	pt.First()

	return pt
}

// String returns a string representation of the APT.
func (apt AnnotatedTree) String() string {
	return apt.leveledStr("", "")
}

func (apt AnnotatedTree) leveledStr(firstPrefix, contPrefix string) string {
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

// Returns the first token of the expression represented by this node in the
// parse tree. All nodes have a first token accessible via the special
// predefined attribute '$ft'; this function serves as a shortcut to getting
// the value from the node attributes with casting and sanity checking handled.
//
// Epsilons are special and return nil. Non-terminals which only have an epsilon
// children also return nil.
//
// Call on pointer because it may update $first if not already set.
func (apt *AnnotatedTree) First() lex.Token {
	// epsilon is a not a token per-se
	if apt.Symbol == "" && apt.Terminal {
		apt.Attributes[string("$ft")] = nil
		return nil
	}

	untyped, ok := apt.Attributes[string("$ft")]

	// if we didn't have it, set it for future calls
	if !ok {
		if apt.Terminal {
			untyped = apt.Source
		} else {
			for i := range apt.Children {
				untyped = apt.Children[i].First()
				if untyped != nil {
					break
				}
			}

			// if all descendants are epsilons, set $ft and return nil
			if untyped == nil {
				apt.Attributes[string("$ft")] = nil
				return nil
			}
		}
		apt.Attributes[string("$ft")] = untyped
	}

	// don't try to do typecast if it's nil
	if untyped == nil {
		return nil
	}

	var first lex.Token
	first, ok = untyped.(lex.Token)
	if !ok {
		panic(fmt.Sprintf("$ft attribute set to non-Token typed value: %v", untyped))
	}

	return first
}

// ID returns the ID of this node in the parse tree. All nodes have an ID
// accessible via the special predefined attribute '$id'; this function serves
// as a shortcut to getting the value from the node attributes with casting and
// sanity checking handled.
//
// If for whatever reason the ID has not been set on this node, IDZero is
// returned.
func (apt AnnotatedTree) ID() aptNodeID {
	var id aptNodeID
	untyped, ok := apt.Attributes["$id"]
	if !ok {
		return id
	}

	id, ok = untyped.(aptNodeID)
	if !ok {
		panic(fmt.Sprintf("$id attribute set to non-APTNodeID typed value: %v", untyped))
	}

	return id
}

// Rule returns the head and production of the grammar rule associated with the
// creation of this node in the parse tree. If apt is for a terminal, prod will
// be empty.
func (apt AnnotatedTree) Rule() (head string, prod []string) {
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
func (apt AnnotatedTree) SymbolOf(rel NodeRelation) (symbol string, ok bool) {
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
func (apt AnnotatedTree) AttributeValueOf(ref AttrRef) (val interface{}, ok bool) {
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
func (apt AnnotatedTree) RelativeNode(rel NodeRelation) (related *AnnotatedTree, ok bool) {
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
func (apt AnnotatedTree) AttributesOf(rel NodeRelation) (attributes nodeAttrs, ok bool) {
	node, ok := apt.RelativeNode(rel)
	if !ok {
		return nil, false
	}
	return node.Attributes, true
}
