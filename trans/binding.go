package trans

import "fmt"

type SDDBinding struct {
	// Synthesized is whether the binding is for a
	Synthesized bool

	// BoundRuleSymbol is the head symbol of the rule the binding is on.
	BoundRuleSymbol string

	// BoundRuleProduction is the list of produced symbols of the rule the
	// binding is on.
	BoundRuleProduction []string

	// Requirements is the attribute references that this binding needs to
	// compute its value. Values corresponding to the references are passed in
	// to calls to Setter via its args slice in the order they are specified
	// here.
	Requirements []AttrRef

	// Dest is the destination.
	Dest AttrRef

	// Setter is name of the hook to call to calculate a value of the node by
	// the binding. A hooks table is used to look up the hook function and call
	// it, when needed.
	Setter string

	// NoFlows is the list of parents that this binding is allowed to not flow
	// up to without causing error.
	NoFlows []string
}

func (bind SDDBinding) Copy() SDDBinding {
	newBind := SDDBinding{
		Synthesized:         bind.Synthesized,
		BoundRuleSymbol:     bind.BoundRuleSymbol,
		BoundRuleProduction: make([]string, len(bind.BoundRuleProduction)),
		Requirements:        make([]AttrRef, len(bind.Requirements)),
		Dest:                bind.Dest,
		Setter:              bind.Setter,
		NoFlows:             make([]string, len(bind.NoFlows)),
	}

	copy(newBind.BoundRuleProduction, bind.BoundRuleProduction)
	copy(newBind.Requirements, bind.Requirements)
	copy(newBind.NoFlows, bind.NoFlows)

	return newBind
}

// Invoke calls the given binding while visiting an annotated parse tree node.
func (bind SDDBinding) Invoke(apt *AnnotatedParseTree, hooksTable HookMap) (val interface{}, invokeErr error) {
	// sanity checks; can we even call this?
	if bind.Setter == "" {
		return nil, hookError{msg: "binding has no setter hook defined"}
	}
	hookFn := hooksTable[bind.Setter]
	if hookFn == nil {
		return nil, hookError{name: bind.Setter, missingHook: true, msg: fmt.Sprintf("no implementation for hook function '%s' was provided", bind.Setter)}
	}

	if bind.Dest.Relation.Type == RelHead && !bind.Synthesized {
		panic("cannot invoke inherited attribute SDD binding on head of rule")
	} else if bind.Dest.Relation.Type != RelHead && bind.Synthesized {
		panic("cannot invoke synthesized attribute SDD binding on production of rule")
	}

	// gather info on the attribute being set
	info := SetterInfo{
		Name:       bind.Dest.Name,
		Synthetic:  bind.Synthesized,
		FirstToken: apt.First(),
	}

	// symbol of who it is for
	var ok bool
	info.GrammarSymbol, ok = apt.SymbolOf(bind.Dest.Relation)
	if !ok {
		// invalid dest
		panic(fmt.Sprintf("bound-to rule does not contain a %s", bind.Dest.Relation.String()))
	}

	// gather args
	args := []interface{}{}
	for i := range bind.Requirements {
		req := bind.Requirements[i]
		reqVal, ok := apt.AttributeValueOf(req)
		if !ok {
			// should never happen, creation of Binding should ensure this.
			_, refNodeExists := apt.AttributesOf(req.Relation)
			if !refNodeExists {
				// reference itself was invalid
				panic(fmt.Sprintf("bound-to rule does not contain a %s", req.Relation.String()))
			} else {
				panic(fmt.Sprintf("attribute %s not yet defined for %s in bound-to-rule", req.Name, req.Relation.String()))
			}
		}

		args = append(args, reqVal)
	}

	// detect panic in deferred function
	defer func() {
		if r := recover(); r != nil {
			invokeErr = hookError{name: bind.Setter, msg: fmt.Sprintf("panicked: %v", r)}
		}
	}()

	// call func
	val, err := hookFn(info, args)
	if err != nil {
		return invokeErr, hookError{name: bind.Setter, msg: err.Error()}
	}

	return val, nil
}

// highly-populated error struct containing inform8ion about an error that
// occured during invocation of a binding due to a problem with the associated
// hook function. hook func could be missing, not set, or could have returned an
// error. if name is empty, then the error is that the hook was set to an empty
// string or never set. If missingHook is set, the the name was set but the
// hook was not found in the hooks table. Otherwise the error is in msg.
type hookError struct {
	// the name of the hook function.
	name string

	missingHook bool

	msg string
}

func (he hookError) Error() string {
	return he.msg
}
