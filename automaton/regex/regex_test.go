package regex

import (
	"testing"

	"github.com/dekarrin/ictiobus/automaton"
	"github.com/stretchr/testify/assert"
)

func Test_RegexToNFA(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect automaton.NFA[string]
	}{
		{
			name:   "function does nothing atm",
			input:  "*#@ (CLEARLY INVALID ! [",
			expect: automaton.NFA[string]{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := RegexToNFA(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}
