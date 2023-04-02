/*
Ictcc runs the ictiobus compiler-compiler on one or more markdown files that
contain 'fishi' codeblocks.

It parses all files given as arguments and outputs a generated frontend for the
language specified by the fishi codeblocks.

Usage:

	ictcc [flags] file1.md file2.md ...

Flags:

	-a/-ast
		Print the AST to stdout before generating the frontend.

	-t/-tree
		Print the parse tree to stdout before generating the frontend.

	-s/-spec
		Print the interpreted language specification to stdout before generating
		the frontend.

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

	-pkg NAME
		Set the name of the package to place generated files in. Defaults to
		'fe'.

	-dest DIR
		Set the destination directory to place generated files in. Defaults to a
		directory named 'fe' in the current working directory.

	-l/lang NAME
		Set the name of the language to generate a frontend for. Defaults to
		"Unspecified".

	-lang-ver VERSION
		Set the version of the language to generate a frontend for. Defaults to
		"v0.0.0".

	-pre-format
		Enable dumping of the fishi filled template files before they are passed
		to the formatter. This allows debugging of the template files when
		editing them, since they must be valid go code to be formatted.

	-tmpl-tokens FILE
		Use the provided file as the template for outputting the generated
		tokens file instead of the default embedded within the binary.

	-tmpl-lexer FILE
		Use the provided file as the template for outputting the generated lexer
		file instead of the default embedded within the binary.

	-tmpl-parser FILE
		Use the provided file as the template for outputting the generated
		parser file instead of the default embedded within the binary.

	-tmpl-sdts FILE
		Use the provided file as the template for outputting the generated SDTS
		file instead of the default embedded within the binary.

	-tmpl-frontend FILE
		Use the provided file as the template for outputting the generated
		frontend file instead of the default embedded within the binary.

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
	"strings"

	"github.com/dekarrin/ictiobus/fishi"
	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/ictiobus/types"
	"github.com/dekarrin/rosed"
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

	// ExitErrGeneration is the code returned as exit status when there is an
	// error creating the generated files.
	ExitErrGeneration

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
	genSpec       bool
	parserCff     string
	lang          string
	dumpPreFormat *bool   = flag.Bool("pre-format", false, "Dump the generated code before running through gofmt")
	pkg           *string = flag.String("pkg", "fe", "The name of the package to place generated files in")
	dest          *string = flag.String("dest", "./fe", "The name of the directory to place the generated package in")
	langVer       *string = flag.String("lang-ver", "v0.0.0", "The version of the language to generate")
	noCache       *bool   = flag.Bool("no-cache", false, "Disable use of cached frontend components, even if available")
	noCacheOutput *bool   = flag.Bool("no-cache-out", false, "Disable writing of cached frontend components, even if one was generated")

	valSDTSOff        *bool = flag.Bool("val-sdts-off", false, "Disable validation of the SDTS of the resulting fishi")
	valSDTSShowTrees  *bool = flag.Bool("val-sdts-trees", false, "Show trees that caused SDTS validation errors")
	valSDTSShowGraphs *bool = flag.Bool("val-sdts-graphs", false, "Show full generated dependency graph output for parse trees that caused SDTS validation errors")
	valSDTSFirstOnly  *bool = flag.Bool("val-sdts-first", false, "Show only the first error found in SDTS validation")
	valSDTSSkip       *int  = flag.Int("val-sdts-skip", 0, "Skip the first N errors found in SDTS validation in output")

	tmplTokens *string = flag.String("tmpl-tokens", "", "A template file to replace the embedded tokens template with")
	tmplLexer  *string = flag.String("tmpl-lexer", "", "A template file to replace the embedded lexer template with")
	tmplParser *string = flag.String("tmpl-parser", "", "A template file to replace the embedded parser template with")
	tmplSDTS   *string = flag.String("tmpl-sdts", "", "A template file to replace the embedded SDTS template with")
	tmplFront  *string = flag.String("tmpl-frontend", "", "A template file to replace the embedded frontend template with")

	lexerTrace  *bool = flag.Bool("debug-lexer", false, "Print the lexer trace to stdout")
	parserTrace *bool = flag.Bool("debug-parser", false, "Print the parser trace to stdout")

	version *bool = flag.Bool("version", false, "Print the version of ictcc and exit")
)

func init() {
	const (
		noGenUsage       = "Do not generate the parser"
		genASTUsage      = "Print the AST of the analyzed fishi"
		genTreeUsage     = "Print the parse trees of each analyzed fishi file"
		genSpecUsage     = "Print the FISHI spec interpreted from the analyzed fishi"
		parserCffUsage   = "Use the specified parser CFF cache file instead of default"
		parserCffDefault = "fishi-parser.cff"
		langUsage        = "The name of the languae being generated"
		langDefault      = "Unspecified"
	)
	flag.BoolVar(&noGen, "no-gen", false, noGenUsage)
	flag.BoolVar(&noGen, "n", false, noGenUsage+" (shorthand)")
	flag.BoolVar(&genAST, "ast", false, genASTUsage)
	flag.BoolVar(&genAST, "a", false, genASTUsage+" (shorthand)")
	flag.BoolVar(&genSpec, "spec", false, genSpecUsage)
	flag.BoolVar(&genSpec, "s", false, genSpecUsage+" (shorthand)")
	flag.BoolVar(&genTree, "tree", false, genTreeUsage)
	flag.BoolVar(&genTree, "t", false, genTreeUsage+" (shorthand)")
	flag.StringVar(&parserCff, "parser", parserCffDefault, parserCffUsage)
	flag.StringVar(&parserCff, "p", parserCffDefault, parserCffUsage+" (shorthand)")
	flag.StringVar(&lang, "lang", langDefault, langUsage)
	flag.StringVar(&lang, "l", langDefault, langUsage+"(shorthand)")
}

func main() {
	// basic function to check if panic is happening and recover it while also
	// preserving possibly-set exit code.
	defer func() {
		if panicErr := recover(); panicErr != nil {
			// we are panicking, make sure we dont lose the panic just because
			// we checked
			panic("unrecoverable panic occured")
		} else {
			os.Exit(returnCode)
		}
	}()

	// gather options and arguments
	invocation := strings.Join(os.Args[1:], " ")

	flag.Parse()

	if *version {
		fmt.Println(GetVersionString())
		return
	}

	// create a spec metadata object
	md := fishi.SpecMetadata{
		Language:       lang,
		Version:        *langVer,
		InvocationArgs: invocation,
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

	cgOpts := fishi.CodegenOptions{
		DumpPreFormat: *dumpPreFormat,
		TemplateFiles: map[string]string{},
	}
	if *tmplTokens != "" {
		cgOpts.TemplateFiles[fishi.ComponentTokens] = *tmplTokens
	}
	if *tmplLexer != "" {
		cgOpts.TemplateFiles[fishi.ComponentLexer] = *tmplLexer
	}
	if *tmplParser != "" {
		cgOpts.TemplateFiles[fishi.ComponentParser] = *tmplParser
	}
	if *tmplSDTS != "" {
		cgOpts.TemplateFiles[fishi.ComponentSDTS] = *tmplSDTS
	}
	if *tmplFront != "" {
		cgOpts.TemplateFiles[fishi.ComponentFrontend] = *tmplFront
	}
	if len(cgOpts.TemplateFiles) == 0 {
		// just nil it
		cgOpts.TemplateFiles = nil
	}

	// now that args are gathered, parse markdown files into an AST
	var joinedAST *fishi.AST

	for _, file := range args {
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
		if res.Tree != nil && genTree {
			fmt.Printf("%s\n", trans.AddAttributes(*res.Tree).String())
		}

		if err != nil {
			// results may be valid even if there is an error
			if joinedAST != nil && genAST {
				fmt.Printf("%s\n", res.AST.String())
			}

			if syntaxErr, ok := err.(*types.SyntaxError); ok {
				fmt.Fprintf(os.Stderr, "%s:\n%s\n", file, syntaxErr.FullMessage())
				returnCode = ExitErrSyntax
			} else {
				fmt.Fprintf(os.Stderr, "%s: %s\n", file, err.Error())
				returnCode = ExitErrOther
			}
			return
		}
	}

	if joinedAST == nil {
		panic("joinedAST is nil; should not be possible")
	}

	if genAST {
		fmt.Printf("%s\n", joinedAST.String())
	}

	// attempt to turn AST into a fishi.Spec

	spec, warnings, err := fishi.NewSpec(*joinedAST)
	// warnings may be valid even if there is an error
	if len(warnings) > 0 {
		for _, warn := range warnings {
			const warnPrefix = "WARN: "
			// indent all except the first line
			warnStr := rosed.Edit(warnPrefix+warn.Message).
				LinesFrom(1).
				IndentOpts(len(warnPrefix), rosed.Options{IndentStr: " "}).
				String()

			fmt.Fprintf(os.Stderr, "%s\n\n", warnStr)
		}
	}
	// now check err
	if err != nil {
		fmt.Printf("\n")
		// TODO: at this point, it would be v nice to have file in the
		// token/syntax error output. Allow specification of file in anyfin that
		// can return a SyntaxError and have all token sources include that.
		if syntaxErr, ok := err.(*types.SyntaxError); ok {
			fmt.Fprintf(os.Stderr, "%s\n", syntaxErr.FullMessage())
			returnCode = ExitErrSyntax
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			returnCode = ExitErrOther
		}
		return
	}

	// we officially have a spec. try to print it if requested
	if genSpec {
		printSpec(spec)
	}

	if !noGen {
		// TODO: jello, first need to create a PRELIM generation along with hooks
		// pkg because that is the only way to validate the SDTS.

		// do processing of the AST here
		err := fishi.GenerateCompilerGo(spec, md, *pkg, *dest, &cgOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			returnCode = ExitErrGeneration
			return
		}
		fmt.Printf("(NOTE: complete frontend generation not implemented yet)\n")
	}

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

		// TODO: sort output pats by priority

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
