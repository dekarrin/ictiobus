package fishi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/fishi/fe"
	"github.com/dekarrin/ictiobus/fishi/syntax"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/ictiobus/types"
	"github.com/gomarkdown/markdown"
	mkast "github.com/gomarkdown/markdown/ast"
	mkparser "github.com/gomarkdown/markdown/parser"
)

type Results struct {
	AST  *AST
	Tree *types.ParseTree
}

type Options struct {
	ParserCFF   string
	ReadCache   bool
	WriteCache  bool
	LexerTrace  bool
	ParserTrace bool
}

// ValidateSimulatedInput generates a lightweight compiler with the spec'd
// frontend in a special directory (".sim" in the local directory) and then runs
// SDTS validation on a variety of parse tree inputs designed to cover all the
// productions of the grammar at least once.
//
// If running validation with the test compiler succeeds, it and the directory
// it was generated in are deleted. If it fails, the directory is left in place
// for inspection.
//
// IRType is required to be set in cgOpts.
//
// valOpts is not required to be set, and if nil will be treated as if it were
// set to an empty struct.
//
// No binary is generated as part of this, but source is which is then executed.
// If PreserveBinarySource is set in cgOpts, the source will be left in the
// .sim directory.
func ValidateSimulatedInput(spec Spec, md SpecMetadata, p ictiobus.Parser, hooks, hooksTable string, cgOpts CodegenOptions, valOpts *trans.ValidationOptions) error {
	pkgName := "sim" + strings.ToLower(md.Language)

	binName := safeTCIdentifierName(md.Language)
	binName = binName[2:] // remove initial "tc".
	binName = strings.ToLower(binName)
	binName = "test" + binName
	genInfo, err := GenerateBinaryMainGo(spec, md, p, hooks, hooksTable, pkgName, ".sim", binName, cgOpts)
	if err != nil {
		return fmt.Errorf("generate test compiler: %w", err)
	}

	err = ExecuteTestCompiler(genInfo, valOpts)
	if err != nil {
		return fmt.Errorf("execute test compiler: %w", err)
	}

	if !cgOpts.PreserveBinarySource {
		// if we got here, no errors. delete the test compiler and its directory
		err = os.RemoveAll(genInfo.Path)
		if err != nil {
			return fmt.Errorf("remove test compiler: %w", err)
		}
	}

	return nil
}

func GetFishiFromMarkdown(mdText []byte) []byte {
	doc := markdown.Parse(mdText, mkparser.New())
	var scanner fishiScanner
	fishi := markdown.Render(doc, scanner)
	return fishi
}

// Preprocess does a preprocess step on the source, which as of now includes
// stripping comments and normalizing end of lines to \n.
func Preprocess(source []byte) []byte {
	toBuf := make([]byte, len(source))
	copy(toBuf, source)
	scanner := bufio.NewScanner(bytes.NewBuffer(toBuf))
	var preprocessed strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "\r\n") || strings.HasPrefix(line, "\n\r") {
			line = line[0 : len(line)-2]
		} else {
			line = strings.TrimSuffix(line, "\n")
		}
		line, _, _ = strings.Cut(line, "#")
		preprocessed.WriteString(line)
		preprocessed.WriteRune('\n')
	}

	return []byte(preprocessed.String())
}

type fishiScanner bool

func (fs fishiScanner) RenderNode(w io.Writer, node mkast.Node, entering bool) mkast.WalkStatus {
	if !entering {
		return mkast.GoToNext
	}

	codeBlock, ok := node.(*mkast.CodeBlock)
	if !ok || codeBlock == nil {
		return mkast.GoToNext
	}

	if strings.ToLower(strings.TrimSpace(string(codeBlock.Info))) == "fishi" {
		w.Write(codeBlock.Literal)
	}
	return mkast.GoToNext
}

func (fs fishiScanner) RenderHeader(w io.Writer, ast mkast.Node) {}
func (fs fishiScanner) RenderFooter(w io.Writer, ast mkast.Node) {}

func ParseMarkdownFile(filename string, opts Options) (Results, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Results{}, err
	}

	res, err := ParseMarkdown(data, opts)
	if err != nil {
		return res, err
	}

	return res, nil
}

func ParseMarkdown(mdText []byte, opts Options) (Results, error) {

	// TODO: read in filename, based on it check for cached version

	// debug steps: output source after preprocess
	// output token stream
	// output grammar constructed
	// output parser table and type

	source := GetFishiFromMarkdown(mdText)
	return Parse(source, opts)
}

// Parse converts the fishi source code provided into an AST.
func Parse(source []byte, opts Options) (Results, error) {
	// get the frontend
	fishiFront, err := GetFrontend(opts)
	if err != nil {
		return Results{}, fmt.Errorf("could not get frontend: %w", err)
	}

	preprocessedSource := Preprocess(source)

	r := Results{}
	// now, try to make a parse tree
	nodes, pt, err := fishiFront.AnalyzeString(string(preprocessedSource))
	r.Tree = pt // need to do this before we return
	if err != nil {
		return r, err
	}
	r.AST = &AST{
		Nodes: nodes,
	}

	return r, nil
}

// GetFrontend gets the frontend for the fishi compiler-compiler. If cffFile is
// provided, it is used to load the cached parser from disk. Otherwise, a new
// frontend is created.
func GetFrontend(opts Options) (ictiobus.Frontend[[]syntax.Block], error) {
	// check for preload
	var preloadedParser ictiobus.Parser
	if opts.ParserCFF != "" && opts.ReadCache {
		var err error
		preloadedParser, err = ictiobus.GetParserFromDisk(opts.ParserCFF)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				preloadedParser = nil
			} else {
				return ictiobus.Frontend[[]syntax.Block]{}, fmt.Errorf("loading cachefile %q: %w", opts.ParserCFF, err)
			}
		}
	}

	feOpts := fe.FrontendOptions{
		LexerTrace:  opts.LexerTrace,
		ParserTrace: opts.ParserTrace,
	}

	fishiFront := fe.Frontend[[]syntax.Block](syntax.HooksTable, feOpts, preloadedParser)

	// check the parser encoding if we generated a new one:
	if preloadedParser == nil && opts.ParserCFF != "" && opts.WriteCache {
		err := ictiobus.SaveParserToDisk(fishiFront.Parser, opts.ParserCFF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writing parser to disk: %s\n", err.Error())
		} else {
			fmt.Printf("wrote parser to %q\n", opts.ParserCFF)
		}
	}

	return fishiFront, nil
}

// ParseFQType parses a fully-qualified type name into its package and type
// along with the name of the package. Any number of leading [] and * are
// allowed, but map types are not supported, although types with an underlying
// map type are supported.
//
// For example, ParseFQType("[]*github.com/ictiobus/fishi.Options") would return
// "github.com/ictiobus/fishi", "[]*fishi.Options", nil.
func ParseFQType(fqType string) (pkg, typeName, pkgName string, err error) {
	fqOriginal := fqType

	preType := ""
	for strings.HasPrefix(fqType, "[]") || strings.HasPrefix(fqType, "*") {
		if strings.HasPrefix(fqType, "[]") {
			preType += "[]"
			fqType = fqType[2:]
		} else {
			preType += "*"
			fqType = fqType[1:]
		}
	}
	typeParts := strings.Split(fqType, ".")
	if len(typeParts) < 2 {
		return "", "", "", fmt.Errorf("invalid fully-qualified type: %s", fqOriginal)
	}
	fqPackage := strings.Join(typeParts[:len(typeParts)-1], ".")
	pkgParts := strings.Split(fqPackage, "/")
	pkgName = pkgParts[len(pkgParts)-1]
	irType := preType + pkgName + "." + typeParts[len(typeParts)-1]

	return fqPackage, irType, pkgName, nil
}
