// Package stack provies a stack structure based off of a Slice.
package stack

import (
	"fmt"
	"reflect"
	"strings"
)

// Stack is a stack. It is backed by a slice where the left-most position is the
// top of the stack. The zero-value for the stack is ready to use.
type Stack[E any] struct {
	Of []E
}

// Push pushes value onto the stack.
func (s *Stack[E]) Push(v E) {
	if s == nil {
		panic("push to nil stack")
	}
	s.Of = append([]E{v}, s.Of...)
}

// Pop pops value off of the stack. If the stack is empty, panics.
func (s *Stack[E]) Pop() E {
	if s == nil {
		panic("pop of nil stack")
	}
	if len(s.Of) == 0 {
		panic("pop of empty stack")
	}

	v := s.Of[0]
	newVals := make([]E, len(s.Of)-1)
	copy(newVals, s.Of[1:])
	s.Of = newVals
	return v
}

// Peek checks the value at the top of the stack. If the stack is empty, panics.
func (s Stack[E]) Peek() E {
	if len(s.Of) == 0 {
		panic("peek of empty stack")
	}
	return s.Of[0]
}

// PeekAt checks the value at the given position of the stack. If the stack is
// not that long, panics.
func (s Stack[E]) PeekAt(p int) E {
	if p >= len(s.Of) {
		panic(fmt.Sprintf("stack index out of range: %d", p))
	}
	return s.Of[p]
}

// Len returns the length of the stack.
func (s Stack[E]) Len() int {
	return len(s.Of)
}

// Empty returns whether the stack is empty.
func (s Stack[E]) Empty() bool {
	return len(s.Of) == 0
}

// String shows the contents of the stack as a simple slice.
func (s Stack[E]) String() string {
	var sb strings.Builder

	sb.WriteString("Stack[")
	for i := range s.Of {
		sb.WriteString(fmt.Sprintf("%v", s.Of[i]))
		if i+1 < len(s.Of) {
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
					other = Stack[E]{Of: *otherSlicePtr}
				}
			} else {
				other = Stack[E]{Of: otherSlice}
			}
		} else if otherPtr == nil {
			return false
		} else {
			other = *otherPtr
		}
	}

	return reflect.DeepEqual(s.Of, other.Of)
}
