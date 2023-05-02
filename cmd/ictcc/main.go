/*
Ictcc runs the ictiobus compiler-compiler on one or more markdown files that
contain 'fishi' codeblocks.

It parses all files given as arguments and outputs a generated frontend for the
language specified by the fishi codeblocks.

Usage:

	ictcc [flags] file1.md file2.md ...

Flags:

	-a/--ast
		Print the AST to stdout before generating the frontend.

	-t/--tree
		Print the parse tree to stdout before generating the frontend.

	-s/--spec
		Print the interpreted language specification to stdout before generating
		the frontend.

	-n/--no-gen
		Do not generate the parser. If this flag is set, the fishi is parsed and
		checked for errors but no other action is taken (unless specified by
		other flags).

	-d/--diag FILE
		Generate a diagnostics binary for the target language. Assuming there
		are no issues with the FISHI spec, this will generate a binary that can
		analyze files written in the target language and output the result of
		frontend analysis. This can be useful for testing out the frontend on
		files quickly and efficiently, as it also includes further options
		useful for debugging purposes, such as debugging lexed tokens and the
		parser itself. Note that by default the diagnostics binary will only
		accept text input in the language accepted by the specified frontend; to
		allow it to perform reading of specialized formats and/or perform
		preprocessing, use the -f/--diag-format-pkg flag.

	-q/--quiet
		Do not show progress messages. This does not affect error messages or
		warning output.

	-f/--diag-format-pkg DIR
		Enable special format reading in the generated diagnostics binary by
		wrapping any io.Reader opened on input files in another io.Reader that
		handles reading the format of the input file. This is performed by
		calling a function in the package located in the specified file, by
		default this function is called 'NewCodeReader' but can be changed by
		specifying the --diag-format-call flag. The function must take an
		io.Reader and return a new io.Reader that reads source code from the
		given io.Reader and performs any preprocessing required on it. This
		allows the diagnostics binary to read files that are not simply text
		files directly ready to be accepted by the frontend. If not set, the
		diagnostics binary will not perform any preprocessing on input files and
		assumes that any input can be directly accepted by the frontend. This
		flag is only useful if -d/--diag is also set.

	-c/--diag-format-call NAME
		Set the name of the function to call in the package specified by
		-f/--diag-format-pkg to get an io.Reader that can read specialized
		formats. Defaults to 'NewCodeReader'. This function is used by the
		diagnostics binary to do format reading and preprocessing on input prior
		to analysis by the frontend. This flag is only useful if -d/--diag is
		also set.

	-l/--lang NAME
		Set the name of the language to generate a frontend for. Defaults to
		"Unspecified".

	-v/--lang-ver VERSION
		Set the version of the language to generate a frontend for. Defaults to
		"v0.0.0".

	-p/--preproc
		Show the output of running preprocessing on input files. This will show
		the exact code that is going to be parsed by ictcc before such parsing
		occurs.

	-F/--fatal WARN_TYPE
		Make warnings of the given type be fatal. If ictcc encounters a warning
		of that type, it will treat it as an error and immediately fail. The
		possible values for the type of warning is as follows: "dupe_human",
		"missing_human", "priority", "unused", "ambig", "validation", "import",
		"val_args", or "all" to make all errors fatal. This option can be passed
		more than once to give multiple warning types. See manual for
		description of when each type of warning could arise. If a warning is
		specified as both fatalized and suppressed by options, treating it as
		fatal takes precedence. Specifying both '-F all' and '-S all' is not
		allowed.

	-S/--suppress WARN_TYPE
		Suppress outout of warnings of the given type. No "WARN" message will be
		printed for that type even if ictcc encounters it. The possible values
		for the type of warning to suppress is as as follows: "dupe_human",
		"missing_human", "priority", "unused", "ambig", "validation", "import",
		"val_args", or "all" to make all errors fatal. This option can be passed
		more than once to give multiple warning types. See manual for
		description of when each type of warning could arise. If a warning is
		specified as both fatalized and suppressed by options, treating it as
		fatal takes precedence. Specifying both '-F all' and '-S all' is not
		allowed.

	--preserve-bin-source
		Do not delete source files for any generated binary after compiling the
		binary.

	--version
		Print the version of the ictiobus compiler-compiler and exit.

	--sim-off
		Disable simulation of the language once built. This will disable SDTS
		validation, as live simulation is the only way to do this due to the
		lack of support for dynamic loading of the hooks package in Go.

	--sim-trees
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show the parse tree(s) that caused the problem.

		Has no effect if -val-sdts-off is set.

	--sim-graphs
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show the full resulting dependency graph(s) that caused
		the issue (if any).

		Has no effect if -val-sdts-off is set.

	--sim-first-err
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, show only the problem(s) found in the first simulated
		parse tree (after any skipped by -val-sdts-skip) and then stop.

		Has no effect if -val-sdts-off is set.

	--sim-skip-errs N
		If problems are detected with the SDTS of the resulting fishi during
		SDTS validation, skip the first N simulated parse trees in the output.
		Combine with -val-sdts-first to view a specific parse tree.

		Has no effect if -val-sdts-off is set.

	--debug-lexer
		Enable debug mode for the lexer and print each token to standard out as
		it is lexed. Note that if the lexer is not in lazy mode, all tokens will
		be lexed before any parsing begins, and so with debug-lexer enabled will
		all be printed to stdout before any parsing begins.

	--debug-parser
		Enable debug mode for the parser and print each step of the parse to
		stdout, including the symbol stack, manipulations of the stack, ACTION
		selected in DFA based on the stack, and other information.

	--pkg NAME
		Set the name of the package to place generated files in. Defaults to
		'fe'.

	--dest DIR
		Set the destination directory to place generated files in. Defaults to a
		directory named 'fe' in the current working directory.

	--prefix PATH
		Set the prefix to use for all generated source files. Defaults to the
		current working directory. If used, generated source files will be be
		output to their location with this prefix instead of in a directory
		(".sim", ".gen", and the generated frontend source package folder)
		located in the current working directory. Combine with
		--preserve-bin-source to aid in debugging. Does not affect diagnostic
		binary output location.

	--debug-templates
		Enable dumping of the fishi filled template files before they are passed
		to the formatter. This allows debugging of the template files when
		editing them, since they must be valid go code to be formatted.

	--tmpl-tokens FILE
		Use the provided file as the template for outputting the generated
		tokens file instead of the default embedded within the binary.

	--tmpl-lexer FILE
		Use the provided file as the template for outputting the generated lexer
		file instead of the default embedded within the binary.

	--tmpl-parser FILE
		Use the provided file as the template for outputting the generated
		parser file instead of the default embedded within the binary.

	--tmpl-sdts FILE
		Use the provided file as the template for outputting the generated SDTS
		file instead of the default embedded within the binary.

	--tmpl-main FILE
		Use the provided file as the template for outputting generated binary
		main file instead of the default embedded within the binary.

	--tmpl-frontend FILE
		Use the provided file as the template for outputting the generated
		frontend file instead of the default embedded within the binary.

	--no-ambig
		Disable the generation of a parser for an ambiguous language. Normally,
		when generating an LR parser, an ambiguous grammar is allowed, with
		shift-reduce conflicts resolved in favor of shift in all cases. If this
		flag is set, that behavior is disabled and an error is returned if the
		given grammar is ambiguous. Note that LL(k) grammars can never be
		ambiguous, so this flag has no effect if explicitly selecting an LL
		parser.

	--ll
		Force the generation of an LL(1) parser instead of the default of trying
		each parser type in sequence from most restrictive to least restrictive
		and using the first one found.

	--slr
		Force the generation of an SLR(1) (simple LR) parser instead of the
		default of trying each parser type in sequence from most restrictive to
		least restrictive and using the first one found.

	--clr
		Force the generation of a CLR(1) (canonical LR) parser instead of the
		default of trying each parser type in sequence from most restrictive to
		least restrictive and using the first one found.

	--lalr
		Force the generation of a LALR(1) (lookahead LR) parser instead of the
		default of trying each parser type in sequence from most restrictive to
		least restrictive and using the first one found.

	--hooks PATH
		Gives the filesystem path to the directory containing the package that
		the hooks table is in. This is required for live validation of simulated
		parse trees, but may be omitted if SDTS validation is disabled.

	--hooks-table NAME
		Gives the expression to retrieve the hooks table from the hooks package,
		relative to the package that it is in. The NAME must be an exported var
		of type map[string]trans.AttributeSetter, or a function call that
		returns such a value. Do not include the package prefix as part of this
		expression; it will be automatically determined. The default value is
		"HooksTable".

	--ir TYPE
		Gives the type of the intermediate representation returned by applying
		the translation scheme to a parse tree. This is required for running
		SDTS validation on simulated parse trees, but may be omitted if SDTS
		validation is not enabled. Regardless, if set, the generated frontend
		file will explicitly return ictiobus.Frontend[TYPE] instead of requiring
		it to be declared at calltime of Frontend(). TYPE must be qualified with
		the fully-qualified package name; e.g.
		"github.com/dekarrin/ictiobus/fishi/syntax.Node". Pointer indirection
		and array/slice notation are allowed; maps are not (but types that have
		map as an underlying type are allowed).

	--dev
		Enable development mode. This will cause generated binaries to use the
		local version of ictiobus instead of the latest release. If this flag is
		enabled, it is assumed that the ictiobus package is located in the
		current working directory.

Each markdown file given is scanned for fishi codeblocks. They are all combined
into a single fishi code block and parsed. Each markdown file is parsed
separately but their resulting ASTs are combined into a single FISHI spec for a
language.

The spec is then used to generate a lexer, parser, and SDTS for the language.
For the parser, if no specific parser is selected via the --ll, --slr, --clr, or
--lalr flags, the parser generator will attempt to generate each type of parser
in sequence from most restrictive to least restrictive (LL(1), simple LR(1),
lookahead LR(1), and canonical LR(1), in that order), and use the first one it
is able to generate. If a specific parser is selected, it will attempt to
generate that one and fail if it is unable to.

If there are any errors, they are displayed and the program exits with a
non-zero exit code. If there are multiple files, they are all attempted to be
parsed, even if a prior one failed, so that as many errors as possible are
shown at once. Note that when multiple files are given, each problem may end up
setting the exit code separately, so if any interpretation of the exit code is
done besides checking for non-zero, it should be noted that it will only be the
correct exit code for the last file parsed.

If files containing cached pre-built components of the frontend are available,
they will be loaded and used unless -no-cache is set. The files are named
'fishi-parser.cff' by default, and the names can be changed with the --parser/-p
flag if desired. Cache invalidation is non-sophisticated and cannot be
automatically detected at this time. To force it to occur, the -no-cache flag
must be manually used (or the file deleted).

If new frontend components are generated from scratch, they will be cached by
saving them to the files mentioned above unless --no-cache-out is set. Note that
if the frontend components are loaded from cache files, they will not be output
to cache files again regardless of whether --no-cache-out is present.

Once the input has been successfully parsed, the parser is generated using the
options provided, unless the -n flag is set, in which case ictcc will
immediately exit with a success code after parsing the input.
*/
package main

import (
	"bufio"
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/spf13/pflag"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/fishi"
	"github.com/dekarrin/ictiobus/fishi/format"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/ictiobus/types"
)

var (
	flagWarnFatal    = pflag.StringArrayP("fatal", "F", nil, "Treat given warning as a fatal error")
	flagWarnSuppress = pflag.StringArrayP("suppress", "S", nil, "Suppress output of given warning")

	flagQuietMode = pflag.BoolP("quiet", "q", false, "Suppress progress messages and other supplementary output")
	flagNoGen     = pflag.BoolP("no-gen", "n", false, "Do not output generated frontend output files. Skips parser generation if --diag is not set.")
	flagGenAST    = pflag.BoolP("ast", "a", false, "Print the AST of the analyzed fishi")
	flagGenTree   = pflag.BoolP("tree", "t", false, "Print the parse trees of each analyzed fishi file")
	flagShowSpec  = pflag.BoolP("spec", "s", false, "Print the FISHI spec interpreted from the analyzed fishi")
	flagLang      = pflag.StringP("lang", "l", "Unspecified", "The name of the languae being generated")
	flagLangVer   = pflag.StringP("lang-ver", "v", "v0.0.0", "The version of the language to generate")
	flagPreproc   = pflag.BoolP("preproc", "p", false, "Print the preprocessed FISHI code before compiling it")

	flagDiagBin        = pflag.StringP("diag", "d", "", "Generate binary that has the generated frontend and uses it to analyze the target language")
	flagDiagFormatPkg  = pflag.StringP("diag-format-pkg", "f", "", "The package containing format functions for the diagnostic binary to call on input prior to passing to frontend analysis")
	flagDiagFormatCall = pflag.StringP("diag-format-call", "c", "NewCodeReader", "The function within the diag-format-pkg to call to open a reader on input prior to passing to frontend analysis")

	flagPathPrefix        = pflag.String("prefix", "", "Path to prepend to path of all generated source files")
	flagPreserveBinSource = pflag.Bool("preserve-bin-source", false, "Preserve the source of any generated binary files")
	flagDebugTemplates    = pflag.Bool("debug-templates", false, "Dump the filled templates before running through gofmt")
	flagPkg               = pflag.String("pkg", "fe", "The name of the package to place generated files in")
	flagDest              = pflag.String("dest", "./fe", "The name of the directory to place the generated package in")
	flagNoCache           = pflag.Bool("no-cache", false, "(UNUSED) Disable use of cached frontend components, even if available")
	flagNoCacheOutput     = pflag.Bool("no-cache-out", false, "(UNUSED) Disable writing of cached frontend components, even if one was generated")
	flagParserCff         = pflag.String("parser", "fishi-parser.cff", "(UNUSED) Use the specified parser CFF cache file instead of default")

	flagSimOff          = pflag.Bool("sim-off", false, "Disable input simulation of the language once built")
	flagSimTrees        = pflag.Bool("sim-trees", false, "Show parse trees that caused errors during simulation")
	flagSimGraphs       = pflag.Bool("sim-graphs", false, "Show full generated dependency graph output for parse trees that caused errors during simulation")
	flagSimFirstErrOnly = pflag.Bool("sim-first-err", false, "Show only the first error found in SDTS validation")
	flagSimSkipErrs     = pflag.Int("sim-skip-errs", 0, "Skip the first N errors found in SDTS validation in output")

	flagTmplTokens = pflag.String("tmpl-tokens", "", "A template file to replace the embedded tokens template with")
	flagTmplLexer  = pflag.String("tmpl-lexer", "", "A template file to replace the embedded lexer template with")
	flagTmplParser = pflag.String("tmpl-parser", "", "A template file to replace the embedded parser template with")
	flagTmplSDTS   = pflag.String("tmpl-sdts", "", "A template file to replace the embedded SDTS template with")
	flagTmplFront  = pflag.String("tmpl-frontend", "", "A template file to replace the embedded frontend template with")
	flagTmplMain   = pflag.String("tmpl-main", "", "A template file to replace the embedded main.go template with")

	flagParserLL      = pflag.Bool("ll", false, "Generate an LL(1) parser")
	flagParserSLR     = pflag.Bool("slr", false, "Generate a simple LR(1) parser")
	flagParserCLR     = pflag.Bool("clr", false, "Generate a canonical LR(1) parser")
	flagParserLALR    = pflag.Bool("lalr", false, "Generate a canonical LR(1) parser")
	flagParserNoAmbig = pflag.Bool("no-ambig", false, "Disallow ambiguity in grammar even if creating a parser that can auto-resolve it")

	flagLexerTrace  = pflag.Bool("debug-lexer", false, "Print the lexer trace to stdout")
	flagParserTrace = pflag.Bool("debug-parser", false, "Print the parser trace to stdout")

	flagHooksPath      = pflag.String("hooks", "", "The path to the hooks directory to use for the generated parser. Required for SDTS validation.")
	flagHooksTableName = pflag.String("hooks-table", "HooksTable", "Function call or name of exported var in 'hooks' that has the hooks table.")

	flagIRType = pflag.String("ir", "", "The fully-qualified type of IR to generate.")

	flagVersion = pflag.Bool("version", false, "Print the version of ictcc and exit")

	flagDev = pflag.Bool("dev", false, "Enable development mode so generated binaries use the version of ictiobus in the current working dir")
)

// DevModeInfo is info on dev mode that is gathered by reading CLI flags.
type DevModeInfo struct {
	// Enabled is whether develepment mode is enabled.
	Enabled bool

	// LocalIctiobusSource is the path to the local ictiobus source code. By
	// default this is taken from the current working directory.
	LocalIctiobusSource string
}

func main() {
	defer preservePanicOrExitWithStatus()

	// gather options and arguments
	invocation := strings.Join(os.Args[1:], " ")

	pflag.Parse()

	if *flagVersion {
		fmt.Println(GetVersionString())
		return
	}

	warnHandler, err := fishi.NewWarnHandlerFromCLI(*flagWarnSuppress, *flagWarnFatal)
	if err != nil {
		errInvalidFlags(err.Error())
		return
	}

	// mutually exclusive and required options for diagnostics bin generation.
	if *flagDiagBin != "" {
		if *flagIRType == "" || *flagHooksPath == "" {
			errInvalidFlags("Diagnostics bin generation requires both --ir and --hooks to be set")
			return
		}

		// you cannot set ONLY the formatting call
		flagInfoDiagFormatCall := pflag.Lookup("diag-format-call")
		// don't error check; all we'd do is panic
		if flagInfoDiagFormatCall.Changed && *flagDiagFormatPkg == "" {
			errInvalidFlags("-c/--diag-format-call cannot be set without -f/--diag-format-pkg")
			return
		}
	} else {
		// otherwise, it makes no sense to set --diag-format-pkg or --diag-format-call; disallow this
		flagInfoDiagFormatPkg := pflag.Lookup("diag-format-pkg")
		flagInfoDiagFormatCall := pflag.Lookup("diag-format-call")
		if flagInfoDiagFormatPkg.Changed {
			errInvalidFlags("-f/--diag-format-pkg cannot be set without -d/--diagnostics-bin")
			return
		}
		if flagInfoDiagFormatCall.Changed {
			errInvalidFlags("-c/--diag-format-call cannot be set without -d/--diagnostics-bin")
			return
		}
	}

	parserType, allowAmbig, err := parserSelectionFromFlags()
	if err != nil {
		errInvalidFlags(err.Error())
		return
	}

	devInfo, err := devModeInfoFromFlags()
	if err != nil {
		errInvalidFlags("--dev: " + err.Error())
		return
	}

	// check args before gathering flags
	args := pflag.Args()

	if len(args) < 1 {
		errNoFiles("No files given to process")
		return
	}

	// create a spec metadata object
	md := fishi.SpecMetadata{
		Language:       *flagLang,
		Version:        *flagLangVer,
		InvocationArgs: invocation,
	}

	fo := fishi.Options{
		ParserCFF:   *flagParserCff,
		ReadCache:   !*flagNoCache,
		WriteCache:  !*flagNoCacheOutput,
		LexerTrace:  *flagLexerTrace,
		ParserTrace: *flagParserTrace,
	}

	cgOpts := fishi.CodegenOptions{
		DumpPreFormat:        *flagDebugTemplates,
		IRType:               *flagIRType,
		TemplateFiles:        map[string]string{},
		PreserveBinarySource: *flagPreserveBinSource,
	}
	if *flagTmplTokens != "" {
		cgOpts.TemplateFiles[fishi.ComponentTokens] = *flagTmplTokens
	}
	if *flagTmplLexer != "" {
		cgOpts.TemplateFiles[fishi.ComponentLexer] = *flagTmplLexer
	}
	if *flagTmplParser != "" {
		cgOpts.TemplateFiles[fishi.ComponentParser] = *flagTmplParser
	}
	if *flagTmplSDTS != "" {
		cgOpts.TemplateFiles[fishi.ComponentSDTS] = *flagTmplSDTS
	}
	if *flagTmplFront != "" {
		cgOpts.TemplateFiles[fishi.ComponentFrontend] = *flagTmplFront
	}
	if *flagTmplMain != "" {
		cgOpts.TemplateFiles[fishi.ComponentMainFile] = *flagTmplMain
	}
	if len(cgOpts.TemplateFiles) == 0 {
		// just nil it
		cgOpts.TemplateFiles = nil
	}

	// now that args are gathered, parse markdown files into an AST
	if !*flagQuietMode {
		files := textfmt.Pluralize(len(args), "FISHI input file", "-s")
		fmt.Printf("Reading %s...\n", files)
	}
	var joinedAST *fishi.AST

	for _, file := range args {
		// if we've been asked to show preprocessed, do that now by directly
		// building the CodeReader and reading the entire file.
		if *flagPreproc {
			err := printPreproc(file)
			if err != nil {
				errOther(err.Error())
				return
			}
		}

		res, err := fishi.ParseMarkdownFile(file, fo)

		if res.AST != nil {
			if joinedAST == nil {
				joinedAST = res.AST
			} else {
				joinedAST.Nodes = append(joinedAST.Nodes, res.AST.Nodes...)
			}
		}

		// parse tree is per-file, so we do this immediately even on error, as
		// it may be useful
		if res.Tree != nil && *flagGenTree {
			fmt.Printf("%s\n", trans.AddAttributes(*res.Tree).String())
		}

		if err != nil {
			// results may be valid even if there is an error
			if joinedAST != nil && *flagGenAST {
				fmt.Printf("%s\n", res.AST.String())
			}

			if syntaxErr, ok := err.(*types.SyntaxError); ok {
				errSyntax(file, syntaxErr)
			} else {
				errOther(fmt.Sprintf("%s: %s", file, err.Error()))
			}
			return
		}
	}

	if joinedAST == nil {
		panic("joinedAST is nil; should not be possible")
	}

	if *flagGenAST {
		fmt.Printf("%s\n", joinedAST.String())
	}

	// attempt to turn AST into a fishi.Spec
	if !*flagQuietMode {
		fmt.Printf("Generating language spec from FISHI...\n")
	}
	spec, warnings, err := fishi.NewSpec(*joinedAST)
	var fatalSpecWarn error
	// warnings may be valid even if there is an error
	if len(warnings) > 0 {
		for _, warn := range warnings {
			if wErr := warnHandler.Handlef("%s\n\n", warn); wErr != nil {
				fatalSpecWarn = wErr
			}
		}
	}
	// now check err
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		if syntaxErr, ok := err.(*types.SyntaxError); ok {
			errSyntax("", syntaxErr)
		} else {
			errOther(err.Error())
		}
		return
	} else if fatalSpecWarn != nil {
		errFatalWarn("fatal warning(s) occured")
		return
	}

	// we officially have a spec. try to print it if requested
	if *flagShowSpec {
		printSpec(spec)
	}

	// if no-gen is set and diagnostics binary not requested, we are done.
	if *flagNoGen && *flagDiagBin == "" {
		if !*flagQuietMode {
			fmt.Printf("(parser generation skipped due to flags)\n")
		}
		return
	}

	// spec completed and no-gen not set; try to create a parser
	var p ictiobus.Parser
	var parserWarns []fishi.Warning
	// if one is selected, use that one
	if parserType != nil {
		if !*flagQuietMode {
			fmt.Printf("Creating %s parser from spec...\n", *parserType)
		}
		p, parserWarns, err = spec.CreateParser(*parserType, allowAmbig)
	} else {
		if !*flagQuietMode {
			fmt.Printf("Creating most restrictive parser from spec...\n")
		}
		p, parserWarns, err = spec.CreateMostRestrictiveParser(allowAmbig)
	}

	// code gen time! 38D
	var fatalParserWarn error
	for _, warn := range parserWarns {
		if wErr := warnHandler.Handle(warn); wErr != nil {
			fatalParserWarn = wErr
		}
	}
	fmt.Fprintf(os.Stderr, "\n")

	if err != nil {
		errParser(err.Error())
		return
	} else if fatalParserWarn != nil {
		errFatalWarn("fatal warning(s) occured")
		return
	}

	if !*flagQuietMode {
		fmt.Printf("Successfully generated %s parser from grammar\n", p.Type().String())
	}

	// create a test compiler and output it
	if !*flagSimOff {
		if *flagIRType == "" {
			warn := fishi.Warning{
				Type:    fishi.WarnValidationArgs,
				Message: "skipping SDTS validation due to missing --ir parameter",
			}

			if wErr := warnHandler.Handle(warn); wErr != nil {
				errFatalWarn(wErr.Error())
				return
			}
		} else {
			if *flagHooksPath == "" {
				warn := fishi.Warning{
					Type:    fishi.WarnValidationArgs,
					Message: "skipping SDTS validation due to missing --hooks parameter",
				}

				if wErr := warnHandler.Handle(warn); wErr != nil {
					errFatalWarn(wErr.Error())
					return
				}
			} else {
				if !*flagQuietMode {
					simGenDir := fishi.SimGenerationDir
					if *flagPathPrefix != "" {
						simGenDir = filepath.Join(*flagPathPrefix, simGenDir)
					}
					fmt.Printf("Generating parser simulation binary in %s...\n", simGenDir)
				}
				di := trans.ValidationOptions{
					ParseTrees:    *flagSimTrees,
					FullDepGraphs: *flagSimGraphs,
					ShowAllErrors: !*flagSimFirstErrOnly,
					SkipErrors:    *flagSimSkipErrs,
				}

				simParams := fishi.SimulatedInputParams{
					Parser:              p,
					HooksPkgDir:         *flagHooksPath,
					HooksExpr:           *flagHooksTableName,
					PathPrefix:          *flagPathPrefix,
					LocalIctiobusSource: devInfo.LocalIctiobusSource,
					Opts:                cgOpts,
					ValidationOpts:      &di,
				}

				err := fishi.ValidateSimulatedInput(spec, md, simParams)
				if err != nil {
					errGeneration(err.Error())
					return
				}
			}
		}
	}

	// generate diagnostics output if requested
	if *flagDiagBin != "" {
		// already checked required flags
		if !*flagQuietMode {
			// tell user if the diagnostic binary cannot do preformatting based
			// on flags
			if *flagDiagFormatPkg == "" {
				fmt.Printf("Format preprocessing disabled in diagnostics bin; set -f to enable\n")
			}

			diagGenDir := fishi.DiagGenerationDir
			if *flagPathPrefix != "" {
				diagGenDir = filepath.Join(*flagPathPrefix, diagGenDir)
			}
			fmt.Printf("Generating diagnostics binary code in %s...\n", diagGenDir)
		}

		// only specify a format call if a format package was specified,
		// otherwise we'll always pass in a non-empty string for the format call
		// even when diagFormatPkg is empty, which is not allowed.
		var formatCall string
		if *flagDiagFormatPkg != "" {
			formatCall = *flagDiagFormatCall
		}

		diagParams := fishi.DiagBinParams{
			Parser:              p,
			HooksPkgDir:         *flagHooksPath,
			HooksExpr:           *flagHooksTableName,
			FormatPkgDir:        *flagDiagFormatPkg,
			FormatCall:          formatCall,
			FrontendPkgName:     *flagPkg,
			BinPath:             *flagDiagBin,
			LocalIctiobusSource: devInfo.LocalIctiobusSource,
			Opts:                cgOpts,
			PathPrefix:          *flagPathPrefix,
		}

		err := fishi.GenerateDiagnosticsBinary(spec, md, diagParams)
		if err != nil {
			errGeneration(err.Error())
			return
		}

		if !*flagQuietMode {
			fmt.Printf("Built diagnostics binary '%s'\n", *flagDiagBin)
		}
	}

	// if we are in no-gen mode, do not output anyfin, glub!
	if *flagNoGen {
		if !*flagQuietMode {
			fmt.Printf("(frontend code output skipped due to flags)\n")
		}
		return
	}

	// assuming it worked, now generate the final output
	if !*flagQuietMode {
		feGenDir := *flagDest
		if *flagPathPrefix != "" {
			feGenDir = filepath.Join(*flagPathPrefix, feGenDir)
		}
		fmt.Printf("Generating compiler frontend in %s...\n", feGenDir)
	}
	feDest := *flagDest
	if *flagPathPrefix != "" {
		feDest = filepath.Join(*flagPathPrefix, feDest)
	}

	// attempt to infer the import path for the frontend package for template
	// fill
	var feImportPath string
	feImportPath, err = inferImportPathFromDir(feDest)
	if err != nil {
		w := fishi.Warning{
			Type:    fishi.WarnImportInference,
			Message: "failed to infer import path for generated frontend: " + err.Error(),
		}

		if wErr := warnHandler.Handlef("%s\n; output will have syntax errors\n", w); wErr != nil {
			errFatalWarn(wErr.Error())
			return
		}

		feImportPath = "FE_IMPORT_PATH"
	}

	err = fishi.GenerateFrontendGo(spec, md, *flagPkg, feDest, feImportPath, &cgOpts)
	if err != nil {
		errGeneration(err.Error())
		return
	}

	parserPath := filepath.Join(feDest, "parser.cff")
	err = ictiobus.SaveParserToDisk(p, parserPath)
	if err != nil {
		errGeneration(err.Error())
		return
	}
}

func printPreproc(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	// dont do direct fs IO
	bufF := bufio.NewReader(f)

	cr, err := format.NewCodeReader(bufF)
	if err != nil {
		return err
	}

	// open a buffered reader on our code reader so we can read it line
	// by line
	bufCR := bufio.NewReader(cr)

	// read file line by line and print each line as it is read
	for {
		line, err := bufCR.ReadString('\n')
		fmt.Print(line)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

func devModeInfoFromFlags() (DevModeInfo, error) {
	dmi := DevModeInfo{}

	if *flagDev {
		dmi.Enabled = true

		// if user wants to enable dev mode, make sure that the current working
		// directory is the root of ictiobus by checking for a go.mod file and
		// then reading it to verify that it is for ictiobus.

		curDir, err := os.Getwd()
		if err != nil {
			return dmi, err
		}
		var modBytes []byte
		if modBytes, err = os.ReadFile(filepath.Join(curDir, "go.mod")); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return dmi, fmt.Errorf("current working directory does not contain a go.mod file")
			} else {
				return dmi, err
			}
		}
		ictiobusModFile, err := modfile.ParseLax("go.mod", modBytes, nil)
		if err != nil {
			return dmi, fmt.Errorf("go.mod: %w", err)
		}

		if ictiobusModFile.Module == nil {
			return dmi, fmt.Errorf("go.mod: module directive is missing")
		}

		if ictiobusModFile.Module.Mod.Path != "github.com/dekarrin/ictiobus" {
			return dmi, fmt.Errorf("local module is %s, not github.com/dekarrin/ictiobus", ictiobusModFile.Module.Mod.Path)
		}

		// all checks done, set the local source path
		dmi.LocalIctiobusSource = curDir
	}

	return dmi, nil
}

// return from flags the parser type selected and whether ambiguity is allowed.
// If no parser type is selected, nil is returned as first arg. if parser type
// does not allow ambiguity, allowAmbig will always be false.
//
// err will be non-nil if there is an invalid combination of CLI flags.
func parserSelectionFromFlags() (t *types.ParserType, allowAmbig bool, err error) {
	// enforce mutual exclusion of cli args
	if (*flagParserLL && (*flagParserCLR || *flagParserSLR || *flagParserLALR)) ||
		(*flagParserCLR && (*flagParserSLR || *flagParserLALR)) ||
		(*flagParserSLR && *flagParserLALR) {

		err = fmt.Errorf("cannot specify more than one parser type")
		return
	}

	allowAmbig = !*flagParserNoAmbig

	if *flagParserLL {
		t = new(types.ParserType)
		*t = types.ParserLL1

		// allowAmbig auto false for LL(1)
		allowAmbig = false
	} else if *flagParserSLR {
		t = new(types.ParserType)
		*t = types.ParserSLR1
	} else if *flagParserCLR {
		t = new(types.ParserType)
		*t = types.ParserCLR1
	} else if *flagParserLALR {
		t = new(types.ParserType)
		*t = types.ParserLALR1
	}

	return
}

func printSpec(spec fishi.Spec) {
	// print tokens
	fmt.Printf("Token Classes:\n")

	// find the longest token class ID
	maxTCLen := 0
	for _, tc := range spec.Tokens {
		if len(tc.ID()) > maxTCLen {
			maxTCLen = len(tc.ID())
		}
	}

	for _, tc := range spec.Tokens {
		// padding
		idPad := strings.Repeat(" ", maxTCLen-len(tc.ID()))
		fmt.Printf("* %s%s - %q\n", tc.ID(), idPad, tc.Human())
	}
	fmt.Printf("\n")

	// print lexer
	fmt.Printf("Lexer Patterns:\n")
	orderedStates := textfmt.OrderedKeys(spec.Patterns)

	for _, state := range orderedStates {
		pats := spec.Patterns[state]

		if state == "" {
			fmt.Printf("All States:\n")
		} else {
			fmt.Printf("State %s:\n", state)
		}

		for _, pat := range pats {
			fmt.Printf("* %s => ", pat.Regex.String())

			switch pat.Action.Type {
			case lex.ActionNone:
				fmt.Printf("(DISCARDED)")
			case lex.ActionScan:
				fmt.Printf("%s", pat.Action.ClassID)
			case lex.ActionState:
				fmt.Printf("GO TO STATE %s", pat.Action.State)
			case lex.ActionScanAndState:
				fmt.Printf("%s THEN GO TO STATE %s", pat.Action.ClassID, pat.Action.State)
			}

			if pat.Priority != 0 {
				fmt.Printf(", PRIORITY %d", pat.Priority)
			}

			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")

	// print grammar in custom way
	fmt.Printf("Grammar:\n")
	nts := spec.Grammar.PriorityNonTerminals()
	// ensure that the start symbol is first
	if nts[0] != spec.Grammar.StartSymbol() {
		startSymIdx := -1

		needle := spec.Grammar.StartSymbol()
		for i, nt := range nts {
			if nt == needle {
				startSymIdx = i
				break
			}
		}

		if startSymIdx == -1 {
			panic("start symbol not found in grammar")
		}

		nts[0], nts[startSymIdx] = nts[startSymIdx], nts[0]
	}

	// find the longest non-terminal name
	maxNTLen := 0
	for _, nt := range nts {
		if len(nt) > maxNTLen {
			maxNTLen = len(nt)
		}
	}

	nextPad := strings.Repeat(" ", maxNTLen)

	for _, nt := range nts {
		// head part is space-padded to align with longest non-terminal name
		r := spec.Grammar.Rule(nt)
		if r.NonTerminal == "" {
			panic("non-terminal not found in grammar")
		}

		headPad := strings.Repeat(" ", maxNTLen-len(r.NonTerminal))

		// first rule will be simply HEAD -> PROD
		ruleStr := fmt.Sprintf("%s%s -> %s\n", r.NonTerminal, headPad, r.Productions[0].String())

		// add alternatives
		for _, prod := range r.Productions[1:] {
			ruleStr += fmt.Sprintf("%s  | %s\n", nextPad, prod.String())
		}

		// print the final rule string
		fmt.Printf("%s", ruleStr)
	}
	fmt.Printf("\n")

	// print translation scheme
	fmt.Printf("Translation Scheme:\n")
	for _, sdd := range spec.TranslationScheme {
		fmt.Printf("* %s: ", sdd.Rule.String())
		lhsStr := sddRefToPrintedString(sdd.Attribute, spec.Grammar, sdd.Rule)
		fmt.Printf("%s = %s(", lhsStr, sdd.Hook)
		for i := range sdd.Args {
			if i != 0 {
				fmt.Printf(", ")
			}
			argStr := sddRefToPrintedString(sdd.Args[i], spec.Grammar, sdd.Rule)
			fmt.Printf("%s", argStr)
		}
		fmt.Printf(")\n")
	}
}

func sddRefToPrintedString(ref trans.AttrRef, g grammar.Grammar, r grammar.Rule) string {
	// which symbol does it refer to?
	var symName string
	if ref.Relation.Type == trans.RelHead {
		symName = "{" + r.NonTerminal + "$^}"
	} else if ref.Relation.Type == trans.RelSymbol {
		sym := r.Productions[0][ref.Relation.Index]
		// now find all indexes of that particular symbol in the rule

		inst := -1
		for i, s := range r.Productions[0] {
			if s == sym {
				inst++
			}
			if i == ref.Relation.Index {
				break
			}
		}
		if inst == -1 {
			// should never happen
			panic("symbol not found in rule")
		}

		symName = fmt.Sprintf("%s$%d", sym, inst)
		if g.IsNonTerminal(sym) {
			symName = "{" + symName + "}"
		}
	} else {
		// find the nth non-terminal in the rule
		curOfType := -1
		symIdx := -1
		for i, sym := range r.Productions[0] {
			if (ref.Relation.Type == trans.RelNonTerminal && g.IsNonTerminal(sym)) || (ref.Relation.Type == trans.RelTerminal && g.IsTerminal(sym)) {
				curOfType++
			}
			if curOfType == ref.Relation.Index {
				symIdx = i
				break
			}
		}

		if symIdx == -1 {
			// should never happen
			panic("non-terminal not found in rule")
		}

		sym := r.Productions[0][symIdx]
		// now find all indexes of that particular symbol in the rule

		inst := -1
		for i, s := range r.Productions[0] {
			if s == sym {
				inst++
			}
			if i == ref.Relation.Index {
				break
			}
		}
		if inst == -1 {
			// should never happen
			panic("symbol not found in rule")
		}

		symName = fmt.Sprintf("%s$%d", sym, inst)
		if g.IsNonTerminal(sym) {
			symName = "{" + symName + "}"
		}
	}

	// now the easy part, add the attribute name
	return fmt.Sprintf("%s.%s", symName, ref.Name)
}

func inferImportPathFromDir(dir string) (string, error) {
	getRealPath := func(path string) (string, error) {
		// first, get the full realpath, absolute:
		nonSym, err := filepath.EvalSymlinks(path)
		if err != nil {
			return "", err
		}
		abs, err := filepath.Abs(nonSym)
		if err != nil {
			return "", err
		}

		return filepath.ToSlash(abs), nil
	}

	// first, get the full realpath, absolute:
	absDir, err := getRealPath(dir)
	if err != nil {
		return "", err
	}

	checkDir := absDir
	candidateImport := ""
	foundGoMod := false
	// climb the directory and check to see if there's a go.mod file

	for !foundGoMod {
		// check if there's a go.mod file in this dir
		goModPath := filepath.Join(checkDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// there's a go.mod file here. read it and get the module name
			goModFile, err := os.ReadFile(goModPath)
			if err != nil {
				return "", err
			}

			// parse the go.mod file
			goMod, err := modfile.Parse(goModPath, goModFile, nil)
			if err != nil {
				return "", err
			}

			// get the module name
			moduleName := goMod.Module.Mod.Path

			// the import path is the module name + the relative path from the module
			// root to the dir.
			relPath, err := filepath.Rel(checkDir, absDir)
			if err != nil {
				return "", err
			}

			candidateImport = filepath.Join(moduleName, relPath)
			foundGoMod = true
		} else {
			// no go.mod file here. check the parent dir.
			parentDir := filepath.Dir(checkDir)
			if parentDir == checkDir {
				// we're at the root of the filesystem. we're done.
				break
			}

			checkDir = parentDir
		}
	}

	if foundGoMod {
		return filepath.ToSlash(candidateImport), nil
	}

	// next, try to get GOPATH:
	goPath, goPathSet := os.LookupEnv("GOPATH")
	if !goPathSet {
		goPath = build.Default.GOPATH
	}
	goPathParts := filepath.SplitList(goPath)

	for _, goPathPart := range goPathParts {
		absGoPathDir, err := getRealPath(goPathPart)
		if err != nil {
			return "", fmt.Errorf("GOPATH: %w", err)
		}
		goSrcPath := filepath.Join(absGoPathDir, "src")

		// check if absGoPathDir + /src is a prefix of absPath
		if strings.HasPrefix(absDir, goSrcPath) {
			// the import path is the path after the prefix.
			relPath, err := filepath.Rel(goSrcPath, absDir)
			if err != nil {
				return "", err
			}

			return filepath.ToSlash(relPath), nil
		}
	}

	// finally, check GOROOT:
	goRoot := runtime.GOROOT()
	absGoRootDir, err := getRealPath(goRoot)
	if err != nil {
		return "", fmt.Errorf("GOROOT: %w", err)
	}
	goSrcPath := filepath.Join(absGoRootDir, "src")

	// check if absGoRootDir + /src is a prefix of absPath
	if strings.HasPrefix(absDir, goSrcPath) {
		// the import path is the path after the prefix.
		relPath, err := filepath.Rel(goSrcPath, absDir)
		if err != nil {
			return "", err
		}

		return filepath.ToSlash(relPath), nil
	}

	return "", fmt.Errorf("path is not within a module, GOPATH, or GOROOT: %s", dir)
}
