// Package trans provides syntax-directed translations of parse trees for the
// ictiobus parser generator. It is involved in the final stage of input
// analysis. It can also serve as an entrypoint with a full-featured translation
// intepreter engine.
package trans

import (
	"fmt"

	"github.com/dekarrin/ictiobus/types"
)

// SetterInfo is struct passed to all bound hooks in a translation scheme to
// provide information on what is being set. It includes the grammar symbol of
// the node it is being set for, the first token of that symbol as it was
// originally lexed, the name of the attribute that the return value of the hook
// will be assigned to, and whether the attribute is synthetic.
type SetterInfo struct {
	// The name of the grammar symbol of the particular node that the attribute
	// is being set on.
	GrammarSymbol string

	// The first token of the grammar symbol that the attribute is being set on.
	FirstToken types.Token

	// The name of the attribute being set.
	Name string

	// Whether the attribute is a synthetic attribute.
	Synthetic bool
}

// Hook takes arguments from other attributes in an annotated parse tree and
// returns a value to set another attribute to. It can return an error if it
// encounters any issues.
type Hook func(info SetterInfo, args []interface{}) (interface{}, error)

// HookMap is a mapping of hook names to hook functions. This is used for
// defining implementation functions for hooks named in a FISHI specification.
type HookMap map[string]Hook

// aptNodeID is the type of the built-in '$id' attribute in annotated parse tree
// nodes.
type aptNodeID uint64

const (
	// aptIDZero is the zero value for an APTNodeID.
	aptIDZero aptNodeID = aptNodeID(0)
)

// String returns the string representation of an APTNodeID.
func (id aptNodeID) String() string {
	return fmt.Sprintf("%d", id)
}

// idGenerator generates unique APTNodeIDs. It should not be used directly; use
// [NewIDGenerator], which will create one that avoids the zero-value of
// APTNodeID.
type idGenerator struct {
	avoidVals []aptNodeID
	seed      aptNodeID
	last      aptNodeID
	started   bool
}

// Creates an IDGenerator that begins with the given seed.
func newIDGenerator(seed int64) idGenerator {
	return idGenerator{
		seed:      aptNodeID(seed),
		avoidVals: []aptNodeID{aptIDZero},
	}
}

// Next generates a unique APTNodeID.
func (idGen *idGenerator) Next() aptNodeID {
	var next aptNodeID
	var valid bool

	for !valid {
		if !idGen.started {
			// then next is set to seed-value
			idGen.started = true
			next = idGen.seed
		} else {
			next = idGen.last + 1
		}
		idGen.last = next

		valid = true
		for i := range idGen.avoidVals {
			if idGen.avoidVals[i] == next {
				valid = false
				break
			}
		}
	}

	return next
}

// nodeAttrs is the type of the attributes map that holds the values of
// attributes on a node of an annotated parse tree.
type nodeAttrs map[string]interface{}

// Copy returns a deep copy of a NodeAttrs.
func (na nodeAttrs) Copy() nodeAttrs {
	newNa := nodeAttrs{}
	for k := range na {
		newNa[k] = na[k]
	}
	return newNa
}
