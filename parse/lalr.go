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

// EmptyLALR1Parser returns a completely empty LALR1Parser, unsuitable for use.
// Generally this should not be used directly except for internal purposes; use
// GenerateLALR1Parser to generate one ready for use
func EmptyLALR1Parser() *lrParser {
	return &lrParser{table: &lalr1Table{}, parseType: types.ParserLALR1}
}

// computeLALR1Kernels computes LALR(1) kernels for grammar g, which must NOT be
// an augmented grammar.
//
// This is an implementation of Algorithm 4.63, "Efficient computation of the
// kernels of the LALR(1) collection of sets of items" from purple dragon book.
func computeLALR1Kernels(g grammar.Grammar) box.SVSet[box.SVSet[grammar.LR1Item]] {
	// we'll also need to know what our start rule and augmented start rules are.
	startSym := g.StartSymbol()
	startSymPrime := g.Augmented().StartSymbol()
	gPrimeStartItem := grammar.LR0Item{NonTerminal: startSymPrime, Right: []string{startSym}}
	gPrimeStartKernel := box.NewSVSet[grammar.LR0Item]()
	gPrimeStartKernel.Set(gPrimeStartItem.String(), gPrimeStartItem)

	gTerminals := g.Terminals()

	// 1. Construct the kernels of the sets of LR(O) items for G.
	lr0Kernels := getLR0Kernels(g)

	calcSponts := map[stateAndItemStr]box.StringSet{}
	calcProps := map[stateAndItemStr][]stateAndItemStr{}

	// special case, lookahead $ is always generated spontaneously for the item
	// S' -> .S in the initial set of items
	calcSponts[stateAndItemStr{state: gPrimeStartKernel.String(), item: gPrimeStartItem.String()}] = box.StringSetOf([]string{"$"})

	for _, lr0KernelName := range lr0Kernels.Elements() {
		IKernelSet := lr0Kernels.Get(lr0KernelName)

		if IKernelSet.Equal(box.StringSetOf([]string{"S-P -> . S"})) {
			fmt.Printf("make debugger do thing\n")
		}

		for _, X := range gTerminals {
			// 2. Apply algorithm 4.62 to the kernel of set of LR(0) items and
			// grammar symbol X to determine which lookaheads are spontaneously
			// generated for kernel items in GOTO(I, X), and from which items in
			// I lookaheads are propagated to kernel items in GOTO(I, X).
			sponts, props := determineLookaheads(g.Augmented(), IKernelSet, X)

			// add them to our pre-calced slice for later use in lookahead
			// table
			for k := range sponts {
				sponSet := sponts[k]
				existing, ok := calcSponts[k]
				if !ok {
					existing = box.NewStringSet()
				}
				existing.AddAll(sponSet)
				calcSponts[k] = existing
			}
			for k := range props {
				propSlice := props[k]
				existing, ok := calcProps[k]
				if !ok {
					existing = make([]stateAndItemStr, 0)
				}
				for i := range propSlice {
					existing = append(existing, propSlice[i])
				}
				calcProps[k] = existing
			}
		}
	}

	// 3. Initialize a table that gives, for each kernel item in each set of
	// items, the associated lookaheads. Initially, each item has associated
	// with it only those lookaheads that we determined in step (2) were
	// generated spontaneously

	// this table holds a slice of passes, each of which map a
	// {LR0Item}.OrderedString() to a slice of passes. Each pass is a
	// slice of the lookaheads found on that pass. Pass 0, aka "INIT" pass in
	// purple dragon book, is the spontaneously generated lookaheads for the
	// item; all other passes are the propagation checks.
	lookaheadCalcTable := []map[stateAndItemStr]box.StringSet{}
	initPass := map[stateAndItemStr]box.StringSet{}
	for k := range calcSponts {
		sponts := calcSponts[k]
		elemSet := box.NewStringSet()
		for _, terminal := range sponts.Elements() {
			elemSet.Add(terminal)
		}
		initPass[k] = elemSet
	}
	lookaheadCalcTable = append(lookaheadCalcTable, initPass)

	/*
		// 4. Make repeated passes over the kernel items in all sets. When we visit
		// an item i, we look up the kernel items to which i propagates its
		// lookaheads, using information tabulated in step (2). The current set of
		// lookaheads for i is added to those already associated with each of the
		// items to which i propagates its lookaheads. We continue making passes
		// over the kernel items until no more new lookaheads are propagated.
		updated := true
		passNum := 1
		for updated {
			updated = false

			prevColumn := lookaheadCalcTable[passNum-1]
			curColumn := map[stateAndItemStr]util.StringSet{}

			// initialy set everyfin to prior column
			for k := range prevColumn {
				curColumn[k] = util.NewStringSet(prevColumn[k])
			}

			for _, lr0KernelName := range lr0Kernels.Elements() {
				IKernelSet := lr0Kernels.Get(lr0KernelName)
				// When we visit an item i, we look up the kernel items to which i
				// propagates its lookaheads, using information tabulated in step
				// (2).
				propagateTo := calcProps[IKernelSet.StringOrdered()]

				// The current set of lookaheads for i is added to those already
				// associated with each of the items to which i propagates its
				// lookaheads.
				curLookaheads := prevColumn[IKernelSet.StringOrdered()]
				for _, toName := range propagateTo.Elements() {
					for _, la := range curLookaheads.Elements() {
						if !curColumn[toName].Has(la) {
							propDest := curColumn[toName]
							propDest.Add(la)
							curColumn[toName] = propDest
							updated = true
						}
					}
				}
			}

			lookaheadCalcTable = append(lookaheadCalcTable, curColumn)
			passNum++
		}*/

	// now collect the final table info into the final result
	//finalPass := lookaheadCalcTable[len(lookaheadCalcTable)-1]
	lalrKernels := box.NewSVSet[box.SVSet[grammar.LR1Item]]()

	// TODO: actually convert the table results to this.
	return lalrKernels

}

type stateAndItemStr struct {
	state string
	item  string
}

// determineLookaheads finds the lookaheads spontaneously generated by items in
// I for kernel items in GOTO(I, X) (jello: g.LR1_GOTO) and the items in I from
// which lookaheads are propagated to kernel items in GOTO(I, X).
//
// g must be an augmented grammar.
// K is the kernel of a set of LR(0) items I. X is a grammar symbol. Returns the
// LALR(1) kernel set generated from the LR(0) item kernel set.
//
// This is an implementation of Algorithm 4.62, "Determining lookaheads", from
// purple dragon book.
//
// "There are two ways a lookahead b can get attached to an LR(0) item
// [B -> ??.??] in some set of LALR(1) items J:"
//
// 1. There is a set of items I, with a kernel item [A -> ??.??, a], and J =
// GOTO(I, X), and the construction of
//
//	GOTO(CLOSURE({[A -> ??.??, a]}), X)
//
// as given in Fig. 4.40 (jello: implemented in g.LR1_CLOSURE and
// g.LR1_GOTO), contains [B -> ??.??, b], regardless of a. Such a lookahead is
// said to be generated *spontaneously* for B -> ??.??.
//
// 2. As a special case, lookahead $ is generated spontaneously for the item
// [S' -> .S] in the initial set of items.
//
// 3. All as (1), but a = b, and GOTO(CLOSURE({[A -> ??.??, b]}), X), as given
// in Fig. 4.40 (jello: again, g.LR1_CLOSURE and g.LR1_GOTO), contains
// [B -> ??.??, b] only because A -> ??.?? has b as one of its associated
// lookaheads. In such a case, we say that lookaheads *propagate* from
// A -> ??.?? in the kernel of I to B -> ??.?? in the kernel of J. Note that
// propagation does not depend on the particular lookahead symbol; either
// all lookaheads propagate from one item to another, or none do.
func determineLookaheads(g grammar.Grammar, K box.SVSet[grammar.LR0Item], X string) (spontaneous map[stateAndItemStr]box.StringSet, propagated map[stateAndItemStr][]stateAndItemStr) {
	// note: '#' in notes stands for any symbol not in the grammar at hand. We
	// will use Grammar.GenerateUniqueName to get one not currently used, and as
	// we require g to be augmented, this should give us somefin OTHER than the
	// added start production.
	nonGrammarSym := g.GenerateUniqueTerminal("#")

	if K.Equal(box.StringSetOf([]string{"S-P -> . S"})) {
		fmt.Printf("make debugger do thing\n")
	}

	spontaneous = map[stateAndItemStr]box.StringSet{}
	propagated = map[stateAndItemStr][]stateAndItemStr{}

	// GOTO will be needed elsewhere
	GOTO_I_X := g.LR0_GOTO(g.LR0_CLOSURE(K), X)

	if GOTO_I_X.Empty() {
		return spontaneous, propagated
	}

	// for ( each item A -> ??.?? in K ) {
	for _, aItemName := range K.Elements() {
		aItem := K.Get(aItemName)

		// J := CLOSURE({[A -> ??.??, #]})
		lr1StartItem := grammar.LR1Item{LR0Item: aItem, Lookahead: nonGrammarSym}
		lr1StartKernels := box.NewSVSet[grammar.LR1Item]()
		lr1StartKernels.Set(lr1StartItem.String(), lr1StartItem)
		J := g.LR1_CLOSURE(lr1StartKernels)

		TRUE_GOTO_I_X := g.LR1_GOTO(J, X)

		// next parts tell us to check condition based on some lookahead in
		// [B -> ??.X??, a] of J ...soooooooo in other words, check all of the
		// items in J
		for _, bItemName := range J.Elements() {
			bItem := J.Get(bItemName)

			newLeft := make([]string, len(bItem.Left))
			copy(newLeft, bItem.Left)

			var newRight []string
			if len(bItem.Right) > 0 {
				newRight = make([]string, len(bItem.Right)-1)
				copy(newRight, bItem.Right[1:])
				newLeft = append(newLeft, bItem.Right[0])
			}

			// shifted item is our [B -> ??X.??]. note that the dot has moved one
			// symbol to the right
			shiftedLR0Item := grammar.LR0Item{
				NonTerminal: bItem.NonTerminal,
				Left:        newLeft,
				Right:       newRight,
			}

			// slightly more complex logic to go through all of TRUE_GOTO
			// and find all items that have the same LR0 as our shifted one
			prodInGoto := false
			for _, elemName := range TRUE_GOTO_I_X.Elements() {
				lr1Item := TRUE_GOTO_I_X.Get(elemName)
				if lr1Item.LR0Item.Equal(shiftedLR0Item) {
					prodInGoto = true
					break
				}
			}
			if !prodInGoto {
				shiftedItemStr := shiftedLR0Item.String()
				fmt.Println(shiftedItemStr)
				continue
			}

			if bItem.Lookahead != nonGrammarSym {
				// if ( [B -> ??.X??, a] is in J, and a is not # )

				// conclude that lookahead a is spontaneously generated for item
				// B -> ??X.?? in GOTO(I, X).
				newItem := grammar.LR1Item{
					LR0Item:   shiftedLR0Item,
					Lookahead: bItem.Lookahead,
				}

				key := stateAndItemStr{
					state: GOTO_I_X.StringOrdered(),
					item:  newItem.LR0Item.String(),
				}

				spontSet, ok := spontaneous[key]
				if !ok {
					spontSet = box.NewStringSet()
				}
				spontSet.Add(bItem.Lookahead)

				spontaneous[key] = spontSet
			} else {
				// if ( [B -> ??.X??, #] is in J )

				// conclude that lookaheads propagate from A -> ??.?? in I to
				// B -> ??X.?? in GOTO(I, X).

				from := stateAndItemStr{
					state: K.StringOrdered(),
					item:  aItem.String(),
				}

				to := stateAndItemStr{
					state: GOTO_I_X.StringOrdered(),
					item:  shiftedLR0Item.String(),
				}

				existingPropagated, ok := propagated[from]
				if !ok {
					existingPropagated = []stateAndItemStr{}
				}
				existingPropagated = append(existingPropagated, to)
				propagated[from] = existingPropagated
			}

		}
	}

	return spontaneous, propagated
}

// g must NOT be an augmented grammar.
func getLR0Kernels(g grammar.Grammar) box.VSet[string, box.SVSet[grammar.LR0Item]] {
	gPrime := g.Augmented()
	itemSets := gPrime.CanonicalLR0Items()

	kernels := box.SVSet[box.SVSet[grammar.LR0Item]]{}

	// okay, now for each state pull out the kernels
	for _, s := range itemSets.Elements() {
		stateVal := itemSets.Get(s)

		kernelItems := box.SVSet[grammar.LR0Item]{}
		for _, stateItemName := range stateVal.Elements() {
			stateItem := stateVal.Get(stateItemName)
			if len(stateItem.Left) > 0 || (len(stateItem.Right) == 1 && stateItem.Right[0] == g.StartSymbol() && stateItem.NonTerminal == gPrime.StartSymbol()) {
				kernelItems.Set(stateItemName, stateItem)
			}
		}
		kernels.Set(kernelItems.StringOrdered(), kernelItems)
	}

	return kernels
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
	data := decbin.EncBinary(lalr1.gPrime)
	data = append(data, decbin.EncString(lalr1.gStart)...)
	data = append(data, decbin.EncMapStringToBinary(lalr1.itemCache)...)
	data = append(data, decbin.EncSliceString(lalr1.gTerms)...)
	data = append(data, decbin.EncSliceString(lalr1.gNonTerms)...)
	data = append(data, decbin.EncBool(lalr1.allowAmbig)...)

	// now the long part, the dfa
	dfaBytes := lalr1.dfa.MarshalBytes(func(s box.SVSet[grammar.LR1Item]) []byte {
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

func (lalr1 *lalr1Table) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	n, err = decbin.DecBinary(data, &lalr1.gPrime)
	if err != nil {
		return fmt.Errorf(".gPrime: %w", err)
	}
	data = data[n:]

	lalr1.gStart, n, err = decbin.DecString(data)
	if err != nil {
		return fmt.Errorf(".gStart: %w", err)
	}
	data = data[n:]

	var ptrMap map[string]*grammar.LR1Item
	ptrMap, n, err = decbin.DecMapStringToBinary[*grammar.LR1Item](data)
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

	lalr1.gTerms, n, err = decbin.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".gTerms: %w", err)
	}
	data = data[n:]

	lalr1.gNonTerms, n, err = decbin.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".gNonTerms: %w", err)
	}
	data = data[n:]

	lalr1.allowAmbig, n, err = decbin.DecBool(data)
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

	lalr1.dfa, err = automaton.UnmarshalDFABytes(dfaBytes, func(b []byte) (box.SVSet[grammar.LR1Item], error) {
		var innerN int
		var innerErr error
		var numEntries int
		set := box.NewSVSet[grammar.LR1Item]()

		numEntries, innerN, innerErr = decbin.DecInt(b)
		if innerErr != nil {
			return nil, fmt.Errorf("get entry count: %w", innerErr)
		}
		b = b[innerN:]
		for i := 0; i < numEntries; i++ {
			var k string
			var v grammar.LR1Item

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

	// 2. State i is constructed from I???. The parsing actions for state i are
	// determined as follows:

	// (a) If [A -> ??.a??, b] is in I??? and GOTO(I???, a) = I???, then set
	// ACTION[i, a] to "shift j." Here a must be a terminal.

	// (b) If [A -> ??., a] is in I???, A != S', then set ACTION[i, a] to "reduce
	// A -> ??".

	// get our set back from current state so we can check it; this is our I???

	// get our set back from current state so we can check it; this is our I???
	itemSet := lalr1.dfa.GetValue(i)

	// we have gauranteed that these dont conflict during construction; still,
	// check it so we can panic if it conflicts
	var alreadySet bool
	var act LRAction

	// Okay, "[some random item] is in I???" is suuuuuuuuper vague. We're
	// basically going to have to check each item and see if it is in the
	// pattern. I *guess* ::::/
	for itemStr := range itemSet {
		item := lalr1.itemCache[itemStr]

		// given item is [A -> ??.??, b]:
		A := item.NonTerminal
		alpha := item.Left
		beta := item.Right
		b := item.Lookahead

		// (a) If [A -> ??.a??, b] is in I??? and GOTO(I???, a) = I???, then set
		// ACTION[i, a] to "shift j." Here a must be a terminal.
		//
		// we'll assume ?? can be ??.
		// ?? can also be ?? but note this ?? is rly ??[1:] from earlier notation
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

		// (b) If [A -> ??., a] is in I???, A != S', then set ACTION[i, a] to
		// "reduce A -> ??".
		//
		// we'll assume ?? can be empty.
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

		// (c) If [S' -> S., $] is in I???, then set ACTION[i, $] to "accept".
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
