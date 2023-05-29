package ictiobus

import (
	"io"
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/stretchr/testify/assert"
)

// mockTokenStream returns a token stream that is used for testing.
type mockTokenStream struct{}

func (ets mockTokenStream) Next() lex.Token {
	return lex.NewToken(lex.MakeDefaultClass("A"), "", 0, 1, "")
}

func (ets mockTokenStream) HasNext() bool {
	return true
}

func (ets mockTokenStream) Peek() lex.Token {
	return lex.NewToken(lex.MakeDefaultClass("A"), "", 0, 1, "")
}

func Test_Frontend_AnalyzeString(t *testing.T) {
	testCases := []struct {
		name      string
		fe        Frontend[int]
		input     string
		expect    int
		expectErr bool
	}{
		{
			name: "SDTS output taken for nonsense input",
			fe: Frontend[int]{
				Lexer: mockLexer{fn: func(r io.Reader) (lex.TokenStream, error) {
					return mockTokenStream{}, nil
				}},
				Parser: mockParser{fn: func(ts lex.TokenStream) (parse.Tree, error) {
					return parse.Tree{}, nil
				}},
				SDTS: mockSDTS{fn: func(t parse.Tree, s ...string) ([]interface{}, []error, error) {
					return []interface{}{8}, nil, nil
				}},
			},
			input:  "alkfjdalksfjd",
			expect: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, _, err := tc.fe.AnalyzeString(tc.input)

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

// mock frontend components below here

type mockLexer struct {
	fn func(io.Reader) (lex.TokenStream, error)
}

func (ml mockLexer) Lex(r io.Reader) (lex.TokenStream, error) {
	return ml.fn(r)
}
func (ml mockLexer) RegisterClass(cl lex.TokenClass, forState string) {}
func (ml mockLexer) AddPattern(pat string, action lex.Action, forState string, priority int) error {
	return nil
}
func (ml mockLexer) FakeLexemeProducer(combine bool, state string) map[string]func() string {
	return nil
}
func (ml mockLexer) SetStartingState(s string)               {}
func (ml mockLexer) RegisterTokenListener(func(t lex.Token)) {}
func (ml mockLexer) StartingState() string                   { return "" }

type mockParser struct {
	fn func(lex.TokenStream) (parse.Tree, error)
}

func (mp mockParser) Parse(s lex.TokenStream) (parse.Tree, error) {
	return mp.fn(s)
}
func (mp mockParser) MarshalBinary() ([]byte, error)       { return nil, nil }
func (mp mockParser) Type() parse.Algorithm                { return parse.LL1 }
func (mp mockParser) TableString() string                  { return "" }
func (mp mockParser) RegisterTraceListener(func(s string)) {}
func (mp mockParser) DFAString() string                    { return "" }
func (mp mockParser) Grammar() grammar.CFG                 { return grammar.CFG{} }
func (mp mockParser) UnmarshalBinary(b []byte) error       { return nil }

type mockSDTS struct {
	fn func(parse.Tree, ...string) ([]interface{}, []error, error)
}

func (ms mockSDTS) Evaluate(p parse.Tree, attrs ...string) ([]interface{}, []error, error) {
	return ms.fn(p, attrs...)
}
func (mp mockSDTS) SetHooks(hooks trans.HookMap) {}
func (mp mockSDTS) BindI(head string, prod []string, attrName string, hook string, withArgs []trans.AttrRef, forProd trans.NodeRelation) error {
	return nil
}
func (mp mockSDTS) Bind(head string, prod []string, attrName string, hook string, withArgs []trans.AttrRef) error {
	return nil
}
func (mp mockSDTS) String() string {
	return "mockSDTS<>"
}
