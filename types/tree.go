package types

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
)

const (
	treeLevelEmpty               = "        "
	treeLevelOngoing             = "  |     "
	treeLevelPrefix              = "  |%s: "
	treeLevelPrefixLast          = `  \%s: `
	treeLevelPrefixNamePadChar   = '-'
	treeLevelPrefixNamePadAmount = 3

	ShortCircuitPrefix = "__ICTIO__:SHORTCIRC:"
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

// ParseTree is a parse tree returned by a parser performing analysis on input
// source code.
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

// MustParseTreeFromDiagram is the same as ParseTreeFromDiagram but panics if
// any error occurs.
func MustParseTreeFromDiagram(s string) *ParseTree {
	pt, err := ParseTreeFromDiagram(s)
	if err != nil {
		panic(err)
	}
	return pt
}

// ParseTreeFromDiagram reads a diagram of a parse tree and returns a ParseTree
// that represents it. In the diagram string s, terminal nodes are enclosed in
// parenthesis brackets, while non-terminal nodes are enclosed in square
// brackets. The diagram is read from left to right, and all whitespace is
// ignored. If a literal parenthesis or square bracket is desired, it must be
// escaped with a backslash. literal backslashes must be escaped with another
// backslash.
//
// For example, the following diagram:
//
//	 [S
//
//				  [NUM
//				    (-)
//				    (2)
//				  ]
//				  (+)
//				  [NUM
//					(3)
//				  ]
//
//		    ]
func ParseTreeFromDiagram(s string) (*ParseTree, error) {
	var err error
	var pt *ParseTree
	st := &box.Stack[*ParseTree]{}

	var curLine int
	var inEscape bool
	var text strings.Builder

	for _, ch := range s {
		// handle escape sequences
		if inEscape {
			if ch != '(' && ch != ')' && ch != '[' && ch != ']' && ch != '\\' && !unicode.IsSpace(ch) {
				err = fmt.Errorf("invalid escape sequence at line %d", curLine)
				return nil, err
			}
			text.WriteRune(ch)
			inEscape = false
			continue
		}

		// inc line number if we hit a newline, before discarding it
		if ch == '\n' {
			curLine++
		}

		// ignore whitespace
		if unicode.IsSpace(ch) {
			continue
		}

		switch ch {
		case '\\':
			inEscape = true
		case '(':
			if st.Len() == 0 {
				// just put it on the stack itself
				st.Push(&ParseTree{Terminal: true})
			} else {
				// make it a child of the top of the stack and push it on.
				parent := st.Pop()

				if parent.Terminal {
					err = fmt.Errorf("unexpected start of term '(' at line %d; cannot have a terminal in a terminal", curLine)
					return nil, err
				}

				// give parent the text we've been building
				if parent.Value == "" {
					parent.Value = text.String()
				}

				child := &ParseTree{Terminal: true}
				parent.Children = append(parent.Children, child)
				st.Push(parent)
				st.Push(child)
			}

			text.Reset()
		case ')':
			if st.Len() == 0 {
				err = fmt.Errorf("unexpected end of term ')' at line %d; not currently in term", curLine)
				return nil, err
			}

			term := st.Pop()
			if !term.Terminal {
				err = fmt.Errorf("unexpected end of term ')' at line %d; not currently in term, did you mean ']'?", curLine)
				return nil, err
			}

			term.Value = text.String()

			if st.Len() == 0 {
				pt = term
			}

			text.Reset()
		case '[':
			if st.Len() == 0 {
				// just put it on the stack itself
				st.Push(&ParseTree{Terminal: false})
			} else {
				// make it a child of the top of the stack and push it on.
				parent := st.Pop()

				if parent.Terminal {
					err = fmt.Errorf("unexpected start of non-term '[' at line %d; cannot have a non-terminal in a terminal", curLine)
					return nil, err
				}

				// give parent the text we've been building
				if parent.Value == "" {
					parent.Value = text.String()
				}

				child := &ParseTree{Terminal: false}
				parent.Children = append(parent.Children, child)
				st.Push(parent)
				st.Push(child)
			}

			text.Reset()
		case ']':
			if st.Len() == 0 {
				err = fmt.Errorf("unexpected end of non-term ']' at line %d; not currently in non-term", curLine)
				return nil, err
			}

			nonTerm := st.Pop()
			if nonTerm.Terminal {
				err = fmt.Errorf("unexpected end of non-term ']' at line %d; not currently in non-term, did you mean ')'?", curLine)
				return nil, err
			}

			if nonTerm.Value == "" {
				nonTerm.Value = text.String()
			}

			if st.Len() == 0 {
				pt = nonTerm
			}

			text.Reset()
		default:
			text.WriteRune(ch)
		}
	}

	if st.Len() > 0 {
		nodeOpenStr := "["
		last := st.Pop()
		if last.Terminal {
			nodeOpenStr = "("
		}

		name := last.Value
		if name == "" {
			name = text.String()
		}

		err = fmt.Errorf("parse tree diagram ends with unclosed node: \"%s%s\"", nodeOpenStr, name)
		return nil, err
	}

	return pt, nil
}

// PTLeaf is a convenience function for creating a new ParseTree that
// represents a terminal symbol. The Source token may or may not be set as
// desired. Note that t's type being ...Token is simply to make it optional;
// only the first such provided t is examined.
func PTLeaf(term string, t ...Token) *ParseTree {
	pt := &ParseTree{Terminal: true, Value: term}
	if len(t) > 0 {
		pt.Source = t[0]
	}
	return pt
}

// PTNode is a convenience function for creating a new ParseTree that
// represents a non-terminal symbol with minimal text.
func PTNode(nt string, children ...*ParseTree) *ParseTree {
	pt := &ParseTree{
		Terminal: false,
		Value:    nt,
		Children: children,
	}
	return pt
}

// Follow takes a path, denoted as a slice of indexes of children to follow,
// starting from the ParseTree it is called on, and returns the descendant tree
// it leads to.
func (pt ParseTree) Follow(path []int) *ParseTree {
	cur := &pt
	for i := range path {
		if path[i] < 0 || path[i] >= len(cur.Children) {
			panic(fmt.Sprintf("cannot follow path[%d]: index out of range: %d", i, path[i]))
		}

		cur = cur.Children[path[i]]
	}

	return cur
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

// PathToDiff returns the point at which the two parse trees diverge, as
// well as whether they diverge at all. If they do not diverge, the returned
// path should not be used.
//
// Finds the path to the point at which the two trees diverge. If set to ignore
// short-circuit, does not include any branches that were inserted via
// shortest-circuit, as detected with "__ICTIO__:SHORTCIRC:" in front of it.
//
// If there are multiple nodes that satisfy the above definition, then the point
// of divergence is the first common ancestor of all such nodes.
//
// This does not consider the Source field, ergo only the structures of the
// trees are compared, not their contents.
//
// Runs in O(n) time with respect to the number of nodes in the trees.
func (pt ParseTree) PathToDiff(t ParseTree, ignoreShortCircuit bool) (path []int, diverges bool) {
	checkStack := &box.Stack[box.HPair[treeNode]]{}
	checkStack.Push(box.HPairOf(treeNode{&t, slices.LList[int]{}}, treeNode{&pt, slices.LList[int]{}}))

	allPoints := [][]int{}

	for !checkStack.Empty() {
		p := checkStack.Pop()
		tn1, tn2 := p.All()
		n1, n2 := tn1.node, tn2.node
		if !ignoreShortCircuit || (!(strings.HasPrefix(n1.Value, ShortCircuitPrefix) && !strings.HasPrefix(n2.Value, ShortCircuitPrefix))) {
			if n1.Terminal != n2.Terminal || n1.Value != n2.Value || len(n1.Children) != len(n2.Children) {
				// diverges here
				allPoints = append(allPoints, tn1.path.Slice())
				// don't check the rest of the children
				continue
			}
		}

		pList := tn1.path // NOTE: may be nil
		for i := len(tn1.node.Children) - 1; i >= 0; i-- {
			nextPath := pList.Add(i)

			tn1Child := tn1.node.Children[i]
			tn2Child := tn2.node.Children[i]

			tn1Item := treeNode{tn1Child, nextPath}
			tn2Item := treeNode{tn2Child, nextPath}

			checkStack.Push(box.HPairOf(tn1Item, tn2Item))
		}
	}

	if len(allPoints) == 0 {
		return nil, false
	}

	if len(allPoints) > 1 {
		// find the first common ancestor

		// check each index
		for i := 0; i < len(allPoints[0]); i++ {

			// check each output
			for j := 1; j < len(allPoints); j++ {
				if allPoints[j-1][i] != allPoints[j][i] {
					// first common ancestor is the parent of the first divergent node
					return allPoints[0][:i], true
				}
			}
		}
	}

	return allPoints[0], true
}

// IsSubTreeOf checks if this ParseTree is a sub-tree of the given parse tree t.
// Does not consider Source for its comparisons, ergo only the structure is
// examined.
//
// This performs a depth-first traversal of t, checking if there is any sub-tree
// in t s.t. pt is exactly equal to that node. Runs in O(n^2) time with respect
// to the number of nodes in the trees.
//
// Returns whether pt is a sub-tree of t, and if so, the path to the first
// node in t where this is the case. The path is represented as a slice of ints
// where each is the child index of the node to traverse to. If it is empty,
// then the root node is the first node where sub is a sub-tree; this is not
// necessarily the same as equality.
func (pt ParseTree) IsSubTreeOf(t ParseTree) (contains bool, path []int) {
	checkStack := &box.Stack[treeNode]{}
	checkStack.Push(treeNode{&t, slices.LList[int]{}})

	for !checkStack.Empty() {
		p := checkStack.Pop()
		startNode := p.node
		pList := p.path

		if pt.Equal(startNode) {
			return true, pList.Slice()
		}

		for i := len(startNode.Children) - 1; i >= 0; i-- {
			nextPath := pList.Add(i)
			checkStack.Push(treeNode{startNode.Children[i], nextPath})
		}
	}

	return false, nil
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

type treeNode struct {
	node *ParseTree
	path slices.LList[int]
}
