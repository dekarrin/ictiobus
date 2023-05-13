package lex

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/syntaxerr"
)

// Token is a lexeme read from text combined with the token class it is as well
// as additional supplementary information gathered during lexing to inform
// error reporting.
type Token interface {
	// Class returns the TokenClass of the Token.
	Class() TokenClass

	// Lexeme returns the text that was lexed as the TokenClass of the Token, as
	// it appears in the source text.
	Lexeme() string

	// LinePos returns the 1-indexed character-of-line that the token appears
	// on in the source text.
	LinePos() int

	// Line returns the 1-indexed line number of the line that the token appears
	// on in the source text.
	Line() int

	// FullLine returns the full of text of the line in source that the token
	// appears on, including both anything that came before the token as well as
	// after it on the line.
	FullLine() string

	// String is the string representation.
	String() string
}

// implementation of Token interface
type lexerToken struct {
	class   TokenClass
	lexed   string
	linePos int
	lineNum int
	line    string
}

func (lt lexerToken) Class() TokenClass {
	return lt.class
}

func (lt lexerToken) Lexeme() string {
	return lt.lexed
}

func (lt lexerToken) LinePos() int {
	return lt.linePos
}

func (lt lexerToken) Line() int {
	return lt.lineNum
}

func (lt lexerToken) FullLine() string {
	return lt.line
}

func (lt lexerToken) String() string {
	// turn all newline chars into \n because we dont want that in the output
	fmtStr := "(%s <%d:%d> \"%s\")"
	content := strings.ReplaceAll(lt.lexed, "\n", "\\n")
	return fmt.Sprintf(fmtStr, strings.ToUpper(lt.class.ID()), lt.lineNum, lt.linePos, content)
}

func NewToken(class TokenClass, lexed string, linePos int, lineNum int, line string) Token {
	return lexerToken{
		class:   class,
		lexed:   lexed,
		linePos: linePos,
		lineNum: lineNum,
		line:    line,
	}
}

// NewSyntaxErrorFromToken uses the location information in the provided token
// to create a SyntaxError with a detailed message on the error and the source
// code which caused it.
func NewSyntaxErrorFromToken(msg string, tok Token) *syntaxerr.SyntaxError {
	return syntaxerr.NewSyntaxError(msg, tok.FullLine(), tok.Lexeme(), tok.Line(), tok.LinePos())
}
