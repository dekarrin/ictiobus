package parse

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/stretchr/testify/assert"
)

func Test_ConstructCanonicalLR1ParseTable(t *testing.T) {
	testCases := []struct {
		name      string
		grammar   string
		expect    string
		ambig     bool
		expectErr bool
	}{
		{
			name: "purple dragon example 4.45",
			grammar: `
				S -> C C ;
				C -> c C | d ;
			`,
			expect: `S  |  A:C        A:D        A:$        |  G:C  G:S
--------------------------------------------------
0  |  s3         s7                    |  2    9  
1  |  s1         s6                    |  4       
2  |  s1         s6                    |  8       
3  |  s3         s7                    |  5       
4  |                        rC -> c C  |          
5  |  rC -> c C  rC -> c C             |          
6  |                        rC -> d    |          
7  |  rC -> d    rC -> d               |          
8  |                        rS -> C C  |          
9  |                        acc        |          `,
		},
		{

			name: "Repetition via epsilon production",
			grammar: `
				S -> A       ;
				A -> A B | ε ;
				B -> a B | b ;
			`,
			ambig: true,
			expect: `S  |  A:A        A:B        A:$        |  G:A  G:B  G:S
-------------------------------------------------------
0  |  rA -> ε    rA -> ε    rA -> ε    |  1         6  
1  |  s3         s5         rS -> A    |       2       
2  |  rA -> A B  rA -> A B  rA -> A B  |               
3  |  s3         s5                    |       4       
4  |  rB -> a B  rB -> a B  rB -> a B  |               
5  |  rB -> b    rB -> b    rB -> b    |               
6  |                        acc        |               `,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)

			// execute
			actual, _, err := constructCanonicalLR1ParseTable(g, tc.ambig)

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

func Test_CanonicalLR1Parse(t *testing.T) {
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
		{

			name: "Repetition via epsilon production",
			grammar: `
				S -> A       ;
				A -> A B | ε ;
				B -> a B | b ;
			`,
			input: []string{"a", "b", "$"},
			ambig: true,
			expect: `( S )
  \---: ( A )
          |---: ( A )
          |       \---: (TERM "")
          \---: ( B )
                  |---: (TERM "a")
                  \---: ( B )
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
			parser, _, err := GenerateCanonicalLR1Parser(g, tc.ambig)
			if !assert.NoError(err, "generating CLR parser failed") {
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
