package trans

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/parse"
)

// sddBinding represents a single binding of a syntax-directed definition to a
// rule in the grammar. It will be executed for all nodes created for that rule.
type sddBinding struct {
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

// String returns the string representation of the sddBinding.
func (bind sddBinding) String() string {
	attrType := "S"
	if !bind.Synthesized {
		attrType = "I"
	}

	prodStr := strings.Join(bind.BoundRuleProduction, " ")
	if prodStr == "" {
		prodStr = "ε"
	}
	rule := bind.BoundRuleSymbol + " -> [" + prodStr + "]"

	dest := bind.Dest.String()

	hook := bind.Setter

	var args string
	argsSlice := slices.Map(bind.Requirements, AttrRef.String)
	if len(argsSlice) > 0 {
		args = strings.Join(argsSlice, ", ")
	}

	fmtStr := `<%s-Attr on="%s" set="%s" hook=%q args=(%s)`
	s := fmt.Sprintf(fmtStr, attrType, rule, dest, hook, args)

	// now add the NoFlows

	if len(bind.NoFlows) > 0 {
		s += " no_flows=[" + strings.Join(bind.NoFlows, ", ") + "]"
	}

	s += ">"
	return s
}

// Copy returns a deep copy of the SDDBinding.
func (bind sddBinding) Copy() sddBinding {
	newBind := sddBinding{
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
//
// listener is called with an event of Type EventHookCall when a hook completes
// exection, regardless of whether it returned an error, as long as it doesn't
// panic. root and pt are used solely as arguments to that event, and are not
// used for any other purpose.
func (bind sddBinding) Invoke(apt *AnnotatedTree, hooksTable HookMap, listener func(Event), root *AnnotatedTree, pt *parse.Tree) (val interface{}, invokeErr error) {
	// sanity checks; can we even call this?
	if bind.Setter == "" {
		return nil, hookError{msg: "binding has no setter hook defined"}
	}
	hookFn := hooksTable[bind.Setter]
	if hookFn == nil {
		return nil, hookError{name: bind.Setter, missingHook: true, msg: fmt.Sprintf("no implementation for hook function '%s' was provided", bind.Setter)}
	}

	if bind.Dest.Rel.Type == RelHead && !bind.Synthesized {
		panic("cannot invoke inherited attribute SDD binding on head of rule")
	} else if bind.Dest.Rel.Type != RelHead && bind.Synthesized {
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
	info.GrammarSymbol, ok = apt.SymbolOf(bind.Dest.Rel)
	if !ok {
		// invalid dest
		panic(fmt.Sprintf("bound-to rule does not contain a %s", bind.Dest.Rel.String()))
	}

	// gather args
	args := []interface{}{}
	for i := range bind.Requirements {
		req := bind.Requirements[i]
		reqVal, ok := apt.AttributeValueOf(req)
		if !ok {
			// should never happen, creation of Binding should ensure this.
			_, refNodeExists := apt.AttributesOf(req.Rel)
			if !refNodeExists {
				// reference itself was invalid
				panic(fmt.Sprintf("bound-to rule does not contain a %s", req.Rel.String()))
			} else {
				errFmt := "attribute %s not defined for %s in bound-to-rule"
				errMsg := fmt.Sprintf(errFmt, req.Name, req.Rel.String())
				return nil, hookError{name: bind.Setter, msg: errMsg}
			}
		}

		args = append(args, reqVal)
	}

	// detect panic in deferred function
	defer func() {
		if r := recover(); r != nil {
			invokeErr = hookError{name: bind.Setter, msg: fmt.Sprintf("hook panicked: %v", r)}
		}
	}()

	// call func
	val, err := hookFn(info, args)

	// emit event
	if listener != nil {
		// first, gather the args and their references
		var argsWithRefs []struct {
			Ref   AttrRef
			Value interface{}
		}
		for i := range args {
			item := struct {
				Ref   AttrRef
				Value interface{}
			}{}
			item.Ref = bind.Requirements[i]
			item.Value = args[i]
			argsWithRefs = append(argsWithRefs, item)
		}

		// next, build the Result struct
		var result struct {
			Value interface{}
			Error error
		}
		result.Value = val
		result.Error = err

		// build the hook struct
		var hookInfo struct {
			Name string
			Args []struct {
				Ref   AttrRef
				Value interface{}
			}
			Node   *AnnotatedTree
			Target AttrRef
			Result struct {
				Value interface{}
				Error error
			}
		}
		hookInfo.Name = bind.Setter
		hookInfo.Args = argsWithRefs
		hookInfo.Node = apt
		hookInfo.Target = bind.Dest
		hookInfo.Result = result

		listener(Event{
			Type:      EventHookCall,
			ParseTree: pt,
			Tree:      root,
			Hook:      &hookInfo,
		})
	}

	// after event emitted, now check the return value and error if the hook
	// returned an error.
	if err != nil {
		return nil, hookError{name: bind.Setter, msg: err.Error()}
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

// Error returns a message describing the error.
func (he hookError) Error() string {
	return he.msg
}
