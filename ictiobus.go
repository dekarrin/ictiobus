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

// HACKING NOTE:
//
// https://jsmachines.sourceforge.net/machines/lalr1.html is an AMAZING tool for
// validating LALR(1) grammars quickly.

import (
	"bufio"
	"encoding"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/decbin"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/ictiobus/translation"
	"github.com/dekarrin/ictiobus/types"
)

// A Lexer represents an in-progress or ready-built lexing engine ready for use.
// It can be stored as a byte representation and retrieved from bytes as well.
type Lexer interface {

	// Lex returns a token stream. The tokens may be lexed in a lazy fashion or
	// an immediate fashion; if it is immediate, errors will be returned at that
	// point. If it is lazy, then error token productions will be returned to
	// the callers of the returned TokenStream at the point where the error
	// occured.
	Lex(input io.Reader) (types.TokenStream, error)
	RegisterClass(cl types.TokenClass, forState string)
	AddPattern(pat string, action lex.Action, forState string, priority int) error

	SetStartingState(s string)
	StartingState() string

	RegisterTokenListener(func(t types.Token))
}

// A Parser represents an in-progress or ready-built parsing engine ready for
// use. It can be stored as a byte representation and retrieved from bytes as
// well.
type Parser interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	// Parse parses input text and returns the parse tree built from it, or a
	// SyntaxError with the description of the problem.
	Parse(stream types.TokenStream) (types.ParseTree, error)

	// Type returns a string indicating what kind of parser was generated. This
	// will be "LL(1)", "SLR(1)", "CLR(1)", or "LALR(1)"
	Type() types.ParserType

	// TableString returns the parsing table as a string.
	TableString() string

	// RegisterTraceListener sets up a function to call when an event occurs.
	// The events are determined by the individual parsers but involve
	// examination of the parser stack or other critical moments that may aid in
	// debugging.
	RegisterTraceListener(func(s string))

	// GetDFA returns a string representation of the DFA for this parser, if one
	// so exists. Will return the empty string if the parser is not of the type
	// to have a DFA.
	GetDFA() string
}

// SDD is a series of syntax-directed definitions bound to syntactic rules of
// a grammar. It is used for evaluation of a parse tree into an intermediate
// representation, or for direct execution.
//
// Strictly speaking, this is closer to an Attribute grammar.
//
// It can be stored as bytes and retrieved as such as well.
type SDD interface {

	// BindInheritedAttribute creates a new SDD binding for setting the value of
	// an inherited attribute with name attrName. The production that the
	// inherited attribute is set on is specified with forProd, which must have
	// its Type set to something other than RelHead (inherited attributes can be
	// set only on production symbols).
	//
	// The binding applies only on nodes in the parse tree created by parsing
	// the grammar rule productions with head symbol head and production symbols
	// prod.
	//
	// The AttributeSetter bindFunc is called when the inherited value attrName
	// is to be set, in order to calculate the new value. Attribute values to
	// pass in as arguments are specified by passing references to the node and
	// attribute name whose value to retrieve in the withArgs slice. Explicitly
	// giving the referenced attributes in this fashion makes it easy to
	// determine the dependency graph for later execution.
	BindInheritedAttribute(head string, prod []string, attrName string, bindFunc translation.AttributeSetter, withArgs []translation.AttrRef, forProd translation.NodeRelation) error

	// BindSynthesizedAttribute creates a new SDD binding for setting the value
	// of a synthesized attribute with name attrName. The attribute is set on
	// the symbol at the head of the rule that the binding is being created for.
	//
	// The binding applies only on nodes in the parse tree created by parsing
	// the grammar rule productions with head symbol head and production symbols
	// prod.
	//
	// The AttributeSetter bindFunc is called when the synthesized value
	// attrName is to be set, in order to calculate the new value. Attribute
	// values to pass in as arguments are specified by passing references to the
	// node and attribute name whose value to retrieve in the withArgs slice.
	// Explicitly giving the referenced attributes in this fashion makes it easy
	// to determine the dependency graph for later execution.
	BindSynthesizedAttribute(head string, prod []string, attrName string, bindFunc translation.AttributeSetter, withArgs []translation.AttrRef) error

	// Bindings returns all bindings defined to apply when at a node in a parse
	// tree created by the rule production with head as its head symbol and prod
	// as its produced symbols. They will be returned in the order they were
	// defined.
	Bindings(head string, prod []string) []translation.SDDBinding

	BindingsFor(head string, prod []string, dest translation.AttrRef) []translation.SDDBinding

	// Evaluate takes a parse tree and executes the semantic actions defined as
	// SDDBindings for a node for each node in the tree and on completion,
	// returns the requested attributes values from the root node. Execution
	// order is automatically determined by taking the dependency graph of the
	// SDD; cycles are not supported. Do note that this does not require the SDD
	// to be S-attributed or L-attributed, only that it not have cycles in its
	// value dependency graph.
	Evaluate(tree types.ParseTree, attributes ...string) ([]interface{}, error)
}

// NewLexer returns a lexer whose Lex method will immediately lex the entire
// input source, finding errors and reporting them and stopping as soon as the
// first lexing error is encountered or the input has been completely lexed.
//
// The TokenStream returned by the Lex function is guaranteed to not have any
// error tokens.
func NewLexer() Lexer {
	return lex.NewLexer(false)
}

// NewLazyLexer returns a Lexer whose Lex method will return a TokenStream that
// is lazily executed; that is to say, calling Next() on the token stream will
// perform only enough lexical analysis to produce the next token. Additionally,
// that TokenStream may produce an error token, which parsers would need to
// handle appropriately.
func NewLazyLexer() Lexer {
	return lex.NewLexer(true)
}

// NewParser returns what is the most flexible and efficient parser in this
// package that can parse the given grammar. The following parsers will be
// attempted to be built, in order, with each subsequent one attempted after the
// prior one fails.
//
// * LALR(1)
// * CLR(1)
// * SLR(1) (not currently attempted due to bugs)
// * LL(1)
//
// Returns an error if no parser can be generated for the given grammar.
//
// allowAmbiguous allows the use of ambiguous grammars in LR parsers. It has no
// effect on LL(1) parser generation; LL(1) grammars must be unambiguous.
func NewParser(g grammar.Grammar, allowAmbiguous bool) (parser Parser, ambigWarns []string, err error) {
	parser, ambigWarns, err = NewLALR1Parser(g, allowAmbiguous)
	if err != nil {
		bigParseGenErr := fmt.Sprintf("LALR(1) generation: %s", err.Error())
		// okay, what about a CLR(1) parser? (though, if LALR doesnt work, dont think CLR will)
		parser, ambigWarns, err = NewCLRParser(g, allowAmbiguous)
		if err != nil {
			bigParseGenErr += fmt.Sprintf("\nCLR(1) generation: %s", err.Error())

			// what about an SLR parser?
			// TODO: SLR fails and panics on some inputs (such as FISHI spec), fix this
			//parser, ambigWarns, err = NewSLRParser(g, allowAmbiguous) lol no SLR(1) currently has an error
			if err != nil {
				//bigParseGenErr += fmt.Sprintf("\nSLR(1) generation: %s", err.Error())

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
func SaveParserToDisk(p Parser, filename string) error {
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

func GetParserFromDisk(filename string) (Parser, error) {
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
func EncodeParserBytes(p Parser) []byte {
	data := decbin.EncString(p.Type().String())
	data = append(data, decbin.EncBinary(p)...)
	return data
}

// DecodeParserBytes takes bytes and returns the Parser encoded within it.
func DecodeParserBytes(data []byte) (p Parser, err error) {
	// first get the string giving the type
	typeStr, n, err := decbin.DecString(data)
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

	_, err = decbin.DecBinary(data[n:], p)
	return p, err
}

// NewLALR1Parser returns an LALR(1) parser that can generate parse trees for
// the given grammar. Returns an error if the grammar is not LALR(1).
func NewLALR1Parser(g grammar.Grammar, allowAmbiguous bool) (parser Parser, ambigWarns []string, err error) {
	return parse.GenerateLALR1Parser(g, allowAmbiguous)
}

// NewSLRParser returns an SLR(1) parser that can generate parse trees for the
// given grammar. Returns an error if the grammar is not SLR(1).
func NewSLRParser(g grammar.Grammar, allowAmbiguous bool) (parser Parser, ambigWarns []string, err error) {
	return parse.GenerateSimpleLRParser(g, allowAmbiguous)
}

// NewLL1Parser returns an LL(1) parser that can generate parse trees for the
// given grammar. Returns an error if the grammar is not LL(1).
func NewLL1Parser(g grammar.Grammar) (parser Parser, err error) {
	return parse.GenerateLL1Parser(g)
}

// NewCLRParser returns a canonical-LR(0) parser that can generate parse trees
// for the given grammar. Returns an error if the grammar is not CLR(1)
func NewCLRParser(g grammar.Grammar, allowAmbiguous bool) (parser Parser, ambigWarns []string, err error) {
	return parse.GenerateCanonicalLR1Parser(g, allowAmbiguous)
}

// NewSDD returns a new Syntax-Directed Definition Scheme.
func NewSDD() SDD {
	return translation.NewSDD()
}

// Frontend is a complete input-to-intermediate representation compiler
// front-end.
type Frontend[E any] struct {
	Lexer       Lexer
	Parser      Parser
	SDT         SDD
	IRAttribute string
}

// AnalyzeString is the same as Analyze but accepts a string as input. It simply
// creates a Reader on s and passes it to Analyze; this method is provided for
// convenience.
func (fe *Frontend[E]) AnalyzeString(s string) (ir E, err error) {
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
func (fe *Frontend[E]) Analyze(r io.Reader) (ir E, err error) {
	// lexical analysis
	tokStream, err := fe.Lexer.Lex(r)
	if err != nil {
		return ir, err
	}

	// syntactic analysis
	parseTree, err := fe.Parser.Parse(tokStream)
	if err != nil {
		return ir, err
	}

	// semantic analysis
	attrVals, err := fe.SDT.Evaluate(parseTree, fe.IRAttribute)
	if err != nil {
		return ir, err
	}

	// all analysis complete, now retrieve the result
	if len(attrVals) != 1 {
		return ir, fmt.Errorf("requested final IR attribute %q from root node but got %d values back", fe.IRAttribute, len(attrVals))
	}
	irUncast := attrVals[0]
	ir, ok := irUncast.(E)
	if !ok {
		// type mismatch; use reflections to collect type for err reporting
		irType := reflect.TypeOf(ir).Name()
		actualType := reflect.TypeOf(irUncast).Name()
		return ir, fmt.Errorf("expected final IR attribute %q to be of type %q at the root node, but result was of type %q", fe.IRAttribute, irType, actualType)
	}

	return ir, nil
}
