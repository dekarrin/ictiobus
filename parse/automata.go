package parse

import (
	"fmt"
	"sort"

	"github.com/dekarrin/ictiobus/automaton"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
)

// constructDFAForLALR1 creates a new DFA whose states are made up of
// the sets of items used in an LALR(1) parser. The grammar of the language that is
// accepted by the parser, g, must be LALR(1) and it must be non-augmented.
// Returns an error if g is not LALR(1).
func constructDFAForLALR1(g grammar.Grammar) (automaton.DFA[box.SVSet[grammar.LR1Item]], error) {
	mergeFunc := func(x1, x2 box.SVSet[grammar.LR1Item]) bool {
		return grammar.CoreSet(x1).Equal(grammar.CoreSet(x2))
	}

	reduceFunc := func(x1, x2 box.SVSet[grammar.LR1Item]) box.SVSet[grammar.LR1Item] {
		if x1 == nil {
			return box.NewSVSet(x2)
		}
		x1.AddAll(x2)
		return x1
	}

	nameFunc := func(x1 box.SVSet[grammar.LR1Item]) string {
		return x1.StringOrdered()
	}

	lr1Dfa := constructDFAForCLR1(g)

	// get an NFA so we can start fixing things
	lalrNfa := automaton.DFAToNFA(lr1Dfa)
	lalrNfa.MergeStatesByValue(mergeFunc, reduceFunc, nameFunc)
	lalrDfa, err := automaton.DeterministicNFAToDFA(lalrNfa)
	if err != nil {
		return automaton.DFA[box.SVSet[grammar.LR1Item]]{}, fmt.Errorf("grammar is not LALR(1); resulted in inconsistent state merges")
	}

	return lalrDfa, nil
}

// constructDFAForCLR1 creates a new DFA whose states are made up of the sets
// of items used in a canonical LR(1) parser. The grammar of the language that
// is accepted by the parser, g, must be LR(1) and it must be non-augmented.
func constructDFAForCLR1(g grammar.Grammar) automaton.DFA[box.SVSet[grammar.LR1Item]] {
	oldStart := g.StartSymbol()
	g = g.Augmented()

	initialItem := grammar.LR1Item{
		LR0Item: grammar.LR0Item{
			NonTerminal: g.StartSymbol(),
			Right:       []string{oldStart},
		},
		Lookahead: "$",
	}

	type transInfo struct {
		input string
		next  string
	}

	startSet := lr1CLOSURE(g, box.SVSet[grammar.LR1Item]{initialItem.String(): initialItem})

	stateSets := box.NewSVSet[box.SVSet[grammar.LR1Item]]()
	stateSets.Set(startSet.StringOrdered(), startSet)
	transitions := map[string]map[string]transInfo{}

	// following algo from http://www.cs.ecu.edu/karl/5220/spr16/Notes/Bottom-up/lr1.html
	updates := true
	for updates {
		updates = false

		// suppose that state q contains set I of LR(1) items
		for _, I := range stateSets {

			for _, item := range I {
				if len(item.Right) == 0 || item.Right[0] == grammar.Epsilon[0] {
					continue // no epsilons, deterministic finite state
				}
				// For each symbol s (either a token or a nonterminal) that
				// immediately follows a dot in an LR(1) item [A → α ⋅ sβ, t] in
				// set I...
				s := item.Right[0]

				// ...let Is be the set of all LR(1) items in I where s
				// immediately follows the dot.
				Is := box.NewSVSet[grammar.LR1Item]()
				for _, checkItem := range I {
					if len(checkItem.Right) >= 1 && checkItem.Right[0] == s {
						newItem := checkItem.Copy()

						// Move the dot to the other side of s in each of them.
						newItem.Left = append(newItem.Left, s)
						newItem.Right = make([]string, len(checkItem.Right)-1)
						copy(newItem.Right, checkItem.Right[1:])

						Is.Set(newItem.String(), newItem)
					}
				}

				// That set [Is] becomes the kernel of state q', and you make a
				// transition from q to q′ on s. As usual, form the closure of
				// the set of LR(1) items in state q'.
				newSet := lr1CLOSURE(g, Is)

				// add to states if not already in it
				if !stateSets.Has(newSet.StringOrdered()) {
					updates = true
					stateSets.Set(newSet.StringOrdered(), newSet)
				}

				// add to transitions if not already in it
				stateTransitions, ok := transitions[I.StringOrdered()]
				if !ok {
					stateTransitions = map[string]transInfo{}
				}
				trans, ok := stateTransitions[s]
				if !ok {
					trans = transInfo{}
				}
				if trans.next != newSet.StringOrdered() {
					updates = true
					trans.input = s
					trans.next = newSet.StringOrdered()
					stateTransitions[s] = trans
					transitions[I.StringOrdered()] = stateTransitions
				}
			}
		}
	}

	// okay, we've actually pre-calculated all DFA items so we can now add them.
	// might be able to optimize to add on-the-fly during above loop but this is
	// easier for the moment.
	dfa := automaton.DFA[box.SVSet[grammar.LR1Item]]{}

	// add states
	stateElems := stateSets.Elements()
	sort.Strings(stateElems)

	for i := range stateElems {
		sName := stateElems[i]
		state := stateSets.Get(sName)
		dfa.AddState(sName, true)
		dfa.SetValue(sName, state)
	}

	// transitions
	for onState, stateTrans := range transitions {
		for _, t := range stateTrans {
			dfa.AddTransition(onState, t.input, t.next)
		}
	}

	// and start
	dfa.Start = startSet.StringOrdered()

	return dfa
}
