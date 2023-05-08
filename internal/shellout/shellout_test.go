package shellout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: These tests may fail if directly executed on windows.

func Test_Shell_Environment(t *testing.T) {
	testCases := []struct {
		name    string
		input   Shell
		cmd     string
		cmdArgs []string
		expect  string
	}{
		{
			name:    "no env set",
			input:   Shell{},
			cmd:     "bash",
			cmdArgs: []string{"-c", "echo -n $test_var"},
			expect:  "",
		},
		{
			name: "env with $test_var set",
			input: Shell{
				Env: []string{"test_var=3"},
			},
			cmd:     "bash",
			cmdArgs: []string{"-c", "echo -n $test_var"},
			expect:  "3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, err := tc.input.Exec(tc.cmd, tc.cmdArgs...)

			if !assert.NoError(err) {
				return
			}

			assert.Equal(tc.expect, actual)
		})
	}
}
