package trans

import (
	"fmt"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/rosed"
)

// Validate runs the SDTS on a fake parse tree derived from the grammar. The
// given attribute will be attempted to be evaluated on the root node.
//
// It will use fake value producer, if provided, to generate lexemes for
// terminals in the tree; otherwise contrived values will be used.
func Validate(sdts SDTS, g grammar.Grammar, attribute string, debug ValidationOptions, fakeValProducer ...map[string]func() string) (warns []string, err error) {
	pts, err := parse.DeriveFullTree(g, fakeValProducer...)
	if err != nil {
		return warns, fmt.Errorf("deriving fake parse tree: %w", err)
	}

	const errIndentStr = "    "

	treeErrs := []box.Pair[error, *parse.Tree]{}

	evalErrToTreeError := func(errFromEval error) error {
		evalErr, ok := errFromEval.(evalError)
		if !ok {
			if errFromEval != nil {
				return errFromEval
			}
		}

		if len(evalErr.depGraphs) > 0 {
			// disconnected depgraph error

			fullMsg := "translation on parse tree resulted in disconnected dependency graphs:"

			for i := range evalErr.unexpectedBreaks {
				br := evalErr.unexpectedBreaks[i]
				fullMsg += fmt.Sprintf("\n* at least one %s.%q in production of (%s -> %s) is unused", br[2], br[3], br[0], br[1])
			}

			if debug.FullDepGraphs {
				fullMsg += "\nDepGraphs:"
				for i := range evalErr.depGraphs {
					dgStr := depGraphString(evalErr.depGraphs[i])
					dgStr = rosed.Edit(dgStr).
						LinesFrom(1).
						IndentOpts(1, rosed.Options{IndentStr: errIndentStr}).
						String()
					fullMsg += fmt.Sprintf("\n"+errIndentStr+"* DepGraph #%d: %s", i+1, dgStr)
				}
			}

			return fmt.Errorf(fullMsg)
		}

		return errFromEval
	}

	for i := range pts {
		var evalWarns []error
		var treeErr error
		_, evalWarns, err = sdts.Evaluate(pts[i], attribute)

		localPT := pts[i]

		if err != nil {
			treeErr = evalErrToTreeError(err)
		}

		for _, ew := range evalWarns {
			ewAsTreeErr := evalErrToTreeError(ew)
			if debug.ParseTrees {
				treeStr := Annotate(localPT).String()
				warns = append(warns, fmt.Sprintf("Failed Tree %d:\n%s\nParse Tree:\n%s", i+1, ewAsTreeErr.Error(), treeStr))
			} else {
				warns = append(warns, fmt.Sprintf("Failed Tree %d: %s", i+1, ewAsTreeErr.Error()))
			}
		}

		if treeErr != nil {
			treeErrs = append(treeErrs, box.PairOf(treeErr, &localPT))
		}
	}

	var finalErr error

	if len(treeErrs) > 0 {
		var treeCountStr string
		var treeCountS string
		var errCountStr string
		var errCountS string
		if len(pts) != 1 {
			treeCountStr = fmt.Sprintf("%d ", len(pts))
			treeCountS = "s"
		}
		if len(treeErrs) != 1 {
			errCountStr = fmt.Sprintf("%d ", len(treeErrs))
			errCountS = "s"
		}

		var fullErrStr string

		for i := range treeErrs {
			if i < debug.SkipErrors {
				continue
			}
			if debug.ParseTrees {
				treeStr := Annotate(*treeErrs[i].Second).String()
				fullErrStr += fmt.Sprintf("\n\nFailed Tree %d:\n%s\nParse Tree:\n%s", i+1, treeErrs[i].First.Error(), treeStr)
			} else {
				fullErrStr += fmt.Sprintf("\n\nFailed Tree %d: %s", i+1, treeErrs[i].First.Error())
			}

			if !debug.ShowAllErrors {
				// count up errors after this one
				otherErrsCount := len(treeErrs) - 1 - i

				// ... and those before this one that were skipped
				otherErrsCount += debug.SkipErrors

				if otherErrsCount > 0 {
					plural := ""
					if otherErrsCount != 1 {
						plural = "s"
					}
					fullErrStr += fmt.Sprintf("\n\n... (and %d more error%s suppressed by options)", otherErrsCount, plural)
				}
				break
			}
		}

		if fullErrStr == "" {
			fullErrStr = "\n... (all error output suppressed by options)"
		}

		fullErrStr = fmt.Sprintf("Running SDTS on %ssimulated parse tree%s got %serror%s:", treeCountStr, treeCountS, errCountStr, errCountS) + fullErrStr

		finalErr = fmt.Errorf("%s", fullErrStr)
	}

	return warns, finalErr
}

// highly populated error struct for examination by validation code and internal
// routines. may make this betta and exported later.
type evalError struct {
	// if this is a disconnected dep graph segments error, this slice will be
	// non-nil and contain the issue nodes.
	depGraphs []*directedGraph[depNode]

	// if this is a disconnected dep graph segments error, this slice will be
	// non-nil and contain the important features of each break. Each element is
	// a string triple containing: the symbol of the parent of the node that
	// caused the break, the production the parent node was made from as a
	// string, the symbol of the node that caused the break, and the name of the
	// attribute that caused the break.
	unexpectedBreaks [][4]string

	// if this is a sort error, this will be true
	sortError bool

	// if this was created due to a missing hook function, this will be set to
	// the name of the hook it is looking for.
	missingHook string

	// if this was created due to a failed hook function, this will be set to
	// the name of the hook it was trying to run.
	failedHook string

	msg string
}

// Error returns the error message.
func (ee evalError) Error() string {
	return ee.msg
}
