package trans

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
)

// directedGraph is a node in a graph whose edges point in one direction to
// another node. This implementation can carry data in the nodes of the graph.
//
// directedGraph's zero-value can be used directly.
type directedGraph[V any] struct {
	// Edges is a set of references to nodes this node goes to (that it has an
	// edge pointing to).
	Edges []*directedGraph[V]

	// InEdges is a set of back-references to nodes that go to (have an edge
	// that point towards) this node.
	InEdges []*directedGraph[V]

	// Data is the value held at this node of the graph.
	Data V
}

// LinkTo creates an out edge from dg to the other graph, and also adds an
// InEdge leading back to dg from other.
func (dg *directedGraph[V]) LinkTo(other *directedGraph[V]) {
	dg.Edges = append(dg.Edges, other)
	other.InEdges = append(other.InEdges, dg)
}

// LinkFrom creates an out edge from the other graph to dg, and also adds an
// InEdge leading back to other from dg.
func (dg *directedGraph[V]) LinkFrom(other *directedGraph[V]) {
	other.LinkTo(dg)
}

// Copy creates a duplicate of this graph. Note that data is copied by value and
// is *not* deeply copied.
func (dg *directedGraph[V]) Copy() *directedGraph[V] {
	return dg.recursiveCopy(map[*directedGraph[V]]bool{})
}

func (dg *directedGraph[V]) recursiveCopy(visited map[*directedGraph[V]]bool) *directedGraph[V] {
	visited[dg] = true
	dgCopy := &directedGraph[V]{Data: dg.Data}

	// check out edges
	for i := range dg.Edges {
		to := dg.Edges[i]
		if _, alreadyVisited := visited[to]; alreadyVisited {
			continue
		}

		toCopy := to.recursiveCopy(visited)
		dgCopy.LinkTo(toCopy)
	}

	// check back edges
	for i := range dg.InEdges {
		from := dg.InEdges[i]
		if _, alreadyVisited := visited[from]; alreadyVisited {
			continue
		}
		fromCopy := from.recursiveCopy(visited)
		dgCopy.LinkFrom(fromCopy)
	}

	return dgCopy
}

// Contains returns whether the graph that the given node is in contains at any
// point the given node. Note that the contains check will check SPECIFICALLY if
// the given address is contained.
func (dg *directedGraph[V]) Contains(other *directedGraph[V]) bool {
	return dg.any(func(dg *directedGraph[V]) bool {
		return dg == other
	}, map[*directedGraph[V]]bool{})
}

func (dg *directedGraph[V]) forEachNode(action func(dg *directedGraph[V]), visited map[*directedGraph[V]]bool) {
	visited[dg] = true
	action(dg)

	// check out edges
	for i := range dg.Edges {
		to := dg.Edges[i]
		if _, alreadyVisited := visited[to]; alreadyVisited {
			continue
		}
		to.forEachNode(action, visited)
	}

	// check back edges
	for i := range dg.InEdges {
		from := dg.InEdges[i]
		if _, alreadyVisited := visited[from]; alreadyVisited {
			continue
		}
		from.forEachNode(action, visited)
	}
}

// AllNodes returns all nodes in the graph, in no particular order.
func (dg *directedGraph[V]) AllNodes() []*directedGraph[V] {
	gathered := new([]*directedGraph[V])
	*gathered = make([]*directedGraph[V], 0)

	onVisit := func(dg *directedGraph[V]) {
		*gathered = append(*gathered, dg)
	}
	dg.forEachNode(onVisit, map[*directedGraph[V]]bool{})

	return *gathered
}

// kahnSort takes the given graph and constructs a topological ordering for its
// nodes such that every node is placed after all nodes that eventually lead
// into it. Fails immediately if there are any cycles in the graph.
//
// This is an implementation of the algorithm published by Arthur B. Kahn in
// "Topological sorting of large networks" in Communications of the ACM, 5 (11),
// in 1962, glub! 38O
func kahnSort[V any](dg *directedGraph[V]) ([]*directedGraph[V], error) {
	// detect cycles first or we may enter an infinite loop
	if dg.HasCycles() {
		return nil, fmt.Errorf("can't apply kahn's algorithm to a graph with cycles")
	}

	// this algorithm involves modifying the graph, which we absolutely do not
	// intend to do, so make a copy and operate on that instead.
	dg = dg.Copy()

	sortedL := []*directedGraph[V]{}
	noIncomingS := box.NewKeySet[*directedGraph[V]]()

	// find all start nodes
	allNodes := dg.AllNodes()
	for i := range allNodes {
		n := allNodes[i]
		if len(n.InEdges) == 0 {
			noIncomingS.Add(n)
		}
	}

	for !noIncomingS.Empty() {
		var n *directedGraph[V]
		// just need to get literally any value from the set
		for nodeInSet := range noIncomingS {
			n = nodeInSet
			break
		}

		noIncomingS.Remove(n)
		sortedL = append(sortedL, n)

		for i := range n.Edges {
			m := n.Edges[i]
			// remove all edges from n to m (instead of just 'the one' bc we
			// have no way of associated a *particular* edge with the in edge
			// on m side and there COULD be dupes)
			newNEdges := []*directedGraph[V]{}
			newMInEdges := []*directedGraph[V]{}
			for j := range n.Edges {
				if n.Edges[j] != m {
					newNEdges = append(newNEdges, n.Edges[j])
				}
			}
			for j := range m.InEdges {
				if m.InEdges[j] != n {
					newMInEdges = append(newMInEdges, m.InEdges[j])
				}
			}
			n.Edges = newNEdges
			m.InEdges = newMInEdges

			if len(m.InEdges) == 0 {
				noIncomingS.Add(m)
			}
		}
	}

	return sortedL, nil
}

// HasCycles returns whether the graph has a cycle in it at any point.
func (dg *directedGraph[V]) HasCycles() bool {
	// if there are no cycles, there is at least one node with no other
	finished := map[*directedGraph[V]]bool{}
	visited := map[*directedGraph[V]]bool{}

	var dfsCycleCheck func(n *directedGraph[V]) bool
	dfsCycleCheck = func(n *directedGraph[V]) bool {
		_, alreadyFinished := finished[n]
		_, alreadyVisited := visited[n]
		if alreadyFinished {
			return false
		}
		if alreadyVisited {
			return true
		}
		visited[n] = true

		for i := range n.Edges {
			cycleFound := dfsCycleCheck(n.Edges[i])
			if cycleFound {
				return true
			}
		}
		finished[n] = true
		return false
	}

	toCheck := dg.AllNodes()
	for i := range toCheck {
		n := toCheck[i]
		if dfsCycleCheck(n) {
			return true
		}
	}

	return false
}

func (dg *directedGraph[V]) any(predicate func(*directedGraph[V]) bool, visited map[*directedGraph[V]]bool) bool {
	visited[dg] = true
	if predicate(dg) {
		return true
	}

	// check out edges
	for i := range dg.Edges {
		to := dg.Edges[i]
		if _, alreadyVisited := visited[to]; alreadyVisited {
			continue
		}
		toMatches := to.any(predicate, visited)
		if toMatches {
			return true
		}
	}

	// check back edges
	for i := range dg.InEdges {
		from := dg.InEdges[i]
		if _, alreadyVisited := visited[from]; alreadyVisited {
			continue
		}
		fromMatches := from.any(predicate, visited)
		if fromMatches {
			return true
		}
	}

	return false
}

type depNode struct {
	Parent    *AnnotatedTree
	Tree      *AnnotatedTree
	Synthetic bool
	Dest      AttrRef
	NoFlows   []string
}

func depGraphString(dg *directedGraph[depNode]) string {
	nodes := dg.AllNodes()
	var sb strings.Builder

	sb.WriteRune('(')

	nodes = slices.SortBy(nodes, func(left, right *directedGraph[depNode]) bool {
		return left.Data.Tree.ID() < right.Data.Tree.ID()
	})

	for i := range nodes {
		n := nodes[i]
		dep := n.Data
		sym := dep.Tree.Symbol

		prd := ""
		for j := range dep.Tree.Children {
			prd += dep.Tree.Children[j].Symbol
			if j+1 < len(dep.Tree.Children) {
				prd += " "
			}
		}

		if prd != "" {
			prd = " -> [" + prd + "]"
		}

		nodeID := dep.Tree.ID()
		nextIDs := []aptNodeID{}

		for j := range n.Edges {
			nextIDs = append(nextIDs, n.Edges[j].Data.Tree.ID())
		}

		var nodeStart string
		if len(nodes) > 1 {
			nodeStart = "\n\t"
		}

		sb.WriteString(fmt.Sprintf("%s(%v: %s%s, <%s>", nodeStart, nodeID, sym, prd, dep.Dest))

		if len(nextIDs) > 0 {
			sb.WriteString(" -> {")
			for j := range nextIDs {
				sb.WriteString(fmt.Sprintf("%v", nextIDs[j]))
				if j+1 < len(nextIDs) {
					sb.WriteString(", ")
				}
			}
			sb.WriteRune('}')
		}

		sb.WriteRune(')')
		if i+1 < len(nodes) {
			sb.WriteRune(',')
		}
	}

	if len(nodes) > 1 {
		sb.WriteRune('\n')
	}

	sb.WriteRune(')')
	return sb.String()
}

// Info on this func from algorithm 5.2.1 of the purple dragon book.
//
// Returns one node from each of the connected sub-graphs of the dependency
// tree. If the entire dependency graph is connected, there will be only 1 item
// in the returned slice.
func depGraph(aptRoot AnnotatedTree, sdts *sdtsImpl) []*directedGraph[depNode] {
	type treeAndParent struct {
		Tree   *AnnotatedTree
		Parent *AnnotatedTree
	}
	// no parent set on first node; it's the root
	treeStack := box.NewStack([]treeAndParent{{Tree: &aptRoot}})

	depNodes := map[aptNodeID]map[string]*directedGraph[depNode]{}

	for treeStack.Len() > 0 {
		curTreeAndParent := treeStack.Pop()
		curTree := curTreeAndParent.Tree
		curParent := curTreeAndParent.Parent

		// what semantic rule would apply to this?
		ruleHead, ruleProd := curTree.Rule()
		binds := sdts.bindingsForRule(ruleHead, ruleProd)

		// sanity check each node on visit to be shore it's got a non-empty ID.
		if curTree.ID() == aptIDZero {
			panic("ID not set on APT node")
		}

		for i := range binds {
			binding := binds[i]
			if len(binding.Requirements) < 1 {
				// we still need to add the binding as a target node so it can
				// be found by other dep nodes

				targetNode, ok := curTree.RelativeNode(binding.Dest.Rel)
				if !ok {
					panic(fmt.Sprintf("relative address cannot be followed: %v", binding.Dest.Rel.String()))
				}
				targetNodeID := targetNode.ID()
				targetNodeDepNodes, ok := depNodes[targetNodeID]
				if !ok {
					targetNodeDepNodes = map[string]*directedGraph[depNode]{}
				}
				targetParent := curParent
				synthTarget := true
				if targetNode.ID() != curTree.ID() {
					// then targetNode MUST be a child of curTreeNode
					targetParent = curTree

					// additionally, it cannot be synthetic because it is not
					// being set at the head of a production
					synthTarget = false
				}

				// specifically, need to address the one for the desired attribute
				toDepNode, ok := targetNodeDepNodes[binding.Dest.Name]
				if !ok {
					toDepNode = &directedGraph[depNode]{Data: depNode{
						Parent:    targetParent,
						Tree:      targetNode,
						Dest:      binding.Dest,
						Synthetic: synthTarget,
						NoFlows:   make([]string, len(binding.NoFlows)),
					}}
					copy(toDepNode.Data.NoFlows, binding.NoFlows)
				}
				// but also, if it DOES already exist we might have created it
				// without knowing whether it is a synthetic attr; either way,
				// check it now
				toDepNode.Data.Synthetic = synthTarget
				toDepNode.Data.Dest = binding.Dest

				targetNodeDepNodes[binding.Dest.Name] = toDepNode
				depNodes[targetNodeID] = targetNodeDepNodes
			}
			for j := range binding.Requirements {
				req := binding.Requirements[j]

				// get the TARGET node (the one whose attr is being set)
				targetNode, ok := curTree.RelativeNode(binding.Dest.Rel)
				if !ok {
					panic(fmt.Sprintf("relative address cannot be followed: %v", binding.Dest.Rel.String()))
				}
				targetNodeID := targetNode.ID()
				targetNodeDepNodes, ok := depNodes[targetNodeID]
				if !ok {
					targetNodeDepNodes = map[string]*directedGraph[depNode]{}
				}
				targetParent := curParent
				synthTarget := true
				if targetNode.ID() != curTree.ID() {
					// then targetNode MUST be a child of curTreeNode
					targetParent = curTree

					// additionally, it cannot be synthetic because it is not
					// being set at the head of a production
					synthTarget = false
				}
				// specifically, need to address the one for the desired attribute
				toDepNode, ok := targetNodeDepNodes[binding.Dest.Name]
				if !ok {
					toDepNode = &directedGraph[depNode]{Data: depNode{
						Parent:    targetParent,
						Tree:      targetNode,
						Dest:      binding.Dest,
						Synthetic: synthTarget,
						NoFlows:   make([]string, len(binding.NoFlows)),
					}}
					copy(toDepNode.Data.NoFlows, binding.NoFlows)
				}
				// but also, if it DOES already exist we might have created it
				// without knowing whether it is a synthetic attr; either way,
				// check it now
				toDepNode.Data.Synthetic = synthTarget
				toDepNode.Data.Dest = binding.Dest

				// get the RELATED node (the one whose attr is used as an argument):
				relNode, ok := curTree.RelativeNode(req.Rel)
				if !ok {
					panic(fmt.Sprintf("relative address cannot be followed: %v", req.Rel.String()))
				}
				relNodeID := relNode.ID()
				relNodeDepNodes, ok := depNodes[relNodeID]
				if !ok {
					relNodeDepNodes = map[string]*directedGraph[depNode]{}
				}
				// specifically, need to address the one for the desired attribute
				fromDepNode, ok := relNodeDepNodes[req.Name]
				if !ok {
					relParent := curParent
					if relNode != curTree {
						// then relNode MUST be a child of curTreeNode
						relParent = curTree
					}
					fromDepNode = &directedGraph[depNode]{Data: depNode{
						// we simply have no idea whether this is a synthetic
						// attribute or not at this time
						Parent: relParent,
						Tree:   relNode,
					}}

					// If the target node is requesting a built-in attribute,
					// nothing will ever come by later to set syntheticness and
					// dest, so set it now. Built-in attribute nodes are always
					// considered synthetic, even when/if inherited attributes
					// are enabled for reel.
					if strings.HasPrefix(req.Name, "$") {
						fromDepNode.Data.Synthetic = true
						fromDepNode.Data.Dest = AttrRef{Rel: NRHead(), Name: req.Name}
					}
				}

				// create the edge; this will modify BOTH dep nodes
				fromDepNode.LinkTo(toDepNode)

				// make shore to assign after modification (shouldn't NEED
				// to due to attrDepNode being ptr-to but do it just to be
				// safe)
				relNodeDepNodes[req.Name] = fromDepNode
				targetNodeDepNodes[binding.Dest.Name] = toDepNode
				depNodes[relNodeID] = relNodeDepNodes
				depNodes[targetNodeID] = targetNodeDepNodes
			}
		}

		// put child nodes on stack in reverse order to get left-first
		for i := len(curTree.Children) - 1; i >= 0; i-- {
			treeStack.Push(treeAndParent{Parent: curTree, Tree: curTree.Children[i]})
		}
	}

	var connectedSubGraphs []*directedGraph[depNode]

	for k := range depNodes {
		idDepNodes := depNodes[k]
		for attrRef := range idDepNodes {
			node := idDepNodes[attrRef]
			var alreadyHaveGraph bool
			if len(node.Edges) > 0 || len(node.InEdges) > 0 {
				// we found a non-empty node, need to check if it's already
				// added

				// first, is this already in a graph we've grabbed? no need to
				// keep it if so
				for i := range connectedSubGraphs {
					prevSub := connectedSubGraphs[i]
					if prevSub.Contains(node) {
						alreadyHaveGraph = true
						break
					}
				}
			}
			if !alreadyHaveGraph {
				connectedSubGraphs = append(connectedSubGraphs, node)
			}
		}
	}

	return connectedSubGraphs
}
