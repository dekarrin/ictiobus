package fishi

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/ictiobus/types"
	"github.com/dekarrin/rosed"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/runenames"
	"golang.org/x/tools/go/packages"
)

// file codegen implements conversion of a fishi spec to a series of go files.
// This is the only way that we can do full validation with a hooks package.

const (
	CommandName = "ictcc"

	GeneratedTokensFilename   = "tokens.ict.go"
	GeneratedLexerFilename    = "lexer.ict.go"
	GeneratedParserFilename   = "parser.ict.go"
	GeneratedSDTSFilename     = "sdts.ict.go"
	GeneratedFrontendFilename = "frontend.ict.go"
)

var (
	underscoreCollapser = regexp.MustCompile(`_+`)
	titleCaser          = cases.Title(language.AmericanEnglish)

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
)

// GenerateCompilerGo generates the source code for a compiler that can handle a
// fishi spec. The source code is placed in the given directory. This does *not*
// copy the hooks package, it only outputs the compiler code.
func GenerateCompilerGo(spec Spec, md SpecMetadata, pkgName, pkgDir string, dumpPreFormat bool) error {
	data := createTemplateFillData(spec, md, pkgName)

	err := os.MkdirAll(pkgDir, 0755)
	if err != nil {
		return fmt.Errorf("creating target dir: %w", err)
	}

	fnMap := template.FuncMap{
		"upperCamel": func(s string) string {
			words := strings.Split(s, "_")
			upperCamel := ""
			for _, word := range words {
				upperCamel += string(titleCaser.String(word))
			}
			return upperCamel
		},
		"quote": func(s string) string {
			return fmt.Sprintf("%q", s)
		},
	}

	renderFiles := []struct {
		name     string
		tmpl     string
		filename string
	}{
		{"tokens", TemplateTokens, GeneratedTokensFilename},
		{"lexer", TemplateLexer, GeneratedLexerFilename},
		{"parser", TemplateParser, GeneratedParserFilename},
		{"sdts", TemplateSDTS, GeneratedSDTSFilename},
		{"frontend", TemplateFrontend, GeneratedFrontendFilename},
	}

	// render the template files
	for _, rf := range renderFiles {
		err = renderTemplateToFile(rf.name, rf.tmpl, fnMap, data, filepath.Join(pkgDir, rf.filename), dumpPreFormat)
		if err != nil {
			return err
		}
	}

	return nil

}

func renderTemplateToFile(name, tmpl string, fns template.FuncMap, data interface{}, dest string, dumpPreFormat bool) error {
	tokTemp, err := template.New(name).Funcs(fns).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing %s template: %w", name, err)
	}

	var tokBuf bytes.Buffer
	if err := tokTemp.Execute(&tokBuf, data); err != nil {
		return fmt.Errorf("generating %s file: %w", name, err)
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
		return fmt.Errorf("formatting %s file: %w", name, err)
	}
	// write the file out
	err = os.WriteFile(dest, formatted, 0666)
	if err != nil {
		return fmt.Errorf("writing %s file: %w", name, err)
	}
	return nil
}

func createTemplateFillData(spec Spec, md SpecMetadata, pkgName string) cgData {
	// fill initial from metadata

	data := cgData{
		FrontendPackage: pkgName,
		Lang:            md.Language,
		Version:         md.Version,
		Command:         CommandName,
		CommandArgs:     md.InvocationArgs,
	}

	// fill classes (also save their cgClass)

	tokCgClasses := map[string]cgClass{}
	for _, class := range spec.Tokens {
		varName := tokenClassVarName(class)

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
	for state := range spec.Patterns {
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

// codegenData for template fill.
type cgData struct {
	FrontendPackage string
	Lang            string
	Version         string
	IRAttribute     string
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

func tokenClassVarName(class types.TokenClass) string {
	nameRunes := []rune{}

	for _, ch := range class.ID() {
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
func copyPackageToTargetAsync(goPackage string, targetDir string) (copyResult chan error, err error) {
	pkgs, err := packages.Load(nil, goPackage)
	if err != nil {
		return nil, fmt.Errorf("scanning package: %w", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected one package, got %d", len(pkgs))
	}

	pkg := pkgs[0]

	// Permissions:
	// rwxr-xr-x = 755

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("creating target dir: %w", err)
	}

	ch := make(chan error)
	go func() {
		for _, file := range pkg.GoFiles {
			baseFilename := filepath.Base(file)

			fileData, err := os.ReadFile(file)
			if err != nil {
				ch <- fmt.Errorf("reading source file %s: %w", baseFilename, err)
				return
			}

			dest := filepath.Join(targetDir, baseFilename)

			// write the file to the dest path
			err = os.WriteFile(dest, fileData, 0644)
			if err != nil {
				ch <- fmt.Errorf("writing source file %s: %w", baseFilename, err)
				return
			}
		}

		ch <- nil
	}()

	return ch, nil
}
