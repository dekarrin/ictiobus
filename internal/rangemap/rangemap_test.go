package rangemap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Range_Count(t *testing.T) {
	testCases := []struct {
		name   string
		input  Range[int]
		expect int
	}{
		{
			name:   "empty range (includes 0)",
			input:  Range[int]{},
			expect: 1,
		},
		{
			name:   "one element range, positive",
			input:  Range[int]{Lo: 1, Hi: 1},
			expect: 1,
		},
		{
			name:   "one element range, negative",
			input:  Range[int]{Lo: -1, Hi: -1},
			expect: 1,
		},
		{
			name:   "multi element range, negative",
			input:  Range[int]{Lo: -4, Hi: -1},
			expect: 4,
		},
		{
			name:   "multi element range, positive",
			input:  Range[int]{Lo: 1, Hi: 4},
			expect: 4,
		},
		{
			name:   "multi element range, start at 0",
			input:  Range[int]{Lo: 0, Hi: 4},
			expect: 5,
		},
		{
			name:   "multi element range, end at 0",
			input:  Range[int]{Lo: -4, Hi: 0},
			expect: 5,
		},
		{
			name:   "multi element range, spanning 0",
			input:  Range[int]{Lo: -4, Hi: 4},
			expect: 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := tc.input.Count()

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_Range_Contains(t *testing.T) {
	testCases := []struct {
		name   string
		r      Range[int]
		input  int
		expect bool
	}{
		{
			name:   "0 in [0, 0]",
			r:      Range[int]{Lo: 0, Hi: 0},
			input:  0,
			expect: true,
		},
		{
			name:   "-1 in [0, 0]",
			r:      Range[int]{Lo: 0, Hi: 0},
			input:  -1,
			expect: false,
		},
		{
			name:   "1 in [0, 0]",
			r:      Range[int]{Lo: 0, Hi: 0},
			input:  1,
			expect: false,
		},
		{
			name:   "3 in [-1, 400]",
			r:      Range[int]{Lo: -1, Hi: 400},
			input:  3,
			expect: true,
		},
		{
			name:   "-1 in [-1, 400]",
			r:      Range[int]{Lo: -1, Hi: 400},
			input:  -1,
			expect: true,
		},
		{
			name:   "400 in [-1, 400]",
			r:      Range[int]{Lo: -1, Hi: 400},
			input:  400,
			expect: true,
		},
		{
			name:   "399 in [-1, 400]",
			r:      Range[int]{Lo: -1, Hi: 400},
			input:  399,
			expect: true,
		},
		{
			name:   "0 in [-1, 400]",
			r:      Range[int]{Lo: -1, Hi: 400},
			input:  0,
			expect: true,
		},
		{
			name:   "-2 in [-1, 400]",
			r:      Range[int]{Lo: -1, Hi: 400},
			input:  -2,
			expect: false,
		},
		{
			name:   "401 in [-1, 400]",
			r:      Range[int]{Lo: -1, Hi: 400},
			input:  401,
			expect: false,
		},
		{
			name:   "0 in [1, 400]",
			r:      Range[int]{Lo: 1, Hi: 400},
			input:  0,
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := tc.r.Contains(tc.input)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_Range_SubsetOf(t *testing.T) {
	testCases := []struct {
		name   string
		r1     Range[int]
		r2     Range[int]
		expect bool
	}{
		{
			name:   "[0, 0] subset of [0, 0]",
			r1:     Range[int]{Lo: 0, Hi: 0},
			r2:     Range[int]{Lo: 0, Hi: 0},
			expect: true,
		},
		{
			name:   "[0, 0] subset of [0, 1]",
			r1:     Range[int]{Lo: 0, Hi: 0},
			r2:     Range[int]{Lo: 0, Hi: 1},
			expect: true,
		},
		{
			name:   "[0, 1] subset of [0, 0]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 0, Hi: 0},
			expect: false,
		},
		{
			name:   "[0, 1] subset of [0, 2]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 0, Hi: 2},
			expect: true,
		},
		{
			name:   "[0, 1] subset of [1, 2]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 1, Hi: 2},
			expect: false,
		},
		{
			name:   "[0, 1] subset of [0, 1]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 0, Hi: 1},
			expect: true,
		},
		{
			name:   "[-1, 3] subset of [-1, 400]",
			r1:     Range[int]{Lo: -1, Hi: 3},
			r2:     Range[int]{Lo: -1, Hi: 400},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := tc.r1.SubsetOf(tc.r2)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_Range_Overlaps(t *testing.T) {
	testCases := []struct {
		name   string
		r1     Range[int]
		r2     Range[int]
		expect bool
	}{
		{
			name:   "[0, 0] overlaps [0, 0]",
			r1:     Range[int]{Lo: 0, Hi: 0},
			r2:     Range[int]{Lo: 0, Hi: 0},
			expect: true,
		},
		{
			name:   "[0, 0] overlaps [0, 1]",
			r1:     Range[int]{Lo: 0, Hi: 0},
			r2:     Range[int]{Lo: 0, Hi: 1},
			expect: true,
		},
		{
			name:   "[0, 1] overlaps [0, 0]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 0, Hi: 0},
			expect: true,
		},
		{
			name:   "[0, 1] overlaps [0, 2]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 0, Hi: 2},
			expect: true,
		},
		{
			name:   "[0, 1] overlaps [1, 2]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 1, Hi: 2},
			expect: true,
		},
		{
			name:   "[0, 1] overlaps [0, 1]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 0, Hi: 1},
			expect: true,
		},
		{
			name:   "[-1, 3] overlaps [-1, 400]",
			r1:     Range[int]{Lo: -1, Hi: 3},
			r2:     Range[int]{Lo: -1, Hi: 400},
			expect: true,
		},
		{
			name:   "[0, 1] overlaps [2, 3]",
			r1:     Range[int]{Lo: 0, Hi: 1},
			r2:     Range[int]{Lo: 2, Hi: 3},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := tc.r1.Overlaps(tc.r2)

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_Range_Compare(t *testing.T) {
	testCases := []struct {
		name   string
		r1     Range[int]
		r2     Range[int]
		expect int
	}{

		// r1 == r2 when r1.Lo == r2.Lo && r1.Hi == r2.Hi:

		// r1.Lo --------------------- r1.Hi
		// r2.Lo --------------------- r2.Hi
		{
			name:   "r1 == r2",
			r1:     Range[int]{Lo: 413, Hi: 612},
			r2:     Range[int]{Lo: 413, Hi: 612},
			expect: 0,
		},
		{
			name:   "r1 == r2 (min)",
			r1:     Range[int]{Lo: 0, Hi: 0},
			r2:     Range[int]{Lo: 0, Hi: 0},
			expect: 0,
		},

		// r1 < r2 when r1.Lo < r2.Lo || (r1.Lo == r2.Lo && r1.Hi < r2.Hi):

		// r1.Lo --------------- r1.Hi
		// r2.Lo --------------------- r2.Hi
		{
			name:   "r1 < r2: start at same point, but r1 is shorter",
			r1:     Range[int]{Lo: 20, Hi: 30},
			r2:     Range[int]{Lo: 20, Hi: 40},
			expect: -1,
		},
		{
			name:   "r1 < r2: start at same point, but r1 is shorter (min)",
			r1:     Range[int]{Lo: 20, Hi: 20},
			r2:     Range[int]{Lo: 20, Hi: 21},
			expect: -1,
		},

		// r1.Lo --------------------- r1.Hi
		//         r2.Lo ------------- r2.Hi
		{
			name:   "r1 < r2: end at same point, but r1 starts first",
			r1:     Range[int]{Lo: 20, Hi: 40},
			r2:     Range[int]{Lo: 30, Hi: 40},
			expect: -1,
		},
		{
			name:   "r1 < r2: end at same point, but r1 starts first (min)",
			r1:     Range[int]{Lo: 20, Hi: 21},
			r2:     Range[int]{Lo: 21, Hi: 21},
			expect: -1,
		},

		// r1.Lo --------------- r1.Hi
		//         r2.Lo ------------- r2.Hi
		{
			name:   "r1 < r2: r1 starts before r2 and r1 ends before r2 with overlap",
			r1:     Range[int]{Lo: 20, Hi: 30},
			r2:     Range[int]{Lo: 25, Hi: 40},
			expect: -1,
		},
		{
			name:   "r1 < r2: r1 starts before r2 and r1 ends before r2 with overlap (min)",
			r1:     Range[int]{Lo: 20, Hi: 21},
			r2:     Range[int]{Lo: 21, Hi: 22},
			expect: -1,
		},

		// r1.Lo ----- r1.Hi
		//                   r2.Lo --- r2.Hi
		{
			name:   "r1 < r2: r1 starts before r2 and r1 ends befre r2 with no overlap",
			r1:     Range[int]{Lo: 20, Hi: 30},
			r2:     Range[int]{Lo: 35, Hi: 40},
			expect: -1,
		},
		{
			name:   "r1 < r2: r1 starts before r2 and r1 ends befre r2 with no overlap (min)",
			r1:     Range[int]{Lo: 20, Hi: 21},
			r2:     Range[int]{Lo: 22, Hi: 23},
			expect: -1,
		},

		// r1.Lo --------------------- r1.Hi
		// 	   r2.Lo ------------ r2.Hi
		{
			name:   "r1 < r2: r1 starts before r2 and r1 ends after r2",
			r1:     Range[int]{Lo: 20, Hi: 40},
			r2:     Range[int]{Lo: 25, Hi: 30},
			expect: -1,
		},
		{
			name:   "r1 < r2: r1 starts before r2 and r1 ends after r2 (min)",
			r1:     Range[int]{Lo: 20, Hi: 21},
			r2:     Range[int]{Lo: 21, Hi: 21},
			expect: -1,
		},

		// r1 > r2 when r1.Lo > r2.Lo || (r1.Lo == r2.Lo && r1.Hi > r2.Hi):

		// r1.Lo --------------------- r1.Hi
		// r2.Lo --------------- r2.Hi
		{
			name:   "r1 > r2: start at same point, but r1 is longer",
			r1:     Range[int]{Lo: 20, Hi: 40},
			r2:     Range[int]{Lo: 20, Hi: 30},
			expect: 1,
		},
		{
			name:   "r1 > r2: start at same point, but r1 is longer (min)",
			r1:     Range[int]{Lo: 20, Hi: 21},
			r2:     Range[int]{Lo: 20, Hi: 20},
			expect: 1,
		},

		//       r1.Lo --------------- r1.Hi
		// r2.Lo --------------------- r2.Hi
		{
			name:   "r1 > r2: end at same point, but r1 starts after r2",
			r1:     Range[int]{Lo: 30, Hi: 40},
			r2:     Range[int]{Lo: 20, Hi: 40},
			expect: 1,
		},
		{
			name:   "r1 > r2: end at same point, but r1 starts after r2 (min)",
			r1:     Range[int]{Lo: 21, Hi: 21},
			r2:     Range[int]{Lo: 20, Hi: 21},
			expect: 1,
		},

		//         r1.Lo ------------- r1.Hi
		// r2.Lo --------------- r2.Hi
		{
			name:   "r1 > r2: r1 starts after r2 and r1 ends after r2 with overlap",
			r1:     Range[int]{Lo: 25, Hi: 40},
			r2:     Range[int]{Lo: 20, Hi: 30},
			expect: 1,
		},
		{
			name:   "r1 > r2: r1 starts after r2 and r1 ends after r2 with overlap (min)",
			r1:     Range[int]{Lo: 21, Hi: 22},
			r2:     Range[int]{Lo: 20, Hi: 21},
			expect: 1,
		},

		//                   r1.Lo --- r1.Hi
		// r2.Lo ----- r2.Hi
		{
			name:   "r1 > r2: r1 starts after r2 and r1 ends after r2 with no overlap",
			r1:     Range[int]{Lo: 35, Hi: 40},
			r2:     Range[int]{Lo: 20, Hi: 30},
			expect: 1,
		},
		{
			name:   "r1 > r2: r1 starts after r2 and r1 ends after r2 with no overlap (min)",
			r1:     Range[int]{Lo: 22, Hi: 23},
			r2:     Range[int]{Lo: 20, Hi: 21},
			expect: 1,
		},

		// 	   r1.Lo ------------ r1.Hi
		// r2.Lo --------------------- r2.Hi
		{
			name:   "r1 > r2: r1 starts after r2 and r1 ends before r2",
			r1:     Range[int]{Lo: 25, Hi: 30},
			r2:     Range[int]{Lo: 20, Hi: 40},
			expect: 1,
		},
		{
			name:   "r1 > r2: r1 starts after r2 and r1 ends before r2 (min)",
			r1:     Range[int]{Lo: 21, Hi: 21},
			r2:     Range[int]{Lo: 20, Hi: 22},
			expect: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := tc.r1.Compare(tc.r2)

			if actual < 0 {
				actual = -1
			} else if actual > 0 {
				actual = 1
			}

			if tc.expect < 0 {
				tc.expect = -1
			} else if tc.expect > 0 {
				tc.expect = 1
			}

			assert.Equal(tc.expect, actual)
		})
	}
}

// cases of r1 > r2:
//
//
//       r1.Lo --------------- r1.Hi
// r2.Lo --------------- r2.Hi
//
//                 r1.Lo ----- r1.Hi
// r2.Lo --- r2.Hi
// ...or to put it another way:
//
// Who is greater in this case?

// r2 is
