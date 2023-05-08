package main

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/stretchr/testify/assert"
)

func Test_sddRefToPrintedString(t *testing.T) {
	testCases := []struct {
		name   string
		ref    trans.AttrRef
		g      grammar.Grammar
		rule   grammar.Rule
		expect string
	}{
		{
			name: "head symbol of rule",
			ref:  trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "Test"},
			g: grammar.MustParse(`
				S -> A | B ;
				A -> a ;
				B -> b | c ;
			`),
			rule:   grammar.MustParseRule("A -> a"),
			expect: "{A$^}.Test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := sddRefToPrintedString(tc.ref, tc.g, tc.rule)

			assert.Equal(tc.expect, actual)
		})
	}
}
