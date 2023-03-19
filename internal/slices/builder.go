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
	d     E
	prev  *LList[E]
	count int
}

func (ll LList[E]) Add(d E) LList[E] {
	newList := LList[E]{d: d, count: ll.count + 1}

	if ll.count != 0 {
		newList.prev = &ll
	}

	return newList
}

func (ll LList[E]) Remove() LList[E] {
	if ll.count == 0 {
		panic("cannot remove from empty list")
	}

	if ll.prev == nil {
		// this is the first node. Return an empty list.
		return LList[E]{}
	}

	return *ll.prev
}

func (ll LList[E]) Empty() bool {
	return ll.count == 0
}

func (ll LList[E]) Slice() []E {
	if ll.Empty() {
		return []E{}
	}

	sl := make([]E, ll.count)
	slIdx := ll.count - 1

	cur := &ll
	for cur != nil {
		sl[slIdx] = cur.d
		slIdx--
		cur = cur.prev
	}

	return sl
}
