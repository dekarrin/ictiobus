package grammar

import (
	"fmt"
	"sort"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/rosed"
)

// LL1Table is a table for LL predictive parsing. It should not be used
// directly and should be obtained by calling NewLL1Table().
type LL1Table struct {
	d *box.Matrix2[string, Production]
}

// NewLL1Table creates an LL1 table initialized with an empty matrix.
func NewLL1Table() LL1Table {
	return LL1Table{
		d: box.NewMatrix2[string, Production](),
	}
}

func (M LL1Table) MarshalBinary() ([]byte, error) {
	var data []byte

	xOrdered := M.d.DefinedXs()
	yOrdered := M.d.DefinedYs()

	sort.Strings(xOrdered)
	sort.Strings(yOrdered)

	data = append(data, rezi.EncInt(M.d.Width())...)
	for _, x := range xOrdered {
		col := map[string]Production{}

		for _, y := range yOrdered {
			var val *Production = M.d.Get(x, y)
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

func (M *LL1Table) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	newM := NewLL1Table()

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

		var ptrMap map[string]*Production
		ptrMap, n, err = rezi.DecMapStringToBinary[*Production](data)
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
func (M LL1Table) Set(A string, a string, alpha Production) {
	M.d.Set(A, a, alpha)
}

// String returns the string representation of the LL1Table.
func (M LL1Table) String() string {
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
func (M LL1Table) Get(A string, a string) Production {
	v := M.d.Get(A, a)
	if v == nil {
		return Error
	}
	return *v
}

// NonTerminals returns all non-terminals used as the X keys for values in this
// table.
func (M LL1Table) NonTerminals() []string {
	xOrdered := M.d.DefinedXs()
	sort.Strings(xOrdered)
	return xOrdered
}

// Terminals returns all terminals used as the Y keys for values in this table.
// Note that the "$" is expected to be present in all LL1 prediction tables.
func (M LL1Table) Terminals() []string {
	yOrdered := M.d.DefinedYs()
	sort.Strings(yOrdered)
	return yOrdered
}

// LLParseTable builds and returns the LL parsing table for the grammar. If it's
// not an LL(1) grammar, returns error.
//
// This is an implementation of Algorithm 4.31, "Construction of a predictive
// parsing table" from the peerple deruuuuugon beeeeeerk. (purple dragon book
// glub)
func (g Grammar) LLParseTable() (M LL1Table, err error) {
	if !g.IsLL1() {
		return M, fmt.Errorf("not an LL(1) grammar")
	}

	nts := g.NonTerminals()
	M = NewLL1Table()

	// For each production A -> α of the grammar, do the following:
	// -purple dragon book
	for _, A := range nts {
		ARule := g.Rule(A)
		for _, alpha := range ARule.Productions {
			FIRSTalpha := box.StringSetOf(g.FIRST(alpha[0]).Elements())

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
				if a != Epsilon[0] {
					M.Set(A, a, alpha)
				}
			}

			// 2. If ε is in FIRST(α), then for each terminal b in FOLLOW(A),
			// add A -> α to M[A, b]. If ε is in FIRST(α) and $ is in FOLLOW(A),
			// add A -> α to M[A, $] as well.
			if FIRSTalpha.Has(Epsilon[0]) {
				for _, b := range g.FOLLOW(A).Elements() {
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
