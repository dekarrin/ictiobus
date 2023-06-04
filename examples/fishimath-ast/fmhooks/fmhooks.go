package fmhooks

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/trans"
)

var (
	HooksTable = trans.HookMap{
		"identity":           hookIdentity,
		"var_node":           hookVarNode,
		"assignment_node":    hookAssignmentNode,
		"lit_node_float":     hookLitNodeFloat,
		"lit_node_int":       hookLitNodeInt,
		"group_node":         hookGroupNode,
		"binary_node_mult":   hookFnForBinaryNode(Multiply),
		"binary_node_div":    hookFnForBinaryNode(Divide),
		"binary_node_add":    hookFnForBinaryNode(Add),
		"binary_node_sub":    hookFnForBinaryNode(Subtract),
		"node_slice_start":   hookNodeSliceStart,
		"node_slice_prepend": hookNodeSlicePrepend,
		"ast":                hookAST,
	}
)

func hookIdentity(_ trans.SetterInfo, args []interface{}) (interface{}, error) { return args[0], nil }

func hookAST(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	nodeSlice, ok := args[0].([]Node)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a []Node")
	}

	return AST{Statements: nodeSlice}, nil
}

func hookVarNode(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	varName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	return VariableNode{Name: varName}, nil
}

func hookAssignmentNode(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	varName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	varVal, ok := args[1].(Node)
	if !ok {
		return nil, fmt.Errorf("arg 2 is not a Node")
	}

	return AssignmentNode{Name: varName, Expr: varVal}, nil
}

func hookLitNodeFloat(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	strVal, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	f64Val, err := strconv.ParseFloat(strVal, 32)
	if err != nil {
		return nil, err
	}

	return LiteralNode{Value: FMValue{vType: Float, f: float32(f64Val)}}, nil
}

func hookLitNodeInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	strVal, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	iVal, err := strconv.Atoi(strVal)
	if err != nil {
		return nil, err
	}

	return LiteralNode{Value: FMValue{vType: Int, i: iVal}}, nil
}

func hookGroupNode(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	exprNode, ok := args[0].(Node)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a Node")
	}

	return GroupNode{Expr: exprNode}, nil
}

func hookFnForBinaryNode(op Operation) trans.Hook {
	fn := func(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
		left, ok := args[0].(Node)
		if !ok {
			return nil, fmt.Errorf("arg 1 is not a Node")
		}

		right, ok := args[1].(Node)
		if !ok {
			return nil, fmt.Errorf("arg 2 is not a Node")
		}

		return BinaryOpNode{
			Left:  left,
			Right: right,
			Op:    op,
		}, nil
	}

	return fn
}

func hookNodeSliceStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	node, ok := args[0].(Node)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a Node")
	}

	return []Node{node}, nil
}

func hookNodeSlicePrepend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	nodeSlice, ok := args[0].([]Node)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a []Node")
	}

	node, ok := args[1].(Node)
	if !ok {
		return nil, fmt.Errorf("arg 2 is not a Node")
	}

	nodeSlice = append([]Node{node}, nodeSlice...)

	return nodeSlice, nil
}

const astTabAmount = 2

// AST is an abstract syntax tree containing a complete representation of input
// written in FISHIMath.
type AST struct {
	Statements []Node
}

// String returns a pretty-print representation of the AST, with depth in the
// tree indicated by indent.
func (ast AST) String() string {
	if len(ast.Statements) < 1 {
		return "AST<>"
	}

	var sb strings.Builder
	sb.WriteString("AST<\n")

	labelFmt := "  STMT #%0*d: "
	largestDigitCount := len(fmt.Sprintf("%d", len(ast.Statements)))

	for i := range ast.Statements {
		label := fmt.Sprintf(labelFmt, largestDigitCount, i+1)
		stmtStr := spaceIndentNewlines(ast.Statements[i].String(), astTabAmount)
		sb.WriteString(label)
		sb.WriteString(stmtStr)
		sb.WriteRune('\n')
	}

	sb.WriteRune('>')
	return sb.String()
}

// FMString returns a string of FISHIMath code that if parsed, would result in
// an equivalent AST. Each statement is put on its own line, but the last line
// will not end in \n.
func (ast AST) FMString() string {
	if len(ast.Statements) < 1 {
		return ""
	}

	var sb strings.Builder

	for i, stmt := range ast.Statements {
		sb.WriteString(stmt.FMString())
		sb.WriteString(" <o^><")

		if i+1 < len(ast.Statements) {
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

// NodeType is the type of a node in the AST.
type NodeType int

const (
	// Literal is a numerical value literal in FISHIMath, such as 3 or 7.284.
	// FISHIMath only supports float32 and int literals, so it will represent
	// one of those.
	Literal NodeType = iota

	// Variable is a variable used in a non-assignment context (i.e. one whose
	// value is being *used*, not set).
	Variable

	// BinaryOp is an operation consisting of two operands and the operation
	// performed on them.
	BinaryOp

	// Assignment is an assignment of a value to a variable in FISHIMath using
	// the value tentacle.
	Assignment

	// Group is an expression grouped with the fishtail and fishhead symbols.
	Group
)

// Operation is a type of operation being performed.
type Operation int

const (
	NoOp Operation = iota
	Add
	Subtract
	Multiply
	Divide
)

// Symbol returns the FISHIMath syntax string that represents the operation. If
// there isn't one for op, "" is returned.
func (op Operation) Symbol() string {
	if op == Add {
		return "+"
	} else if op == Subtract {
		return "-"
	} else if op == Multiply {
		return "*"
	} else if op == Divide {
		return "/"
	}

	return ""
}

func (op Operation) String() string {
	if op == Add {
		return "addition"
	} else if op == Subtract {
		return "subtraction"
	} else if op == Multiply {
		return "multiplication"
	} else if op == Divide {
		return "division"
	}

	return fmt.Sprintf("operation(code=%d)", int(op))
}

// Node is a node of the AST. It can be converted to the actual type it is by
// calling the appropriate function.
type Node interface {
	// Type returns the type of thing this node is. Whether or not the other
	// functions return valid values depends on the type of AST Node returned
	// here.
	Type() NodeType

	// AsLiteral returns this Node as a LiteralNode. Panics if Type() is not
	// Literal.
	AsLiteral() LiteralNode

	// AsVariable returns this Node as a VariableNode. Panics if Type() is not
	// Variable.
	AsVariable() VariableNode

	// AsBinaryOperation returns this Node as a BinaryOpNode. Panics if Type()
	// is not BinaryOp.
	AsBinaryOp() BinaryOpNode

	// AsAssignment returns this Node as an AssignmentNode. Panics if Type() is
	// not Assignment.
	AsAssignment() AssignmentNode

	// AsGroup returns this Node as a GroupNode. Panics if Type() is not Group.
	AsGroup() GroupNode

	// FMString converts this Node into FISHIMath code that would produce an
	// equivalent Node.
	FMString() string

	// String returns a human-readable string representation of this Node, which
	// will vary based on what Type() is.
	String() string
}

// LiteralNode is an AST node representing a numerical constant used in
// FISHIMath.
type LiteralNode struct {
	Value FMValue
}

func (n LiteralNode) Type() NodeType               { return Literal }
func (n LiteralNode) AsLiteral() LiteralNode       { return n }
func (n LiteralNode) AsVariable() VariableNode     { panic("Type() is not Variable") }
func (n LiteralNode) AsBinaryOp() BinaryOpNode     { panic("Type() is not BinaryOp") }
func (n LiteralNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n LiteralNode) AsGroup() GroupNode           { panic("Type() is not Group") }

func (n LiteralNode) FMString() string {
	return n.Value.String()
}

func (n LiteralNode) String() string {
	return fmt.Sprintf("[LITERAL value=%v]", n.Value)
}

// VariableNode is an AST node representing the use of a variable's value in
// FISHIMath. It does *not* represent assignment to a variable, as that is done
// with an AssignmentNode.
type VariableNode struct {
	Name string
}

func (n VariableNode) Type() NodeType               { return Variable }
func (n VariableNode) AsLiteral() LiteralNode       { panic("Type() is not Literal") }
func (n VariableNode) AsVariable() VariableNode     { return n }
func (n VariableNode) AsBinaryOp() BinaryOpNode     { panic("Type() is not BinaryOp") }
func (n VariableNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n VariableNode) AsGroup() GroupNode           { panic("Type() is not Group") }

func (n VariableNode) FMString() string {
	return n.Name
}

func (n VariableNode) String() string {
	return fmt.Sprintf("[VARIABLE name=%v]", n.Name)
}

// BinaryOpNode is an AST node representing a binary operation in FISHIMath. It
// has a left operand, a right operand, and an operation to perform on them.
type BinaryOpNode struct {
	Left  Node
	Right Node
	Op    Operation
}

func (n BinaryOpNode) Type() NodeType               { return BinaryOp }
func (n BinaryOpNode) AsLiteral() LiteralNode       { panic("Type() is not Literal") }
func (n BinaryOpNode) AsVariable() VariableNode     { panic("Type() is not Variable") }
func (n BinaryOpNode) AsBinaryOp() BinaryOpNode     { return n }
func (n BinaryOpNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n BinaryOpNode) AsGroup() GroupNode           { panic("Type() is not Group") }

func (n BinaryOpNode) FMString() string {
	return fmt.Sprintf("%s %s %s", n.Left.FMString(), n.Op.Symbol(), n.Right.FMString())
}

func (n BinaryOpNode) String() string {
	const (
		leftStart  = "  left:  "
		rightStart = "  right: "
	)

	leftStr := spaceIndentNewlines(n.Left.String(), astTabAmount)
	rightStr := spaceIndentNewlines(n.Right.String(), astTabAmount)

	return fmt.Sprintf("[BINARY_OPERATION type=%v\n%s%s\n%s%s\n]", n.Op.String(), leftStart, leftStr, rightStart, rightStr)
}

// AssignmentNode is an AST node representing the assignment of an expression to
// a variable in FISHIMath. Name is the name of the variable, Expr is the
// expression being assigned to it.
type AssignmentNode struct {
	Name string
	Expr Node
}

func (n AssignmentNode) Type() NodeType               { return Assignment }
func (n AssignmentNode) AsLiteral() LiteralNode       { panic("Type() is not Literal") }
func (n AssignmentNode) AsVariable() VariableNode     { panic("Type() is not Variable") }
func (n AssignmentNode) AsBinaryOp() BinaryOpNode     { panic("Type() is not BinaryOp") }
func (n AssignmentNode) AsAssignment() AssignmentNode { return n }
func (n AssignmentNode) AsGroup() GroupNode           { panic("Type() is not Group") }

func (n AssignmentNode) FMString() string {
	return fmt.Sprintf("%s =o %s", n.Name, n.Expr.FMString())
}

func (n AssignmentNode) String() string {
	const (
		exprStart = "  expr:  "
	)

	exprStr := spaceIndentNewlines(n.Expr.String(), astTabAmount)

	return fmt.Sprintf("[ASSIGNMENT name=%q\n%s%s\n]", n.Name, exprStart, exprStr)
}

// GroupNode is an AST node representing an expression grouped by the fishtail
// and fishhead symbols in FISHIMath. Expr is the expression in the group.
type GroupNode struct {
	Expr Node
}

func (n GroupNode) Type() NodeType               { return Assignment }
func (n GroupNode) AsLiteral() LiteralNode       { panic("Type() is not Literal") }
func (n GroupNode) AsVariable() VariableNode     { panic("Type() is not Variable") }
func (n GroupNode) AsBinaryOp() BinaryOpNode     { panic("Type() is not BinaryOp") }
func (n GroupNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n GroupNode) AsGroup() GroupNode           { return n }

func (n GroupNode) FMString() string {
	return fmt.Sprintf(">{ %s '}", n.Expr.FMString())
}

func (n GroupNode) String() string {
	const (
		exprStart = "  expr:  "
	)

	exprStr := spaceIndentNewlines(n.Expr.String(), astTabAmount)

	return fmt.Sprintf("[GROUP\n%s%s\n]", exprStart, exprStr)
}

func spaceIndentNewlines(str string, amount int) string {
	if strings.Contains(str, "\n") {
		// need to pad every newline
		pad := " "
		for len(pad) < amount {
			pad += " "
		}
		str = strings.ReplaceAll(str, "\n", "\n"+pad)
	}
	return str
}

// ValueType is the type of a value in FISHIMath. Only Float and Int are
// supported.
type ValueType int

const (
	// Int is an integer of at least 32 bits.
	Int ValueType = iota

	// Float is a floating point number represented as an IEEE-754 single
	// precision (32-bit) float.
	Float
)

// FMValue is a typed value used in FISHIMath. The type of value it holds is
// querable with Type(). Int() or Float() can be called on it to get the value
// as that type, otherwise Interface() can be called to return the exact value
// as whatever type it is.
type FMValue struct {
	vType ValueType
	i     int
	f     float32
}

// Int returns the value of v as an int, converting if Type() is not Int.
func (v FMValue) Int() int {
	if v.vType == Float {
		return int(math.Round(float64(v.f)))
	}
	return v.i
}

// Float returns the value of v as a float32, converting if Type() is not Float.
func (v FMValue) Float() float32 {
	if v.vType == Int {
		return float32(v.i)
	}
	return v.f
}

// Interface returns the value held within this as its native Go type.
func (v FMValue) Interface() interface{} {
	if v.vType == Float {
		return v.f
	}
	return v.i
}

// Type returns the type of this value.
func (v FMValue) Type() ValueType {
	return v.vType
}

// String returns the string representation of an FMValue.
func (v FMValue) String() string {
	if v.vType == Float {
		str := fmt.Sprintf("%.7f", v.f)
		// remove extra 0's...
		str = strings.TrimRight(str, "0")
		// ...but there should be at least one 0 if nothing else
		if strings.HasSuffix(str, ".") {
			str = str + "0"
		}
		return str
	}
	return fmt.Sprintf("%d", v.i)
}
