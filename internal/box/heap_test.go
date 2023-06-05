package box

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Heap_Add(t *testing.T) {
	testCases := []struct {
		name   string
		heap   *Heap[int]
		add    int
		expect string
	}{
		{
			name:   "add to empty",
			heap:   NewMaxHeap[int](),
			add:    2,
			expect: `(2)`,
		},
		{
			name: "add larger value to heap of len 1",
			heap: NewMaxHeap(2),
			add:  4,
			expect: `(4)
  \---: (2)`,
		},
		{
			name: "add larger value to heap of len 2",
			heap: NewMaxHeap(2, 4),
			add:  6,
			expect: `(6)
  |---: (2)
  \---: (4)`,
		},
		{
			name: "add larger value to heap of len 3",
			heap: NewMaxHeap(2, 4, 6),
			add:  32,
			expect: `(32)
  |---: (6)
  |       \---: (2)
  \---: (4)`,
		},
		{
			name: "add larger value to heap of len 4",
			heap: NewMaxHeap(2, 4, 6, 32),
			add:  60,
			expect: `(60)
  |---: (32)
  |       |---: (2)
  |       \---: (6)
  \---: (4)`,
		},
		{
			name: "add larger value to heap of len 5",
			heap: NewMaxHeap(2, 4, 6, 32, 60),
			add:  65,
			expect: `(65)
  |---: (32)
  |       |---: (2)
  |       \---: (6)
  \---: (60)
          \---: (4)`,
		},
		{
			name: "add larger value to heap of len 6",
			heap: NewMaxHeap(2, 4, 6, 32, 60, 65),
			add:  70,
			expect: `(70)
  |---: (32)
  |       |---: (2)
  |       \---: (6)
  \---: (65)
          |---: (4)
          \---: (60)`,
		},
		{
			name: "add medium value to heap of len 6",
			heap: NewMaxHeap(2, 4, 6, 32, 60, 65),
			add:  20,
			expect: `(65)
  |---: (32)
  |       |---: (2)
  |       \---: (6)
  \---: (60)
          |---: (4)
          \---: (20)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			tc.heap.Add(tc.add)

			actualStr := tc.heap.String()
			assert.Equal(tc.expect, actualStr)
		})
	}
}

func Test_Heap_PushManyThenPopAll(t *testing.T) {
	// just make sure it doesn't panic on us
	assert := assert.New(t)

	for numAdds := 0; numAdds < 100; numAdds++ {
		const spacingFactor = 3

		h := NewMaxHeap[int]()

		for i := 0; i < numAdds; i++ {
			addMe := i*spacingFactor + rand.Intn(spacingFactor*2)
			h.Add(addMe)
		}

		assert.Equal(numAdds, h.count)
		assert.Equal(numAdds, h.Len())

		for i := 0; i < numAdds; i++ {
			h.Pop()
		}

		assert.Equal(0, h.Len())
	}
}

func Test_Heap_Pop(t *testing.T) {
	testCases := []struct {
		name        string
		heap        *Heap[int]
		expectValue int
		expectHeap  string
	}{
		{
			name:        "heap len 6",
			heap:        NewMaxHeap(8, 8, 12, 2, 59, 8),
			expectValue: 59,
			expectHeap: `(12)
  |---: (8)
  |       |---: (2)
  |       \---: (8)
  \---: (8)`,
		},
		{
			name:        "heap len 7",
			heap:        NewMaxHeap(8, 8, 12, 2, 59, 8, 88),
			expectValue: 88,
			expectHeap: `(59)
  |---: (12)
  |       |---: (2)
  |       \---: (8)
  \---: (8)
          \---: (8)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual := tc.heap.Pop()

			actualHeapStr := tc.heap.String()

			assert.Equal(tc.expectValue, actual)
			assert.Equal(tc.expectHeap, actualHeapStr)
		})
	}
}
