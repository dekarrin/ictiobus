package matrix

// Matrix2 is a 2d mapping of coordinates to values. Do not use a Matrix2 by
// itself, use NewMatrix2.
type Matrix2[EX, EY comparable, V any] map[EX]map[EY]V

func NewMatrix2[EX, EY comparable, V any]() Matrix2[EX, EY, V] {
	return map[EX]map[EY]V{}
}

func (m2 Matrix2[EX, EY, V]) Set(x EX, y EY, value V) {
	if m2 == nil {
		panic("assignment to nil Matrix2")
	}

	col, colExists := m2[x]
	if !colExists {
		col = map[EY]V{}
		m2[x] = col
	}
	col[y] = value
	m2[x] = col
}

// Get gets a pointer to the value pointed to by the coordinates. If not at
// those coordinates, nil is returned. Updating the value of the pointer will
// not update the matrix; use Set for that.
func (m2 Matrix2[EX, EY, V]) Get(x EX, y EY) *V {
	if m2 == nil {
		return nil
	}

	col, colExists := m2[x]
	if !colExists {
		return nil
	}

	val, valExists := col[y]
	if !valExists {
		return nil
	}

	return &val
}
