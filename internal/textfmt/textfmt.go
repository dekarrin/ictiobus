package textfmt

import (
	"sort"
	"strings"
	"unicode"
)

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

// ArticleFor returns the article for the given string. It will be capitalized
// the same as the string. If definite is true, the returned value will be "the"
// capitalized as described; otherwise, it will be "a"/"an" capitalized as
// described.
func ArticleFor(s string, definite bool) string {
	sRunes := []rune(s)

	if len(sRunes) < 1 {
		return ""
	}

	leadingUpper := unicode.IsUpper(sRunes[0])
	allCaps := leadingUpper
	if leadingUpper && len(sRunes) > 1 {
		allCaps = unicode.IsUpper(sRunes[1])
	}

	art := ""
	if definite {
		if allCaps {
			art = "THE"
		} else if leadingUpper {
			art = "The"
		} else {
			art = "the"
		}
	} else {
		if allCaps || leadingUpper {
			art = "A"
		} else {
			art = "a"
		}

		sUpperRunes := []rune(strings.ToUpper(s))
		first := sUpperRunes[0]
		if first == 'A' || first == 'E' || first == 'I' || first == 'O' || first == 'U' {
			if allCaps {
				art += "N"
			} else {
				art += "n"
			}
		}
	}

	return art
}

// TruncateWith truncates s to maxLen. If s is longer than maxLen, it is
// truncated and the cont string is placed after it. If it is shorter or equal
// to maxLen, it is returned unchanged.
func TruncateWith(s string, maxLen int, cont string) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + cont
}
