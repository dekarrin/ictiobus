package syntax

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/stretchr/testify/assert"
)

func TestMakeFishispec(t *testing.T) {
	testCases := []struct {
		name      string
		info      trans.SetterInfo
		args      []interface{}
		expect    interface{}
		expectErr bool
	}{
		{
			name: "one grammar block",
			info: trans.SetterInfo{
				GrammarSymbol: "FISHISPEC",
				FirstToken:    lex.NewToken(lex.NewTokenClass("bluh", "A Bluh"), "", 0, 0, "bluh"),
				Name:          "glub",
				Synthetic:     true,
			},
			args: []interface{}{[]Block{GrammarBlock{
				Content: []GrammarContent{
					{
						Rules: []GrammarRule{
							{
								Rule: grammar.Rule{
									NonTerminal: "A",
									Productions: []grammar.Production{{"B"}, {"B", "C"}},
								},
							},
						},
					},
				},
			}},
			},
			expect: AST{Nodes: []Block{GrammarBlock{
				Content: []GrammarContent{
					{
						Rules: []GrammarRule{
							{
								Rule: grammar.Rule{
									NonTerminal: "A",
									Productions: []grammar.Production{{"B"}, {"B", "C"}},
								},
							},
						},
					},
				},
			}}},
		},
		{
			name: "error",
			info: trans.SetterInfo{
				GrammarSymbol: "FISHISPEC",
				FirstToken:    lex.NewToken(lex.NewTokenClass("bluh", "A Bluh"), "", 0, 0, "bluh"),
				Name:          "glub",
				Synthetic:     true,
			},
			args:      []interface{}{3},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := sdtsFnMakeFishispec(tc.info, tc.args)

			if tc.expectErr {
				assert.Error(err)
				return
			}

			if !assert.NoError(err) {
				return
			}
			assert.Equal(tc.expect, actual)
		})
	}
}
