package trans

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/parse"
)

// sdts.go contains the implementation of a Syntax-Directed Translation Scheme.

type sdtsImpl struct {
	hooks    HookMap
	bindings map[string]map[string][]sddBinding
	listener func(Event)
}

// String returns the string representation of the sdtsImpl.
func (sdts *sdtsImpl) String() string {
	if sdts == nil {
		return "<nil>"
	}

	var sb strings.Builder
	sb.WriteString("sdtsImpl<")
	if sdts.bindings == nil {
		sb.WriteString("(nil bindings) ")
	} else {
		sb.WriteRune('\n')

		// need ordered listing for debug output
		rNames := slices.Keys(sdts.bindings)
		sort.Strings(rNames)
		for _, rName := range rNames {
			rProds := slices.Keys(sdts.bindings[rName])
			sort.Strings(rProds)
			for _, rProd := range rProds {
				sddsForRule := sdts.bindings[rName][rProd]
				for _, sdd := range sddsForRule {
					sb.WriteRune('\t')
					sb.WriteString(sdd.String())
					sb.WriteRune('\n')
				}
			}
		}
		sb.WriteString("\t-- ")
	}
	sb.WriteString("hook implementations:")

	if len(sdts.hooks) == 0 {
		sb.WriteString(" (none)")
	} else {
		// uses this inste
		hooks := slices.Keys(sdts.hooks)
		alphaHooks := slices.SortBy(hooks, func(l, r string) bool {
			return l < r
		})

		sb.WriteString("\n")
		for _, hook := range alphaHooks {
			sb.WriteRune('\t')
			sb.WriteString(hook)
			sb.WriteRune('\n')
		}
	}

	sb.WriteRune('>')

	return sb.String()
}

// RegisterListener registers a callback function to be called whenever an event
// occurs during translation. This can be used for debugging. If nil is passed
// as the listener, it will disable sending events to it.
func (sdts *sdtsImpl) RegisterListener(listener func(Event)) {
	sdts.listener = listener

}

// SetHooks sets the hook mapping containing implementations of the hooks in the
// SDTS. It must be called before calling Evaluate.
func (sdts *sdtsImpl) SetHooks(hooks HookMap) {
	// only create a new map if we don't already have one
	if sdts.hooks == nil {
		sdts.hooks = HookMap{}
	}

	// add each hook to the map
	for k, v := range hooks {
		if v == nil {
			// do not add a nil mapping.

			// but if we already have a non-nil mapping and were given a nil
			// one, remove it
			delete(sdts.hooks, k)
		} else {
			sdts.hooks[k] = v
		}
	}
}

// Evaluate executes the entire syntax-directed translation scheme on the given
// parse tree, and returns the requested attributes from the root of the
// generated AnnotatedTree. The parse tree is first annotated to produce an
// AnnotatedTree, then nodes are visited in an order determined by
// dependency graphs created on all attributes defined for the SDTS that have
// qualifying nodes in the APT. When each node is visited, any associated SDD
// bindings are called on it.
//
// This function requires that SetHooks be called with a valid set of
// implementations for all hooks that will be called.
func (sdts *sdtsImpl) Evaluate(tree parse.Tree, attributes ...string) (vals []interface{}, warns []error, err error) {
	// don't check for no hooks being set because it's possible we are going to
	// be handed an empty parse tree, which will fail for other reasons first
	// or perhaps will not fail at all.

	// first get an annotated parse tree
	root := Annotate(tree)
	depGraphs := depGraph(root, sdts)
	var unexpectedBreaks [][4]string

	if len(depGraphs) > 1 {
		// first, eliminate all depGraphs whose head has a noFlow that applies
		// to it.
		updatedDepGraphs := []*directedGraph[depNode]{}
		for i := range depGraphs {
			var isRoot bool
			var hasUnexpectedBreaks bool

			allNodes := depGraphs[i].AllNodes()

			for j := range allNodes {
				node := allNodes[j]

				if len(node.Edges) == 0 {
					// then either it must be root, or it must have a noFlow that matches
					if node.Data.Parent == nil {
						isRoot = true
						break
					}

					// TODO GHI #101: things are wonky for inherited, check those separately,
					// might need to not assume that Parent is the parent of the
					// node for the rule the actual binding was set on. Synthesized should be fine though.
					nodeParentSymbol := node.Data.Parent.Symbol

					// check for parent in NoFlows
					if slices.In(nodeParentSymbol, node.Data.NoFlows) {
						// then this node does not contribute to unexpected breaks
						continue
					}

					parentProdStr := slices.Reduce(node.Data.Parent.Children, "", func(idx int, item *AnnotatedTree, accum string) string {
						return accum + " " + item.Symbol
					})
					parentProdStr = strings.TrimSpace(parentProdStr)

					unexpectedBreaks = append(unexpectedBreaks, [4]string{nodeParentSymbol, parentProdStr, node.Data.Tree.Symbol, node.Data.Dest.Name})

					// otherwise, if we got here, it's not an expected break.
					// no need to check further
					hasUnexpectedBreaks = true
					break
				}
			}

			// if it is the root, we keep it no matter what. otherwise, only
			// consider it if it has unexpected breaks; else theres no reason to
			// even evaluate them so we can just drop that graph.
			if isRoot || hasUnexpectedBreaks {
				updatedDepGraphs = append(updatedDepGraphs, depGraphs[i])
			}
		}

		depGraphs = updatedDepGraphs
	}

	var singleAttrRoot *directedGraph[depNode]
	// if it's *still* more than 1, scan to see if it's only one attrRoot; that is a warning, not an error, unless
	// asked to be.
	if len(depGraphs) > 1 {
		// if exactly one is root with IR, we can just use that.
		var multipleAttrRoots bool

		for i := range depGraphs {
			allNodes := depGraphs[i].AllNodes()
			for j := range allNodes {
				node := allNodes[j]

				// must be at the end of the eval chain, be for the root of the APT, and be for one of the attributes
				// listed.
				if len(node.Edges) == 0 && node.Data.Parent == nil && slices.In(node.Data.Dest.Name, attributes) {
					if singleAttrRoot != nil {
						// this is an error; can't have multiple attr-bearing with roots depgraphs, let later things
						// catch it
						multipleAttrRoots = true
						break
					}
					// note: taking a shortcut here, strictlly speaking we shouldn't consider this "done" until we
					// have found EACH attribute, otherwise there will be an error popping up as each attribute is
					// evaluated later. But this is fine for now, glub.
					singleAttrRoot = depGraphs[i]
					break
				}
			}

			if multipleAttrRoots {
				singleAttrRoot = nil
				break
			}
		}
	}

	// now we deffin8ly have a pro8lem!!!!!!!!
	if len(depGraphs) > 1 {
		// sort unexpectedBreaks
		unexpectedBreaks = slices.SortBy(unexpectedBreaks, func(left, right [4]string) bool {
			if left[0] == right[0] {
				if left[1] == right[1] {
					if left[2] == right[2] {
						return left[3] < right[3]
					} else {
						return left[2] < right[2]
					}
				} else {
					return left[1] < right[1]
				}
			} else {
				return left[0] < right[0]
			}
		})

		// and only keep distinct items
		unexpectedBreaks = slices.Distinct(unexpectedBreaks, func(b1, b2 [4]string) bool {
			return b1[0] == b2[0] && b1[1] == b2[1] && b1[2] == b2[2] && b1[3] == b2[3]
		})

		if singleAttrRoot != nil {
			warns = append(warns, evalError{
				msg:              "applying SDTS to tree results in evaluation dependency graph with undeclared disconnected segments",
				depGraphs:        depGraphs,
				unexpectedBreaks: unexpectedBreaks,
			})
			depGraphs = []*directedGraph[depNode]{singleAttrRoot}
		} else {
			return nil, warns, evalError{
				msg:              "applying SDTS to tree results in evaluation dependency graph with multiple disconnected root segments",
				depGraphs:        depGraphs,
				unexpectedBreaks: unexpectedBreaks,
			}
		}
	}

	visitOrder, err := kahnSort(depGraphs[0])
	if err != nil {
		return nil, warns, evalError{
			msg:       fmt.Sprintf("sorting SDTS dependency graph: %s", err.Error()),
			sortError: true,
		}
	}

	// we now have an annotated tree. tell listeners
	sdts.emitEvent(Event{
		Type:      EventAnnotation,
		ParseTree: &tree,
		Tree:      &root,
	})

	for i := range visitOrder {
		depNode := visitOrder[i].Data

		nodeTree := depNode.Tree
		synthetic := depNode.Synthetic
		treeParent := depNode.Parent

		var invokeOn *AnnotatedTree
		if synthetic {
			invokeOn = nodeTree
		} else {
			invokeOn = treeParent
		}

		nodeRuleHead, nodeRuleProd := nodeTree.Rule()

		bindingsToExec := sdts.bindingsForAttr(nodeRuleHead, nodeRuleProd, depNode.Dest)
		for j := range bindingsToExec {
			binding := bindingsToExec[j]
			value, err := binding.Invoke(invokeOn, sdts.hooks, sdts.emitEvent, &root, &tree)

			if err != nil {
				attrTypeStr := "synthetic"
				if !synthetic {
					attrTypeStr = "inherited"
				}
				if hookErr, ok := err.(hookError); ok {
					hName := hookErr.name
					if hookErr.name == "" {
						hName = "?"
					}

					errMsg := fmt.Sprintf("%s binding %s = %s(", attrTypeStr, binding.Dest.String(), hName)
					for k := range binding.Requirements {
						errMsg += binding.Requirements[k].String()
						if k+1 < len(binding.Requirements) {
							errMsg += ", "
						}
					}
					errMsg += fmt.Sprintf(") for rule %s -> %s", nodeRuleHead, nodeRuleProd)

					if hookErr.name == "" {
						return nil, warns, evalError{
							msg: fmt.Sprintf("%s: no hook set on binding", errMsg),
						}
					} else if hookErr.missingHook {
						return nil, warns, evalError{
							missingHook: hName,
							msg:         fmt.Sprintf("%s: '%s' is not in the provided hooks table", errMsg, hookErr.name),
						}
					} else {
						return nil, warns, evalError{
							failedHook: hName,
							msg:        fmt.Sprintf("%s: %s", errMsg, hookErr.Error()),
						}
					}
				} else {
					return nil, warns, err
				}
			}

			// now actually set the value on the attribute
			nodeTree.Attributes[depNode.Dest.Name] = value
		}
	}

	// gather requested attributes from root
	attrValues := make([]interface{}, len(attributes))
	for i := range attributes {
		val, ok := root.Attributes[attributes[i]]
		if !ok {
			return nil, warns, evalError{
				msg:       fmt.Sprintf("SDTS does not set attribute %q on root node", attributes[i]),
				sortError: true,
			}
		}
		attrValues[i] = val
	}

	return attrValues, warns, nil
}

// BindSynthesizedAttribute adds a binding to the SDTS for a synthesized
// attribute.
func (sdts *sdtsImpl) Bind(head string, prod []string, attrName string, hook string, withArgs []AttrRef) error {
	// sanity checks; can we even call this?
	if hook == "" {
		return fmt.Errorf("cannot bind to empty hook")
	}

	// check args
	argErrs := ""
	for i := range withArgs {
		req := withArgs[i]
		if !req.Rel.ValidFor(head, prod) {
			argErrs += fmt.Sprintf("\n* bound-to-rule does not have a %s", req.Rel.String())
		}
	}
	if len(argErrs) > 0 {
		return fmt.Errorf("bad arguments:%s", argErrs)
	}

	// get storage slice
	bindingsForHead, ok := sdts.bindings[head]
	if !ok {
		bindingsForHead = map[string][]sddBinding{}
	}
	defer func() { sdts.bindings[head] = bindingsForHead }()

	prodStr := strings.Join(prod, " ")
	existingBindings, ok := bindingsForHead[prodStr]
	if !ok {
		existingBindings = make([]sddBinding, 0)
	}
	defer func() { bindingsForHead[prodStr] = existingBindings }()

	// build the binding
	bind := sddBinding{
		Synthesized:         true,
		BoundRuleSymbol:     head,
		BoundRuleProduction: make([]string, len(prod)),
		Requirements:        make([]AttrRef, len(withArgs)),
		Setter:              hook,
		Dest:                AttrRef{Rel: NodeRelation{Type: RelHead}, Name: attrName},
	}

	copy(bind.BoundRuleProduction, prod)
	copy(bind.Requirements, withArgs)
	existingBindings = append(existingBindings, bind)

	// defers will assign back up to map

	return nil
}

// BindI adds a binding to the SDTS for an inherited attribute.
// At this time, inherited attributes are not supported.
func (sdts *sdtsImpl) BindI(head string, prod []string, attrName string, hook string, withArgs []AttrRef, forProd NodeRelation) error {
	// sanity checks; can we even call this?
	if hook == "" {
		return fmt.Errorf("cannot bind to empty hook")
	}

	// check forProd
	if forProd.Type == RelHead {
		return fmt.Errorf("inherited attributes not allowed to be defined on production heads")
	}
	if !forProd.ValidFor(head, prod) {
		return fmt.Errorf("bad target symbol: bound-to-rule does not have a %s", forProd.String())
	}

	// check args
	argErrs := ""
	for i := range withArgs {
		req := withArgs[i]
		if !req.Rel.ValidFor(head, prod) {
			argErrs += fmt.Sprintf("\n* bound-to-rule does not have a %s", req.Rel.String())
		}
	}
	if len(argErrs) > 0 {
		return fmt.Errorf("bad arguments:%s", argErrs)
	}

	// get storage slice
	bindingsForHead, ok := sdts.bindings[head]
	if !ok {
		bindingsForHead = map[string][]sddBinding{}
	}
	defer func() { sdts.bindings[head] = bindingsForHead }()

	prodStr := strings.Join(prod, " ")
	existingBindings, ok := bindingsForHead[prodStr]
	if !ok {
		existingBindings = make([]sddBinding, 0)
	}
	defer func() { bindingsForHead[prodStr] = existingBindings }()

	// build the binding
	bind := sddBinding{
		Synthesized:         true,
		BoundRuleSymbol:     head,
		BoundRuleProduction: make([]string, len(prod)),
		Requirements:        make([]AttrRef, len(withArgs)),
		Setter:              hook,
		Dest:                AttrRef{Rel: forProd, Name: attrName},
	}

	copy(bind.BoundRuleProduction, prod)
	copy(bind.Requirements, withArgs)
	existingBindings = append(existingBindings, bind)

	// defers will assign back up to map

	return nil
}

// SetNoFlow explicitly marks a particular binding as expected to not 'flow' up
// a dependency graph; that is, it will not consider it a warnable issue if
// the value produced by it is not used by another part of the translation
// scheme.
func (sdts *sdtsImpl) SetNoFlow(synth bool, head string, prod []string, attrName string, forProd NodeRelation, which int, ifParent string) error {
	prodStr := strings.Join(prod, " ")

	var attrTypeName string
	if synth {
		attrTypeName = "synthesized"
	} else {
		attrTypeName = "inherited"
		// check forProd
		if forProd.Type == RelHead {
			return fmt.Errorf("inherited attributes can never be defined on production heads")
		}
		if !forProd.ValidFor(head, prod) {
			return fmt.Errorf("(%s -> %s) nodes do not have a %s", head, prodStr, forProd.String())
		}
	}

	// get storage slice
	bindingsForHead, ok := sdts.bindings[head]
	if !ok {
		return fmt.Errorf("no bindings present for head %s", head)
	}

	existingBindings, ok := bindingsForHead[prodStr]
	if !ok {
		return fmt.Errorf("no bindings present for rule (%s -> %s)", head, prodStr)
	}

	// get only the bindings for the attribute we're interested in, and track the
	// original index of it so we can update it later.
	candidateBindings := make([]box.Pair[sddBinding, int], 0)
	for i := range existingBindings {
		bind := existingBindings[i]
		if bind.Dest.Name == attrName {
			candidateBindings = append(candidateBindings, box.PairOf(bind, i))
		}
	}
	if len(candidateBindings) == 0 {
		return fmt.Errorf("rule (%s -> %s) does not have any bindings for attribute %s", head, prodStr, attrName)
	}

	// filter the bindings by synthesized or inherited
	candidateBindings = slices.Filter(candidateBindings, func(item box.Pair[sddBinding, int]) bool {
		return item.First.Synthesized
	})
	if len(candidateBindings) == 0 {
		return fmt.Errorf("rule (%s -> %s) does not have any %s attributes", head, prodStr, attrTypeName)
	}

	// filter the candidates by forProd, if applicable
	if !synth {
		candidateBindings = slices.Filter(candidateBindings, func(item box.Pair[sddBinding, int]) bool {
			return item.First.Dest.Rel == forProd
		})
	}
	if len(candidateBindings) == 0 {
		return fmt.Errorf("rule (%s -> %s) does not have any inherited attributes for attribute %s on %s", head, prodStr, attrName, forProd.String())
	}

	if which < 0 {
		// apply to all synthesized/inherited bindings as appropriate
		for i := range candidateBindings {
			bind := existingBindings[candidateBindings[i].Second]
			bind.NoFlows = append(bind.NoFlows, ifParent)
			existingBindings[candidateBindings[i].Second] = bind
		}
	} else {
		if which >= len(candidateBindings) {
			return fmt.Errorf("rule does not have binding matching criteria with index %d; highest index is %d", which, len(candidateBindings)-1)
		}
		bind := existingBindings[candidateBindings[which].Second]
		bind.NoFlows = append(bind.NoFlows, ifParent)
		existingBindings[candidateBindings[which].Second] = bind
	}

	bindingsForHead[prodStr] = existingBindings
	sdts.bindings[head] = bindingsForHead

	return nil
}

func (sdts *sdtsImpl) emitEvent(e Event) {
	if sdts.listener != nil {
		sdts.listener(e)
	}
}

// bindingsForAttr returns all bindings defined to apply when at a node in a parse
// tree created by the rule production with head as its head symbol and prod
// as its produced symbols, and when setting the attribute referred to by
// dest. They will be returned in the order they were defined.
func (sdts *sdtsImpl) bindingsForAttr(head string, prod []string, attrRef AttrRef) []sddBinding {
	allForRule := sdts.bindingsForRule(head, prod)

	matchingBindings := []sddBinding{}

	for i := range allForRule {
		if allForRule[i].Dest == attrRef {
			matchingBindings = append(matchingBindings, allForRule[i])
		}
	}

	return matchingBindings
}

// bindingsForRule returns all bindings defined to apply when at a node in a parse
// tree created by the rule production with head as its head symbol and prod
// as its produced symbols. They will be returned in the order they were
// defined.
func (sdts *sdtsImpl) bindingsForRule(head string, prod []string) []sddBinding {
	forHead, ok := sdts.bindings[head]
	if !ok {
		return nil
	}

	symStr := strings.Join(prod, " ")
	forProd, ok := forHead[symStr]
	if !ok {
		return nil
	}

	targetBindings := make([]sddBinding, len(forProd))
	copy(targetBindings, forProd)

	return targetBindings
}
