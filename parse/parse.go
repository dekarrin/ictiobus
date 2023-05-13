// Package parse provides parser construction and functionality. It contains
// everything needed to generate parsers based on LL(1), SLR(1), LR(1), or
// LALR(1) grammars, which are able to produce parse trees when given a stream
// of tokens as input.
package parse

import (
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
)

// IsLL1 returns whether the grammar is LL(1).
func IsLL1(g grammar.Grammar) bool {
	nts := g.NonTerminals()
	for _, A := range nts {
		AiRule := g.Rule(A)

		// we'll need this later, glubglub 38)
		followSetA := box.StringSetOf(findFOLLOWSet(g, A).Elements())

		// Whenever A -> α | β are two distinct productions of G:
		// -purple dragon book
		for i := range AiRule.Productions {
			for j := i + 1; j < len(AiRule.Productions); j++ {
				alphaFIRST := g.FIRST(AiRule.Productions[i][0])
				betaFIRST := g.FIRST(AiRule.Productions[j][0])

				aFSet := box.StringSetOf(alphaFIRST.Elements())
				bFSet := box.StringSetOf(betaFIRST.Elements())

				// 1. For no terminal a do both α and β derive strings beginning
				// with a.
				//
				// 2. At most of of α and β derive the empty string.
				//
				//
				// ...or in other words, FIRST(α) and FIRST(β) are disjoint
				// sets.
				// -purple dragon book

				if !aFSet.DisjointWith(bFSet) {
					return false
				}

				// 3. If β =*> ε, then α does not derive any string beginning
				// with a terminal in FOLLOW(A). Likewise, if α =*> ε, then β
				// does not derive any string beginning with a terminal in
				// FOLLOW(A).
				//
				//
				// ...or in other words, if ε is in FIRST(β), then FIRST(α) and
				// FOLLOW(A) are disjoint sets, and likewise if ε is in
				// FIRST(α).
				// -perple dergon berk. (Purple dragon book)
				if bFSet.Has(grammar.Epsilon[0]) {
					if !followSetA.DisjointWith(aFSet) {
						return false
					}
				}
				if aFSet.Has(grammar.Epsilon[0]) {
					if !followSetA.DisjointWith(bFSet) {
						return false
					}
				}
			}

		}
	}

	return true
}

// findFOLLOWSet is the used to get the findFOLLOWSet set of symbol X for generating
// various types of parsers.
func findFOLLOWSet(g grammar.Grammar, X string) box.Set[string] {
	return recursiveFOLLOWSet(g, X, box.NewStringSet())
}

func recursiveFOLLOWSet(g grammar.Grammar, X string, prevFollowChecks box.Set[string]) box.Set[string] {
	if X == "" {
		// there is no follow set. return empty set
		return box.NewStringSet()
	}
	followSet := box.NewStringSet()
	if X == g.StartSymbol() {
		followSet.Add("$")
	}

	A := g.NonTerminals()
	for i := range A {
		AiRule := g.Rule(A[i])

		for _, prod := range AiRule.Productions {
			if prod.HasSymbol(X) {
				// how many occurances of X are there? that says how many times
				// we need to do this, so find them
				var Xcount int
				for k := range prod {
					if prod[k] == X {
						Xcount++
					}
				}

				// do this for each occurance of X
				for Xoccurance := 0; Xoccurance < Xcount; Xoccurance++ {
					//alpha := []string{}
					beta := []string{}
					var doneWithAlpha bool
					var Xencounter int
					for k := range prod {
						if prod[k] == X {
							Xencounter++
							if Xencounter > Xoccurance && !doneWithAlpha {
								// only count this as end of alpha if we are at the
								// occurance of X we are looking for
								doneWithAlpha = true
								continue
							}
						}
						if !doneWithAlpha {
							//alpha = append(alpha, prod[k])
						} else {
							beta = append(beta, prod[k])
						}
					}

					// we now have our alpha, X, and beta

					// is there a FIRST in beta that isnt exclusively delta,
					// its firsts are in X's FOLLOW. Stop checking at the first
					// in beta that is NOT reducible to eps.
					for b := range beta {
						betaFirst := g.FIRST(beta[b])

						for _, k := range betaFirst.Elements() {
							if k != grammar.Epsilon[0] {
								followSet.Add(k)
							}
						}

						if !betaFirst.Has(grammar.Epsilon[0]) {
							// stop looping
							break
						}
					}

					// if X "can be" at the end of the production (i.e. if
					// either X is the final symbol of the production or if all
					// symbols following X are non-terminals with epsilon in
					// their FIRST sets), then FOLLOW(A) is in FOLLOW(X), where
					// A is the non-terminal producing X.
					canBeAtEnd := true
					for b := range beta {
						betaFirst := g.FIRST(beta[b])
						if !betaFirst.Has(grammar.Epsilon[0]) {
							canBeAtEnd = false
							break
						}
					}
					if canBeAtEnd {
						// dont infinitely recurse; if the producer is the
						// symbol, there's no need to add the FOLLOW from it bc
						// we are CURRENTLY calculating it.
						//
						// similarly, track the symbols we are going through.
						// don't recheck for the same one.
						if A[i] != X && !prevFollowChecks.Has(A[i]) {
							prevFollowChecks.Add(X)
							followA := recursiveFOLLOWSet(g, A[i], prevFollowChecks)
							for _, k := range followA.Elements() {
								followSet.Add(k)
							}
						}
					}
				}
			}
		}
	}

	return followSet
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
