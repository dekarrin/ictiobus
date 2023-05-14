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
					return nil, nil
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

func Test_EncodeDecodeParserBytes(t *testing.T) {
	testCases := []struct {
		name  string
		ctor  func(grammar.Grammar, bool) (parse.Parser, []string, error)
		g     string
		ambig bool
	}{
		{
			name: "CLR parser",
			ctor: NewCLRParser,
			g: `
				S -> C C ;
				C -> c C | d ;
			`,
		},
		{
			name: "SLR parser",
			ctor: NewSLRParser,
			g: `
				S -> C C ;
				C -> c C | d ;
			`,
		},
		{
			name: "LALR parser",
			ctor: NewLALRParser,
			g: `
				S -> C C ;
				C -> c C | d ;
			`,
		},
		{
			name: "LL parser",
			ctor: func(g grammar.Grammar, b bool) (parse.Parser, []string, error) {
				p, err := NewLLParser(g)
				return p, nil, err
			},
			g: `
				S -> C C ;
				C -> c C | d ;
			`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			g, err := grammar.Parse(tc.g)
			if err != nil {
				t.Fatalf("Error parsing grammar: %v", err)
			}

			p, _, err := tc.ctor(g, tc.ambig)
			if err != nil {
				t.Fatalf("Error creating parser: %v", err)
			}

			b := EncodeParserBytes(p)
			p2, err := DecodeParserBytes(b)
			if err != nil {
				t.Fatalf("Error decoding parser: %v", err)
			}

			assert.Equal(p.Type(), p2.Type(), "type of decoded parser does not match original parser")
			assert.Equal(p.TableString(), p2.TableString(), "parsing table of decoded parser does not match original parser")
			assert.Equal(p.DFAString(), p2.DFAString(), "DFA of decoded parser does not match original parser")
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
func (mp mockParser) Type() parse.Algorithm                { return parse.AlgoLL1 }
func (mp mockParser) TableString() string                  { return "" }
func (mp mockParser) RegisterTraceListener(func(s string)) {}
func (mp mockParser) DFAString() string                    { return "" }
func (mp mockParser) Grammar() grammar.Grammar             { return grammar.Grammar{} }
func (mp mockParser) UnmarshalBinary(b []byte) error       { return nil }

type mockSDTS struct {
	fn func(parse.Tree, ...string) ([]interface{}, []error, error)
}

func (ms mockSDTS) Evaluate(p parse.Tree, attrs ...string) ([]interface{}, []error, error) {
	return ms.fn(p, attrs...)
}
func (mp mockSDTS) SetHooks(hooks trans.HookMap) {}
func (mp mockSDTS) BindInheritedAttribute(head string, prod []string, attrName string, hook string, withArgs []trans.AttrRef, forProd trans.NodeRelation) error {
	return nil
}
func (mp mockSDTS) BindSynthesizedAttribute(head string, prod []string, attrName string, hook string, withArgs []trans.AttrRef) error {
	return nil
}
func (mp mockSDTS) SetNoFlow(synth bool, head string, prod []string, attrName string, forProd trans.NodeRelation, which int, ifParent string) error {
	return nil
}
func (mp mockSDTS) Validate(grammar grammar.Grammar, attribute string, debug trans.ValidationOptions, fakeValProducer ...map[string]func() string) (warns []string, err error) {
	return nil, nil
}
