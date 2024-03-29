package parse

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/lex"

	"github.com/stretchr/testify/assert"
)

func Test_ConstructLALR1ParseTable(t *testing.T) {
	testCases := []struct {
		name      string
		grammar   string
		expect    string
		ambig     bool
		expectErr bool
	}{
		{
			name: "purple dragon LALR(1) example grammar 4.55",
			grammar: `
				S -> C C ;
				C -> c C | d ;
			`,
			expect: `S  |  A:C        A:D        A:$        |  G:C  G:S
--------------------------------------------------
0  |  s1         s4                    |  2    6  
1  |  s1         s4                    |  3       
2  |  s1         s4                    |  5       
3  |  rC -> c C  rC -> c C  rC -> c C  |          
4  |  rC -> d    rC -> d    rC -> d    |          
5  |                        rS -> C C  |          
6  |                        acc        |          `,
		},
		{
			name: "Repetition via epsilon production",
			grammar: `
				S -> A       ;
				A -> B b     ;
				B -> B a     ;
				B -> ε       ;
			`,
			ambig: true,
			expect: `S  |  A:A        A:B        A:$        |  G:A  G:B  G:S
-------------------------------------------------------
0  |  rB -> ε    rB -> ε               |  4    1    5  
1  |  s3         s2                    |               
2  |                        rA -> B b  |               
3  |  rB -> B a  rB -> B a             |               
4  |                        rS -> A    |               
5  |                        acc        |               `,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)

			// execute
			actual, _, err := constructLALR1ParseTable(g, tc.ambig)

			// assert
			if tc.expectErr {
				assert.Error(err)
				return
			}
			if !assert.NoError(err) {
				return
			}
			assert.Equal(tc.expect, actual.String())
		})
	}

}

func Test_LALR1Parse(t *testing.T) {
	testCases := []struct {
		name      string
		grammar   string
		input     []string
		expect    string
		ambig     bool
		expectErr bool
	}{
		{
			name: "purple dragon example 4.45",
			grammar: `
				E -> E + T | T ;
				T -> T * F | F ;
				F -> ( E ) | id ;
			`,
			input: []string{"(", "id", "+", "id", ")", "*", "id", lex.TokenEndOfText.ID()},
			expect: `( E )
  \---: ( T )
          |---: ( T )
          |       \---: ( F )
          |               |---: (TERM "(")
          |               |---: ( E )
          |               |       |---: ( E )
          |               |       |       \---: ( T )
          |               |       |               \---: ( F )
          |               |       |                       \---: (TERM "id")
          |               |       |---: (TERM "+")
          |               |       \---: ( T )
          |               |               \---: ( F )
          |               |                       \---: (TERM "id")
          |               \---: (TERM ")")
          |---: (TERM "*")
          \---: ( F )
                  \---: (TERM "id")`,
		},
		{
			name: "Repetition via epsilon production",
			grammar: `
				S -> A       ;
				A -> B b     ;
				B -> B a     ;
				B -> ε       ;
			`,
			input: []string{"a", "b", "$"},
			ambig: true,
			expect: `( S )
  \---: ( A )
          |---: ( B )
          |       |---: ( B )
          |       |       \---: (TERM "")
          |       \---: (TERM "a")
          \---: (TERM "b")`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)
			stream := mockTokens(tc.input...)

			// execute
			parser, _, err := GenerateLALR1Parser(g, tc.ambig)
			if !assert.NoError(err, "generating LALR parser failed") {
				return
			}
			actual, err := parser.Parse(stream)

			// assert
			if tc.expectErr {
				assert.Error(err)
				return
			}
			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual.String())

		})
	}
}
