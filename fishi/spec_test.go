package fishi

import (
	"regexp"
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/ictiobus/types"
)

func Test_Spec_ValidateSDTS_disconnectedGraph(t *testing.T) {
	// A -> A + NUM
	// NUM -> id | int

	spec := Spec{
		Tokens: []types.TokenClass{
			lex.NewTokenClass("+", "plus sign"),
			lex.NewTokenClass("var", "variable"),
			lex.NewTokenClass("int", "integer"),
		},
		Patterns: map[string][]Pattern{
			"": []Pattern{
				{Regex: regexp.MustCompile(`\s+`), Action: lex.Discard()},
				{Regex: regexp.MustCompile(`\+`), Action: lex.LexAs("+")},
				{Regex: regexp.MustCompile(`\d+`), Action: lex.LexAs("int")},
				{Regex: regexp.MustCompile(`[A-Za-z_]+`), Action: lex.LexAs("var")},
			},
		},
		Grammar: grammar.MustParse(`
			E -> E + F     ;
			F -> var | int ;
		`),
		TranslationScheme: []SDD{
			{
				Attribute: trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "val"},
				Rule:      grammar.MustParseRule("E -> E + F"),
			},
		},
	}
}
