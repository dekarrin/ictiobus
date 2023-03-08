package grammar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LR0Item_MarshalUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name  string
		input LR0Item
	}{
		{
			name:  "empty",
			input: LR0Item{},
		},
		{
			name: "nil left, right",
			input: LR0Item{
				NonTerminal: "A",
			},
		},
		{
			name: "empty left, right",
			input: LR0Item{
				NonTerminal: "A",
				Left:        []string{},
				Right:       []string{},
			},
		},
		{
			name: "empty non-terminal",
			input: LR0Item{
				Left:  []string{"item"},
				Right: []string{},
			},
		},
		{
			name: "fully-populated value",
			input: LR0Item{
				NonTerminal: "Hello",
				Left:        []string{"item"},
				Right:       []string{"", "A", "B"},
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

			actualPtr := &LR0Item{}
			err = actualPtr.UnmarshalBinary(encoded)
			if !assert.NoError(err, "UnmarshalBinary failed") {
				return
			}

			actual := *actualPtr

			assert.Equal(tc.input, actual)
		})
	}
}

func Test_LR1Item_MarshalUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name  string
		input LR1Item
	}{
		{
			name:  "empty",
			input: LR1Item{},
		},
		{
			name: "empty LR0Item",
			input: LR1Item{
				Lookahead: "gamma",
			},
		},
		{
			name: "fully-populated value",
			input: LR1Item{
				LR0Item: LR0Item{
					NonTerminal: "TEST",
					Left:        []string{},
					Right:       []string{"A", "B", "C"},
				},
				Lookahead: "gamma",
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

			actualPtr := &LR1Item{}
			err = actualPtr.UnmarshalBinary(encoded)
			if !assert.NoError(err, "UnmarshalBinary failed") {
				return
			}

			actual := *actualPtr

			assert.Equal(tc.input, actual)
		})
	}
}
