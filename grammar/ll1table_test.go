package grammar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LL1Table_MarshalUnmarshalBinary(t *testing.T) {
	type entry struct {
		A     string     // non-terminal
		a     string     // terminal
		alpha Production // production
	}

	withEntries := func(entries ...entry) LL1Table {
		result := NewLL1Table()

		for _, entry := range entries {
			result.Set(entry.A, entry.a, entry.alpha)
		}

		return result
	}

	testCases := []struct {
		name  string
		input LL1Table
	}{
		{
			name:  "empty",
			input: NewLL1Table(),
		},
		{
			name: "one entry",
			input: withEntries(
				entry{"S", "a", Production{"A", "B"}},
			),
		},
		{
			name: "multiple entries, one is nil",
			input: withEntries(
				entry{"S", "a", Production{"A", "B"}},
				entry{"A", "d", Production{"A"}},
				entry{"A", "d", Production{}},
				entry{"$", "e", nil},
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			encoded, err := tc.input.MarshalBinary()
			if !assert.NoError(err, "MarshalBinary failed") {
				return
			}

			actual := NewLL1Table()

			actualPtr := &actual
			err = actualPtr.UnmarshalBinary(encoded)
			if !assert.NoError(err, "UnmarshalBinary failed") {
				return
			}

			actual = *actualPtr

			assert.Equal(tc.input, actual)
			assert.Equal(tc.input.String(), actual.String())
		})
	}
}
