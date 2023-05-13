package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_scanMarkdownForFishiBlocks(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name: "fishi and text",
			input: "Test block\n" +
				"only include the fishi block\n" +
				"```fishi\n" +
				"%%tokens\n" +
				"\n" +
				"%token test\n" +
				"```\n",
			expect: "%%tokens\n" +
				"\n" +
				"%token test\n",
		},
		{
			name: "two fishi blocks",
			input: "Test block\n" +
				"only include the fishi blocks\n" +
				"```fishi\n" +
				"%%tokens\n" +
				"\n" +
				"%token test\n" +
				"```\n" +
				"some more text\n" +
				"```fishi\n" +
				"\n" +
				"%token 7\n" +
				"%%actions\n" +
				"\n" +
				"%set go\n" +
				"```\n" +
				"other text\n",
			expect: "%%tokens\n" +
				"\n" +
				"%token test\n" +
				"\n" +
				"%token 7\n" +
				"%%actions\n" +
				"\n" +
				"%set go\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := scanMarkdownForFishiBlocks([]byte(tc.input))

			assert.Equal(tc.expect, string(actual))
		})
	}
}
