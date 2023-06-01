// Package box contains generic types and interfaces for various container data
// types. Stacks, sets, matrixes, and other such types which primarily hold
// values can be found here.
//
// This package does not contain manipulation functions to operate directly on
// slices; use the github.com/dekarrin/ictiobus/internal/slices library for
// that.
package box

import "fmt"

type heapNode[E any] struct {
	parent      *heapNode[E]
	left        *heapNode[E]
	right       *heapNode[E]
	v           E
	buildingIdx int // index within parent heap's buildingLevel slice. no meaning if node is not in it.
}

// Heap is a data structure that is optimized for retrieval of the element with
// the highest value. Both insertion and deletion is O(log n). CompareFunc is
// used to compare two values for sorting; it returns whether the left argument
// is "less than" (should be sorted as before) the right argument. CompareFunc
// must be set to a function or it will by default perform string conversion of
// values with %v in Printf and compare the results.
//
// The zero-value for a Heap is ready for immediate use, but it is recommended
// that a pointer be obtained if it is going to be passed to other functions
// outside the one it is used in.
type Heap[E any] struct {
	// CompareFunc is used to determine ordering of elements.
	CompareFunc func(l, r E) bool

	// root is the root
	root *heapNode[E]

	// openBotNode is the leftmost node with at least one open space in the
	// first level that is not completely filled, starting from the top.
	openBuildingNode *heapNode[E]

	// buildingLevel is all nodes in the level being "built"; openBotNode is a
	// pointer to the first element in this slice with at least one space in it.
	buildingLevel []*heapNode[E]

	// lastElem is the rightmost leaf node of the lowest level.
	lastElem *heapNode[E]

	// tracks number of elements.
	count int
}

// Elements returns all elements in the heap. They will be returned in the order
// that they are visited in a leftmost depth-first traversal of the heap.
func (h *Heap[E]) Elements() []E {
	if h == nil || h.root == nil {
		return nil
	}

	var elems []E

	nodeStack := NewStack([]*heapNode[E]{h.root})

	for nodeStack.Len() > 0 {
		n := nodeStack.Pop()

		elems = append(elems, n.v)

		if n.right != nil {
			nodeStack.Push(n.right)
		}
		if n.left != nil {
			nodeStack.Push(n.left)
		}
	}

	return elems
}

// Len returns the number of elements in the heap. Element count is tracked so
// this operation is O(1).
func (h *Heap[E]) Len() int {
	if h == nil {
		return 0
	}
	return h.count
}

// Peek returns the element at the top of the heap. This is the "leftmost" of
// values according to CompareFunc. This is an O(1) operation.
func (h *Heap[E]) Peek() E {
	if h == nil || h.root == nil {
		panic("peek empty heap")
	}

	return h.root.v
}

// Pop removes the element at the top of the heap and returns it. This is the
// "leftmost" of values according to CompareFunc.
func (h *Heap[E]) Pop() E {
	if h == nil || h.root == nil {
		panic("pop empty heap")
	}

	oldRootValue := h.root.v

	h.count--

	// special case - count of 1
	if h.root.left == nil && h.root.right == nil {
		h.root = nil
		h.buildingLevel = nil
		h.lastElem = nil
		h.openBuildingNode = nil
	}

	// replace the root element with the rightmost bottom level element.
	lastElem := h.lastElem
	h.root.v = lastElem.v

	// as this is a replacement, we need to remove the last element's prior node
	// and update relevant pointers
	if lastElem.parent.right == lastElem {
		lastElem.parent.right = nil

		// if lastElem was the end of the current building level, then we just
		// destroyed the building level and need the *parent* level to become
		// the new one.
		if lastElem == h.buildingLevel[len(h.buildingLevel)-1] {
			var newLevel []*heapNode[E]
			for _, bn := range h.buildingLevel {
				bn.buildingIdx = len(newLevel)
				newLevel = append(newLevel, bn)
			}
			h.buildingLevel = newLevel

			// the last in the building level automatically becomes the new open
			// one.
			h.openBuildingNode = h.buildingLevel[len(h.buildingLevel)-1]
		} else {
			// otherwise, need to update the openBuilding node because we just
			// opened a space in our parent, so they are the new one.
			h.openBuildingNode = lastElem.parent
		}

		// finally, make sibling to the left be the new lastElem
		h.lastElem = lastElem.parent.left
	} else {
		// lastElem is on the *left*, which means it can never be the end of a
		// buildingLevel. Glubbin' nice! 38D
		lastElem.parent.left = nil

		// eliminating the left means both openBuildingNode and buildingLevel
		// are both still valid, no need to update. but we do need to update
		// h.lastElem. It becomes either the right child of the buildingNode
		// *before* lastElem's parent, or if lastElem's parent is the first node
		// in the level, the last node in the building level.
		if lastElem.parent.buildingIdx == 0 {
			h.lastElem = h.buildingLevel[len(h.buildingLevel)-1]
		} else {
			lastElemParentSib := h.buildingLevel[lastElem.parent.buildingIdx-1]
			h.lastElem = lastElemParentSib.right
		}
	}

	// node has been removed from bottom and all metadata has been properly
	// updated. Restore the heap property.
	//
	// curNode must be lt both of its items. if not, it is swapped with the more
	// lt of the two.

	curNode := h.root
	for {
		// edge case: only right is filled.
		if curNode.left == nil && curNode.right == nil {
			// nothing to do, this is a leaf
			break
		}

		// edge case: only left is filled
		if curNode.right == nil {
			// check against left side only

			if h.comesBefore(curNode.v, curNode.left.v) {
				break
			}

			// otherwise, we need to swap
			curNode.v, curNode.left.v = curNode.left.v, curNode.v
			curNode = curNode.left
			continue
		}

		// non-edge case: both left and right are filled.

		if h.comesBefore(curNode.left.v, curNode.right.v) {
			// swap with the left side
			curNode.v, curNode.left.v = curNode.left.v, curNode.v
			curNode = curNode.left
		} else {
			// swap with right side
			curNode.v, curNode.right.v = curNode.right.v, curNode.v
			curNode = curNode.right
		}
	}

	return oldRootValue

}

// Add adds a new element to the heap. This is an O(log n) operation.
func (h *Heap[E]) Add(elem E) {
	if h == nil {
		panic("add element to nil heap")
	}

	node := &heapNode[E]{v: elem}
	h.lastElem = node

	h.count++

	// special initial case - no elements yet defined
	if h.root == nil {
		h.root = node
		h.lastElem = node
		h.openBuildingNode = node
		node.buildingIdx = 0
		h.buildingLevel = []*heapNode[E]{node}
		return
	}

	// add node to bottom, initially preserving binary shape
	buildingNode := h.openBuildingNode
	if buildingNode.left == nil {
		buildingNode.left = node
		node.parent = buildingNode

		// presumably the right is still open and openBuildingNode is still valid.
	} else if buildingNode.right == nil {
		buildingNode.right = node
		node.parent = buildingNode

		// this node is now filled and openBuildingNode must be updated.

		// is there a sibling to the right?
		if buildingNode.buildingIdx+1 < len(h.buildingLevel) {
			// yes, that sibling is the new open

			sibling := h.buildingLevel[buildingNode.buildingIdx+1]
			h.openBuildingNode = sibling
		} else {
			// need to start new building level consisting of all child nodes
			// of current level

			var nextLevel []*heapNode[E]
			for _, bn := range h.buildingLevel {
				bn.left.buildingIdx = len(nextLevel)
				bn.right.buildingIdx = len(nextLevel) + 1
				nextLevel = append(nextLevel, bn.left, bn.right)
			}
			h.buildingLevel = nextLevel
			h.openBuildingNode = nextLevel[0]
		}
	}

	// node has now been added to bottom and all metadata has been properly
	// updated. Restore the heap property.
	curNode := node
	for curNode.parent != nil && h.comesBefore(curNode.v, curNode.parent.v) {
		// swap the two values; by swaping the values only instead of entire
		// nodes, we preserve existing pointers to nodes that are based on their
		// relative structure.
		curNode.v, curNode.parent.v = curNode.parent.v, curNode.v
		curNode = curNode.parent
	}

	// done.
}

func (h *Heap[E]) comesBefore(l, r E) bool {
	if h.CompareFunc != nil {
		return h.CompareFunc(l, r)
	}

	leftVal := fmt.Sprintf("%v", l)
	rightVal := fmt.Sprintf("%v", r)

	return leftVal <= rightVal
}
