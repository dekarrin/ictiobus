package syntaxerr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	testCases := []struct {
		name       string
		sourceLine string
		source     string
		pos        int
		line       int
	}{
		{
			name:       "normal creation",
			sourceLine: "a := 27 + 3",
			source:     "27",
			pos:        5,
			line:       300,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := New("test msg", tc.sourceLine, tc.source, tc.line, tc.pos)

			assert.Equal(tc.line, actual.Line())
			assert.Equal(tc.source, actual.Source())
			assert.Equal(tc.pos, actual.Position())
		})
	}

}
