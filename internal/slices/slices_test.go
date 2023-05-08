package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_In(t *testing.T) {
	testCases := []struct {
		name   string
		target string
		input  []string
		expect bool
	}{
		{
			name:   "exists",
			target: "3",
			input:  []string{"2", "dlakjfdk", "3"},
			expect: true,
		},
		{
			name:   "not exists in nil slice",
			target: "3",
			input:  nil,
			expect: false,
		},
		{
			name:   "not exists in empty slice",
			target: "3",
			input:  []string{},
			expect: false,
		},
		{
			name:   "not exists",
			target: "3",
			input:  []string{"procyon", "regulus", "vega", "polaris", "castor", "pollux", "capella", "spica", "arcturus"},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := In(tc.target, tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}
