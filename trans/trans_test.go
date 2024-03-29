package trans

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IDGenerator_Next(t *testing.T) {
	testCases := []struct {
		name       string
		seed       int64
		extraTimes int
		expect     aptNodeID
	}{
		{
			name:       "First call gives number",
			seed:       12,
			extraTimes: 0,
			expect:     aptNodeID(12),
		},
		{
			name:       "multiple calls",
			seed:       12,
			extraTimes: 20,
			expect:     aptNodeID(32),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			idGen := newIDGenerator(tc.seed)

			for i := 0; i < tc.extraTimes; i++ {
				idGen.Next()
			}

			actual := idGen.Next()

			assert.Equal(tc.expect, actual)
		})
	}
}
