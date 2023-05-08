package textfmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_TruncateWith(t *testing.T) {
	testCases := []struct {
		name   string
		s      string
		maxLen int
		cont   string
		expect string
	}{
		{
			name:   "not needed",
			s:      "test",
			maxLen: 5,
			cont:   "...",
			expect: "test",
		},
		{
			name:   "needed",
			s:      "test of a long string",
			maxLen: 10,
			cont:   "...",
			expect: "test of a ...",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := TruncateWith(tc.s, tc.maxLen, tc.cont)

			assert.Equal(tc.expect, actual)
		})
	}
}
