package grammar

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/internal/slices"
)

// Production is a production of a grammar rule.
type Production []string

var (
	// Epsilon represents the empty production.
	Epsilon = Production{""}

	// Error is a production used to indicate an error.
	Error = Production{}
)

// Copy returns a deep-copied duplicate of this production.
func (p Production) Copy() Production {
	p2 := make(Production, len(p))
	copy(p2, p)

	return p2
}

// AllLR0Items returns all LR0 items of the production. Note: a Production does not
// know what non-terminal produces it, so the NonTerminal field of the returned
// LR0Items will be blank.
func (p Production) AllLR0Items() []LR0Item {
	if p.Equal(Epsilon) {
		// not a typo: don't return an empty slice of items, return a slice
		// containing a single empty item ("NonTerminal -> .")
		return []LR0Item{{}}
	}

	items := []LR0Item{}
	for dot := 0; dot < len(p); dot++ {
		item := LR0Item{
			Left:  p[:dot],
			Right: p[dot:],
		}
		items = append(items, item)
	}

	// finally, add the single dot for the end
	items = append(items, LR0Item{Left: p})

	return items
}

// Equal returns whether Rule is equal to another value. It will not be equal
// if the other value cannot be cast to Production or *Production.
func (p Production) Equal(o any) bool {
	other, ok := o.(Production)
	if !ok {
		// also okay if its the pointer value, as long as its non-nil
		otherPtr, ok := o.(*Production)
		if !ok {
			// also okay if it's a string slice
			otherSlice, ok := o.([]string)

			if !ok {
				// also okay if it's a ptr to string slice
				otherSlicePtr, ok := o.(*[]string)
				if !ok {
					return false
				} else if otherSlicePtr == nil {
					return false
				} else {
					other = Production(*otherSlicePtr)
				}
			} else {
				other = Production(otherSlice)
			}
		} else if otherPtr == nil {
			return false
		} else {
			other = *otherPtr
		}
	}

	if len(p) != len(other) {
		return false
	} else {
		for i := range p {
			if p[i] != other[i] {
				return false
			}
		}
	}

	return true
}

// String returns the string representation of the Production.
func (p Production) String() string {
	// if it's an epsilon production output that symbol only
	if p.Equal(Epsilon) {
		return "ε"
	}
	// separate each by space and call it good

	var sb strings.Builder

	for i := range p {
		sb.WriteString(p[i])
		if i+1 < len(p) {
			sb.WriteRune(' ')
		}

	}

	return sb.String()
}

// IsUnit returns whether this production is a unit production.
func (p Production) IsUnit() bool {
	return len(p) == 1 && !p.Equal(Epsilon) && strings.ToUpper(p[0]) == p[0]
}

// HasSymbol returns whether the production has the given symbol in it.
func (p Production) HasSymbol(sym string) bool {
	return slices.In(sym, p)
}

// MarshalBinary converts p into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (p Production) MarshalBinary() ([]byte, error) {
	return rezi.EncSliceString([]string(p)), nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into p. All
// of p's fields will be replaced by the fields decoded from data.
func (p *Production) UnmarshalBinary(data []byte) error {
	strSlice, _, err := rezi.DecSliceString(data)
	if err != nil {
		return err
	}

	*p = strSlice
	return nil
}

// Rule is a set of derivation rules of a grammar for some NonTerminal.
type Rule struct {
	NonTerminal string
	Productions []Production
}

// MustParseRule is like parseRule but panics if it can't.
func MustParseRule(r string) Rule {
	rule, err := ParseRule(r)
	if err != nil {
		panic(err.Error())
	}
	return rule
}

// ParseRule parses a Rule from a string like "S -> X | Y".
func ParseRule(r string) (Rule, error) {
	r = strings.TrimSpace(r)
	sides := strings.Split(r, "->")
	if len(sides) != 2 {
		return Rule{}, fmt.Errorf("not a rule of form 'NONTERM -> SYMBOL SYMBOL | SYMBOL ...': %q", r)
	}
	nonTerminal := strings.TrimSpace(sides[0])

	if nonTerminal == "" {
		return Rule{}, fmt.Errorf("empty nonterminal name not allowed for production rule")
	}

	// ensure that it isnt an illegal char, only things used should be 'A-Z',
	// '_', and '-'
	for _, ch := range nonTerminal {
		if ('A' > ch || ch > 'Z') && ch != '_' && ch != '-' {
			return Rule{}, fmt.Errorf("invalid nonterminal name %q; must only be chars A-Z, \"_\", or \"-\"", nonTerminal)
		}
	}

	parsedRule := Rule{NonTerminal: nonTerminal}

	productionsString := strings.TrimSpace(sides[1])
	prodStrings := strings.Split(productionsString, "|")
	for _, p := range prodStrings {
		parsedProd := Production{}
		// split by spaces
		p = strings.TrimSpace(p)
		symbols := strings.Split(p, " ")
		for _, sym := range symbols {
			sym = strings.TrimSpace(sym)

			if sym == "" {
				return Rule{}, fmt.Errorf("empty symbol not allowed")
			}

			if strings.ToLower(sym) == "ε" {
				// epsilon production
				parsedProd = Epsilon
				continue
			} else {
				parsedProd = append(parsedProd, sym)
			}
		}

		parsedRule.Productions = append(parsedRule.Productions, parsedProd)
	}

	return parsedRule, nil
}

// MarshalBinary converts r into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (r Rule) MarshalBinary() ([]byte, error) {
	data := rezi.EncString(r.NonTerminal)
	data = append(data, rezi.EncSliceBinary(r.Productions)...)
	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into r. All
// of r's fields will be replaced by the fields decoded from data.
func (r *Rule) UnmarshalBinary(data []byte) error {
	var n int
	var err error

	r.NonTerminal, n, err = rezi.DecString(data)
	if err != nil {
		return err
	}
	data = data[n:]

	prodSl, _, err := rezi.DecSliceBinary[*Production](data)
	if err != nil {
		return err
	}

	if prodSl == nil {
		r.Productions = nil
	} else {
		r.Productions = make([]Production, len(prodSl))
		for i := range prodSl {
			if prodSl[i] != nil {
				r.Productions[i] = *prodSl[i]
			}
		}
	}

	return nil
}

// Returns all LRItems in the Rule with their NonTerminal field properly set.
func (r Rule) LRItems() []LR0Item {
	items := []LR0Item{}
	for _, p := range r.Productions {
		prodItems := p.AllLR0Items()
		for i := range prodItems {
			item := prodItems[i]
			item.NonTerminal = r.NonTerminal
			prodItems[i] = item
		}
		items = append(items, prodItems...)
	}
	return items
}

// Copy returns a deep-copy duplicate of the given Rule.
func (r Rule) Copy() Rule {
	r2 := Rule{
		NonTerminal: r.NonTerminal,
		Productions: make([]Production, len(r.Productions)),
	}

	for i := range r.Productions {
		r2.Productions[i] = r.Productions[i].Copy()
	}

	return r2
}

// String returns the string representation of a rule.
func (r Rule) String() string {
	var sb strings.Builder

	sb.WriteString(r.NonTerminal)
	sb.WriteString(" -> ")

	for i := range r.Productions {
		sb.WriteString(r.Productions[i].String())
		if i+1 < len(r.Productions) {
			sb.WriteString(" | ")
		}
	}

	return sb.String()
}

// ReplaceProduction returns a rule that does not include the given production
// and subsitutes the given production(s) for it. If no productions are given
// the specified production is simply removed. If the specified production
// does not exist, the replacements are added to the end of the rule.
func (r Rule) ReplaceProduction(p Production, replacements ...Production) Rule {
	var addedReplacements bool
	newProds := []Production{}
	for i := range r.Productions {
		if !r.Productions[i].Equal(p) {
			newProds = append(newProds, r.Productions[i])
		} else if len(replacements) > 0 {
			newProds = append(newProds, replacements...)
			addedReplacements = true
		}
	}
	if !addedReplacements {
		newProds = append(newProds, replacements...)
	}

	r.Productions = newProds
	return r
}

// Equal returns whether Rule is equal to another value. It will not be equal
// if the other value cannot be casted to a Rule or *Rule.
func (r Rule) Equal(o any) bool {
	other, ok := o.(Rule)
	if !ok {
		// also okay if its the pointer value, as long as its non-nil
		otherPtr, ok := o.(*Rule)
		if !ok {
			return false
		} else if otherPtr == nil {
			return false
		}
		other = *otherPtr
	}

	if r.NonTerminal != other.NonTerminal {
		return false
	} else if !slices.EqualSlices(r.Productions, other.Productions) {
		return false
	}

	return true
	// cant do util.EqualSlices here because Productions is a slice of []string
}

// CanProduce returns whether this rule can produce the given Production.
func (r Rule) CanProduce(p Production) bool {
	for _, alt := range r.Productions {
		if alt.Equal(p) {
			return true
		}
	}
	return false
}

// CanProduceSymbol whether any alternative in productions produces the
// given term/non-terminal
func (r Rule) CanProduceSymbol(termOrNonTerm string) bool {
	for _, alt := range r.Productions {
		for _, sym := range alt {
			if sym == termOrNonTerm {
				return true
			}
		}
	}
	return false
}

// HasProduction returns whether the rule has a production of the exact sequence
// of symbols entirely.
func (r Rule) HasProduction(prod Production) bool {
	for _, alt := range r.Productions {
		if len(alt) == len(prod) {
			eq := true
			for i := range alt {
				if alt[i] != prod[i] {
					eq = false
					break
				}
			}
			if eq {
				return true
			}
		}
	}
	return false
}

// UnitProductions returns all productions from the Rule that are unit
// productions; i.e. are of the form A -> B where both A and B are
// non-terminals.
func (r Rule) UnitProductions() []Production {
	prods := []Production{}

	for _, alt := range r.Productions {
		if alt.IsUnit() {
			prods = append(prods, alt)
		}
	}

	return prods
}
