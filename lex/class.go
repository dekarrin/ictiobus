package lex

import (
	"strings"

	"github.com/dekarrin/ictiobus/internal/rezi"
)

// TokenClass is the class of a token in ictiobus compiler frontends. This is
// how tokens are represented in grammar, and can be considered the 'type' of a
// lexed token.
type TokenClass interface {
	// ID returns the ID of the token class. The ID must uniquely identify the
	// token within all terminals of a grammar.
	ID() string

	// Human returns a human-readable name for the token class, for use in
	// contexts such as error reporting.
	Human() string

	// Equal returns whether the TokenClass equals another. If two IDs are the
	// same, Equal must return true. TOOD: can't we replace all uses with a call
	// to ID() then? check this once move is done.
	Equal(o any) bool
}

var (
	TokenUndefined = MakeDefaultClass("<ictiobus_undefined_token>")
	TokenError     = MakeDefaultClass("<ictioubus_error>")
	TokenEndOfText = MakeDefaultClass("$")
)

// MakeDefaultClass takes a string and returns a token that both uses the
// lower-case version of the string as its ID and the un-modified string as its
// Human-readable string.
func MakeDefaultClass(s string) TokenClass {
	return NewTokenClass(strings.ToLower(s), s)
}

// implementation of TokenClass interface.
type lexerClass struct {
	id   string
	name string
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into lc.
// All of lc's fields will be replaced by the fields decoded from data.
func (lc *lexerClass) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	lc.id, n, err = rezi.DecString(data)
	if err != nil {
		return err
	}
	data = data[n:]

	lc.name, _, err = rezi.DecString(data)
	if err != nil {
		return err
	}

	return nil
}

// MarshalBinary converts lc into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (lc *lexerClass) MarshalBinary() ([]byte, error) {
	data := rezi.EncString(lc.id)
	data = append(data, rezi.EncString(lc.name)...)
	return data, nil
}

func (lc *lexerClass) ID() string {
	return lc.id
}

func (lc *lexerClass) Human() string {
	return lc.name
}

func (lc *lexerClass) Equal(o any) bool {
	other, ok := o.(TokenClass)
	if !ok {
		otherPtr, ok := o.(*TokenClass)
		if !ok {
			return false
		}
		if otherPtr == nil {
			return false
		}
		other = *otherPtr
	}

	return other.ID() == lc.ID()
}

// NewTokenClass creates a new token class with the given id and human readable
// name.
func NewTokenClass(id string, human string) TokenClass {
	return &lexerClass{id: id, name: human}
}
