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
		"identity":          hookIdentity,
		"int":               hookInt,
		"float":             hookFloat,
		"multiply":          hookMultiply,
		"divide":            hookDivide,
		"add":               hookAdd,
		"subtract":          hookSubtract,
		"read_var":          hookReadVar,
		"write_var":         hookWriteVar,
		"num_slice_start":   hookNumSliceStart,
		"num_slice_prepend": hookNumSlicePrepend,
	}
)

var (
	symbolTable = map[string]FMValue{}
)

func hookIdentity(_ trans.SetterInfo, args []interface{}) (interface{}, error) { return args[0], nil }

func hookFloat(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	literalText, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	f64Val, err := strconv.ParseFloat(literalText, 32)
	if err != nil {
		return nil, err
	}
	fVal := float32(f64Val)

	return FMFloat(fVal), nil
}

func hookInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	literalText, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	iVal, err := strconv.Atoi(literalText)
	if err != nil {
		return nil, err
	}

	return FMInt(iVal), nil
}

func hookMultiply(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	v1, v2, err := getBinaryArgsCoerced(args)
	if err != nil {
		return nil, err
	}

	if v1.IsFloat {
		return FMFloat(v1.Float() * v2.Float()), nil
	}
	return FMInt(v1.Int() * v2.Int()), nil
}

func hookDivide(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	v1, v2, err := getBinaryArgsCoerced(args)
	if err != nil {
		return nil, err
	}

	// if one of them is a float (which will have them both coerced to float),
	// OR if we're about to divide by zero, do IEEE-754 math.
	if v1.IsFloat || v2.Int() == 0 {
		return FMFloat(v1.Float() / v2.Float()), nil
	}

	return FMInt(v1.Int() / v2.Int()), nil
}

func hookAdd(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	v1, v2, err := getBinaryArgsCoerced(args)
	if err != nil {
		return nil, err
	}

	if v1.IsFloat {
		return FMFloat(v1.Float() + v2.Float()), nil
	}
	return FMInt(v1.Int() + v2.Int()), nil
}

func hookSubtract(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	v1, v2, err := getBinaryArgsCoerced(args)
	if err != nil {
		return nil, err
	}

	if v1.IsFloat {
		return FMFloat(v1.Float() - v2.Float()), nil
	}
	return FMInt(v1.Int() - v2.Int()), nil
}

func hookReadVar(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	varName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	varVal := symbolTable[varName]

	return varVal, nil
}

func hookWriteVar(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	varName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not a string")
	}

	varVal, ok := args[1].(FMValue)
	if !ok {
		return nil, fmt.Errorf("arg 2 is not an FMValue")
	}

	symbolTable[varName] = varVal

	return varVal, nil
}

func hookNumSliceStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	v, ok := args[0].(FMValue)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not an FMValue")
	}

	return []FMValue{v}, nil
}

func hookNumSlicePrepend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	vSlice, ok := args[0].([]FMValue)
	if !ok {
		return nil, fmt.Errorf("arg 1 is not an []FMValue")
	}

	v, ok := args[1].(FMValue)
	if !ok {
		return nil, fmt.Errorf("arg 2 is not an FMValue")
	}

	vSlice = append([]FMValue{v}, vSlice...)

	return vSlice, nil
}

func getBinaryArgsCoerced(args []interface{}) (left, right FMValue, err error) {
	v1, ok := args[0].(FMValue)
	if !ok {
		return left, right, fmt.Errorf("arg 1 is not an FMValue")
	}

	v2, ok := args[1].(FMValue)
	if !ok {
		return left, right, fmt.Errorf("arg 2 is not an FMValue")
	}

	// if one is a float, they are now both floats
	if v1.IsFloat && !v2.IsFloat {
		v2 = FMFloat(v2.Float())
	} else if v2.IsFloat && !v1.IsFloat {
		v1 = FMFloat(v1.Float())
	}

	return v1, v2, nil
}

// FMValue is a calculated result from FISHIMath. It holds either a float32 or
// int and is convertible to either. The type of value it holds is querable with
// IsFloat. Int() or Float() can be called on it to get the value as that type.
type FMValue struct {
	IsFloat bool
	i       int
	f       float32
}

// FMFloat creates a new FMValue that holds a float32 value.
func FMFloat(v float32) FMValue {
	return FMValue{IsFloat: true, f: v}
}

// FMInt creates a new FMValue that holds an int value.
func FMInt(v int) FMValue {
	return FMValue{i: v}
}

// Int returns the value of v as an int, converting if necessary from a float.
func (v FMValue) Int() int {
	if v.IsFloat {
		return int(math.Round(float64(v.f)))
	}
	return v.i
}

// Float returns the value of v as a float32, converting if necessary from an
// int.
func (v FMValue) Float() float32 {
	if !v.IsFloat {
		return float32(v.i)
	}
	return v.f
}

// String returns the string representation of an FMValue.
func (v FMValue) String() string {
	if v.IsFloat {
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
