package types

// DebugInfo is a struct passed to certain functions that selects the items to
// be included in error output for debugging purposes while building a compiler.
type DebugInfo struct {
	// ParseTrees is whether to include parse trees in error output.
	ParseTrees bool

	// FullDepGraphs is whether to include full dependency graphs in error
	// output.
	FullDepGraphs bool
}
