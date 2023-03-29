// Package slices contains functions useful for operating on slices.
package slices

import "sort"

// In returns whether s is present in the given slice by checking each item
// in order.
func In[V comparable](s V, slice []V) bool {
	for i := range slice {
		if slice[i] == s {
			return true
		}
	}
	return false
}

// CustomComparable is an interface for items that may be checked against
// arbitrary other objects. In practice most will attempt to typecast to their
// own type and immediately return false if the argument is not the same, but in
// theory this allows for comparison to multiple types of things.
type CustomComparable interface {
	Equal(other any) bool
}

// EqualSlices checks that the two slices contain the same items in the same
// order. Equality of items is checked by items in the slices are equal by
// calling the custom Equal function on each element. In particular, Equal is
// called on elements of sl1 with elements of sl2 passed in as the argument.
func EqualSlices[T CustomComparable](sl1 []T, sl2 []T) bool {
	if len(sl1) != len(sl2) {
		return false
	}

	for i := range sl1 {
		if !sl1[i].Equal(sl2[i]) {
			return false
		}
	}

	return true
}

// LongestCommonPrefix gets the longest prefix that the two slices have in
// common.
func LongestCommonPrefix[T comparable](sl1 []T, sl2 []T) []T {
	var pref []T

	minLen := len(sl1)
	if minLen > len(sl2) {
		minLen = len(sl2)
	}

	for i := 0; i < minLen; i++ {
		if sl1[i] != sl2[i] {
			break
		}
		pref = append(pref, sl1[i])
	}

	return pref
}

// HasPrefix returns whether the given slice has the given prefix. If prefix
// is empty or nil this will always be true regardless of sl's value.
func HasPrefix[T comparable](sl []T, prefix []T) bool {
	if len(prefix) > len(sl) {
		return false
	}

	if len(prefix) == 0 {
		return true
	}

	for i := range prefix {
		if sl[i] != prefix[i] {
			return false
		}
	}

	return true
}

type sorter[E any] struct {
	src []E
	lt  func(left, right E) bool
}

func (s sorter[E]) Len() int {
	return len(s.src)
}

func (s sorter[E]) Swap(i, j int) {
	s.src[i], s.src[j] = s.src[j], s.src[i]
}

func (s sorter[E]) Less(i, j int) bool {
	return s.lt(s.src[i], s.src[j])
}

// SortBy takes the items and uses the provided function to sort the list. The
// function should return true if left is less than (comes before) right.
//
// items will not be modified.
func SortBy[E any](items []E, lt func(left E, right E) bool) []E {
	if len(items) == 0 || lt == nil {
		return items
	}

	s := sorter[E]{
		src: make([]E, len(items)),
		lt:  lt,
	}

	copy(s.src, items)
	sort.Sort(s)
	return s.src
}

// Any returns the first item in the slice that satisfies the given predicate.
// If there is no item, the zero value is returned and ok is false.
func Any[E any](sl []E, fn func(item E) bool) (item E, ok bool) {
	for _, it := range sl {
		if fn(it) {
			return it, true
		}
	}
	return item, false
}

// Filter returns a slice with only those items from sl that filterFn returns
// true for. If filterFn is nil, returns a nil slice.
func Filter[E any](sl []E, filterFn func(item E) bool) []E {
	if filterFn == nil || sl == nil {
		return nil
	}

	out := []E{}
	for _, item := range sl {
		if filterFn(item) {
			out = append(out, item)
		}
	}

	return out
}

func Reduce[E any, V any](sl []E, initial V, fn func(idx int, item E, accum V) V) V {
	accum := initial
	for i, item := range sl {
		accum = fn(i, item, accum)
	}
	return accum
}
