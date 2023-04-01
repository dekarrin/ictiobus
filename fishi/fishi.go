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
	ParserCFF         string
	ReadCache         bool
	WriteCache        bool
	SDTSValidate      bool
	SDTSValShowTrees  bool
	SDTSValShowGraphs bool
	SDTSValAllTrees   bool
	SDTSValSkipTrees  int
	LexerTrace        bool
	ParserTrace       bool
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
func GetFrontend(opts Options) (ictiobus.Frontend[[]syntax.ASTBlock], error) {
	// check for preload
	var preloadedParser ictiobus.Parser
	if opts.ParserCFF != "" && opts.ReadCache {
		var err error
		preloadedParser, err = ictiobus.GetParserFromDisk(opts.ParserCFF)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				preloadedParser = nil
			} else {
				return ictiobus.Frontend[[]syntax.ASTBlock]{}, fmt.Errorf("loading cachefile %q: %w", opts.ParserCFF, err)
			}
		}
	}

	feOpts := fe.FrontendOptions{
		LexerTrace:  opts.LexerTrace,
		ParserTrace: opts.ParserTrace,
	}

	fishiFront := fe.Frontend[[]syntax.ASTBlock](syntax.HooksTable, feOpts, preloadedParser)

	// check the parser encoding if we generated a new one:
	if preloadedParser == nil && opts.ParserCFF != "" && opts.WriteCache {
		err := ictiobus.SaveParserToDisk(fishiFront.Parser, opts.ParserCFF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writing parser to disk: %s\n", err.Error())
		} else {
			fmt.Printf("wrote parser to %q\n", opts.ParserCFF)
		}
	}

	// validate our SDTS if we were asked to
	if opts.SDTSValidate {
		valProd := fishiFront.Lexer.FakeLexemeProducer(true, "")

		di := trans.ValidationOptions{
			ParseTrees:    opts.SDTSValShowTrees,
			FullDepGraphs: opts.SDTSValShowGraphs,
			ShowAllErrors: opts.SDTSValAllTrees,
			SkipErrors:    opts.SDTSValSkipTrees,
		}

		sddErr := fishiFront.SDT.Validate(fishiFront.Parser.Grammar(), fishiFront.IRAttribute, di, valProd)
		if sddErr != nil {
			return ictiobus.Frontend[[]syntax.ASTBlock]{}, fmt.Errorf("sdd validation error: %w", sddErr)
		}
	}

	return fishiFront, nil
}
