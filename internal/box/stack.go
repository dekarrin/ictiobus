package box

import (
	"fmt"
	"reflect"
	"strings"
)

// Stack is a FIFO stack of items. It is backed by a slice where the left-most
// position is the top of the stack. The zero-value for the stack is an empty
// stack, ready to use.
type Stack[E any] struct {
	of []E
}

// NewStack creates a stack whose contents are set to the given slice, with the
// element at index 0 acting as the top of the stack and end of the slice as the
// bottom. The new stack will take ownership of elems and the caller should not
// use it anymore.
func NewStack[E any](elems []E) *Stack[E] {
	return &Stack[E]{
		of: elems,
	}
}

// Push pushes value onto the stack.
func (s *Stack[E]) Push(v E) {
	if s == nil {
		panic("push to nil stack")
	}
	s.of = append([]E{v}, s.of...)
}

// Pop pops value off of the stack. If the stack is empty, panics.
func (s *Stack[E]) Pop() E {
	if s == nil {
		panic("pop of nil stack")
	}
	if len(s.of) == 0 {
		panic("pop of empty stack")
	}

	v := s.of[0]
	newVals := make([]E, len(s.of)-1)
	copy(newVals, s.of[1:])
	s.of = newVals
	return v
}

// Peek checks the value at the top of the stack. If the stack is empty, panics.
func (s *Stack[E]) Peek() E {
	if s == nil || len(s.of) == 0 {
		panic("peek of empty stack")
	}
	return s.of[0]
}

// PeekAt checks the value at the given position of the stack. If the stack is
// not that long, panics.
func (s *Stack[E]) PeekAt(p int) E {
	if s == nil || p >= len(s.of) {
		panic(fmt.Sprintf("stack index out of range: %d", p))
	}
	return s.of[p]
}

// Len returns the length of the stack.
func (s *Stack[E]) Len() int {
	if s == nil {
		return 0
	}

	return len(s.of)
}

// Empty returns whether the stack is empty.
func (s *Stack[E]) Empty() bool {
	if s == nil {
		return true
	}

	return len(s.of) == 0
}

// Elements returns the items in the stack as a slice, with the top of the stack
// at the 0th index. Modifying the returned slice will have no effect on the
// Stack.
func (s *Stack[E]) Elements() []E {
	if s == nil {
		return nil
	}

	elems := make([]E, len(s.of))
	copy(elems, s.of)
	return elems
}

// String shows the contents of the stack as a simple slice.
func (s Stack[E]) String() string {
	var sb strings.Builder

	sb.WriteString("Stack[")
	for i := range s.of {
		sb.WriteString(fmt.Sprintf("%v", s.of[i]))
		if i+1 < len(s.of) {
			sb.WriteRune(',')
			sb.WriteRune(' ')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

// Equal returns whether two stacks have exactly the same contents in the same
// order. It supports comparing to other stacks, slices, and pointers to either.
func (s Stack[E]) Equal(o any) bool {
	other, ok := o.(Stack[E])
	if !ok {
		// also okay if its the pointer value, as long as its non-nil
		otherPtr, ok := o.(*Stack[E])
		if !ok {
			// also okay if it's a slice
			otherSlice, ok := o.([]E)

			if !ok {
				// also okay if it's a ptr to slice
				otherSlicePtr, ok := o.(*[]E)
				if !ok {
					return false
				} else if otherSlicePtr == nil {
					return false
				} else {
					other = Stack[E]{of: *otherSlicePtr}
				}
			} else {
				other = Stack[E]{of: otherSlice}
			}
		} else if otherPtr == nil {
			return false
		} else {
			other = *otherPtr
		}
	}

	return reflect.DeepEqual(s.of, other.of)
}
