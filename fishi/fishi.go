package fishi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/types"
	"github.com/gomarkdown/markdown"
	mkast "github.com/gomarkdown/markdown/ast"
	mkparser "github.com/gomarkdown/markdown/parser"
)

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

func ReadFishiMdFile(filename string, validateSDTS bool) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = ProcessFishiMd(filename, data, validateSDTS)
	if err != nil {
		return err
	}

	return nil
}

func ProcessFishiMd(filename string, mdText []byte, validateSDTS bool) error {

	// TODO: read in filename, based on it check for cached version

	var preloadedParser ictiobus.Parser
	var cachefile string
	if filename != "" {
		filenameBaseName := filepath.Base(filename)
		filenameExt := filepath.Ext(filename)
		filenamePreExt := filenameBaseName[:len(filenameBaseName)-len(filenameExt)]
		cachefileDir := filepath.Dir(filename)
		cachefileName := filenamePreExt + "-parser.cff"
		cachefile = filepath.Join(cachefileDir, cachefileName)

		var err error
		preloadedParser, err = ictiobus.GetParserFromDisk(cachefile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				preloadedParser = nil
			} else {
				return fmt.Errorf("loading cachefile %q: %w", cachefile, err)
			}
		}
	}

	// debug steps: output source after preprocess
	// output token stream
	// output grammar constructed
	// output parser table and type

	fishiSource := GetFishiFromMarkdown(mdText)
	fishiSource = Preprocess(fishiSource)
	//fishi := bytes.NewBuffer(fishiSource)

	lx := CreateBootstrapLexer()

	var parser ictiobus.Parser
	if preloadedParser != nil {
		parser = preloadedParser

		fmt.Printf("successfully loaded %s parser\n", parser.Type().String())
	} else {
		var ambigWarns []string
		parser, ambigWarns := CreateBootstrapParser()
		for i := range ambigWarns {
			fmt.Printf("warn: ambiguous grammar: %s\n", ambigWarns[i])
		}

		fmt.Printf("successfully built %s parser\n", parser.Type().String())

		parserBytes := ictiobus.EncodeParserBytes(parser)
		_, err := ictiobus.DecodeParserBytes(parserBytes)
		if err != nil {
			fmt.Printf("FAILED TO DECODE IMMEDIATELY: %s\n", err.Error())
		}

		if cachefile != "" {
			err := ictiobus.SaveParserToDisk(parser, cachefile)
			if err != nil {
				fmt.Printf("writing parser to disk: %s\n", err.Error())
			} else {
				fmt.Printf("wrote parser to %q\n", cachefile)
			}
		}
	}

	sdd := CreateBootstrapSDD()

	// but does the thing work? grab a val producer and the grammar and find out
	valProd := lx.FakeLexemeProducer(true, "")
	g := parser.Grammar()
	if validateSDTS {
		sddErr := sdd.Validate(g, "ast", types.DebugInfo{}, valProd)
		if sddErr != nil {
			return fmt.Errorf("sdd validation error: %w", sddErr)
		}
	}

	frontEnd := ictiobus.Frontend[AST]{
		Lexer:       lx,
		Parser:      parser,
		SDT:         sdd,
		IRAttribute: "ast",
	}

	/*dfa := parser.GetDFA()
	if dfa != "" {
		fmt.Printf("%s\n", dfa)
	}*/

	// now, try to make a parse tree for your own grammar
	fishiTest := `%%actions

			%symbol


			{hey}
			%prod  %index 8

		%action {thing}.thing %hook thing
			%prod {some}

		%action {thing}.thing %hook thing
			%prod {test}

			%action {thing}.thing %hook thing
		%prod {ye}
		%action {thing}.thing %hook thing

				%symbol {yo}%prod + {EAT} ext

		%action {thing}.thing %hook thing
		%%tokens
		[somefin]

		%stateshift   someState

		%%tokens

			glub  %discard


			[some]{FREEFORM}idk[^bullshit]text\*
			%discard

			[more]b*shi{2,4}   %stateshift glub
		%token lovely %human "Something nice"
			%priority 1

		%state this

		[yo] %discard
		
		%%grammar
		%state glub
		{RULE} =   {SOMEBULLSHIT}

					%%grammar
					{RULE}=                           {WOAH} | n
					{R2}				= =+  {DAMN} cool | okaythen + 2 | {}
					                 | {SOMEFIN ELSE}

					%state someState

					{ANOTHER}=		{HMM}
	



		%%actions

		%symbol {text-element}
		%prod FREEFORM_TEXT
		%action {text-element}.str
		%hook identity  %with FREEFORM_TEXT.$text

		%prod ESCSEQ
		%action {text-element}.str
		%hook unescape  %with ESCSEQ.$test		`

	frontEnd.Debug = types.DebugInfo{}
	ast, err := frontEnd.AnalyzeString(fishiTest)
	if err != nil {
		return err
	}

	fmt.Printf("AST read from data:\n")
	fmt.Printf(ast.String() + "\n")

	return nil
}
