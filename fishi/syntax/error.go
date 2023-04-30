package syntax

import "fmt"

type HookArgError struct {
	ArgNum  int
	Args    []interface{}
	Message string
}

func (e *HookArgError) Error() string {
	return fmt.Sprintf("arg[%d]: %s", e.ArgNum, e.Message)
}

func NewArgTypeError(args []interface{}, argNum int, expectedType string) *HookArgError {
	return NewArgError(args, argNum, "expected type to be %s, got %T", expectedType, args[argNum])
}

func NewArgError(args []interface{}, argNum int, msg string, a ...interface{}) *HookArgError {
	if len(a) > 0 {
		msg = fmt.Sprintf(msg, a...)
	}
	return &HookArgError{
		ArgNum:  argNum,
		Args:    args,
		Message: msg,
	}
}
