package grammar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Production_MarshalUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name  string
		input Production
	}{
		{
			name:  "empty",
			input: Production{},
		},
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "one item",
			input: Production{"item"},
		},
		{
			name:  "two items",
			input: Production{"item1", "item2"},
		},
		{
			name:  "three items",
			input: Production{"item1", "item2", "item3"},
		},
		{
			name:  "epsilon",
			input: Epsilon,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			encoded, err := tc.input.MarshalBinary()
			if !assert.NoError(err, "MarshalBinary failed") {
				return
			}

			actualPtr := &Production{}
			err = actualPtr.UnmarshalBinary(encoded)
			if !assert.NoError(err, "UnmarshalBinary failed") {
				return
			}

			actual := *actualPtr

			assert.Equal(tc.input, actual)
		})
	}
}

func Test_Rule_MarshalUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name  string
		input Rule
	}{
		{
			name:  "empty",
			input: Rule{},
		},
		{
			name: "nil productions",
			input: Rule{
				NonTerminal: "NonTerm",
				Productions: nil,
			},
		},
		{
			name: "empty productions",
			input: Rule{
				NonTerminal: "NonTerm",
				Productions: []Production{},
			},
		},
		{
			name: "one production",
			input: Rule{
				NonTerminal: "TEST",
				Productions: []Production{
					{"item"},
				},
			},
		},
		{
			name: "one production, no non-terminal",
			input: Rule{
				Productions: []Production{
					{"item"},
				},
			},
		},
		{
			name: "two productions",
			input: Rule{
				NonTerminal: "TEST",
				Productions: []Production{
					{"item"},
					{"S", "p", "L"},
				},
			},
		},
		{
			name: "multiple productions",
			input: Rule{
				NonTerminal: "TEST",
				Productions: []Production{
					{"item"},
					{"S", "p", "L"},
					{},
					Epsilon,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			encoded, err := tc.input.MarshalBinary()
			if !assert.NoError(err, "MarshalBinary failed") {
				return
			}

			actualPtr := &Rule{}
			err = actualPtr.UnmarshalBinary(encoded)
			if !assert.NoError(err, "UnmarshalBinary failed") {
				return
			}

			actual := *actualPtr

			assert.Equal(tc.input, actual)
		})
	}
}
