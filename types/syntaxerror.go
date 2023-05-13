package types

import (
	"fmt"
	"strings"
)

// SyntaxError is an error returned when there is a problem with the syntax of
// analyzed code. For reporting errors to an end-user, calling FullMessage is
// recommended over Error, as it will output context and location of the error.
type SyntaxError struct {
	sourceLine string
	source     string

	// line that error occured on, 1-indexed.
	line int

	// position in line of error, 1-indexed.
	pos     int
	message string
}

// NewSyntaxError creates a new SyntaxError with its properties set.
func NewSyntaxError(msg string, sourceLine string, source string, line int, pos int) *SyntaxError {
	return &SyntaxError{
		sourceLine: sourceLine,
		source:     source,
		line:       line,
		pos:        pos,
		message:    msg,
	}
}

// Error returns the string representation of the SyntaxError.
func (se SyntaxError) Error() string {
	if se.line == 0 {
		return fmt.Sprintf("syntax error: %s", se.message)
	}

	return fmt.Sprintf("syntax error: around line %d, char %d: %s", se.line, se.pos, se.message)
}

// Source returns the exact text of the specific source code that caused the
// issue. If no particular source was the cause (such as for unexpected EOF
// errors), this will return an empty string.
func (se SyntaxError) Source() string {
	return se.source
}

// Line returns the line the error occured on. Lines are 1-indexed. This will
// return 0 if the line is not set.
func (se SyntaxError) Line() int {
	return se.line
}

// Position returns the character position that the error occured on. Character
// positions are 1-indexed. This will return 0 if the character position is not
// set.
func (se SyntaxError) Position() int {
	return se.pos
}

// FullMessage shows the complete message of the error string along with the
// offending line and a cursor to the problem position in a formatted way.
func (se SyntaxError) FullMessage() string {
	errMsg := se.Error()

	if se.line != 0 {
		errMsg = se.SourceLineWithCursor() + "\n" + errMsg
	}

	return errMsg
}

// MessageForFile returns the full error message in the format of
// filename:line:pos: message, followed by the syntax error itself.
func (se SyntaxError) MessageForFile(filename string) string {
	var msg string

	if se.line != 0 {
		msg = fmt.Sprintf("%s:%d:%d: %s\n%s", filename, se.line, se.pos, se.message, se.SourceLineWithCursor())
	} else {
		msg = fmt.Sprintf("%s: %s", filename, msg)
	}

	return msg
}

// SourceLineWithCursor returns the source offending code on one line and
// directly under it a cursor showing where the error occured.
//
// Returns a blank string if no source line was provided for the error (such as
// for unexpected EOF errors).
func (se SyntaxError) SourceLineWithCursor() string {
	if se.sourceLine == "" {
		return ""
	}

	cursorLine := ""
	// pos will be 1-indexed.
	for i := 0; i < se.pos-1 && i < len(se.sourceLine); i++ {
		if se.sourceLine[i] == '\t' {
			cursorLine += "    "
		} else {
			cursorLine += " "
		}
	}

	return strings.ReplaceAll(se.sourceLine, "\t", "    ") + "\n" + cursorLine + "^"
}
