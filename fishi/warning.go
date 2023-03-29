package fishi

type WarnType int

const (
	WarnDuplicateHumanDefs WarnType = iota
	WarnMissingHumanDef
	WarnPriorityZero
	WarnUnusedTerminal
)

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
