package fetoken

import (
	"testing"

	"github.com/dekarrin/ictiobus/trans"
	"github.com/stretchr/testify/assert"
)

func TestTokensExist(t *testing.T) {
	testCases := []struct {
		name      string
		info      trans.SetterInfo
		args      []interface{}
		expect    interface{}
		expectErr bool
	}{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual interface{}
			var err error

			if tc.expectErr {
				assert.Error(err)
				return
			}

			if !assert.NoError(err) {
				return
			}
			assert.Equal(tc.expect, actual)
		})
	}
}
