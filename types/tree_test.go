package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Help(t *testing.T) {
	st := MustParseTreeFromDiagram("[S (+) (+)]")

	fmt.Println(st.String())

	assert.True(t, false)
}

func Test_ParseTree_IsSubTreeOf(t *testing.T) {

	testsCases := []struct {
		name       string
		tree       *ParseTree
		t          *ParseTree
		expect     bool
		expectPath []int
	}{
		{
			name:       "zero tree is a subtree of zero tree",
			tree:       &ParseTree{},
			t:          &ParseTree{},
			expect:     true,
			expectPath: []int{},
		},
		{
			name: "zero tree is not a subtree of empty node with children",
			tree: &ParseTree{},
			t: &ParseTree{Children: []*ParseTree{
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
			tree:   &ParseTree{},
			t:      &ParseTree{Value: "root", Terminal: true},
			expect: false,
		},
		{
			name:   "zero tree is not a subtree of epsilon terminal node",
			tree:   &ParseTree{},
			t:      &ParseTree{Value: "", Terminal: true},
			expect: false,
		},
		{
			name:   "zero tree is not a subtree of non-terminal node with Value set",
			tree:   &ParseTree{},
			t:      &ParseTree{Value: "S", Terminal: false},
			expect: false,
		},
	}

	for _, tc := range testsCases {
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
