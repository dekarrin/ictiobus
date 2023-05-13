package lex

// ActionType is a type of action that the lexer can take.
type ActionType int

const (
	ActionNone ActionType = iota
	ActionScan
	ActionState
	ActionScanAndState
)

// Action is an action for the lexer to take when it matches a defined regex
// pattern.
type Action struct {
	Type    ActionType
	ClassID string
	State   string
}

// SwapState returns a lexer action that indicates that the lexer should swap
// to the given state.
func SwapState(toState string) Action {
	return Action{
		Type:  ActionState,
		State: toState,
	}
}

// LexAs returns a lexer action that indicates that the lexer should take the
// source text that it matched against and lex it as a token of the given token
// class.
func LexAs(classID string) Action {
	return Action{
		Type:    ActionScan,
		ClassID: classID,
	}
}

// LexAndSwapState returns a lexer action that indicates that the lexer should
// take the source text that it matched against and lex it as a token of the
// given token class, and then it should swap to the new state.
func LexAndSwapState(classID string, newState string) Action {
	return Action{
		Type:    ActionScanAndState,
		ClassID: classID,
		State:   newState,
	}
}

// Discard returns a lexer action that indicates that it should take no action
// and effectively discard the text it matched against.
func Discard() Action {
	return Action{}
}
