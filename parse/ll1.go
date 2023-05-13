package parse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dekarrin/rosed"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/types"
)

type ll1Parser struct {
	table ll1Table
	g     grammar.Grammar
	trace func(s string)
}

// Grammar returns the grammar that was used to generate the parser.
func (ll *ll1Parser) Grammar() grammar.Grammar {
	return ll.g
}

// DFAString would normally return a string representation of the DFA that
// drives the parser, but LL(1) parsers do not generally construct a DFA, and
// so this returns a string indicating such.
func (ll *ll1Parser) DFAString() string {
	return "(LL top-down parser does not use a DFA)"
}

// RegisterTraceListener sets a function to be called with messages that
// indicate what action the parser is taking. It is useful for debug purposes.
func (ll *ll1Parser) RegisterTraceListener(listener func(s string)) {
	ll.trace = listener
}

// TableString returns the parser table as a string.
func (ll *ll1Parser) TableString() string {
	return ll.table.String()
}

// MarshalBinary converts ll into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (ll *ll1Parser) MarshalBinary() ([]byte, error) {
	data := rezi.EncBinary(ll.table)
	data = append(data, rezi.EncBinary(ll.g)...)
	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into ll.
// All of ll's fields will be replaced by the fields decoded from data.
func (ll *ll1Parser) UnmarshalBinary(data []byte) error {
	n, err := rezi.DecBinary(data, &ll.table)
	if err != nil {
		return fmt.Errorf("table: %w", err)
	}
	data = data[n:]

	_, err = rezi.DecBinary(data, &ll.g)
	if err != nil {
		return fmt.Errorf("g: %w", err)
	}

	return nil
}

// EmptyLL1Parser returns a completely empty LL1 parser, unsuitable for use.
// Generally this should not be used directly except for internal purposes; use
// GenerateLL1Parser to generate one ready for use.
func EmptyLL1Parser() *ll1Parser {
	return &ll1Parser{}
}

// GenerateLL1Parser generates a parser for LL1 grammar g. The grammar must
// already be LL1 or convertible to an LL1 grammar.
//
// The returned parser parses the input using LL(k) parsing rules on the
// context-free Grammar g (k=1). The grammar must already be LL(1); it will not
// be forced to it.
func GenerateLL1Parser(g grammar.Grammar) (*ll1Parser, error) {
	M, err := generateLL1ParseTable(g)
	if err != nil {
		return &ll1Parser{}, err
	}
	return &ll1Parser{table: M, g: g.Copy()}, nil
}

// Type returns the type of the parser. This will be ParserLL1 for an
// LL(1)-parser.
func (ll1 *ll1Parser) Type() types.ParserType {
	return types.ParserLL1
}

func (ll1 ll1Parser) notifyPopped(s string) {
	if ll1.trace != nil {
		ll1.trace(fmt.Sprintf("popped %q", s))
	}
}

func (ll1 ll1Parser) notifyPushed(s string) {
	if ll1.trace != nil {
		ll1.trace(fmt.Sprintf("pushed %q", s))
	}
}

// Parse takes a stream of tokens and parses it into a parse tree. If any syntax
// errors are encountered, an empty parse tree and a *types.SyntaxError is
// returned.
func (ll1 *ll1Parser) Parse(stream types.TokenStream) (types.ParseTree, error) {
	symStack := box.NewStack([]string{ll1.g.StartSymbol(), "$"})
	next := stream.Peek()
	X := symStack.Peek()
	ll1.notifyPopped(X)
	pt := types.ParseTree{Value: ll1.g.StartSymbol()}
	ptStack := box.NewStack([]*types.ParseTree{&pt})

	node := ptStack.Peek()
	for X != "$" { /* stack is not empty */
		if strings.ToLower(X) == X {
			stream.Next()

			// is terminals
			t := ll1.g.Term(X)
			if next.Class().ID() == t.ID() {
				node.Terminal = true
				node.Source = next
				symStack.Pop()
				X = symStack.Peek()
				ll1.notifyPopped(X)
				ptStack.Pop()
				node = ptStack.Peek()
			} else {
				return pt, types.NewSyntaxErrorFromToken(fmt.Sprintf("There should be a %s here, but it was %q!", t.Human(), next.Lexeme()), next)
			}

			next = stream.Peek()
		} else {
			nextProd := ll1.table.Get(X, ll1.g.TermFor(next.Class()))
			if nextProd.Equal(grammar.Error) {
				return pt, types.NewSyntaxErrorFromToken(fmt.Sprintf("It doesn't make any sense to put a %q here!", next.Class().Human()), next)
			}

			symStack.Pop()
			ptStack.Pop()
			for i := len(nextProd) - 1; i >= 0; i-- {
				if nextProd[i] != grammar.Epsilon[0] {
					symStack.Push(nextProd[i])
					ll1.notifyPushed(nextProd[i])
				}

				child := &types.ParseTree{Value: nextProd[i]}
				if nextProd[i] == grammar.Epsilon[0] {
					child.Terminal = true
				}
				node.Children = append([]*types.ParseTree{child}, node.Children...)

				if nextProd[i] != grammar.Epsilon[0] {
					ptStack.Push(child)
				}
			}

			X = symStack.Peek()
			ll1.notifyPopped(X)

			// node stack will always be one smaller than symbol stack bc
			// glub, we dont put a node onto the stack for "$".
			if X != "$" {
				node = ptStack.Peek()
			}
		}
	}

	return pt, nil
}

// ll1Table is a table for LL predictive parsing. It should not be used
// directly and should be obtained by calling newLLTable().
type ll1Table struct {
	d *box.Matrix2[string, grammar.Production]
}

func newLL1Table() ll1Table {
	return ll1Table{
		d: box.NewMatrix2[string, grammar.Production](),
	}
}

// generateLL1ParseTable builds and returns the LL parsing table for the grammar.
// If it's not an LL(1) grammar, returns error.
//
// This is an implementation of Algorithm 4.31, "Construction of a predictive
// parsing table" from the peerple deruuuuugon beeeeeerk. (purple dragon book
// glub)
func generateLL1ParseTable(g grammar.Grammar) (M ll1Table, err error) {
	if !IsLL1(g) {
		return M, fmt.Errorf("not an LL(1) grammar")
	}

	nts := g.NonTerminals()
	M = newLL1Table()

	// For each production A -> α of the grammar, do the following:
	// -purple dragon book
	for _, A := range nts {
		ARule := g.Rule(A)
		for _, alpha := range ARule.Productions {
			FIRSTalpha := box.StringSetOf(findFIRSTSet(g, alpha[0]).Elements())

			// 1. For each terminal a in FIRST(A), add A -> α to M[A, a].
			// -purple dragon book
			//
			// (this LOOKS like a typo in that actually following terminology
			// in these comments, FIRST(A) means "FIRST OF ALL PRODUCTIONS OF
			// A" but specifically in this section of the book, this
			// terminalogy means ONLY the first set of production we are looking
			// at. So really this is a in FIRST(α) by the convention used in
			// these comments, but purple dragon calls it FIRST(A), which is
			// technically correct within the bounds of "For each production
			// A -> α").
			for a := range FIRSTalpha {
				if a != grammar.Epsilon[0] {
					M.Set(A, a, alpha)
				}
			}

			// 2. If ε is in FIRST(α), then for each terminal b in FOLLOW(A),
			// add A -> α to M[A, b]. If ε is in FIRST(α) and $ is in FOLLOW(A),
			// add A -> α to M[A, $] as well.
			if FIRSTalpha.Has(grammar.Epsilon[0]) {
				for _, b := range findFOLLOWSet(g, A).Elements() {
					// we cover the $ case automatically by not rly caring about
					// them bein glubbin terminals to begin w. W3 SH3LL H4V3
					// 33LQU4L1TY >38]
					M.Set(A, b, alpha)
				}
			}
		}
	}

	return M, nil
}

// MarshalBinary converts M into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (M ll1Table) MarshalBinary() ([]byte, error) {
	var data []byte

	xOrdered := M.d.DefinedXs()
	yOrdered := M.d.DefinedYs()

	sort.Strings(xOrdered)
	sort.Strings(yOrdered)

	data = append(data, rezi.EncInt(M.d.Width())...)
	for _, x := range xOrdered {
		col := map[string]grammar.Production{}

		for _, y := range yOrdered {
			var val *grammar.Production = M.d.Get(x, y)
			if val == nil {
				continue
			}
			col[y] = *val
		}

		data = append(data, rezi.EncString(x)...)
		data = append(data, rezi.EncMapStringToBinary(col)...)
	}

	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into M. All
// of M's fields will be replaced by the fields decoded from data.
func (M *ll1Table) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	newM := newLL1Table()

	var numEntries int
	numEntries, n, err = rezi.DecInt(data)
	if err != nil {
		return err
	}
	data = data[n:]

	for i := 0; i < numEntries; i++ {
		var x string
		x, n, err = rezi.DecString(data)
		if err != nil {
			return err
		}
		data = data[n:]

		var ptrMap map[string]*grammar.Production
		ptrMap, n, err = rezi.DecMapStringToBinary[*grammar.Production](data)
		if err != nil {
			return err
		}
		data = data[n:]

		for y := range ptrMap {
			newM.d.Set(x, y, *ptrMap[y])
		}
	}

	*M = newM
	return nil
}

// Set sets the production to use given symbol A and input symbol a.
func (M ll1Table) Set(A string, a string, alpha grammar.Production) {
	M.d.Set(A, a, alpha)
}

// String returns the string representation of the LL1Table.
func (M ll1Table) String() string {
	data := [][]string{}

	terms := M.Terminals()
	nts := M.NonTerminals()

	topRow := []string{""}
	topRow = append(topRow, terms...)
	data = append(data, topRow)

	for i := range nts {
		dataRow := []string{nts[i]}
		for j := range terms {
			prod := M.Get(nts[i], terms[j])
			dataRow = append(dataRow, prod.String())
		}
		data = append(data, dataRow)
	}

	return rosed.Edit("").
		InsertTableOpts(0, data, 80, rosed.Options{
			TableBorders: true,
			TableHeaders: true,
		}).
		String()
}

// Get returns an empty Production if it does not exist, or the one at the
// given coords.
func (M ll1Table) Get(A string, a string) grammar.Production {
	v := M.d.Get(A, a)
	if v == nil {
		return grammar.Error
	}
	return *v
}

// NonTerminals returns all non-terminals used as the X keys for values in this
// table.
func (M ll1Table) NonTerminals() []string {
	xOrdered := M.d.DefinedXs()
	sort.Strings(xOrdered)
	return xOrdered
}

// Terminals returns all terminals used as the Y keys for values in this table.
// Note that the "$" is expected to be present in all LL1 prediction tables.
func (M ll1Table) Terminals() []string {
	yOrdered := M.d.DefinedYs()
	sort.Strings(yOrdered)
	return yOrdered
}
