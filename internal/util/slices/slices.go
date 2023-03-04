// Package slices contains functions useful for operating on slices.
package slices

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
