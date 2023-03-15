// Package translation holds constructs involved in the final stage of langage
// processing. It can also serve as an entrypoint with a full-featured
// translation intepreter engine.
package translation

type APTNodeID uint64

const (
	IDZero APTNodeID = APTNodeID(0)
)

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

type AttributeSetter func(symbol string, name string, args []interface{}) interface{}
