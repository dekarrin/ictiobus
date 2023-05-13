// Package trans provides syntax-directed translations of parse trees for the
// ictiobus parser generator. It is involved in the final stage of input
// analysis. It can also serve as an entrypoint with a full-featured translation
// intepreter engine.
package trans

import (
	"fmt"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/types"
)

// SDTS is a series of syntax-directed translations bound to syntactic rules of
// a grammar. It is used for evaluation of a parse tree into an intermediate
// representation, or for direct execution.
//
// Strictly speaking, this is closer to an Attribute grammar.
type SDTS interface {
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

	// BindInheritedAttribute creates a new SDTS binding for setting the value
	// of an inherited attribute with name attrName. The production that the
	// inherited attribute is set on is specified with forProd, which must have
	// its Type set to something other than RelHead (inherited attributes can be
	// set only on production symbols).
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
	BindInheritedAttribute(head string, prod []string, attrName string, hook string, withArgs []AttrRef, forProd NodeRelation) error

	// BindSynthesizedAttribute creates a new SDTS binding for setting the value
	// of a synthesized attribute with name attrName. The attribute is set on
	// the symbol at the head of the rule that the binding is being created for.
	//
	// The binding applies only on nodes in the parse tree created by parsing
	// the grammar rule productions with head symbol head and production symbols
	// prod.
	//
	// The AttributeSetter bound to hook is called when the synthesized value
	// attrName is to be set, in order to calculate the new value. Attribute
	// values to pass in as arguments are specified by passing references to the
	// node and attribute name whose value to retrieve in the withArgs slice.
	// Explicitly giving the referenced attributes in this fashion makes it easy
	// to determine the dependency graph for later execution.
	BindSynthesizedAttribute(head string, prod []string, attrName string, hook string, withArgs []AttrRef) error

	// SetNoFlow sets a binding to be explicitly allowed to not be required to
	// flow up to a particular parent. This will prevent it from causing an
	// error if it results in a disconnected dependency graph if the node of
	// that binding has the given parent.
	//
	// - forProd is only used if synth is false. It specifies the production
	// that the binding to match must apply to.
	// - which is the index of the binding to set it on, if multiple match the
	// prior criteria. Set to -1 or less to set it on all matching bindings.
	// - ifParent is the symbol that the parent of the node must be for no flow
	// to be considered acceptable.
	SetNoFlow(synth bool, head string, prod []string, attrName string, forProd NodeRelation, which int, ifParent string) error

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
	Evaluate(tree types.ParseTree, attributes ...string) (vals []interface{}, warns []error, err error)

	// Validate checks whether this SDTS is valid for the given grammar. It will
	// create a simulated parse tree that contains a node for every rule of the
	// given grammar and will attempt to evaluate it, returning an error if
	// there is any issue running the bindings.
	//
	// fakeValProducer should be a map of token class IDs to functions that can
	// produce fake values for the given token class. This is used to simulate
	// actual lexemes in the parse tree. If not provided, entirely contrived
	// values will be used, which may not behave as expected with the SDTS. To
	// get one that will use the configured regexes of tokens used for lexing,
	// call FakeLexemeProducer on a Lexer.
	Validate(grammar grammar.Grammar, attribute string, debug ValidationOptions, fakeValProducer ...map[string]func() string) (warns []string, err error)
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
