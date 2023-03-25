package fishi

import (
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/types"

	"github.com/stretchr/testify/assert"
)

func Test_WithFakeInput(t *testing.T) {
	assert := assert.New(t)

	_, actual := Execute([]byte(testInput), Options{ParserCFF: "../fishi-parser.cff", ReadCache: true, WriteCache: true, SDTSValidate: true})

	assert.NoError(actual)

	if actual != nil {
		actualSynt, ok := actual.(*types.SyntaxError)
		if ok {
			fmt.Println(actualSynt.FullMessage())
		}
	}
}

func Test_SelfHostedMarkdown(t *testing.T) {
	assert := assert.New(t)

	_, actual := ExecuteMarkdownFile("../fishi.md", Options{ParserCFF: "../fishi-parser.cff", ReadCache: true, WriteCache: true, SDTSValidate: true})

	assert.NoError(actual)

	if actual != nil {
		actualSynt, ok := actual.(*types.SyntaxError)
		if ok {
			fmt.Println(actualSynt.FullMessage())
		}
	}

	assert.False(true)
}

func Test_GetFishiFromMarkdown(t *testing.T) {
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
				"%action go\n" +
				"```\n" +
				"other text\n",
			expect: "%%tokens\n" +
				"\n" +
				"%token test\n" +
				"\n" +
				"%token 7\n" +
				"%%actions\n" +
				"\n" +
				"%action go\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := GetFishiFromMarkdown([]byte(tc.input))

			assert.Equal(tc.expect, string(actual))
		})
	}
}

const (
	testInput = `%%actions
	
						%symbol
	
	
						{hey}
						%prod  %index 8
	
					%action {thing}.thing %hook thing
						%prod {}
	
					%action {thing}.thing %hook thing
						%prod {test} this {THING}
	
						%action {thing}.thing %hook thing
					%prod {ye} + {A}
	
					%action {thing}.thing %hook thing
	
							%symbol {yo}%prod + {EAT} ext
	
					%action {thing}.thing %hook thing
					%%tokens
					[somefin]
	
					%stateshift   someState
	
			%%tokens
	
			%!%[more]%!%bluggleb*shi{2,4}   %stateshift glub
			%token lovely %human Something for this
	
				%%tokens
	
					glub  %discard
	
	
					[some]{FREEFORM}idk[^bullshit]text\*
					%discard
	
					%!%[more]%!%bluggleb*shi{2,4}   %stateshift glub
				%token lovely %human Something nice
					%priority 1
	
				%state this
	
				[yo] %discard
	
				%%grammar
				%state glub
				{RULE} =   {SOMEBULLSHIT}
	
							%%grammar
							{RULE}=                           {WOAH} | n
							{R2}				= =+  {DAMN} cool | okaythen + 2 | {}
											 | {SOMEFIN ELSE}
	
							%state someState
	
							{ANOTHER}=		{HMM}
	
	
	
	
				%%actions
	
				%symbol {text-element}
				%prod FREEFORM_TEXT
				%action {text-element}.str
				%hook identity  %with FREEFORM_TEXT.$text
	
				%prod ESCSEQ
				%action {text-element}.str
				%hook unescape  %with ESCSEQ.$test
	
	
				%symbol {OTHER}
				%prod EHHH
				%action {OTHER}.str
				%hook identity  %with FREEFORM_TEXT.$text
	
				%prod ESCSEQ
				%action {text-element}$12.str
				%hook unescape  %with ESCSEQ.$test
	
				%state someGoodState
	
				%symbol {text-element}
				%prod FREEFORM_TEXT
				%action {text-element}.str
				%hook identity  %with FREEFORM_TEXT.$text
	
				%prod ESCSEQ
				%action {text-element}.str
				%hook unescape  %with ESCSEQ.$test
	
				`
)
