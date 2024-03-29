package automaton

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/internal/textfmt"
)

// NFA is a non-deterministic finite automaton. It holds 'values' within its
// states; supplementary information associated with each state that is not
// required for the NFA to function.
type NFA[E any] struct {
	order  uint64
	states map[string]nfaState[E]

	// Start is the starting state of the NFA.
	Start string
}

// nfaTransitionTo holds a transition from one state to another in an NFA. It
// holds info on the state the transition is from, the input it transitions on,
// and its priority within the transition table of the state it comes from.
// Priority matters because in an NFA, there could be two different transitions
// using the same input.
type nfaTransitionTo struct {
	from  string
	input string
	index int
}

// AcceptingStates returns the set of all states that are accepting. The order
// of items in the returned slice is guaranteed to be stable.
func (nfa NFA[E]) AcceptingStates() []string {
	accepting := []string{}
	allStates := nfa.States()
	for i := range allStates {
		if nfa.states[allStates[i]].accepting {
			accepting = append(accepting, allStates[i])
		}
	}
	sort.Strings(accepting)

	return accepting
}

// allTransitionsTo gets all transitions to the given state.
func (nfa NFA[E]) allTransitionsTo(toState string) []nfaTransitionTo {
	if _, ok := nfa.states[toState]; !ok {
		// Gr8! We are done.
		return []nfaTransitionTo{}
	}

	transitions := []nfaTransitionTo{}

	s := nfa.States()

	for _, sName := range s {
		state := nfa.states[sName]
		for k := range state.transitions {
			for i := range state.transitions[k] {
				if state.transitions[k][i].next == toState {
					trans := nfaTransitionTo{
						from:  sName,
						input: k,
						index: i,
					}
					transitions = append(transitions, trans)
				}
			}
		}
	}

	return transitions
}

// Copy returns a deeply-copied duplicate of this NFA.
func (nfa NFA[E]) Copy() NFA[E] {
	copied := NFA[E]{
		Start:  nfa.Start,
		states: make(map[string]nfaState[E]),
		order:  nfa.order,
	}

	for k := range nfa.states {
		copied.states[k] = nfa.states[k].Copy()
	}

	return copied
}

// States returns all names of states in the NFA. Ordering of the states in the
// returned slice is guaranteed to be stable.
func (nfa NFA[E]) States() []string {
	states := []string{}

	for k := range nfa.states {
		states = append(states, k)
	}

	sort.Strings(states)

	return states
}

// NFAToDFA converts the NFA into a deterministic finite automaton accepting the
// same strings. The function reduceFn is called for all state values being
// combined; the first time it is called for a state, the reduced argument will
// be the zero-value of E2, and each subsequent time it is called for the same
// state, reduced will be the prior value that was returned from reduceFn.
//
// This is an implementation of algorithm 3.20 from the purple dragon book.
func NFAToDFA[E1, E2 any](nfa NFA[E1], reduceFn func(reduced E2, next E1) E2) DFA[E2] {
	inputSymbols := nfa.InputSymbols()

	Dstart := nfa.epsilonClosure(nfa.Start)

	markedStates := box.NewStringSet()
	Dstates := map[string]box.StringSet{}
	Dstates[Dstart.StringOrdered()] = Dstart

	// these are Dstates but represented in actual format for placement into
	// our implement8ion of DFAs, which is also where transition function info
	// and acceptance info is stored.
	dfa := DFA[E2]{
		states: map[string]dfaState[E2]{},
	}

	// initially, ε-closure(s₀) is the only state in Dstates, and it is unmarked
	for {
		// get unmarked states in Dstates
		DstateNames := box.StringSetOf(textfmt.OrderedKeys(Dstates))
		unmarkedStates := DstateNames.Difference(markedStates)

		if unmarkedStates.Len() < 1 {
			break
		}

		// make the conversion deterministic so output can be easier to examine.
		unmarkedStatesElements := unmarkedStates.Elements()
		sort.Strings(unmarkedStatesElements)

		// while ( there is an unmarked state T in Dstates )
		for _, Tname := range unmarkedStatesElements {
			T := Dstates[Tname]

			// mark T
			markedStates.Add(Tname)

			// (need to reduce the value of every item to get a set of them)
			var combinedValue E2
			for nfaStateName := range T {
				val := nfa.GetValue(nfaStateName)
				combinedValue = reduceFn(combinedValue, val)
			}

			newDFAState := dfaState[E2]{name: Tname, value: combinedValue, transitions: map[string]faTransition{}}

			if T.Any(func(v string) bool {
				return nfa.states[v].accepting
			}) {
				newDFAState.accepting = true
			}

			// for ( each input symbol a )
			for _, a := range inputSymbols {
				// (but like, glub, not the epsilon symbol itself)
				if a == "" {
					continue
				}

				U := nfa.epsilonClosureOfSet(nfa.moveSet(T, a))

				// if its not a symbol that the state can transition on, U will
				// be empty, skip it
				if U.Empty() {
					continue
				}

				// if U is not in Dstates
				if !DstateNames.Has(U.StringOrdered()) {
					// add U as an unmarked state to Dstates
					DstateNames.Add(U.StringOrdered())
					Dstates[U.StringOrdered()] = U
				}

				// Dtran[T, a] = U
				newDFAState.transitions[a] = faTransition{input: a, next: U.StringOrdered()}
			}

			// add it to our working DFA states as well
			newDFAState.ordering = dfa.order
			dfa.order++

			dfa.states[Tname] = newDFAState

			if dfa.Start == "" {
				// then T is our starting state.
				dfa.Start = Tname
			}
		}

	}
	return dfa
}

// InputSymbols returns the set of all input symbols processed by some
// transition in the NFA. The ordering of items in the returned slice is
// guaranteed to be stable.
func (nfa NFA[E]) InputSymbols() []string {
	symbols := box.NewStringSet()
	for sName := range nfa.states {
		st := nfa.states[sName]

		for a := range st.transitions {
			symbols.Add(a)
		}
	}

	symbolsSlice := symbols.Elements()
	sort.Strings(symbolsSlice)

	return symbolsSlice
}

// moveSet returns the set of states reachable with one transition from some state
// in X on input a. Purple dragon book calls this function moveSet(T, a) and it is
// on page 153 as part of algorithm 3.20.
func (nfa NFA[E]) moveSet(X box.Set[string], a string) box.StringSet {
	moves := box.NewStringSet()

	for _, s := range X.Elements() {
		stateItem, ok := nfa.states[s]
		if !ok {
			continue
		}

		transitions := stateItem.transitions[a]

		for _, t := range transitions {
			moves.Add(t.next)
		}
	}

	return moves
}

// MergeStatesByValue performs a merge of states, using the mergeIfValuesCond
// function to determine when to merge based on their values and the reduceFunc
// to combine them together, and nameFn for the name of the new state.
//
// reduceFn is called to merge two items together. It will be called in order of
// elements encountered, and the first call will always have the merged be the
// zero value of the NFA's value-type.
//
// nameFn is called with the fully-merged value to get the name for the state.
func (nfa *NFA[E]) MergeStatesByValue(mergeCondFn func(x1, x2 E) bool, reduceFn func(merged, next E) E, nameFn func(E) string) {
	// counter for unique state name
	newStateNum := 0

	// now start merging states
	updated := true
	for updated {
		updated = false

		alreadyMerged := box.NewStringSet()
		orderedStateElements := nfa.States()
		stateVals := map[string]E{}
		for _, name := range orderedStateElements {
			stateVals[name] = nfa.GetValue(name)
		}

		for _, stateName := range orderedStateElements {
			if alreadyMerged.Has(stateName) {
				continue
			}

			mergeWith := []string{}
			firstElement := stateVals[stateName]

			// need to find ALL to merge w or this is gonna get wild REEL quick
			for _, otherStateName := range orderedStateElements {
				if stateName == otherStateName {
					continue
				}

				otherElement := stateVals[otherStateName]

				// Note: we do NOT enforce an ordering in general on which
				// states are merged first. this could cause issues; doing them
				// in an arbitrary order

				// check their cores
				if mergeCondFn(firstElement, otherElement) {
					mergeWith = append(mergeWith, otherStateName)
				}
			}

			// now we merge any that have been queued to do so
			if len(mergeWith) > 0 {
				updated = true
				alreadyMerged.Add(stateName)
				destState := nfa.states[stateName]
				var zeroVal E
				mergedVal := reduceFn(zeroVal, firstElement)

				for i := range mergeWith {
					alreadyMerged.Add(mergeWith[i])

					mergedVal = reduceFn(mergedVal, stateVals[mergeWith[i]])

				}

				// We COULD tell what new name of state would be NOW, but to keep
				// things from overlapping during the process we will be setting
				// to a unique number and updating after all merges are complete
				// (at which point there should be 0 conflicting state names).
				newStateName := fmt.Sprintf("%d", newStateNum)
				newStateNum++
				destState.name = nameFn(mergedVal)
				destState.value = mergedVal

				// and so we can rewrite transitions from the old states to the
				// new one
				for i := range mergeWith {
					transitionsToMerged := nfa.allTransitionsTo(mergeWith[i])

					for j := range transitionsToMerged {
						trans := transitionsToMerged[j]
						from := trans.from
						sym := trans.input
						idx := trans.index

						// rewrite the transition to new state
						nfa.states[from].transitions[sym][idx] = faTransition{input: sym, next: newStateName}
					}

					// also, check to see if we need to update start
					if nfa.Start == mergeWith[i] {
						nfa.Start = newStateName
					}
				}

				// also rewrite any transitions to the merged-to state
				transitionsToDestState := nfa.allTransitionsTo(stateName)
				for j := range transitionsToDestState {
					trans := transitionsToDestState[j]
					from := trans.from
					sym := trans.input
					idx := trans.index

					// rewrite the transition to new state
					nfa.states[from].transitions[sym][idx] = faTransition{input: sym, next: newStateName}
				}

				// also, check to see if we need to update start
				if nfa.Start == stateName {
					nfa.Start = newStateName
				}

				// finally, enshore that any transitions we lose by deleting the
				// old state are added to the new state. this SHOULD collapse to
				// a single state by the time that things are done if it is
				// indeed an LALR(1) grammar
				for i := range mergeWith {
					lostTransitions := nfa.states[mergeWith[i]].transitions
					for sym := range lostTransitions {
						transForSym := lostTransitions[sym]
						destTransForSym, ok := destState.transitions[sym]
						if !ok {
							destTransForSym = []faTransition{}
						}

						for j := range transForSym {
							// is this already in the dest? don't add it if so
							faTrans := transForSym[j]

							inDestTrans := false
							for k := range destTransForSym {
								destFATrans := destTransForSym[k]
								if destFATrans == faTrans {
									inDestTrans = true
									break
								}
							}
							if !inDestTrans {
								destTransForSym = append(destTransForSym, faTrans)
							}
						}
						destState.transitions[sym] = destTransForSym
					}
				}

				// with those updated, we can now delete the old states from
				// the DFA
				for i := range mergeWith {
					delete(nfa.states, mergeWith[i])
				}

				// unshore if this condition is proven not to happen, either
				// way it's 8AD so checking
				if _, ok := nfa.states[newStateName]; ok {
					panic(fmt.Sprintf("merged state name conflicts w state %q already in DFA", newStateName))
				}

				// enshore the updated new state is stored...
				nfa.states[newStateName] = destState

				// ...and, finally, remove the old version of it
				delete(nfa.states, stateName)
			}

			// did we just update? if so, all of the pre-cached info on states
			// and names and such is invalid due to modifying the DFA, and
			// therefore must be regenerated before checking anyfin else.
			//
			// they will be auto-regenerated by the parent loop
			if updated {
				break
			}
		}
	}

	// prior to conversion to dfa, go through and update the auto-numbered states
	nfaStates := nfa.States()
	for _, stateName := range nfaStates {
		st := nfa.states[stateName]

		// we keep the name pre-calculated in .name, so check if there's a mismatch
		if st.name != stateName {
			newStateName := st.name
			transitionsToMerged := nfa.allTransitionsTo(stateName)

			for j := range transitionsToMerged {
				trans := transitionsToMerged[j]
				from := trans.from
				sym := trans.input
				idx := trans.index

				// rewrite the transition to new state
				nfa.states[from].transitions[sym][idx] = faTransition{input: sym, next: newStateName}
			}

			// also, check to see if we need to update start
			if nfa.Start == stateName {
				nfa.Start = newStateName
			}

			// and now, swap the name for the reel one
			nfa.states[newStateName] = st
			delete(nfa.states, stateName)
		}
	}
}

// DeterministicNFAToDFA creates a DFA from an NFA by copying all of its states
// and transitions exactly as they are. This performs no merges and assumes the
// given NFA[E] is already de-facto deterministic. It will return an error if
// the NFA contains any epsilon transitions or any transitions from the same
// state to two different states based on the same input symbol.
func DeterministicNFAToDFA[E any](nfa NFA[E]) (DFA[E], error) {
	dfa := DFA[E]{
		Start:  nfa.Start,
		states: map[string]dfaState[E]{},
	}

	nfaNames := textfmt.OrderedKeys(nfa.states)

	for _, sName := range nfaNames {
		nState := nfa.states[sName]

		dState := dfaState[E]{
			ordering:    nState.ordering,
			name:        nState.name,
			value:       nState.value,
			transitions: map[string]faTransition{},
			accepting:   nState.accepting,
		}

		for sym := range nState.transitions {
			nTransList := nState.transitions[sym]

			goesTo := ""
			for i := range nTransList {
				if nTransList[i].next == "" {
					return DFA[E]{}, fmt.Errorf("state %q has empty transition-to for %q", nState.name, sym)
				}
				if goesTo == "" {
					// first time we are seeing this, set it now
					goesTo = nTransList[i].next
					dState.transitions[sym] = faTransition{
						input: sym,
						next:  nTransList[i].next,
					}
				} else {
					// if there's more transitions, they simply need to go to the
					// same place.
					if nTransList[i].next != goesTo {
						return DFA[E]{}, fmt.Errorf("state %q has non-deterministic transition for symbol %q", nState.name, sym)
					}
				}
			}
		}

		dfa.states[sName] = dState
	}

	return dfa, nil
}

// epsilonClosureOfSet gives the set of states reachable from some state in
// X using one or more ε-moves.
func (nfa NFA[E]) epsilonClosureOfSet(X box.Set[string]) box.StringSet {
	allClosures := box.NewStringSet()

	for _, s := range X.Elements() {
		closures := nfa.epsilonClosure(s)
		allClosures.AddAll(closures)
	}

	return allClosures
}

// epsilonClosure gives the set of states reachable from state using one or more
// ε-moves.
func (nfa NFA[E]) epsilonClosure(s string) box.StringSet {
	stateItem, ok := nfa.states[s]
	if !ok {
		return nil
	}

	closure := box.NewStringSet()
	checkingStates := &box.Stack[nfaState[E]]{}
	checkingStates.Push(stateItem)

	for checkingStates.Len() > 0 {
		checking := checkingStates.Pop()

		if closure.Has(checking.name) {
			// we've already checked it. skip.
			continue
		}

		// add it to the closure and then check it for recursive closures
		closure.Add(checking.name)

		epsilonMoves, hasEpsilons := checking.transitions[""]
		if !hasEpsilons {
			continue
		}

		for _, move := range epsilonMoves {
			stateName := move.next
			state, ok := nfa.states[stateName]
			if !ok {
				// should never happen unless someone manually adds to
				// unexported properties; AddTransition ensures that only valid
				// and followable transitions are allowed to be added.
				panic(fmt.Sprintf("points to invalid state: %q", stateName))
			}

			checkingStates.Push(state)
		}
	}

	return closure
}

// String returns the string representation of an NFA.
func (nfa NFA[E]) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<START: %q, STATES:", nfa.Start))

	orderedStates := textfmt.OrderedKeys(nfa.states)
	orderedStates = slices.SortBy(orderedStates, func(s1, s2 string) bool {
		n1, err := strconv.Atoi(s1)
		if err != nil {
			// fallback; str comparison
			return s1 < s2
		}

		n2, err := strconv.Atoi(s2)
		if err != nil {
			// fallback; str comparison
			return s1 < s2
		}

		return n1 < n2
	})

	for i := range orderedStates {
		sb.WriteString("\n\t")
		sb.WriteString(nfa.states[orderedStates[i]].String())

		if i+1 < len(nfa.states) {
			sb.WriteRune(',')
		} else {
			sb.WriteRune('\n')
		}
	}

	sb.WriteRune('>')

	return sb.String()
}

// NumberStates renames all states to each have a unique name based on an
// increasing number sequence. The starting state is guaranteed to be numbered
// 0; beyond that, the states are put in the order they were added.
func (nfa *NFA[E]) NumberStates() {
	if _, ok := nfa.states[nfa.Start]; !ok {
		panic("can't number states of NFA with no start state set")
	}
	origStateNames := nfa.States()

	// make shore to pull out starting state and place at front
	startIdx := -1
	for i := range origStateNames {
		if origStateNames[i] == nfa.Start {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		panic("couldn't find starting state; should never happen")
	}

	origStateNames = append(origStateNames[:startIdx], origStateNames[startIdx+1:]...)
	origStateNames = slices.SortBy(origStateNames, func(s1, s2 string) bool {
		return nfa.states[s1].ordering < nfa.states[s2].ordering
	})
	origStateNames = append([]string{nfa.Start}, origStateNames...)

	numMapping := map[string]string{}
	for i := range origStateNames {
		name := origStateNames[i]
		newName := fmt.Sprintf("%d", i)
		numMapping[name] = newName
	}

	// to keep things simple, instead of searching for every instance of each
	// name which is an expensive operation, we'll just build an entirely new
	// NFA using our mapping rules to adjust names as we go, then steal its
	// states map.

	newNfa := NFA[E]{
		states: make(map[string]nfaState[E]),
		Start:  numMapping[nfa.Start],
	}

	// first, add the initial states
	for _, name := range origStateNames {
		st := nfa.states[name]
		newName := numMapping[name]
		newNfa.AddState(newName, st.accepting)

		newSt := newNfa.states[newName]
		newSt.ordering = st.ordering
		newNfa.states[newName] = newSt

		newNfa.SetValue(newName, st.value)

		// transitions come later, need to add all states *first*
	}

	// add initial transitions
	for _, name := range origStateNames {
		st := nfa.states[name]
		from := numMapping[name]

		for sym := range st.transitions {
			symTrans := st.transitions[sym]
			for i := range symTrans {
				t := symTrans[i]
				to := numMapping[t.next]
				newNfa.AddTransition(from, sym, to)
			}
		}
	}

	// oh ya, just gonna go ahead and sneeeeeeeak this on away from ya
	nfa.states = newNfa.states
	nfa.Start = newNfa.Start
}

// Join combines two NFAs into a single one. The argument fromToOther gives the
// method of joining the two NFAs; it is a slice of triples, each of which gives
// a state from the original nfa, the symbol to transition on, and a state in
// the provided NFA to go to on receiving that symbol.
//
// The original NFAs are not modified. The resulting NFA's start state is the
// same as the original NFA's start state.
//
// In order to prevent conflicts, all state names in the resulting NFA will be
// named according to a scheme that namespaces them by which NFA they came from;
// states that came from the original NFA will be changed to be called
// '1:ORIGNAL_NAME' in the resulting NFA, and states that came from the provided
// NFA will be changed to be called '2:ORIGINAL_NAME' in the resulting NFA, with
// 'ORIGINAL_NAME' replaced with the actual original name of the state.
//
// After the resulting NFA is created, all state names listed in addAccept will
// be changed to accepting states in the resulting NFA. Likewise, all state
// names listed in removeAccept will be changed to no longer be accepting in the
// resulting DFA.
//
// Note that because addAccept and removeAccept are applied to the resulting NFA
// after creation, they must use the state-naming convention mentioned above,
// while states mentioned in fromToOther should use the original names of the
// states.
func (nfa NFA[E]) Join(other NFA[E], fromToOther [][3]string, otherToFrom [][3]string, addAccept []string, removeAccept []string) (NFA[E], error) {
	if len(fromToOther) < 1 {
		return NFA[E]{}, fmt.Errorf("need to provide at least one mapping in fromToOther")
	}

	joined := NFA[E]{
		states: make(map[string]nfaState[E]),
		Start:  "1:" + nfa.Start,
	}

	addAcceptSet := box.StringSetOf(addAccept)
	removeAcceptSet := box.StringSetOf(removeAccept)

	nfaStateNames := joined.States()

	// first, add the initial states
	for _, stateName := range nfaStateNames {
		st := nfa.states[stateName]
		newName := "1:" + stateName

		accept := st.accepting
		if addAcceptSet.Has(newName) {
			accept = true
		} else if removeAcceptSet.Has(newName) {
			accept = false
		}
		joined.AddState(newName, accept)
		joined.SetValue(newName, st.value)

		// transitions come later, need to add all states *first*
	}

	// add initial transitions
	for _, stateName := range nfaStateNames {
		st := nfa.states[stateName]
		from := "1:" + stateName

		for sym := range st.transitions {
			symTrans := st.transitions[sym]
			for i := range symTrans {
				t := symTrans[i]
				to := "1:" + t.next
				joined.AddTransition(from, sym, to)
			}
		}
	}

	// next, do the same for the second NFA
	otherStateNames := other.States()

	for _, stateName := range otherStateNames {
		st := other.states[stateName]
		newName := "2:" + stateName

		accept := st.accepting
		if addAcceptSet.Has(newName) {
			accept = true
		} else if removeAcceptSet.Has(newName) {
			accept = false
		}
		joined.AddState(newName, accept)
		joined.SetValue(newName, st.value)

		// transitions come later, need to add all states *first*
	}

	// add other transitions
	for _, stateName := range otherStateNames {
		st := other.states[stateName]
		from := "2:" + stateName

		for sym := range st.transitions {
			symTrans := st.transitions[sym]
			for i := range symTrans {
				t := symTrans[i]
				to := "2:" + t.next
				joined.AddTransition(from, sym, to)
			}
		}
	}

	// already did accept adjustment on the fly, now it's time to link the
	// states together
	for i := range fromToOther {
		link := fromToOther[i]
		from := "1:" + link[0]
		sym := link[1]
		to := "2:" + link[2]
		joined.AddTransition(from, sym, to)
	}
	for i := range otherToFrom {
		link := otherToFrom[i]
		from := "2:" + link[0]
		sym := link[1]
		to := "1:" + link[2]
		joined.AddTransition(from, sym, to)
	}

	return joined, nil
}

// AddState adds a new state to the NFA.If the state already exists, whether it
// is accepting is updated to match the value provided.
func (nfa *NFA[E]) AddState(state string, accepting bool) {
	if s, ok := nfa.states[state]; ok {
		s.accepting = accepting
		// Gr8! We are done.
		return
	}

	newState := nfaState[E]{
		ordering:    nfa.order,
		name:        state,
		transitions: make(map[string][]faTransition),
		accepting:   accepting,
	}
	nfa.order++

	if nfa.states == nil {
		nfa.states = map[string]nfaState[E]{}
	}

	nfa.states[state] = newState
}

// SetValue sets the value associated with a state of the NFA.
func (nfa *NFA[E]) SetValue(state string, v E) {
	s, ok := nfa.states[state]
	if !ok {
		panic(fmt.Sprintf("setting value on non-existing state: %q", state))
	}
	s.value = v
	nfa.states[state] = s
}

// GetValue gets the value associated with a state of the NFA.
func (nfa *NFA[E]) GetValue(state string) E {
	s, ok := nfa.states[state]
	if !ok {
		panic(fmt.Sprintf("getting value on non-existing state: %q", state))
	}
	return s.value
}

// AddTransition adds a new transition to the NFA. The transition will occur
// when input is encountered while the NFA is in state fromState, and will cause
// it to move to toState. Both fromState and toState must exist, or else a panic
// will occur. If the given transition already exists, this function has no
// effect.
func (nfa *NFA[E]) AddTransition(fromState string, input string, toState string) {
	curFromState, ok := nfa.states[fromState]

	if !ok {
		// Can't let you do that, Starfox
		panic(fmt.Sprintf("add transition from non-existent state %q", fromState))
	}
	if _, ok := nfa.states[toState]; !ok {
		// I'm afraid I can't do that, Dave
		panic(fmt.Sprintf("add transition to non-existent state %q", toState))
	}

	curInputTransitions, ok := curFromState.transitions[input]
	if !ok {
		curInputTransitions = make([]faTransition, 0)
	}

	newTransition := faTransition{
		input: input,
		next:  toState,
	}

	curInputTransitions = append(curInputTransitions, newTransition)

	curFromState.transitions[input] = curInputTransitions
	nfa.states[fromState] = curFromState
}
