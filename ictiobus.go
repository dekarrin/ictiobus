// Package ictiobus contains parsers and parser-generator constructs used as
// part of research into compiling techniques. It is the tunascript compilers
// pulled out after it turned from "small knowledge gaining side-side project"
// into a full-blown compilers and translators learning and research project.
//
// It's based off of the name for the buffalo fish due to the buffalo's relation
// with bison. Naturally, bison due to its popularity as a parser-generator
// tool.
//
// This will probably never be as good as bison, so consider using that. This is
// for research and does not seek to replace existing toolchains in any
// practical fashion.
package ictiobus

// DEV NOTE:
//
// https://jsmachines.sourceforge.net/machines/lalr1.html is an AMAZING tool for
// validating LALR(1) grammars quickly.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/ictiobus/trans"
	"github.com/dekarrin/ictiobus/types"
)

// NewLexer returns a lexer whose Lex method will immediately lex the entire
// input source, finding errors and reporting them and stopping as soon as the
// first lexing error is encountered or the input has been completely lexed.
//
// The TokenStream returned by the Lex function is guaranteed to not have any
// error tokens.
func NewLexer() lex.Lexer {
	return lex.NewLexer(false)
}

// NewLazyLexer returns a Lexer whose Lex method will return a TokenStream that
// is lazily executed; that is to say, calling Next() on the token stream will
// perform only enough lexical analysis to produce the next token. Additionally,
// that TokenStream may produce an error token, which parsers would need to
// handle appropriately.
func NewLazyLexer() lex.Lexer {
	return lex.NewLexer(true)
}

// NewParser returns what is the most flexible and efficient parser in this
// package that can parse the given grammar. The following parsers will be
// attempted to be built, in order, with each subsequent one attempted after the
// prior one fails.
//
// * LALR(1)
// * CLR(1)
// * SLR(1)
// * LL(1)
//
// Returns an error if no parser can be generated for the given grammar.
//
// allowAmbiguous allows the use of ambiguous grammars in LR parsers. It has no
// effect on LL(1) parser generation; LL(1) grammars must be unambiguous.
func NewParser(g grammar.Grammar, allowAmbiguous bool) (parser parse.Parser, ambigWarns []string, err error) {
	parser, ambigWarns, err = NewLALR1Parser(g, allowAmbiguous)
	if err != nil {
		bigParseGenErr := fmt.Sprintf("LALR(1) generation: %s", err.Error())
		// okay, what about a CLR(1) parser? (though, if LALR doesnt work, dont think CLR will)
		parser, ambigWarns, err = NewCLRParser(g, allowAmbiguous)
		if err != nil {
			bigParseGenErr += fmt.Sprintf("\nCLR(1) generation: %s", err.Error())

			// what about an SLR parser?
			parser, ambigWarns, err = NewSLRParser(g, allowAmbiguous)
			if err != nil {
				bigParseGenErr += fmt.Sprintf("\nSLR(1) generation: %s", err.Error())

				// LL?
				ambigWarns = nil
				parser, err = NewLL1Parser(g)
				if err != nil {
					bigParseGenErr += fmt.Sprintf("\nLL(1) generation: %s", err.Error())

					return nil, nil, fmt.Errorf("generating parser:\n%s", bigParseGenErr)
				}
			}
		}
	}

	return parser, ambigWarns, nil
}

// SaveParserToDisk stores the parser in a binary file format on disk.
func SaveParserToDisk(p parse.Parser, filename string) error {
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	bufWriter := bufio.NewWriter(fp)

	allBytes := EncodeParserBytes(p)
	_, err = bufWriter.Write(allBytes)
	if err != nil {
		return err
	}
	err = bufWriter.Flush()
	if err != nil {
		return err
	}

	return nil
}

// GetParserFromDisk retrieves the parser from a binary file on disk.
func GetParserFromDisk(filename string) (parse.Parser, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	p, err := DecodeParserBytes(data)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// EncodeParserBytes takes a parser and returns the encoded bytes.
func EncodeParserBytes(p parse.Parser) []byte {
	data := rezi.EncString(p.Type().String())
	data = append(data, rezi.EncBinary(p)...)
	return data
}

// DecodeParserBytes takes bytes and returns the Parser encoded within it.
func DecodeParserBytes(data []byte) (p parse.Parser, err error) {
	// first get the string giving the type
	typeStr, n, err := rezi.DecString(data)
	if err != nil {
		return nil, fmt.Errorf("read parser type: %w", err)
	}
	parserType, err := types.ParseParserType(typeStr)
	if err != nil {
		return nil, fmt.Errorf("decode parser type: %w", err)
	}

	// set var's concrete type by getting an empty copy
	switch parserType {
	case types.ParserLL1:
		p = parse.EmptyLL1Parser()
	case types.ParserSLR1:
		p = parse.EmptySLR1Parser()
	case types.ParserLALR1:
		p = parse.EmptyLALR1Parser()
	case types.ParserCLR1:
		p = parse.EmptyCLR1Parser()
	default:
		panic("should never happen: parsed parserType is not valid")
	}

	_, err = rezi.DecBinary(data[n:], p)
	return p, err
}

// NewLALR1Parser returns an LALR(1) parser that can generate parse trees for
// the given grammar. Returns an error if the grammar is not LALR(1).
func NewLALR1Parser(g grammar.Grammar, allowAmbiguous bool) (parser parse.Parser, ambigWarns []string, err error) {
	return parse.GenerateLALR1Parser(g, allowAmbiguous)
}

// NewSLRParser returns an SLR(1) parser that can generate parse trees for the
// given grammar. Returns an error if the grammar is not SLR(1).
func NewSLRParser(g grammar.Grammar, allowAmbiguous bool) (parser parse.Parser, ambigWarns []string, err error) {
	return parse.GenerateSLR1Parser(g, allowAmbiguous)
}

// NewLL1Parser returns an LL(1) parser that can generate parse trees for the
// given grammar. Returns an error if the grammar is not LL(1).
func NewLL1Parser(g grammar.Grammar) (parser parse.Parser, err error) {
	return parse.GenerateLL1Parser(g)
}

// NewCLRParser returns a canonical-LR(0) parser that can generate parse trees
// for the given grammar. Returns an error if the grammar is not CLR(1)
func NewCLRParser(g grammar.Grammar, allowAmbiguous bool) (parser parse.Parser, ambigWarns []string, err error) {
	return parse.GenerateCLR1Parser(g, allowAmbiguous)
}

// NewSDTS returns a new Syntax-Directed Translation Scheme.
func NewSDTS() trans.SDTS {
	return trans.NewSDTS()
}

// Frontend is a complete input-to-intermediate representation compiler
// front-end.
type Frontend[E any] struct {
	Lexer       lex.Lexer
	Parser      parse.Parser
	SDT         trans.SDTS
	IRAttribute string

	// Language is the name of the langauge that the frontend is for. It must be
	// set by the user.
	Language string

	// Version is the version of the frontend for the language. It must be set
	// by the user. This does not necessarily indicate the version of the
	// language itself; its possible that a later frontend may result in the
	// exact same semantics and syntax of the language whilst using a different
	// grammar.
	Version string
}

// AnalyzeString is the same as Analyze but accepts a string as input. It simply
// creates a Reader on s and passes it to Analyze; this method is provided for
// convenience.
//
// The parse tree may be valid even if there is an error, in which case pt will
// be non-nil.
func (fe *Frontend[E]) AnalyzeString(s string) (ir E, pt *parse.ParseTree, err error) {
	r := strings.NewReader(s)
	return fe.Analyze(r)
}

// Analyze takes the text in reader r and performs the phases necessary to
// produce an intermediate representation of it. First, in the lexical analysis
// phase, it lexes the input read from r to produce a stream of tokens. This
// stream is consumed by the syntactic analysis phase to produce a parse tree.
// Finally, in the semantic analysis phase, the actions of the syntax-directed
// translation scheme are applied to the parse tree to produce the final
// intermediate representation.
//
// If there is a problem with the input, it will be returned in a SyntaxError
// containing information about the location where it occured in the source text
// s.
//
// The parse tree may be valid even if there is an error, in which case pt will
// be non-nil.
func (fe *Frontend[E]) Analyze(r io.Reader) (ir E, pt *parse.ParseTree, err error) {
	// lexical analysis
	tokStream, err := fe.Lexer.Lex(r)
	if err != nil {
		return ir, nil, err
	}

	// syntactic analysis
	parseTree, err := fe.Parser.Parse(tokStream)
	if err != nil {
		return ir, &parseTree, err
	}

	// semantic analysis (discard warns at this stage)
	attrVals, _, err := fe.SDT.Evaluate(parseTree, fe.IRAttribute)
	if err != nil {
		return ir, &parseTree, err
	}

	// all analysis complete, now retrieve the result
	if len(attrVals) != 1 {
		return ir, &parseTree, fmt.Errorf("requested final IR attribute %q from root node but got %d values back", fe.IRAttribute, len(attrVals))
	}
	irUncast := attrVals[0]
	var ok bool
	ir, ok = irUncast.(E)
	if !ok {
		// type mismatch; use reflections to collect type for err reporting
		return ir, &parseTree, fmt.Errorf("expected final IR attribute %q to be of type %T at the root node, but result was of type %T", fe.IRAttribute, ir, irUncast)
	}

	return ir, &parseTree, nil
}
