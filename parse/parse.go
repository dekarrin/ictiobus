// Package parse provides language parsers for the ictiobus parser generator.
// These parsers implement [Parser] and are invoked by calling their Parse()
// method with a [lex.TokenStream] to get tokens from. The parser will read
// tokens from the stream and apply syntactic analysis to try and produce a
// parse tree, represented as a [Tree].
//
// This package currently provides an LL(1) parser, a Simple LR(1) parser, a
// Canonical LR(1) parser, and an LALR(1) parser, as well as the means to
// generate each from a context-free grammar describing the accepted language.
// The exact type of parser needed depends on the grammar.
package parse

import (
	"bufio"
	"encoding"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/dekarrin/ictiobus/automaton"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/rosed"
)

// A Parser represents an in-progress or ready-built parsing engine ready for
// use. It can be stored as a byte representation and retrieved from bytes as
// well.
type Parser interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	// Parse parses input text and returns the parse tree built from it, or a
	// SyntaxError with the description of the problem.
	Parse(stream lex.TokenStream) (Tree, error)

	// Type returns a string indicating what kind of parser was generated. This
	// will be "LL(1)", "SLR(1)", "CLR(1)", or "LALR(1)"
	Type() Algorithm

	// TableString returns the parsing table as a string.
	TableString() string

	// RegisterTraceListener sets up a function to call when an event occurs.
	// The events are determined by the individual parsers but involve
	// examination of the parser stack or other critical moments that may aid in
	// debugging.
	RegisterTraceListener(func(s string))

	// DFAString returns a string representation of the DFA for this parser, if one
	// so exists. Will return the empty string if the parser is not of the type
	// to have a DFA.
	DFAString() string

	// Grammar returns the grammar that this parser can parse.
	Grammar() grammar.CFG
}

// Algorithm is a classification of parsers in ictiobus.
type Algorithm string

const (
	LL1   Algorithm = "LL(1)"
	SLR1  Algorithm = "SLR(1)"
	CLR1  Algorithm = "CLR(1)"
	LALR1 Algorithm = "LALR(1)"
)

// String returns the string representation of a ParserType.
func (pt Algorithm) String() string {
	return string(pt)
}

// ParseAlgorithm parses a string containing the name of an Algorithm.
func ParseAlgorithm(s string) (Algorithm, error) {
	switch s {
	case LL1.String():
		return LL1, nil
	case SLR1.String():
		return SLR1, nil
	case CLR1.String():
		return CLR1, nil
	case LALR1.String():
		return LALR1, nil
	default:
		return LL1, fmt.Errorf("not a valid ParserType: %q", s)
	}
}

// IsLL1 returns whether the grammar is LL(1).
func IsLL1(g grammar.CFG) bool {
	nts := g.NonTerminals()
	for _, A := range nts {
		AiRule := g.Rule(A)

		// we'll need this later, glubglub 38)
		followSetA := box.StringSetOf(findFOLLOWSet(g, A).Elements())

		// Whenever A -> α | β are two distinct productions of G:
		// -purple dragon book
		for i := range AiRule.Productions {
			for j := i + 1; j < len(AiRule.Productions); j++ {
				alphaFIRST := findFIRSTSet(g, AiRule.Productions[i][0])
				betaFIRST := findFIRSTSet(g, AiRule.Productions[j][0])

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

// findFIRSTSet returns the findFIRSTSet set of symbol X in the grammar.
func findFIRSTSet(g grammar.CFG, X string) box.Set[string] {
	return firstSetSafeRecurse(g, X, box.NewStringSet())
}

func firstSetSafeRecurse(g grammar.CFG, X string, seen box.StringSet) box.Set[string] {
	seen.Add(X)
	if strings.ToLower(X) == X {
		// terminal or epsilon
		return box.NewStringSet(map[string]bool{X: true})
	} else {
		firsts := box.NewStringSet()
		r := g.Rule(X)

		for ntIdx := range r.Productions {
			Y := r.Productions[ntIdx]
			var gotToEnd bool
			for k := 0; k < len(Y); k++ {
				if !seen.Has(Y[k]) {
					firstY := findFIRSTSet(g, Y[k])
					for _, str := range firstY.Elements() {
						if str != "" {
							firsts.Add(str)
						}
					}
					if firstY.Len() == 1 && firstY.Has(grammar.Epsilon[0]) {
						firsts.Add(grammar.Epsilon[0])
					}
					if !firstY.Has(grammar.Epsilon[0]) {
						// if its not, then break
						break
					}
					if k+1 >= len(Y) {
						gotToEnd = true
					}
				}
			}
			if gotToEnd {
				firsts.Add(grammar.Epsilon[0])
			}
		}
		return firsts
	}
}

// findFIRSTSetString is identical to FIRST but for a string of symbols rather than
// just one.
func findFIRSTSetString(g grammar.CFG, X ...string) box.Set[string] {
	first := box.NewStringSet()
	epsilonPresent := false
	for i := range X {
		fXi := findFIRSTSet(g, X[i])
		epsilonPresent = false
		for _, j := range fXi.Elements() {
			if j != grammar.Epsilon[0] {
				first.Add(j)
			} else {
				epsilonPresent = true
			}
		}
		if !epsilonPresent {
			break
		}
	}
	if epsilonPresent {
		first.Add(grammar.Epsilon[0])
	}

	return first
}

// findFOLLOWSet is the used to get the findFOLLOWSet set of symbol X for generating
// various types of parsers.
func findFOLLOWSet(g grammar.CFG, X string) box.Set[string] {
	return recursiveFOLLOWSet(g, X, box.NewStringSet())
}

func recursiveFOLLOWSet(g grammar.CFG, X string, prevFollowChecks box.Set[string]) box.Set[string] {
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
				// how many occurrences of X are there? that says how many times
				// we need to do this, so find them
				var Xcount int
				for k := range prod {
					if prod[k] == X {
						Xcount++
					}
				}

				// do this for each occurence of X
				for Xoccurrence := 0; Xoccurrence < Xcount; Xoccurrence++ {
					//alpha := []string{}
					beta := []string{}
					var doneWithAlpha bool
					var Xencounter int
					for k := range prod {
						if prod[k] == X {
							Xencounter++
							if Xencounter > Xoccurrence && !doneWithAlpha {
								// only count this as end of alpha if we are at the
								// occurrence of X we are looking for
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
						betaFirst := findFIRSTSet(g, beta[b])

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
						betaFirst := findFIRSTSet(g, beta[b])
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
func lr1CLOSURE(g grammar.CFG, I box.SVSet[grammar.LR1Item]) box.SVSet[grammar.LR1Item] {
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
					for _, b := range findFIRSTSetString(g, fullArgs...).Elements() {
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

// outputSetValuedDFA writes a pretty-print representation of a DFA whose values
// in its states are box.SVSets of some type that implements fmt.Stringer. The
// representation is written to w.
func outputSetValuedDFA[E fmt.Stringer](w io.Writer, dfa automaton.DFA[box.SVSet[E]]) {
	// lol let's get some buffering here
	bw := bufio.NewWriter(w)

	bw.WriteString("DFA:\n")
	bw.WriteString("\tStart: ")
	bw.WriteRune('"')
	bw.WriteString(dfa.Start)
	bw.WriteString("\"\n")

	// now get ordered states
	orderedStates := dfa.States()
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
		layout := rosed.Editor{Options: tabOpts}

		// get name and accepting data
		nameCell := fmt.Sprintf("%q", orderedStates[i])
		if dfa.IsAccepting(orderedStates[i]) {
			nameCell = "(" + nameCell + ")"
		}
		nameData := [][]string{{nameCell}}

		// get item data for the state, in deterministic ordering
		itemData := [][]string{}
		items := dfa.GetValue(orderedStates[i])

		lrItemNames := items.Elements()
		sort.Strings(lrItemNames)

		for i := range lrItemNames {
			it := items.Get(lrItemNames[i])
			cell := fmt.Sprintf("[%s]", it.String())
			itemData = append(itemData, []string{cell})
		}

		// okay, finally, get transitions, in deterministic ordering
		transData := [][]string{}

		transOrdered := dfa.GetTransitions(orderedStates[i])
		transOrdered = slices.SortBy(transOrdered, func(left, right [2]string) bool {
			return left[0] < right[0]
		})

		for _, t := range transOrdered {
			cell := fmt.Sprintf("%q ==> %q", t[0], t[1])
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
