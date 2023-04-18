package fe

import (
	"bytes"
	"testing"

	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/types"
	"github.com/stretchr/testify/assert"

	. "github.com/dekarrin/ictiobus/fishi/fe/fetoken"
)

func Test_Fishi_Lexer_AttrRef_Terminal(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect []types.Token
	}{
		{
			name: "single attr ref",
			input: `%%actions
					someAttrRef.value`,
			expect: []types.Token{
				lex.NewToken(TCHeaderActions, "%%actions", 0, 0, ""),
				lex.NewToken(TCAttrRef, "someAttrRef.value", 0, 0, ""),
				lex.NewToken(types.TokenEndOfText, "", 0, 0, ""),
			},
		},
	}

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
			tokCount := len(toks)

			// only check count, token class, and lexeme.
			if !assert.Len(toks, len(tc.expect), "different number of tokens") {
				if tokCount < len(tc.expect) {
					tokCount = len(tc.expect)
				}
			}

			for i := 0; i < tokCount; i++ {
				if !assert.Equal(tc.expect[i].Class().ID(), toks[i].Class().ID(), "different token class for token #%d") {
					return
				}
				assert.Equal(tc.expect[i].Lexeme(), toks[i].Lexeme(), "different lexemes for token #%d")
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
