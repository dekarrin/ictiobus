package fe_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dekarrin/ictiobus/fishi"
	"github.com/dekarrin/ictiobus/fishi/fe"
	"github.com/dekarrin/ictiobus/fishi/format"
	"github.com/dekarrin/ictiobus/fishi/syntax"
)

// Test_Fishi_Spec ensures that the current frontend can produces a spec on its
// own source code correctly.
func Test_Fishi_Spec(t *testing.T) {
	assert := assert.New(t)

	// open fishi.md
	fileR, err := os.Open("../../docs/fishi.md")
	if !assert.NoError(err) {
		return
	}

	// bring in frontend
	frontend := fe.Frontend(syntax.HooksTable, nil)

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

	// ignore warnings as we convert the AST to a spec; just make sure we dont
	// error out
	_, _, err = fishi.NewSpec(ast)

	assert.NoError(err)
}
