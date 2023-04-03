package fishi

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/rosed"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/runenames"
)

// file codegen implements conversion of a fishi spec to a series of go files.
// This is the only way that we can do full validation with a hooks package.

const (
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
	GeneratedTokensFilename   = "tokens.ict.go"
	GeneratedLexerFilename    = "lexer.ict.go"
	GeneratedParserFilename   = "parser.ict.go"
	GeneratedSDTSFilename     = "sdts.ict.go"
	GeneratedFrontendFilename = "frontend.ict.go"
	GeneratedMainFilename     = "main.ict.go"
)

// Default template strings for each component of the generated compiler.
var (
	//go:embed templates/tokens.go.tmpl
	TemplateTokens string

	//go:embed templates/lexer.go.tmpl
	TemplateLexer string

	//go:embed templates/parser.go.tmpl
	TemplateParser string

	//go:embed templates/sdts.go.tmpl
	TemplateSDTS string

	//go:embed templates/frontend.go.tmpl
	TemplateFrontend string

	//go:embed templates/main.go.tmpl
	TemplateMainFile string
)

var (
	underscoreCollapser = regexp.MustCompile(`_+`)
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
		ComponentTokens:   TemplateTokens,
		ComponentLexer:    TemplateLexer,
		ComponentParser:   TemplateParser,
		ComponentSDTS:     TemplateSDTS,
		ComponentFrontend: TemplateFrontend,
		ComponentMainFile: TemplateMainFile,
	}
)

type CodegenOptions struct {
	// DumpPreFormat will dump the generated code before it is formatted.
	DumpPreFormat bool

	// TemplateFiles is a map of template names to a path to a custom template
	// file for that template. If entries are detected under the key of one of
	// the ComponentX constants, the path in it is parsed as a template file and
	// used for outputting the generated code for that component instead of the
	// default embedded template.
	TemplateFiles map[string]string

	// IRType is the fully-qualified type of the intermediate representation in
	// the frontend. This is used to make the Frontend function return a
	// specific type instead of requiring an explicit type instantiation when
	// called.
	IRType string
}

// GeneratedCodeInfo contains information about the generated code.
type GeneratedCodeInfo struct {
	// MainFile is the path to the main executable file, relative to Path.
	MainFile string

	// Path is the location of the root of the generated code.
	Path string
}

// GenerateTestCompiler generates (but does not yet run) a test compiler for the
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
func GenerateTestCompiler(spec Spec, md SpecMetadata, p ictiobus.Parser, hooksPkgDir string, hooksExpr string, opts CodegenOptions) (GeneratedCodeInfo, error) {
	if opts.IRType == "" {
		return GeneratedCodeInfo{}, fmt.Errorf("IRType must be set in options")
	}

	irFQPackage, irType, irPackage, irErr := ParseFQType(opts.IRType)
	if irErr != nil {
		return GeneratedCodeInfo{}, fmt.Errorf("parsing IRType: %w", irErr)
	}

	gci := GeneratedCodeInfo{}
	var fePkgName = "sim" + strings.ToLower(md.Language)

	// what is the name of our hooks package? find out by reading the first go
	// file in the package.
	hooksDirItems, err := os.ReadDir(hooksPkgDir)
	if err != nil {
		return gci, fmt.Errorf("reading hooks package dir: %w", err)
	}

	var hooksPkgName string
	for _, item := range hooksDirItems {
		if !item.IsDir() && strings.ToLower(filepath.Ext(item.Name())) == ".go" {
			// read the file to find the package name
			goFilePath := filepath.Join(hooksPkgDir, item.Name())
			goFile, err := os.Open(goFilePath)
			if err != nil {
				return gci, fmt.Errorf("reading go file in hooks package: %w", err)
			}

			// buffered reading
			r := bufio.NewReader(goFile)

			// now find the package name in the file
			for hooksPkgName == "" {
				str, err := r.ReadString('\n')
				strTrimmed := strings.TrimSpace(str)

				// is it a line starting with "package"?
				if strings.HasPrefix(strTrimmed, "package") {
					lineItems := strings.Split(strTrimmed, " ")
					if len(lineItems) == 2 {
						hooksPkgName = lineItems[1]
						break
					}
				}

				// ofc if err is somefin else
				if err != nil {
					if err == io.EOF {
						break
					}
					return gci, fmt.Errorf("reading go file in hooks package: %w", err)
				}
			}
		}

		if hooksPkgName != "" {
			break
		}
	}
	if hooksPkgName == "" {
		return gci, fmt.Errorf("could not find package name for hooks package; make sure files are gofmt'd")
	}
	if hooksPkgName == fePkgName {
		// double it to avoid name collision
		fePkgName += "_" + fePkgName
	}

	safePkgIdent := func(s string) string {
		s = safeTCIdentifierName(s)
		s = s[2:] // remove initial "tc".
		return strings.ToLower(s)
	}

	// create a temporary directory to save things in
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("ictcc-test-%s", safePkgIdent(md.Language)))
	if err != nil {
		return gci, fmt.Errorf("creating temp dir: %w", err)
	}

	// start copying the hooks package
	hooksDestPath := filepath.Join(tmpDir, "internal", hooksPkgName)
	hooksDone, err := copyDirToTargetAsync(hooksPkgDir, hooksDestPath)
	if err != nil {
		return gci, fmt.Errorf("copying hooks package: %w", err)
	}

	// generate the compiler code
	fePkgPath := filepath.Join(tmpDir, "internal", fePkgName)
	err = GenerateCompilerGo(spec, md, fePkgName, fePkgPath, &opts)
	if err != nil {
		return gci, fmt.Errorf("generating compiler: %w", err)
	}

	// since GenerateCompilerGo ensures the directory exists, we can now copy
	// the encoded parser into it as well.
	parserPath := filepath.Join(fePkgPath, "parser.cff")
	err = ictiobus.SaveParserToDisk(p, parserPath)
	if err != nil {
		return gci, fmt.Errorf("writing parser: %w", err)
	}

	// only fill in the ir package import if ir's package is not not the same as the hooks package
	if irPackage == hooksPkgName {
		irFQPackage = ""
	}

	// export template with main file
	mainFillData := cgMainData{
		BinPkg:            "github.com/dekarrin/ictiobus/langtest/" + strings.ToLower(safePkgIdent(md.Language)),
		BinName:           "test" + strings.ToLower(safePkgIdent(md.Language)),
		BinVersion:        "(simulator version)",
		HooksPkg:          hooksPkgName,
		HooksTableExpr:    hooksExpr,
		FrontendPkg:       fePkgName,
		IRTypePackage:     irFQPackage,
		IRType:            irType,
		IncludeSimulation: true,
	}
	fnMap := createFuncMap()
	renderFiles := map[string]codegenTemplate{
		ComponentMainFile: {nil, GeneratedMainFilename},
	}

	// initialize templates
	err = initTemplates(renderFiles, fnMap, opts.TemplateFiles)
	if err != nil {
		return gci, err
	}

	// finally, render the main file
	rf := renderFiles[ComponentMainFile]
	mainFileRelPath := rf.outFile
	err = renderTemplateToFile(rf.tmpl, mainFillData, filepath.Join(tmpDir, mainFileRelPath), opts.DumpPreFormat)
	if err != nil {
		return gci, err
	}

	// wait for the hooks package to be copied; we'll need it for go mod tidy
	<-hooksDone

	// shell out to run go module stuff
	goModInitCmd := exec.Command("go", "mod", "init", mainFillData.BinPkg)
	goModInitCmd.Env = os.Environ()
	goModInitCmd.Dir = tmpDir
	goModInitOutput, err := goModInitCmd.CombinedOutput()
	if err != nil {
		return gci, fmt.Errorf("initializing generated module with binary: %w\n%s", err, string(goModInitOutput))
	}
	goModTidyCmd := exec.Command("go", "mod", "tidy")
	goModTidyCmd.Env = os.Environ()
	goModTidyCmd.Dir = tmpDir
	goModTidyOutput, err := goModTidyCmd.CombinedOutput()
	if err != nil {
		return gci, fmt.Errorf("tidying generated module with binary: %w\n%s", err, string(goModTidyOutput))
	}

	// if we got here, all output has been written to the temp dir.
	gci.Path = tmpDir
	gci.MainFile = mainFileRelPath

	return gci, nil
}

// GenerateCompilerGo generates the source code for a compiler that can handle a
// fishi spec. The source code is placed in the given directory. This does *not*
// copy the hooks package, it only outputs the compiler code.
//
// If opts is nil, the default options will be used.
func GenerateCompilerGo(spec Spec, md SpecMetadata, pkgName, pkgDir string, opts *CodegenOptions) error {
	if opts == nil {
		opts = &CodegenOptions{}
	}

	data := createTemplateFillData(spec, md, pkgName, opts.IRType)

	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		return fmt.Errorf("creating target dir: %w", err)
	}

	fnMap := createFuncMap()

	renderFiles := map[string]codegenTemplate{
		ComponentTokens:   {nil, GeneratedTokensFilename},
		ComponentLexer:    {nil, GeneratedLexerFilename},
		ComponentParser:   {nil, GeneratedParserFilename},
		ComponentSDTS:     {nil, GeneratedSDTSFilename},
		ComponentFrontend: {nil, GeneratedFrontendFilename},
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
		"upperCamel": safeTCIdentifierName,
		"quote": func(s string) string {
			return fmt.Sprintf("%q", s)
		},
		"rquote": func(s string) string {
			s = strings.ReplaceAll(s, "`", "` + \"`\" + `")
			return fmt.Sprintf("`%s`", s)
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

			// TODO: p shore it's not actually necessary to reassign the results
			// of calling Parse(); check l8er.
			rf.tmpl, err = rf.tmpl.Parse(templateStr)
			if err != nil {
				return fmt.Errorf("parsing custom %s template %s: %w", comp, fileBasename, err)
			}
		} else {
			// use default embedded template string

			// TODO: p shore it's not actually necessary to reassign the results
			// of calling Parse(); check l8er.
			rf.tmpl, err = rf.tmpl.Parse(defaultTemplates[comp])
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

func createTemplateFillData(spec Spec, md SpecMetadata, pkgName string, fqIRType string) cgData {
	// fill initial from metadata
	data := cgData{
		FrontendPackage: pkgName,
		Lang:            md.Language,
		Version:         md.Version,
		Command:         CommandName,
		CommandArgs:     md.InvocationArgs,
	}

	// if IR type is specified, use it
	if fqIRType != "" {
		irPkg, irType, _, err := ParseFQType(fqIRType)
		if err == nil {
			data.IRPackage = irPkg
			data.IRType = irType
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
			// TODO: kind of fragile, directly putting code in as a string,
			// probably should be templated, but works for now
			switch pat.Action.Type {
			case lex.ActionScan:
				tokData := tokCgClasses[pat.Action.ClassID]
				entry.Action = fmt.Sprintf("lex.LexAs(%s.ID())", tokData.Name)
			case lex.ActionScanAndState:
				tokData := tokCgClasses[pat.Action.ClassID]
				entry.Action = fmt.Sprintf("lex.LexAndSwapState(%s.ID(), %q)", tokData.Name, pat.Action.State)
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

	nts := spec.Grammar.PriorityNonTerminals()
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
			Synthetic:   sdd.Attribute.Relation.Type == trans.RelHead,
			ForRelType:  sdd.Attribute.Relation.Type.GoString(),
			ForRelIndex: sdd.Attribute.Relation.Index,
		}

		// fill symbols from the only production
		for _, sym := range sdd.Rule.Productions[0] {
			sdtsData.Symbols = append(sdtsData.Symbols, sym)
		}

		// fill args
		for _, arg := range sdd.Args {
			argData := cgArg{
				RelType:   arg.Relation.Type.GoString(),
				RelIndex:  arg.Relation.Index,
				Attribute: arg.Name,
			}
			sdtsData.Args = append(sdtsData.Args, argData)
		}

		bData.Productions = append(bData.Productions, sdtsData)
	}
	// now add all the bindings to the data in order
	for _, b := range bindingOrder {
		data.Bindings = append(data.Bindings, *b)
	}

	// done, return finished data
	return data
}

type codegenTemplate struct {
	tmpl    *template.Template
	outFile string
}

// codegen data for template fill of main.go
type cgMainData struct {
	BinPkg            string
	BinName           string
	BinVersion        string
	HooksPkg          string
	HooksTableExpr    string
	FrontendPkg       string
	IRTypePackage     string
	IRType            string
	IncludeSimulation bool
}

// codegenData for template fill.
type cgData struct {
	FrontendPackage string
	Lang            string
	Version         string
	IRAttribute     string
	IRType          string
	IRPackage       string
	Command         string
	CommandArgs     string
	Classes         []cgClass
	Patterns        cgPatterns
	Rules           []cgRule
	Bindings        []cgBinding
}

type cgPatterns struct {
	DefaultState     cgStatePatterns
	NonDefaultStates []cgStatePatterns
}

type cgStatePatterns struct {
	State   string
	Classes []cgClass
	Entries []cgPatternEntry
}

type cgPatternEntry struct {
	Regex    string
	Action   string
	Priority int
}

type cgBinding struct {
	Head        string
	Productions []cgSDTSProd
}

type cgSDTSProd struct {
	Symbols     []string
	Attribute   string
	Hook        string
	Args        []cgArg
	Synthetic   bool
	ForRelType  string
	ForRelIndex int
}

type cgArg struct {
	RelType   string
	RelIndex  int
	Attribute string
}

type cgRule struct {
	Head        string
	Productions []cgGramProd
}

type cgGramProd struct {
	Symbols []string
}

type cgClass struct {
	Name  string
	ID    string
	Human string
}

func safeTCIdentifierName(str string) string {
	nameRunes := []rune{}

	for _, ch := range str {
		if ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z') || ('0' <= ch && ch <= '9') || ch == '_' {
			nameRunes = append(nameRunes, ch)
		} else if ch == '-' || unicode.IsSpace(ch) {
			nameRunes = append(nameRunes, '_')
		} else {
			// can we get a symbol name?
			chName := runenames.Name(ch)

			// how many words is the rune? you get up to 3.
			words := strings.Split(chName, " ")
			if len(words) > 3 {
				// we only want the first 3 words
				words = words[:3]
			}
			chName = strings.Join(words, "_")
			nameRunes = append(nameRunes, []rune(chName)...)
		}
	}

	// collapse all runs of underscores
	name := underscoreCollapser.ReplaceAllString(string(nameRunes), "_")
	// trim leading and trailing underscores
	name = strings.Trim(name, "_")

	fullName := "TC"
	// split by underscores and do a title case on each word
	words := strings.Split(name, "_")
	for _, word := range words {
		fullName += string(titleCaser.String(word))
	}

	return fullName
}

// asynchronounsly copy the package to the target directory. returns non-nil
// error if scanning the package and directory creation was successful. later,
// pushes the first error that occurs while copying the contents of a file to
// the channel, or nil to the channel if the copy was successful.
//
// TODO: this should be a go dir, not a package.
func copyDirToTargetAsync(srcDir string, targetDir string) (copyResult chan error, err error) {
	/*
		pkgs, err := packages.Load(nil, goPackage)
		if err != nil {
			return nil, fmt.Errorf("scanning package: %w", err)
		}
		if len(pkgs) != 1 {
			return nil, fmt.Errorf("expected one package, got %d", len(pkgs))
		}

		pkg := pkgs[0]

		// Permissions:
		// rwxr-xr-x = 755*/

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
