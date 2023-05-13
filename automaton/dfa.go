package automaton

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/rosed"
)

// DFA is a deterministic finite automaton. It holds 'values' within its states;
// supplementary information associated with each state that is not required for
// the DFA to function.
type DFA[E any] struct {
	order  uint64
	states map[string]dfaState[E]

	// Start is the starting state of the DFA.
	Start string
}

// MarshalBytes converts a DFA into a slice of bytes that can be decoded with
// UnmarshalDFABytes. The value held within its states is encoded to bytes using
// the provided conversion function.
func (dfa DFA[E]) MarshalBytes(conv func(E) []byte) []byte {
	data := rezi.EncInt(int(dfa.order))
	data = append(data, rezi.EncString(dfa.Start)...)

	if dfa.states == nil {
		data = append(data, rezi.EncInt(-1)...)
	} else {
		stateNames := textfmt.OrderedKeys(dfa.states)
		data = append(data, rezi.EncInt(len(stateNames))...)
		for _, stateName := range stateNames {
			data = append(data, rezi.EncString(stateName)...)
			stateBytes := dfa.states[stateName].MarshalBytes(conv)
			data = append(data, rezi.EncInt(len(stateBytes))...)
			data = append(data, stateBytes...)
		}
	}

	return data
}

// UnmarshalDFABytes takes a slice of bytes created by MarshalBytes and decodes
// it into a new DFA. The values held within its states is decoded from bytes
// using the provided conversion function.
func UnmarshalDFABytes[E any](data []byte, conv func([]byte) (E, error)) (DFA[E], error) {
	var dfa DFA[E]
	var n int
	var err error

	var iVal int
	iVal, n, err = rezi.DecInt(data)
	if err != nil {
		return dfa, fmt.Errorf(".order: %w", err)
	}
	dfa.order = uint64(iVal)
	data = data[n:]

	dfa.Start, n, err = rezi.DecString(data)
	if err != nil {
		return dfa, fmt.Errorf(".Start: %w", err)
	}
	data = data[n:]

	var numStates int
	numStates, n, err = rezi.DecInt(data)
	if err != nil {
		return dfa, fmt.Errorf(".states: %w", err)
	}
	data = data[n:]
	if numStates > -1 {
		dfa.states = map[string]dfaState[E]{}
		for i := 0; i < numStates; i++ {
			var name string
			var stateBytesLen int
			var state dfaState[E]

			name, n, err = rezi.DecString(data)
			if err != nil {
				return dfa, fmt.Errorf(".states[%d]: %w", i, err)
			}
			data = data[n:]

			stateBytesLen, n, err = rezi.DecInt(data)
			if err != nil {
				return dfa, fmt.Errorf(".states[%s]: value bytes len: %w", name, err)
			}
			data = data[n:]
			stateBytes := data[:stateBytesLen]
			state, err = unmarshalDFAStateBytes(stateBytes, conv)
			if err != nil {
				return dfa, fmt.Errorf(".states[%s]: %w", name, err)
			}
			data = data[stateBytesLen:]

			dfa.states[name] = state
		}
	} else {
		dfa.states = nil
	}

	return dfa, nil
}

// Copy returns a deeply-copied duplicate of this DFA.
func (dfa DFA[E]) Copy() DFA[E] {
	copied := DFA[E]{
		Start:  dfa.Start,
		states: make(map[string]dfaState[E]),
		order:  dfa.order,
	}

	for k := range dfa.states {
		copied.states[k] = dfa.states[k].Copy()
	}

	return copied
}

// TransformDFA converts the values held within a DFA to new values. The
// transform function is called on each state's value and used to create a new
// one in a new DFA. The original DFA is not modified.
func TransformDFA[E1, E2 any](dfa DFA[E1], transform func(old E1) E2) DFA[E2] {
	copied := DFA[E2]{
		states: make(map[string]dfaState[E2]),
		Start:  dfa.Start,
		order:  dfa.order,
	}

	for k := range dfa.states {
		oldState := dfa.states[k]
		copiedState := dfaState[E2]{
			name:        oldState.name,
			value:       transform(oldState.value),
			transitions: make(map[string]faTransition),
			accepting:   oldState.accepting,
			ordering:    oldState.ordering,
		}

		for sym := range oldState.transitions {
			copiedState.transitions[sym] = oldState.transitions[sym]
		}

		copied.states[k] = copiedState
	}

	return copied
}

// DFAToNFA converts the DFA into an equivalent non-deterministic finite
// automaton type. Note that the type change doesn't suddenly make usage
// non-deterministic but it does allow for non-deterministic transitions to be
// added.
func DFAToNFA[E any](dfa DFA[E]) NFA[E] {
	nfa := NFA[E]{
		Start:  dfa.Start,
		states: map[string]nfaState[E]{},
		order:  dfa.order,
	}

	for sName := range dfa.states {
		dState := dfa.states[sName]

		nState := nfaState[E]{
			ordering:    dState.ordering,
			name:        dState.name,
			value:       dState.value,
			transitions: map[string][]faTransition{},
			accepting:   dState.accepting,
		}

		for sym := range dState.transitions {
			dTrans := dState.transitions[sym]
			nState.transitions[sym] = []faTransition{{input: dTrans.input, next: dTrans.next}}
		}

		nfa.states[sName] = nState
	}

	return nfa
}

// NumberStates renames all states to each have a unique name based on an
// increasing number sequence. The starting state is guaranteed to be numbered
// 0; beyond that, the states are put in order they were added.
func (dfa *DFA[E]) NumberStates() {
	if _, ok := dfa.states[dfa.Start]; !ok {
		panic("can't number states of DFA with no start state set")
	}
	origStateNames := textfmt.OrderedKeys(dfa.States())

	// make shore to pull out starting state and place at front
	startIdx := -1
	for i := range origStateNames {
		if origStateNames[i] == dfa.Start {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		panic("couldn't find starting state; should never happen")
	}

	origStateNames = append(origStateNames[:startIdx], origStateNames[startIdx+1:]...)
	origStateNames = slices.SortBy(origStateNames, func(s1, s2 string) bool {
		return dfa.states[s1].ordering < dfa.states[s2].ordering
	})
	origStateNames = append([]string{dfa.Start}, origStateNames...)

	numMapping := map[string]string{}
	for i := range origStateNames {
		name := origStateNames[i]
		newName := fmt.Sprintf("%d", i)
		numMapping[name] = newName
	}

	// to keep things simple, instead of searching for every instance of each
	// name which is an expensive operation, we'll just build an entirely new
	// DFA using our mapping rules to adjust names as we go, then steal its
	// states map.

	newDfa := &DFA[E]{
		states: make(map[string]dfaState[E]),
		Start:  numMapping[dfa.Start],
	}

	// first, add the initial states
	for _, name := range origStateNames {
		st := dfa.states[name]
		newName := numMapping[name]
		newDfa.AddState(newName, st.accepting)

		newSt := newDfa.states[newName]
		newSt.ordering = st.ordering
		newDfa.states[newName] = newSt

		newDfa.SetValue(newName, st.value)

		// transitions come later, need to add all states *first*
	}

	// add initial transitions
	for _, name := range origStateNames {
		st := dfa.states[name]
		from := numMapping[name]

		for sym := range st.transitions {
			t := st.transitions[sym]
			to := numMapping[t.next]
			newDfa.AddTransition(from, sym, to)
		}
	}

	// oh ya, just gonna go ahead and sneeeeeeeak this on away from ya
	dfa.states = newDfa.states
	dfa.Start = newDfa.Start
}

// SetValue sets the value associated with a state of the DFA.
func (dfa *DFA[E]) SetValue(state string, v E) {
	s, ok := dfa.states[state]
	if !ok {
		panic(fmt.Sprintf("setting value on non-existing state: %q", state))
	}
	s.value = v
	dfa.states[state] = s
}

// GetValue gets the value associated with a state of the DFA.
func (dfa *DFA[E]) GetValue(state string) E {
	s, ok := dfa.states[state]
	if !ok {
		panic(fmt.Sprintf("getting value on non-existing state: %q", state))
	}
	return s.value
}

// IsAccepting returns whether the given state is an accepting (terminating)
// state. Returns false if the state does not exist.
func (dfa DFA[E]) IsAccepting(state string) bool {
	s, ok := dfa.states[state]
	if !ok {
		return false
	}

	return s.accepting
}

// Validate immediately returns an error if it finds the following:
//
// Any state impossible to reach (no transitions to it).
// Any transition leading to a state that doesn't exist.
// A start that isn't a state that exists.
func (dfa DFA[E]) Validate() error {
	errs := ""
	// all states must be reachable somehow. Must be reachable by some other
	// state if not the start state.
	for sName := range dfa.states {
		if sName == dfa.Start {
			continue
		}

		atLeastOneTransitionTo := false
		for otherName := range dfa.states {
			if otherName == sName {
				continue
			}

			st := dfa.states[otherName]

			for i := range st.transitions {
				if st.transitions[i].next == sName {
					atLeastOneTransitionTo = true
					break
				}
			}

			if atLeastOneTransitionTo {
				break
			}
		}
		if !atLeastOneTransitionTo {
			errs += fmt.Sprintf("\nno transitions to non-start state %q", sName)
		}
	}

	// all transitions must lead to an existing state
	for sName := range dfa.states {
		// dont skip if the starting state; this applies to that state too
		st := dfa.states[sName]

		for symbol := range st.transitions {
			nextState := st.transitions[symbol].next

			if _, ok := dfa.states[nextState]; !ok {
				errs += fmt.Sprintf("\nstate %q transitions to non-existing state: %q", sName, st.transitions[symbol])
			}
		}
	}

	// finally, start must be a reel state that exists
	if _, ok := dfa.states[dfa.Start]; !ok {
		errs += fmt.Sprintf("\nstart state does not exist: %q", dfa.Start)
	}

	if len(errs) > 0 {
		errs = errs[1:]
		return fmt.Errorf(errs)
	}

	return nil
}

// States returns all states in the dfa.
func (dfa DFA[E]) States() box.StringSet {
	states := box.NewStringSet()

	for k := range dfa.states {
		states.Add(k)
	}

	return states
}

// Next returns the next state of the DFA, given a current state and an input.
// Will return "" if state is not an existing state or if there is no transition
// from the given state on the given input.
func (dfa DFA[E]) Next(fromState string, input string) string {
	state, ok := dfa.states[fromState]
	if !ok {
		return ""
	}

	transition, ok := state.transitions[input]
	if !ok {
		return ""
	}

	return transition.next
}

// AllTransitionsTo gets all transitions to the given state. It returns a slice
// of 2-tuples that each contain the originating state name followed by the
// input that causes transitions to toState.
func (dfa DFA[E]) AllTransitionsTo(toState string) [][2]string {
	if _, ok := dfa.states[toState]; !ok {
		// Gr8! We are done.
		return [][2]string{}
	}

	transitions := [][2]string{}

	s := dfa.States()

	for _, sName := range s.Elements() {
		state := dfa.states[sName]
		for k := range state.transitions {
			if state.transitions[k].next == toState {
				trans := [2]string{sName, k}
				transitions = append(transitions, trans)
			}
		}
	}

	return transitions
}

// RemoveState removes a state from the DFA. The state can only be removed if
// there are not currently any transitions to it; trying to remove a state that
// has transitions to it will cause a panic.
func (dfa *DFA[E]) RemoveState(state string) {
	_, ok := dfa.states[state]
	if !ok {
		// Gr8! We are done.
		return
	}

	// is this allowed?
	transitionsTo := dfa.AllTransitionsTo(state)

	if len(transitionsTo) > 0 {
		panic("can't remove state that is currently traversed to")
	}

	delete(dfa.states, state)
}

// AddState adds a new state to the DFA.
func (dfa *DFA[E]) AddState(state string, accepting bool) {
	if _, ok := dfa.states[state]; ok {
		// Gr8! We are done.
		return
	}

	newState := dfaState[E]{
		ordering:    dfa.order,
		name:        state,
		transitions: make(map[string]faTransition),
		accepting:   accepting,
	}
	dfa.order++

	if dfa.states == nil {
		dfa.states = map[string]dfaState[E]{}
	}

	dfa.states[state] = newState
}

// RemoveTransition removes a transition from the DFA. The transition that uses
// input to transition from fromState to toState is removed. If there is not
// currently a transition that satisfies that, this function has no effect.
func (dfa *DFA[E]) RemoveTransition(fromState string, input string, toState string) {
	curFromState, ok := dfa.states[fromState]
	if !ok {
		// Gr8! We are done.
		return
	}

	curTrans, ok := curFromState.transitions[input]
	if !ok {
		// Done early
		return
	}

	if curTrans.next != toState {
		// already not here
		return
	}

	// otherwise, remove the relation
	delete(curFromState.transitions, input)
}

// AddTransition adds a new transition to the DFA. The transition will occur
// when input is encountered while the DFA is in state fromState, and will cause
// it to move to toState. Both fromState and toState must exist, or else a panic
// will occur. If the given transition already exists, this function has no
// effect.
func (dfa *DFA[E]) AddTransition(fromState string, input string, toState string) {
	curFromState, ok := dfa.states[fromState]

	if !ok {
		// Can't let you do that, Starfox
		panic(fmt.Sprintf("add transition from non-existent state %q", fromState))
	}
	if _, ok := dfa.states[toState]; !ok {
		// I'm afraid I can't do that, Dave
		panic(fmt.Sprintf("add transition to non-existent state %q", toState))
	}

	trans := faTransition{
		input: input,
		next:  toState,
	}

	curFromState.transitions[input] = trans
	dfa.states[fromState] = curFromState
}

// String returns the string representation of a DFA.
func (dfa DFA[E]) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<START: %q, STATES:", dfa.Start))

	orderedStates := textfmt.OrderedKeys(dfa.states)
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
		sb.WriteString(dfa.states[orderedStates[i]].String())

		if i+1 < len(dfa.states) {
			sb.WriteRune(',')
		} else {
			sb.WriteRune('\n')
		}
	}

	sb.WriteRune('>')

	return sb.String()
}

// OutputSetValuedDFA writes a pretty-print representation of a DFA whose values
// in its states are box.SVSets of some type that implements fmt.Stringer. The
// representation is written to w.
func OutputSetValuedDFA[E fmt.Stringer](w io.Writer, dfa DFA[box.SVSet[E]]) {
	// lol let's get some buffering here
	bw := bufio.NewWriter(w)

	bw.WriteString("DFA:\n")
	bw.WriteString("\tStart: ")
	bw.WriteRune('"')
	bw.WriteString(dfa.Start)
	bw.WriteString("\"\n")

	// now get ordered states
	orderedStates := textfmt.OrderedKeys(dfa.states)
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

	tabOpts := rosed.Options{TableBorders: true}

	bw.WriteString("\tStates:")
	// write out each state in a reasonable way
	for i := range orderedStates {
		bw.WriteString("\n")
		state := dfa.states[orderedStates[i]]
		layout := rosed.Editor{Options: tabOpts}

		// get name and accepting data
		nameCell := fmt.Sprintf("%q", state.name)
		if state.accepting {
			nameCell = "(" + nameCell + ")"
		}
		nameData := [][]string{{nameCell}}

		// get item data for the state, in deterministic ordering
		itemData := [][]string{}
		items := state.value

		lrItemNames := items.Elements()
		sort.Strings(lrItemNames)

		for i := range lrItemNames {
			it := items.Get(lrItemNames[i])
			cell := fmt.Sprintf("[%s]", it.String())
			itemData = append(itemData, []string{cell})
		}

		// okay, finally, get transitions, in deterministic ordering
		transData := [][]string{}
		transOrdered := textfmt.OrderedKeys(state.transitions)

		for i := range transOrdered {
			t := state.transitions[transOrdered[i]]

			cell := fmt.Sprintf("%q ==> %q", t.input, t.next)
			transData = append(transData, []string{cell})
		}

		layout = layout.
			InsertTable(rosed.End, nameData, 80)

		if len(itemData) > 0 {
			layout = layout.
				LinesFrom(-1).
				Delete(0, 81).
				Commit().
				InsertTable(rosed.End, itemData, 80)
		}

		if len(transData) > 0 {
			layout = layout.
				LinesFrom(-1).
				Delete(0, 81).
				Commit().
				InsertTable(rosed.End, transData, 80)
		}

		str := layout.
			Indent(1).
			String()

		bw.WriteString(str)
	}
	if len(orderedStates) == 0 {
		bw.WriteString(" (none)\n")
	}

	bw.Flush()
}

// ValueString returns the string representation of the DFA with its states'
// values included in the output.
func (dfa DFA[E]) ValueString() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("<START: %q, STATES:", dfa.Start))

	orderedStates := textfmt.OrderedKeys(dfa.states)
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
		sb.WriteString(dfa.states[orderedStates[i]].ValueString())

		if i+1 < len(dfa.states) {
			sb.WriteRune(',')
		} else {
			sb.WriteRune('\n')
		}
	}

	sb.WriteRune('>')

	return sb.String()
}

// NewLALR1ViablePrefixDFA creates a new DFA whose states are made up of the
// sets of items used in an LALR(1) parser. The grammar of the language that is
// accepted by the parser, g, must be LALR(1) and it must be non-augmented.
// Returns an error if g is not LALR(1).
func NewLALR1ViablePrefixDFA(g grammar.Grammar) (DFA[box.SVSet[grammar.LR1Item]], error) {
	lr1Dfa := NewLR1ViablePrefixDFA(g)

	// get an NFA so we can start fixing things
	lalrNfa := DFAToNFA(lr1Dfa)

	// counter for unique state name
	newStateNum := 0

	// now start merging states
	updated := true
	for updated {
		updated = false

		alreadyMerged := box.NewStringSet()
		states := lalrNfa.States()
		stateVals := map[string]box.SVSet[grammar.LR1Item]{}
		orderedStateElements := states.Elements()
		sort.Strings(orderedStateElements)
		for _, name := range orderedStateElements {
			stateVals[name] = lalrNfa.GetValue(name)
		}

		for _, stateName := range orderedStateElements {
			if alreadyMerged.Has(stateName) {
				continue
			}

			mergeWith := []string{}
			coreSet := grammar.CoreSet(stateVals[stateName])

			// need to find ALL to merge w or this is gonna get wild REEL quick
			for _, otherStateName := range orderedStateElements {
				if stateName == otherStateName {
					continue
				}

				otherCoreSet := grammar.CoreSet(stateVals[otherStateName])

				// Note: we do NOT enforce an ordering in general on which
				// states are merged first. this could cause issues; doing them
				// in an arbitrary order

				// check their cores
				if coreSet.Equal(otherCoreSet) {
					mergeWith = append(mergeWith, otherStateName)
				}
			}

			// now we merge any that have been queued to do so
			if len(mergeWith) > 0 {
				updated = true
				alreadyMerged.Add(stateName)
				destState := lalrNfa.states[stateName]
				mergedStateSet := box.NewSVSet(stateVals[stateName])

				for i := range mergeWith {
					alreadyMerged.Add(mergeWith[i])
					mergedStateSet.AddAll(stateVals[mergeWith[i]])
				}

				// We COULD tell what new name of state would be NOW, but to keep
				// things from overlapping during the process we will be setting
				// to a unique number and updating after all merges are complete
				// (at which point there should be 0 conflicting state names).
				newStateName := fmt.Sprintf("%d", newStateNum)
				newStateNum++
				destState.name = mergedStateSet.StringOrdered()
				destState.value = mergedStateSet

				// and so we can rewrite transitions from the old states to the
				// new one
				for i := range mergeWith {
					transitionsToMerged := lalrNfa.AllTransitionsTo(mergeWith[i])

					for j := range transitionsToMerged {
						trans := transitionsToMerged[j]
						from := trans.from
						sym := trans.input
						idx := trans.index

						// rewrite the transition to new state
						lalrNfa.states[from].transitions[sym][idx] = faTransition{input: sym, next: newStateName}
					}

					// also, check to see if we need to update start
					if lalrNfa.Start == mergeWith[i] {
						lalrNfa.Start = newStateName
					}
				}

				// also rewrite any transitions to the merged-to state
				transitionsToDestState := lalrNfa.AllTransitionsTo(stateName)
				for j := range transitionsToDestState {
					trans := transitionsToDestState[j]
					from := trans.from
					sym := trans.input
					idx := trans.index

					// rewrite the transition to new state
					lalrNfa.states[from].transitions[sym][idx] = faTransition{input: sym, next: newStateName}
				}

				// also, check to see if we need to update start
				if lalrNfa.Start == stateName {
					lalrNfa.Start = newStateName
				}

				// finally, enshore that any transitions we lose by deleting the
				// old state are added to the new state. this SHOULD collapse to
				// a single state by the time that things are done if it is
				// indeed an LALR(1) grammar
				for i := range mergeWith {
					lostTransitions := lalrNfa.states[mergeWith[i]].transitions
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
					delete(lalrNfa.states, mergeWith[i])
				}

				// unshore if this condition is proven not to happen, either
				// way it's 8AD so checking
				if _, ok := lalrNfa.states[newStateName]; ok {
					panic(fmt.Sprintf("merged state name conflicts w state %q already in DFA", newStateName))
				}

				// enshore the updated new state is stored...
				lalrNfa.states[newStateName] = destState

				// ...and, finally, remove the old version of it
				delete(lalrNfa.states, stateName)
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
	lalrStates := lalrNfa.States().Elements()
	for _, stateName := range lalrStates {
		st := lalrNfa.states[stateName]

		// we keep the name pre-calculated in .name, so check if there's a mismatch
		if st.name != stateName {
			newStateName := st.name
			transitionsToMerged := lalrNfa.AllTransitionsTo(stateName)

			for j := range transitionsToMerged {
				trans := transitionsToMerged[j]
				from := trans.from
				sym := trans.input
				idx := trans.index

				// rewrite the transition to new state
				lalrNfa.states[from].transitions[sym][idx] = faTransition{input: sym, next: newStateName}
			}

			// also, check to see if we need to update start
			if lalrNfa.Start == stateName {
				lalrNfa.Start = newStateName
			}

			// and now, swap the name for the reel one
			lalrNfa.states[newStateName] = st
			delete(lalrNfa.states, stateName)
		}
	}

	lalrDfa, err := directNFAToDFA(lalrNfa)
	if err != nil {
		return DFA[box.SVSet[grammar.LR1Item]]{}, fmt.Errorf("grammar is not LALR(1); resulted in inconsistent state merges")
	}

	return lalrDfa, nil
}

// NewLR1ViablePrefixDFA creates a new DFA whose states are made up of the sets
// of items used in a canonical LR(1) parser. The grammar of the language that
// is accepted by the parser, g, must be LR(1) and it must be non-augmented.
func NewLR1ViablePrefixDFA(g grammar.Grammar) DFA[box.SVSet[grammar.LR1Item]] {
	oldStart := g.StartSymbol()
	g = g.Augmented()

	initialItem := grammar.LR1Item{
		LR0Item: grammar.LR0Item{
			NonTerminal: g.StartSymbol(),
			Right:       []string{oldStart},
		},
		Lookahead: "$",
	}

	startSet := lr1CLOSURE(g, box.SVSet[grammar.LR1Item]{initialItem.String(): initialItem})

	stateSets := box.NewSVSet[box.SVSet[grammar.LR1Item]]()
	stateSets.Set(startSet.StringOrdered(), startSet)
	transitions := map[string]map[string]faTransition{}

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
					stateTransitions = map[string]faTransition{}
				}
				trans, ok := stateTransitions[s]
				if !ok {
					trans = faTransition{}
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
	dfa := DFA[box.SVSet[grammar.LR1Item]]{}

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

// lr1CLOSURE is the closure function used for constructing LR(1) item sets for
// use in a parser DFA.
//
// Note: this actually takes the grammar for each production B -> gamma in G,
// not G'. It's assumed this function is only called on a g.Augmented()
// instance.
func lr1CLOSURE(g grammar.Grammar, I box.SVSet[grammar.LR1Item]) box.SVSet[grammar.LR1Item] {
	Iset := I.Copy()
	I = Iset.(box.SVSet[grammar.LR1Item])

	updated := true
	for updated {
		updated = false
		for _, it := range I {
			if len(it.Right) >= 1 {
				B := it.Right[0]
				ruleB := g.Rule(B)
				if ruleB.NonTerminal == "" {
					continue
				}

				for _, gamma := range ruleB.Productions {
					fullArgs := make([]string, len(it.Right[1:]))
					copy(fullArgs, it.Right[1:])
					fullArgs = append(fullArgs, it.Lookahead)
					for _, b := range g.FIRST_STRING(fullArgs...).Elements() {
						if strings.ToLower(b) != b {
							continue // terminals only
						}

						var newItem grammar.LR1Item

						// SPECIAL CASE: if we're dealing with an epsilon, our
						// item will look like "A -> .". normally we are adding
						// a dot at the START of an item added in the LR1
						// CLOSURE func, but since "A -> ." should always be
						// treated as "at the end", we add a special item with
						// only the dot, and no left or right.
						if gamma.Equal(grammar.Epsilon) {
							newItem = grammar.LR1Item{LR0Item: grammar.LR0Item{NonTerminal: B}, Lookahead: b}
						} else {
							newItem = grammar.LR1Item{
								LR0Item: grammar.LR0Item{
									NonTerminal: B,
									Right:       gamma,
								},
								Lookahead: b,
							}
						}
						if !I.Has(newItem.String()) {
							I.Set(newItem.String(), newItem)
							updated = true
						}
					}
				}
			}
		}
	}
	return I
}
