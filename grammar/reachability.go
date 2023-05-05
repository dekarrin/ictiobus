package grammar

import (
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/slices"
)

func (g Grammar) ReachableFrom(start string, end string) (bool, []box.Pair[string, Production]) {
	if !g.IsNonTerminal(start) {
		return false, nil
	}
	if !g.IsNonTerminal(end) && !g.IsTerminal(end) {
		return false, nil
	}

	// run reachability algorithm, but instead of starting at the start symbol,
	// start with each production of it.

	reached := box.NewSVSet[slices.LList[box.Pair[string, Production]]]()

	r := g.Rule(start)
	for _, p := range r.Productions {
		for _, sym := range p {
			var path slices.LList[box.Pair[string, Production]]
			path = path.Add(box.PairOf(start, p))

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
					path = path.Add(box.PairOf(k, prod))

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
