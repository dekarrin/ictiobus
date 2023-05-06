package fetoken

import (
	"testing"

	"github.com/dekarrin/ictiobus/types"
	"github.com/stretchr/testify/assert"
)

func TestByID(t *testing.T) {
	testCases := []struct {
		name   string
		id     string
		expect types.TokenClass
	}{
		{
			name:   "token does not exist",
			id:     "vriskaNEPETAkanayaKarkat",
			expect: nil,
		},
		{
			name:   "terminal exists",
			id:     "term",
			expect: TCTerm,
		},
		{
			name:   "non-terminal exists",
			id:     "nonterm",
			expect: TCNonterm,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := ByID(tc.id)

			assert.Equal(tc.expect, actual)
		})
	}
}
