package syntax

import "fmt"

// HooksArgError is returned by an SDTS hook function when there is a problem
// with one of its arguments. It should not be used directly; initialize it with
// [NewArgTypeError] or [NewArgError].
type hookArgError struct {
	ArgNum  int
	Args    []interface{}
	Message string
}

// Error returns a string representation of the error.
func (e *hookArgError) Error() string {
	return fmt.Sprintf("arg[%d]: %s", e.ArgNum, e.Message)
}

// newArgTypeError returns a new error that describes a type mismatch of an SDTS
// hook argument.
func newArgTypeError(args []interface{}, argNum int, expectedType string) *hookArgError {
	return newArgError(args, argNum, "expected type to be %s, got %T", expectedType, args[argNum])
}

// NewArgTypeError returns a new error that describes some kind of an error with
// an SDTS hook argument.
func newArgError(args []interface{}, argNum int, msg string, a ...interface{}) *hookArgError {
	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	return &hookArgError{
		ArgNum:  argNum,
		Args:    args,
		Message: msg,
	}
}
