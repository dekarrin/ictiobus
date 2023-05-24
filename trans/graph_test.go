package trans

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_depGraph(t *testing.T) {
	testCases := []struct {
		name     string
		apt      *AnnotatedParseTree
		bindings []sddBinding
		expect   []string
	}{
		{
			name: "no dependencies",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
				APTLeaf(6, "int"),
			),
			bindings: []sddBinding{
				{
					Synthesized:         true,
					BoundRuleSymbol:     "A",
					BoundRuleProduction: []string{"B", "B", "int", "C", "int"},
					Requirements:        nil,
					Dest:                AttrRef{Relation: NodeRelation{Type: RelHead}, Name: "test"},
					Setter:              "constant_builder",
				},
			},
			expect: []string{`((1: "A" [B B int C int] <head symbol["test"]>))`},
		},
		{
			name: "dep on built-in via rel-symbol",
			apt: APTNode(1, "A",
				APTNode(2, "B"),
				APTNode(3, "B"),
				APTLeaf(4, "int"),
				APTNode(5, "C"),
				APTLeaf(6, "int"),
			),
			bindings: []sddBinding{
				{
					Synthesized:         true,
					BoundRuleSymbol:     "A",
					BoundRuleProduction: []string{"B", "B", "int", "C", "int"},
					Requirements:        []AttrRef{{Relation: NodeRelation{}}},
					Dest:                AttrRef{Relation: NodeRelation{Type: RelHead}, Name: "test"},
					Setter:              "constant_builder",
				},
			},
			expect: []string{`((1: "A" [B B int C int] <head symbol["test"]>))`},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// setup
			sdts := &sdtsImpl{}
			if len(tc.bindings) > 0 {
				sdts.bindings = map[string]map[string][]sddBinding{}
			}
			for _, b := range tc.bindings {
				bindsForHead, ok := sdts.bindings[b.BoundRuleSymbol]
				if !ok {
					bindsForHead = map[string][]sddBinding{}
				}

				prodStr := strings.Join(b.BoundRuleProduction, " ")

				bindsForProd, ok := bindsForHead[prodStr]
				if !ok {
					bindsForProd = []sddBinding{}
				}

				bindsForProd = append(bindsForProd, b)

				bindsForHead[prodStr] = bindsForProd
				sdts.bindings[b.BoundRuleSymbol] = bindsForHead
			}

			// exec
			actuals := depGraph(*tc.apt, sdts)

			actualStrs := make([]string, len(actuals))
			for i, dg := range actuals {
				actualStrs[i] = depGraphString(dg)
			}

			// assert-by-string
			assert.Equal(tc.expect, actualStrs)
		})
	}
}
