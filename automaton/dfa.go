package automaton

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/internal/textfmt"
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
	origStateNames := dfa.States()

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

// States returns all names of states in the DFA. Ordering of the states in the
// returned slice is guaranteed to be stable.
func (dfa DFA[E]) States() []string {
	states := []string{}

	for k := range dfa.states {
		states = append(states, k)
	}

	sort.Strings(states)

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

// allTransitionsTo gets all transitions to the given state. It returns a slice
// of 2-tuples that each contain the originating state name followed by the
// input that causes transitions to toState.
func (dfa DFA[E]) allTransitionsTo(toState string) [][2]string {
	if _, ok := dfa.states[toState]; !ok {
		// Gr8! We are done.
		return [][2]string{}
	}

	transitions := [][2]string{}

	s := dfa.States()

	for _, sName := range s {
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
	transitionsTo := dfa.allTransitionsTo(state)

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

// GetTransitions returns the transitions out of state fromState. They are
// returned as a list of [2]string; the first item is the input symbol, the
// second is the the state the DFA transitions to on that input.
//
// If given a state that does not exist, a slice of len 0 is returned.
func (dfa DFA[E]) GetTransitions(fromState string) [][2]string {
	from, ok := dfa.states[fromState]
	if !ok {
		return nil
	}

	var transitions [][2]string
	for k := range from.transitions {
		t := from.transitions[k]
		transitions = append(transitions, [2]string{t.input, t.next})
	}

	return transitions
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
