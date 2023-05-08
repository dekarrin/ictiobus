package tmatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AnyStrMap(t *testing.T) {
	testCases := []struct {
		name          string
		inputActual   map[string]int
		inputExpected []map[string]int
		expectErr     bool
	}{
		{
			name:        "matches",
			inputActual: map[string]int{"T": 1, "S": 2},
			inputExpected: []map[string]int{
				{"T": 2, "S": 2},
				{"T": 1, "S": 2},
				{"T": 2, "S": 1},
			},
			expectErr: false,
		},
		{
			name:        "not matches",
			inputActual: map[string]int{"T": 1, "S": 2},
			inputExpected: []map[string]int{
				{"T": 2, "S": 2},
				{"T": 2, "S": 1},
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			err := AnyStrMap(tc.inputActual, tc.inputExpected)

			if tc.expectErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
