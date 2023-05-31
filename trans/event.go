package trans

import "github.com/dekarrin/ictiobus/parse"

// EventType is a type of notable event that occurs as part of execution of a
// translation scheme. It is used as the Type of an Event to indicate which of
// its arguments are valid.
type EventType int

const (
	// EventAnnotation events are emitted immediately after the tree from the
	// parsing phase has been successfully annotated with built-in properties.
	// This will be before any event hook is called on the annotated tree.
	EventAnnotation EventType = iota

	// EventHookCall events are emitted after a hook completes execution either
	// successfully or with an error. It is emitted even if the hook returned an
	// error, however it is *not* emitted if a hook panics.
	EventHookCall
)

// Event holds information on a notable event that occurs as part of execution
// of a translation scheme. It is the main argument for calls to a listener
// given with RegisterListener.
type Event struct {

	// Type gives the type of the event that has occurred.
	Type EventType

	// ParseTree is a pointer to the original parse tree passed to the
	// translation scheme for execution.
	ParseTree *parse.Tree

	// Tree is a pointer to the root node of the full annotated parse tree that
	// the translation scheme is being executed on.
	Tree *AnnotatedTree

	// Hook contains information on the hook that was called. It will only
	// contain valid values when Type is EventHookCall.
	Hook *struct {
		// Name is the name of the hook that was called.
		Name string

		// Args contains the arguments that were passed to the hook when calling
		// it, in the order they were given.
		Args []struct {
			// Ref is a reference to the attribute whose value was used as the
			// argument value, relative to Node of the Hook.
			Ref AttrRef

			// Value is the value of the argument.
			Value interface{}
		}

		// Node is a pointer to the node of the tree that the hook was executed
		// on.
		Node *AnnotatedTree

		// Target is a reference to the attribute that the result of the hook
		// was assigned to. It is relative to Node.
		Target AttrRef

		// Result contains the return values of the called hook.
		Result struct {

			// Value is the main value returned by the hook. By convention, this
			// is only valid if Error is nil, but it is up to the hook
			// implementation to determine that.
			Value interface{}

			// Error is the error value returned by the hook.
			Error error
		}
	}
}
