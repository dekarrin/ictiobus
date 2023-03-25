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
	"github.com/dekarrin/ictiobus/types"
	"github.com/gomarkdown/markdown"
	mkast "github.com/gomarkdown/markdown/ast"
	mkparser "github.com/gomarkdown/markdown/parser"
)

type Results struct {
	AST  *AST
	Tree *types.ParseTree
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

func ExecuteMarkdownFile(filename string, useCache bool) (Results, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Results{}, err
	}

	res, err := ExecuteMarkdown(data, useCache)
	if err != nil {
		return res, err
	}

	return res, nil
}

func ExecuteMarkdown(mdText []byte, useCache bool) (Results, error) {

	// TODO: read in filename, based on it check for cached version

	// debug steps: output source after preprocess
	// output token stream
	// output grammar constructed
	// output parser table and type

	source := GetFishiFromMarkdown(mdText)
	return Execute(source, "fishi-parser.cff", useCache)
}

// Execute executes the fishi source code provided.
func Execute(source []byte, compiledParserFilename string, useCache bool) (Results, error) {
	// get the frontend
	fishiFront, err := GetFrontend(compiledParserFilename, true, useCache)
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

// GetFrontend gets the frontend for the fishi compiler-compiler. If cffFile is
// provided, it is used to load the cached parser from disk. Otherwise, a new
// frontend is created.
func GetFrontend(cffFile string, validateSDTS bool, useCff bool) (ictiobus.Frontend[AST], error) {
	// check for preload
	var preloadedParser ictiobus.Parser
	if cffFile != "" && useCff {
		var err error
		preloadedParser, err = ictiobus.GetParserFromDisk(cffFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				preloadedParser = nil
			} else {
				return ictiobus.Frontend[AST]{}, fmt.Errorf("loading cachefile %q: %w", cffFile, err)
			}
		}
	}

	fishiFront := Frontend(preloadedParser)

	// check the parser encoding if we generated a new one:
	if preloadedParser == nil {
		parserBytes := ictiobus.EncodeParserBytes(fishiFront.Parser)
		_, err := ictiobus.DecodeParserBytes(parserBytes)
		if err != nil {
			fmt.Printf("FAILED TO DECODE IMMEDIATELY: %s\n", err.Error())
		}

		if cffFile != "" {
			err := ictiobus.SaveParserToDisk(fishiFront.Parser, cffFile)
			if err != nil {
				fmt.Printf("writing parser to disk: %s\n", err.Error())
			} else {
				fmt.Printf("wrote parser to %q\n", cffFile)
			}
		}
	}

	// validate our SDTS if we were asked to
	if validateSDTS {
		valProd := fishiFront.Lexer.FakeLexemeProducer(true, "")
		sddErr := fishiFront.SDT.Validate(fishiFront.Parser.Grammar(), fishiFront.IRAttribute, types.DebugInfo{}, valProd)
		if sddErr != nil {
			return ictiobus.Frontend[AST]{}, fmt.Errorf("sdd validation error: %w", sddErr)
		}
	}

	return fishiFront, nil
}
