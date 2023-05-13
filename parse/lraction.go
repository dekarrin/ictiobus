package parse

import (
	"fmt"

	"github.com/dekarrin/ictiobus/grammar"
)

func isShiftReduceConlict(act1, act2 lrAction) (isSR bool, shiftAct lrAction) {
	if act1.Type == lrReduce && act2.Type == lrShift {
		return true, act2
	}
	if act2.Type == lrReduce && act1.Type == lrShift {
		return true, act1
	}

	return false, act1
}

func makeLRConflictError(act1, act2 lrAction, onInput string) error {
	if act1.Type == lrReduce && act2.Type == lrShift || act1.Type == lrShift && act2.Type == lrReduce {
		// shift-reduce conflict

		reduceRule := ""
		if act1.Type == lrReduce {
			reduceRule = act1.Symbol + " -> " + act1.Production.String()
		} else {
			reduceRule = act2.Symbol + " -> " + act2.Production.String()
		}
		return fmt.Errorf("shift/reduce conflict detected on terminal %q (shift or reduce %s)", onInput, reduceRule)
	} else if act1.Type == lrReduce && act2.Type == lrReduce {
		// reduce-reduce conflict

		reduce1 := act1.Symbol + " -> " + act1.Production.String()
		reduce2 := act2.Symbol + " -> " + act2.Production.String()
		return fmt.Errorf("reduce/reduce conflict detected on terminal %q (reduce %s or reduce %s)", onInput, reduce1, reduce2)
	} else if act1.Type == lrAccept || act2.Type == lrAccept {
		nonAcceptAct := act2

		if act2.Type == lrAccept {
			nonAcceptAct = act1
		}

		// accept-? conflict
		if nonAcceptAct.Type == lrShift {
			return fmt.Errorf("accept/shift conflict detected on terminal %q", onInput)
		} else if nonAcceptAct.Type == lrReduce {
			reduce := nonAcceptAct.Symbol + " -> " + nonAcceptAct.Production.String()
			return fmt.Errorf("accept/reduce conflict detected on terminal %q (accept or reduce %s)", onInput, reduce)
		}
	} else if act1.Type == lrShift && act2.Type == lrShift {
		return fmt.Errorf("(!) shift/shift conflict on terminal %q", onInput)
	}
	return fmt.Errorf("LR action conflict on terminal %q (%s or %s)", onInput, act1.String(), act2.String())
}

// lrActionType is a type of action for a shift-reduce LR-parser to perform.
type lrActionType int

const (
	lrShift lrActionType = iota
	lrReduce
	lrAccept
	lrError
)

// String returns the string representation of an LRActionType.
func (lt lrActionType) String() string {
	switch lt {
	case lrShift:
		return "SHIFT"
	case lrReduce:
		return "REDUCE"
	case lrAccept:
		return "ACCEPT"
	case lrError:
		return "ERROR"
	default:
		return fmt.Sprintf("LRActionType<%d>", lt)
	}
}

// lrAction is an action for a shift-reduce LR parser to perform. At any point,
// it decides based on what symbols have been seen and a DFA whether to shift an
// an input symbol onto the stack, reduce the currently read set of input
// symbols, accept the complete input, or produce an error.
type lrAction struct {
	// Type is the type of the LRAction.
	Type lrActionType

	// Production is used when Type is LRReduce. It is the production which
	// should be reduced; the β of A -> β.
	Production grammar.Production

	// Symbol is used when Type is LRReduce. It is the symbol to reduce the
	// production to; the A of A -> β.
	Symbol string

	// State is the state to shift to. It is used only when Type is LRShift.
	State string
}

// String returns a string representation of the LRAction.
func (act lrAction) String() string {
	switch act.Type {
	case lrAccept:
		return "ACTION<accept>"
	case lrError:
		return "ACTION<error>"
	case lrReduce:
		return fmt.Sprintf("ACTION<reduce %s -> %s>", act.Symbol, act.Production.String())
	case lrShift:
		return fmt.Sprintf("ACTION<shift %s>", act.State)
	default:
		return "ACTION<unknown>"
	}
}

// Equal returns true if o is an LRAction or a *LRAction whose properties are
// the same as act.
func (act lrAction) Equal(o any) bool {
	other, ok := o.(lrAction)
	if !ok {
		otherPtr := o.(*lrAction)
		if !ok {
			return false
		}
		if otherPtr == nil {
			return false
		}
		other = *otherPtr
	}

	if act.Type != other.Type {
		return false
	} else if !act.Production.Equal(other.Production) {
		return false
	} else if act.State != other.State {
		return false
	} else if act.Symbol != other.Symbol {
		return false
	}

	return true
}
