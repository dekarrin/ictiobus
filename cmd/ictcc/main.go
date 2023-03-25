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

	-cache-off
		Disable use of any cached frontend components, even if available.

	-val-sdts-off
		Disable validatione of the SDTS of the resulting fishi.

	-val-sdts-trees
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show the parse tree(s) that caused the problem.

	-val-sdts-graphs
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show the full resulting dependency graph(s) that caused
		the issue (if any).

	-debug-lexer
		Enable debug mode for the lexer and print each token to standard out as
		it is lexed. Note that if the lexer is not in lazy mode, all tokens will
		be lexed before any parsing begins, and so with debug-lexer enabled will
		all be printed to stdout before any parsing begins.

	-debug-parser
		Enable debug mode for the parser and print each step of the parse to
		stdout, including the symbol stack, manipulations of the stack, ACTION
		selected in DFA based on the stack, and other information.
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
	noGen             bool
	genAST            bool
	genTree           bool
	noCache           *bool = flag.Bool("cache-off", false, "Disable use of cached frontend components, even if available")
	noValidateSDTS    *bool = flag.Bool("val-sdts-off", false, "Disable validation of the SDTS of the resulting fishi")
	showSDTSValTrees  *bool = flag.Bool("val-sdts-trees", false, "Show trees that caused SDTS validation errors")
	showSDTSValGraphs *bool = flag.Bool("val-sdts-graphs", false, "Show full generated dependency graph output for parse trees that caused SDTS validation errors")
	lexerTrace        *bool = flag.Bool("debug-lexer", false, "Print the lexer trace to stdout")
	parserTrace       *bool = flag.Bool("debug-parser", false, "Print the parser trace to stdout")
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

	fo := fishi.Options{
		ParserCFF:      "fishi-parser.cff",
		UseCache:       !*noCache,
		ValidateSDTS:   !*noValidateSDTS,
		ShowSDTSTrees:  *showSDTSValTrees,
		ShowSDTSGraphs: *showSDTSValGraphs,
		LexerTrace:     *lexerTrace,
		ParserTrace:    *parserTrace,
	}

	for _, file := range args {

		res, err := fishi.ExecuteMarkdownFile(file, fo)

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
