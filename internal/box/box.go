// Package box contains generic types and interfaces for various container data
// types. Stacks, sets, matrixes, and other such types which primarily hold
// values can be found here.
//
// This package does not contain manipulation functions to operate directly on
// slices; use the github.com/dekarrin/ictiobus/internal/slices library for
// that.
package box

import (
	"fmt"
	"sort"
)

// Container is the interface that all types in box satisfy. It has an accessor
// for obtaining the elements within it (not guaranteed to be in any order by
// the interface, but implementations may have specific ordering), and for
// getting the number of elements in it.
type Container[E any] interface {
	// Elements returns a slice of the elements. They are not guaranteed to be
	// in any particular order.
	Elements() []E

	// Len returns the number of elements in the Container. This will always
	// match len(Elements()), but can sometimes be cheaper to call.
	Len() int
}

type namedSortable[V any] struct {
	val  V
	name string
}

// Alphabetized returns the ordered elements of the given container. The string
// representation '%v' of each element is used for comparison.
func Alphabetized[V any](c Container[V]) []V {
	// convert them all to string and order that.

	toSort := []namedSortable[V]{}

	for _, item := range c.Elements() {
		itemStr := fmt.Sprintf("%v", item)
		toSort = append(toSort, namedSortable[V]{val: item, name: itemStr})
	}

	sortFunc := func(i, j int) bool {
		return toSort[i].name < toSort[j].name
	}

	sort.Slice(toSort, sortFunc)

	sortedVals := make([]V, len(toSort))
	for i := range toSort {
		sortedVals[i] = toSort[i].val
	}

	return sortedVals
}
