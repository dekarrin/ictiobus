package box

// Sequence is a container whose elements have an ordering. Calling Elements()
// on a Sequence is garaunteed to return the elements in the same order they
type Sequence[E any] interface {
	Container[E]
}
