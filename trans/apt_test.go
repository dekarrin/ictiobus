package trans

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_APT_RelativeNode(t *testing.T) {
	testCases := []struct {
		name          string
		apt           *AnnotatedTree
		rel           NodeRelation
		expect        *AnnotatedTree
		expectNoMatch bool
	}{
		{
			name: "no-match: symbol",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
			),
			rel:           NodeRelation{Type: RelSymbol, Index: 10},
			expectNoMatch: true,
		},
		{
			name: "head symbol",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
			),
			rel: NodeRelation{Type: RelHead},
			expect: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
			),
		},
		{
			name: "1st production symbol",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
			),
			rel:    NodeRelation{Type: RelSymbol, Index: 0},
			expect: ATNode(2, "B"),
		},
		{
			name: "2nd production symbol",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
			),
			rel:    NodeRelation{Type: RelSymbol, Index: 1},
			expect: ATNode(3, "B"),
		},
		{
			name: "3rd production symbol",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
			),
			rel:    NodeRelation{Type: RelSymbol, Index: 2},
			expect: ATLeaf(4, "int"),
		},
		{
			name: "1st non-terminal",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
				ATLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelNonTerminal, Index: 0},
			expect: ATNode(2, "B"),
		},
		{
			name: "2nd non-terminal",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
				ATLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelNonTerminal, Index: 1},
			expect: ATNode(3, "B"),
		},
		{
			name: "3rd non-terminal",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
				ATLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelNonTerminal, Index: 2},
			expect: ATNode(5, "C"),
		},
		{
			name: "1st terminal",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
				ATLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelTerminal, Index: 0},
			expect: ATLeaf(4, "int"),
		},
		{
			name: "2nd terminal",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
				ATLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelTerminal, Index: 1},
			expect: ATLeaf(6, "int"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, ok := tc.apt.RelativeNode(tc.rel)

			if !assert.Equal(!tc.expectNoMatch, ok) {
				return
			}
			if tc.expectNoMatch {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}
}
