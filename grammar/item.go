package grammar

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
)

// LR0Item is an LR(0) item from a grammar. This is an object used for
// generating handle recognition systems for bottom-up parsers. Each LR0Item
// has a NonTerminal of the rule it relates to, and a left and right side of
// a dot. The dot denotes handle recognition; all symbols on the left have been
// seen, and all symbols on the right have yet to be seen.
type LR0Item struct {
	// NonTerminal is the symbol at the head of the grammar rule that this item
	// is for.
	NonTerminal string

	// Left is all symbols in the production of the rule that are to the left of
	// the dot.
	Left []string

	// Right is all symbols in the production of the rule that are to the right
	// of the dot.
	Right []string
}

// MarshalBinary converts lr0 into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (lr0 LR0Item) MarshalBinary() ([]byte, error) {
	data := rezi.EncString(lr0.NonTerminal)
	data = append(data, rezi.EncSliceString(lr0.Left)...)
	data = append(data, rezi.EncSliceString(lr0.Right)...)

	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into lr0.
// All of lr0's fields will be replaced by the fields decoded from data.
func (lr0 *LR0Item) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	lr0.NonTerminal, n, err = rezi.DecString(data)
	if err != nil {
		return fmt.Errorf(".NonTerminal: %w", err)
	}
	data = data[n:]

	lr0.Left, n, err = rezi.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".Left: %w", err)
	}
	data = data[n:]

	lr0.Right, _, err = rezi.DecSliceString(data)
	if err != nil {
		return fmt.Errorf(".Right: %w", err)
	}

	return nil
}

// Equal returns whether the given LR0Item is equal to another LR0Item or
// *LR0Item.
func (lr0 LR0Item) Equal(o any) bool {
	other, ok := o.(LR0Item)
	if !ok {
		otherPtr, ok := o.(*LR0Item)
		if !ok {
			return false
		}
		if otherPtr == nil {
			return false
		}
		other = *otherPtr
	}

	if lr0.NonTerminal != other.NonTerminal {
		return false
	} else if len(lr0.Left) != len(other.Left) {
		return false
	} else if len(lr0.Right) != len(other.Right) {
		return false
	}

	// now check the left and right
	for i := range lr0.Left {
		if lr0.Left[i] != other.Left[i] {
			return false
		}
	}
	for i := range lr0.Right {
		if lr0.Right[i] != other.Right[i] {
			return false
		}
	}

	return true
}

// LR1Item is an LR(1) item from a grammar. This is the same as an LR0Item, but
// with a Lookahead symbol that determines when the item can be used.
type LR1Item struct {
	LR0Item
	Lookahead string
}

// MarshalBinary converts lr1 into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (lr1 LR1Item) MarshalBinary() ([]byte, error) {
	data := rezi.EncBinary(lr1.LR0Item)
	data = append(data, rezi.EncString(lr1.Lookahead)...)
	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into lr1.
// All of lr1's fields will be replaced by the fields decoded from data.
func (lr1 *LR1Item) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	n, err = rezi.DecBinary(data, &lr1.LR0Item)
	if err != nil {
		return fmt.Errorf(".LR0Item: %w", err)
	}
	data = data[n:]

	lr1.Lookahead, _, err = rezi.DecString(data)
	if err != nil {
		return fmt.Errorf(".Left: %w", err)
	}

	return nil
}

// CoreSet returns the set of cores in a set of LR1 items. A core of an LR1 item
// is simply the LR0 portion of it.
func CoreSet(s box.VSet[string, LR1Item]) box.SVSet[LR0Item] {
	cores := box.NewSVSet[LR0Item]()
	for _, elem := range s.Elements() {
		lr1 := s.Get(elem)
		cores.Set(lr1.LR0Item.String(), lr1.LR0Item)
	}

	return cores
}

// Equal returns whether the given LR0Item is equal to another LR1Item or
// *LR1Item.
func (lr1 LR1Item) Equal(o any) bool {
	other, ok := o.(LR1Item)
	if !ok {
		otherPtr, ok := o.(*LR1Item)
		if !ok {
			return false
		}
		if otherPtr == nil {
			return false
		}
		other = *otherPtr
	}

	if !lr1.LR0Item.Equal(other.LR0Item) {
		return false
	} else if lr1.Lookahead != other.Lookahead {
		return false
	}

	return true
}

// Copy returns a deep copy of the LR1Item.
func (lr1 LR1Item) Copy() LR1Item {
	lrCopy := LR1Item{}
	lrCopy.NonTerminal = lr1.NonTerminal
	lrCopy.Left = make([]string, len(lr1.Left))
	copy(lrCopy.Left, lr1.Left)
	lrCopy.Right = make([]string, len(lr1.Right))
	copy(lrCopy.Right, lr1.Right)
	lrCopy.Lookahead = lr1.Lookahead

	return lrCopy
}

// MustParseLR0Item is identical to ]ParseLR0Item], but panics if any parse
// error occurs.
func MustParseLR0Item(s string) LR0Item {
	i, err := ParseLR0Item(s)
	if err != nil {
		panic(err.Error())
	}
	return i
}

// MustParseLR1Item is identical to [ParseLR1Item], but panics if any parse
// error occurs.
func MustParseLR1Item(s string) LR1Item {
	i, err := ParseLR1Item(s)
	if err != nil {
		panic(err.Error())
	}
	return i
}

// ParseLR1Item parses a string of the form "NONTERM -> ALPHA.BETA" into an
// LR0Item.
func ParseLR0Item(s string) (LR0Item, error) {
	sides := strings.Split(s, "->")
	if len(sides) != 2 {
		return LR0Item{}, fmt.Errorf("not an item of form 'NONTERM -> ALPHA.BETA': %q", s)
	}
	nonTerminal := strings.TrimSpace(sides[0])

	if nonTerminal == "" {
		return LR0Item{}, fmt.Errorf("empty nonterminal name not allowed for item")
	}

	parsedItem := LR0Item{
		NonTerminal: nonTerminal,
	}

	productionsString := strings.TrimSpace(sides[1])
	prodStrings := strings.Split(productionsString, ".")
	if len(prodStrings) != 2 {
		return LR0Item{}, fmt.Errorf("item must have exactly one dot")
	}

	alphaStr := strings.TrimSpace(prodStrings[0])
	betaStr := strings.TrimSpace(prodStrings[1])

	alphaSymbols := strings.Split(alphaStr, " ")
	betaSymbols := strings.Split(betaStr, " ")

	var parsedAlpha, parsedBeta []string

	for _, aSym := range alphaSymbols {
		aSym = strings.TrimSpace(aSym)

		if aSym == "" {
			continue
		}

		if strings.ToLower(aSym) == "ε" {
			// epsilon production
			aSym = ""
		}

		parsedAlpha = append(parsedAlpha, aSym)
	}

	for _, bSym := range betaSymbols {
		bSym = strings.TrimSpace(bSym)

		if bSym == "" {
			continue
		}

		if strings.ToLower(bSym) == "ε" {
			// epsilon production
			bSym = ""
		}

		parsedBeta = append(parsedBeta, bSym)
	}

	parsedItem.Left = parsedAlpha
	parsedItem.Right = parsedBeta

	return parsedItem, nil
}

// ParseLR1Item parses a string of the form "NONTERM -> ALPHA.BETA, LOOKAHEAD"
// into an LR1Item.
func ParseLR1Item(s string) (LR1Item, error) {
	sides := strings.Split(s, ",")
	if len(sides) != 2 {
		return LR1Item{}, fmt.Errorf("not an item of form 'NONTERM -> ALPHA.BETA, a': %q", s)
	}

	item := LR1Item{}
	var err error
	item.LR0Item, err = ParseLR0Item(sides[0])
	if err != nil {
		return item, err
	}

	item.Lookahead = strings.TrimSpace(sides[1])

	return item, nil
}

// String returns the string representation of an LR0 Item.
func (item LR0Item) String() string {
	nonTermPhrase := ""
	if item.NonTerminal != "" {
		nonTermPhrase = fmt.Sprintf("%s -> ", item.NonTerminal)
	}

	left := strings.Join(item.Left, " ")
	right := strings.Join(item.Right, " ")

	if len(left) > 0 {
		left = left + " "
	}
	if len(right) > 0 {
		right = " " + right
	}

	return fmt.Sprintf("%s%s.%s", nonTermPhrase, left, right)
}

// String returns the string representation of an LR1 Item.
func (item LR1Item) String() string {
	return fmt.Sprintf("%s, %s", item.LR0Item.String(), item.Lookahead)
}
