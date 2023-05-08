package parse

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/stretchr/testify/assert"
)

func Test_ConstructSimpleLRParseTable(t *testing.T) {
	testCases := []struct {
		name      string
		grammar   string
		expect    string
		expectErr bool
	}{
		{
			name: "purple dragon example 4.45",
			grammar: `
				E -> E + T | T ;
				T -> T * F | F ;
				F -> ( E ) | id ;
			`,
			expect: `S   |  A:(  A:)          A:*          A:+          A:ID  A:$          |  G:E  G:F  G:T
--------------------------------------------------------------------------------------
0   |  s1                                          s6                 |  4    7    5  
1   |  s1                                          s6                 |  9    7    5  
2   |       rF -> ( E )  rF -> ( E )  rF -> ( E )        rF -> ( E )  |               
3   |       rT -> T * F  rT -> T * F  rT -> T * F        rT -> T * F  |               
4   |                                 s8                 acc          |               
5   |       rE -> T      s10          rE -> T            rE -> T      |               
6   |       rF -> id     rF -> id     rF -> id           rF -> id     |               
7   |       rT -> F      rT -> F      rT -> F            rT -> F      |               
8   |  s1                                          s6                 |       7    11 
9   |       s2                        s8                              |               
10  |  s1                                          s6                 |       3       
11  |       rE -> E + T  s10          rE -> E + T        rE -> E + T  |               `,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)

			// execute
			actual, _, err := constructSimpleLRParseTable(g, false)

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

func Test_SLR1Parse(t *testing.T) {
	testCases := []struct {
		name      string
		grammar   string
		input     []string
		expect    string
		expectErr bool
	}{
		{
			name: "purple dragon example 4.45",
			grammar: `
				E -> E + T | T ;
				T -> T * F | F ;
				F -> ( E ) | id ;
				`,
			input: []string{"id", "*", "id", "+", "id", "$"},
			expect: `( E )
  |---: ( E )
  |       \---: ( T )
  |               |---: ( T )
  |               |       \---: ( F )
  |               |               \---: (TERM "id")
  |               |---: (TERM "*")
  |               \---: ( F )
  |                       \---: (TERM "id")
  |---: (TERM "+")
  \---: ( T )
          \---: ( F )
                  \---: (TERM "id")`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)
			stream := mockTokens(tc.input...)

			// execute
			parser, _, err := GenerateSimpleLRParser(g, false)
			assert.NoError(err, "generating SLR parser failed")
			actual, err := parser.Parse(stream)

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
