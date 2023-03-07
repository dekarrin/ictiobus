package decbin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncSliceString(t *testing.T) {
	testCases := []struct {
		name   string
		input  []string
		expect []byte
	}{
		{
			name:   "empty",
			input:  []string{},
			expect: []byte{0x00},
		},
		{
			name:   "one",
			input:  []string{"one"},
			expect: []byte{0x01, 0x03, 0x6f, 0x6e, 0x65},
		},
		{
			name:   "two",
			input:  []string{"one", "two"},
			expect: []byte{0x02, 0x03, 0x6f, 0x6e, 0x65, 0x03, 0x74, 0x77, 0x6f},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncSliceString(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecSliceString(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue []string
		expectRead  int
		expectError bool
	}{
		{
			name:        "empty",
			input:       []byte{0x00},
			expectValue: []string{},
			expectRead:  1,
		},
		{
			name:        "one",
			input:       []byte{0x01, 0x03, 0x6f, 0x6e, 0x65},
			expectValue: []string{"one"},
			expectRead:  5,
		},
		{
			name:        "two",
			input:       []byte{0x02, 0x03, 0x6f, 0x6e, 0x65, 0x03, 0x74, 0x77, 0x6f},
			expectValue: []string{"one", "two"},
			expectRead:  9,
		},
		{
			name:        "not enough bytes",
			input:       []byte{0x01, 0x03, 0x6f, 0x6e},
			expectValue: nil,
			expectRead:  0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecSliceString(tc.input)

			assert.Equal(tc.expectValue, actualValue)
			assert.Equal(tc.expectRead, actualRead)
			if tc.expectError {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
