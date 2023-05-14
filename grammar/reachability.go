package grammar

import (
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
)

// ReachableFrom returns whether the symbol end can be derived from symbol start
// with any combination of derivations. Returns whether it can be reached, and
// if it can, additional returns a set of Rules that each contain exactly one
// production; each is a derivation that must be done to a symbol of the prior
// derived to string to ultimately end up at end.
func (g CFG) ReachableFrom(start string, end string) (bool, []Rule) {
	if !g.IsNonTerminal(start) {
		return false, nil
	}
	if !g.IsNonTerminal(end) && !g.IsTerminal(end) {
		return false, nil
	}

	// run reachability algorithm, but instead of starting at the start symbol,
	// start with each production of it.

	reached := box.NewSVSet[slices.LList[Rule]]()

	r := g.Rule(start)
	for _, p := range r.Productions {
		for _, sym := range p {
			var path slices.LList[Rule]
			path = path.Add(Rule{NonTerminal: start, Productions: []Production{p}})

			// if path is any of the not-via's, skip it.

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
					path = path.Add(Rule{NonTerminal: k, Productions: []Production{prod}})

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
