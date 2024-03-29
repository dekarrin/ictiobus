package automaton

import (
	"testing"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/stretchr/testify/assert"
)

func Test_NFA_EpsilonClosure(t *testing.T) {
	testCases := []struct {
		name      string
		nfa       map[string][]string
		nfaStart  string
		nfaAccept []string
		forState  string
		expect    []string
	}{
		{
			name: "aiken example - B",
			nfa: map[string][]string{
				"A": {
					"=(ε)=> H",
					"=(ε)=> B",
				},
				"B": {
					"=(ε)=> C",
					"=(ε)=> D",
				},
				"C": {
					"=(1)=> E",
				},
				"D": {
					"=(0)=> F",
				},
				"E": {
					"=(ε)=> G",
				},
				"F": {
					"=(ε)=> G",
				},
				"G": {
					"=(ε)=> A",
					"=(ε)=> H",
				},
				"H": {
					"=(ε)=> I",
				},
				"I": {
					"=(1)=> J",
				},
				"J": {},
			},
			nfaAccept: []string{"J"},
			nfaStart:  "A",
			forState:  "B",
			expect:    []string{"B", "C", "D"},
		},
		{
			name: "aiken example - G",
			nfa: map[string][]string{
				"A": {
					"=(ε)=> H",
					"=(ε)=> B",
				},
				"B": {
					"=(ε)=> C",
					"=(ε)=> D",
				},
				"C": {
					"=(1)=> E",
				},
				"D": {
					"=(0)=> F",
				},
				"E": {
					"=(ε)=> G",
				},
				"F": {
					"=(ε)=> G",
				},
				"G": {
					"=(ε)=> A",
					"=(ε)=> H",
				},
				"H": {
					"=(ε)=> I",
				},
				"I": {
					"=(1)=> J",
				},
				"J": {},
			},
			nfaAccept: []string{"J"},
			nfaStart:  "A",
			forState:  "G",
			expect:    []string{"A", "B", "C", "D", "G", "H", "I"},
		},
		{
			name: "aiken example, recursive variant - G",
			nfa: map[string][]string{
				"A": {
					"=(ε)=> H",
					"=(ε)=> B",
				},
				"B": {
					"=(ε)=> C",
					"=(ε)=> D",
				},
				"C": {
					"=(ε)=> E",
				},
				"D": {
					"=(0)=> F",
				},
				"E": {
					"=(ε)=> G",
				},
				"F": {
					"=(ε)=> G",
				},
				"G": {
					"=(ε)=> A",
					"=(ε)=> H",
				},
				"H": {
					"=(ε)=> I",
				},
				"I": {
					"=(1)=> J",
				},
				"J": {},
			},
			nfaAccept: []string{"J"},
			nfaStart:  "A",
			forState:  "G",
			expect:    []string{"A", "B", "C", "D", "G", "H", "I", "E"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			nfa := buildNFA(tc.nfa, tc.nfaStart, tc.nfaAccept)
			expectSet := box.StringSetOf(tc.expect)

			// execute
			actual := nfa.epsilonClosure(tc.forState)

			// assert
			assert.True(actual.Equal(expectSet))
		})
	}
}

func Test_NFA_ToDFA(t *testing.T) {
	testCases := []struct {
		name         string
		nfa          map[string][]string
		nfaStart     string
		nfaAccept    []string
		expect       map[string][]string
		expectStart  string
		expectAccept []string
	}{
		{
			name: "aiken example (lexical analysis)",
			nfa: map[string][]string{
				"A": {
					"=(ε)=> H",
					"=(ε)=> B",
				},
				"B": {
					"=(ε)=> C",
					"=(ε)=> D",
				},
				"C": {
					"=(1)=> E",
				},
				"D": {
					"=(0)=> F",
				},
				"E": {
					"=(ε)=> G",
				},
				"F": {
					"=(ε)=> G",
				},
				"G": {
					"=(ε)=> A",
					"=(ε)=> H",
				},
				"H": {
					"=(ε)=> I",
				},
				"I": {
					"=(1)=> J",
				},
				"J": {},
			},
			nfaAccept: []string{"J"},
			nfaStart:  "A",
			expect: map[string][]string{
				"{A, B, C, D, H, I}": {
					"=(0)=> {A, B, C, D, F, G, H, I}",
					"=(1)=> {A, B, C, D, E, G, H, I, J}",
				},
				"{A, B, C, D, F, G, H, I}": {
					"=(0)=> {A, B, C, D, F, G, H, I}",
					"=(1)=> {A, B, C, D, E, G, H, I, J}",
				},
				"{A, B, C, D, E, G, H, I, J}": {
					"=(0)=> {A, B, C, D, F, G, H, I}",
					"=(1)=> {A, B, C, D, E, G, H, I, J}",
				},
			},
			expectStart:  "{A, B, C, D, H, I}",
			expectAccept: []string{"{A, B, C, D, E, G, H, I, J}"},
		},
		{
			name: "aiken example (recognizing viable prefixes)",
			nfa: map[string][]string{
				// first row from vid
				"T -> . ( E )": {
					"=(()=> T -> ( . E )",
				},
				"T -> ( . E )": {
					"=(ε)=> E -> . T",
					"=(ε)=> E -> . T + E",
					"=(E)=> T -> ( E . )",
				},
				"T -> ( E . )": {
					"=())=> T -> ( E ) .",
				},
				"T -> ( E ) .": {},

				// 2nd row from vid
				"E-P -> E .": {},
				"E -> . T + E": {
					"=(ε)=> T -> . ( E )",
					"=(T)=> E -> T . + E",
					"=(ε)=> T -> . int",
					"=(ε)=> T -> . int * T",
				},
				"E -> T . + E": {
					"=(+)=> E -> T + . E",
				},
				"E -> T + . E": {
					"=(ε)=> E -> . T + E",
					"=(E)=> E -> T + E .",
					"=(ε)=> E -> . T",
				},

				// 3rd row from vid
				"E-P -> . E": {
					"=(E)=> E-P -> E .",
					"=(ε)=> E -> . T + E",
					"=(ε)=> E -> . T",
				},
				"T -> . int": {
					"=(int)=> T -> int .",
				},
				"T -> int .":   {},
				"E -> T + E .": {},

				// 4th row from vid
				"E -> . T": {
					"=(ε)=> T -> . int",
					"=(ε)=> T -> . int * T",
					"=(T)=> E -> T .",
					"=(ε)=> T -> . ( E )",
				},
				"T -> int . * T": {
					"=(*)=> T -> int * . T",
				},

				// 5th row from vid
				"E -> T .": {},
				"T -> . int * T": {
					"=(int)=> T -> int . * T",
				},
				"T -> int * . T": {
					"=(ε)=> T -> . int",
					"=(T)=> T -> int * T .",
					"=(ε)=> T -> . ( E )",
					"=(ε)=> T -> . int * T",
				},
				"T -> int * T .": {},
			},
			nfaAccept: []string{
				// (all of them)
				"T -> . ( E )", "T -> ( . E )", "T -> ( E . )", "T -> ( E ) .",
				"E-P -> E .", "E -> . T + E", "E -> T . + E", "E -> T + . E",
				"E-P -> . E", "T -> . int", "T -> int .", "E -> T + E .",
				"E -> . T", "T -> int . * T", "E -> T .", "T -> . int * T",
				"T -> int * . T", "T -> int * T .",
			},
			nfaStart: "E-P -> . E",
			expect: map[string][]string{
				"{E -> . T, E -> . T + E, E -> T + . E, T -> . ( E ), T -> . int, T -> . int * T}": {
					"=(E)=> {E -> T + E .}",
					"=(()=> {E -> . T, E -> . T + E, T -> ( . E ), T -> . ( E ), T -> . int, T -> . int * T}",
					"=(int)=> {T -> int ., T -> int . * T}",
					"=(T)=> {E -> T ., E -> T . + E}",
				},
				"{E -> . T, E -> . T + E, T -> ( . E ), T -> . ( E ), T -> . int, T -> . int * T}": {
					"=(()=> {E -> . T, E -> . T + E, T -> ( . E ), T -> . ( E ), T -> . int, T -> . int * T}",
					"=(E)=> {T -> ( E . )}",
					"=(int)=> {T -> int ., T -> int . * T}",
					"=(T)=> {E -> T ., E -> T . + E}",
				},
				"{E -> . T, E -> . T + E, E-P -> . E, T -> . ( E ), T -> . int, T -> . int * T}": {
					"=(T)=> {E -> T ., E -> T . + E}",
					"=(int)=> {T -> int ., T -> int . * T}",
					"=(()=> {E -> . T, E -> . T + E, T -> ( . E ), T -> . ( E ), T -> . int, T -> . int * T}",
					"=(E)=> {E-P -> E .}",
				},
				"{T -> . ( E ), T -> . int, T -> . int * T, T -> int * . T}": {
					"=(T)=> {T -> int * T .}",
					"=(()=> {E -> . T, E -> . T + E, T -> ( . E ), T -> . ( E ), T -> . int, T -> . int * T}",
					"=(int)=> {T -> int ., T -> int . * T}",
				},
				"{E -> T ., E -> T . + E}": {
					"=(+)=> {E -> . T, E -> . T + E, E -> T + . E, T -> . ( E ), T -> . int, T -> . int * T}",
				},
				"{T -> int ., T -> int . * T}": {
					"=(*)=> {T -> . ( E ), T -> . int, T -> . int * T, T -> int * . T}",
				},
				"{E -> T + E .}":   {},
				"{E-P -> E .}":     {},
				"{T -> int * T .}": {},
				"{T -> ( E . )}": {
					"=())=> {T -> ( E ) .}",
				},
				"{T -> ( E ) .}": {},
			},
			expectStart: "{E -> . T, E -> . T + E, E-P -> . E, T -> . ( E ), T -> . int, T -> . int * T}",
			expectAccept: []string{
				"{E -> . T, E -> . T + E, E -> T + . E, T -> . ( E ), T -> . int, T -> . int * T}",
				"{E -> . T, E -> . T + E, T -> ( . E ), T -> . ( E ), T -> . int, T -> . int * T}",
				"{E -> . T, E -> . T + E, E-P -> . E, T -> . ( E ), T -> . int, T -> . int * T}",
				"{T -> . ( E ), T -> . int, T -> . int * T, T -> int * . T}",
				"{E -> T ., E -> T . + E}",
				"{T -> int ., T -> int . * T}",
				"{E -> T + E .}",
				"{E-P -> E .}",
				"{T -> int * T .}",
				"{T -> ( E . )}",
				"{T -> ( E ) .}",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			nfa := buildNFA(tc.nfa, tc.nfaStart, tc.nfaAccept)
			expect := buildDFA(tc.expect, tc.expectStart, tc.expectAccept)

			// execute
			actual := NFAToDFA(*nfa, func(soFar box.SVSet[string], elem2 string) box.SVSet[string] {
				if soFar == nil {
					soFar = box.NewSVSet[string]()
				}
				soFar.Set(elem2, elem2)
				return soFar
			})

			// assert
			assert.Equal(expect.String(), actual.String())
		})
	}
}

func buildNFA(from map[string][]string, start string, acceptingStates []string) *NFA[string] {
	nfa := &NFA[string]{}

	acceptSet := box.StringSetOf(acceptingStates)

	for k := range from {
		nfa.AddState(k, acceptSet.Has(k))
		nfa.SetValue(k, k)
	}

	// add transitions AFTER all states are already in or it will cause a panic
	for k := range from {
		for i := range from[k] {
			input, next := MustParseTransition(from[k][i])
			nfa.AddTransition(k, input, next)
		}
	}

	nfa.Start = start

	return nfa
}
