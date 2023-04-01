package trans

// ValidationOptions is a struct passed to SDTS.Validate that selects the items
// to be included in the text of returned errors when problems are encountered.
type ValidationOptions struct {
	// ParseTrees is whether to include parse trees in error output.
	ParseTrees bool

	// FullDepGraphs is whether to include full dependency graphs in error
	// output.
	FullDepGraphs bool

	// ShowAllErrors is whether to show all validation errors in error output.
	// If false, only the first error (after any skipped) will be shown.
	ShowAllErrors bool

	// SkipErrors is the number of initial errors to skip in error output, if
	// any.
	SkipErrors int
}
