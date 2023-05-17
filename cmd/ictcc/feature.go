package main

import (
	"fmt"
	"strconv"
	"strings"
)

// ExpFeature is an experimental feature flag for enabling features not normally
// enabled.
type ExpFeature int

const (
	FeatureNone ExpFeature = iota
	FeatureInheritedAttributes
)

// ExpFeatureAll() returns a slice of all the ExpFeature constants.
func ExpFeatureAll() []ExpFeature {
	// When a new ExpFeature is added, it must be added to this slice manually.
	fs := []ExpFeature{
		FeatureNone,
		FeatureInheritedAttributes,
	}

	return fs
}

// Short returns the short-code of an ExpFeature.
func (f ExpFeature) Short() string {
	switch f {
	case FeatureNone:
		return "none"
	case FeatureInheritedAttributes:
		return "inherited-attributes"
	default:
		return fmt.Sprintf("%d", int(f))
	}
}

// String returns the string representation of an ExpFeature.
func (f ExpFeature) String() string {
	switch f {
	case FeatureNone:
		return "FeatureNone"
	case FeatureInheritedAttributes:
		return "FeatureInheritedAttributes"
	default:
		return fmt.Sprintf("ExpFeature(%d)", int(f))
	}
}

// ParseShortExpFeature parses an ExpFeature from the given short string. The
// short string may be a string consisting of the same string as the Short()
// method returns, or be an integer represented as a string that corresponds to
// the desired feature.
func ParseShortExpFeature(s string) (ExpFeature, error) {
	sLower := strings.ToLower(s)

	// if it is one of the short strings, use that
	for _, f := range ExpFeatureAll() {
		if sLower == f.Short() {
			return f, nil
		}
	}

	// else, try to parse an int from it
	intVal, err := strconv.Atoi(s)
	if err != nil {
		return FeatureNone, fmt.Errorf("not a feature short name and not an int: %q", s)
	}

	for _, f := range ExpFeatureAll() {
		if intVal == int(f) {
			return f, nil
		}
	}

	return FeatureNone, fmt.Errorf("not a valid feature type code: %d", intVal)
}
