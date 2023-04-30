// Package translation holds constructs involved in the final stage of langage
// processing. It can also serve as an entrypoint with a full-featured
// translation intepreter engine.
package trans

import (
	"fmt"

	"github.com/dekarrin/ictiobus/types"
)

type APTNodeID uint64

const (
	IDZero APTNodeID = APTNodeID(0)
)

func (id APTNodeID) String() string {
	return fmt.Sprintf("%d", id)
}

// IDGenerator should not be used directly, use NewIDGenerator. This will
// generate one that avoids the zero-value of APTNodeID.
type IDGenerator struct {
	avoidVals []APTNodeID
	seed      APTNodeID
	last      APTNodeID
	started   bool
}

func NewIDGenerator(seed int64) IDGenerator {
	return IDGenerator{
		seed:      APTNodeID(seed),
		avoidVals: []APTNodeID{IDZero},
	}
}

func (idGen *IDGenerator) Next() APTNodeID {
	var next APTNodeID
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

type AttrName string

func (nan AttrName) String() string {
	return string(nan)
}

type NodeAttrs map[string]interface{}

func (na NodeAttrs) Copy() NodeAttrs {
	newNa := NodeAttrs{}
	for k := range na {
		newNa[k] = na[k]
	}
	return newNa
}

type NodeValues struct {
	Attributes NodeAttrs

	Terminal bool

	Symbol string
}

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

// Hook is a setter of an attribute.
type Hook func(info SetterInfo, args []interface{}) (interface{}, error)
