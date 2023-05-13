package automaton

import (
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/stretchr/testify/assert"
)

func Test_FATransition_MarshalUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name  string
		input faTransition
	}{
		{
			name:  "empty",
			input: faTransition{},
		},
		{
			name: "empty input",
			input: faTransition{
				next: "Kanaya Maryam",
			},
		},
		{
			name: "empty next",
			input: faTransition{
				input: "Vriska Serket",
			},
		},
		{
			name: "fully-populated value",
			input: faTransition{
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

			actualPtr := &faTransition{}
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
		input dfaState[dummy]
	}{
		{
			name:  "empty",
			input: dfaState[dummy]{},
		},
		{
			name: "fully-populated",
			input: dfaState[dummy]{
				ordering: 28921,
				name:     "bizarrely long name",
				value: dummy{
					val1: "vriska serket",
					val2: 88888888,
				},
				transitions: map[string]faTransition{
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
				data := rezi.EncString(d.val1)
				data = append(data, rezi.EncInt(d.val2)...)
				return data
			})

			actual, err := unmarshalDFAStateBytes(encoded, func(data []byte) (dummy, error) {
				var d dummy
				var n int
				var err error

				d.val1, n, err = rezi.DecString(data)
				if err != nil {
					return d, fmt.Errorf(".val1: %w", err)
				}
				data = data[n:]

				d.val2, _, err = rezi.DecInt(data)
				if err != nil {
					return d, fmt.Errorf(".val2: %w", err)
				}

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
