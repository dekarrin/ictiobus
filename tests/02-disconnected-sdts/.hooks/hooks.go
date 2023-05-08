// Package hooks contains a set of hooks for the simplemath expression language.
package hooks

import (
	"fmt"
	"strconv"

	"github.com/dekarrin/ictiobus/trans"
)

var (
	HooksTable = trans.HookMap{
		"int":          hookInt,
		"identity":     hookIdentity,
		"add":          hookAdd,
		"mult":         hookMult,
		"lookup_value": hookLookupValue,
		"constant-1":   hookConstant(1),
		"constant-2":   hookConstant(2),
	}
)

func hookConstant(val interface{}) trans.Hook {
	return func(info trans.SetterInfo, args []interface{}) (interface{}, error) { return val, nil }
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
