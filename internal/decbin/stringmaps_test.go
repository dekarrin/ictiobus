package decbin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncMapStringToInt(t *testing.T) {
	testCases := []struct {
		name   string
		input  map[string]int
		expect []byte
	}{
		{
			name:   "empty",
			input:  map[string]int{},
			expect: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncMapStringToInt(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}
