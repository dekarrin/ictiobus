package main

import (
	"fmt"
	"strconv"
	"strings"
)

// expFeature is an experimental feature flag for enabling features not normally
// enabled.
type expFeature int

const (
	featureNone expFeature = iota
	featureInheritedAttributes
)

// expFeatureAll() returns a slice of all the ExpFeature constants.
func expFeatureAll() []expFeature {
	// When a new ExpFeature is added, it must be added to this slice manually.
	fs := []expFeature{
		featureNone,
		featureInheritedAttributes,
	}

	return fs
}

// Short returns the short-code of an ExpFeature.
func (f expFeature) Short() string {
	switch f {
	case featureNone:
		return "none"
	case featureInheritedAttributes:
		return "inherited-attributes"
	default:
		return fmt.Sprintf("%d", int(f))
	}
}

// String returns the string representation of an ExpFeature.
func (f expFeature) String() string {
	switch f {
	case featureNone:
		return "FeatureNone"
	case featureInheritedAttributes:
		return "FeatureInheritedAttributes"
	default:
		return fmt.Sprintf("ExpFeature(%d)", int(f))
	}
}

// parseShortExpFeature parses an ExpFeature from the given short string. The
// short string may be a string consisting of the same string as the Short()
// method returns, or be an integer represented as a string that corresponds to
// the desired feature.
func parseShortExpFeature(s string) (expFeature, error) {
	sLower := strings.ToLower(s)

	// if it is one of the short strings, use that
	for _, f := range expFeatureAll() {
		if sLower == f.Short() {
			return f, nil
		}
	}

	// else, try to parse an int from it
	intVal, err := strconv.Atoi(s)
	if err != nil {
		return featureNone, fmt.Errorf("not a feature short name and not an int: %q", s)
	}

	for _, f := range expFeatureAll() {
		if intVal == int(f) {
			return f, nil
		}
	}

	return featureNone, fmt.Errorf("not a valid feature type code: %d", intVal)
}
