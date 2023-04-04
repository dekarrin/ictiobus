package fe

import (
	"bytes"
	"testing"

	"github.com/dekarrin/ictiobus/types"
	"github.com/stretchr/testify/assert"
)

func Test_Fishi_Lexer_AttrRef_Terminal(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect []types.Token
	}{}

	lx := CreateBootstrapLexer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			r := bytes.NewReader([]byte(tc.input))

			// execute
			tokStream, err := lx.Lex(r)

			// verify
			if !assert.NoError(err) {
				return
			}

			// collect the tokens
			toks := gatherTokens(tokStream)

			// validate them
			for i, tok := range toks {

			}
		})
	}
}

func gatherTokens(stream types.TokenStream) []types.Token {
	allTokens := []types.Token{}

	for stream.HasNext() {
		allTokens = append(allTokens, stream.Next())
	}

	return allTokens
}
