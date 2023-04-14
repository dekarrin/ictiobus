package fishi

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dekarrin/ictiobus/fishi/format"
	"github.com/dekarrin/ictiobus/types"

	"github.com/stretchr/testify/assert"
)

func Test_WithFakeInput(t *testing.T) {
	assert := assert.New(t)

	r, err := format.NewCodeReader(bytes.NewReader([]byte(testInput)))
	if !assert.NoError(err) {
		return
	}

	_, actual := Parse(r, Options{ParserCFF: "../fishi-parser.cff", ReadCache: true, WriteCache: true})

	assert.NoError(actual)

	if actual != nil {
		actualSynt, ok := actual.(*types.SyntaxError)
		if ok {
			fmt.Println(actualSynt.FullMessage())
		}
	}
}

func Test_SelfHostedMarkdown_Spec(t *testing.T) {
	assert := assert.New(t)

	res, err := ParseMarkdownFile("../fishi.md", Options{ParserCFF: "../fishi-parser.cff", ReadCache: true, WriteCache: true})
	if !assert.NoError(err) {
		return
	}

	_, _, actualErr := NewSpec(*res.AST)

	if actualErr != nil {
		actualSynt, ok := actualErr.(*types.SyntaxError)
		if ok {
			fmt.Println(actualSynt.FullMessage())
		}
	}
}

func Test_SelfHostedMarkdown(t *testing.T) {
	assert := assert.New(t)

	_, actual := ParseMarkdownFile("../fishi.md", Options{ParserCFF: "../fishi-parser.cff", ReadCache: true, WriteCache: true})

	assert.NoError(actual)

	if actual != nil {
		actualSynt, ok := actual.(*types.SyntaxError)
		if ok {
			fmt.Println(actualSynt.FullMessage())
		}
	}
}

const (
	testInput = `%%actions
	
						%symbol
	
	
						{hey}
						%prod  %index 8
	
					%set {thing}.thing %hook thing
						%prod {}
	
					%set {thing}.thing %hook thing
						%prod {test} this {THING}
	
						%set {thing}.thing %hook thing
					%prod {ye} + {A}
	
					%set {thing}.thing %hook thing
	
							%symbol {yo}%prod + {EAT} ext
	
					%set {thing}.thing %hook thing
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
				%set {text-element}.str
				%hook identity  %with {0}.$text
	
				%prod ESCSEQ
				%set {text-element}.str
				%hook unescape  %with {.}.$test
	
	
				%symbol {OTHER}
				%prod EHHH
				%set {OTHER}.str
				%hook identity  %with {9}.$text
	
				%prod ESCSEQ
				%set {text-element$12}.str
				%hook unescape  %with {^}.$test
	
				%state someGoodState
	
				%symbol {text-element}
				%prod FREEFORM_TEXT
				%set {text-element}.str
				%hook identity  %with {ANON$12}.$text
	
				%prod ESCSEQ
				%set {text-element}.str
				%hook unescape  %with {ESCSEQ}.$test
	
				`
)
