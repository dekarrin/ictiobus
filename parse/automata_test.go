package parse

import (
	"testing"

	"github.com/dekarrin/ictiobus/automaton"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/stretchr/testify/assert"
)

func Test_constructDFAForLALR1(t *testing.T) {
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
			actual, err := constructDFAForLALR1(g)
			if !assert.NoError(err) {
				return
			}

			// assert
			assert.Equal(tc.expect, actual.String())
		})
	}

}

func Test_constructDFAForSLR1(t *testing.T) {
	testCases := []struct {
		name        string
		grammar     string
		expect      map[string][]string
		expectStart string
	}{
		{
			name: "aiken example",
			grammar: `
				E -> T + E | T ;
				T -> int * T | int | ( E ) ;
			`,
			expect: map[string][]string{
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
			expectStart: "E-P -> . E",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)
			nfa := buildLR0NFA(tc.expect, tc.expectStart)
			expect := nfa.ToDFA()

			// execute
			actual := constructDFAForSLR1(g)

			// assert
			assert.Equal(expect.String(), actual.String())
		})
	}
}

func buildLR0NFA(from map[string][]string, start string) *automaton.NFA[string] {
	nfa := &automaton.NFA[string]{}

	for k := range from {
		stateItem := grammar.MustParseLR0Item(k)
		nfa.AddState(stateItem.String(), true)
	}

	fromKeys := textfmt.OrderedKeys(from)

	for _, k := range fromKeys {
		fromItem := grammar.MustParseLR0Item(k)
		for i := range from[k] {
			input, next := automaton.MustParseTransition(from[k][i])
			toItem := grammar.MustParseLR0Item(next)
			nfa.AddTransition(fromItem.String(), input, toItem.String())
		}
	}

	nfa.Start = start

	return nfa
}
