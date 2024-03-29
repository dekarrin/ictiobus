package format

import (
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
	mkast "github.com/gomarkdown/markdown/ast"
	mkparser "github.com/gomarkdown/markdown/parser"
)

type fishiScanner struct{}

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

// scanMarkdownForFishiBlocks finds all FISHI codeblocks in the given
// Markdown bytes. The blocks' contents are appended together and returned as a
// single slice of bytes.
func scanMarkdownForFishiBlocks(mdText []byte) []byte {
	doc := markdown.Parse(mdText, mkparser.New())
	var scanner fishiScanner
	fishi := markdown.Render(doc, scanner)
	return fishi
}
