package grammar

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/box"
	. "github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
)

func (g Grammar) ReachableFrom(start string, end string) (bool, []Pair[string, Production]) {
	if !g.IsNonTerminal(start) {
		return false, nil
	}
	if !g.IsNonTerminal(end) && !g.IsTerminal(end) {
		return false, nil
	}

	// run reachability algorithm, but instead of starting at the start symbol,
	// start with each production of it.

	reached := box.NewSVSet[slices.LList[Pair[string, Production]]]()

	r := g.Rule(start)
	for _, p := range r.Productions {
		for _, sym := range p {
			var path slices.LList[Pair[string, Production]]
			path = path.Add(PairOf(start, p))

			if sym == end {
				return true, path.Slice()
			}
			reached.Add(sym)
			reached.Set(sym, path)
		}
	}

	updated := true
	for updated {
		updated = false
		for k := range reached {
			rule := g.Rule(k)
			if rule.NonTerminal != k {
				// terminal; don't check it
				continue
			}
			for _, prod := range rule.Productions {
				for _, sym := range prod {
					var path = reached.Get(k)
					path = path.Add(PairOf(k, prod))

					if sym == end {
						return true, path.Slice()
					}
					if !reached.Has(sym) {
						reached.Add(sym)
						reached.Set(sym, path)
						updated = true
					}
				}
			}
		}
	}

	return false, nil
}

func (g Grammar) ReachesMultiple(start string, targets ...ReachGoal) (bool, [][]Pair[string, Production]) {
	if len(targets) == 0 {
		panic("no targets given")
	}

	type entry struct {
		Symbol    string
		Path      slices.LList[Pair[string, Production]]
		Solutions []reachSol
	}

	reached := []entry{}

	// start with each production of the start symbol

	r := g.Rule(start)
	for _, p := range r.Productions {
		for _, sym := range p {
			var path slices.LList[Pair[string, Production]]
			path = path.Add(PairOf(start, p))

			ent := entry{
				Symbol: sym,
				Path:   path,
			}

			for i, t := range targets {
				if t.Matches(sym) {
					var sol reachSol
					sol.Goal = i
					sol.Symbol = sym
					sol.Derivation = path
					ent.Solutions = append(ent.Solutions, sol)
				}
			}

			if len(ent.Solutions) == len(targets) {
				// we have a solution for each target. convert to returnable
				// format and return.
				paths := make([][]Pair[string, Production], len(targets))
				for _, sol := range ent.Solutions {
					paths[sol.Goal] = sol.Derivation.Slice()
				}
				return true, paths
			}

			if g.IsNonTerminal(ent.Symbol) {
				// keep non-terminals only
				reached = append(reached, ent)
			}
		}
	}

	// TODO: PRIOR TO TESTING MAKE SHORE SOLUTION IS BEING ADDED AS IT SEES THE
	// SYMBOL NOT WHEN IT ACTUALLY ADDS THAT SYMBOL; SOLVED SYMBOLS SHOULD BE
	// SKIPPED. (but also we should add it as a "not solved" one where it is
	// followed as well)

	// need to maintain a seperate map for fast lookup
	type derivEntryPath struct {
		sym       string
		sols      []reachSol
		derivPath *slices.LList[Pair[string, Production]]
	}
	var keptDerivations = make(map[string][]derivEntryPath)

	updated := true
	for updated {
		updated = false
		nextReachList := []entry{}
		for _, ent := range reached {
			rule := g.Rule(ent.Symbol)

			// add each production of the symbol to the path
			for _, p := range rule.Productions {
				for _, sym := range p {
					var path slices.LList[Pair[string, Production]]
					path = path.Add(PairOf(start, p))

					ent := entry{
						Symbol: sym,
						Path:   path,
					}

					for i, t := range targets {
						if t.Matches(sym) {
							var sol reachSol
							sol.Goal = i
							sol.Symbol = sym
							sol.Derivation = path
							ent.Solutions = append(ent.Solutions, sol)
						}
					}

					if len(ent.Solutions) == len(targets) {
						// we have a solution for each target. convert to returnable
						// format and return.
						paths := make([][]Pair[string, Production], len(targets))
						for _, sol := range ent.Solutions {
							paths[sol.Goal] = sol.Derivation.Slice()
						}
						return true, paths
					}

					// keep non-terminals only
					if g.IsTerminal(ent.Symbol) {
						continue
					}

					addEntry := false
					addNewEntryToReached := false
					// if we already have a path to this symbol, use the
					// following criteria to determine what to do with new entry:
					kept, haveExistingPath := keptDerivations[ent.Symbol]
					if haveExistingPath {
						for i, oldEntry := range kept {
							oldHasUnique, newHasUnique := checkUniqueSols(oldEntry.sols, ent.Solutions)

							// 1. if the new entry's solution set is the same as the old
							//    entry for that symbol:
							if !oldHasUnique && !newHasUnique {
								if oldEntry.derivPath.Len() > ent.Path.Len() {
									//    - if the new entry has a shorter path, replace the old
									//    seen entry with the new one and update all other
									//    entries which previously used the old entry to use
									//    this one as it is strictly better.

									// to accomplish this, we'll set add entry as true but not addToReached
									addEntry = true
									addNewEntryToReached = false

									newDeriveEntry := oldEntry

									// update the POINTED to derive path linked list
									// entry, which automatically updates everyfin that points to it.
									*newDeriveEntry.derivPath = ent.Path

									// create new map entry
									newKept := make([]derivEntryPath, len(kept))
									copy(newKept, kept)
									newKept[i] = newDeriveEntry
									keptDerivations[ent.Symbol] = newKept
									break
								} else {
									// if the new entry has an equal or longer path, do not add it and check no more
									// entries for this symbol.
									addEntry = false
									addNewEntryToReached = false
									break
								}
							}

							// 2. if the new entry's solution set contains targets not
							//	  already reached by the old entry, and the old entry
							//	  contains targets not reached by the new entry, keep
							//    both. However, check against other kept entries as well.
							if oldHasUnique && newHasUnique {
								addEntry = true
								addNewEntryToReached = true
							}

							// 3. if the new entry's solution set contains targets not
							//	  already reached by the old entry, and the old entry
							//	  contains no targets not reached by the new entry,
							// 	  replace the old entry with the new one and update all
							//    other entries which previously used the old entry to
							//	  use this one.
							if !oldHasUnique && newHasUnique {
								addEntry = true
								addNewEntryToReached = false

								newDeriveEntry := oldEntry

								// update the POINTED to derive path linked list
								// entry, which automatically updates everyfin that points to it.
								*newDeriveEntry.derivPath = ent.Path

								// create new map entry
								newKept := make([]derivEntryPath, len(kept))
								copy(newKept, kept)
								newKept[i] = newDeriveEntry
								keptDerivations[ent.Symbol] = newKept
								break
							}

							// 4. if the new entry's solution set contains no targets
							//	  not already reached by the old entry, do nothing. The
							//    old entry is strictly better. check no other entries;
							// 	  do not add the new entry.
							if oldHasUnique && !newHasUnique {
								addEntry = false
								addNewEntryToReached = false
								break
							}
						}
					} else {
						addEntry = true
						addNewEntryToReached = true
					}

					if addEntry {
						updated = true
						nextReachList = append(nextReachList, ent)

						if addNewEntryToReached {
							oldEnts, ok := keptDerivations[ent.Symbol]
							if !ok {
								oldEnts = make([]derivEntryPath, 0)
							}
							oldEnts = append(oldEnts, derivEntryPath{
								derivPath: &ent.Path,
								sols:      ent.Solutions,
								sym:       ent.Symbol,
							})
							keptDerivations[ent.Symbol] = oldEnts
						}
					}

					reached = nextReachList
				}
			}
		}
	}

	return false, nil

}

func checkUniqueSols(left, right []reachSol) (leftHasUnique, rightHasUnique bool) {
	// special case check
	if len(left) == 0 {
		if len(right) == 0 {
			return false, false
		} else {
			return false, true
		}
	} else if len(right) == 0 && len(left) > 0 {
		return true, false
	}
	inBoth := map[int]bool{}

	for _, l := range left {
		for _, r := range right {
			if l.Goal == r.Goal {
				inBoth[l.Goal] = true
			} else {
				leftHasUnique = true
			}
		}
	}

	// now check the right
	for _, r := range right {
		if _, ok := inBoth[r.Goal]; !ok {
			rightHasUnique = true
			break
		}
	}

	return leftHasUnique, rightHasUnique
}

type ReachGoal struct {
	symbols []string
}

func ReachAny(s ...string) ReachGoal {
	if len(s) == 0 {
		panic("no symbols given")
	}

	r := ReachGoal{symbols: make([]string, len(s))}
	copy(r.symbols, s)
	return r
}

func ReachSymbol(s string) ReachGoal {
	return ReachGoal{symbols: []string{s}}
}

func (r ReachGoal) Matches(s string) bool {
	for _, sym := range r.symbols {
		if sym == s {
			return true
		}
	}
	return false
}

// solution in a reachability problem.
type reachSol struct {
	// Goal is the index of the sought-after term that this solution reaches.
	Goal int

	// Symbol is the specific name of the symbol that this solution reaches.
	Symbol string

	// Derivation is the sequence of derivation steps that this solution uses to
	// go fron the goal to the start.
	Derivation slices.LList[Pair[string, Production]]
}

func (rs reachSol) String() string {
	var sb strings.Builder
	sb.WriteString("{Found (")
	sb.WriteString(fmt.Sprintf("%d", rs.Goal))
	sb.WriteString(", ")
	sb.WriteString(rs.Symbol)
	sb.WriteString("); [")
	der := rs.Derivation.Slice()
	for i, p := range der {
		sb.WriteRune('(')
		sb.WriteString(p.First)
		sb.WriteString(" -> ")
		sb.WriteString(p.Second.String())
		sb.WriteRune(')')
		if i+1 < len(der) {
			sb.WriteString(", ")
		}
	}
	sb.WriteString("]}")

	return sb.String()
}
