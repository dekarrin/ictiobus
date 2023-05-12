package automaton

import (
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
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

func Test_NewLALR1ViablePrefixDFA(t *testing.T) {
	testCases := []struct {
		name        string
		grammar     string
		expect      string
		expectStart string
	}{
		{
			name: "2-rule ex from https://www.cs.york.ac.uk/fp/lsa/lectures/lalr.pdf",
			grammar: `
				S -> C C ;
				C -> c C | d ;
			`,
			expect: `<START: "{C -> . c C, c, C -> . c C, d, C -> . d, c, C -> . d, d, S -> . C C, $, S-P -> . S, $}", STATES:
	(({C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d} [=(C)=> {C -> c C ., $, C -> c C ., c, C -> c C ., d}, =(c)=> {C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d}, =(d)=> {C -> d ., $, C -> d ., c, C -> d ., d}])),
	(({C -> . c C, $, C -> . d, $, S -> C . C, $} [=(C)=> {S -> C C ., $}, =(c)=> {C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d}, =(d)=> {C -> d ., $, C -> d ., c, C -> d ., d}])),
	(({C -> . c C, c, C -> . c C, d, C -> . d, c, C -> . d, d, S -> . C C, $, S-P -> . S, $} [=(C)=> {C -> . c C, $, C -> . d, $, S -> C . C, $}, =(S)=> {S-P -> S ., $}, =(c)=> {C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d}, =(d)=> {C -> d ., $, C -> d ., c, C -> d ., d}])),
	(({C -> c C ., $, C -> c C ., c, C -> c C ., d} [])),
	(({C -> d ., $, C -> d ., c, C -> d ., d} [])),
	(({S -> C C ., $} [])),
	(({S-P -> S ., $} []))
>`,
		},
		{
			name: "purple dragon 'efficient' LALR construction grammar",
			grammar: `
					S -> L = R | R ;
					L -> * R | id ;
					R -> L ;
			`,
			expect: `<START: "{L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, S -> . L = R, $, S -> . R, $, S-P -> . S, $}", STATES:
	(({L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =} [=(*)=> {L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =}, =(L)=> {R -> L ., $, R -> L ., =}, =(R)=> {L -> * R ., $, L -> * R ., =}, =(id)=> {L -> id ., $, L -> id ., =}])),
	(({L -> * R ., $, L -> * R ., =} [])),
	(({L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, S -> . L = R, $, S -> . R, $, S-P -> . S, $} [=(*)=> {L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =}, =(L)=> {R -> L ., $, S -> L . = R, $}, =(R)=> {S -> R ., $}, =(S)=> {S-P -> S ., $}, =(id)=> {L -> id ., $, L -> id ., =}])),
	(({L -> . * R, $, L -> . id, $, R -> . L, $, S -> L = . R, $} [=(*)=> {L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =}, =(L)=> {R -> L ., $, R -> L ., =}, =(R)=> {S -> L = R ., $}, =(id)=> {L -> id ., $, L -> id ., =}])),
	(({L -> id ., $, L -> id ., =} [])),
	(({R -> L ., $, R -> L ., =} [])),
	(({R -> L ., $, S -> L . = R, $} [=(=)=> {L -> . * R, $, L -> . id, $, R -> . L, $, S -> L = . R, $}])),
	(({S -> L = R ., $} [])),
	(({S -> R ., $} [])),
	(({S-P -> S ., $} []))
>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)

			// execute
			actual, err := NewLALR1ViablePrefixDFA(g)
			if !assert.NoError(err) {
				return
			}

			// assert
			assert.Equal(tc.expect, actual.String())
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
