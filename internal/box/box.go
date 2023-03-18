package box

import (
	"fmt"
	"sort"
)

type Container[E any] interface {
	// Elements returns a slice of the elements. They are not guaranteed to be
	// in any particular order.
	Elements() []E
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
