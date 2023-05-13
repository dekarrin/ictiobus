package parse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseTree_PathToDiff(t *testing.T) {
	testCases := []struct {
		name          string
		tree          *Tree
		t             *Tree
		ignoreSC      bool
		expect        []int
		expectDiverge bool
	}{
		{
			name:          "zero tree - has no divergence",
			tree:          &Tree{},
			t:             &Tree{},
			expectDiverge: false,
		},
		{
			name:          "simple tree - divergence at root - value",
			tree:          MustParseTreeFromDiagram("[A]"),
			t:             MustParseTreeFromDiagram("[S]"),
			expect:        []int{},
			expectDiverge: true,
		},
		{
			name:          "simple tree - divergence at root - num children",
			tree:          MustParseTreeFromDiagram("[A (a)]"),
			t:             MustParseTreeFromDiagram("[A (a) (b)]"),
			expect:        []int{},
			expectDiverge: true,
		},
		{
			name:          "simple tree - divergence at root - type",
			tree:          MustParseTreeFromDiagram("[A]"),
			t:             MustParseTreeFromDiagram("(A)"),
			expect:        []int{},
			expectDiverge: true,
		},
		{
			name:          "one point of divergence",
			tree:          MustParseTreeFromDiagram("[S  [A (a)]  [B (b)]]"),
			t:             MustParseTreeFromDiagram("[S  [A (b)]  [B (b)]]"),
			expect:        []int{0, 0},
			expectDiverge: true,
		},
		{
			name:          "multiple points of divergence",
			tree:          MustParseTreeFromDiagram("[S  [A (a) (b)]  [B (b)]]"),
			t:             MustParseTreeFromDiagram("[S  [A (b) (a)]  [B (b)]]"),
			expect:        []int{0},
			expectDiverge: true,
		},
		{
			name:          "multiple points of divergence - up to root",
			tree:          MustParseTreeFromDiagram("[S  [A (a)]  [B (b)]]"),
			t:             MustParseTreeFromDiagram("[S  [A (b)]  [A (b)]]"),
			expect:        []int{},
			expectDiverge: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actual, actualDiverge := tc.tree.PathToDiff(*tc.t, tc.ignoreSC)

			assert.Equal(tc.expectDiverge, actualDiverge)
			if tc.expectDiverge {
				assert.Equal(tc.expect, actual)
			}
		})
	}

}

func Test_ParseTree_IsSubTreeOf(t *testing.T) {
	testCases := []struct {
		name       string
		tree       *Tree
		t          *Tree
		expect     bool
		expectPath []int
	}{
		{
			name:       "zero tree is a subtree of zero tree",
			tree:       &Tree{},
			t:          &Tree{},
			expect:     true,
			expectPath: []int{},
		},
		{
			name: "zero tree is not a subtree of empty node with children",
			tree: &Tree{},
			t: &Tree{Children: []*Tree{
				PTLeaf(""),
				PTNode("A"),
				{Value: "X"},
			}},
			expect: false,
		},
		{
			name: "not a sub-tree, t is empty",
			tree: PTNode("S",
				PTNode("A", PTLeaf("a")),
				PTNode("B", PTLeaf("b")),
			),
			t:      PTNode(""),
			expect: false,
		},
		{
			name: "not a sub-tree, t is completely different",
			tree: MustParseTreeFromDiagram(`
				[S
					[A (a)]
					[B (b)]
				]
			`),
			t: MustParseTreeFromDiagram(`
				[E
					[E
						[T
							[T
								[F
									(\()
									[E
										[T
											[F
												(id)
									]]]
									(\))
							]]
							(*)
							[F
								(id)
					]]]
					(+)
					[T
						[F
							(id)
				]]]
			`),
			expect: false,
		},
		{
			name: "not a sub-tree, t is same symbols",
			tree: MustParseTreeFromDiagram(`
				[E
					[E (*)]
					[F (id)]
				]
			`),
			t: MustParseTreeFromDiagram(`
				[E
					[E
						[T
							[T
								[F
									(\()
									[E
										[T
											[F
												(id)
									]]]
									(\))
							]]
							(*)
							[F
								(id)
					]]]
					(+)
					[T
						[F
							(id)
				]]]
			`),
			expect: false,
		},
		{
			name:       "sub-tree match, small search tree",
			tree:       MustParseTreeFromDiagram("[E (*)]"),
			t:          MustParseTreeFromDiagram("[F [E (*)] (+) ]"),
			expect:     true,
			expectPath: []int{0},
		},
		{
			name:       "sub-tree match, exactly equal",
			tree:       MustParseTreeFromDiagram("[E (*) [F (id)]]"),
			t:          MustParseTreeFromDiagram("[E (*) [F (id)]]"),
			expect:     true,
			expectPath: []int{},
		},
		{
			name: "sub-tree match, in large tree",
			tree: MustParseTreeFromDiagram("[E (*) [F (id)]]"),
			t: MustParseTreeFromDiagram(`
				[E
					[F
						(\()
						[E
							(+)
							[E
								[F
									(id)
							]]
							(\))
					]]
					(*)
					[F
						[F
							(&)
							[E (+) [F]]
						]
						(id)
					    [E (*) [F (id)]]
					(+)
					($)
				]]
			`),
			expect:     true,
			expectPath: []int{2, 2},
		},
		{
			name:   "off by one additional child",
			tree:   MustParseTreeFromDiagram("[E (*) [F (id)]]"),
			t:      MustParseTreeFromDiagram("[E (*) [F (id)] (-)]"),
			expect: false,
		},
		{
			name:   "zero tree is not a subtree of terminal node",
			tree:   &Tree{},
			t:      &Tree{Value: "root", Terminal: true},
			expect: false,
		},
		{
			name:   "zero tree is not a subtree of epsilon terminal node",
			tree:   &Tree{},
			t:      &Tree{Value: "", Terminal: true},
			expect: false,
		},
		{
			name:   "zero tree is not a subtree of non-terminal node with Value set",
			tree:   &Tree{},
			t:      &Tree{Value: "S", Terminal: false},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			actualContains, actualPath := tc.tree.IsSubTreeOf(*tc.t)

			assert.Equal(tc.expect, actualContains)

			if tc.expect {
				// only check path if we expect the result to have been true,
				// otherwise the returned path should not be used.
				assert.Equal(tc.expectPath, actualPath)
			}
		})
	}
}
