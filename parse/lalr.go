package parse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dekarrin/ictiobus/automaton"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/types"
	"github.com/dekarrin/rosed"
)

// EmptyLALR1Parser returns a completely empty LALR1Parser, unsuitable for use.
// Generally this should not be used directly except for internal purposes; use
// GenerateLALR1Parser to generate one ready for use
func EmptyLALR1Parser() *lrParser {
	return &lrParser{table: &lalr1Table{}, parseType: types.ParserLALR1}
}

// GenerateLALR1Parser returns a parser that uses the set of canonical
// LR(1) items from g to parse input in language g. The provided language must
// be in LR(1) or else the a non-nil error is returned.
//
// allowAmbig allows the use of ambiguous grammars; in cases where there is a
// shift-reduce conflict, shift will be preferred. If the grammar is detected as
// ambiguous, the 2nd arg 'ambiguity warnings' will be filled with each
// ambiguous case detected.
func GenerateLALR1Parser(g grammar.Grammar, allowAmbig bool) (*lrParser, []string, error) {
	table, ambigWarns, err := constructLALR1ParseTable(g, allowAmbig)
	if err != nil {
		return &lrParser{}, nil, err
	}

	return &lrParser{table: table, parseType: types.ParserLALR1, gram: g}, ambigWarns, nil
}

// constructLALR1ParseTable constructs the LALR(1) table for G.
// It augments grammar G to produce G', then the canonical collection of sets of
// LR(1) items of G' is used to construct a table with applicable GOTO and
// ACTION columns.
//
// This is an implementation of Algorithm 4.59, "An easy, but space-consuming
// LALR table construction", from the purple dragon book. In the comments, most
// of which is lifted directly from the textbook, GOTO[i, A] refers to the vaue
// of the table's GOTO column at state i, symbol A, while GOTO(i, A) refers to
// the "precomputed GOTO function for grammar G'".
//
// allowAmbig allows the use of an ambiguous grammar; in this case, shift/reduce
// conflicts are resolved by preferring shift. Grammars which result in
// reduce/reduce conflicts will still be rejected. If the grammar is detected as
// ambiguous, the 2nd arg 'ambiguity warnings' will be filled with each
// ambiguous case detected.
func constructLALR1ParseTable(g grammar.Grammar, allowAmbig bool) (LRParseTable, []string, error) {
	dfa, _ := automaton.NewLALR1ViablePrefixDFA(g)
	dfa.NumberStates()

	table := &lalr1Table{
		gPrime:     g.Augmented(),
		gTerms:     g.Terminals(),
		gStart:     g.StartSymbol(),
		gNonTerms:  g.NonTerminals(),
		dfa:        dfa,
		itemCache:  map[string]grammar.LR1Item{},
		allowAmbig: allowAmbig,
	}

	// collect item cache from the states of our lr1 DFA
	allStates := textfmt.OrderedKeys(table.dfa.States())
	for _, dfaStateName := range allStates {
		itemSet := table.dfa.GetValue(dfaStateName)
		for k := range itemSet {
			table.itemCache[k] = itemSet[k]
		}
	}

	// check that we dont hit conflicts in ACTION
	var ambigWarns []string
	for i := range dfa.States() {
		fromState := fmt.Sprintf(" (from DFA state %q)", textfmt.TruncateWith(i, 4, "..."))
		for _, a := range table.gPrime.Terminals() {
			itemSet := table.dfa.GetValue(i)
			var matchFound bool
			var act LRAction
			for itemStr := range itemSet {
				item := table.itemCache[itemStr]
				A := item.NonTerminal
				alpha := item.Left
				beta := item.Right
				b := item.Lookahead
				if table.gPrime.IsTerminal(a) && len(beta) > 0 && beta[0] == a {
					j, err := table.Goto(i, a)
					if err == nil {
						// match found
						shiftAct := LRAction{Type: LRShift, State: j}
						if matchFound && !shiftAct.Equal(act) {
							if allowAmbig {
								ambigWarns = append(ambigWarns, makeLRConflictError(act, shiftAct, a).Error()+fromState)
								act = shiftAct
							} else {
								return nil, ambigWarns, fmt.Errorf("grammar is not LALR(1): %w", makeLRConflictError(act, shiftAct, a))
							}
						} else {
							act = shiftAct
							matchFound = true
						}
					}
				}

				if len(beta) == 0 && A != table.gPrime.StartSymbol() && a == b {
					reduceAct := LRAction{Type: LRReduce, Symbol: A, Production: grammar.Production(alpha)}
					if matchFound && !reduceAct.Equal(act) {
						if isSRConflict, _ := isShiftReduceConlict(act, reduceAct); isSRConflict && allowAmbig {
							// do nothing; new action is a reduce so it's already resolved
							ambigWarns = append(ambigWarns, makeLRConflictError(act, reduceAct, a).Error()+fromState)
						} else {
							return nil, ambigWarns, fmt.Errorf("grammar is not LALR(1): %w", makeLRConflictError(act, reduceAct, a))
						}
					} else {
						act = reduceAct
						matchFound = true
					}
				}

				if a == "$" && b == "$" && A == table.gPrime.StartSymbol() && len(alpha) == 1 && alpha[0] == table.gStart && len(beta) == 0 {
					newAct := LRAction{Type: LRAccept}
					if matchFound && !newAct.Equal(act) {
						return nil, ambigWarns, fmt.Errorf("grammar is not LALR(1): %w", makeLRConflictError(act, newAct, a))
					}
					act = newAct
					matchFound = true
				}
			}
		}
	}

	return table, ambigWarns, nil
}

type lalr1Table struct {
	gPrime     grammar.Grammar
	gStart     string
	dfa        automaton.DFA[box.SVSet[grammar.LR1Item]]
	itemCache  map[string]grammar.LR1Item
	gTerms     []string
	gNonTerms  []string
	allowAmbig bool
}

func (lalr1 *lalr1Table) MarshalBinary() ([]byte, error) {
	data := rezi.EncBinary(lalr1.gPrime)
	data = append(data, rezi.EncString(lalr1.gStart)...)
	data = append(data, rezi.EncMapStringToBinary(lalr1.itemCache)...)
	data = append(data, rezi.EncSliceString(lalr1.gTerms)...)
	data = append(data, rezi.EncSliceString(lalr1.gNonTerms)...)
	data = append(data, rezi.EncBool(lalr1.allowAmbig)...)

	// now the long part, the dfa
	dfaBytes := lalr1.dfa.MarshalBytes(func(s box.SVSet[grammar.LR1Item]) []byte {
		keys := s.Elements()
		sort.Strings(keys)

		innerData := rezi.EncInt(len(keys))
		for _, k := range keys {
			v := s.Get(k)

			innerData = append(innerData, rezi.EncString(k)...)
			innerData = append(innerData, rezi.EncBinary(v)...)
		}

		return innerData
	})
	data = append(data, rezi.EncInt(len(dfaBytes))...)
	data = append(data, dfaBytes...)

	return data, nil
}

func (lalr1 *lalr1Table) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	n, err = rezi.DecBinary(data, &lalr1.gPrime)
	if err != nil {
		return fmt.Errorf(".gPrime: %w", err)
	}
	data = data[n:]

	lalr1.gStart, n, err = rezi.DecString(data)
	if err != nil {
		return fmt.Errorf(".gStart: %w", err)
	}
	data = data[n:]

	var ptrMap map[string]*grammar.LR1Item
	ptrMap, n, err = rezi.DecMapStringToBinary[*grammar.LR1Item](data)
	if err != nil {
		return fmt.Errorf(".itemCache: %w", err)
	}
	lalr1.itemCache = map[string]grammar.LR1Item{}
	for k := range ptrMap {
		if ptrMap[k] != nil {
			lalr1.itemCache[k] = *ptrMap[k]
		} else {
			lalr1.itemCache[k] = grammar.LR1Item{}
		}
	}
	data = data[n:]

	lalr1.gTerms, n, err = rezi.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".gTerms: %w", err)
	}
	data = data[n:]

	lalr1.gNonTerms, n, err = rezi.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".gNonTerms: %w", err)
	}
	data = data[n:]

	lalr1.allowAmbig, n, err = rezi.DecBool(data)
	if err != nil {
		return fmt.Errorf(".allowAmbig: %w", err)
	}
	data = data[n:]

	var dfaBytesLen int
	dfaBytesLen, n, err = rezi.DecInt(data)
	if err != nil {
		// TODO: rename all .dfa-ish fields to actually be dfa
		return fmt.Errorf(".dfa: %w", err)
	}
	data = data[n:]
	if len(data) < dfaBytesLen {
		// TODO: make all "not enough bytes" messages be unexpected EOF
		return fmt.Errorf(".dfa: unexpected EOF")
	}
	dfaBytes := data[:dfaBytesLen]

	lalr1.dfa, err = automaton.UnmarshalDFABytes(dfaBytes, func(b []byte) (box.SVSet[grammar.LR1Item], error) {
		var innerN int
		var innerErr error
		var numEntries int
		set := box.NewSVSet[grammar.LR1Item]()

		numEntries, innerN, innerErr = rezi.DecInt(b)
		if innerErr != nil {
			return nil, fmt.Errorf("get entry count: %w", innerErr)
		}
		b = b[innerN:]
		for i := 0; i < numEntries; i++ {
			var k string
			var v grammar.LR1Item

			k, innerN, innerErr = rezi.DecString(b)
			if innerErr != nil {
				return nil, fmt.Errorf("entry[%d]: %w", i, innerErr)
			}
			b = b[innerN:]

			innerN, innerErr = rezi.DecBinary(b, &v)
			if innerErr != nil {
				return nil, fmt.Errorf("entry[%s]: %w", k, innerErr)
			}
			b = b[innerN:]

			set.Add(k)
			set.Set(k, v)
		}

		return set, nil
	})
	if err != nil {
		return fmt.Errorf(".dfa: %w", err)
	}

	return nil
}

func (lalr1 *lalr1Table) GetDFA() string {
	var sb strings.Builder
	automaton.OutputSetValuedDFA(&sb, lalr1.dfa)
	return sb.String()
}

func (lalr1 *lalr1Table) Action(i, a string) LRAction {
	// Algorithm 4.59, which we are using for construction of the LALR(1) parse
	// table, explicitly mentions to construct the Action table as it is done
	// in Algorithm 4.56.

	// step 2 of algorithm 4.56, "Construction of canonical-LR parsing tables",
	// for reference:

	// 2. State i is constructed from Iᵢ. The parsing actions for state i are
	// determined as follows:

	// (a) If [A -> α.aβ, b] is in Iᵢ and GOTO(Iᵢ, a) = Iⱼ, then set
	// ACTION[i, a] to "shift j." Here a must be a terminal.

	// (b) If [A -> α., a] is in Iᵢ, A != S', then set ACTION[i, a] to "reduce
	// A -> α".

	// get our set back from current state so we can check it; this is our Iᵢ

	// get our set back from current state so we can check it; this is our Iᵢ
	itemSet := lalr1.dfa.GetValue(i)

	// we have gauranteed that these dont conflict during construction; still,
	// check it so we can panic if it conflicts
	var alreadySet bool
	var act LRAction

	// Okay, "[some random item] is in Iᵢ" is suuuuuuuuper vague. We're
	// basically going to have to check each item and see if it is in the
	// pattern. I *guess* ::::/
	for itemStr := range itemSet {
		item := lalr1.itemCache[itemStr]

		// given item is [A -> α.β, b]:
		A := item.NonTerminal
		alpha := item.Left
		beta := item.Right
		b := item.Lookahead

		// (a) If [A -> α.aβ, b] is in Iᵢ and GOTO(Iᵢ, a) = Iⱼ, then set
		// ACTION[i, a] to "shift j." Here a must be a terminal.
		//
		// we'll assume α can be ε.
		// β can also be ε but note this β is rly β[1:] from earlier notation
		// used to assign beta (beta := item.Right).
		if lalr1.gPrime.IsTerminal(a) && len(beta) > 0 && beta[0] == a {
			j, err := lalr1.Goto(i, a)

			// it's okay if we get an error; it just means there is no
			// transition defined (i think, glub, the purple dragon book's
			// method of constructing GOTO would have it returning an empty
			// set in this case but unshore), so it is not a match.
			if err == nil {
				// match found
				shiftAct := LRAction{Type: LRShift, State: j}
				if alreadySet && !shiftAct.Equal(act) {
					// assuming shift/shift conflicts do not occur, and assuming
					// we have just created a shift, we must be in a
					// shift/reduce conflict here.
					if lalr1.allowAmbig {
						// this is fine, resolve in favor of shift
						act = shiftAct
					} else {
						panic(fmt.Sprintf("grammar is not LALR(1): %s", makeLRConflictError(act, shiftAct, a).Error()))
					}
				} else {
					act = shiftAct
					alreadySet = true
				}
			}
		}

		// (b) If [A -> α., a] is in Iᵢ, A != S', then set ACTION[i, a] to
		// "reduce A -> α".
		//
		// we'll assume α can be empty.
		// the beta we previously retrieved MUST be empty.
		// further, lookahead b MUST be a.
		if len(beta) == 0 && A != lalr1.gPrime.StartSymbol() && a == b {
			reduceAct := LRAction{Type: LRReduce, Symbol: A, Production: grammar.Production(alpha)}
			if alreadySet && !reduceAct.Equal(act) {
				if isSRConflict, _ := isShiftReduceConlict(act, reduceAct); isSRConflict && lalr1.allowAmbig {
					// we are in a shift/reduce conflict; the prior def is a
					// shift. resolve this in favor of shift by simply not assigning
					// the new one.
				} else {
					panic(fmt.Sprintf("grammar is not LALR(1): %s", makeLRConflictError(act, reduceAct, a).Error()))
				}
			} else {
				act = reduceAct
				alreadySet = true
			}
		}

		// (c) If [S' -> S., $] is in Iᵢ, then set ACTION[i, $] to "accept".
		if a == "$" && b == "$" && A == lalr1.gPrime.StartSymbol() && len(alpha) == 1 && alpha[0] == lalr1.gStart && len(beta) == 0 {
			acceptAct := LRAction{Type: LRAccept}
			if alreadySet && !acceptAct.Equal(act) {
				panic(fmt.Sprintf("grammar is not LALR(1): %s", makeLRConflictError(act, acceptAct, a).Error()))
			}
			act = acceptAct
			alreadySet = true
		}
	}

	// if we haven't found one, error
	if !alreadySet {
		act.Type = LRError
	}

	return act
}

func (lalr1 *lalr1Table) Goto(state, symbol string) (string, error) {
	newState := lalr1.dfa.Next(state, symbol)
	if newState == "" {
		return "", fmt.Errorf("GOTO[%q, %q] is an error entry", state, symbol)
	}
	return newState, nil
}

func (lalr1 *lalr1Table) Initial() string {
	return lalr1.dfa.Start
}

func (lalr1 *lalr1Table) String() string {
	// need mapping of state to indexes
	stateRefs := map[string]string{}

	// need to gaurantee order
	stateNames := lalr1.dfa.States().Elements()
	sort.Strings(stateNames)

	// put the initial state first
	for i := range stateNames {
		if stateNames[i] == lalr1.dfa.Start {
			old := stateNames[0]
			stateNames[0] = stateNames[i]
			stateNames[i] = old
			break
		}
	}
	for i := range stateNames {
		stateRefs[stateNames[i]] = fmt.Sprintf("%d", i)
	}

	allTerms := make([]string, len(lalr1.gTerms))
	copy(allTerms, lalr1.gTerms)
	allTerms = append(allTerms, "$")

	// okay now do data setup
	data := [][]string{}

	// set up the headers
	headers := []string{"S", "|"}

	for _, t := range allTerms {
		headers = append(headers, fmt.Sprintf("A:%s", t))
	}

	headers = append(headers, "|")

	for _, nt := range lalr1.gNonTerms {
		headers = append(headers, fmt.Sprintf("G:%s", nt))
	}
	data = append(data, headers)

	// now need to do each state
	for stateIdx := range stateNames {
		i := stateNames[stateIdx]
		row := []string{stateRefs[i], "|"}

		for _, t := range allTerms {
			act := lalr1.Action(i, t)

			cell := ""
			switch act.Type {
			case LRAccept:
				cell = "acc"
			case LRReduce:
				// reduces to the state that corresponds with the symbol
				cell = fmt.Sprintf("r%s -> %s", act.Symbol, act.Production.String())
			case LRShift:
				cell = fmt.Sprintf("s%s", stateRefs[act.State])
			case LRError:
				// do nothing, err is blank
			}

			row = append(row, cell)
		}

		row = append(row, "|")

		for _, nt := range lalr1.gNonTerms {
			var cell = ""

			gotoState, err := lalr1.Goto(i, nt)
			if err == nil {
				cell = stateRefs[gotoState]
			}

			row = append(row, cell)
		}

		data = append(data, row)
	}

	// This used to be 120 width. Glu88in' *8et* on that. lol.
	return rosed.
		Edit("").
		InsertTableOpts(0, data, 10, rosed.Options{
			TableHeaders:             true,
			NoTrailingLineSeparators: true,
		}).
		String()
}
