package types

import (
	"strings"

	"github.com/dekarrin/ictiobus/internal/decbin"
)

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

type simpleTokenClass string

func (class *simpleTokenClass) UnmarshalBinary(data []byte) error {
	s, _, err := decbin.DecString(data)
	if err != nil {
		return err
	}

	*class = simpleTokenClass(s)
	return nil
}

func (class *simpleTokenClass) MarshalBinary() ([]byte, error) {
	return decbin.EncString(string(*class)), nil
}

func (class *simpleTokenClass) ID() string {
	return strings.ToLower(string(*class))
}

func (class *simpleTokenClass) Human() string {
	return string(*class)
}

func (class *simpleTokenClass) Equal(o any) bool {
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

	return other.ID() == class.ID()
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
	tc := simpleTokenClass(s)
	return &tc
}
