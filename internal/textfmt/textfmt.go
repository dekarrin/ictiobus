package textfmt

import (
	"math"
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

// Pluralize returns the singular string if count is 1, otherwise the plural
// string. If the plural string starts with '-', it is treated as a suffix and
// the pluralization is done by removing the leading '-' and appending it to the
// singular string.
func Pluralize(count int, sing, plural string) string {
	if count == 1 {
		return sing
	}

	if strings.HasPrefix(plural, "-") {
		return sing + plural[1:]
	}

	return plural
}

// OrdinalSuf returns the suffix for the ordinal version of the given number,
// e.g. "rd" for 3, "st" for 51, etc.
func OrdinalSuf(a int) string {
	// first, if negative, just give the prefix for the positive
	if a < 0 {
		a *= -1
	}

	// special exception for the english language: 11th, 12th, and 13th break
	// the "by 1's place" rule, and need to be allowed for explicitly
	if a == 11 || a == 12 || a == 13 {
		return "th"
	}

	finalDigit := a

	if a > 9 {
		// it all depends on the final digit
		nextTen := int(math.Floor(float64(a)/10) * 10)
		finalDigit = a - nextTen
	}

	if finalDigit == 1 {
		return "st"
	} else if finalDigit == 2 {
		return "nd"
	} else if finalDigit == 3 {
		return "rd"
	} else {
		return "th"
	}
}
