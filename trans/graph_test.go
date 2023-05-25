package trans

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_depGraph(t *testing.T) {
	testCases := []struct {
		name     string
		apt      *AnnotatedTree
		bindings []sddBinding
		expect   []string
	}{
		/*{
			name: "no dependencies",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B"),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
				ATLeaf(6, "int"),
			),
			bindings: []sddBinding{
				{
					Synthesized:         true,
					BoundRuleSymbol:     "A",
					BoundRuleProduction: []string{"B", "B", "int", "C", "int"},
					Requirements:        nil,
					Dest:                AttrRef{Rel: NRHead(), Name: "test"},
				},
			},
			expect: []string{`((1: A -> [B B int C int] <[head symbol].test>))`},
		},*/
		{
			name: "1-step dep terminating on built-in via rel-symbol",
			apt: ATNode(1, "A",
				ATNode(2, "B"),
				ATNode(3, "B", ATLeaf(7, "int")),
				ATLeaf(4, "int"),
				ATNode(5, "C"),
				ATLeaf(6, "int"),
			),
			bindings: []sddBinding{
				/*{
					Synthesized:         true,
					BoundRuleSymbol:     "B",
					BoundRuleProduction: []string{"int"},
					Requirements:        []AttrRef{{Rel: NRSymbol(0), Name: "$text"}},
					Dest:                AttrRef{Rel: NRHead(), Name: "feature"},
					Setter:              "constant_builder",
				},*/
				{
					Synthesized:         true,
					BoundRuleSymbol:     "A",
					BoundRuleProduction: []string{"B", "B", "int", "C", "int"},
					Requirements:        []AttrRef{{Rel: NRSymbol(2), Name: "$text"}},
					Dest:                AttrRef{Rel: NRHead(), Name: "test"},
					Setter:              "constant_builder",
				},
			},
			expect: []string{`((1: "A" [B B int C int] <[head symbol].test>))`},
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
			assert.Len(actualStrs, len(tc.expect))
			minLen := len(actualStrs)
			if len(tc.expect) < minLen {
				minLen = len(tc.expect)
			}

			for i := 0; i < minLen; i++ {
				assert.Equal(tc.expect[i], actualStrs[i], "dep graph index %d does not match expected", i)
			}
		})
	}
}
