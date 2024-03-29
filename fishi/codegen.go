package fishi

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/shellout"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/rosed"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/runenames"
)

// file codegen implements conversion of a fishi spec to a series of go files.
// This is the only way that we can do full validation with a hooks package.

const (
	// CommandName is the name of the ictiobus binary.
	CommandName = "ictcc"
)

// Names of each component of the generated compiler. Each component
// represents one file that is generated.
const (
	ComponentTokens   = "tokens"
	ComponentLexer    = "lexer"
	ComponentParser   = "parser"
	ComponentSDTS     = "sdts"
	ComponentFrontend = "frontend"
	ComponentMainFile = "main"
)

// Names of each file that is generated.
const (
	generatedTokensFilename   = "tokens.ict.go"
	generatedLexerFilename    = "lexer.ict.go"
	generatedParserFilename   = "parser.ict.go"
	generatedSDTSFilename     = "sdts.ict.go"
	generatedFrontendFilename = "frontend.ict.go"
	generatedMainFilename     = "main.ict.go"
)

// Default template strings for each component of the generated compiler.
var (
	//go:embed templates/tokens.go.tmpl
	templateTokens string

	//go:embed templates/lexer.go.tmpl
	templateLexer string

	//go:embed templates/parser.go.tmpl
	templateParser string

	//go:embed templates/sdts.go.tmpl
	templateSDTS string

	//go:embed templates/frontend.go.tmpl
	templateFrontend string

	//go:embed templates/main.go.tmpl
	templateMainFile string
)

var (
	underscoreCollapser = regexp.MustCompile(`_+`)
	unsafeRunenameChars = regexp.MustCompile(`[^A-Za-z0-9_]`)
	titleCaser          = cases.Title(language.AmericanEnglish)

	// order in which components of the generated compiler (files) are created.
	// not all will be present in all cases; this only enforces an ordering on
	// rendered templates to aid in debugging.
	codegenOrder = []string{
		ComponentTokens,
		ComponentLexer,
		ComponentParser,
		ComponentSDTS,
		ComponentFrontend,
		ComponentMainFile,
	}

	defaultTemplates = map[string]string{
		ComponentTokens:   templateTokens,
		ComponentLexer:    templateLexer,
		ComponentParser:   templateParser,
		ComponentSDTS:     templateSDTS,
		ComponentFrontend: templateFrontend,
		ComponentMainFile: templateMainFile,
	}
)

// executeTestCompiler runs the compiler pointed to by gci in validation mode.
//
// If valOptions is nil, the default validation options are used.
func executeTestCompiler(gci GeneratedCodeInfo, valOptions *trans.ValidationOptions, warnOpts *WarnHandler, quietMode bool) error {
	if valOptions == nil {
		valOptions = &trans.ValidationOptions{}
	}
	if warnOpts == nil {
		warnOpts = NewWarnHandler()
	}

	args := []string{"run", gci.MainFile, "--sim"}
	if valOptions.FullDepGraphs {
		args = append(args, "--sim-graphs")
	}
	if valOptions.ParseTrees {
		args = append(args, "--sim-trees")
	}
	if !valOptions.ShowAllErrors {
		args = append(args, "--sim-first-err")
	}
	if valOptions.SkipErrors != 0 {
		args = append(args, "--sim-skip-errs", fmt.Sprintf("%d", valOptions.SkipErrors))
	}
	if quietMode {
		args = append(args, "-q")
	}

	for _, wt := range WarnTypeAll() {
		if warnOpts.HandlingType(wt) == WarnHandlingFatal {
			args = append(args, "-F", wt.Short())
		} else if warnOpts.HandlingType(wt) == WarnHandlingSuppress {
			args = append(args, "-S", wt.Short())
		}
	}

	env := append(os.Environ(), "GOWORK=off")
	return shellout.ExecFG(gci.Path, env, "go", args...)
}

// GenerateDiagnosticsBinary generates a binary that can read input written in
// the language specified by the given Spec and SpecMetadata and print out basic
// information about the analysis, with the goal of printing out the
// constructed intermediate representation from analyzed files.
//
// The args formatPkgDir and formatCall are used to specify preformatting for
// code that that the generated binary will analyze. If set, io.Readers that are
// opened on code input will be passed to the function specified by formatCall.
// If formatCall is set, formatPkgDir must also be set, even if it is already
// specified by another parameter. formatCall must be the name of a function
// within the package specified by formatPkgDir which takes an io.Reader and
// returns a new io.Reader that wraps the one passed in and returns preformatted
// code ready to be analyzed by the generated frontend.
//
// localSource only needed if doing dev in ictiobus; otherwise latest ictiobus
// published version is used.
func GenerateDiagnosticsBinary(spec Spec, md SpecMetadata, params DiagBinParams) error {
	if len(spec.Tokens) == 0 {
		return fmt.Errorf("spec defines no tokens")
	}
	if len(spec.Patterns) == 0 {
		return fmt.Errorf("spec defines no lexer patterns")
	}
	if len(spec.Grammar.NonTerminals()) == 0 {
		return fmt.Errorf("spec defines no grammar rules")
	}
	if len(spec.TranslationScheme) == 0 {
		return fmt.Errorf("spec defines no translation scheme actions")
	}

	binName := filepath.Base(params.BinPath)

	outDir := DiagGenerationDir
	if params.PathPrefix != "" {
		outDir = filepath.Join(params.PathPrefix, outDir)
	}

	gci, err := generateBinaryMainGo(spec, md, mainBinaryParams{
		Parser:              params.Parser,
		HooksPkgDir:         params.HooksPkgDir,
		HooksExpr:           params.HooksExpr,
		FormatPkgDir:        params.FormatPkgDir,
		FormatCall:          params.FormatCall,
		FrontendPkgName:     params.FrontendPkgName,
		GenPath:             outDir,
		BinName:             binName,
		Opts:                params.Opts,
		LocalIctiobusSource: params.LocalIctiobusSource,
	})
	if err != nil {
		return err
	}

	env := os.Environ()
	env = append(env, "CGO_ENABLED=0", "GOWORK=off")
	if err := shellout.ExecFG(gci.Path, env, "go", "build", "-o", binName, gci.MainFile); err != nil {
		return err
	}

	// Move it to the target location.
	if err := os.Rename(filepath.Join(gci.Path, binName), params.BinPath); err != nil {
		return err
	}

	// unless requested to preserve the source, remove the generated source
	// directory.
	if !params.Opts.PreserveBinarySource {
		if err := os.RemoveAll(gci.Path); err != nil {
			return err
		}
	}

	return nil
}

// generateBinaryMainGo generates (but does not yet run) a test compiler for the
// given spec and pre-created parser, using the provided hooks package. Once it
// is created, it will be able to be executed by calling go run on the provided
// mainfile from the outDir.
//
// It will be created in a temporary file; caller must do os.RemoveAll() on the
// returned path when done.
//
// hooksExpr must be set to an exported identifier or func call that can be
// called from within the hooks package with a type of
// map[string]trans.AttributeSetter. It can be a function call, constant name,
// or var name.
//
// opts must be non-nil and IRType must be set.
func generateBinaryMainGo(spec Spec, md SpecMetadata, params mainBinaryParams) (GeneratedCodeInfo, error) {
	if params.Opts.IRType == "" {
		return GeneratedCodeInfo{}, fmt.Errorf("IRType must be set in options")
	}
	if params.FormatPkgDir != "" && params.FormatCall == "" {
		return GeneratedCodeInfo{}, fmt.Errorf("formatCall must be set if formatPkgDir is set")
	}
	if params.FormatPkgDir == "" && params.FormatCall != "" {
		return GeneratedCodeInfo{}, fmt.Errorf("formatPkgDir must be set if formatCall is set")
	}

	// we need a separate import for the format package only if it's not the same
	// as the hooks package.
	separateFormatImport := params.FormatPkgDir != params.HooksPkgDir && params.FormatPkgDir != ""

	irFQPackage, irType, irPackage, irTypeBare, irErr := parseFQType(params.Opts.IRType)
	if irErr != nil {
		return GeneratedCodeInfo{}, fmt.Errorf("parsing IRType: %w", irErr)
	}

	gci := GeneratedCodeInfo{}

	hooksPkgName, err := readPackageName(params.HooksPkgDir)
	if err != nil {
		return gci, fmt.Errorf("reading hooks package name: %w", err)
	}

	var formatPkgName string
	if params.FormatPkgDir != "" {
		formatPkgName, err = readPackageName(params.FormatPkgDir)
		if err != nil {
			return gci, fmt.Errorf("reading format package name: %w", err)
		}
	}

	if hooksPkgName == params.FrontendPkgName {
		// double it to avoid name collision
		params.FrontendPkgName += "_" + params.FrontendPkgName
	}

	// only worry about formatPkgName if the dir is not the same as hooks.
	if params.FormatPkgDir != params.HooksPkgDir && formatPkgName == params.FrontendPkgName {
		// double it to avoid name collision
		params.FrontendPkgName += "_" + params.FrontendPkgName
	}

	safePkgIdent := func(s string) string {
		s = safeTCIdentifierName(s)
		s = s[2:] // remove initial "tc".
		return strings.ToLower(s)
	}

	err = os.MkdirAll(params.GenPath, 0766)
	if err != nil {
		return gci, fmt.Errorf("creating dir for generated code: %w", err)
	}

	// erase anyfin currently in the old genpath; old code might make this code
	// break.
	err = clearDirContents(params.GenPath)
	if err != nil {
		return gci, fmt.Errorf("clearing dir for generated code: %w", err)
	}

	// start copying the hooks package
	hooksDestPath := filepath.Join(params.GenPath, "internal", hooksPkgName)
	hooksDone, err := copyDirToTargetAsync(params.HooksPkgDir, hooksDestPath)
	if err != nil {
		return gci, fmt.Errorf("copying hooks package: %w", err)
	}
	// start copying the format package if set and if it's not the same as the
	// hooks package.
	var formatDone <-chan error
	if separateFormatImport {
		formatDestPath := filepath.Join(params.GenPath, "internal", formatPkgName)
		formatDone, err = copyDirToTargetAsync(params.FormatPkgDir, formatDestPath)
		if err != nil {
			return gci, fmt.Errorf("copying format package: %w", err)
		}
	}

	// generate the compiler code
	fePkgPath := filepath.Join(params.GenPath, "internal", params.FrontendPkgName)
	binPkg := "github.com/dekarrin/ictiobus/langexec/" + safePkgIdent(md.Language)
	fePkgImport := filepath.ToSlash(filepath.Join(binPkg, "internal", params.FrontendPkgName))

	feOpts := params.Opts
	// if we copied the IR, we need to change the import in frontend files so
	// that they refer to the correct one.
	if irPackage == hooksPkgName {
		feIRPkg := filepath.ToSlash(filepath.Join(binPkg, "internal", hooksPkgName))
		feFQir := makeFQType(feIRPkg, irType)
		feOpts.IRType = feFQir
	}
	err = GenerateFrontendGo(spec, md, params.FrontendPkgName, fePkgPath, fePkgImport, &feOpts)
	if err != nil {
		return gci, fmt.Errorf("generating compiler: %w", err)
	}

	// since GenerateCompilerGo ensures the directory exists, we can now copy
	// the encoded parser into it as well.
	parserPath := filepath.Join(fePkgPath, "parser.cff")
	err = parse.WriteFile(params.Parser, parserPath)
	if err != nil {
		return gci, fmt.Errorf("writing parser: %w", err)
	}

	var irIsBuiltIn bool
	// only fill in the ir package import if ir's package is not the builtin
	// package and if it is not the same as the hooks package
	if irFQPackage == "builtin" {
		irFQPackage = ""
		irType = irTypeBare
		irIsBuiltIn = true
	} else if irPackage == hooksPkgName {
		irFQPackage = ""
	}

	// export template with main file
	mainFillData := cgMainData{
		BinPkg:            binPkg,
		BinName:           params.BinName,
		Version:           md.Version,
		Lang:              md.Language,
		HooksPkg:          hooksPkgName,
		HooksTableExpr:    params.HooksExpr,
		FormatPkg:         formatPkgName,
		FormatCall:        params.FormatCall,
		FrontendPkgImport: fePkgImport,
		TokenPkgName:      params.FrontendPkgName + "token",
		ImportFormatPkg:   separateFormatImport,
		FrontendPkg:       params.FrontendPkgName,
		IRTypePackage:     irFQPackage,
		IRType:            irType,
		IRIsBuiltInType:   irIsBuiltIn,
		IncludeSimulation: true,
	}
	fnMap := createFuncMap()
	renderFiles := map[string]codegenTemplate{
		ComponentMainFile: {nil, generatedMainFilename},
	}

	// initialize templates
	err = initTemplates(renderFiles, fnMap, params.Opts.TemplateFiles)
	if err != nil {
		return gci, err
	}

	// finally, render the main file
	rf := renderFiles[ComponentMainFile]
	mainFileRelPath := rf.outFile
	err = renderTemplateToFile(rf.tmpl, mainFillData, filepath.Join(params.GenPath, mainFileRelPath), params.Opts.DumpPreFormat)
	if err != nil {
		return gci, err
	}

	// wait for the hooks package to be copied; we'll need it for go mod tidy
	<-hooksDone

	// if we have a format package, wait for it to be copied
	if formatDone != nil {
		<-formatDone
	}

	// wipe any existing go module stuff
	err = os.RemoveAll(filepath.Join(params.GenPath, "go.mod"))
	if err != nil {
		return gci, fmt.Errorf("removing existing go.mod: %w", err)
	}
	err = os.RemoveAll(filepath.Join(params.GenPath, "go.sum"))
	if err != nil {
		return gci, fmt.Errorf("removing existing go.sum: %w", err)
	}
	err = os.RemoveAll(filepath.Join(params.GenPath, "vendor"))
	if err != nil {
		return gci, fmt.Errorf("removing existing vendor directory: %w", err)
	}

	// shell out to run go module stuff
	shell := shellout.Shell{Dir: params.GenPath, Env: os.Environ()}
	goModInitOutput, err := shell.Exec("go", "mod", "init", mainFillData.BinPkg)
	if err != nil {
		return gci, fmt.Errorf("initializing generated module with binary: %w\n%s", err, goModInitOutput)
	}

	if params.LocalIctiobusSource != "" {
		// make shore we use the latest version of ictiobus in the generated code
		goRepOutput, err := shell.Exec("go", "mod", "edit", "-replace", "github.com/dekarrin/ictiobus="+params.LocalIctiobusSource)
		if err != nil {
			return gci, fmt.Errorf("replacing ictiobus with local source: %w\n%s", err, goRepOutput)
		}
	}

	goModTidyOutput, err := shell.Exec("go", "mod", "tidy")
	if err != nil {
		return gci, fmt.Errorf("tidying generated module with binary: %w\n%s", err, goModTidyOutput)
	}

	// if we got here, all output has been written to the temp dir.
	gci.Path = params.GenPath
	gci.MainFile = mainFileRelPath

	return gci, nil
}

func clearDirContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// GenerateFrontendGo generates the source code for a compiler frontend that can
// handle a fishi spec. The source code is placed in the given directory. This
// does *not* copy the hooks package, it only outputs the frontend code.
//
// If opts is nil, the default options will be used.
func GenerateFrontendGo(spec Spec, md SpecMetadata, pkgName, pkgDir string, pkgImport string, opts *CodegenOptions) error {
	if len(spec.Tokens) == 0 {
		return fmt.Errorf("spec defines no tokens")
	}
	if len(spec.Patterns) == 0 {
		return fmt.Errorf("spec defines no lexer patterns")
	}
	if len(spec.Grammar.NonTerminals()) == 0 {
		return fmt.Errorf("spec defines no grammar rules")
	}
	if len(spec.TranslationScheme) == 0 {
		return fmt.Errorf("spec defines no translation scheme actions")
	}

	if opts == nil {
		opts = &CodegenOptions{}
	}

	data := createTemplateFillData(spec, md, pkgName, opts.IRType, pkgImport)

	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		return fmt.Errorf("creating target dir: %w", err)
	}

	err = os.MkdirAll(filepath.Join(pkgDir, pkgName+"token"), 0755)
	if err != nil {
		return fmt.Errorf("creating target dir: %w", err)
	}

	fnMap := createFuncMap()

	renderFiles := map[string]codegenTemplate{
		ComponentTokens:   {nil, filepath.Join(pkgName+"token", generatedTokensFilename)},
		ComponentLexer:    {nil, generatedLexerFilename},
		ComponentParser:   {nil, generatedParserFilename},
		ComponentSDTS:     {nil, generatedSDTSFilename},
		ComponentFrontend: {nil, generatedFrontendFilename},
	}

	// initialize templates
	err = initTemplates(renderFiles, fnMap, opts.TemplateFiles)
	if err != nil {
		return err
	}

	// finally, render the template files
	for _, comp := range codegenOrder {
		rf, ok := renderFiles[comp]
		if !ok {
			continue
		}

		err = renderTemplateToFile(rf.tmpl, data, filepath.Join(pkgDir, rf.outFile), opts.DumpPreFormat)
		if err != nil {
			return err
		}
	}

	return nil

}

func createFuncMap() template.FuncMap {
	return template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"upperCamel": safeTCIdentifierName,
		"quote": func(s string) string {
			return fmt.Sprintf("%q", s)
		},
		"rquote": func(s string) string {
			s = strings.ReplaceAll(s, "`", "` + \"`\" + `")
			return fmt.Sprintf("`%s`", s)
		},
		"title": func(s string) string {
			return titleCaser.String(s)
		},
		"with_article": func(def bool, s string) string {
			return textfmt.ArticleFor(s, def) + " " + s
		},
	}
}

func initTemplates(renderFiles map[string]codegenTemplate, fnMap template.FuncMap, customTemplateFiles map[string]string) error {
	// initialize the templates and parse the template for each, either from the
	// default embedded template string or from the specified custom template
	// file on disk.
	for _, comp := range codegenOrder {
		var err error

		rf, ok := renderFiles[comp]
		if !ok {
			continue
		}

		rf.tmpl = template.New(comp).Funcs(fnMap)

		if customTemplatePath, ok := customTemplateFiles[comp]; ok {
			// custom template file specified, load it
			fileBasename := filepath.Base(customTemplatePath)

			// avoid use of Template.ParseFiles because according to docs it
			// relies on the template having the same name as the basename of at
			// least one of the files, which is not going to be the case here.
			templateBytes, err := os.ReadFile(customTemplatePath)
			if err != nil {
				return fmt.Errorf("loading custom %s template %s: %w", comp, fileBasename, err)
			}

			templateStr := string(templateBytes)

			_, err = rf.tmpl.Parse(templateStr)
			if err != nil {
				return fmt.Errorf("parsing custom %s template %s: %w", comp, fileBasename, err)
			}
		} else {
			// use default embedded template string
			_, err = rf.tmpl.Parse(defaultTemplates[comp])
			if err != nil {
				return fmt.Errorf("parsing default %s template: %w", comp, err)
			}
		}

		renderFiles[comp] = rf
	}

	return nil
}

func renderTemplateToFile(tmpl *template.Template, data interface{}, dest string, dumpPreFormat bool) error {
	var tokBuf bytes.Buffer
	if err := tmpl.Execute(&tokBuf, data); err != nil {
		return fmt.Errorf("generating %s file: %w", tmpl.Name(), err)
	}
	if dumpPreFormat {
		fmt.Printf("\n=== %s ===\n", dest)
		preFmt := rosed.Edit(tokBuf.String()).
			Apply(func(idx int, line string) []string {
				// add line numbers
				return []string{fmt.Sprintf("%4d: %s", idx+1, line)}
			}).
			String()
		fmt.Printf("%s\n", preFmt)
	}
	formatted, err := format.Source(tokBuf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting %s file: %w", tmpl.Name(), err)
	}
	// write the file out
	err = os.WriteFile(dest, formatted, 0666)
	if err != nil {
		return fmt.Errorf("writing %s file: %w", tmpl.Name(), err)
	}
	return nil
}

func createTemplateFillData(spec Spec, md SpecMetadata, pkgName string, fqIRType string, fePkgImport string) cgData {
	// fill initial from metadata
	data := cgData{
		FrontendPackage:   pkgName,
		TokenPkgName:      pkgName + "token",
		FrontendPkgImport: fePkgImport,
		Lang:              md.Language,
		Version:           md.Version,
		Command:           CommandName,
		CommandArgs:       md.InvocationArgs,
	}

	// if IR type is specified, use it
	if fqIRType != "" {
		irPkg, irType, _, irTypeBare, err := parseFQType(fqIRType)
		if err == nil {
			if irPkg == "builtin" {
				// if it's a builtin type there's no need to do the import and
				// we must also strip the package name from the type
				data.IRType = irTypeBare
			} else {
				data.IRPackage = irPkg
				data.IRType = irType
			}
		}
	}

	// fill classes (also save their cgClass)

	tokCgClasses := map[string]cgClass{}
	for _, class := range spec.Tokens {
		varName := safeTCIdentifierName(class.ID())

		classData := cgClass{
			Name:  varName,
			ID:    class.ID(),
			Human: class.Human(),
		}

		tokCgClasses[class.ID()] = classData

		data.Classes = append(data.Classes, classData)
	}

	// fill patterns

	tokMap := spec.ClassMap()

	for _, state := range textfmt.OrderedKeys(spec.Patterns) {
		cgStateData := cgStatePatterns{State: state}
		statePats := spec.Patterns[state]

		seenToks := box.NewStringSet()
		for _, pat := range statePats {
			entry := cgPatternEntry{
				Priority: pat.Priority,
				Regex:    pat.Regex.String(),
			}

			// figure out our action string
			switch pat.Action.Type {
			case lex.ActionScan:
				tokData := tokCgClasses[pat.Action.ClassID]
				entry.Action = fmt.Sprintf("lex.LexAs(%s.%s.ID())", data.TokenPkgName, tokData.Name)
			case lex.ActionScanAndState:
				tokData := tokCgClasses[pat.Action.ClassID]
				entry.Action = fmt.Sprintf("lex.LexAndSwapState(%s.%s.ID(), %q)", data.TokenPkgName, tokData.Name, pat.Action.State)
			case lex.ActionState:
				entry.Action = fmt.Sprintf("lex.SwapState(%q)", pat.Action.State)
			case lex.ActionNone:
				entry.Action = "lex.Discard()"
			}

			// register any token class used in the pattern
			if pat.Action.Type == lex.ActionScan || pat.Action.Type == lex.ActionScanAndState {
				if !seenToks.Has(pat.Action.ClassID) {
					tok := tokMap[pat.Action.ClassID]
					tokData := tokCgClasses[tok.ID()]
					seenToks.Add(tok.ID())
					cgStateData.Classes = append(cgStateData.Classes, tokData)
				}
			}

			cgStateData.Entries = append(cgStateData.Entries, entry)
		}

		if state == "" {
			data.Patterns.DefaultState = cgStateData
		} else {
			data.Patterns.NonDefaultStates = append(data.Patterns.NonDefaultStates, cgStateData)
		}
	}

	// fill rules

	nts := spec.Grammar.NonTerminalsByPriority()
	for _, nt := range nts {
		gRule := spec.Grammar.Rule(nt)
		rData := cgRule{Head: gRule.NonTerminal}
		for _, p := range gRule.Productions {
			pData := cgGramProd{}
			for _, sym := range p {
				pData.Symbols = append(pData.Symbols, sym)
			}
			rData.Productions = append(rData.Productions, pData)
		}

		data.Rules = append(data.Rules, rData)
	}

	// fill bindings

	// group all the bindings for node with same head together
	bindingData := map[string]*cgBinding{}
	// ... but also, preserve the order of the bindings
	bindingOrder := []*cgBinding{}
	for _, sdd := range spec.TranslationScheme {
		bData, ok := bindingData[sdd.Rule.NonTerminal]
		if !ok {
			bData = &cgBinding{
				Head: sdd.Rule.NonTerminal,
			}
			bindingData[sdd.Rule.NonTerminal] = bData
			bindingOrder = append(bindingOrder, bData)
		}

		sdtsData := cgSDTSProd{
			Attribute:   sdd.Attribute.Name,
			Hook:        sdd.Hook,
			Synthetic:   sdd.Attribute.Rel.Type == trans.RelHead,
			ForRelType:  sdd.Attribute.Rel.Type.GoString(),
			ForRelIndex: sdd.Attribute.Rel.Index,
		}

		// fill symbols from the only production
		for _, sym := range sdd.Rule.Productions[0] {
			sdtsData.Symbols = append(sdtsData.Symbols, sym)
		}

		// fill args
		for _, arg := range sdd.Args {
			argData := cgArg{
				RelType:   arg.Rel.Type.GoString(),
				RelIndex:  arg.Rel.Index,
				Attribute: arg.Name,
			}
			sdtsData.Args = append(sdtsData.Args, argData)
		}

		bData.Productions = append(bData.Productions, sdtsData)
	}
	// now add all the bindings to the data in order
	// (first one is the default IR attribute)
	for _, b := range bindingOrder {
		data.Bindings = append(data.Bindings, *b)
		if data.IRAttribute == "" {
			data.IRAttribute = b.Productions[0].Attribute
		}
	}

	// done, return finished data
	return data
}

func safeTCIdentifierName(str string) string {
	nameRunes := []rune{}

	getSymbolName := func(ch rune) string {
		// can we get a symbol name?
		chName := runenames.Name(ch)

		chName = unsafeRunenameChars.ReplaceAllString(chName, "_")

		// how many words is the rune? you get up to 3.
		words := strings.Split(chName, " ")
		if len(words) > 3 {
			// we only want the first 3 words
			words = words[:3]
		}
		chName = strings.Join(words, "_")

		return chName
	}

	for _, ch := range str {
		if ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z') || ('0' <= ch && ch <= '9') || ch == '_' {
			nameRunes = append(nameRunes, ch)
		} else if ch == '-' || unicode.IsSpace(ch) {
			nameRunes = append(nameRunes, '_')
		} else {
			chName := getSymbolName(ch)
			nameRunes = append(nameRunes, []rune(chName)...)
		}
	}

	// collapse all runs of underscores
	name := underscoreCollapser.ReplaceAllString(string(nameRunes), "_")
	// trim leading and trailing underscores
	name = strings.Trim(name, "_")

	// did we just get an empty string? if so, go back and actually do symbol
	// names. should only happen if it consists only of "-" and whitespace
	// chars
	if name == "" {
		newNameRunes := []rune{}

		for _, ch := range str {
			chName := getSymbolName(ch)
			newNameRunes = append(newNameRunes, []rune(chName)...)
		}

		name = string(newNameRunes)
	}

	fullName := "TC"
	// split by underscores and do a title case on each word
	words := strings.Split(name, "_")
	for _, word := range words {
		fullName += string(titleCaser.String(word))
	}

	if fullName == "TC" {
		panic("assertion failed: generated token name has no content besides 'TC'")
	}

	return fullName
}

// asynchronounsly copy the package to the target directory. returns non-nil
// error if scanning the package and directory creation was successful. later,
// pushes the first error that occurs while copying the contents of a file to
// the channel, or nil to the channel if the copy was successful.
func copyDirToTargetAsync(srcDir string, targetDir string) (copyResult chan error, err error) {
	// read the list of files in the source directory
	srcFiles, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("reading source dir %q: %w", srcDir, err)
	}

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("creating target dir: %w", err)
	}

	ch := make(chan error)
	go func() {
		var recursions []chan error

		for _, dirEntry := range srcFiles {
			file := filepath.Join(srcDir, dirEntry.Name())
			if dirEntry.IsDir() {
				// do it recursively.
				recursion, err := copyDirToTargetAsync(file, filepath.Join(targetDir, dirEntry.Name()))
				if err != nil {
					ch <- err
					return
				}

				recursions = append(recursions, recursion)
			} else {
				fileData, err := os.ReadFile(file)
				if err != nil {
					ch <- fmt.Errorf("reading source file %s: %w", file, err)
					return
				}

				dest := filepath.Join(targetDir, dirEntry.Name())

				// write the file to the dest path
				err = os.WriteFile(dest, fileData, 0644)
				if err != nil {
					ch <- fmt.Errorf("writing source file %s: %w", dest, err)
					return
				}
			}
		}

		// block until all the recursions are done
		for _, recursion := range recursions {
			err := <-recursion
			if err != nil {
				ch <- err
				return
			}
		}

		ch <- nil
	}()

	return ch, nil
}

func readPackageName(dir string) (string, error) {
	// what is the name of our hooks package? find out by reading the first go
	// file in the package.
	dirItems, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var pkgName string
	for _, item := range dirItems {
		if !item.IsDir() && strings.ToLower(filepath.Ext(item.Name())) == ".go" {
			// read the file to find the package name
			goFilePath := filepath.Join(dir, item.Name())
			goFile, err := os.Open(goFilePath)
			if err != nil {
				return "", err
			}

			// buffered reading
			r := bufio.NewReader(goFile)

			// now find the package name in the file
			for pkgName == "" {
				str, err := r.ReadString('\n')
				strTrimmed := strings.TrimSpace(str)

				// is it a line starting with "package"?
				if strings.HasPrefix(strTrimmed, "package") {
					lineItems := strings.Split(strTrimmed, " ")
					if len(lineItems) == 2 {
						pkgName = lineItems[1]
						break
					}
				}

				// ofc if err is somefin else
				if err != nil {
					if err == io.EOF {
						break
					}
					return "", err
				}
			}
		}

		if pkgName != "" {
			break
		}
	}
	if pkgName == "" {
		return "", fmt.Errorf("could not find package name; make sure files are gofmt'd")
	}

	return pkgName, nil
}
