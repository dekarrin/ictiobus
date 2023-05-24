package trans

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_APT_RelativeNode(t *testing.T) {
	testCases := []struct {
		name          string
		apt           *AnnotatedParseTree
		rel           NodeRelation
		expect        *AnnotatedParseTree
		expectNoMatch bool
	}{
		{
			name: "no-match: symbol",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
			),
			rel:           NodeRelation{Type: RelSymbol, Index: 10},
			expectNoMatch: true,
		},
		{
			name: "head symbol",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
			),
			rel: NodeRelation{Type: RelHead},
			expect: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
			),
		},
		{
			name: "1st production symbol",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
			),
			rel:    NodeRelation{Type: RelSymbol, Index: 0},
			expect: APTNode(2, "B"),
		},
		{
			name: "2nd production symbol",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
			),
			rel:    NodeRelation{Type: RelSymbol, Index: 1},
			expect: APTNode(3, "B"),
		},
		{
			name: "3rd production symbol",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
			),
			rel:    NodeRelation{Type: RelSymbol, Index: 2},
			expect: APTLeaf(4, "int"),
		},
		{
			name: "1st non-terminal",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
				APTLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelNonTerminal, Index: 0},
			expect: APTNode(2, "B"),
		},
		{
			name: "2nd non-terminal",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
				APTLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelNonTerminal, Index: 1},
			expect: APTNode(3, "B"),
		},
		{
			name: "3rd non-terminal",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
				APTLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelNonTerminal, Index: 2},
			expect: APTNode(5, "C"),
		},
		{
			name: "1st terminal",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
				APTLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelTerminal, Index: 0},
			expect: APTLeaf(4, "int"),
		},
		{
			name: "2nd terminal",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
				APTLeaf(6, "int"),
			),
			rel:    NodeRelation{Type: RelTerminal, Index: 1},
			expect: APTLeaf(6, "int"),
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
