// Package trans provides syntax-directed translations of parse trees for the
// ictiobus parser generator. It is involved in the final stage of input
// analysis. A complete [SDTS]'s Evaluate() function is called with a
// [parse.Tree] as input to produce the final result of the analysis, the
// intermediate representation.
//
// At this time, while there are function stubs and supposed availability of
// inherited attributes in an SDTS, only S-attributed attribute grammars are
// supported at this time. Attempting to use inherited attributes will result in
// untested and undefined behavior.
package trans

import (
	"fmt"

	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
)

// SDTS is a series of syntax-directed translations bound to syntactic rules of
// a grammar. It is used for evaluation of a parse tree into an intermediate
// representation, or for direct execution.
//
// This is a representation of the additions to a grammar which would make it an
// attribute gramamr.
type SDTS interface {
	// Evaluate takes a parse tree and executes the semantic actions defined as
	// SDDBindings for a node for each node in the tree and on completion,
	// returns the requested attributes values from the root node. Execution
	// order is automatically determined by taking the dependency graph of the
	// SDTS; cycles are not supported. Do note that this does not require the
	// SDTS to be S-attributed or L-attributed, only that it not have cycles in
	// its value dependency graph.
	//
	// Warn errors are provided in the slice of error and can be populated
	// regardless of whether the final (actual) error is non-nil.
	Evaluate(tree parse.Tree, attributes ...string) (vals []interface{}, warns []error, err error)

	// SetHooks sets the hook table for mapping SDTS hook names as used in a
	// call to BindSynthesizedAttribute or BindInheritedAttribute to their
	// actual implementations.
	//
	// Because the map from strings to function pointers, this hook map must be
	// set at least once before the SDTS is used. It is recommended to set it
	// every time the SDTS is loaded as soon as it is loaded.
	//
	// Calling it multiple times will add to the existing hook table, not
	// replace it entirely. If there are any duplicate hook names, the last one
	// set will be the one that is used.
	SetHooks(hooks HookMap)

	// BindI creates a new SDTS binding for setting the value of an inherited
	// attribute with name attrName. The production that the inherited attribute
	// is set on is specified with forProd, which must have its Type set to
	// something other than RelHead (inherited attributes can be set only on
	// production symbols).
	//
	// Inherited properties on SDTSs are not supported or tested at this time.
	// This function should only be called for experimental purposes.
	//
	// The binding applies only on nodes in the parse tree created by parsing
	// the grammar rule productions with head symbol head and production symbols
	// prod.
	//
	// The AttributeSetter bound to hook is called when the inherited value
	// attrName is to be set, in order to calculate the new value. Attribute
	// values to pass in as arguments are specified by passing references to the
	// node and attribute name whose value to retrieve in the withArgs slice.
	// Explicitlygiving the referenced attributes in this fashion makes it easy
	// to determine the dependency graph for later execution.
	BindI(head string, prod []string, attrName string, hook string, withArgs []AttrRef, forProd NodeRelation) error

	// Bind creates a new SDTS binding for setting the value of a synthesized
	// attribute with name attrName. The attribute is set on the symbol at the
	// head of the rule that the binding is being created for.
	//
	// The binding will be applied to nodes in the parse tree created by parsing
	// the grammar rule productions with head symbol head and production symbols
	// prod.
	//
	// The Hook implementation bound to hook is called when the synthesized
	// value attrName is to be set, in order to calculate the new value.
	// Attribute values to pass in as arguments are specified by passing
	// references to the node and attribute name whose value to retrieve in the
	// withArgs slice. Explicitly giving the referenced attributes in this
	// fashion makes it easy to determine the dependency graph for later
	// execution.
	Bind(head string, prod []string, attrName string, hook string, withArgs []AttrRef) error

	// String returns a string representation of the SDTS.
	String() string
}

// NewSDTS creates a new, empty Syntax-Directed Translation Scheme.
func NewSDTS() SDTS {
	impl := sdtsImpl{
		bindings: map[string]map[string][]sddBinding{},
	}
	return &impl
}

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
	FirstToken lex.Token

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
