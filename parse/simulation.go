package parse

import (
	"fmt"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/lex"
)

// DeriveFullTree derives a parse tree based on the grammar that is guaranteed
// to contain every rule at least once. It is *not* garunteed to have each rule
// only once, although it will make a best effort to minimize duplications.
//
// The fakeValProducer map, if provided, will be used to assign a lexed
// value to each synthesized terminal node that has a class whose ID matches
// a key in it. If the map is not provided or does not contain a key for a
// token class matching a terminal being generated, a default string will be
// automatically generated for it.
func DeriveFullTree(g grammar.Grammar, fakeValProducer ...map[string]func() string) ([]ParseTree, error) {
	// create a function to get a value for any terminal from the
	// fakeValProducer, falling back on default behavior if none is provided or
	// if a token class is not found in the fakeValProducer.
	var lineNo int
	makeTermSource := func(forTerm string) lex.Token {
		class := g.Term(forTerm)
		val := fmt.Sprintf("<SIMULATED %s>", class.ID())
		if len(fakeValProducer) > 0 && fakeValProducer[0] != nil {
			if fvp, ok := fakeValProducer[0][class.ID()]; ok {
				val = fvp()
			}
		}
		lineNo++
		t := lex.NewToken(class, val, 11, lineNo, fmt.Sprintf("<fakeLine>%s</fakeLine>", val))
		return t
	}

	// we need to get a list of all productions that we have yet to create
	toCreate := map[string][]grammar.Production{}
	uncoveredSymbols := box.NewStringSet()
	for _, nonTerm := range g.NonTerminals() {
		rule := g.Rule(nonTerm)
		prods := make([]grammar.Production, len(rule.Productions))
		copy(prods, rule.Productions)
		toCreate[nonTerm] = prods
		uncoveredSymbols.Add(nonTerm)
	}

	// we also need to get the 'shortest' production for each non-terminal; this
	// is the production that eventually results in the fewest number of
	// non-terminals being used in the final parse tree.
	shortestDerivation, err := createFewestNonTermsAlternationsTable(g)
	if err != nil {
		return nil, fmt.Errorf("failed to get shortest derivation table for grammar: %w", err)
	}

	// put in all symbols that we are totally done with
	// TODO GHI #90: we can probs replace using both coveredSymbols and uncoveredSymbols
	// by only using uncoveredSymbols; it's the only one we will iterate on.
	coveredSymbols := box.NewStringSet()

	// this function will be called for non-terminals who have fully cleared
	// their productions from toCreate during an iteration.
	CLEARED_NT_IS_COVERED := func(checkNT string) bool {
		// if it has ever been covered, it is covered
		if coveredSymbols.Has(checkNT) {
			return true
		}

		ntRule := g.Rule(checkNT)
		for _, p := range ntRule.Productions {
			for _, sym := range p {
				if g.IsTerminal(sym) {
					// terminals are not considered as part of the check, as
					// they will be covered by the non-terminal having done so
					// in clearing its toCreate entry.
					continue
				}

				if coveredSymbols.Has(sym) {
					// this symbol is covered, so we can skip it
					continue
				}

				// ... but dont return not covered if the symbol can recurse
				// back to checkNT
				if sym == checkNT {
					// this symbol can recurse back to NT. do not check it
					// for coverage, as it will be covered by nature of
					// being higher up in the tree.
					continue
				}

				if reachable, _ := g.ReachableFrom(sym, checkNT); reachable {
					continue
				}

				// if we are here, then the sym is a non-terminal not in
				// coveredSymbols that will not recurse to checkNT. checkNT is
				// not covered.
				return false
			}
		}

		return true
	}

	finalTrees := []ParseTree{}

	for len(toCreate) > 0 {
		root := &ParseTree{
			Terminal: false,
			Value:    g.StartSymbol(),
		}
		treeStack := box.NewStack([]*ParseTree{root})

		for treeStack.Len() > 0 {
			pt := treeStack.Pop()

			if pt.Terminal {
				continue
			}

			// select the production to do
			options, ok := toCreate[pt.Value]
			if !ok {
				// we have already created all the productions for this non-terminal
				// so we need to select the production that results in the fewest
				// non-terminals being used and be done.
				shortest := deriveShortestTree(g, pt.Value, shortestDerivation, makeTermSource)
				pt.Children = shortest.Children
				continue
			}

			option := options[0]
			if option.Equal(grammar.Epsilon) {
				// epsilon prod must be specifically added to done here since it
				// would not be added to the done list otherwise
				pt.Children = []*ParseTree{{Terminal: true}}
			} else {
				pt.Children = make([]*ParseTree, len(option))
				for i, sym := range option {
					if g.IsTerminal(sym) {
						pt.Children[i] = &ParseTree{
							Value:    sym,
							Terminal: true,
							Source:   makeTermSource(sym),
						}
					} else {
						pt.Children[i] = &ParseTree{
							Value: sym,
						}
					}
				}
			}

			// push children onto stack in reverse order so they are popped in
			// the correct order (left to right)
			for i := len(pt.Children) - 1; i >= 0; i-- {
				treeStack.Push(pt.Children[i])
			}

			toCreate[pt.Value] = options[1:]
			if len(toCreate[pt.Value]) == 0 {
				delete(toCreate, pt.Value)
				if CLEARED_NT_IS_COVERED(pt.Value) {
					uncoveredSymbols.Remove(pt.Value)
					coveredSymbols.Add(pt.Value)
				}
			}

		}
		// add the root to our list of final trees
		finalTrees = append(finalTrees, *root)

		// now here's the reel trick; if we didn't actually end up covering all
		// productions of toCreate, we will re-populate it... but ONLY if
		// toCreate has at least one remaining entry that is not in our
		// "covered" table.
		//
		// Additionally, we will not repopulate the ones that were incomplete,
		// so next time through we are guaranteed to hit the next item.
		if len(toCreate) > 0 {
			// prior to running checks, run through our list of uncoveredSymbols
			// that have been cleared in this iteration and see if they are NOW
			// covered. We will need to retry on each iteration.
			clearedButUncovered := box.NewStringSet()
			for _, nt := range uncoveredSymbols.Elements() {
				if _, ok := toCreate[nt]; !ok {
					// if it's not currently covered but it is no longer in toCreate,
					// then its CLEARED_NT_IS_COVERED check was false. but more
					// may have happened since then, so we need to check it
					// again.
					//
					// TODO GHI #90: actually we probably don't need the original check
					// at all. This would cover it. But we could retain cleared
					// and update THAT as we go.
					clearedButUncovered.Add(nt)
				}
			}
			updatedClearedButUncovered := true
			for updatedClearedButUncovered {
				updatedClearedButUncovered = false
				for _, nt := range clearedButUncovered.Elements() {
					if CLEARED_NT_IS_COVERED(nt) {
						coveredSymbols.Add(nt)
						uncoveredSymbols.Remove(nt)
						clearedButUncovered.Remove(nt)
						updatedClearedButUncovered = true
					}
				}
			}

			// now everyfin should be current. proceed with normal check.

			stillUncovered := box.NewStringSet()

			for sym := range toCreate {
				// if sym is in the coveredSymbols, then it *is* complete,
				// actually; just from a prior iteration
				if !coveredSymbols.Has(sym) {
					stillUncovered.Add(sym)
				}
			}

			if stillUncovered.Empty() {
				// if we have nothing in stillUncovered, then we have covered all
				// items across all iterations, so we are done. set toCreate to
				// empty so we exit the loop.
				toCreate = map[string][]grammar.Production{}
			} else {
				// repopulate all items that were covered in some prior iteration
				for _, nonTerm := range g.NonTerminals() {
					if !stillUncovered.Has(nonTerm) {
						rule := g.Rule(nonTerm)
						prods := make([]grammar.Production, len(rule.Productions))
						copy(prods, rule.Productions)
						toCreate[nonTerm] = prods
					}
				}
			}
		}
	}

	// clean up all trees that are subtrees and equal to other trees.
	updated := true
	for updated {
		updated = false
		removeItems := make([]int, 0)
		for i := 1; i < len(finalTrees); i += 2 {
			containsLeft, pathLeft := finalTrees[i-1].IsSubTreeOf(finalTrees[i])
			if containsLeft {
				if len(pathLeft) == 0 {
					// it is a subtree at root, aka identical. just keep one.
					removeItems = append(removeItems, i)
					updated = true
				} else {
					// first final tree is a subtree of the second, so we remove
					// it
					removeItems = append(removeItems, i-1)
				}
			} else {
				containsRight, pathRight := finalTrees[i].IsSubTreeOf(finalTrees[i-1])
				if containsRight {
					if len(pathRight) == 0 {
						// it is a subtree at root, aka identical. just keep one.
						removeItems = append(removeItems, i)
						updated = true
					} else {
						// second final tree is a subtree of the first, so we remove
						// it
						removeItems = append(removeItems, i)
					}
				}
			}

			// otherwise, don't remove either
		}

		if len(removeItems) > 0 {
			updatedFinal := make([]ParseTree, 0)
			for i := 0; i < len(finalTrees); i++ {
				if !slices.In(i, removeItems) {
					updatedFinal = append(updatedFinal, finalTrees[i])
				}
			}
			finalTrees = updatedFinal
			updated = true
		}
	}

	return finalTrees, nil
}

// deriveShortest returns the parse tree for the given symbol with the fewest
// possible number of nodes.
func deriveShortestTree(g grammar.Grammar, sym string, shortestDerivation map[string]grammar.Production, tokMaker func(term string) lex.Token) *ParseTree {
	root := &ParseTree{
		Value: sym,
	}

	treeStack := box.NewStack([]*ParseTree{root})

	for treeStack.Len() > 0 {
		pt := treeStack.Pop()

		if pt.Terminal {
			continue
		}

		option := shortestDerivation[pt.Value]
		if option.Equal(grammar.Epsilon) {
			pt.Children = []*ParseTree{{Terminal: true}}
		} else {
			pt.Children = make([]*ParseTree, len(option))
			for i, sym := range option {
				if g.IsTerminal(sym) {
					pt.Children[i] = &ParseTree{
						Value:    sym,
						Terminal: true,
						Source:   tokMaker(sym),
					}
				} else {
					pt.Children[i] = &ParseTree{
						Value: sym,
					}
				}
			}
		}

		// push children onto stack in reverse order so they are popped in
		// the correct order (left to right)
		for i := len(pt.Children) - 1; i >= 0; i-- {
			treeStack.Push(pt.Children[i])
		}
	}

	return root
}

// createFewestNonTermsAlternationsTable returns a map of non-terminals to the
// production in their associated rule that using to derive would result in the
// fewest new non-terminals being derived for that rule.
func createFewestNonTermsAlternationsTable(g grammar.Grammar) (map[string]grammar.Production, error) {
	// DEFINITION OF SCORE:
	// S(rule) = Floor(S(prod1) + S(prod2) + ... + S(prodN))
	// S(prod) = num N of non terminals in a production + S(nonterm1) + S(nonterm2) + ... + S(nontermN)

	// type to let us store either the int value or a reference to a score
	type valOrRef struct {
		val int
		ref string
	}

	// this will let us quickly check the minimum possible value for a calculation
	// and whether it is constant
	minPossibleVal := func(calcSlice []valOrRef) (int, bool) {
		valSoFar := 0
		constant := true
		for _, v := range calcSlice {
			if v.ref != "" {
				constant = false
			} else {
				valSoFar += v.val
			}
		}
		return valSoFar, constant
	}

	// make shore we are operating on a sane grammar
	err := g.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid grammar: %w", err)
	}

	// use this to "unconvert" a production from its string represntation
	prodStringToProd := map[string]grammar.Production{}

	// Create entries in a table for each non-terminal for which there is a rule
	// in the grammar, initially with no value set. Each item will be the score.
	// Additionally, create a "shortest alt" table from a rule to the production
	// that has the lowest score
	shortest := map[string]grammar.Production{}
	scores := map[string]int{}
	calcs := map[string]map[string][]valOrRef{}

	// For each production of each rule, create an entry in a calculations table
	// that specifies the calculation to carry out as specified above. Not all
	// will be fully solvable in this step. That's okay, we're just getting the
	// calculation steps.
	nts := g.NonTerminals()
	for _, nt := range nts {
		r := g.Rule(nt)
		ntMap, ok := calcs[nt]
		if !ok {
			ntMap = map[string][]valOrRef{}
			calcs[nt] = ntMap
		}
		for _, p := range r.Productions {
			calcEntry := []valOrRef{}
			for _, sym := range p {
				if g.IsNonTerminal(sym) {
					calcEntry = append(calcEntry, valOrRef{val: 1})
					calcEntry = append(calcEntry, valOrRef{ref: sym})
				}
			}

			// if the production is only terminals, then the score is 0
			if len(calcEntry) == 0 {
				calcEntry = append(calcEntry, valOrRef{val: 0})
			}

			prodStringToProd[p.String()] = p
			ntMap[p.String()] = calcEntry
		}
	}

	// Next, for each rule, if there are productions whose calculation is
	// self-referential, eliminate them because they will never be the smallest.
	// If this results in a rule being totally eliminated, the grammar has an
	// inescapable derivation cycle and should not be considered valid.
	for nt := range calcs {
		ntCalcs := calcs[nt]

		// find all the productions that are self-referential
		for prodStr, calc := range ntCalcs {
			for _, term := range calc {
				if term.ref == nt {
					delete(ntCalcs, prodStr)
					break
				}
			}
		}
		calcs[nt] = ntCalcs

		// check if this results in all prods being removed
		if len(ntCalcs) == 0 {
			return nil, fmt.Errorf("rule %s has an inescapable derivation cycle", nt)
		}
	}

	// Repeat the following until the calculations table is empty:
	for len(calcs) > 0 {
		modified := false

		// For each rule R in the calculations table, if there is a production
		// for it which has a single constant value that is indisputably the
		// lowest value, set that constant as the value of S(R) and the
		// alternation as the value of the rule in shortest table, then
		// eliminate all productions of R from the calculations table.
		for nt := range calcs {
			var remove bool
			ntCalcs := calcs[nt]

			for prodStr, calc := range ntCalcs {
				if len(calc) == 1 && calc[0].ref == "" {
					// we have a single constant value. check if it is the lowest
					minVal := calc[0].val
					isSmallest := true

					for prodStr2, calc2 := range ntCalcs {
						if prodStr2 == prodStr {
							// don't check self
							continue
						}

						// TODO GHI #91: efficiency. use otherIsConstant to immediately
						// move to THAT being the candidate instead of just
						// giving up.
						otherMinVal, _ := minPossibleVal(calc2)
						if otherMinVal < minVal {
							isSmallest = false
							break
						}
					}

					if isSmallest {
						// set that constant as the value of S(R) and the
						// alternation as the value of the rule in shortest
						// table
						scores[nt] = calc[0].val
						prod, ok := prodStringToProd[prodStr]
						if !ok {
							// should never happen
							panic(fmt.Sprintf("prod for %s not found in prodStringToProd", prodStr))
						}
						shortest[nt] = prod
						remove = true
					}
				}
				if remove {
					// we already found the smallest and know we need to remove
					// the nt, no need to keep looking
					break
				}
			}
			if remove {
				delete(calcs, nt)
				modified = true
			}
		}

		// For each entry in the calculations table, replace the value of all
		// known S(rule) expressions referred to with their actual value.
		for nt := range calcs {
			ntCalcs := calcs[nt]

			for _, calc := range ntCalcs {
				// Replace the value of all known S(rule) expressions referred
				// to with their actual value.
				for i := range calc {
					if calc[i].ref == "" {
						// constant value, nothing to do
						continue
					}

					refedScore, ok := scores[calc[i].ref]
					if !ok {
						// we don't know the score yet, so we can't replace it
						continue
					}

					// replace the ref with the score
					calc[i] = valOrRef{val: refedScore}
					modified = true
				}
			}
		}

		// Next, for each entry in the calculations table, if all items in the
		// calculation are constants, replace the entry with a single constant
		// that is the sum of all the constants.
		for nt := range calcs {
			ntCalcs := calcs[nt]

			for prodStr, calc := range ntCalcs {
				allConstants := true
				var sum int
				for i := range calc {
					if calc[i].ref != "" {
						allConstants = false
						break
					} else {
						sum += calc[i].val
					}
				}
				if allConstants {
					modified = true

					// replace the entry with a single constant that is the sum
					// of all the constants
					ntCalcs[prodStr] = []valOrRef{{val: sum}}
				}
			}
		}

		// cycle check; if a pass results in no changes, we have an indirect cycle
		// and cannot continue.
		if !modified {
			return nil, fmt.Errorf("indirect derivation cycle detected")
		}
	}

	return shortest, nil
}
