// Package fm contains a scripting engine for FISHIMath that uses an Ictiobus
// frontend for its analysis phase.
package fm

import (
	"fmt"
	"io"

	"github.com/dekarrin/ictfishimath_ast/fmfront"
	"github.com/dekarrin/ictfishimath_ast/fmhooks"
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/syntaxerr"
)

// Interpreter is a scripting engine that parses and executes FISHIMath
// statements.
type Interpreter struct {

	// Variables is the currently active variables in the interpreter. Any
	// values set here will be reflected during the next call to Exec.
	Variables map[string]fmhooks.FMValue

	// LastResult is the result of the last statement that was successfully
	// executed.
	LastResult fmhooks.FMValue

	// File is the name of the file currently being executed by the engine. This
	// is used in error reporting and is optional to set.
	File string

	// InitialVars is what Variables is set to whenenver Reset is called. It
	// will not be modified by use of this Interpreter.
	InitialVars map[string]fmhooks.FMValue

	fe ictiobus.Frontend[fmhooks.AST]
}

// InitEnvironment initializes the interpreter environment. All defined symbols
// and variables are removed and reset to those defined in InitialVars, and
// LastResult is reset. interp.File is not modified.
func (interp *Interpreter) InitEnvironment() {
	interp.Variables = map[string]fmhooks.FMValue{}
	interp.LastResult = fmhooks.FMValue{}

	if interp.InitialVars != nil {
		for k := range interp.InitialVars {
			interp.Variables[k] = interp.InitialVars[k]
		}
	}
}

// Eval parses the given string as FISHIMath code and applies it immediately.
// Returns a non-nil error if there is a syntax error in the text. The value of
// the last valid statement will be in interp.LastResult after Eval returns.
func (interp *Interpreter) Eval(code string) error {
	ast, err := interp.Parse(code)
	if err != nil {
		return err
	}

	interp.Exec(ast)
	return nil
}

// EvalReader parses the contents of a Reader as FISHIMath code and applies it
// immediately. Returns a non-nil error if there is a syntax error in the text
// or if there is an error reading bytes from the Reader. The value of the last
// valid statement will be in interp.LastResult after EvalReader returns.
func (interp *Interpreter) EvalReader(r io.Reader) error {
	ast, err := interp.ParseReader(r)
	if err != nil {
		return err
	}

	interp.Exec(ast)
	return nil
}

// Exec executes all mathematical statements contained in the AST and returns
// the result of the last statement. Additionally, interp.LastResult is set to
// that result. If no statements are in the AST, the returned FMValue will be
// the zero value and interp.LastResult will not be altered.
func (interp *Interpreter) Exec(ast fmhooks.AST) fmhooks.FMValue {
	if interp.Variables == nil {
		interp.Variables = map[string]fmhooks.FMValue{}
	}

	if len(ast.Statements) < 1 {
		return fmhooks.FMValue{}
	}

	var lastResult fmhooks.FMValue
	for i := range ast.Statements {
		stmt := ast.Statements[i]
		lastResult = interp.execNode(stmt)
	}

	return lastResult
}

// execNode executes the mathematical expression contained in the AST node and
// returns the result of the final one. This will also set interp.LastResult to
// that value.
func (interp *Interpreter) execNode(n fmhooks.Node) (result fmhooks.FMValue) {
	defer func() {
		interp.LastResult = result
	}()

	switch n.Type() {
	case fmhooks.Assignment:
		an := n.AsAssignment()
		name := an.Name
		value := interp.execNode(an.Expr)
		interp.Variables[name] = value
		return value
	case fmhooks.BinaryOp:
		bon := n.AsBinaryOp()
		left := interp.execNode(bon.Left)
		right := interp.execNode(bon.Right)

		switch bon.Op {
		case fmhooks.Add:
			return left.Add(right)
		case fmhooks.Subtract:
			return left.Subtract(right)
		case fmhooks.Multiply:
			return left.Multiply(right)
		case fmhooks.Divide:
			return left.Divide(right)
		default:
			panic(fmt.Sprintf("should never happen: unknown operation type: %v", bon.Op))
		}
	case fmhooks.Group:
		gn := n.AsGroup()
		return interp.execNode(gn.Expr)
	case fmhooks.Literal:
		ln := n.AsLiteral()
		return ln.Value
	case fmhooks.Variable:
		vn := n.AsVariable()
		return interp.Variables[vn.Name]
	default:
		panic(fmt.Sprintf("should never happen: unknown AST node type: %v", n.Type()))
	}
}

// Parse parses (but does not execute) FISHIMath code. The code is converted
// into an AST for further examination.
func (interp *Interpreter) Parse(code string) (ast fmhooks.AST, err error) {
	interp.initFrontend()

	ast, _, err = interp.fe.AnalyzeString(code)
	if err != nil {

		// wrap syntax errors so user of the Interpreter doesn't have to check
		// for a special syntax error just to get the detailed syntax err info
		if synErr, ok := err.(*syntaxerr.Error); ok {
			return ast, fmt.Errorf("%s", synErr.MessageForFile(interp.File))
		}
	}

	return ast, err
}

// ParseReader parses (but does not execute) FISHIMath code in the given reader.
// The entire contents of the Reader are read as FM code, which is returned as
// an AST for further examination.
func (interp *Interpreter) ParseReader(r io.Reader) (ast fmhooks.AST, err error) {
	interp.initFrontend()

	ast, _, err = interp.fe.Analyze(r)
	if err != nil {

		// wrap syntax errors so user of the Interpreter doesn't have to check
		// for a special syntax error just to get the detailed syntax err info
		if synErr, ok := err.(*syntaxerr.Error); ok {
			return ast, fmt.Errorf("%s", synErr.MessageForFile(interp.File))
		}
	}

	return ast, err
}

// initializes the frontend in member fe so that it can be used. If frontend is
// already initialized, this function does nothing. interp.fe can be safely used
// after calling this function.
func (interp *Interpreter) initFrontend() {
	// if IR attribute is blank, fe is by-extension not yet set, because
	// Ictiobus-generated frontends will never have an empty IRAttribute.
	if interp.fe.IRAttribute == "" {
		interp.fe = fmfront.Frontend(fmhooks.HooksTable, nil)
	}
}
