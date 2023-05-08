// Package hooks contains a set of hooks for the simplemath expression language.
package hooks

import (
	"fmt"
	"strconv"

	"github.com/dekarrin/ictiobus/trans"
)

type opChain struct {
	Type string
	Arg  int
	Next *opChain
}

var (
	HooksTable = trans.HookMap{
		"int":          hookInt,
		"identity":     hookIdentity,
		"add":          hookAdd,
		"mult":         hookMult,
		"lookup_value": hookLookupValue,
		"mult_chain":   hookChain("mult"),
		"add_chain":    hookChain("add"),
		"empty_chain":  hookEmptyChain,
		"eval_chain":   hookEvalChain,
	}
)

func hookEvalChain(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	left, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("left arg is not an int: %v", args[0])
	}

	right, ok := args[1].(opChain)
	if !ok {
		return nil, fmt.Errorf("right arg is not an opchain: %v", args[1])
	}

	curVal := left

	next := &right
	for next != nil {
		if next.Type == "mult" {
			curVal *= next.Arg
		} else if next.Type == "add" {
			curVal += next.Arg
		} else if next.Type != "empty" {
			return nil, fmt.Errorf("chain type is not add, mult, or empty: %s", next.Type)
		}
		next = next.Next
	}

	return curVal, nil
}

func hookEmptyChain(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	return opChain{Type: "empty"}, nil
}

func hookChain(t string) trans.Hook {
	return func(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
		arg, ok := args[0].(int)
		if !ok {
			return nil, fmt.Errorf("first arg is not an int: %v", args[0])
		}

		next, ok := args[1].(opChain)
		if !ok {
			return nil, fmt.Errorf("second arg is not an opchain: %v", args[1])
		}

		return opChain{Type: t, Arg: arg, Next: &next}, nil
	}
}

func hookInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	intSeq, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("int() value is not a string: %v", args[0])
	}

	return strconv.Atoi(intSeq)
}

func hookIdentity(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	return args[0], nil
}

func hookLookupValue(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	varName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("var name is not a string: %v", args[0])
	}

	return len(varName), nil
}

func hookAdd(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	left, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("left side is not an int: %v", args[0])
	}

	right, ok := args[1].(int)
	if !ok {
		return nil, fmt.Errorf("right side is not an int: %v", args[1])
	}

	return left + right, nil
}

func hookMult(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
	left, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("left side is not an int: %v", args[0])
	}

	right, ok := args[1].(int)
	if !ok {
		return nil, fmt.Errorf("right side is not an int: %v", args[1])
	}

	return left * right, nil
}
