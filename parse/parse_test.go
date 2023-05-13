package parse

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/stretchr/testify/assert"
)

func Test_IsLL1(t *testing.T) {
	testCases := []struct {
		name   string
		g      string
		expect bool
	}{
		{
			name:   "empty grammar",
			expect: true,
		},
		{
			name: "example 1 - S",
			g: `
				S -> T A          ;
				A -> plus T A | ε ;
				T -> F B          ;
				B -> mult F B | ε ;
				F -> lp S rp | id ;
			`,
			expect: true,
		},
		{
			name: "same string in two prods",
			g: `
				S -> a | a b ;
			`,
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)

			g := grammar.MustParse(tc.g)

			// execute
			actual := IsLL1(g)

			// assert
			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_findFOLLOWSet(t *testing.T) {
	const (
		example1Grammar = `	
			S -> a B D h ;
			B -> c C     ;
			C -> b C | ε ;
			D -> E F     ;
			E -> g | ε   ;
			F -> f | ε   ;
		`
		aikenOperationsGrammar = `
			S -> T X                     ;
			T -> lparen S rparen | int Y ;
			X -> plus S | ε              ;
			Y -> times T | ε             ;
		`
	)

	testCases := []struct {
		name   string
		g      string
		follow string
		expect []string
	}{
		{
			name: "empty grammar",
		},
		{
			name:   "example 1 - S",
			g:      example1Grammar,
			follow: "S",
			expect: []string{"$"},
		},
		{
			name:   "example 1 - B",
			g:      example1Grammar,
			follow: "B",
			expect: []string{"g", "f", "h"},
		},
		{
			name:   "example 1 - C",
			g:      example1Grammar,
			follow: "C",
			expect: []string{"g", "f", "h"},
		},
		{
			name:   "example 1 - D",
			g:      example1Grammar,
			follow: "D",
			expect: []string{"h"},
		},
		{
			name:   "example 1 - E",
			g:      example1Grammar,
			follow: "E",
			expect: []string{"f", "h"},
		},
		{
			name:   "example 1 - F",
			g:      example1Grammar,
			follow: "F",
			expect: []string{"h"},
		},
		{
			name:   "example 1 - a",
			g:      example1Grammar,
			follow: "a",
			expect: []string{"c"},
		},
		{
			name:   "example 1 - h",
			g:      example1Grammar,
			follow: "h",
			expect: []string{"$"},
		},
		{
			name:   "example 1 - c",
			g:      example1Grammar,
			follow: "c",
			expect: []string{"b", "g", "f", "h"},
		},
		{
			name:   "example 1 - b",
			g:      example1Grammar,
			follow: "b",
			expect: []string{"b", "g", "f", "h"},
		},
		{
			name:   "example 1 - g",
			g:      example1Grammar,
			follow: "g",
			expect: []string{"f", "h"},
		},
		{
			name:   "example 1 - f",
			g:      example1Grammar,
			follow: "f",
			expect: []string{"h"},
		},
		{
			name:   "aiken operations - S",
			g:      aikenOperationsGrammar,
			follow: "S",
			expect: []string{"$", "rparen"},
		},
		{
			name:   "aiken operations - X",
			g:      aikenOperationsGrammar,
			follow: "X",
			expect: []string{"$", "rparen"},
		},
		{
			name:   "aiken operations - T",
			g:      aikenOperationsGrammar,
			follow: "T",
			expect: []string{"plus", "$", "rparen"},
		},
		{
			name:   "aiken operations - Y",
			g:      aikenOperationsGrammar,
			follow: "Y",
			expect: []string{"plus", "$", "rparen"},
		},
		{
			name:   "aiken operations - (",
			g:      aikenOperationsGrammar,
			follow: "lparen",
			expect: []string{"lparen", "int"},
		},
		{
			name:   "aiken operations - )",
			g:      aikenOperationsGrammar,
			follow: "rparen",
			expect: []string{"rparen", "plus", "$"},
		},
		{
			name:   "aiken operations - +",
			g:      aikenOperationsGrammar,
			follow: "plus",
			expect: []string{"lparen", "int"},
		},
		{
			name:   "aiken operations - *",
			g:      aikenOperationsGrammar,
			follow: "times",
			expect: []string{"lparen", "int"},
		},
		{
			name:   "aiken operations - int",
			g:      aikenOperationsGrammar,
			follow: "int",
			expect: []string{"times", "plus", "$", "rparen"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			expectMap := map[string]bool{}
			for i := range tc.expect {
				expectMap[tc.expect[i]] = true
			}

			g := grammar.MustParse(tc.g)

			// execute
			actual := findFOLLOWSet(g, tc.follow)

			// assert
			assert.Equal(textfmt.OrderedKeys(expectMap), box.Alphabetized[string](actual))
		})
	}
}
