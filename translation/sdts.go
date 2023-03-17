package translation

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/types"
)

// sdts.go contains the implementation of a Syntax-Directed Translation Scheme.
// TODO: update terminology to match SDTS; we use SDD improperly here.

type sdtsImpl struct {
	bindings map[string]map[string][]SDDBinding
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
		return nil, evalError{
			msg:       "applying SDD to tree results in evaluation dependency graph with disconnected segments",
			depGraphs: depGraphs,
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

	forProd, ok := forHead[strings.Join(prod, " ")]
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

// Validate runs the SDTS on a fake parse tree derived from the grammar. The
// given attribute will be attempted to be evaluated on the root node.
//
// It will use fake value producer, if provided, to generate lexemes for
// terminals in the tree; otherwise contrived values will be used.
func (sdts *sdtsImpl) Validate(g grammar.Grammar, attribute string, fakeValProducer ...map[string]func() string) error {
	pts, err := g.DeriveFullTree(fakeValProducer...)
	if err != nil {
		return fmt.Errorf("deriving fake parse tree: %w", err)
	}

	// TODO: use *all* parse trees, not just the first one
	_, err = sdts.Evaluate(pts[0], attribute)

	evalErr, ok := err.(evalError)
	if !ok {
		return err
	}

	// TODO: betta explanation of what happened using the info in the error
	if len(evalErr.depGraphs) > 0 {
		// disconnected depgraph error

		fullMsg := "translation on fake parse tree resulted in disconnected dependency graphs:"

		for i := range evalErr.depGraphs {
			fullMsg += fmt.Sprintf("\n* %s", DepGraphString(evalErr.depGraphs[i]))
		}

		return fmt.Errorf(fullMsg)
	}

	// TODO: betta message for kahn sort error

	return err
}

func NewSDTS() *sdtsImpl {
	impl := sdtsImpl{
		map[string]map[string][]SDDBinding{},
	}
	return &impl
}

// highly populated error struct for examination by validation code and internal
// routines. may make this betta and exported later.
type evalError struct {
	// if this is a disconnected dep graph segments error, this slice will be
	// non-nil and contain the issue nodes.
	depGraphs []*DirectedGraph[DepNode]

	// if this is a sort error, this will be true
	sortError bool

	msg string
}

func (ee evalError) Error() string {
	return ee.msg
}
