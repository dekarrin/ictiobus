// Package tmatch provides some quick and dirty matching functions with detailed
// mismatch error messages for use with unit tests. It will probably eventually
// be replaced with a well-vetted library like gomatch or gomega in the future.
package tmatch

import (
	"fmt"

	"github.com/dekarrin/ictiobus/internal/textfmt"
)

// CompFunc is a comparison function that returns true if the two values are
// equal. Used for defining custom comparison for types that do not fulfill
// the 'comparable' type constraint.
type CompFunc func(v1, v2 any) bool

// Comparer convertes the given function into a CompFunc. The returned CompFunc
// ret to ensure the types are exactly as expected, and if not, considers
// it a mismatch.
func Comparer[E1 any, E2 any](fn func(v1 E1, v2 E2) bool) CompFunc {
	return func(v1, v2 any) bool {
		c1, ok := v1.(E1)
		if !ok {
			return false
		}
		c2, ok := v2.(E2)
		if !ok {
			return false
		}

		return fn(c1, c2)
	}
}

// AnyStrMapV returns an error if the actual map does not match any of the maps
// in expect, using vMatches to match elements.
func AnyStrMapV[E any](actual map[string]E, expect []map[string]E, vMatches CompFunc) error {
	foundAny := false
	for _, expectMap := range expect {
		candidateFailedMatch := false
		// check that no key is present in one that is not present in the other
		for key := range actual {
			if _, ok := expectMap[key]; !ok {
				candidateFailedMatch = true
				break
			}
		}
		if candidateFailedMatch {
			continue
		}
		for key := range expectMap {
			if _, ok := actual[key]; !ok {
				candidateFailedMatch = true
				break
			}
		}
		if candidateFailedMatch {
			continue
		}

		// keys are all the same, now check values
		for key := range actual {
			if !vMatches(actual[key], expectMap[key]) {
				candidateFailedMatch = true
				break
			}
		}
		if !candidateFailedMatch {
			foundAny = true
			break
		}
	}

	if !foundAny {
		errMsg := "actual does not match any expected:\n     "
		if len(expect) > 9 {
			errMsg += " "
		}
		errMsg += "actual: "
		if actual == nil {
			errMsg += "nil"
		} else {
			errMsg += "{"
			ordered := textfmt.OrderedKeys(actual)
			for i, k := range ordered {
				errMsg += fmt.Sprintf("%s: %v", k, actual[k])
				if i+1 < len(ordered) {
					errMsg += ", "
				}
			}
			errMsg += "}"
		}
		errMsg += "\n"
		for i, expectMap := range expect {
			errMsg += fmt.Sprintf("expected[%d]: ", i)
			if expectMap == nil {
				errMsg += "nil"
			} else {
				errMsg += "{"
				ordered := textfmt.OrderedKeys(expectMap)
				for j, k := range ordered {
					errMsg += fmt.Sprintf("%s: %v", k, expectMap[k])
					if j+1 < len(ordered) {
						errMsg += ", "
					}
				}
				errMsg += "}"
			}
			errMsg += "\n"
		}
		return fmt.Errorf(errMsg)
	}
	return nil
}

// MatchAnyStrMap checks if the actual string map matches any of the expected
// maps. All maps must have a comparable value type. If the value type is not
// comparable, use MatchAnyStrMapV instead to provide a custom equality check.
func AnyStrMap[E comparable](actual map[string]E, expect []map[string]E) error {

	return AnyStrMapV(actual, expect, Comparer(func(v1, v2 E) bool {
		return v1 == v2
	}))
}
