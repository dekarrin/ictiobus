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
		expect   []*directedGraph[depNode]
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
			actual := depGraph(*tc.apt, sdts)

			// assert
			assert.Equal(tc.expect, actual)
		})
	}
}
