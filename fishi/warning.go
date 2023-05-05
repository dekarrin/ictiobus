package fishi

import (
	"fmt"
	"strconv"
	"strings"
)

type WarnType int

const (
	WarnNone WarnType = iota
	WarnDuplicateHumanDefs
	WarnMissingHumanDef
	WarnPriorityZero
	WarnUnusedTerminal
	WarnAmbiguousGrammar
	WarnValidation
	WarnValidationArgs
	WarnImportInference
)

// WarnTypeAll() returns a slice of all the WarnType constants. Must be added
// manually.
func WarnTypeAll() []WarnType {
	wts := []WarnType{
		WarnNone,
		WarnDuplicateHumanDefs,
		WarnPriorityZero,
		WarnUnusedTerminal,
		WarnAmbiguousGrammar,
		WarnValidation,
		WarnValidationArgs,
		WarnImportInference,
	}

	return wts
}

func (wt WarnType) Short() string {
	switch wt {
	case WarnNone:
		return "none"
	case WarnDuplicateHumanDefs:
		return "dupe_human"
	case WarnMissingHumanDef:
		return "missing_human"
	case WarnPriorityZero:
		return "priority"
	case WarnUnusedTerminal:
		return "unused"
	case WarnAmbiguousGrammar:
		return "ambig"
	case WarnValidation:
		return "validation"
	case WarnImportInference:
		return "import"
	case WarnValidationArgs:
		return "val_args"
	default:
		return fmt.Sprintf("%d", int(wt))
	}
}

func (wt WarnType) String() string {
	switch wt {
	case WarnNone:
		return "WarnNone"
	case WarnDuplicateHumanDefs:
		return "WarnDuplicateHumanDefs"
	case WarnMissingHumanDef:
		return "WarnMissingHumanDef"
	case WarnPriorityZero:
		return "WarnPriorityZero"
	case WarnUnusedTerminal:
		return "WarnUnusedTerminal"
	case WarnAmbiguousGrammar:
		return "WarnAmbiguousGrammar"
	case WarnValidation:
		return "WarnValidation"
	case WarnImportInference:
		return "WarnImportInference"
	case WarnValidationArgs:
		return "WarnValidationArgs"
	default:
		return fmt.Sprintf("WarnType(%d)", int(wt))
	}
}

// ParseShortWarnType parses a WarnType from the given short string. The short
// string may be a string consisting of the same string as the Short() method
// returns, or be an integer represented as a string that corresponds to the
// desired warning.
func ParseShortWarnType(s string) (WarnType, error) {
	sLower := strings.ToLower(s)

	// if it is one of the short strings, use that
	for _, wt := range WarnTypeAll() {
		if sLower == wt.Short() {
			return wt, nil
		}
	}

	// else, try to parse an int from it
	intVal, err := strconv.Atoi(s)
	if err != nil {
		return WarnNone, fmt.Errorf("not a warning short name and not an int: %q", s)
	}

	for _, wt := range WarnTypeAll() {
		if intVal == int(wt) {
			return wt, nil
		}
	}

	return WarnNone, fmt.Errorf("not a valid warning type code: %d", intVal)
}

// Warning is a warning that is generated when processing an AST. It is not an
// error per-se, but fulfills the error interface so that it can be treated as
// one by caller if desired.
type Warning struct {
	Type    WarnType
	Message string
}

func (w Warning) Error() string {
	return w.Message
}
