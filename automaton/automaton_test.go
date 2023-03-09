package automaton

import (
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/internal/decbin"
	"github.com/stretchr/testify/assert"
)

func Test_FATransition_MarshalUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name  string
		input FATransition
	}{
		{
			name:  "empty",
			input: FATransition{},
		},
		{
			name: "empty input",
			input: FATransition{
				next: "Kanaya Maryam",
			},
		},
		{
			name: "empty next",
			input: FATransition{
				input: "Vriska Serket",
			},
		},
		{
			name: "fully-populated value",
			input: FATransition{
				input: "a",
				next:  "b",
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

			actualPtr := &FATransition{}
			err = actualPtr.UnmarshalBinary(encoded)
			if !assert.NoError(err, "UnmarshalBinary failed") {
				return
			}

			actual := *actualPtr

			assert.Equal(tc.input, actual)
		})
	}
}

func Test_DFAState_MarshalUnmarshalBinary(t *testing.T) {
	type dummy struct {
		val1 string
		val2 int
	}

	testCases := []struct {
		name  string
		input DFAState[dummy]
	}{
		{
			name:  "empty",
			input: DFAState[dummy]{},
		},
		{
			name: "fully-populated",
			input: DFAState[dummy]{
				ordering: 28921,
				name:     "bizarrely long name",
				value: dummy{
					val1: "vriska serket",
					val2: 88888888,
				},
				transitions: map[string]FATransition{
					"a": {input: "a", next: "b"},
				},
				accepting: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			encoded := tc.input.MarshalBytes(func(d dummy) []byte {
				data := decbin.EncString(d.val1)
				data = append(data, decbin.EncInt(d.val2)...)
				return data
			})

			actual, err := UnmarshalDFAStateBytes(encoded, func(data []byte) (dummy, error) {
				var d dummy
				var n int
				var err error

				d.val1, n, err = decbin.DecString(data)
				if err != nil {
					return d, fmt.Errorf(".val1: %w", err)
				}
				data = data[n:]

				d.val2, n, err = decbin.DecInt(data)
				if err != nil {
					return d, fmt.Errorf(".val2: %w", err)
				}
				data = data[n:]

				return d, nil
			})
			if !assert.NoError(err, "UnmarshalDFAStateBytes failed") {
				return
			}

			// glub glub 38D v happy fishy
			assert.Equal(tc.input, actual)
		})
	}
}
