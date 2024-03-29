package lex

import (
	"io"
)

type immediateTokenStream struct {
	tokens []Token
	cur    int
}

// ImmediatelyLex returns a token stream that goes through the entire input in
// the provided reader and lexes all of them, returning them as a TokenStream
// which will return them one at a time as its Next() function is called. If any
// lexing errors occur, they will be immediately returned as a non-nil error.
func (lx *lexerTemplate) ImmediatelyLex(input io.Reader) (TokenStream, error) {
	// an immediate lexer is simply a 'lazy' lexer that just, keeps going. so
	// make one of those.
	lazyCore, err := lx.LazyLex(input)
	if err != nil {
		return nil, err
	}

	lexedTokens := []Token{}

	for lazyCore.HasNext() {
		tok := lazyCore.Next()

		// if it's an error token, capture that and turn it into a proper
		// 'syntax error' style error (technically it's a lexical specification
		// error but lets not split hairs over that)
		if tok.Class().ID() == TokenError.ID() {
			// stop. do not allow panic mode to continue, lexing has failed.

			// create a new token to hold all values of tok except lexeme so we
			// don't put the lexeme of "err message" into the actual token
			// shown when the error is displayed to end user
			tokWrap := lexerToken{
				class:   tok.Class(),
				linePos: tok.LinePos(),
				line:    tok.FullLine(),
				lineNum: tok.Line(),
			}

			return nil, NewSyntaxErrorFromToken(tok.Lexeme(), tokWrap)
		}

		lexedTokens = append(lexedTokens, tok)
	}

	// and we are now done with the pre-lex.
	return &immediateTokenStream{tokens: lexedTokens}, nil
}

// Next returns the next token in the stream and advances the stream by one
// token. If at the end of the stream, this will return a token whose Class()
// is types.TokenEndOfText. If an error in lexing occurs, it will return a token
// whose Class() is types.TokenError and whose lexeme is a message explaining
// the error.
func (lx *immediateTokenStream) Next() Token {
	n := lx.tokens[lx.cur]
	lx.cur++
	return n
}

// Peek returns the next token in the stream without advancing the stream.
func (lx *immediateTokenStream) Peek() Token {
	return lx.tokens[lx.cur]
}

// HasNext returns whether the stream has any additional tokens.
func (lx *immediateTokenStream) HasNext() bool {
	return lx.Remaining() > 0
}

// Remaining returns the number of tokens remaining in the token stream.
func (lx *immediateTokenStream) Remaining() int {
	return len(lx.tokens) - lx.cur
}
