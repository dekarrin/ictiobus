package translation

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/stack"
	"github.com/dekarrin/ictiobus/lex"
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
	depGraphs := DepGraph(root, sdts)
	if len(depGraphs) > 1 {
		return nil, fmt.Errorf("applying SDD to tree results in evaluation dependency graph with disconnected segments")
	}
	visitOrder, err := KahnSort(depGraphs[0])
	if err != nil {
		return nil, fmt.Errorf("sorting SDD dependency graph: %w", err)
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
			return nil, fmt.Errorf("SDD does not set attribute %q on root node", attributes[i])
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

func (sdts *sdtsImpl) Validate(g grammar.Grammar, fakeValProducer ...map[string]func() string) error {
	// validate the grammar so we know that any non-terminal we get from it is
	// valid and produced within the grammar, and anything we get that is not
	// a non-terminal is a terminal used within the grammar.
	err := g.Validate()
	if err != nil {
		return fmt.Errorf("grammar is not valid: %w", err)
	}

	// create a function to get a value for any terminal from the
	// fakeValProducer, falling back on default behavior if none is provided or
	// if a token class is not found in the fakeValProducer.
	makeTermValue := func(class types.TokenClass) string {
		if len(fakeValProducer) > 0 {
			if fvp, ok := fakeValProducer[0][class.ID()]; ok {
				return fvp()
			}
		}
		return fmt.Sprintf("<SIMULATED %s>", class.ID())
	}

	root := &types.ParseTree{}
	treeStack := stack.Stack[*types.ParseTree]{Of: []*types.ParseTree{root}}
	sym := g.StartSymbol()
	var lineNo int

	for {
		pt := treeStack.Pop()

		isTerm := g.Rule(sym).NonTerminal == ""
		pt.Value = sym
		if isTerm {
			termClass := g.Term(sym)
			val := makeTermValue(termClass)
			pt.Terminal = true
			pt.Source = lex.NewToken(termClass, val, 11, lineNo, fmt.Sprintf("<fakeLine>%s</fakeLine>", val))
			pt.Children = nil
			lineNo++
		} else {
			pt.Terminal = false
			pt.Source = nil
			pt.Children = make([]*types.ParseTree, 0)

			// need to make sure we add the children to the parse tree.
			r := g.Rule(sym)

			// what are the possible alternations?
			for i := range r.Productions {
				alt := r.Productions[i]

			}
		}
	}

	return nil
}

func NewSDTS() *sdtsImpl {
	impl := sdtsImpl{
		map[string]map[string][]SDDBinding{},
	}
	return &impl
}
