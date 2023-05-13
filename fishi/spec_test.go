package fishi

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/stretchr/testify/assert"
)

func Test_Spec_ValidateSDTS_disconnectedGraph(t *testing.T) {
	hooksList := map[string]trans.Hook{
		"add": func(info trans.SetterInfo, args []interface{}) (interface{}, error) {
			left, ok := args[0].(int)
			if !ok {
				return nil, fmt.Errorf("bad type for expr left")
			}
			right, ok := args[1].(int)
			if !ok {
				return nil, fmt.Errorf("bad type for expr right")
			}

			return left + right, nil
		},
		"ident": func(info trans.SetterInfo, args []interface{}) (interface{}, error) { return args[0], nil },
		"var_value": func(info trans.SetterInfo, args []interface{}) (interface{}, error) {
			varName := coerceToString(args[0])
			return len(varName), nil
		},
		"int": func(info trans.SetterInfo, args []interface{}) (interface{}, error) {
			return strconv.Atoi(coerceToString(args[0]))
		},
		"constant-1": func(info trans.SetterInfo, args []interface{}) (interface{}, error) { return 1, nil },
	}

	testCases := []struct {
		name            string
		spec            Spec
		opts            trans.ValidationOptions
		expectErr       bool
		expectWarnTypes []WarnType
	}{
		{
			name: "normal spec",
			spec: Spec{
				Tokens: []lex.TokenClass{
					lex.NewTokenClass("+", "plus sign"),
					lex.NewTokenClass("var", "variable"),
					lex.NewTokenClass("int", "integer"),
				},
				Patterns: map[string][]Pattern{
					"": {
						{Regex: regexp.MustCompile(`\s+`), Action: lex.Discard()},
						{Regex: regexp.MustCompile(`\+`), Action: lex.LexAs("+")},
						{Regex: regexp.MustCompile(`\d+`), Action: lex.LexAs("int")},
						{Regex: regexp.MustCompile(`[A-Za-z_]+`), Action: lex.LexAs("var")},
					},
				},
				Grammar: grammar.MustParse(`
					E -> E + F | F ;
					F -> var | int ;
				`),
				TranslationScheme: []SDD{
					{
						Attribute: trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "val"},
						Rule:      grammar.MustParseRule("E -> E + F"),
						Hook:      "add",
						Args: []trans.AttrRef{
							{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "val"},
							{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 2}, Name: "val"},
						},
					},
					{
						Attribute: trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "val"},
						Rule:      grammar.MustParseRule("E -> F"),
						Hook:      "ident",
						Args: []trans.AttrRef{
							{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "val"},
						},
					},
					{
						Attribute: trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "val"},
						Rule:      grammar.MustParseRule("F -> var"),
						Hook:      "var_value",
						Args: []trans.AttrRef{
							{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
						},
					},
					{
						Attribute: trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "val"},
						Rule:      grammar.MustParseRule("F -> int"),
						Hook:      "int",
						Args: []trans.AttrRef{
							{Relation: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
						},
					},
				},
			},
		},
		{
			name: "SDTS has disconnected depgraph but still allowed",
			spec: Spec{
				Tokens: []lex.TokenClass{
					lex.NewTokenClass("+", "plus sign"),
					lex.NewTokenClass("*", "mult sign"),
					lex.NewTokenClass("var", "variable"),
					lex.NewTokenClass("int", "integer"),
				},
				Patterns: map[string][]Pattern{
					"": {
						{Regex: regexp.MustCompile(`\s+`), Action: lex.Discard()},
						{Regex: regexp.MustCompile(`\+`), Action: lex.LexAs("+")},
						{Regex: regexp.MustCompile(`\*`), Action: lex.LexAs("*")},
						{Regex: regexp.MustCompile(`\d+`), Action: lex.LexAs("int")},
						{Regex: regexp.MustCompile(`[A-Za-z_]+`), Action: lex.LexAs("var")},
					},
				},
				Grammar: grammar.MustParse(`
					E -> E + F | F ;
					F -> var | int ;
				`),
				TranslationScheme: []SDD{
					{
						Attribute: trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "val"},
						Rule:      grammar.MustParseRule("E -> E + F"),
						Hook:      "constant-1",
					},
					{
						Attribute: trans.AttrRef{Relation: trans.NodeRelation{Type: trans.RelHead}, Name: "val"},
						Rule:      grammar.MustParseRule("E -> F"),
						Hook:      "constant-1",
					},
				},
			},
			expectWarnTypes: []WarnType{WarnValidation},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)

			// exec
			warns, err := tc.spec.ValidateSDTS(tc.opts, hooksList)

			actualWarnTypesMap := map[WarnType]struct{}{}
			for _, w := range warns {
				actualWarnTypesMap[w.Type] = struct{}{}
			}
			actualWarnTypes := []WarnType{}
			for wt := range actualWarnTypesMap {
				actualWarnTypes = append(actualWarnTypes, wt)
			}

			// assert
			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
			assert.ElementsMatch(tc.expectWarnTypes, actualWarnTypes)
		})
	}
}

func coerceToString(a interface{}) string {
	// is it just a string? return if so
	if str, ok := a.(string); ok {
		return str
	}

	// otherwise, is it a stringer? call String() and return if so
	if str, ok := a.(fmt.Stringer); ok {
		return str.String()
	}

	// otherwise, is it an error? call Error() and return if so
	if err, ok := a.(error); ok {
		return err.Error()
	}

	// finally, if none of those, get the default formatting and return that
	return fmt.Sprintf("%v", a)
}
