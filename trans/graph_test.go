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
						mockSBind("A", []string{"B", "B", "int", "C", "int"}, AttrRef{Rel: NRHead(), Name: "test"}),
					},
					expect: []string{`((1: A -> [B B int C int], {head symbol}.test))`},
				},
				{
					name: "1-step dep terminating on built-in via rel-symbol to terminal",
					apt: ATNode(1, "A",
						ATNode(2, "B"),
						ATNode(3, "B",
							ATLeaf(7, "int"),
						),
						ATLeaf(4, "int"),
						ATNode(5, "C"),
						ATLeaf(6, "int"),
					),
					bindings: []sddBinding{
						mockSBind("A", []string{"B", "B", "int", "C", "int"}, AttrRef{Rel: NRHead(), Name: "test"},
							AttrRef{Rel: NRSymbol(2), Name: "$text"},
						),
					},
					expect: []string{`(
			(1: A -> [B B int C int], {head symbol}.test),
			(4: int, {head symbol}.$text => [1])
		)`},
				},
				{
					name: "2-step dep terminating on built-in via rel-symbol to terminal",
					apt: ATNode(1, "A",
						ATNode(2, "B"),
						ATNode(3, "B",
							ATLeaf(7, "int"),
						),
						ATLeaf(4, "int"),
						ATNode(5, "C"),
						ATLeaf(6, "int"),
					),
					bindings: []sddBinding{
						mockSBind("B", []string{"int"}, AttrRef{Rel: NRHead(), Name: "feature"},
							AttrRef{Rel: NRSymbol(0), Name: "$text"},
						),
						mockSBind("A", []string{"B", "B", "int", "C", "int"}, AttrRef{Rel: NRHead(), Name: "test"},
							AttrRef{Rel: NRSymbol(1), Name: "feature"},
						),
					},
					expect: []string{`(
			(1: A -> [B B int C int], {head symbol}.test),
			(3: B -> [int], {head symbol}.feature => [1]),
			(7: int, {head symbol}.$text => [3])
		)`},
				},
				{
					name: "1-step dep terminating on built-in via rel-terminal to terminal",
					apt: ATNode(1, "A",
						ATNode(2, "B"),
						ATNode(3, "B",
							ATLeaf(7, "int"),
						),
						ATLeaf(4, "int"),
						ATNode(5, "C"),
						ATLeaf(6, "int"),
					),
					bindings: []sddBinding{
						mockSBind("A", []string{"B", "B", "int", "C", "int"}, AttrRef{Rel: NRHead(), Name: "test"},
							AttrRef{Rel: NRTerminal(0), Name: "$text"},
						),
					},
					expect: []string{`(
			(1: A -> [B B int C int], {head symbol}.test),
			(4: int, {head symbol}.$text => [1])
		)`},
				},
				{
					name: "1-step dep terminating on non-built-in via rel-nonterminal to nonterminal",
					apt: ATNode(1, "A",
						ATNode(2, "B"),
						ATNode(3, "B",
							ATLeaf(7, "int"),
						),
						ATLeaf(4, "int"),
						ATNode(5, "C"),
						ATLeaf(6, "int"),
					),
					bindings: []sddBinding{
						mockSBind("B", []string{"int"}, AttrRef{Rel: NRHead(), Name: "someVal"}),
						mockSBind("A", []string{"B", "B", "int", "C", "int"}, AttrRef{Rel: NRHead(), Name: "test"},
							AttrRef{Rel: NRNonTerminal(1), Name: "someVal"},
						),
					},
					expect: []string{`(
			(1: A -> [B B int C int], {head symbol}.test),
			(3: B -> [int], {head symbol}.someVal => [1])
		)`},
				},*/
		{
			name: "FISHIMATH (eval version) breaking test case for GHI-128",
			apt: ATNode(1, "FISHIMATH",
				ATNode(2, "STATEMENTS",
					ATNode(3, "STMT",
						ATNode(4, "EXPR",
							ATLeaf(5, "id"),
							ATLeaf(6, "tentacle"),
							ATNode(7, "EXPR",
								ATNode(8, "SUM",
									ATNode(9, "PRODUCT",
										ATNode(10, "TERM",
											ATLeaf(11, "fishtail"),
											ATNode(12, "EXPR",
												ATNode(13, "TERM",
													ATNode(14, "PRODUCT",
														ATNode(15, "TERM",
															ATLeaf(16, "id"),
														),
													),
												),
											),
											ATLeaf(17, "fishhead"),
										),
										ATLeaf(18, "*"),
										ATNode(19, "PRODUCT",
											ATNode(20, "TERM",
												ATLeaf(21, "int"),
											),
											ATLeaf(22, "/"),
											ATNode(23, "PRODUCT",
												ATNode(24, "TERM",
													ATLeaf(25, "float"),
												),
											),
										),
									),
								),
							),
						),
						ATLeaf(26, "shark"),
					),
					ATNode(27, "STATEMENTS",
						ATNode(28, "STMT",
							ATNode(29, "EXPR",
								ATNode(30, "SUM",
									ATNode(31, "PRODUCT",
										ATNode(32, "TERM",
											ATLeaf(33, "id"),
										),
									),
								),
							),
							ATNode(34, "shark"),
						),
					),
				),
			),
			bindings: []sddBinding{
				mockSBind("FISHIMATH", []string{"STATEMENTS"}, AttrRef{Rel: NRHead(), Name: "ir"},
					AttrRef{Rel: NRNonTerminal(0), Name: "result"},
				),

				mockSBind("STATEMENTS", []string{"STMT", "STATEMENTS"}, AttrRef{Rel: NRHead(), Name: "result"},
					AttrRef{Rel: NRSymbol(1), Name: "result"},
					AttrRef{Rel: NRSymbol(0), Name: "value"},
				),
				mockSBind("STATEMENTS", []string{"STMT"}, AttrRef{Rel: NRHead(), Name: "result"},
					AttrRef{Rel: NRSymbol(0), Name: "value"},
				),

				mockSBind("STMT", []string{"EXPR", "shark"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "value"},
				),

				mockSBind("EXPR", []string{"id", "tentacle", "EXPR"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "$text"},
					AttrRef{Rel: NRSymbol(2), Name: "value"},
				),
				mockSBind("EXPR", []string{"SUM"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "value"},
				),

				mockSBind("SUM", []string{"PRODUCT", "+", "EXPR"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRNonTerminal(0), Name: "value"},
					AttrRef{Rel: NRNonTerminal(1), Name: "value"},
				),
				mockSBind("SUM", []string{"PRODUCT", "-", "EXPR"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRNonTerminal(0), Name: "value"},
					AttrRef{Rel: NRNonTerminal(1), Name: "value"},
				),
				mockSBind("SUM", []string{"PRODUCT"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "value"},
				),

				mockSBind("PRODUCT", []string{"TERM", "*", "PRODUCT"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRNonTerminal(0), Name: "value"},
					AttrRef{Rel: NRNonTerminal(1), Name: "value"},
				),
				mockSBind("PRODUCT", []string{"TERM", "/", "PRODUCT"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRNonTerminal(0), Name: "value"},
					AttrRef{Rel: NRNonTerminal(1), Name: "value"},
				),
				mockSBind("PRODUCT", []string{"TERM"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "value"},
				),

				mockSBind("TERM", []string{"fishtail", "EXPR", "fishhead"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(1), Name: "value"},
				),
				mockSBind("TERM", []string{"int"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "$text"},
				),
				mockSBind("TERM", []string{"float"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "$text"},
				),
				mockSBind("TERM", []string{"id"}, AttrRef{Rel: NRHead(), Name: "value"},
					AttrRef{Rel: NRSymbol(0), Name: "$text"},
				),
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

// mockSBind creates a synthesized sddBinding with the given args with Setter
// set to a mock value.
func mockSBind(sym string, prod []string, dest AttrRef, reqs ...AttrRef) sddBinding {
	b := sddBinding{
		Synthesized:         true,
		BoundRuleSymbol:     sym,
		BoundRuleProduction: prod,
		Dest:                dest,
		Requirements:        reqs,
		Setter:              "mock_setter",
	}

	return b
}
