// Package grammar implements context-free grammars and associated constructs.
package grammar

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/dekarrin/ictiobus/lex"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/internal/slices"
	"github.com/dekarrin/ictiobus/internal/textfmt"
)

var (
	// expressionGrammar_4_1 is the Grammar corresponding to expression
	// grammar (4.1) from the dragon book.
	//
	// It has start symbol E, non-terminals {E, T, F}, terminals {+, *, (, ),
	// id}, and the following rules:
	//
	// E -> E + T | T
	// T -> T * F | F
	// F -> ( E ) | id
	//
	ExpressionGrammar_4_1 = MustParse(`
		E -> E + T | T;
		T -> T * F | F;
		F -> ( E ) | id;
	`)
)

// Grammar for tunascript language, used by a parsing algorithm to create a
// parse tree from some input.
type Grammar struct {
	rulesByName map[string]int

	// main rules store, not just doing a simple map bc
	// rules may have order that matters
	rules     []Rule
	terminals map[string]lex.TokenClass

	// Start is the name of the start symbol. If not set, It is assumed to be S.
	Start string
}

type marshaledTokenClass struct {
	id    string
	human string
}

// MarshalBinary converts m into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (m marshaledTokenClass) MarshalBinary() ([]byte, error) {
	data := rezi.EncString(m.id)
	data = append(data, rezi.EncString(m.human)...)
	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into m. All
// of m's fields will be replaced by the fields decoded from data.
func (m *marshaledTokenClass) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	m.id, n, err = rezi.DecString(data)
	if err != nil {
		return err
	}
	data = data[n:]

	m.human, _, err = rezi.DecString(data)
	if err != nil {
		return err
	}

	return nil
}

// MarshalBinary converts g into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (g Grammar) MarshalBinary() ([]byte, error) {
	data := rezi.EncMapStringToInt(g.rulesByName)
	rulesData := rezi.EncSliceBinary(g.rules)
	data = append(data, rulesData...)

	serializedTerminals := map[string]marshaledTokenClass{}
	for k := range g.terminals {
		serializedTerminals[k] = marshaledTokenClass{
			id:    g.terminals[k].ID(),
			human: g.terminals[k].Human(),
		}
	}

	data = append(data, rezi.EncMapStringToBinary(serializedTerminals)...)
	data = append(data, rezi.EncString(g.Start)...)
	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into g. All
// of g's fields will be replaced by the fields decoded from data.
func (g *Grammar) UnmarshalBinary(data []byte) error {
	var n int
	var err error

	g.rulesByName, n, err = rezi.DecMapStringToInt(data)
	if err != nil {
		return fmt.Errorf("rulesByName: %w", err)
	}
	data = data[n:]

	rulesSl, n, err := rezi.DecSliceBinary[*Rule](data)
	if err != nil {
		return fmt.Errorf("rules: %w", err)
	}
	if rulesSl == nil {
		g.rules = nil
	} else {
		g.rules = make([]Rule, len(rulesSl))
		for i := range rulesSl {
			if rulesSl[i] != nil {
				g.rules[i] = *rulesSl[i]
			}
		}
	}
	data = data[n:]

	var serializedTerminals map[string]*marshaledTokenClass
	serializedTerminals, n, err = rezi.DecMapStringToBinary[*marshaledTokenClass](data)
	if err != nil {
		return fmt.Errorf("terminals: %w", err)
	}
	data = data[n:]

	if serializedTerminals != nil {
		g.terminals = map[string]lex.TokenClass{}
		for k := range serializedTerminals {
			g.terminals[k] = lex.NewTokenClass(serializedTerminals[k].id, serializedTerminals[k].human)
		}
	}

	g.Start, _, err = rezi.DecString(data)
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	return nil
}

// Terminals returns an ordered list of the terminals in the grammar.
func (g Grammar) Terminals() []string {
	return textfmt.OrderedKeys(g.terminals)
}

// Augmented returns a new grammar that is a copy of this one but with the start
// symbol S changed to a new rule, S' -> S.
func (g Grammar) Augmented() Grammar {
	// get a copy, this will modify g
	g = g.Copy()

	oldStart := g.StartSymbol()
	dummySym := g.GenerateUniqueName(oldStart)

	g.AddRule(dummySym, []string{oldStart})
	g.Start = dummySym

	return g
}

// IsTerminal returns whether the given symbol is a terminal.
func (g Grammar) IsTerminal(sym string) bool {
	_, ok := g.terminals[sym]
	return ok
}

// IsNonTerminal returns whether the given symbol is a non-terminal.
func (g Grammar) IsNonTerminal(sym string) bool {
	_, ok := g.rulesByName[sym]
	return ok
}

// LR0Items returns all LR0 Items in the grammar.
func (g Grammar) LR0Items() []LR0Item {
	nonTerms := g.NonTerminals()

	items := []LR0Item{}
	for _, nt := range nonTerms {
		r := g.Rule(nt)
		items = append(items, r.LRItems()...)
	}
	return items
}

// Copy makes a duplicate deep copy of the grammar.
func (g Grammar) Copy() Grammar {
	g2 := Grammar{
		rulesByName: make(map[string]int, len(g.rulesByName)),
		rules:       make([]Rule, len(g.rules)),
		terminals:   make(map[string]lex.TokenClass, len(g.terminals)),
		Start:       g.Start,
	}

	for k := range g.rulesByName {
		g2.rulesByName[k] = g.rulesByName[k]
	}

	for i := range g.rules {
		g2.rules[i] = g.rules[i].Copy()
	}

	for k := range g.terminals {
		g2.terminals[k] = g.terminals[k]
	}

	return g2
}

// StartSymbol returns the defined start symbol for the grammar. If one is set
// in g.Start, that is returned, otherwise "S" is.
func (g Grammar) StartSymbol() string {
	if g.Start == "" {
		return "S"
	} else {
		return g.Start
	}
}

// String returns a string representation of the grammar.
func (g Grammar) String() string {
	return fmt.Sprintf("(%q, R=%q)", textfmt.OrderedKeys(g.terminals), g.rules)
}

// Rule returns the grammar rule for the given nonterminal symbol.
// If there is no rule defined for that nonterminal, a Rule with an empty
// NonTerminal field is returned; else it will be the same string as the one
// passed in to the function.
func (g Grammar) Rule(nonterminal string) Rule {
	if g.rulesByName == nil {
		return Rule{}
	}

	if curIdx, ok := g.rulesByName[nonterminal]; !ok {
		return Rule{}
	} else {
		return g.rules[curIdx]
	}
}

// Term returns the tokenClass that the given terminal symbol maps to. If the
// given terminal symbol is not defined as a terminal symbol in this grammar,
// the special TokenClass UndefinedToken is returned.
func (g Grammar) Term(terminal string) lex.TokenClass {
	if g.terminals == nil {
		return lex.TokenUndefined
	}

	if class, ok := g.terminals[terminal]; !ok {
		return lex.TokenUndefined
	} else {
		return class
	}
}

// AddTerm adds the given terminal along with the tokenClass that corresponds to
// it; tokens must be of that class in order to match the terminal.
//
// The mapping of terminal symbol IDs to tokenClasses must be 1-to-1; i.e. It is
// an error to map multiple terms to the same tokenClass, and it is an error to
// map the same term to multiple tokenClasses.
//
// As a result, redefining the same term will cause the old one to be removed,
// and during validation if multiple terminals are matched to the same
// tokenClass it will be considered an error.
//
// It is an error to map any terminal to types.TokenUndefined or
// types.TokenEndOfText and attempting to do so will panic immediately.
func (g *Grammar) AddTerm(terminal string, class lex.TokenClass) {
	if terminal == "" {
		panic("empty terminal not allowed")
	}

	if class.ID() == lex.TokenEndOfText.ID() {
		panic("can't add out-of-band signal TokenEndOfText as defined terminal")
	}

	// ensure that it isnt an illegal char, only things used should be 'a-z',
	// '_', and '-'
	for _, ch := range terminal {
		if unicode.IsSpace(ch) || ch == '.' || ch == '|' {
			panic(fmt.Sprintf("invalid terminal name %q; must only be lower-case chars or symbols with no whitespace or periods or bars", terminal))
		}
	}
	if terminal == "$" {
		// we cant use this as the terminal name, ever.
		panic("invalid terminal name '$'; cant use the name of the end-of-text token")
	}

	if class.ID() == lex.TokenUndefined.ID() {
		panic("cannot explicitly map a terminal to TokenUndefined")
	}

	if g.terminals == nil {
		g.terminals = map[string]lex.TokenClass{}
	}

	g.terminals[terminal] = class
}

// RemoveUnusedTerminals removes all terminals that are not currently used by
// any rule.
func (g *Grammar) RemoveUnusedTerminals() {
	producedTerms := box.NewStringSet()
	terms := g.Terminals()

	for i := range g.rules {
		rule := g.rules[i]
		for _, alt := range rule.Productions {
			for _, sym := range alt {
				// if its empty its the empty non-terminal (episilon production) so skip
				if sym == "" {
					continue
				}
				if strings.ToUpper(sym) != sym {
					producedTerms.Add(sym)
				}
			}
		}
	}

	// drop every term that isn't in use
	for _, term := range terms {
		if _, ok := producedTerms[term]; !ok {
			g.RemoveTerm(term)
		}
	}

}

// RemoveTerm eliminates the given terminal from the grammar. The terminal
// will no longer be considered a valid symbol for a rule in the Grammar to
// produce.
//
// If the grammar already does not contain the given nonterminal this function
// has no effect.
func (g *Grammar) RemoveTerm(t string) {
	// is this rule even present?
	delete(g.terminals, t)
}

// RemoveRule eliminates all productions of the given nonterminal from the
// grammar. The nonterminal will no longer be considered to be a part of the
// Grammar.
//
// If the grammar already does not contain the given non-terminal this function
// has no effect.
func (g *Grammar) RemoveRule(nonterminal string) {
	// is this rule even present?

	ruleIdx, ok := g.rulesByName[nonterminal]
	if !ok {
		// that was easy
		return
	}

	// delete name -> index mapping
	delete(g.rulesByName, nonterminal)

	// delete from main store
	if ruleIdx+1 < len(g.rules) {
		g.rules = append(g.rules[:ruleIdx], g.rules[ruleIdx+1:]...)

		// Hold on, we just need to adjust the indexes across this quick...
		for i := ruleIdx; i < len(g.rules); i++ {
			r := g.rules[i]
			g.rulesByName[r.NonTerminal] = i
		}
	} else {
		g.rules = g.rules[:ruleIdx]
	}
}

// AddRule adds the given production for a nonterminal. If the nonterminal has
// already been given, the production is added as an alternative for that
// nonterminal with lower priority than all others already added.
//
// All rules require at least one symbol in the production. For episilon
// production, give only the empty string.
func (g *Grammar) AddRule(nonterminal string, production []string) {
	if nonterminal == "" {
		panic("empty nonterminal name not allowed for production rule")
	}

	// ensure that it isnt an illegal char, only things used should be 'A-Z',
	// '_', and '-'
	for _, ch := range nonterminal {
		if ('A' > ch || ch > 'Z') && ch != '_' && ch != '-' {
			panic(fmt.Sprintf("invalid nonterminal name %q; must only be chars A-Z, \"_\", or \"-\"", nonterminal))
		}
	}

	if len(production) < 1 {
		panic("for epsilon production give empty string; all rules must have productions")
	}

	// check that epsilon, if given, is by itself
	if len(production) != 1 {
		for _, sym := range production {
			if sym == "" {
				panic("episilon production only allowed as sole production of an alternative")
			}
		}
	}

	if g.rulesByName == nil {
		g.rulesByName = map[string]int{}
	}

	curIdx, ok := g.rulesByName[nonterminal]
	if !ok {
		g.rules = append(g.rules, Rule{NonTerminal: nonterminal})
		curIdx = len(g.rules) - 1
		g.rulesByName[nonterminal] = curIdx
	}

	curRule := g.rules[curIdx]
	curRule.Productions = append(curRule.Productions, production)
	g.rules[curIdx] = curRule
}

// NonTerminals returns list of all the non-terminal symbols. All will be upper
// case.
func (g Grammar) NonTerminals() []string {
	return textfmt.OrderedKeys(g.rulesByName)
}

// NonTerminalsByPriority returns list of all the non-terminal symbols in the order
// they were defined in. All will be upper case.
func (g Grammar) NonTerminalsByPriority() []string {
	termNames := []string{}
	for _, r := range g.rules {
		termNames = append(termNames, r.NonTerminal)
	}

	return termNames
}

// NonTerminalsByReversePriority returns list of all the non-terminal symbols in
// reverse order from the order they were defined in. This is handy because it
// can have the effect of causing iteration to do so in a manner that a human
// might do looking at a grammar, reversed.
func (g Grammar) NonTerminalsByReversePriority() []string {
	termNames := []string{}
	for _, r := range g.rules {
		termNames = append([]string{r.NonTerminal}, termNames...)
	}

	return termNames
}

// UnitProductions returns all production rules that are of the form A -> B,
// where A and B are both non-terminals. The returned list contains rules
// mapping the non-terminal to the other non-terminal; all other productions
// from the grammar will not be present.
func (g Grammar) UnitProductions() []Rule {
	allUnitProductions := []Rule{}

	for _, nonTerm := range g.NonTerminals() {
		rule := g.Rule(nonTerm)
		ruleUnitProds := rule.UnitProductions()
		if len(ruleUnitProds) > 0 {
			allUnitProductions = append(allUnitProductions, Rule{NonTerminal: nonTerm, Productions: ruleUnitProds})
		}
	}

	return allUnitProductions
}

// HasUnreachables returns whether the grammar currently has unreachle
// non-terminals.
func (g Grammar) HasUnreachableNonTerminals() bool {
	for _, nonTerm := range g.NonTerminals() {
		if nonTerm == g.StartSymbol() {
			continue
		}

		reachable := false
		for _, otherNonTerm := range g.NonTerminals() {
			if otherNonTerm == nonTerm {
				continue
			}

			r := g.Rule(otherNonTerm)
			if r.CanProduceSymbol(nonTerm) {
				reachable = true
				break
			}
		}

		if !reachable {
			return true
		}

	}

	return false
}

// UnreachableNonTerminals returns all non-terminals (excluding the start
// symbol) that are currently unreachable due to not being produced by any other
// grammar rule.
func (g Grammar) UnreachableNonTerminals() []string {
	unreachables := []string{}

	for _, nonTerm := range g.NonTerminals() {
		if nonTerm == g.StartSymbol() {
			continue
		}

		reachable := false
		for _, otherNonTerm := range g.NonTerminals() {
			if otherNonTerm == nonTerm {
				continue
			}

			r := g.Rule(otherNonTerm)
			if r.CanProduceSymbol(nonTerm) {
				reachable = true
				break
			}
		}

		if !reachable {
			unreachables = append(unreachables, nonTerm)
		}
	}

	return unreachables
}

// RemoveUnitProductions returns a Grammar that derives strings equivalent to
// this one but with all unit production rules removed.
func (g Grammar) RemoveUnitProductions() Grammar {
	for _, nt := range g.NonTerminals() {
		rule := g.Rule(nt)
		resolvedSymbols := map[string]bool{}
		for len(rule.UnitProductions()) > 0 {
			newProds := []Production{}
			for _, p := range rule.Productions {
				if p.IsUnit() && p[0] != nt {
					hoistedRule := g.Rule(p[0])
					includedHoistedProds := []Production{}
					for _, hoistedProd := range hoistedRule.Productions {
						if len(hoistedProd) == 1 && hoistedProd[0] == nt {
							// dont add
						} else if rule.CanProduce(hoistedProd) {
							// dont add
						} else if _, ok := resolvedSymbols[p[0]]; ok {
							// dont add
						} else {
							includedHoistedProds = append(includedHoistedProds, hoistedProd)
						}
					}

					newProds = append(newProds, includedHoistedProds...)
					resolvedSymbols[p[0]] = true
				} else {
					newProds = append(newProds, p)
				}
			}
			rule.Productions = newProds
		}

		g.rules[g.rulesByName[rule.NonTerminal]] = rule
	}

	// okay, now just remove the unreachable ones (not strictly necessary for
	// all interpretations of unit production removal but lets do it anyways for
	// simplicity)
	g = g.RemoveUreachableNonTerminals()

	return g
}

// RemoveUnreachableNonTerminals returns a grammar with all unreachable
// non-terminals removed.
func (g Grammar) RemoveUreachableNonTerminals() Grammar {
	for g.HasUnreachableNonTerminals() {
		for _, nt := range g.UnreachableNonTerminals() {
			g.RemoveRule(nt)
		}
	}
	return g
}

// RemoveEpsilons returns a grammar that derives strings equivalent to the first
// one (with the exception of the empty string) but with all epsilon productions
// automatically eliminated.
//
// Call Validate before this or it may go poorly.
func (g Grammar) RemoveEpsilons() Grammar {
	// run this in a loop until all vars have epsilon propagated out

	propagated := map[string]bool{}
	// first find all of the non-terminals that have epsilon productions

	for {
		// find the first non-terminal with an epsilon production
		toPropagate := ""
		for _, A := range g.NonTerminals() {
			ruleIdx := g.rulesByName[A]
			rule := g.rules[ruleIdx]

			if rule.HasProduction(Epsilon) {
				toPropagate = A
				break
			}
		}

		// if we didn't find any non-terminals with epsilon productions then
		// there are none remaining and we are done.
		if toPropagate == "" {
			break
		}

		// let's call the non-terminal whose epsilons are about to be propegated
		// up 'A'
		A := toPropagate

		// for each of those, remove them from all others
		producesA := map[string]bool{}

		ruleA := g.Rule(A)
		// find all non-terms that produce this, not including self
		for _, B := range g.NonTerminals() {
			ruleIdx := g.rulesByName[B]
			rule := g.rules[ruleIdx]

			// does b produce A?
			if rule.CanProduceSymbol(A) {
				producesA[B] = true
			}
		}

		// okay, now for each production that produces A...
		for B := range producesA {
			ruleB := g.Rule(B)

			if len(ruleA.Productions) == 1 {
				// if A is ONLY an epsilon producer, B can safely eliminate every
				// A from its productions.

				// remove all As from B productions. if it was a unit production,
				// replace it with an epsilon production
				for i, bProd := range ruleB.Productions {
					var newProd Production
					if len(bProd) == 1 && bProd[0] == A {
						newProd = Epsilon
					} else {
						for _, sym := range bProd {
							if sym != A {
								newProd = append(newProd, sym)
							}
						}
					}
					ruleB.Productions[i] = newProd
				}
			} else {
				// general algorithm, summarized in video:
				// https://www.youtube.com/watch?v=j9cNTlGkyZM

				// for each production of b
				var newProds []Production
				for _, bProd := range ruleB.Productions {
					if slices.In(A, bProd) {
						// gen all permutations of A being epsi for that
						// production
						// AsA -> AsA, sA, s, As
						// AAsA -> AAsA, AsA, AsA,
						rewrittenEpsilons := getEpsilonRewrites(A, bProd)

						newProds = append(newProds, rewrittenEpsilons...)
					} else {
						// keep it as-is
						newProds = append(newProds, bProd)
					}
				}

				// if B has already propagated epsilons up we can immediately
				// remove any epsilons it just received
				if _, propagatedEpsilons := propagated[B]; propagatedEpsilons {
					newProds = removeEpsilons(newProds)
				}

				ruleB.Productions = newProds
			}

			if A == B {
				// update our A rule if we need to
				ruleA = ruleB
			}

			ruleBIdx := g.rulesByName[B]
			g.rules[ruleBIdx] = ruleB
		}

		// A is now 'covered'; if it would get an epsilon propagated to it
		// it can remove it directly bc it having an epsilon prod has already
		// been propagated up.
		propagated[A] = true
		ruleA.Productions = removeEpsilons(ruleA.Productions)
		g.rules[g.rulesByName[A]] = ruleA
	}

	// did we just make any rules empty? probably should double-check that.

	// A may be unused by this point, may want to fix that
	return g
}

// RemoveLeftRecursion returns a grammar that has no left recursion, suitable
// for operations on by a top-down parsing method.
//
// This will force immediate removal of epsilon-productions and unit-productions
// as well, as this algorithem only works on CFGs without those.
//
// This is an implementation of Algorithm 4.19 from the purple dragon book,
// "Eliminating left recursion".
func (g Grammar) RemoveLeftRecursion() Grammar {
	// precond: grammar must have no epsilon productions or unit productions
	g = g.RemoveEpsilons().RemoveUnitProductions()

	grammarUpdated := true
	for grammarUpdated {
		grammarUpdated = false

		// arrange the nonterminals in some order A₁, A₂, ..., Aₙ.
		A := g.NonTerminalsByReversePriority()
		for i := range A {
			AiRule := g.Rule(A[i])
			for j := 0; j < i; j++ {
				AjRule := g.Rule(A[j])

				// replace each production of the form Aᵢ -> Aⱼγ by the
				// productions Aᵢ -> δ₁γ | δ₂γ | ... | δₖγ, where
				// Aⱼ -> δ₁ | δ₂ | ... | δₖ are all current Aⱼ productions

				newProds := []Production{}
				for k := range AiRule.Productions {
					if AiRule.Productions[k][0] == A[j] { // if rule is Aᵢ -> Aⱼγ (γ may be ε)
						grammarUpdated = true
						gamma := AiRule.Productions[k][1:]
						deltas := AjRule.Productions

						// add replacement rules
						for d := range deltas {
							deltaProd := deltas[d]
							newProds = append(newProds, append(deltaProd, gamma...))
						}
					} else {
						// add it unchanged
						newProds = append(newProds, AiRule.Productions[k])
					}
				}

				// persist the changes
				AiRule.Productions = newProds
				g.rules[g.rulesByName[A[i]]] = AiRule
			}

			// eliminate the immediate left recursion

			// first, group the productions as
			//
			// A -> Aα₁ | Aα₂ | ... | Aαₘ | β₁ | β₂ | βₙ
			//
			// where no βᵢ starts with an A.
			//
			// ^ That was purple dragon book. 8ut transl8ed, *I* say...
			// "put all the immediate left recursive productions first."
			alphas := []Production{}
			betas := []Production{}
			for k := range AiRule.Productions {
				if AiRule.Productions[k][0] == AiRule.NonTerminal {
					alphas = append(alphas, AiRule.Productions[k][1:])
				} else {
					betas = append(betas, AiRule.Productions[k])
				}
			}

			if len(alphas) > 0 {
				grammarUpdated = true

				// then, replace the A-productions by
				//
				// A  -> β₁A' | β₂A' | ... | βₙA'
				// A' -> α₁A' | α₂A' | ... | αₘA' | ε
				//
				// (purple dragon book)

				if len(betas) < 1 {

					// if we have zero betas, we need to have A produce A' only.
					// but if that's the case, then A -> A' becomes a
					// unit production and since we would be creating A' now, we
					// know A is the only non-term that would produce it,
					// therefore there is no point in putting in a new term and
					// we can immediately just shove all the A' rules into A
					newARule := Rule{NonTerminal: AiRule.NonTerminal}

					for _, a := range alphas {
						newARule.Productions = append(newARule.Productions, append(a, AiRule.NonTerminal))
					}
					// also add epsilon
					newARule.Productions = append(newARule.Productions, Epsilon)

					// update A
					AiRule = newARule
					g.rules[g.rulesByName[A[i]]] = AiRule
				} else {
					APrime := g.GenerateUniqueName(AiRule.NonTerminal)
					newARule := Rule{NonTerminal: AiRule.NonTerminal}
					newAprimeRule := Rule{NonTerminal: APrime}

					for _, b := range betas {
						newARule.Productions = append(newARule.Productions, append(b, APrime))
					}
					for _, a := range alphas {
						newAprimeRule.Productions = append(newAprimeRule.Productions, append(a, APrime))
					}
					// also add epsilon to A'
					newAprimeRule.Productions = append(newAprimeRule.Productions, Epsilon)

					// update A
					AiRule = newARule
					g.rules[g.rulesByName[A[i]]] = AiRule

					// insert A' immediately after A (convention)
					// shouldn't be modifying what we are iterating over bc we are
					// iterating over a pre-retrieved list of nonterminals
					AiIndex := g.rulesByName[A[i]]

					g.insertRule(newAprimeRule, AiIndex)
				}
			}
		}
	}

	g = g.RemoveUreachableNonTerminals()

	return g
}

func (g *Grammar) insertRule(r Rule, idx int) {
	// explicitly copy the end of the slice because trying to
	// save a post list and then modifying has lead to aliasing
	// issues in past
	var postList []Rule = make([]Rule, len(g.rules)-(idx+1))
	copy(postList, g.rules[idx+1:])
	g.rules = append(g.rules[:idx+1], r)
	g.rules = append(g.rules, postList...)

	// update indexes
	for i := idx + 1; i < len(g.rules); i++ {
		g.rulesByName[g.rules[i].NonTerminal] = i
	}
}

// LeftFactor returns a new Grammar equivalent to this one but with all unclear
// alternative choices for a top-down parser are left factored to equivalent
// pairs of statements.
//
// This is an implementation of Algorithm 4.21 from the purple dragon book,
// "Left factoring a grammar".
func (g Grammar) LeftFactor() Grammar {
	changes := true
	for changes {
		changes = false
		A := g.NonTerminals()
		for i := range A {
			AiRule := g.Rule(A[i])
			// find the longest common prefix α common to two or more of Aᵢ's
			// alternatives

			alpha := []string{}
			for j := range AiRule.Productions {
				checkingAlt := AiRule.Productions[j]

				for k := j + 1; k < len(AiRule.Productions); k++ {
					againstAlt := AiRule.Productions[k]
					longestPref := slices.LongestCommonPrefix(checkingAlt, againstAlt)

					// in this case we will simply always take longest between two
					// because anyfin else would require far more intense searching.
					// if more than one matches that, well awesome we'll pick that
					// up too!! 38D

					if len(longestPref) > len(alpha) {
						alpha = longestPref
					}
				}
			}

			if len(alpha) > 0 && !Epsilon.Equal(alpha) {
				// there is a non-trivial common prefix
				changes = true

				// Replace all of the A-productions A -> αβ₁ | αβ₂ | ... | αβₙ | γ,
				// where γ represents all alternatives that do not begin with α,
				// by:
				//
				// A  -> αA' | γ
				// A' -> β₁ | β₂ | ... | βₙ
				//
				// Where A' is a new-non-terminal.
				gamma := []Production{}
				betas := []Production{}

				for _, alt := range AiRule.Productions {
					if slices.HasPrefix(alt, alpha) {
						beta := alt[len(alpha):]
						if len(beta) == 0 {
							beta = Epsilon
						}
						betas = append(betas, beta)
					} else {
						gamma = append(gamma, alt)
					}
				}

				APrime := g.GenerateUniqueName(AiRule.NonTerminal)
				APrimeRule := Rule{NonTerminal: APrime, Productions: betas}

				AiRule.Productions = append([]Production{append(Production(alpha), APrime)}, gamma...)
				// update A
				g.rules[g.rulesByName[A[i]]] = AiRule

				// insert A' immediately after A (convention)
				// shouldn't be modifying what we are iterating over bc we are
				// iterating over a pre-retrieved list of nonterminals
				AiIndex := g.rulesByName[A[i]]
				g.insertRule(APrimeRule, AiIndex)
			}
		}
	}

	return g
}

// MustParse is identical to [Parse] but panics if an error is encountered.
func MustParse(gr string) Grammar {
	g, err := Parse(gr)
	if err != nil {
		panic(err.Error())
	}
	return g
}

// Parse parses a 'grammar string' into a Grammar object. The string must have
// a semicolon between rules, spaces between each symbol, non-terminals must
// contain at least one upper-case letter. Epsilon "ε" is used for the epsilon
// production. Example:
//
//	S -> A | B ;
//	A -> a | ε ;
//	B -> A b | c ;
func Parse(gr string) (Grammar, error) {
	lines := strings.Split(gr, ";")

	var g Grammar
	onFirst := true
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		rule, err := ParseRule(line)
		if err != nil {
			return Grammar{}, err
		}

		if onFirst {
			// this becomes the start symbol
			g.Start = rule.NonTerminal
			onFirst = false
		}

		for _, p := range rule.Productions {
			for _, sym := range p {
				if strings.ToLower(sym) == sym && sym != "" {
					tc := lex.NewTokenClass(strings.ToLower(sym), sym)
					g.AddTerm(tc.ID(), tc)
				}
			}
			g.AddRule(rule.NonTerminal, p)
		}
	}

	return g, nil
}

// TermFor returns the term used in the grammar to represent the given
// TokenClass. If tc is not a TokenClass in the grammar, "" is returned.
func (g Grammar) TermFor(tc lex.TokenClass) string {
	if tc.ID() == lex.TokenEndOfText.ID() {
		return "$"
	}
	for k := range g.terminals {
		if g.terminals[k].Equal(tc) {
			return k
		}
	}
	return ""
}

// GenerateUniqueName generates a name for a non-terminal gauranteed to be
// unique within the grammar, based on original if one is provided.
func (g Grammar) GenerateUniqueName(original string) string {
	newName := original + "-P"
	existingRule := g.Rule(newName)
	for existingRule.NonTerminal != "" {
		newName += "P"
		existingRule = g.Rule(newName)
	}

	return newName
}

// GenerateUniqueTerminal generates a name for a terminal gauranteed to be
// unique within the grammar, based on the given original if one is provided.
func (g Grammar) GenerateUniqueTerminal(original string) string {
	newName := original
	addedHyphen := false
	existingTerm := g.Term(newName)
	for existingTerm.ID() != lex.TokenUndefined.ID() {
		if !addedHyphen {
			newName += "-"
			addedHyphen = true
		}
		newName += "p"
		existingTerm = g.Term(newName)
	}

	return newName
}

// removeEpsilons removes all epsilon-only productions from a list of
// productions and returns the result.
func removeEpsilons(from []Production) []Production {
	newProds := []Production{}

	for i := range from {
		if !from[i].Equal(Epsilon) {
			newProds = append(newProds, from[i])
		}
	}

	return newProds
}

func getEpsilonRewrites(epsilonableNonterm string, prod Production) []Production {
	// how many times does it occur?
	var numOccurances int
	for i := range prod {
		if prod[i] == epsilonableNonterm {
			numOccurances++
		}
	}

	if numOccurances == 0 {
		return []Production{prod}
	}

	// generate all numbers of that binary bitsize

	perms := int(math.Pow(2, float64(numOccurances)))

	// we're using the bitfield of above perms to denote which A should be "on"
	// and which should be "off" in the resulting string.

	newProds := []Production{}

	epsilonablePositions := make([]string, numOccurances)
	for i := perms - 1; i >= 0; i-- {
		// fill positions from the bitfield making up the cur permutation num
		for j := range epsilonablePositions {
			if ((i >> j) & 1) > 0 {
				epsilonablePositions[j] = epsilonableNonterm
			} else {
				epsilonablePositions[j] = ""
			}
		}

		// build a new production
		newProd := Production{}
		var curEpsilonable int
		for j := range prod {
			if prod[j] == epsilonableNonterm {
				pos := epsilonablePositions[curEpsilonable]
				if pos != "" {
					newProd = append(newProd, pos)
				}
				curEpsilonable++
			} else {
				newProd = append(newProd, prod[j])
			}
		}
		if len(newProd) == 0 {
			newProd = Epsilon
		}
		newProds = append(newProds, newProd)
	}

	// now eliminate every production that is a duplicate
	uniqueNewProds := []Production{}
	seenProductions := map[string]bool{}
	for i := range newProds {
		str := strings.Join(newProds[i], " ")

		if _, alreadySeen := seenProductions[str]; alreadySeen {
			continue
		}

		uniqueNewProds = append(uniqueNewProds, newProds[i])
		seenProductions[str] = true
	}

	return uniqueNewProds
}

// Validates that the current rules form a complete grammar with no
// missing definitions. TODO: should also dupe-check rules.
func (g Grammar) Validate() error {
	if g.rulesByName == nil {
		g.rulesByName = map[string]int{}
	}

	// a grammar needs at least one rule and at least one terminal or it makes
	// no sense.
	if len(g.rules) < 1 {
		return fmt.Errorf("no rules defined in grammar")
	} else if len(g.terminals) < 1 {
		return fmt.Errorf("no terminals defined in grammar")
	}

	producedNonTerms := map[string]bool{}
	producedTerms := map[string]bool{}

	// make sure all non-terminals produce either defined
	// non-terminals or defined terminals
	orderedTermKeys := textfmt.OrderedKeys(g.terminals)

	errStr := ""

	for i := range g.rules {
		rule := g.rules[i]
		for _, alt := range rule.Productions {
			for _, sym := range alt {
				// if its empty its the empty non-terminal (episilon production) so skip
				if sym == "" {
					continue
				}
				if g.IsNonTerminal(sym) {
					// non-terminal
					if _, ok := g.rulesByName[sym]; !ok {
						errStr += fmt.Sprintf("ERR: no production defined for nonterminal %q produced by %q\n", sym, rule.NonTerminal)
					}
					producedNonTerms[sym] = true
				} else {
					// terminal
					if _, ok := g.terminals[sym]; !ok {
						errStr += fmt.Sprintf("ERR: undefined terminal %q produced by %q\n", sym, rule.NonTerminal)
					}
					producedTerms[sym] = true
				}
			}
		}
	}

	// make sure every defined terminal is used and that each maps to a distinct
	// token class
	seenClasses := map[string]string{}
	for _, term := range orderedTermKeys {
		if _, ok := producedTerms[term]; !ok {
			errStr += fmt.Sprintf("ERR: terminal %q is not produced by any rule\n", term)
		}

		cl := g.terminals[term]
		if mappedBy, alreadySeen := seenClasses[cl.ID()]; alreadySeen {
			errStr += fmt.Sprintf("ERR: terminal %q maps to same class %q as terminal %q", term, cl.Human(), mappedBy)
		}
		seenClasses[cl.ID()] = term
	}

	// make sure every non-term is used
	for _, r := range g.rules {
		// S is used by default, don't check that one
		if r.NonTerminal == g.StartSymbol() {
			continue
		}

		if _, ok := producedNonTerms[r.NonTerminal]; !ok {
			errStr += fmt.Sprintf("ERR: non-terminal %q not produced by any rule\n", r.NonTerminal)
		}
	}

	// make sure we HAVE an S
	if _, ok := g.rulesByName[g.StartSymbol()]; !ok {
		errStr += fmt.Sprintf("ERR: no rules defined for productions of start symbol '%s'", g.StartSymbol())
	}

	if len(errStr) > 0 {
		// chop off trailing newline
		errStr = errStr[:len(errStr)-1]
		return fmt.Errorf(errStr)
	}

	return nil
}
