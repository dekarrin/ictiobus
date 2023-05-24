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
			name: "head symbol",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B"},
					{Symbol: "B"},
					{Symbol: "int", Terminal: true},
					{Symbol: "C"},
				},
			},
			rel: NodeRelation{Type: RelHead},
			expect: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B"},
					{Symbol: "B"},
					{Symbol: "int", Terminal: true},
					{Symbol: "C"},
				},
			},
		},
		{
			name: "1st production symbol",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
				},
			},
			rel:    NodeRelation{Type: RelSymbol, Index: 0},
			expect: &AnnotatedParseTree{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
		},
		{
			name: "2nd production symbol",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
				},
			},
			rel:    NodeRelation{Type: RelSymbol, Index: 1},
			expect: &AnnotatedParseTree{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
		},
		{
			name: "3rd production symbol",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
				},
			},
			rel:    NodeRelation{Type: RelSymbol, Index: 2},
			expect: &AnnotatedParseTree{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
		},
		{
			name: "1st non-terminal",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 5}},
				},
			},
			rel:    NodeRelation{Type: RelNonTerminal, Index: 0},
			expect: &AnnotatedParseTree{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
		},
		{
			name: "2nd non-terminal",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 5}},
				},
			},
			rel:    NodeRelation{Type: RelNonTerminal, Index: 1},
			expect: &AnnotatedParseTree{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
		},
		{
			name: "3rd non-terminal",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 5}},
				},
			},
			rel:    NodeRelation{Type: RelNonTerminal, Index: 2},
			expect: &AnnotatedParseTree{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
		},
		{
			name: "1st terminal",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 5}},
				},
			},
			rel:    NodeRelation{Type: RelTerminal, Index: 0},
			expect: &AnnotatedParseTree{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
		},
		{
			name: "2nd terminal",
			apt: &AnnotatedParseTree{Symbol: "A",
				Children: []*AnnotatedParseTree{
					{Symbol: "B", Attributes: nodeAttrs{"$id": 1}},
					{Symbol: "B", Attributes: nodeAttrs{"$id": 2}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 3}},
					{Symbol: "C", Attributes: nodeAttrs{"$id": 4}},
					{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 5}},
				},
			},
			rel:    NodeRelation{Type: RelTerminal, Index: 1},
			expect: &AnnotatedParseTree{Symbol: "int", Terminal: true, Attributes: nodeAttrs{"$id": 5}},
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
