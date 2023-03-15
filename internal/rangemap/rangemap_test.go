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

func Test_RangeMap_Count(t *testing.T) {
	testCases := []struct {
		name   string
		rm     RangeMap[int]
		expect int
	}{
		{
			name:   "empty",
			rm:     RangeMap[int]{},
			expect: 0,
		},
		{
			name:   "non-empty",
			rm:     RangeMap[int]{count: 20},
			expect: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := tc.rm.Count()

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_RangeMap_Call(t *testing.T) {
	testCases := []struct {
		name        string
		rm          RangeMap[int]
		input       int
		expect      int
		expectPanic bool
	}{
		{
			name: "one range, input within domain",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
				},
				count: 11,
			},
			input:  5,
			expect: 15,
		},
		{
			name: "one range, input 0",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
				},
				count: 11,
			},
			input:  0,
			expect: 10,
		},
		{
			name: "one range, input max",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
				},
				count: 11,
			},
			input:  10,
			expect: 20,
		},
		{
			name: "two ranges",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
					{Lo: 11, Hi: 13},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
					{Lo: 200, Hi: 202},
				},
				count: 14,
			},
			input:  0,
			expect: 10,
		},
		{
			name: "two ranges, input specifies start of second",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
					{Lo: 11, Hi: 13},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
					{Lo: 200, Hi: 202},
				},
				count: 14,
			},
			input:  11,
			expect: 200,
		},
		{
			name: "two ranges, input max",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
					{Lo: 11, Hi: 13},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
					{Lo: 200, Hi: 202},
				},
				count: 14,
			},
			input:  13,
			expect: 202,
		},
		{
			name: "two ranges, input middle of second",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
					{Lo: 11, Hi: 13},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
					{Lo: 200, Hi: 202},
				},
				count: 14,
			},
			input:  12,
			expect: 201,
		},
		{
			name: "panics: one range, input < 0",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
				},
				count: 11,
			},
			input:       -1,
			expectPanic: true,
		},
		{
			name: "panics: one range, input >= Count()",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 10},
				},
				ranges: []Range[int]{
					{Lo: 10, Hi: 20},
				},
				count: 11,
			},
			input:       11,
			expectPanic: true,
		},
		{
			name:        "panics: empty RangeMap",
			rm:          RangeMap[int]{},
			input:       10,
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual int
			if tc.expectPanic {
				assert.Panics(func() {
					actual = tc.rm.Call(tc.input)
				})
			} else {
				actual = tc.rm.Call(tc.input)
			}

			assert.Equal(tc.expect, actual)
		})
	}
}

func Test_RangeMap_Intersection(t *testing.T) {
	testCases := []struct {
		name        string
		rm          RangeMap[int]
		input       RangeMap[int]
		expect      RangeMap[int]
		expectPanic bool
	}{
		{
			name:   "empty and empty - no intersections",
			rm:     RangeMap[int]{},
			input:  RangeMap[int]{},
			expect: RangeMap[int]{},
		},
		{
			name: "basic test, one with another",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{0, 5},
					{6, 10},
				},
				ranges: []Range[int]{
					{20, 25},
					{36, 40},
				},
				count: 11,
			},
			input: RangeMap[int]{
				domains: []Range[int]{
					{0, 6},
				},
				ranges: []Range[int]{
					{22, 28},
				},
				count: 7,
			},
			expect: RangeMap[int]{
				domains: []Range[int]{
					{0, 3},
				},
				ranges: []Range[int]{
					{22, 25},
				},
				count: 4,
			},
		},
		{
			name: "no intersections",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{0, 5},
					{6, 10},
				},
				ranges: []Range[int]{
					{20, 25},
					{36, 40},
				},
				count: 11,
			},
			input: RangeMap[int]{
				domains: []Range[int]{
					{0, 2},
					{3, 5},
					{6, 11},
				},
				ranges: []Range[int]{
					{26, 28},
					{30, 32},
					{45, 50},
				},
				count: 12,
			},
			expect: RangeMap[int]{},
		},
		{
			name: "multiple intersections",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{0, 12},
					{13, 27},
					{28, 38},
				},
				ranges: []Range[int]{
					{8, 20},
					{36, 50},
					{60, 70},
				},
				count: 39,
			},
			input: RangeMap[int]{
				domains: []Range[int]{
					{0, 12},
					{13, 15},
					{16, 18},
					{19, 23},
					{24, 24},
					{25, 27},
				},
				ranges: []Range[int]{
					{8, 20},
					{36, 38},
					{40, 42},
					{48, 52},
					{62, 62},
					{208, 210},
				},
				count: 28,
			},
			expect: RangeMap[int]{
				domains: []Range[int]{
					{0, 12},
					{13, 15},
					{16, 18},
					{19, 21},
					{22, 22},
				},
				ranges: []Range[int]{
					{8, 20},
					{36, 38},
					{40, 42},
					{48, 50},
					{62, 62},
				},
				count: 23,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			var actual *RangeMap[int]
			if tc.expectPanic {
				assert.Panics(func() {
					actual = tc.rm.Intersection(&tc.input)
				})
			} else {
				actual = tc.rm.Intersection(&tc.input)
			}

			assert.Equal(&tc.expect, actual)
		})
	}
}

func Test_RangeMap_Add(t *testing.T) {
	testCases := []struct {
		name        string
		rm          RangeMap[int]
		start       int
		end         int
		expect      RangeMap[int]
		expectPanic bool
	}{
		{
			name:  "add (0, 2) to empty",
			rm:    RangeMap[int]{},
			start: 0,
			end:   2,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 2},
				},
				ranges: []Range[int]{
					{Lo: 0, Hi: 2},
				},
				count: 3,
			},
		},
		{
			name:  "add (413, 612) to empty",
			rm:    RangeMap[int]{},
			start: 413,
			end:   612,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 199},
				},
				ranges: []Range[int]{
					{Lo: 413, Hi: 612},
				},
				count: 200,
			},
		},
		{
			name: "new before one existing: add (1, 8) to existing {(413, 612)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 199},
				},
				ranges: []Range[int]{
					{Lo: 413, Hi: 612},
				},
				count: 200,
			},
			start: 1,
			end:   8,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 207},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 413, Hi: 612},
				},
				count: 208,
			},
		},
		{
			name: "new after one existing: add (413, 612) to existing {(1, 8)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
			start: 413,
			end:   612,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 207},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 413, Hi: 612},
				},
				count: 208,
			},
		},
		{
			name: "new after two existing: add (615, 620) to existing {(1, 8), (413, 612)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 207},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 413, Hi: 612},
				},
				count: 208,
			},
			start: 615,
			end:   620,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 207},
					{Lo: 208, Hi: 213},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 413, Hi: 612},
					{Lo: 615, Hi: 620},
				},
				count: 214,
			},
		},
		{
			name: "new before two existing: add (1, 8) to existing {(413, 612), (615, 620)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 199},
					{Lo: 200, Hi: 205},
				},
				ranges: []Range[int]{
					{Lo: 413, Hi: 612},
					{Lo: 615, Hi: 620},
				},
				count: 206,
			},
			start: 1,
			end:   8,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 207},
					{Lo: 208, Hi: 213},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 413, Hi: 612},
					{Lo: 615, Hi: 620},
				},
				count: 214,
			},
		},
		{
			name: "new between two existing: add (413, 612) to existing {(1, 8), (615, 620)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 13},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 615, Hi: 620},
				},
				count: 14,
			},
			start: 413,
			end:   612,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 207},
					{Lo: 208, Hi: 213},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 413, Hi: 612},
					{Lo: 615, Hi: 620},
				},
				count: 214,
			},
		},

		// 	   r1.Lo ------------ r1.Hi
		// e2.Lo --------------------- e2.Hi
		{
			name: "add range completely within existing range: add (2, 6) to existing {(1, 8)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
			start: 2,
			end:   6,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
		},

		// r1.Lo -------------------- r1.Hi
		//     e2.Lo ----------- e2.Hi
		{
			name: "add range that completely contains existing range: add (1, 8) to existing {(2, 6)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
				},
				ranges: []Range[int]{
					{Lo: 2, Hi: 6},
				},
				count: 5,
			},
			start: 1,
			end:   8,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
		},

		// r1.Lo -------------------- r1.Hi
		// e2.Lo -------------------- e2.Hi
		{
			name: "add range that completely equals existing range: add (1, 8) to existing {(1, 8)}",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
			start: 1,
			end:   8,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
		},

		{
			name: "start of r is inside an existing range, but end is outside all ranges (no other overlap)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
			start: 2,
			end:   10,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 9},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 10},
				},
				count: 10,
			},
		},
		{
			name: "start of r is inside an existing range, but end is outside all ranges (no other overlap, one other item before)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 1},
					{Lo: 2, Hi: 9},
				},
				ranges: []Range[int]{
					{Lo: -2, Hi: -1},
					{Lo: 1, Hi: 8},
				},
				count: 10,
			},
			start: 2,
			end:   10,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 1},
					{Lo: 2, Hi: 11},
				},
				ranges: []Range[int]{
					{Lo: -2, Hi: -1},
					{Lo: 1, Hi: 10},
				},
				count: 12,
			},
		},
		{
			name: "start of r is inside an existing range, but end is outside all ranges (overlap one)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 13},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 615, Hi: 620},
				},
				count: 14,
			},
			start: 2,
			end:   621,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 620},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 621},
				},
				count: 621,
			},
		},
		{
			name: "start of r is inside an existing range, but end is outside all ranges (overlap one, one other item before)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 13},
					{Lo: 14, Hi: 16},
				},
				ranges: []Range[int]{
					{Lo: -8, Hi: -1},
					{Lo: 615, Hi: 620},
					{Lo: 628, Hi: 630},
				},
				count: 17,
			},
			start: 616,
			end:   631,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 24},
				},
				ranges: []Range[int]{
					{Lo: -8, Hi: -1},
					{Lo: 615, Hi: 631},
				},
				count: 25,
			},
		},
		{
			name: "start of r is inside an existing range, but end is outside all ranges (overlap one, one other item after)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 13},
					{Lo: 14, Hi: 19},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 615, Hi: 620},
					{Lo: 650, Hi: 655},
				},
				count: 19,
			},
			start: 2,
			end:   621,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 620},
					{Lo: 621, Hi: 626},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 621},
					{Lo: 650, Hi: 655},
				},
				count: 626,
			},
		},
		{
			name: "start of r is inside an existing range, but end is outside all ranges (overlap two)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
					{Lo: 8, Hi: 13},
					{Lo: 14, Hi: 23},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
					{Lo: 615, Hi: 620},
					{Lo: 1001, Hi: 1010},
				},
				count: 24,
			},
			start: 2,
			end:   1011,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 1010},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 1011},
				},
				count: 1011,
			},
		},

		{
			name: "end of r is inside an existing range, but start is outside all ranges (no other overlap)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 7},
				},
				ranges: []Range[int]{
					{Lo: 1, Hi: 8},
				},
				count: 8,
			},
			start: -10,
			end:   2,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 18},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: 8},
				},
				count: 19,
			},
		},
		{
			name: "end of r is inside an existing range, but start is outside all ranges (no other overlap, one other item before)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
					{Lo: 5, Hi: 12},
				},
				ranges: []Range[int]{
					{Lo: -16, Hi: -12},
					{Lo: 1, Hi: 8},
				},
				count: 13,
			},
			start: -10,
			end:   2,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
					{Lo: 5, Hi: 23},
				},
				ranges: []Range[int]{
					{Lo: -16, Hi: -12},
					{Lo: -10, Hi: 8},
				},
				count: 24,
			},
		},
		{
			name: "end of r is inside an existing range, but start is outside all ranges (overlap one)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
					{Lo: 5, Hi: 12},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: -6},
					{Lo: 1, Hi: 8},
				},
				count: 13,
			},
			start: -11,
			end:   2,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 19},
				},
				ranges: []Range[int]{
					{Lo: -11, Hi: 8},
				},
				count: 20,
			},
		},
		{
			name: "end of r is inside an existing range, but start is outside all ranges (overlap one, one other item before)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
					{Lo: 5, Hi: 12},
					{Lo: 13, Hi: 18},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: -6},
					{Lo: 1, Hi: 8},
					{Lo: 10, Hi: 15},
				},
				count: 19,
			},
			start: -4,
			end:   12,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
					{Lo: 5, Hi: 24},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: -6},
					{Lo: -4, Hi: 15},
				},
				count: 25,
			},
		},
		{
			name: "end of r is inside an existing range, but start is outside all ranges (overlap one, one other item after)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
					{Lo: 5, Hi: 12},
					{Lo: 13, Hi: 18},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: -6},
					{Lo: 1, Hi: 8},
					{Lo: 650, Hi: 655},
				},
				count: 19,
			},
			start: -11,
			end:   2,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 19},
					{Lo: 20, Hi: 25},
				},
				ranges: []Range[int]{
					{Lo: -11, Hi: 8},
					{Lo: 650, Hi: 655},
				},
				count: 26,
			},
		},
		{
			name: "end of r is inside an existing range, but start is outside all ranges (overlap two)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 4},
					{Lo: 5, Hi: 12},
					{Lo: 13, Hi: 18},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: -6},
					{Lo: 1, Hi: 8},
					{Lo: 615, Hi: 620},
				},
				count: 19,
			},
			start: -11,
			end:   616,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 631},
				},
				ranges: []Range[int]{
					{Lo: -11, Hi: 620},
				},
				count: 632,
			},
		},

		{
			name: "start of r is inside an existing range, and end is inside other existing range (no other overlap)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 13},
					{Lo: 14, Hi: 21},
				},
				ranges: []Range[int]{
					{Lo: -7, Hi: 6},
					{Lo: 8, Hi: 15},
				},
				count: 22,
			},
			start: 1,
			end:   10,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 22},
				},
				ranges: []Range[int]{
					{Lo: -7, Hi: 15},
				},
				count: 23,
			},
		},
		{
			name: "start of r is inside an existing range, and end is inside other existing range (no other overlap, one other item before)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 3},
					{Lo: 4, Hi: 17},
					{Lo: 18, Hi: 25},
				},
				ranges: []Range[int]{
					{Lo: -12, Hi: -9},
					{Lo: -7, Hi: 6},
					{Lo: 8, Hi: 15},
				},
				count: 26,
			},
			start: 1,
			end:   10,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 3},
					{Lo: 4, Hi: 26},
				},
				ranges: []Range[int]{
					{Lo: -12, Hi: -9},
					{Lo: -7, Hi: 15},
				},
				count: 27,
			},
		},
		{
			name: "start of r is inside an existing range, and end is inside other existing range (overlap one)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 6},
					{Lo: 7, Hi: 10},
					{Lo: 11, Hi: 16},
				},
				ranges: []Range[int]{
					{Lo: 0, Hi: 6},
					{Lo: 8, Hi: 11},
					{Lo: 15, Hi: 20},
				},
				count: 17,
			},
			start: 1,
			end:   17,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 20},
				},
				ranges: []Range[int]{
					{Lo: 0, Hi: 20},
				},
				count: 21,
			},
		},
		{
			name: "start of r is inside an existing range, and end is inside other existing range (overlap one, one other item before)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 3},
					{Lo: 4, Hi: 10},
					{Lo: 11, Hi: 14},
					{Lo: 15, Hi: 20},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: -7},
					{Lo: 0, Hi: 6},
					{Lo: 8, Hi: 11},
					{Lo: 15, Hi: 20},
				},
				count: 21,
			},
			start: 1,
			end:   17,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 3},
					{Lo: 4, Hi: 24},
				},
				ranges: []Range[int]{
					{Lo: -10, Hi: -7},
					{Lo: 0, Hi: 20},
				},
				count: 25,
			},
		},
		{
			name: "start of r is inside an existing range, and end is inside other existing range (overlap one, one other item after)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 6},
					{Lo: 7, Hi: 10},
					{Lo: 11, Hi: 16},
					{Lo: 17, Hi: 20},
				},
				ranges: []Range[int]{
					{Lo: 0, Hi: 6},
					{Lo: 8, Hi: 11},
					{Lo: 15, Hi: 20},
					{Lo: 26, Hi: 29},
				},
				count: 21,
			},
			start: 1,
			end:   17,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 20},
					{Lo: 21, Hi: 24},
				},
				ranges: []Range[int]{
					{Lo: 0, Hi: 20},
					{Lo: 26, Hi: 29},
				},
				count: 25,
			},
		},
		{
			name: "start of r is inside an existing range, and end is inside other existing range (overlap two)",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 6},
					{Lo: 7, Hi: 10},
					{Lo: 11, Hi: 16},
					{Lo: 17, Hi: 20},
				},
				ranges: []Range[int]{
					{Lo: 0, Hi: 6},
					{Lo: 8, Hi: 11},
					{Lo: 15, Hi: 20},
					{Lo: 26, Hi: 29},
				},
				count: 21,
			},
			start: 1,
			end:   28,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 29},
				},
				ranges: []Range[int]{
					{Lo: 0, Hi: 29},
				},
				count: 30,
			},
		},
		{
			name: "normalize adjacent result ranges",
			rm: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 2},
					{Lo: 3, Hi: 4},
				},
				ranges: []Range[int]{
					{Lo: 2, Hi: 4},
					{Lo: 6, Hi: 7},
				},
				count: 5,
			},
			start: 5,
			end:   6,
			expect: RangeMap[int]{
				domains: []Range[int]{
					{Lo: 0, Hi: 5},
				},
				ranges: []Range[int]{
					{Lo: 2, Hi: 7},
				},
				count: 6,
			},
		},
	}

	// TODO: find out why above test panics, then add final test cases for
	// last case in Add.

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			if tc.expectPanic {
				assert.Panics(func() {
					tc.rm.Add(tc.start, tc.end)
				})
				return
			}
			tc.rm.Add(tc.start, tc.end)

			assert.Equal(tc.expect, tc.rm)
		})
	}
}
