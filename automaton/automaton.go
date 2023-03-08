package automaton

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dekarrin/ictiobus/internal/decbin"
	"github.com/dekarrin/ictiobus/internal/textfmt"
)

type FATransition struct {
	input string
	next  string
}

func (t FATransition) MarshalBinary() ([]byte, error) {
	data := decbin.EncString(t.input)
	data = append(data, decbin.EncString(t.next)...)
	return data, nil
}

func (t *FATransition) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	t.input, n, err = decbin.DecString(data)
	if err != nil {
		return fmt.Errorf(".input: %w", err)
	}
	data = data[n:]

	t.next, _, err = decbin.DecString(data)
	if err != nil {
		return fmt.Errorf(".next: %w", err)
	}

	return nil
}

func (t FATransition) String() string {
	inp := t.input
	if inp == "" {
		inp = "ε"
	}
	return fmt.Sprintf("=(%s)=> %s", inp, t.next)
}

func mustParseFATransition(s string) FATransition {
	t, err := parseFATransition(s)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func parseFATransition(s string) (FATransition, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, " ", 2)

	left, right := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

	if len(left) < 3 {
		return FATransition{}, fmt.Errorf("not a valid FATransition: left len < 3: %q", left)
	}

	if left[0] != '=' {
		return FATransition{}, fmt.Errorf("not a valid FATransition: left[0] != '=': %q", left)
	}
	if left[1] != '(' {
		return FATransition{}, fmt.Errorf("not a valid FATransition: left[1] != '(': %q", left)
	}
	left = left[2:]
	// also chop off the ending arrow
	if len(left) < 4 {
		return FATransition{}, fmt.Errorf("not a valid left: len(chopped) < 4: %q", left)
	}
	if left[len(left)-1] != '>' {
		return FATransition{}, fmt.Errorf("not a valid left: chopped[-1] != '>': %q", left)
	}
	if left[len(left)-2] != '=' {
		return FATransition{}, fmt.Errorf("not a valid left: chopped[-2] != '=': %q", left)
	}
	if left[len(left)-3] != ')' {
		return FATransition{}, fmt.Errorf("not a valid left: chopped[-3] != ')': %q", left)
	}
	input := left[:len(left)-3]
	if input == "ε" {
		input = ""
	}

	// next is EASY af
	next := right
	if next == "" {
		return FATransition{}, fmt.Errorf("not a valid FATransition: bad next: %q", s)
	}

	return FATransition{
		input: input,
		next:  next,
	}, nil
}

type DFAState[E any] struct {
	ordering    uint64
	name        string
	value       E
	transitions map[string]FATransition
	accepting   bool
}

func (ds DFAState[E]) MarshalBytes(conv func(E) []byte) []byte {
	data := decbin.EncInt(int(ds.ordering))
	data = append(data, decbin.EncString(ds.name)...)

	convData := conv(ds.value)

	data = append(data, decbin.EncInt(len(convData))...)
	data = append(data, convData...)
	data = append(data, decbin.EncMapStringToBinary(ds.transitions)...)
	data = append(data, decbin.EncBool(ds.accepting)...)
	return data
}

func UnmarshalDFAStateBytes[E any](data []byte, conv func([]byte) (E, error)) (DFAState[E], error) {
	var ds DFAState[E]
	var n int
	var err error

	var iVal int
	iVal, n, err = decbin.DecInt(data)
	if err != nil {
		return ds, fmt.Errorf(".ordering: %w", err)
	}
	data = data[n:]
	ds.ordering = uint64(iVal)

	ds.name, n, err = decbin.DecString(data)
	if err != nil {
		return ds, fmt.Errorf(".name: %w", err)
	}
	data = data[n:]

	var convLen int
	convLen, n, err = decbin.DecInt(data)
	if err != nil {
		return ds, fmt.Errorf("get value data len: %w", err)
	}
	data = data[n:]
	if len(data) < convLen {
		return ds, fmt.Errorf(".value: not enough bytes")
	}
	convData := data[:convLen]
	ds.value, err = conv(convData)
	if err != nil {
		return ds, fmt.Errorf(".value: %w", err)
	}
	data = data[convLen:]

	var ptrMap map[string]*FATransition
	ptrMap, n, err = decbin.DecMapStringToBinary[*FATransition](data)
	if err != nil {
		return ds, fmt.Errorf(".transitions: %w", err)
	}
	if ptrMap != nil {
		ds.transitions = map[string]FATransition{}
		for k := range ptrMap {
			if ptrMap[k] != nil {
				ds.transitions[k] = *ptrMap[k]
			} else {
				ds.transitions[k] = FATransition{}
			}
		}
	} else {
		ds.transitions = nil
	}
	data = data[n:]

	ds.accepting, _, err = decbin.DecBool(data)
	if err != nil {
		return ds, fmt.Errorf(".accepting: %w", err)
	}

	return ds, nil
}

func (ds DFAState[E]) Copy() DFAState[E] {
	copied := DFAState[E]{
		ordering:    ds.ordering,
		name:        ds.name,
		value:       ds.value,
		transitions: make(map[string]FATransition),
		accepting:   ds.accepting,
	}

	for k := range ds.transitions {
		copied.transitions[k] = ds.transitions[k]
	}

	return copied
}

func (ns DFAState[E]) String() string {
	var moves strings.Builder

	inputs := textfmt.OrderedKeys(ns.transitions)

	for i, input := range inputs {
		moves.WriteString(ns.transitions[input].String())
		if i+1 < len(inputs) {
			moves.WriteRune(',')
			moves.WriteRune(' ')
		}
	}

	str := fmt.Sprintf("(%s [%s])", ns.name, moves.String())

	if ns.accepting {
		str = "(" + str + ")"
	}

	return str
}

func (ns DFAState[E]) ValueString() string {
	var moves strings.Builder

	inputs := textfmt.OrderedKeys(ns.transitions)

	for i, input := range inputs {
		moves.WriteString(ns.transitions[input].String())
		if i+1 < len(inputs) {
			moves.WriteRune(',')
			moves.WriteRune(' ')
		}
	}

	str := fmt.Sprintf("(%s %q [%s])", ns.name, fmt.Sprintf("%v", ns.value), moves.String())

	if ns.accepting {
		str = "(" + str + ")"
	}

	return str
}

type NFAState[E any] struct {
	ordering    uint64
	name        string
	value       E
	transitions map[string][]FATransition
	accepting   bool
}

func (ns NFAState[E]) Copy() NFAState[E] {
	copied := NFAState[E]{
		ordering:    ns.ordering,
		name:        ns.name,
		value:       ns.value,
		transitions: make(map[string][]FATransition),
		accepting:   ns.accepting,
	}

	for k := range ns.transitions {
		trans := ns.transitions[k]
		transCopy := make([]FATransition, len(trans))
		copy(transCopy, trans)
		copied.transitions[k] = transCopy
	}

	return copied
}

func (ns NFAState[E]) String() string {
	var moves strings.Builder

	inputs := textfmt.OrderedKeys(ns.transitions)

	for i, input := range inputs {
		var tStrings []string

		for _, t := range ns.transitions[input] {
			tStrings = append(tStrings, t.String())
		}

		sort.Strings(tStrings)

		for tIdx, t := range tStrings {
			moves.WriteString(t)
			if tIdx+1 < len(tStrings) || i+1 < len(inputs) {
				moves.WriteRune(',')
				moves.WriteRune(' ')
			}
		}
	}

	str := fmt.Sprintf("(%s [%s])", ns.name, moves.String())

	if ns.accepting {
		str = "(" + str + ")"
	}

	return str
}

func (ns NFAState[E]) ValueString() string {
	var moves strings.Builder

	inputs := textfmt.OrderedKeys(ns.transitions)

	for i, input := range inputs {
		var tStrings []string

		for _, t := range ns.transitions[input] {
			tStrings = append(tStrings, t.String())
		}

		sort.Strings(tStrings)

		for tIdx, t := range tStrings {
			moves.WriteString(t)
			if tIdx+1 < len(tStrings) || i+1 < len(inputs) {
				moves.WriteRune(',')
				moves.WriteRune(' ')
			}
		}
	}

	str := fmt.Sprintf("(%s %q [%s])", ns.name, fmt.Sprintf("%v", ns.value), moves.String())

	if ns.accepting {
		str = "(" + str + ")"
	}

	return str
}
