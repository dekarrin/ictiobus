package parse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dekarrin/ictiobus/automaton"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/decbin"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/types"
	"github.com/dekarrin/rosed"
)

// EmptySLR1Parser returns a completely empty SLR1Parser, unsuitable for use.
// Generally this should not be used directly except for internal purposes; use
// GenerateSimpleLRParser to generate one ready for use
func EmptySLR1Parser() *lrParser {
	return &lrParser{table: &slrTable{}, parseType: types.ParserSLR1}
}

// GenerateSimpleLRParser returns a parser that uses SLR bottom-up parsing to
// parse languages in g. It will return an error if g is not an SLR(1) grammar.
//
// allowAmbig allows the use of ambiguous grammars; in cases where there is a
// shift-reduce conflict, shift will be preferred. If the grammar is detected as
// ambiguous, the 2nd arg 'ambiguity warnings' will be filled with each
// ambiguous case detected.
func GenerateSimpleLRParser(g grammar.Grammar, allowAmbig bool) (*lrParser, []string, error) {
	table, ambigWarns, err := constructSimpleLRParseTable(g, allowAmbig)
	if err != nil {
		return &lrParser{}, ambigWarns, err
	}

	return &lrParser{table: table, parseType: types.ParserSLR1, gram: g}, ambigWarns, nil
}

// constructSimpleLRParseTable constructs the SLR(1) table for G. It augments
// grammar G to produce G', then the canonical collection of sets of items of G'
// is used to construct a table with applicable GOTO and ACTION columns.
//
// This is an implementation of Algorithm 4.46, "Constructing an SLR-parsing
// table", from the purple dragon book. In the comments, most of which is lifted
// directly from the textbook, GOTO[i, A] refers to the vaue of the table's
// GOTO column at state i, symbol A, while GOTO(i, A) refers to the "precomputed
// GOTO function for grammar G'".
//
// allowAmbig allows the use of an ambiguous grammar; in this case, shift/reduce
// conflicts are resolved by preferring shift. Grammars which result in
// reduce/reduce conflicts will still be rejected. If the grammar is detected as
// ambiguous, the 2nd arg 'ambiguity warnings' will be filled with each
// ambiguous case detected.
func constructSimpleLRParseTable(g grammar.Grammar, allowAmbig bool) (LRParseTable, []string, error) {
	// we will skip a few steps here and simply grab the LR0 DFA for G' which
	// will pretty immediately give us our GOTO() function, since as purple
	// dragon book mentions, "intuitively, the GOTO function is used to define
	// the transitions in the LR(0) automaton for a grammar."
	lr0Automaton := automaton.NewLR0ViablePrefixNFA(g).ToDFA()
	lr0Automaton.NumberStates()

	table := &slrTable{
		gPrime:     g.Augmented(),
		gStart:     g.StartSymbol(),
		gTerms:     g.Terminals(),
		gNonTerms:  g.NonTerminals(),
		lr0:        lr0Automaton,
		itemCache:  map[string]grammar.LR0Item{},
		allowAmbig: allowAmbig,
	}

	for _, item := range table.gPrime.LR0Items() {
		table.itemCache[item.String()] = item
	}

	// check ahead to see if we would get conflicts in ACTION function
	var ambigWarns []string
	for i := range lr0Automaton.States() {
		fromState := fmt.Sprintf(" (from DFA state %q)", textfmt.TruncateWith(i, 4, "..."))
		for _, a := range table.gPrime.Terminals() {
			itemSet := table.lr0.GetValue(i)
			var matchFound bool
			var act LRAction
			for itemStr := range itemSet {
				item := table.itemCache[itemStr]
				A := item.NonTerminal
				alpha := item.Left
				beta := item.Right

				var followA box.ISet[string]
				if A != table.gPrime.StartSymbol() {
					// we'll need this later, glub 38)
					followA = table.gPrime.FOLLOW(A)
				}

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
								return nil, ambigWarns, fmt.Errorf("grammar is not SLR(1): %w", makeLRConflictError(act, shiftAct, a))
							}
						} else {
							act = shiftAct
							matchFound = true
						}
					}
				}

				if len(beta) == 0 && A != table.gPrime.StartSymbol() && followA.Has(a) {
					reduceAct := LRAction{Type: LRReduce, Symbol: A, Production: grammar.Production(alpha)}
					if matchFound && !reduceAct.Equal(act) {
						if isSRConflict, _ := isShiftReduceConlict(act, reduceAct); isSRConflict && allowAmbig {
							// do nothing; new action is a reduce so it's already resolved
							ambigWarns = append(ambigWarns, makeLRConflictError(act, reduceAct, a).Error()+fromState)
						} else {
							return nil, ambigWarns, fmt.Errorf("grammar is not SLR(1): %w", makeLRConflictError(act, reduceAct, a))
						}
					} else {
						act = reduceAct
						matchFound = true
					}
				}

				if a == "$" && A == table.gPrime.StartSymbol() && len(alpha) == 1 && alpha[0] == table.gStart && len(beta) == 0 {
					newAct := LRAction{Type: LRAccept}
					if matchFound && !newAct.Equal(act) {
						return nil, ambigWarns, fmt.Errorf("grammar is not SLR(1): %w", makeLRConflictError(act, newAct, a))
					}
					act = newAct
					matchFound = true
				}
			}
		}
	}

	return table, ambigWarns, nil
}

type slrTable struct {
	gPrime     grammar.Grammar
	gStart     string
	lr0        automaton.DFA[box.SVSet[grammar.LR0Item]]
	itemCache  map[string]grammar.LR0Item
	gTerms     []string
	gNonTerms  []string
	allowAmbig bool
}

func (slr *slrTable) MarshalBinary() ([]byte, error) {
	data := decbin.EncBinary(slr.gPrime)
	data = append(data, decbin.EncString(slr.gStart)...)
	data = append(data, decbin.EncMapStringToBinary(slr.itemCache)...)
	data = append(data, decbin.EncSliceString(slr.gTerms)...)
	data = append(data, decbin.EncSliceString(slr.gNonTerms)...)
	data = append(data, decbin.EncBool(slr.allowAmbig)...)

	// now the long part, the dfa
	dfaBytes := slr.lr0.MarshalBytes(func(s box.SVSet[grammar.LR0Item]) []byte {
		keys := s.Elements()
		sort.Strings(keys)

		innerData := decbin.EncInt(len(keys))
		for _, k := range keys {
			v := s.Get(k)

			innerData = append(innerData, decbin.EncString(k)...)
			innerData = append(innerData, decbin.EncBinary(v)...)
		}

		return innerData
	})
	data = append(data, decbin.EncInt(len(dfaBytes))...)
	data = append(data, dfaBytes...)

	return data, nil
}

func (slr *slrTable) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	n, err = decbin.DecBinary(data, &slr.gPrime)
	if err != nil {
		return fmt.Errorf(".gPrime: %w", err)
	}
	data = data[n:]

	slr.gStart, n, err = decbin.DecString(data)
	if err != nil {
		return fmt.Errorf(".gStart: %w", err)
	}
	data = data[n:]

	var ptrMap map[string]*grammar.LR0Item
	ptrMap, n, err = decbin.DecMapStringToBinary[*grammar.LR0Item](data)
	if err != nil {
		return fmt.Errorf(".itemCache: %w", err)
	}
	slr.itemCache = map[string]grammar.LR0Item{}
	for k := range ptrMap {
		if ptrMap[k] != nil {
			slr.itemCache[k] = *ptrMap[k]
		} else {
			slr.itemCache[k] = grammar.LR0Item{}
		}
	}
	data = data[n:]

	slr.gTerms, n, err = decbin.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".gTerms: %w", err)
	}
	data = data[n:]

	slr.gNonTerms, n, err = decbin.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".gNonTerms: %w", err)
	}
	data = data[n:]

	slr.allowAmbig, n, err = decbin.DecBool(data)
	if err != nil {
		return fmt.Errorf(".allowAmbig: %w", err)
	}
	data = data[n:]

	var dfaBytesLen int
	dfaBytesLen, n, err = decbin.DecInt(data)
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

	slr.lr0, err = automaton.UnmarshalDFABytes(dfaBytes, func(b []byte) (box.SVSet[grammar.LR0Item], error) {
		var innerN int
		var innerErr error
		var numEntries int
		set := box.NewSVSet[grammar.LR0Item]()

		numEntries, innerN, innerErr = decbin.DecInt(b)
		if innerErr != nil {
			return nil, fmt.Errorf("get entry count: %w", innerErr)
		}
		b = b[innerN:]
		for i := 0; i < numEntries; i++ {
			var k string
			var v grammar.LR0Item

			k, innerN, innerErr = decbin.DecString(b)
			if innerErr != nil {
				return nil, fmt.Errorf("entry[%d]: %w", i, innerErr)
			}
			b = b[innerN:]

			innerN, innerErr = decbin.DecBinary(b, &v)
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

func (slr *slrTable) GetDFA() string {
	var sb strings.Builder
	automaton.OutputSetValuedDFA(&sb, slr.lr0)
	return sb.String()
}

func (slr *slrTable) String() string {
	// need mapping of state to indexes
	stateRefs := map[string]string{}

	// need to gaurantee order
	stateNames := slr.lr0.States().Elements()
	sort.Strings(stateNames)

	// put the initial state first
	for i := range stateNames {
		if stateNames[i] == slr.lr0.Start {
			old := stateNames[0]
			stateNames[0] = stateNames[i]
			stateNames[i] = old
			break
		}
	}
	for i := range stateNames {
		stateRefs[stateNames[i]] = fmt.Sprintf("%d", i)
	}

	allTerms := make([]string, len(slr.gTerms))
	copy(allTerms, slr.gTerms)
	allTerms = append(allTerms, "$")

	// okay now do data setup
	data := [][]string{}

	// set up the headers
	headers := []string{"S", "|"}

	for _, t := range allTerms {
		headers = append(headers, fmt.Sprintf("A:%s", t))
	}

	headers = append(headers, "|")

	for _, nt := range slr.gNonTerms {
		headers = append(headers, fmt.Sprintf("G:%s", nt))
	}
	data = append(data, headers)

	// now need to do each state
	for stateIdx := range stateNames {
		i := stateNames[stateIdx]
		row := []string{stateRefs[i], "|"}

		for _, t := range allTerms {
			act := slr.Action(i, t)

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

		for _, nt := range slr.gNonTerms {
			var cell = ""

			gotoState, err := slr.Goto(i, nt)
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

func (slr *slrTable) Initial() string {
	return slr.lr0.Start
}

func (slr *slrTable) Goto(state, symbol string) (string, error) {
	// as purple  dragon book mentions, "intuitively, the GOTO function is used
	// to define the transitions in the LR(0) automaton for a grammar." We will
	// take advantage of the corollary; we already have the automaton defined,
	// so consequently the transitions of it can be used to derive the value of
	// GOTO(i, a).

	// assume the state is the concatenated items in the set. Up to caller to
	// enshore this is the glubbin case.

	// step 3 of algorithm 4.46, "Constructing an SLR-parsing table", for
	// reference

	// 3. The goto transitions for state i are constructed for all nonterminals
	// A using the rule: If GOTO(I???, A) = I???, then GOTO[i, A] = j.

	newState := slr.lr0.Next(state, symbol)

	if newState == "" {
		return "", fmt.Errorf("GOTO[%q, %q] is an error entry", state, symbol)
	}
	return newState, nil
}

func (slr *slrTable) Action(i, a string) LRAction {
	// step 2 of algorithm 4.46, "Constructing an SLR-parsing table", for
	// reference

	// 2. State i is constructed from I???. The parsing actions for state i are
	// determined as follows:

	// get our set back from current state so we can check it; this is our I???
	itemSet := slr.lr0.GetValue(i)

	// we have gauranteed that these dont conflict during construction; still,
	// check it so we can panic if it conflicts
	var alreadySet bool
	var act LRAction

	// Okay, "[some random item] is in I???" is suuuuuuuuper vague. We're
	// basically going to have to check each item and see if it is in the
	// pattern. I *guess* ::::/
	for itemStr := range itemSet {
		item := slr.itemCache[itemStr]

		// given item is [A -> ??.??]:
		A := item.NonTerminal
		alpha := item.Left
		beta := item.Right

		var followA box.ISet[string]
		if A != slr.gPrime.StartSymbol() {
			// we'll need this later, glub 38)
			followA = slr.gPrime.FOLLOW(A)
		}

		// (a) If [A -> ??.a??] is in I??? and GOTO(I???, a) = I???, then set
		// ACTION[i, a] to "shift j." Here a must be a terminal.
		//
		// we'll assume ?? can be ??.
		// ?? can also be ?? but note this ?? is rly ??[1:] from earlier notation
		// used to assign beta (beta := item.Right).
		if slr.gPrime.IsTerminal(a) && len(beta) > 0 && beta[0] == a {
			j, err := slr.Goto(i, a)

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
					if slr.allowAmbig {
						// this is fine, resolve in favor of shift
						act = shiftAct
					} else {
						panic(fmt.Sprintf("grammar is not SLR(1): %s", makeLRConflictError(act, shiftAct, a).Error()))
					}
				} else {
					act = shiftAct
					alreadySet = true
				}
			}
		}

		// (b) If [A -> ??.] is in I???, then set ACTION[i, a] to "reduce A -> ??"
		// for all a in FOLLOW(A); here A may not be S'.
		//
		// we'll assume ?? can be empty.
		// the beta we previously retrieved MUST be empty
		if len(beta) == 0 && A != slr.gPrime.StartSymbol() && followA.Has(a) {
			reduceAct := LRAction{Type: LRReduce, Symbol: A, Production: grammar.Production(alpha)}
			if alreadySet && !reduceAct.Equal(act) {
				if isSRConflict, _ := isShiftReduceConlict(act, reduceAct); isSRConflict && slr.allowAmbig {
					// we are in a shift/reduce conflict; the prior def is a
					// shift. resolve this in favor of shift by simply not assigning
					// the new one.
				} else {
					panic(fmt.Sprintf("grammar is not SLR(1): %s", makeLRConflictError(act, reduceAct, a).Error()))
				}
			} else {
				act = reduceAct
				alreadySet = true
			}
		}

		// (c) If [S' -> S.] is in I???, then set ACTION[i, $] to "accept".
		if a == "$" && A == slr.gPrime.StartSymbol() && len(alpha) == 1 && alpha[0] == slr.gStart && len(beta) == 0 {
			acceptAct := LRAction{Type: LRAccept}
			if alreadySet && !acceptAct.Equal(act) {
				panic(fmt.Sprintf("grammar is not SLR(1): %s", makeLRConflictError(act, acceptAct, a).Error()))
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
