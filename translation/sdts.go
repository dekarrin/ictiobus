package translation

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/types"
)

// sdts.go contains the implementation of a Syntax-Directed Translation Scheme.
// TODO: update terminology to match SDTS; we use SDD improperly here.

type sdtsImpl struct {
	hooks    map[string]AttributeSetter
	bindings map[string]map[string][]SDDBinding
}

func (sdts *sdtsImpl) SetHooks(hooks map[string]AttributeSetter) {
	// only create a new map if we don't already have one
	if sdts.hooks == nil {
		sdts.hooks = map[string]AttributeSetter{}
	}

	// add each hook to the map
	for k, v := range hooks {
		sdts.hooks[k] = v
	}
}

func (sdts *sdtsImpl) BindingsFor(head string, prod []string, attrRef AttrRef) []SDDBinding {
	allForRule := sdts.Bindings(head, prod)

	matchingBindings := []SDDBinding{}

	for i := range allForRule {
		if allForRule[i].Dest == attrRef {
			matchingBindings = append(matchingBindings, allForRule[i])
		}
	}

	return matchingBindings
}

func (sdts *sdtsImpl) Evaluate(tree types.ParseTree, attributes ...string) ([]interface{}, error) {
	// first get an annotated parse tree
	root := AddAttributes(tree)
	// TODO: allow the annotated parse tree to be printed for debug output
	depGraphs := DepGraph(root, sdts)

	// TODO: this is actually fine as long as we got exactly ONE with the root
	// node but is probably not intended. we should warn, not error.
	//
	// specifically, also check to see if a disconnected graph in fact has a parent
	// with no SDT bindings and thus no connection to the child.
	if len(depGraphs) > 1 {
		// first, eliminate all depGraphs whose head has a noFlow that applies
		// to it.
		unexpectedBreaks := [][4]string{}
		updatedDepGraphs := []*DirectedGraph[DepNode]{}
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

					// TODO: things are wonky for inherited, check those separately,
					// might need to not assume that Parent is the parent of the
					// node for the rule the actual binding was set on. Synthesized should be fine though.
					nodeParentSymbol := node.Data.Parent.Symbol

					// check for parent in NoFlows
					if slices.In(nodeParentSymbol, node.Data.NoFlows) {
						// then this node does not contribute to unexpected breaks
						continue
					}

					parentProdStr := slices.Reduce(node.Data.Parent.Children, "", func(idx int, item *AnnotatedParseTree, accum string) string {
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

		// if it's *still* more than 1, we have a problem.
		if len(depGraphs) > 1 {
			return nil, evalError{
				msg:              "applying SDD to tree results in evaluation dependency graph with undeclared disconnected segments",
				depGraphs:        depGraphs,
				unexpectedBreaks: unexpectedBreaks,
			}
		}
	}
	visitOrder, err := KahnSort(depGraphs[0])
	if err != nil {
		return nil, evalError{
			msg:       fmt.Sprintf("sorting SDD dependency graph: %s", err.Error()),
			sortError: true,
		}
	}

	for i := range visitOrder {
		depNode := visitOrder[i].Data

		nodeTree := depNode.Tree
		synthetic := depNode.Synthetic
		treeParent := depNode.Parent

		var invokeOn *AnnotatedParseTree
		if synthetic {
			invokeOn = nodeTree
		} else {
			invokeOn = treeParent
		}

		nodeRuleHead, nodeRuleProd := nodeTree.Rule()

		bindingsToExec := sdts.BindingsFor(nodeRuleHead, nodeRuleProd, depNode.Dest)
		for j := range bindingsToExec {
			binding := bindingsToExec[j]
			value := binding.Invoke(invokeOn)

			// now actually set the value on the attribute
			nodeTree.Attributes[depNode.Dest.Name] = value
		}
	}

	// gather requested attributes from root
	attrValues := make([]interface{}, len(attributes))
	for i := range attributes {
		val, ok := root.Attributes[attributes[i]]
		if !ok {
			return nil, evalError{
				msg:       fmt.Sprintf("SDD does not set attribute %q on root node", attributes[i]),
				sortError: true,
			}
		}
		attrValues[i] = val
	}

	return attrValues, nil
}

func (sdts *sdtsImpl) Bindings(head string, prod []string) []SDDBinding {
	forHead, ok := sdts.bindings[head]
	if !ok {
		return nil
	}

	symStr := strings.Join(prod, " ")
	forProd, ok := forHead[symStr]
	if !ok {
		return nil
	}

	targetBindings := make([]SDDBinding, len(forProd))
	copy(targetBindings, forProd)

	return targetBindings
}

func (sdts *sdtsImpl) BindSynthesizedAttribute(head string, prod []string, attrName string, bindFunc AttributeSetter, withArgs []AttrRef) error {
	// sanity checks; can we even call this?
	if bindFunc == nil {
		return fmt.Errorf("cannot bind nil bindFunc")
	}

	// check args
	argErrs := ""
	for i := range withArgs {
		req := withArgs[i]
		if !req.Relation.ValidFor(head, prod) {
			argErrs += fmt.Sprintf("\n* bound-to-rule does not have a %s", req.Relation.String())
		}
	}
	if len(argErrs) > 0 {
		return fmt.Errorf("bad arguments:%s", argErrs)
	}

	// get storage slice
	bindingsForHead, ok := sdts.bindings[head]
	if !ok {
		bindingsForHead = map[string][]SDDBinding{}
	}
	defer func() { sdts.bindings[head] = bindingsForHead }()

	prodStr := strings.Join(prod, " ")
	existingBindings, ok := bindingsForHead[prodStr]
	if !ok {
		existingBindings = make([]SDDBinding, 0)
	}
	defer func() { bindingsForHead[prodStr] = existingBindings }()

	// build the binding
	bind := SDDBinding{
		Synthesized:         true,
		BoundRuleSymbol:     head,
		BoundRuleProduction: make([]string, len(prod)),
		Requirements:        make([]AttrRef, len(withArgs)),
		Setter:              bindFunc,
		Dest:                AttrRef{Relation: NodeRelation{Type: RelHead}, Name: attrName},
	}

	copy(bind.BoundRuleProduction, prod)
	copy(bind.Requirements, withArgs)
	existingBindings = append(existingBindings, bind)

	// defers will assign back up to map

	return nil
}

func (sdts *sdtsImpl) BindInheritedAttribute(head string, prod []string, attrName string, bindFunc AttributeSetter, withArgs []AttrRef, forProd NodeRelation) error {
	// sanity checks; can we even call this?
	if bindFunc == nil {
		return fmt.Errorf("cannot bind nil bindFunc")
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
		if !req.Relation.ValidFor(head, prod) {
			argErrs += fmt.Sprintf("\n* bound-to-rule does not have a %s", req.Relation.String())
		}
	}
	if len(argErrs) > 0 {
		return fmt.Errorf("bad arguments:%s", argErrs)
	}

	// get storage slice
	bindingsForHead, ok := sdts.bindings[head]
	if !ok {
		bindingsForHead = map[string][]SDDBinding{}
	}
	defer func() { sdts.bindings[head] = bindingsForHead }()

	prodStr := strings.Join(prod, " ")
	existingBindings, ok := bindingsForHead[prodStr]
	if !ok {
		existingBindings = make([]SDDBinding, 0)
	}
	defer func() { bindingsForHead[prodStr] = existingBindings }()

	// build the binding
	bind := SDDBinding{
		Synthesized:         true,
		BoundRuleSymbol:     head,
		BoundRuleProduction: make([]string, len(prod)),
		Requirements:        make([]AttrRef, len(withArgs)),
		Setter:              bindFunc,
		Dest:                AttrRef{Relation: forProd, Name: attrName},
	}

	copy(bind.BoundRuleProduction, prod)
	copy(bind.Requirements, withArgs)
	existingBindings = append(existingBindings, bind)

	// defers will assign back up to map

	return nil
}

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
	candidateBindings := make([]box.Pair[SDDBinding, int], 0)
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
	candidateBindings = slices.Filter(candidateBindings, func(item box.Pair[SDDBinding, int]) bool {
		return item.First.Synthesized
	})
	if len(candidateBindings) == 0 {
		return fmt.Errorf("rule (%s -> %s) does not have any %s attributes", head, prodStr, attrTypeName)
	}

	// filter the candidates by forProd, if applicable
	if !synth {
		candidateBindings = slices.Filter(candidateBindings, func(item box.Pair[SDDBinding, int]) bool {
			return item.First.Dest.Relation == forProd
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

func NewSDTS() *sdtsImpl {
	impl := sdtsImpl{
		bindings: map[string]map[string][]SDDBinding{},
	}
	return &impl
}
