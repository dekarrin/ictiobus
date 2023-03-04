package textfmt

import "sort"

// OrderedKeys returns the keys of m, ordered a particular way. The order is
// guaranteed to be the same on every run.
//
// As of this writing, the order is alphabetical, but this function does not
// guarantee this will always be the case.
func OrderedKeys[V any](m map[string]V) []string {
	var keys []string
	var idx int

	keys = make([]string, len(m))
	idx = 0

	for k := range m {
		keys[idx] = k
		idx++
	}

	sort.Strings(keys)

	return keys
}
