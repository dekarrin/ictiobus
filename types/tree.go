package types

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/stack"
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

type ParseTree struct {
	// Terminal is whether thie node is for a terminal symbol.
	Terminal bool

	// Value is the symbol at this node.
	Value string

	// Source is only available when Terminal is true.
	Source Token

	// Children is all children of the parse tree.
	Children []*ParseTree
}

// String returns a prettified representation of the entire parse tree suitable
// for use in line-by-line comparisons of tree structure. Two parse trees are
// considered semantcally identical if they produce identical String() output.
func (pt ParseTree) String() string {
	return pt.leveledStr("", "")
}

// Copy returns a duplicate, deeply-copied parse tree.
func (pt ParseTree) Copy() ParseTree {
	newPt := ParseTree{
		Terminal: pt.Terminal,
		Value:    pt.Value,
		Source:   pt.Source,
		Children: make([]*ParseTree, len(pt.Children)),
	}

	for i := range pt.Children {
		if pt.Children[i] != nil {
			newChild := pt.Children[i].Copy()
			newPt.Children[i] = &newChild
		}
	}

	return newPt
}

// Equal returns whether the parseTree is equal to the given object. If the
// given object is not a parseTree, returns false, else returns whether the two
// parse trees have the exact same structure.
//
// Does not consider the Source field, ergo only the structures of the trees are
// compared, not their contents.
//
// Runs in O(n) time with respect to the number of nodes in the trees.
func (pt ParseTree) Equal(o any) bool {
	other, ok := o.(ParseTree)
	if !ok {
		// also okay if its the pointer value, as long as its non-nil
		otherPtr, ok := o.(*ParseTree)
		if !ok {
			return false
		} else if otherPtr == nil {
			return false
		}
		other = *otherPtr
	}

	if pt.Terminal != other.Terminal {
		return false
	} else if pt.Value != other.Value {
		return false
	} else {
		// check every sub tree
		if len(pt.Children) != len(other.Children) {
			return false
		}

		for i := range pt.Children {
			if !pt.Children[i].Equal(other.Children[i]) {
				return false
			}
		}
	}
	return true
}

// Checks if the given ParseTree contains sub as a sub-tree. Does not consider
// Source for its comparisons, ergo only the structure is examined.
//
// This performs a depth-first traversal of the parse tree, checking if sub is
// equal at every point. Runs in O(n^2) time with respect to the number of nodes
// in the trees.
//
// Returns whether sub is a sub-tree of pt, and if so, the path to the first
// node in pt where this is the case. The path is represented as a slice of ints
// where each is the child index of the node to traverse to. If it is empty,
// then the root node is the first node where sub is a sub-tree; this is not
// necessarily the same as equality.
func (pt ParseTree) ContainsSubTree(sub ParseTree) (isSubTree bool, path []int) {
	type pair struct {
		node *ParseTree
		path []int
	}

	checkStack := stack.Stack[pair]{}
	checkStack.Push(pair{&pt, []int{}})

	for !checkStack.Empty() {
		p := checkStack.Pop()
		startNode := p.node
		path := p.path

		// add any node but the root to the path.
		if idx != -1 {
			buildingPath = append(buildingPath, idx)
		}

		if startNode.Equal(sub) {
			return true, path
		}

		for i := len(startNode.Children) - 1; i >= 0; i-- {
			checkStack.Push(startNode.Children[i])
		}
	}

	return false
}

func (pt ParseTree) leveledStr(firstPrefix, contPrefix string) string {
	var sb strings.Builder

	sb.WriteString(firstPrefix)
	if pt.Terminal {
		sb.WriteString(fmt.Sprintf("(TERM %q)", pt.Value))
	} else {
		sb.WriteString(fmt.Sprintf("( %s )", pt.Value))
	}

	for i := range pt.Children {
		sb.WriteRune('\n')
		var leveledFirstPrefix string
		var leveledContPrefix string
		if i+1 < len(pt.Children) {
			leveledFirstPrefix = contPrefix + makeTreeLevelPrefix("")
			leveledContPrefix = contPrefix + treeLevelOngoing
		} else {
			leveledFirstPrefix = contPrefix + makeTreeLevelPrefixLast("")
			leveledContPrefix = contPrefix + treeLevelEmpty
		}
		itemOut := pt.Children[i].leveledStr(leveledFirstPrefix, leveledContPrefix)
		sb.WriteString(itemOut)
	}

	return sb.String()
}
