package grammar

import (
	"fmt"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/decbin"
	"github.com/dekarrin/ictiobus/internal/matrix"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/rosed"
)

// TODO: this should probably be in a different package, probably `parse`,
// specifically in the rather light ll1.go file, glub.

type LL1Table matrix.Matrix2[string, string, Production]

func (M LL1Table) MarshalBinary() ([]byte, error) {
	var data []byte
	xOrdered := textfmt.OrderedKeys(M)

	data = append(data, decbin.EncInt(len(M))...)
	for _, x := range xOrdered {
		col := M[x]
		data = append(data, decbin.EncString(x)...)
		data = append(data, decbin.EncMapStringToBinary(col)...)
	}

	return data, nil
}

func (M *LL1Table) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	newM := LL1Table{}

	var numEntries int
	numEntries, n, err = decbin.DecInt(data)
	if err != nil {
		return err
	}
	data = data[n:]

	for i := 0; i < numEntries; i++ {
		var x string
		x, n, err = decbin.DecString(data)
		if err != nil {
			return err
		}
		data = data[n:]

		var ptrMap map[string]*Production
		ptrMap, n, err = decbin.DecMapStringToBinary[*Production](data)
		if err != nil {
			return err
		}
		data = data[n:]

		newMap := map[string]Production{}
		for k := range ptrMap {
			newMap[k] = *ptrMap[k]
		}
		newM[x] = newMap
	}

	*M = newM
	return nil
}

func (M LL1Table) Set(A string, a string, alpha Production) {
	matrix.Matrix2[string, string, Production](M).Set(A, a, alpha)
}

func (M LL1Table) String() string {
	data := [][]string{}

	terms := M.Terminals()
	nts := M.NonTerminals()

	topRow := []string{""}
	for i := range terms {
		topRow = append(topRow, terms[i])
	}
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
		}).
		String()
}

// Get returns an empty Production if it does not exist, or the one at the
// given coords.
func (M LL1Table) Get(A string, a string) Production {
	v := matrix.Matrix2[string, string, Production](M).Get(A, a)
	if v == nil {
		return Error
	}
	return *v
}

// NonTerminals returns all non-terminals used as the X keys for values in this
// table.
func (M LL1Table) NonTerminals() []string {
	return textfmt.OrderedKeys(M)
}

// Terminals returns all terminals used as the Y keys for values in this table.
// Note that the "$" is expected to be present in all LL1 prediction tables.
func (M LL1Table) Terminals() []string {
	termSet := map[string]bool{}

	for k := range M {
		subMap := map[string]map[string]Production(M)[k]

		for term := range subMap {
			termSet[term] = true
		}
	}

	return textfmt.OrderedKeys(termSet)
}

func NewLL1Table() LL1Table {
	return LL1Table(matrix.NewMatrix2[string, string, Production]())
}

// LLParseTable builds and returns the LL parsing table for the grammar. If it's
// not an LL(1) grammar, returns error.
//
// This is an implementation of Algorithm 4.31, "Construction of a predictive
// parsing table" from the peerple deruuuuugon beeeeeerk. (purple dragon book
// glub)
func (g Grammar) LLParseTable() (M LL1Table, err error) {
	if !g.IsLL1() {
		return nil, fmt.Errorf("not an LL(1) grammar")
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
