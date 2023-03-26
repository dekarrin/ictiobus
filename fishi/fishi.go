package fishi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

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

func ExecuteMarkdownFile(filename string, opts Options) (Results, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Results{}, err
	}

	res, err := ExecuteMarkdown(data, opts)
	if err != nil {
		return res, err
	}

	return res, nil
}

func ExecuteMarkdown(mdText []byte, opts Options) (Results, error) {

	// TODO: read in filename, based on it check for cached version

	// debug steps: output source after preprocess
	// output token stream
	// output grammar constructed
	// output parser table and type

	source := GetFishiFromMarkdown(mdText)
	return Execute(source, opts)
}

// Execute executes the fishi source code provided.
func Execute(source []byte, opts Options) (Results, error) {
	// get the frontend
	fishiFront, err := GetFrontend(opts)
	if err != nil {
		return Results{}, fmt.Errorf("could not get frontend: %w", err)
	}

	preprocessedSource := Preprocess(source)

	r := Results{}
	// now, try to make a parse tree
	ast, pt, err := fishiFront.AnalyzeString(string(preprocessedSource))
	r.Tree = pt // need to do this before we return
	if err != nil {
		return r, err
	}
	r.AST = &ast

	return r, nil
}
