package box

import (
	"fmt"

	"github.com/dekarrin/ictiobus/internal/slices"
)

// Matrix2 is a 2d mapping of coordinates to values. Values are "sparsish"; an
// X-coordinate specified for some Y is considered "created" for all Ys.
//
// Do not use a Matrix2 by itself, use NewMatrix2.
type Matrix2[K comparable, V any] struct {
	data  map[K]map[K]V
	count int
}

// NewMatrix2 creates a new matrix ready to be used.
func NewMatrix2[K comparable, V any]() *Matrix2[K, V] {
	m2 := Matrix2[K, V]{
		data: make(map[K]map[K]V),
	}
	return &m2
}

// Delete removes a coordinate pair. If it was the last instance of that X for
// all Ys or if it was the last instance of that Y for all Xs, that last
// coordinate's slots are considered removed.
//
// Returns whether a value had existed at the given coordinates.
func (m2 *Matrix2[K, V]) Delete(x, y K) bool {
	if m2 == nil {
		panic("delete from nil Matrix2")
	}

	col, colExists := m2.data[x]
	if !colExists {
		return false
	}

	_, valExists := col[y]
	if !valExists {
		return false
	}

	delete(col, y)
	if len(col) == 0 {
		delete(m2.data, x)
	} else {
		m2.data[x] = col
	}
	m2.count--

	return true
}

// Has returns whether there is currently a value defined at the given
// coordinates. Strictly, it returns whether Get(x, y) would return a nil
// pointer; it will return false if either there is no slot defined for the
// coordinates or if there is a slot but it is not set.
func (m2 *Matrix2[K, V]) Has(x, y K) bool {
	if m2 == nil {
		return false
	}

	col, colExists := m2.data[x]
	if !colExists {
		return false
	}

	_, valExists := col[y]
	return valExists
}

// HasSlot returns whether there is currently a slot defined for the given
// coordinates. Strictly, this is whether any Y value defined has a value for
// the given X, and whether any X has a value for the given Y.
func (m2 *Matrix2[K, V]) HasSlot(x, y K) bool {
	if m2 == nil {
		return false
	}

	col, colExists := m2.data[x]
	if !colExists {
		// x is invalid for all Ys; return false
		return false
	}

	// if the x is valid, y does not necessarily need to be valid for THAT x,
	// only some X. But we will check the given X first
	_, valExists := col[y]
	if valExists {
		return true
	}

	for checkX := range m2.data {
		if checkX == x {
			// already checked that X
			continue
		}

		if _, valExists := m2.data[checkX][y]; valExists {
			return true
		}
	}

	return false
}

// Set sets the value at the given coordinates.
func (m2 *Matrix2[K, V]) Set(x, y K, value V) {
	if m2 == nil {
		panic("assignment to nil Matrix2")
	}

	col, colExists := m2.data[x]
	if !colExists {
		col = map[K]V{}
		m2.data[x] = col
	}

	// does somefin already exist there? increment count if not
	_, valExists := col[y]

	col[y] = value
	m2.data[x] = col

	if !valExists {
		m2.count++
	}
}

// Get gets a pointer to the value pointed to by the coordinates. If not at
// those coordinates, nil is returned. Updating the value of the pointer will
// not update the matrix; use Set for that.
func (m2 *Matrix2[K, V]) Get(x, y K) *V {
	if m2 == nil {
		return nil
	}

	col, colExists := m2.data[x]
	if !colExists {
		return nil
	}

	val, valExists := col[y]
	if !valExists {
		return nil
	}

	return &val
}

// GetDefault gets the value stored at the coordinates. If nothing is defined
// for those coordinates, the default value is returned.
func (m2 *Matrix2[K, V]) GetDefault(x, y K, def V) V {
	if m2 == nil {
		return def
	}

	col, colExists := m2.data[x]
	if !colExists {
		return def
	}

	val, valExists := col[y]
	if !valExists {
		return def
	}

	return val
}

// Width returns the number of defined items for X.
func (m2 *Matrix2[K, V]) Width() int {
	if m2 == nil {
		return 0
	}

	return len(m2.data)
}

// Height returns the number of defined items for Y.
func (m2 *Matrix2[K, V]) Height() int {
	if m2 == nil {
		return 0
	}

	seen := make(map[K]struct{})

	for colName := range m2.data {
		for y := range m2.data[colName] {
			seen[y] = struct{}{}
		}
	}

	return len(seen)
}

// DefinedXs returns the names of all X coordinates currently defined.
func (m2 *Matrix2[K, V]) DefinedXs() []K {
	if m2 == nil {
		return nil
	}

	xNames := []K{}
	for x := range m2.data {
		xNames = append(xNames, x)
	}

	return xNames
}

// DefinedYs returns the names of all Y coordinates currently defined.
func (m2 *Matrix2[K, V]) DefinedYs() []K {
	if m2 == nil {
		return nil
	}

	ySeen := map[K]struct{}{}

	for x := range m2.data {
		for y := range m2.data[x] {
			ySeen[y] = struct{}{}
		}
	}

	defined := []K{}
	for seen := range ySeen {
		defined = append(defined, seen)
	}

	return defined
}

// Len returns the number of total slots in the matrix currently defined by
// rows and columns. This will be width * height and will additionally be the
// length of the slice returned by Elements().
func (m2 *Matrix2[K, V]) Len() int {
	return m2.Height() * m2.Width()
}

// Count returns the number of slots in the matrix with a defined value in them.
// This will be at most Len() and possibly shorter if a value has not been
// defined for every combination of coordinates.
func (m2 *Matrix2[K, V]) Count() int {
	return m2.count
}

// Elements returns all the items in the matrix in sequence. Each row is
// returned in order from the 0th to the last. If there is nothing defined at a
// given coordinate pair, the zero-value for V will be used.
//
// Ordering is defined by the following: If K is a basic numeric type (int,
// int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, byte, rune,
// float32, float64) (complex numbers cannot be directly compared), then it will
// be from the smallest to the largest letter. Otherwise, the coordinates will
// be in collation order of the strings returned by calling fmt.Sprintf("%v") on
// them.
//
// For instance, a Matrix2[int, int] with the following contents:
//
// 1 8 7 6
// 2   4 6
// 0     1
//
// Would return the following from Elements(): []int{1, 8, 7, 6, 2, 0, 4, 6, 0,
// 0, 0, 1}.
func (m2 *Matrix2[K, V]) Elements() []V {
	if m2 == nil {
		return nil
	}

	sortedXs := slices.SortBy(m2.DefinedXs(), m2.matrixKeyCompFunc)
	sortedYs := slices.SortBy(m2.DefinedYs(), m2.matrixKeyCompFunc)

	var allItems []V

	var zeroVal V

	for _, y := range sortedYs {
		for _, x := range sortedXs {
			elem := m2.GetDefault(x, y, zeroVal)
			allItems = append(allItems, elem)
		}
	}

	return allItems
}

func (m2 *Matrix2[K, V]) matrixKeyCompFunc(left, right K) bool {
	var utLeft interface{} = left
	var utRight interface{} = right

	switch utLeft.(type) {
	case int:
		vL := utLeft.(int)
		vR := utRight.(int)
		return vL < vR
	case int8:
		vL := utLeft.(int8)
		vR := utRight.(int8)
		return vL < vR
	case int16:
		vL := utLeft.(int16)
		vR := utRight.(int16)
		return vL < vR
	case int32:
		vL := utLeft.(int32)
		vR := utRight.(int32)
		return vL < vR
	case int64:
		vL := utLeft.(int64)
		vR := utRight.(int64)
		return vL < vR
	case uint:
		vL := utLeft.(uint)
		vR := utRight.(uint)
		return vL < vR
	case uint8:
		vL := utLeft.(uint8)
		vR := utRight.(uint8)
		return vL < vR
	case uint16:
		vL := utLeft.(uint16)
		vR := utRight.(uint16)
		return vL < vR
	case uint32:
		vL := utLeft.(uint32)
		vR := utRight.(uint32)
		return vL < vR
	case float32:
		vL := utLeft.(float32)
		vR := utRight.(float32)
		return vL < vR
	case float64:
		vL := utLeft.(float64)
		vR := utRight.(float64)
		return vL < vR
	default:
		vL := fmt.Sprintf("%v", left)
		vR := fmt.Sprintf("%v", left)
		return vL < vR
	}
}
