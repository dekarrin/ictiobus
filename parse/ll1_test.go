package parse

import (
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/types"
	"github.com/stretchr/testify/assert"
)

func Test_LL1PredictiveParse(t *testing.T) {
	testCases := []struct {
		name      string
		grammar   string
		input     []string
		expect    string
		expectErr bool
	}{
		{
			name: "aiken expression LL1 sample",
			grammar: `
				S -> T X ;

				T -> ( S )
				   | int Y ;

				X -> + S
				   | ε ;

				Y -> * T
				   | ε ;
			`,
			input: []string{
				"int", "*", "int", types.TokenEndOfText.ID(),
			},
			expect: "( S )\n" +
				`  |---: ( T )` + "\n" +
				`  |       |---: (TERM "int")` + "\n" +
				`  |       \---: ( Y )` + "\n" +
				`  |               |---: (TERM "*")` + "\n" +
				`  |               \---: ( T )` + "\n" +
				`  |                       |---: (TERM "int")` + "\n" +
				`  |                       \---: ( Y )` + "\n" +
				`  |                               \---: (TERM "")` + "\n" +
				`  \---: ( X )` + "\n" +
				`          \---: (TERM "")`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)
			stream := mockTokens(tc.input...)
			ll1, err := GenerateLL1Parser(g)
			if !assert.NoError(err) {
				return
			}

			// execute
			actual, err := ll1.Parse(stream)

			// assert
			if tc.expectErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			assert.Equal(tc.expect, actual.String())
		})
	}
}

func Test_createLL1ParseTable(t *testing.T) {
	testCases := []struct {
		name   string
		g      string
		expect map[string]map[string]grammar.Production
	}{
		{
			name: "aiken example",
			g: `
				S -> T X                        ;
				T -> lparen S rparen | int Y    ;
				X -> p S | ε                    ;
				Y -> m T | ε                    ;
			`,
			expect: map[string]map[string]grammar.Production{
				"S": {"int": grammar.Production{"T", "X"}, "lparen": grammar.Production{"T", "X"}},
				"X": {"p": grammar.Production{"p", "S"}, "rparen": grammar.Epsilon, "$": grammar.Epsilon},
				"T": {"int": grammar.Production{"int", "Y"}, "lparen": grammar.Production{"lparen", "S", "rparen"}},
				"Y": {"m": grammar.Production{"m", "T"}, "p": grammar.Epsilon, "rparen": grammar.Epsilon, "$": grammar.Epsilon},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.g)
			llTab := newLLParseTable()
			for x := range tc.expect {
				for y := range tc.expect[x] {
					llTab.Set(x, y, tc.expect[x][y])
				}
			}

			expect := llTab

			// execute
			actual, err := createLLParseTable(g)

			// assert
			assert.NoError(err)
			if err != nil {
				return
			}

			actualNTs := actual.NonTerminals()
			expectedNTs := expect.NonTerminals()
			if !assert.ElementsMatch(expectedNTs, actualNTs, "non-terminals set not equal") {
				fmt.Printf("Actual produced table:\n" + actual.String())
				return
			}

			actualTerms := actual.Terminals()
			expectedTerms := expect.Terminals()
			if !assert.ElementsMatch(expectedTerms, actualTerms, "terminals set not equal") {
				fmt.Printf("Actual produced table:\n" + actual.String())
				return
			}

			// check each of the entries
			for i := range expectedNTs {
				for j := range expectedTerms {
					A := expectedNTs[i]
					a := expectedTerms[j]
					expectEntry := expect.Get(A, a)
					actualEntry := actual.Get(A, a)
					assert.Equalf(expectEntry, actualEntry, "incorrect entry in M[%q, %q]", A, a)
				}
			}
		})
	}
}

func Test_LL1Table_MarshalUnmarshalBinary(t *testing.T) {
	type entry struct {
		A     string             // non-terminal
		a     string             // terminal
		alpha grammar.Production // production
	}

	withEntries := func(entries ...entry) LL1Table {
		result := newLLParseTable()

		for _, entry := range entries {
			result.Set(entry.A, entry.a, entry.alpha)
		}

		return result
	}

	testCases := []struct {
		name  string
		input LL1Table
	}{
		{
			name:  "empty",
			input: newLLParseTable(),
		},
		{
			name: "one entry",
			input: withEntries(
				entry{"S", "a", grammar.Production{"A", "B"}},
			),
		},
		{
			name: "multiple entries, one is nil",
			input: withEntries(
				entry{"S", "a", grammar.Production{"A", "B"}},
				entry{"A", "d", grammar.Production{"A"}},
				entry{"A", "d", grammar.Production{}},
				entry{"$", "e", nil},
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			encoded, err := tc.input.MarshalBinary()
			if !assert.NoError(err, "MarshalBinary failed") {
				return
			}

			actual := newLLParseTable()

			actualPtr := &actual
			err = actualPtr.UnmarshalBinary(encoded)
			if !assert.NoError(err, "UnmarshalBinary failed") {
				return
			}

			actual = *actualPtr

			assert.Equal(tc.input, actual)
			assert.Equal(tc.input.String(), actual.String())
		})
	}
}
