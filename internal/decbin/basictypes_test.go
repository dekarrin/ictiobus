package decbin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncBool(t *testing.T) {
	testCases := []struct {
		name   string
		input  bool
		expect []byte
	}{
		{
			name:   "true",
			input:  true,
			expect: []byte{0x01},
		},
		{
			name:   "false",
			input:  false,
			expect: []byte{0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := EncBool(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_DecBool(t *testing.T) {
	testCases := []struct {
		name        string
		input       []byte
		expectValue bool
		expectRead  int
		expectError bool
	}{
		{
			name:        "true from exact value",
			input:       []byte{0x01},
			expectValue: true,
			expectRead:  1,
		},
		{
			name:        "true from sequence",
			input:       []byte{0x01, 0x00},
			expectValue: true,
			expectRead:  1,
		},
		{
			name:        "false from exact value",
			input:       []byte{0x00},
			expectValue: false,
			expectRead:  1,
		},
		{
			name:        "false from sequence",
			input:       []byte{0x00, 0x01},
			expectValue: false,
			expectRead:  1,
		},
		{
			name:        "error from exact value - 0x02",
			input:       []byte{0x02},
			expectError: true,
		},
		{
			name:        "error from exact value - 0xff",
			input:       []byte{0xff},
			expectError: true,
		},
		{
			name:        "error from sequence",
			input:       []byte{0x25, 0xab, 0xcc},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualValue, actualRead, err := DecBool(tc.input)
			if tc.expectError {
				if !assert.Error(err) {
					return
				}
			} else if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expectValue, actualValue)
			assert.Equal(tc.expectRead, actualRead, "num read bytes does not match expected")
		})
	}
}
