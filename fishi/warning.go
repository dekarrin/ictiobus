package fishi

import "fmt"

type WarnType int

const (
	WarnDuplicateHumanDefs WarnType = iota
	WarnMissingHumanDef
	WarnPriorityZero
	WarnUnusedTerminal
	WarnAmbiguousGrammar
	WarnValidation
)

func (wt WarnType) String() string {
	switch wt {
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
	default:
		return fmt.Sprintf("WarnType(%d)", int(wt))
	}
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
