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

	-p/-parser FILE
		Set the location of the pre-compiled parser cache to the given CFF
		format file as opposed to the default of './parser.cff'.

	-no-cache
		Disable the loading of any cached frontend components, even if a
		pre-built one is available.

	-no-cache-out
		Disable writing of any frontend components cache, even if a component
		was built by the invocation.

	-version
		Print the version of the ictiobus compiler-compiler and exit.

	-val-sdts-off
		Disable validatione of the SDTS of the resulting fishi.

	-val-sdts-trees
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show the parse tree(s) that caused the problem.

		Has no effect if -val-sdts-off is set.

	-val-sdts-graphs
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show the full resulting dependency graph(s) that caused
		the issue (if any).

		Has no effect if -val-sdts-off is set.

	-val-sdts-first
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show only the problem(s) found in the first simulated
		parse tree (after any skipped by -val-sdts-skip) and then stop.

		Has no effect if -val-sdts-off is set.

	-val-sdts-skip N
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, skip the first N simulated parse trees in the output.
		Combine with -val-sdts-first to view a specific parse tree.

		Has no effect if -val-sdts-off is set.

	-debug-lexer
		Enable debug mode for the lexer and print each token to standard out as
		it is lexed. Note that if the lexer is not in lazy mode, all tokens will
		be lexed before any parsing begins, and so with debug-lexer enabled will
		all be printed to stdout before any parsing begins.

	-debug-parser
		Enable debug mode for the parser and print each step of the parse to
		stdout, including the symbol stack, manipulations of the stack, ACTION
		selected in DFA based on the stack, and other information.

Each markdown file given is scanned for fishi codeblocks. They are all combined
into a single fishi code block and parsed. Each markdown file is parsed
separately but their resulting ASTs are combined into a single list of FISHI
specs for a language.

If there are any errors, they are displayed and the program exits with a
non-zero exit code. If there are multiple files, they are all attempted to be
parsed, even if a prior one failed, so that as many errors as possible are
shown at once. Note that when multiple files are given, each problem may end up
setting the exit code separately, so if any interpretation of the exit code is
done besides checking for non-zero, it should be noted that it will only be the
correct exit code for the last file parsed.

If files containing cached pre-built components of the frontend are available,
they will be loaded and used unless -no-cache is set. The files are named
'fishi-parser.cff' by default, and the names can be changed with the -parser/-p
flag if desired. Cache invalidation is non-sophisticated and cannot be
automatically detected at this time. To force it to occur, the -no-cache flag
must be manually used (or the file deleted).

If new frontend components are generated from scratch, they will be cached by
saving them to the files mentioned above unless -no-cache-out is set. Note that
if the frontend components are loaded from cache files, they will not be output
to cache files again regardless of whether -no-cache-out is present.

Once the input has been successfully parsed, the parser is generated using the
options provided, unless the -n flag is set, in which case ictcc will
immediately exit with a success code after parsing the input.
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
	noGen         bool
	genAST        bool
	genTree       bool
	parserCff     string
	noCache       *bool = flag.Bool("no-cache", false, "Disable use of cached frontend components, even if available")
	noCacheOutput *bool = flag.Bool("no-cache-out", false, "Disable writing of cached frontend components, even if one was generated")

	valSDTSOff        *bool = flag.Bool("val-sdts-off", false, "Disable validation of the SDTS of the resulting fishi")
	valSDTSShowTrees  *bool = flag.Bool("val-sdts-trees", false, "Show trees that caused SDTS validation errors")
	valSDTSShowGraphs *bool = flag.Bool("val-sdts-graphs", false, "Show full generated dependency graph output for parse trees that caused SDTS validation errors")
	valSDTSFirstOnly  *bool = flag.Bool("val-sdts-first", false, "Show only the first error found in SDTS validation")
	valSDTSSkip       *int  = flag.Int("val-sdts-skip", 0, "Skip the first N errors found in SDTS validation in output")

	lexerTrace  *bool = flag.Bool("debug-lexer", false, "Print the lexer trace to stdout")
	parserTrace *bool = flag.Bool("debug-parser", false, "Print the parser trace to stdout")

	version *bool = flag.Bool("version", false, "Print the version of ictcc and exit")
)

func init() {
	const (
		noGenUsage       = "Do not generate the parser"
		genASTUsage      = "Print the AST of the analyzed fishi"
		genTreeUsage     = "Print the parse tree of the analyzed fishi"
		parserCffUsage   = "Use the specified parser CFF cache file instead of default"
		parserCffDefault = "fishi-parser.cff"
	)
	flag.BoolVar(&noGen, "no-gen", false, noGenUsage)
	flag.BoolVar(&noGen, "n", false, noGenUsage+" (shorthand)")
	flag.BoolVar(&genAST, "ast", false, genASTUsage)
	flag.BoolVar(&genAST, "a", false, genASTUsage+" (shorthand)")
	flag.BoolVar(&genTree, "tree", false, genTreeUsage)
	flag.BoolVar(&genTree, "t", false, genTreeUsage+" (shorthand)")
	flag.StringVar(&parserCff, "parser", parserCffDefault, parserCffUsage)
	flag.StringVar(&parserCff, "p", parserCffDefault, parserCffUsage+" (shorthand)")
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

	if *version {
		fmt.Println(GetVersionString())
		return
	}

	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "No files given to process")
		returnCode = ExitErrNoFiles
		return
	}

	fo := fishi.Options{
		ParserCFF:         parserCff,
		ReadCache:         !*noCache,
		WriteCache:        !*noCacheOutput,
		SDTSValidate:      !*valSDTSOff,
		SDTSValShowTrees:  *valSDTSShowTrees,
		SDTSValShowGraphs: *valSDTSShowGraphs,
		SDTSValAllTrees:   !*valSDTSFirstOnly,
		SDTSValSkipTrees:  *valSDTSSkip,
		LexerTrace:        *lexerTrace,
		ParserTrace:       *parserTrace,
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
