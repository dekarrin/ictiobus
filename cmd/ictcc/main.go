/*
Ictcc runs the ictiobus compiler-compiler on one or more markdown files that
contain 'fishi' codeblocks.

It parses all files given as arguments and outputs a generated frontend for the
language specified by the fishi codeblocks.

Usage:

	ictcc [flags] file1.md file2.md ...

Flags:

	-a/-ast
		Print the AST to stdout before generating the parser.

	-t/-tree
		Print the parse tree to stdout before generating the parser.

	-n/-no-gen
		Do not generate the parser. If this flag is set, the fishi is parsed and
		checked for errors but no other action is taken (unless specified by
		other flags).
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dekarrin/ictiobus/fishi"
	"github.com/dekarrin/ictiobus/translation"
	"github.com/dekarrin/ictiobus/types"
)

const (
	// ExitSuccess is the exit code for a successful run.
	ExitSuccess = iota

	// ExitErrNoFiles is the code returned as exit status when no files are
	// provided to the invocation.
	ExitErrNoFiles

	// ExitErrSyntax is the code returned as exit status when a syntax error
	// occurs.
	ExitErrSyntax

	// ExitErrOther is a generic error code for any other error.
	ExitErrOther
)

var (
	returnCode = ExitSuccess
)

var (
	noGen   bool
	genAST  bool
	genTree bool
)

func init() {
	const (
		noGenUsage   = "Do not generate the parser"
		genASTUsage  = "Print the AST of the analyzed fishi"
		genTreeUsage = "Print the parse tree of the analyzed fishi"
	)
	flag.BoolVar(&noGen, "no-gen", false, noGenUsage)
	flag.BoolVar(&noGen, "n", false, noGenUsage)
	flag.BoolVar(&genAST, "ast", false, genASTUsage)
	flag.BoolVar(&genAST, "a", false, genASTUsage)
	flag.BoolVar(&genTree, "tree", false, genTreeUsage)
	flag.BoolVar(&genTree, "t", false, genTreeUsage)
}

func main() {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// we are panicking, make sure we dont lose the panic just because
			// we checked
			panic("unrecoverable panic occured")
		} else {
			os.Exit(returnCode)
		}
	}()

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "No files given to process")
		returnCode = ExitErrNoFiles
		return
	}

	for _, file := range args {

		res, err := fishi.ExecuteMarkdownFile(file)

		// results may be valid even if there is an error
		if res.AST != nil && genAST {
			fmt.Printf("%s\n", res.AST.String())
		}

		if res.Tree != nil && genTree {
			fmt.Printf("%s\n", translation.AddAttributes(*res.Tree).String())
		}

		if err != nil {
			if syntaxErr, ok := err.(*types.SyntaxError); ok {
				fmt.Fprintf(os.Stderr, "%s:\n%s\n", file, syntaxErr.FullMessage())
				returnCode = ExitErrSyntax
			} else {
				fmt.Fprintf(os.Stderr, "%s: %s\n", file, err.Error())
				returnCode = ExitErrOther
			}
			return
		}

		if !noGen {
			// do processing of the AST here
			time.Sleep(100 * time.Millisecond) // so they don't interleave
			fmt.Printf("(frontend generation not implemented yet)\n")
		}
	}

}
