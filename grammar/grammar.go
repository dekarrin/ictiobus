// Package grammar implements context-free grammars and associated constructs.
package grammar

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/types"

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
	terminals map[string]types.TokenClass

	// Start is the name of the start symbol. If not set, It is assumed to be S.
	Start string
}

type marshaledTokenClass struct {
	id    string
	human string
}

func (m marshaledTokenClass) MarshalBinary() ([]byte, error) {
	data := rezi.EncString(m.id)
	data = append(data, rezi.EncString(m.human)...)
	return data, nil
}

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
		g.terminals = map[string]types.TokenClass{}
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

// LR1_CLOSURE is the closure function used for constructing LR(1) item sets for
// use in a parser DFA.
//
// Note: this actually takes the grammar for each production B -> gamma in G,
// not G'. It's assumed this function is only called on a g.Augmented()
// instance.
func (g Grammar) LR1_CLOSURE(I box.SVSet[LR1Item]) box.SVSet[LR1Item] {
	Iset := I.Copy()
	I = Iset.(box.SVSet[LR1Item])

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

						var newItem LR1Item

						// SPECIAL CASE: if we're dealing with an epsilon, our
						// item will look like "A -> .". normally we are adding
						// a dot at the START of an item added in the LR1
						// CLOSURE func, but since "A -> ." should always be
						// treated as "at the end", we add a special item with
						// only the dot, and no left or right.
						if gamma.Equal(Epsilon) {
							newItem = LR1Item{LR0Item: LR0Item{NonTerminal: B}, Lookahead: b}
						} else {
							newItem = LR1Item{
								LR0Item: LR0Item{
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

// Copy makes a duplicate deep copy of the grammar.
func (g Grammar) Copy() Grammar {
	g2 := Grammar{
		rulesByName: make(map[string]int, len(g.rulesByName)),
		rules:       make([]Rule, len(g.rules)),
		terminals:   make(map[string]types.TokenClass, len(g.terminals)),
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
// the special TokenClass types.UndefinedToken is returned.
func (g Grammar) Term(terminal string) types.TokenClass {
	if g.terminals == nil {
		return types.TokenUndefined
	}

	if class, ok := g.terminals[terminal]; !ok {
		return types.TokenUndefined
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
func (g *Grammar) AddTerm(terminal string, class types.TokenClass) {
	if terminal == "" {
		panic("empty terminal not allowed")
	}

	if class.ID() == types.TokenEndOfText.ID() {
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

	if class.ID() == types.TokenUndefined.ID() {
		panic("cannot explicitly map a terminal to TokenUndefined")
	}

	if g.terminals == nil {
		g.terminals = map[string]types.TokenClass{}
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

// PriorityNonTerminals returns list of all the non-terminal symbols in the order
// they were defined in. All will be upper case.
func (g Grammar) PriorityNonTerminals() []string {
	termNames := []string{}
	for _, r := range g.rules {
		termNames = append(termNames, r.NonTerminal)
	}

	return termNames
}

// ReversePriorityNonTerminals returns list of all the non-terminal symbols in
// reverse order from the order they were defined in. This is handy because it
// can have the effect of causing iteration to do so in a manner that a human
// might do looking at a grammar, reversed.
func (g Grammar) ReversePriorityNonTerminals() []string {
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

// deriveShortest returns the parse tree for the given symbol with the fewest
// possible number of nodes.
func (g Grammar) deriveShortestTree(sym string, shortestDerivation map[string]Production, tokMaker func(term string) types.Token) *types.ParseTree {
	root := &types.ParseTree{
		Value: sym,
	}

	treeStack := box.NewStack([]*types.ParseTree{root})

	for treeStack.Len() > 0 {
		pt := treeStack.Pop()

		if pt.Terminal {
			continue
		}

		option := shortestDerivation[pt.Value]
		if option.Equal(Epsilon) {
			pt.Children = []*types.ParseTree{{Terminal: true}}
		} else {
			pt.Children = make([]*types.ParseTree, len(option))
			for i, sym := range option {
				if g.IsTerminal(sym) {
					pt.Children[i] = &types.ParseTree{
						Value:    sym,
						Terminal: true,
						Source:   tokMaker(sym),
					}
				} else {
					pt.Children[i] = &types.ParseTree{
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

// DeriveFullTree derives a parse tree based on the grammar that is guaranteed
// to contain every rule at least once. It is *not* garunteed to have each rule
// only once, although it will make a best effort to minimize duplications.
//
// The fakeValProducer map, if provided, will be used to assign a lexed
// value to each synthesized terminal node that has a class whose ID matches
// a key in it. If the map is not provided or does not contain a key for a
// token class matching a terminal being generated, a default string will be
// automatically generated for it.
func (g Grammar) DeriveFullTree(fakeValProducer ...map[string]func() string) ([]types.ParseTree, error) {
	// create a function to get a value for any terminal from the
	// fakeValProducer, falling back on default behavior if none is provided or
	// if a token class is not found in the fakeValProducer.
	var lineNo int
	makeTermSource := func(forTerm string) types.Token {
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
	toCreate := map[string][]Production{}
	uncoveredSymbols := box.NewStringSet()
	for _, nonTerm := range g.NonTerminals() {
		rule := g.Rule(nonTerm)
		prods := make([]Production, len(rule.Productions))
		copy(prods, rule.Productions)
		toCreate[nonTerm] = prods
		uncoveredSymbols.Add(nonTerm)
	}

	// we also need to get the 'shortest' production for each non-terminal; this
	// is the production that eventually results in the fewest number of
	// non-terminals being used in the final parse tree.
	shortestDerivation, err := g.CreateFewestNonTermsAlternationsTable()
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

	finalTrees := []types.ParseTree{}

	for len(toCreate) > 0 {
		root := &types.ParseTree{
			Terminal: false,
			Value:    g.StartSymbol(),
		}
		treeStack := box.NewStack([]*types.ParseTree{root})

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
				shortest := g.deriveShortestTree(pt.Value, shortestDerivation, makeTermSource)
				pt.Children = shortest.Children
				continue
			}

			option := options[0]
			if option.Equal(Epsilon) {
				// epsilon prod must be specifically added to done here since it
				// would not be added to the done list otherwise
				pt.Children = []*types.ParseTree{{Terminal: true}}
			} else {
				pt.Children = make([]*types.ParseTree, len(option))
				for i, sym := range option {
					if g.IsTerminal(sym) {
						pt.Children[i] = &types.ParseTree{
							Value:    sym,
							Terminal: true,
							Source:   makeTermSource(sym),
						}
					} else {
						pt.Children[i] = &types.ParseTree{
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
				toCreate = map[string][]Production{}
			} else {
				// repopulate all items that were covered in some prior iteration
				for _, nonTerm := range g.NonTerminals() {
					if !stillUncovered.Has(nonTerm) {
						rule := g.Rule(nonTerm)
						prods := make([]Production, len(rule.Productions))
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
			updatedFinal := make([]types.ParseTree, 0)
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

// CreateFewestNonTermsAlternationsTable returns a map of non-terminals to the
// production in their associated rule that using to derive would result in the
// fewest new non-terminals being derived for that rule.
func (g Grammar) CreateFewestNonTermsAlternationsTable() (map[string]Production, error) {
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
	prodStringToProd := map[string]Production{}

	// Create entries in a table for each non-terminal for which there is a rule
	// in the grammar, initially with no value set. Each item will be the score.
	// Additionally, create a "shortest alt" table from a rule to the production
	// that has the lowest score
	shortest := map[string]Production{}
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
		A := g.ReversePriorityNonTerminals()
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

// recursiveFindFollowSet
func (g Grammar) recursiveFindFollowSet(X string, prevFollowChecks box.Set[string]) box.Set[string] {
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
					alpha := []string{}
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
							alpha = append(alpha, prod[k])
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
							if k != Epsilon[0] {
								followSet.Add(k)
							}
						}

						if !betaFirst.Has(Epsilon[0]) {
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
						if !betaFirst.Has(Epsilon[0]) {
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
							followA := g.recursiveFindFollowSet(A[i], prevFollowChecks)
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
					tc := types.MakeDefaultClass(sym)
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
func (g Grammar) TermFor(tc types.TokenClass) string {
	if tc.ID() == types.TokenEndOfText.ID() {
		return "$"
	}
	for k := range g.terminals {
		if g.terminals[k].Equal(tc) {
			return k
		}
	}
	return ""
}

// IsLL1 returns whether the grammar is LL(1).
func (g Grammar) IsLL1() bool {
	nts := g.NonTerminals()
	for _, A := range nts {
		AiRule := g.Rule(A)

		// we'll need this later, glubglub 38)
		followSetA := box.StringSetOf(g.FOLLOW(A).Elements())

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
				if bFSet.Has(Epsilon[0]) {
					if !followSetA.DisjointWith(aFSet) {
						return false
					}
				}
				if aFSet.Has(Epsilon[0]) {
					if !followSetA.DisjointWith(bFSet) {
						return false
					}
				}
			}

		}
	}

	return true
}

// FOLLOW is the used to get the FOLLOW set of symbol X for generating various
// types of parsers.
func (g Grammar) FOLLOW(X string) box.Set[string] {
	return g.recursiveFindFollowSet(X, box.NewStringSet())
}

// FIRST_STRING is identical to FIRST but for a string of symbols rather than
// just one.
func (g Grammar) FIRST_STRING(X ...string) box.Set[string] {
	first := box.NewStringSet()
	epsilonPresent := false
	for i := range X {
		fXi := g.FIRST(X[i])
		epsilonPresent = false
		for _, j := range fXi.Elements() {
			if j != Epsilon[0] {
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
		first.Add(Epsilon[0])
	}

	return first
}

// FIRST returns the FIRST set of symbol X in the grammar.
func (g Grammar) FIRST(X string) box.Set[string] {
	return g.firstSetSafeRecurse(X, box.NewStringSet())
}

func (g Grammar) firstSetSafeRecurse(X string, seen box.StringSet) box.Set[string] {
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
					firstY := g.FIRST(Y[k])
					for _, str := range firstY.Elements() {
						if str != "" {
							firsts.Add(str)
						}
					}
					if firstY.Len() == 1 && firstY.Has(Epsilon[0]) {
						firsts.Add(Epsilon[0])
					}
					if !firstY.Has(Epsilon[0]) {
						// if its not, then break
						break
					}
					if k+1 >= len(Y) {
						gotToEnd = true
					}
				}
			}
			if gotToEnd {
				firsts.Add(Epsilon[0])
			}
		}
		return firsts
	}
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
	for existingTerm.ID() != types.TokenUndefined.ID() {
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
