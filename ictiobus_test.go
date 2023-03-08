package ictiobus

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/stretchr/testify/assert"
)

func Test_EncodeDecodeParserBytes(t *testing.T) {
	testCases := []struct {
		name  string
		ctor  func(grammar.Grammar, bool) (Parser, []string, error)
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
			assert.Equal(p.GetDFA(), p2.GetDFA(), "DFA of decoded parser does not match original parser")
		})
	}
}
