package box

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlphabetized_int(t *testing.T) {
	testCases := []struct {
		name   string
		input  Container[int]
		expect []int
	}{
		{
			name:   "no elements - set",
			input:  NewKeySet[int](),
			expect: []int{},
		},
		{
			name:   "three elements - set",
			input:  NewKeySet(map[int]bool{6: false, 284: false, 2: false}),
			expect: []int{2, 284, 6},
		},
		{
			name:   "hpair",
			input:  HPairOf(2, 2000),
			expect: []int{2, 2000},
		},
		{
			name:   "htriple",
			input:  HTripleOf(123, 456, 8),
			expect: []int{123, 456, 8},
		},
		{
			name:   "hquad",
			input:  HQuadrupleOf(2, 2000, 284, 1),
			expect: []int{1, 2, 2000, 284},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := Alphabetized(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func TestAlphabetized_string(t *testing.T) {
	testCases := []struct {
		name   string
		input  Container[string]
		expect []string
	}{
		{
			name:   "no elements - set",
			input:  NewKeySet[string](),
			expect: []string{},
		},
		{
			name:   "three elements - set",
			input:  NewKeySet(map[string]bool{"this": false, "": false, "vriska": false, "nepeta": false}),
			expect: []string{"", "nepeta", "this", "vriska"},
		},
		{
			name:   "hpair",
			input:  HPairOf("kanaya", "aradia"),
			expect: []string{"aradia", "kanaya"},
		},
		{
			name:   "htriple",
			input:  HTripleOf("tavros", "sollux", "equius"),
			expect: []string{"equius", "sollux", "tavros"},
		},
		{
			name:   "hquad",
			input:  HQuadrupleOf("terezi", "eridan", "feferi", "aradia"),
			expect: []string{"aradia", "eridan", "feferi", "terezi"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := Alphabetized(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}
