/*
Ictcc produces compiler frontends written in Go from frontend specifications
written in FISHI.

Usage:

	ictcc [flags] FILE ...

Ictcc reads in the provided FISHI code, either from a file specified as its
args, from CLI flag -C, from stdin by specifying file "-", or some combination
of the above. All FISHI read is combined into a single spec, which is then used
to generate a compiler frontend that is output as Go code.

All input must be UTF-8 encoded markdown-formatted text that contains code
blocks marked with the label `fishi`; only those codeblocks are read for FISHI
source code. The contents of all such codeblocks for an input are concatenated
together to form the "FISHI part" of an input. This concatenated series of FISHI
statements then has comment stripping and line normalization applied to it
before it is parsed into an AST.

When all inputs have been successfully parsed, their ASTs are joined into a
single one by concatenation in the order the inputs they were parsed from were
given, and that AST is then interpreted into a language spec.

This language spec is then used to create a lexer, parser, and then translation
scheme for the language described in the spec. The parser algorithm will be the
one specified by CLI flags; otherwise, the most restrictive one supported that
can handle the grammar is used.

If the --ir and --hooks options are provided, the generated frontend is then
validated by building it into a simulation binary which then simulates language
input against the frontend, covering every possible production in the grammar.
Any issues found at this stage are output; otherwise, the binary and its sources
are deleted.

The Go code for the generated frontend is then placed in a local directory;
"./fe" by default, which can be changed with --dest. The name of the package it
is placed in, "fe" by default, can be changed with the --pkg flag. The language
metadata, retrievable from the generated frontend, can be set by using the -l
and -v flags.

If an error occurs while parsing any of the FISHI, ictcc will still try to parse
any remaining input files for error reporting purposes, but will ultimately fail
to produce generated code. All files must contain parsable FISHI.

Flags:

	-a, --ast
		Print the AST of successfully read FISHI files to stdout.

	-c, --diag-format-call NAME
		Call the function called NAME in the package given by --diag-format-pkg
		when obtaining a code io.Reader in a generated diagnostics binary. This
		is "NewCodeReader" by default. --diag-format-call has no effect unless
		--diag-format-pkg and --diag are also set.

	--clr
		Generate a Canonical LR(k) parser. Mutually exclusive with --ll, --slr,
		and --lalr.

	-C, --command CODE
		Read the FISHI markdown document in CODE before any other input is read.

	-d, --diag FILE
		Generate a diagnostics binary from the spec and output it to the path
		FILE. This binary will contain a self-contained version of the generated
		frontend and can be used to validate it by attempting to use it to parse
		input files in the language the frontend was generated for. This flag
		requires the --ir and --hooks flags to also be set. By default the
		generated binary will not do any preprocessing of input files; to enable
		it, use the --diag-format-pkg flag.

	--debug-lexer
		Print each token as it lexed from FISHI input.

	--debug-parser
		Print each step the parser takes as it parsers FISHI input.

	--debug-templates
		Dump templates after they are filled for codegen but before they are
		formatted by gofmt, along with line numbers for easy reference.

	--dest PATH
		Place the generated Go files in a package rooted at PATH. The default
		value is "./fe".

	--dev
		Enable the use of and reference to ictiobus code located in the current
		working directory as it is currently written as opposed to using the
		latest release version of ictiobus. If environment variable
		ICTIOBUS_SOURCE is set, that will be used instead of the current working
		directory.

	-D, --dfa
		Print a detailed representation of the DFA that is constructed for the
		generated parser to stdout.

	--exp FEATURE
		Enable experimental or untested feature FEATURE. The allowed values for
		FEATURE are as follows for this version of ictcc: "inherited-attributes"
		and "all".

	-f, --diag-format-pkg PATH
		Enable format reading in generated diagnostic binary specified with
		--diag by using the io.Reader provided by the Go package located at
		PATH. This package must provide a function that matches the signature
		"NewCodeReader(io.Reader) (io.Reader, error)", though the returned type
		can be any type that implements io.Reader. The name of that function can
		be selected with --diag-format-call. --diag-format-pkg has no effect
		unless --diag is also set.

	-F, --fatal WARNTYPE
		Treat WARNTYPE warnings as fatal. If the specified type of warning is
		encountered, ictcc will output it as though it were an error and
		immediately halt. Valid values for WARNTYPE are "dupe-human",
		"missing-human", "priority", "unused", "ambig", "validation", "import",
		"val-args", "exp-inherited-attributes", and "all". This flag may be
		specified multiple times and in conjunction with -S flags; if both -F
		and -S are specified for a warning, -F takes precedence.

	--hooks PATH
		Retrieve the hooks table binding translation scheme hooks to their
		implementations from the Go package located in the directory specified
		by PATH. The package must contain an exported var named "HooksTable" of
		type trans.HookMap. The name of the var searched for can be set with
		--hooks-table if needed.

	--hooks-table NAME
		Set the name of the exported hooks table variable in the Go package
		located at the path specified by --hooks. NAME must be the name of an
		exported var of type trans.HookMap. The default value is "HooksTable".

	--ir TYPE
		Set the type of the IR returned by the generated frontend to TYPE. TYPE
		must be either an unqualified basic type (such as "int" or "float32"),
		or, if using requires importing a package, the import path of the
		package, followed by a dot, followed by the name of the type (such as
		"github.com/dekarrin/ictiobus/fishi/syntax.AST", or
		"*crypto/x509/pkix.Name"). Packages with a different name than the last
		component of their import path are not supported at this time. Pointer
		types and slice types are both supported; map types are not. If --ir is
		provided, the Frontend() function in the generated Go package will be
		prefilled with this type, making it so callers of Frontend() do not need
		to supply it at runtime.

	-l, --lang NAME
		Set the language name in the metadata of the generated frontend to NAME.
		The default value is "Unspecified".

	--lalr
		Generate an LALR(k) parser. Mutually exclusive with --ll, --slr, and
		--clr.

	--ll
		Generate an LL(k) parser. Mutually exclusive with --lalr, --slr, and
		--clr.

	-n, --no-gen
		Do not output a Go package with source code files that contain the
		generated frontend. If no other options that would cause spec processing
		are provided, this will cause ictcc to stop after the spec has been
		read.

	--no-ambig
		Disallow generation for specs that define an ambiguous context-free
		grammar.

	--pkg NAME
		Set the name of the package the generated Go source files will be placed
		in. The default value is "fe".

	--prefix PATH
		Prefix the path of all generated source files with PATH. This includes
		source files used as part of creating binaries as well as the output
		directory specified by --dest. This does not affect the location of the
		diagnostics binary specified with --diag.

	--preserve-bin-source
		Do not delete source files that are generated in the process of
		producing a binary (simulation or diagnostics), even if the binary is
		successfully built.

	-P, --preproc
		Show input FISHI after preprocessing is executed on it; this will be the
		FISHI that is directly provided to the lexer after it is gathered from
		codeblocks in the input markdown document.

	-q, --quiet
		Enable quiet mode; do not output progress or supplemantary messages.
		Output specifically requested via other flags or caused by warnings or
		errors is not affected by this flag.

	-s, --spec
		Print a formatted listing of the complete spec out once it is read from
		FISHI input files.

	--sim-first-err
		Print only the first error returned from language input simulation,
		after any that are skipped by --sim-skip-errs.

	--sim-graphs
		Print the full dependency graph info for any issue found during language
		input simulation that involves translation scheme dependency graphs.

	--sim-off
		Disable language input simulation, even if --ir and --hooks flags are
		provided.

	--sim-skip-errs N
		Skip outputting the first N errors encountered during language input
		simulation. Note that simulation errors will still cause ictcc to halt
		generation even if their output is suppressed.

	--sim-trees
		Print the parse trees of any inputs found to cause issues during
		language input simulation.

	--slr
		Generate a Simple LR(k) parser. Mutually exclusive with --ll, --lalr,
		and --clr.

	-S, --suppress WARNTYPE
		Suppress the output of WARNTYPE warnings. If the specified type of
		warning is encountered, ictcc will ignore it. Valid values for WARNTYPE
		are the same as for --fatal. This flag may be specified multiple times
		and in conjunction with -F flags; if both -F and -S are specified for a
		warning, -F takes precedence.

	-t, --tree
		Print the parse tree of successfully parsed FISHI files to stdout.

	--tmpl-frontend FILE
		Use the contents of FILE as the template to generate frontend.ict.go
		with during codegen.

	--tmpl-lexer FILE
		Use the contents of FILE as the template to generate lexer.ict.go with
		during codegen.

	--tmpl-main FILE
		Use the contents of FILE as the template to generate main.go with during
		codegen for binaries.

	--tmpl-parser FILE
		Use the contents of FILE as the template to generate parser.ict.go with
		during codegen.

	--tmpl-sdts FILE
		Use the contents of FILE as the template to generate sdts.ict.go with
		during codegen.

	--tmpl-tokens FILE
		Use the contents of FILE as the template to generate tokens.ict.go with
		during codegen.

	-T, --parse-table
		Print the parse table of the parser generated from the spec to stdout.

	-v, --lang-ver VERSION
		Set the language version in the metadata of the generated frontend to
		VERSION. The default value is "v0.0".

	--version
		Print the current version of ictcc and then exit.
*/
package main

import (
	"bufio"
	"bytes"
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

	"github.com/dekarrin/ictiobus/fishi"
	"github.com/dekarrin/ictiobus/fishi/format"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/ictiobus/syntaxerr"
	"github.com/dekarrin/ictiobus/trans"
)

var (
	flagWarnFatal    = pflag.StringArrayP("fatal", "F", nil, "Treat given warning as a fatal error")
	flagWarnSuppress = pflag.StringArrayP("suppress", "S", nil, "Suppress output of given warning")

	flagExp        = pflag.StringArray("exp", nil, "Experimental feature to enable")
	flagCommand    = pflag.StringP("command", "C", "", "Code to execute before any source code files are read")
	flagQuietMode  = pflag.BoolP("quiet", "q", false, "Suppress progress messages and other supplementary output")
	flagNoGen      = pflag.BoolP("no-gen", "n", false, "Do not output generated frontend output files")
	flagGenAST     = pflag.BoolP("ast", "a", false, "Print the AST of the analyzed fishi")
	flagGenTree    = pflag.BoolP("tree", "t", false, "Print the parse trees of each analyzed fishi file")
	flagShowSpec   = pflag.BoolP("spec", "s", false, "Print the FISHI spec interpreted from the analyzed fishi")
	flagLang       = pflag.StringP("lang", "l", "Unspecified", "The name of the languae being generated")
	flagLangVer    = pflag.StringP("lang-ver", "v", "v0.0", "The version of the language to generate")
	flagPreproc    = pflag.BoolP("preproc", "P", false, "Print the preprocessed FISHI code before compiling it")
	flagParseTable = pflag.BoolP("parse-table", "T", false, "Print the parse table used by the generated parser")
	flagDFA        = pflag.BoolP("dfa", "D", false, "Print the complete DFA of the parser")

	flagDiagBin        = pflag.StringP("diag", "d", "", "Generate binary that has the generated frontend and uses it to analyze the target language")
	flagDiagFormatPkg  = pflag.StringP("diag-format-pkg", "f", "", "The package containing format functions for the diagnostic binary to call on input prior to passing to frontend analysis")
	flagDiagFormatCall = pflag.StringP("diag-format-call", "c", "NewCodeReader", "The function within the diag-format-pkg to call to open a reader on input prior to passing to frontend analysis")

	flagPathPrefix        = pflag.String("prefix", "", "Path to prepend to path of all generated source files")
	flagPreserveBinSource = pflag.Bool("preserve-bin-source", false, "Preserve the source of any generated binary files")
	flagDebugTemplates    = pflag.Bool("debug-templates", false, "Dump the filled templates before running through gofmt")
	flagPkg               = pflag.String("pkg", "fe", "The name of the package to place generated files in")
	flagDest              = pflag.String("dest", "./fe", "The name of the directory to place the generated package in")

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

	flagHooksPath      = pflag.String("hooks", "", "The path to the hooks directory to use for the generated parser. Required for SDTS validation")
	flagHooksTableName = pflag.String("hooks-table", "HooksTable", "Function call or name of exported var in 'hooks' that has the hooks table")

	flagIRType = pflag.String("ir", "", "The fully-qualified type of IR to generate")

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

	expFeatures, err := experimentalFeaturesFromFlags()
	if err != nil {
		errInvalidFlags("--exp: " + err.Error())
		return
	}

	// check args before gathering flags
	args := pflag.Args()

	if len(args) < 1 && *flagCommand == "" {
		errNoFiles("No files given to process")
		return
	}

	// create a spec metadata object
	md := fishi.SpecMetadata{
		Language:       *flagLang,
		Version:        *flagLangVer,
		InvocationArgs: invocation,
	}

	fo := &fishi.Options{
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

	// now that args are gathered, parse any CLI commands and markdown files
	// into an AST
	var joinedAST *fishi.AST

	if *flagCommand != "" {
		var cmdReader io.Reader
		cmdBuf := bytes.NewBuffer([]byte(*flagCommand))

		// do preloading step of reading and immediately outputting the preprocessed file
		if *flagPreproc {
			err := printPreproc(cmdBuf)
			if err != nil {
				errOther(err.Error())
				return
			}

			// then reset the buffer
			cmdBuf = bytes.NewBuffer([]byte(*flagCommand))
		}

		cmdReader = cmdBuf

		cmdRes, cmdErr := fishi.ParseMarkdown(cmdReader, fo)

		if cmdRes.AST != nil {
			if joinedAST == nil {
				joinedAST = cmdRes.AST
			} else {
				joinedAST.Nodes = append(joinedAST.Nodes, cmdRes.AST.Nodes...)
			}
		}

		// parse tree is per-file, so we do this immediately even on error, as
		// it may be useful
		if cmdRes.Tree != nil && *flagGenTree {
			fmt.Printf("%s\n", trans.Annotate(*cmdRes.Tree).String())
		}

		if cmdErr != nil {
			// results may be valid even if there is an error
			if joinedAST != nil && *flagGenAST {
				fmt.Printf("%s\n", cmdRes.AST.String())
			}

			if syntaxErr, ok := cmdErr.(*syntaxerr.Error); ok {
				errSyntax("<COMMAND>", syntaxErr)
			} else {
				errOther(fmt.Sprintf("%s: %s", "<COMMAND>", err.Error()))
			}
			return
		}
	}

	if len(args) > 0 {
		if !*flagQuietMode {
			files := textfmt.Pluralize(len(args), "FISHI input file", "-s")
			fmt.Printf("Reading %s...\n", files)
		}

		var haveReadStdin bool
		for _, file := range args {
			if file == "-" {
				if haveReadStdin {
					continue
				} else {
					haveReadStdin = true
				}
			}

			// can't just use os.Stdin because the preprocess print could clober it.
			// instead, declare a reader and just use that IF we end up needing it.
			var rewoundStdinReader io.Reader

			// if we've been asked to show preprocessed, do that now by directly
			// building the CodeReader and reading the entire file.
			if *flagPreproc {
				// if we just read stdin err we are going to need the 'rewound'
				// version.
				rewoundStdinReader, err = printPreprocFile(file)
				if err != nil {
					errOther(err.Error())
					return
				}
			}

			var res fishi.Results

			if file == "-" {
				// read from stdin
				var readFrom io.Reader

				readFrom = os.Stdin
				if rewoundStdinReader != nil {
					// stdin already read by preprocess, so use the buffer we tee'd
					// to instead of os.Stdin directly
					readFrom = rewoundStdinReader
				}
				res, err = fishi.ParseMarkdown(readFrom, fo)
			} else {
				res, err = fishi.ParseMarkdownFile(file, fo)
			}

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
				fmt.Printf("%s\n", trans.Annotate(*res.Tree).String())
			}

			if err != nil {
				// results may be valid even if there is an error
				if joinedAST != nil && *flagGenAST {
					fmt.Printf("%s\n", res.AST.String())
				}

				if syntaxErr, ok := err.(*syntaxerr.Error); ok {
					errSyntax(file, syntaxErr)
				} else {
					errOther(fmt.Sprintf("%s: %s", file, err.Error()))
				}
				return
			}
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
			// is it an experimental feature for inherited attr?
			if warn.Type == fishi.WarnEFInheritedAttributes {
				if _, enabled := expFeatures[featureInheritedAttributes]; !enabled {
					errOther(warn.Message)
					return
				}
			}
			if wErr := warnHandler.Handlef("%s\n\n", warn); wErr != nil {
				fatalSpecWarn = wErr
			}
		}
	}
	// now check err
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		if syntaxErr, ok := err.(*syntaxerr.Error); ok {
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

	// if no-gen is set and diagnostics binary not requested and DFA not
	// requested, we are done.
	if *flagNoGen && *flagDiagBin == "" && !*flagDFA && !*flagParseTable {
		if !*flagQuietMode {
			fmt.Printf("(parser generation skipped due to flags)\n")
		}
		return
	}

	// spec completed and no-gen not set; try to create a parser
	var p parse.Parser
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

	var fatalParserWarn error
	for _, warn := range parserWarns {
		if wErr := warnHandler.Handle(warn); wErr != nil {
			fatalParserWarn = wErr
		}
	}

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

	// code gen time! 38D

	// output dfa if requested
	if *flagDFA {
		fmt.Printf("%s\n", p.DFAString())
	}

	// output parse table if requested
	if *flagParseTable {
		fmt.Printf("%s\n", p.TableString())
	}

	// create a test compiler and output it if either codegen or diagnostic bin
	// is enabled.
	if (!*flagNoGen || *flagDiagBin != "") && !*flagSimOff {
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
					WarningHandler:      warnHandler,
					QuietMode:           *flagQuietMode,
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
	err = parse.WriteFile(p, parserPath)
	if err != nil {
		errGeneration(err.Error())
		return
	}
}

func printPreproc(r io.Reader) error {
	cr, err := format.NewCodeReader(r)
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

// printPreprocFile will return a 'rewound' version if and only if it is operating
// on stdin (file is "-"). Else, rewoundStdin will be nil.
func printPreprocFile(file string) (rewoundStdin io.Reader, err error) {
	var f io.Reader

	if file == "-" {
		// actually, read from stdin.
		var buf *bytes.Buffer
		f = io.TeeReader(os.Stdin, buf)
		rewoundStdin = buf
	} else {
		fileReader, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer fileReader.Close()
		f = fileReader
	}

	// dont do direct fs IO
	bufF := bufio.NewReader(f)

	err = printPreproc(bufF)
	if err != nil {
		return nil, err
	}

	return rewoundStdin, nil
}

func experimentalFeaturesFromFlags() (map[expFeature]struct{}, error) {
	if *flagExp == nil {
		return nil, nil
	}

	enabled := map[expFeature]struct{}{}
	for i := range *flagExp {
		expStr := strings.ToLower((*flagExp)[i])
		if expStr == "all" {
			all := expFeatureAll()
			for _, f := range all {
				enabled[f] = struct{}{}
			}
			return enabled, nil
		}
		exp, err := parseShortExpFeature(expStr)
		if err != nil {
			return nil, err
		}
		if exp == featureNone {
			return nil, fmt.Errorf("cannot select feature 'none'")
		}
		enabled[exp] = struct{}{}
	}

	return enabled, nil
}

func devModeInfoFromFlags() (DevModeInfo, error) {
	dmi := DevModeInfo{}

	if *flagDev {
		dmi.Enabled = true

		// if user wants to enable dev mode, make sure that the current working
		// directory is the root of ictiobus by checking for a go.mod file and
		// then reading it to verify that it is for ictiobus, or if
		// ICTIOBUS_SOURCE is set, use that instead of current working dir.

		var sourceDir string
		var usedCwd bool
		if os.Getenv("ICTIOBUS_SOURCE") != "" {
			sourceDir = os.Getenv("ICTIOBUS_SOURCE")
		} else {
			curDir, err := os.Getwd()
			if err != nil {
				return dmi, err
			}
			sourceDir = curDir
			usedCwd = true
		}

		var err error
		var modBytes []byte
		if modBytes, err = os.ReadFile(filepath.Join(sourceDir, "go.mod")); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				errDir := sourceDir
				if usedCwd {
					errDir = "current working directory"
				}

				return dmi, fmt.Errorf("%s does not contain a go.mod file", errDir)
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
		dmi.LocalIctiobusSource = sourceDir
	}

	return dmi, nil
}

// return from flags the parser type selected and whether ambiguity is allowed.
// If no parser type is selected, nil is returned as first arg. if parser type
// does not allow ambiguity, allowAmbig will always be false.
//
// err will be non-nil if there is an invalid combination of CLI flags.
func parserSelectionFromFlags() (t *parse.Algorithm, allowAmbig bool, err error) {
	// enforce mutual exclusion of cli args
	if (*flagParserLL && (*flagParserCLR || *flagParserSLR || *flagParserLALR)) ||
		(*flagParserCLR && (*flagParserSLR || *flagParserLALR)) ||
		(*flagParserSLR && *flagParserLALR) {

		err = fmt.Errorf("cannot specify more than one parser type")
		return
	}

	allowAmbig = !*flagParserNoAmbig

	if *flagParserLL {
		t = new(parse.Algorithm)
		*t = parse.LL1

		// allowAmbig auto false for LL(1)
		allowAmbig = false
	} else if *flagParserSLR {
		t = new(parse.Algorithm)
		*t = parse.SLR1
	} else if *flagParserCLR {
		t = new(parse.Algorithm)
		*t = parse.CLR1
	} else if *flagParserLALR {
		t = new(parse.Algorithm)
		*t = parse.LALR1
	}

	return
}

func printSpec(spec fishi.Spec) {
	// print tokens
	fmt.Printf("Token Classes:\n")
	if len(spec.Tokens) == 0 {
		fmt.Printf("(no tokens defined)\n")
	} else {
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
	}
	fmt.Printf("\n")

	// print lexer
	fmt.Printf("Lexer Patterns:\n")
	orderedStates := textfmt.OrderedKeys(spec.Patterns)
	if len(orderedStates) == 0 {
		fmt.Printf("(no patterns defined)\n")
	} else {
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
	}
	fmt.Printf("\n")

	// print grammar in custom way
	fmt.Printf("Grammar:\n")
	nts := spec.Grammar.NonTerminalsByPriority()
	if len(nts) == 0 {
		fmt.Printf("(no rules defined)\n")
	} else {
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
	}
	fmt.Printf("\n")

	// print translation scheme
	fmt.Printf("Translation Scheme:\n")
	if len(spec.TranslationScheme) == 0 {
		fmt.Printf("(no actions defined)\n")
	} else {
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
}

func sddRefToPrintedString(ref trans.AttrRef, g grammar.CFG, r grammar.Rule) string {
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
			if errors.Is(err, os.ErrNotExist) {
				// just assume it's not a symlink; it doesn't yet exist
				nonSym = path
			} else {
				return "", err
			}
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
