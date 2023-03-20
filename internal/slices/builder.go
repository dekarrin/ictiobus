package slices

// LList allows building a slice by appending elements one at a time,
// with multiple refering to a common prefix to avoid copying. Each Add will
// return a new LinkedBuilder with the new element appended.
//
// Call Slice() to create slice built out of the linked list.
//
// LList is immutable; each Add returns a new LList.
//
// The zero value is a valid empty list.
type LList[E any] struct {
	d      E
	prev   *LList[E]
	filled bool
}

func (ll LList[E]) Add(d E) LList[E] {
	newList := LList[E]{d: d, filled: true}

	if ll.filled {
		newList.prev = &ll
	}

	return newList
}

func (ll LList[E]) Remove() LList[E] {
	if !ll.filled {
		panic("cannot remove from empty list")
	}

	if ll.prev == nil {
		// this is the first node. Return an empty list.
		return LList[E]{}
	}

	return *ll.prev
}

// Len gets the number of elements in the list. Note that this is an O(n)
// operation due to LL's not tracking their length.
func (ll LList[E]) Len() int {
	if !ll.filled {
		return 0
	}

	count := 0
	cur := &ll
	for cur != nil {
		count++
		cur = cur.prev
	}

	return count
}

func (ll LList[E]) Empty() bool {
	return !ll.filled
}

func (ll LList[E]) Slice() []E {
	if ll.Empty() {
		return []E{}
	}

	// find number of elements
	curr := &ll
	count := 0
	for curr != nil {
		count++
		curr = curr.prev
	}

	sl := make([]E, count)
	slIdx := count - 1

	cur := &ll
	for cur != nil {
		sl[slIdx] = cur.d
		slIdx--
		cur = cur.prev
	}

	return sl
}
