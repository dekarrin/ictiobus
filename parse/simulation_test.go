package parse

import (
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/tmatch"
	"github.com/stretchr/testify/assert"
)

func Test_Grammar_DeriveFullTree(t *testing.T) {
	testCases := []struct {
		name      string
		input     grammar.CFG
		expect    []Tree
		expectErr bool
	}{
		{
			name: "minimal grammar",
			input: grammar.MustParse(`
				S -> a   ;
			`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "a", Terminal: true},
				}},
			},
		},
		{
			name: "1 rule, multi-production (terms only)",
			input: grammar.MustParse(`
					S -> a | b  ;
				`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "a", Terminal: true},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "b", Terminal: true},
				}},
			},
		},
		{
			name: "minimal 2 rule grammar",
			input: grammar.MustParse(`
					S -> B  ;
					B -> b  ;
				`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "B", Children: []*Tree{
						{Value: "b", Terminal: true},
					}},
				}},
			},
		},
		{
			name: "directly recursive grammar",
			input: grammar.MustParse(`
					S -> S B | B  ;
					B -> b        ;
				`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "S", Children: []*Tree{
						{Value: "B", Children: []*Tree{
							{Value: "b", Terminal: true},
						}},
					}},
					{Value: "B", Children: []*Tree{
						{Value: "b", Terminal: true},
					}},
				}},
			},
		},
		{
			name: "indirectly recursive grammar",
			input: grammar.MustParse(`
					S -> B | a ;
					B -> S b   ;
				`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "B", Children: []*Tree{
						{Value: "S", Children: []*Tree{
							{Value: "a", Terminal: true},
						}},
						{Value: "b", Terminal: true},
					}},
				}},
			},
		},
		{
			name: "lower rule impossible to fill in one try",
			input: grammar.MustParse(`
					S   -> BL        ;
					BL  -> a | b | c ;
				`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "BL", Children: []*Tree{
						{Value: "a", Terminal: true},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "BL", Children: []*Tree{
						{Value: "b", Terminal: true},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "BL", Children: []*Tree{
						{Value: "c", Terminal: true},
					}},
				}},
			},
		},
		{
			name: "lower rule impossible to fill in one try and second try makes third symbol unreachable",
			input: grammar.MustParse(`
					S  -> BL     ;
					BL -> A | b  ;
					A  -> a      ;
				`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "BL", Children: []*Tree{
						{Value: "A", Children: []*Tree{
							{Value: "a", Terminal: true},
						}},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "BL", Children: []*Tree{
						{Value: "b", Terminal: true},
					}},
				}},
			},
		},
		{
			name: "lower rule is unreachable on second try and recurses",
			input: grammar.MustParse(`
				S -> A | B ;
				B -> b     ;
				A -> a | S ;
			`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "A", Children: []*Tree{
						{Value: "a", Terminal: true},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "B", Children: []*Tree{
						{Value: "b", Terminal: true},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "A", Children: []*Tree{
						{Value: "S", Children: []*Tree{
							{Value: "B", Children: []*Tree{
								{Value: "b", Terminal: true},
							}},
						}},
					}},
				}},
			},
		},
		{
			name: "2nd alt is never reached",
			input: grammar.MustParse(`
				S -> A | B ;
				A -> a | b ;
				B -> c | d ;
			`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "A", Children: []*Tree{
						{Value: "a", Terminal: true},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "B", Children: []*Tree{
						{Value: "c", Terminal: true},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "A", Children: []*Tree{
						{Value: "b", Terminal: true},
					}},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "B", Children: []*Tree{
						{Value: "d", Terminal: true},
					}},
				}},
			},
		},
		{
			name: "expr grammar",
			input: grammar.MustParse(`
						E -> E + T | T   ;
						T -> T * F | F   ;
						F -> ( E ) | id  ;
					`),
			expect: []Tree{
				{Value: "E", Children: []*Tree{
					{Value: "E", Children: []*Tree{
						{Value: "T", Children: []*Tree{
							{Value: "T", Children: []*Tree{
								{Value: "F", Children: []*Tree{
									{Value: "(", Terminal: true},
									{Value: "E", Children: []*Tree{
										{Value: "T", Children: []*Tree{
											{Value: "F", Children: []*Tree{
												{Value: "id", Terminal: true},
											}},
										}},
									}},
									{Value: ")", Terminal: true},
								}},
							}},
							{Value: "*", Terminal: true},
							{Value: "F", Children: []*Tree{
								{Value: "id", Terminal: true},
							}},
						}},
					}},
					{Value: "+", Terminal: true},
					{Value: "T", Children: []*Tree{
						{Value: "F", Children: []*Tree{
							{Value: "id", Terminal: true},
						}},
					}},
				}},
			},
		},
		{
			name: "grammar with epsilon",
			input: grammar.MustParse(`
						S -> S a | B   ;
						B -> b | ε     ;
					`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "S", Children: []*Tree{
						{Value: "B", Children: []*Tree{
							{Value: "b", Terminal: true},
						}},
					}},
					{Value: "a", Terminal: true},
				}},
				{Value: "S", Children: []*Tree{
					{Value: "S", Children: []*Tree{
						{Value: "B", Children: []*Tree{
							{Value: "", Terminal: true},
						}},
					}},
					{Value: "a", Terminal: true},
				}},
			},
		},
		{
			name: "a* grammar",
			input: grammar.MustParse(`
						S -> S a | ε   ;
					`),
			expect: []Tree{
				{Value: "S", Children: []*Tree{
					{Value: "S", Children: []*Tree{
						{Value: "", Terminal: true},
					}},
					{Value: "a", Terminal: true},
				}},
			},
		},
		{
			name: "inescapable derivation cycle in single rule",
			input: grammar.MustParse(`
				S -> S a | S b  ;
			`),
			expectErr: true,
		},
		{
			name: "multi-rule inescapable derivation cycle",
			input: grammar.MustParse(`
				S -> A a | b B  ;
				A -> S d		;
				B -> c S		;
			`),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := DeriveFullTree(tc.input)
			if tc.expectErr {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			assert.Len(actual, len(tc.expect))

			limit := len(tc.expect)
			if len(actual) < limit {
				limit = len(actual)
			}
			for i := 0; i < limit; i++ {
				assert.Equal(tc.expect[i].String(), actual[i].String())
			}
		})
	}
}

func Test_Grammar_createFewestNonTermsAlternationsTable(t *testing.T) {
	testCases := []struct {
		name        string
		input       grammar.CFG
		expect      map[string]grammar.Production
		expectOneOf []map[string]grammar.Production // because this is testing a non-deterministic algorithm, there may be multiple possible outputs
		expectErr   bool
	}{
		{
			name: "inescapable derivation cycle in single rule",
			input: grammar.MustParse(`
				S -> S a | S b  ;
			`),
			expectErr: true,
		},
		{
			name: "multi-rule inescapable derivation cycle",
			input: grammar.MustParse(`
				S -> A a | b B  ;
				A -> S d		;
				B -> c S		;
			`),
			expectErr: true,
		},
		{
			name: "single rule",
			input: grammar.MustParse(`
				E -> id ;
			`),
			expect: map[string]grammar.Production{
				"E": {"id"},
			},
		},
		{
			name: "simple expr grammar",
			input: grammar.MustParse(`
				E -> E + T | T ;
				T -> T * F | F ;
				F -> ( E ) | id ;
			`),
			expect: map[string]grammar.Production{
				"E": {"T"}, "T": {"F"}, "F": {"id"},
			},
		},
		{
			name: "same score on rule",
			input: grammar.MustParse(`
				E -> E + T | T ;
				T -> T * F | F ;
				F -> ( E ) | id | num;
			`),
			expectOneOf: []map[string]grammar.Production{
				{"E": {"T"}, "T": {"F"}, "F": {"id"}},
				{"E": {"T"}, "T": {"F"}, "F": {"num"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// make sure we didnt accidentally make an invalid test
			if !tc.expectErr && tc.expect == nil && tc.expectOneOf == nil {
				panic(fmt.Sprintf("test case %s does not specify expectErr, expect, or expectOneOf", tc.name))
			}

			actual, err := createFewestNonTermsAlternationsTable(tc.input)
			if tc.expectErr {
				assert.Error(err)
				return
			} else if !assert.NoError(err) {
				return
			}

			// if only one, check that one
			if tc.expect != nil {
				assert.Equal(tc.expect, actual)
			} else {
				// otherwise, check that it is one of the possible ones
				assertErr := tmatch.AnyStrMapV(actual, tc.expectOneOf, tmatch.Comparer(grammar.Production.Equal))
				assert.NoError(assertErr)
			}
		})
	}

}
