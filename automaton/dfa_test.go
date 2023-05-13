package automaton

import (
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/stretchr/testify/assert"
)

func Test_DFA_MarshalUnmarshalBinary(t *testing.T) {
	type dummy struct {
		val1 string
		val2 int
	}

	testCases := []struct {
		name  string
		input DFA[dummy]
	}{
		{
			name:  "empty",
			input: DFA[dummy]{},
		},
		{
			name: "fully populated",
			input: DFA[dummy]{
				order: 285039842,
				Start: "Feferi Peixes",
				states: map[string]dfaState[dummy]{
					"Nepeta Leijon": {
						ordering: 28921,
						name:     "bizarrely long name",
						value: dummy{
							val1: "nepeta leijon",
							val2: 88888888,
						},
						transitions: map[string]faTransition{
							"a": {input: "a", next: "b"},
						},
						accepting: true,
					},
					"Feferi Peixes": {
						ordering: 413,
						name:     "Feferi Peixes",
						value: dummy{
							val1: "feferi peixes",
							val2: 6188,
						},
						transitions: map[string]faTransition{},
					},
					"Karkat Vantas": {
						ordering: 612,
						name:     "Karkat Vantas",
						value: dummy{
							val1: "karkat vantas",
							val2: 8888,
						},
						transitions: map[string]faTransition{
							"a": {input: "a", next: "Feferi Peixes"},
							"b": {input: "b", next: "Nepeta Leijon"},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			encoded := tc.input.MarshalBytes(func(d dummy) []byte {
				data := rezi.EncString(d.val1)
				data = append(data, rezi.EncInt(d.val2)...)
				return data
			})

			actual, err := UnmarshalDFABytes(encoded, func(data []byte) (dummy, error) {
				var d dummy
				var n int
				var err error

				d.val1, n, err = rezi.DecString(data)
				if err != nil {
					return d, fmt.Errorf(".val1: %w", err)
				}
				data = data[n:]

				d.val2, _, err = rezi.DecInt(data)
				if err != nil {
					return d, fmt.Errorf(".val2: %w", err)
				}

				return d, nil
			})
			if !assert.NoError(err, "UnmarshalDFAStateBytes failed") {
				return
			}

			// glub glub 38D v happy fishy
			assert.Equal(tc.input, actual)
		})
	}
}

func buildDFA(from map[string][]string, start string, acceptingStates []string) *DFA[string] {
	dfa := &DFA[string]{}

	acceptSet := box.StringSetOf(acceptingStates)

	for k := range from {
		dfa.AddState(k, acceptSet.Has(k))
		dfa.SetValue(k, k)
	}

	// add transitions AFTER all states are already in or it will cause a panic
	for k := range from {
		for i := range from[k] {
			transition := mustParseFATransition(from[k][i])
			dfa.AddTransition(k, transition.input, transition.next)
		}
	}

	dfa.Start = start

	return dfa
}
