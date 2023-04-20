package fe_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dekarrin/ictiobus/fishi"
	"github.com/dekarrin/ictiobus/fishi/fe"
	"github.com/dekarrin/ictiobus/fishi/format"
	"github.com/dekarrin/ictiobus/fishi/syntax"
	"github.com/dekarrin/ictiobus/internal/textfmt"
)

// Test_Fishi_Spec ensures that the current frontend produces a spec on its own
// source code correctly. It MUST be updated any time changes to fishi.md are
// made. Having this test enshores that an invalid frontend is not inadvertantly
// generated and that fishi.md always has correct syntax for producing the
// current frontend.
func Test_Fishi_Spec(t *testing.T) {
	// expected spec is here
	expected := fishi.Spec{}

	assert := assert.New(t)

	// open fishi.md
	fileR, err := os.Open("../../fishi.md")
	if !assert.NoError(err) {
		return
	}

	// bring in frontend
	frontend := fe.Frontend(syntax.HooksTable, fe.FrontendOptions{})

	// open pre-processing stream
	r, err := format.NewCodeReader(fileR)
	if !assert.NoError(err) {
		return
	}

	// get AST via frontend analysis
	ast, _, err := frontend.Analyze(r)
	if !assert.NoError(err) {
		return
	}

	// ignore warnings as we convert the AST to a spec; we only care about the
	// resulting spec output.
	actual, _, err := fishi.NewSpec(ast)

	// compare tokens
	assert.Len(actual.Tokens, len(expected.Tokens), "incorrect number of tokens produced")
	tokLen := len(actual.Tokens)
	if tokLen > len(expected.Tokens) {
		tokLen = len(expected.Tokens)
	}
	for i := 0; i < tokLen; i++ {
		actualTok := actual.Tokens[i]
		expectedTok := expected.Tokens[i]

		assert.Equal(expectedTok.Human(), actualTok.Human(), "Tokens[%d]: name mismatch", i)
		assert.Equal(expectedTok.ID(), actualTok.ID(), "Tokens[%d]: ID mismatch", i)
	}

	// compare patterns
	actualPatStates := textfmt.OrderedKeys(actual.Patterns)
	expectedPatStates := textfmt.OrderedKeys(expected.Patterns)
	if assert.Equal(expectedPatStates, actualPatStates, "produced lexer states difer") {
		// only check the patterns if there are the same states
		for _, state := range actualPatStates {
			actualPats := actual.Patterns[state]
			expectedPats := expected.Patterns[state]

			assert.Len(actualPats, len(expectedPats), "incorrect number of lexer patterns for state %q", state)
			patLen := len(actualPats)
			if patLen > len(expectedPats) {
				patLen = len(expectedPats)
			}
			for i := 0; i < patLen; i++ {
				actualPat := actualPats[i]
				expectedPat := expectedPats[i]

				assert.Equal(expectedPat.Action, actualPat.Action, "Patterns[%q][%d]: action mismatch", state, i)
				assert.Equal(expectedPat.Priority, actualPat.Priority, "Patterns[%q][%d]: priority mismatch", state, i)
				assert.NotNil(actualPat.Regex, "Patterns[%q][%d]: regex is nil", state, i)
				assert.Equal(expectedPat.Regex.String(), actualPat.Regex.String(), "Patterns[%q][%d]: regex mismatch", state, i)
			}
		}
	}

	// compare grammar rules
	actualNonTerms := actual.Grammar.PriorityNonTerminals()
	expectedNonTerms := expected.Grammar.PriorityNonTerminals()
	if assert.Equal(actualNonTerms, expectedNonTerms, "produced grammar rules differ") {
		for _, nt := range actualNonTerms {
			actualRules := actual.Grammar.Rule(nt)
			expectedRules := expected.Grammar.Rule(nt)

			assert.Equal(expectedRules, actualRules, "Grammar productions for %q mismatch", nt)
		}
	}

	// compare translation schemes
	assert.Len(actual.TranslationScheme, len(expected.TranslationScheme), "incorrect number of SDDs in translation scheme")
	tsLen := len(actual.TranslationScheme)
	if tsLen > len(expected.TranslationScheme) {
		tsLen = len(expected.TranslationScheme)
	}
	for i := 0; i < tsLen; i++ {
		actualSDD := actual.TranslationScheme[i]
		expectedSDD := expected.TranslationScheme[i]

		assert.Equal(expectedSDD.Args, actualSDD.Args, "TranslationScheme[%d]: args mismatch", i)
		assert.Equal(expectedSDD.Attribute, actualSDD.Attribute, "TranslationScheme[%d]: attribute mismatch", i)
		assert.Equal(expectedSDD.Hook, actualSDD.Hook, "TranslationScheme[%d]: hook mismatch", i)
		assert.Equal(expectedSDD.Rule, actualSDD.Rule, "TranslationScheme[%d]: rule mismatch", i)
	}
}
