package ictiobus

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/stretchr/testify/assert"
)

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
			ctor: NewLALR1Parser,
			g: `
				S -> C C ;
				C -> c C | d ;
			`,
		},
		{
			name: "LL parser",
			ctor: func(g grammar.Grammar, b bool) (parse.Parser, []string, error) {
				p, err := NewLL1Parser(g)
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
