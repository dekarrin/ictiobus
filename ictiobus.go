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
	"encoding"
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

// A Lexer represents an in-progress or ready-built lexing engine ready for use.
// It can be stored as a byte representation and retrieved from bytes as well.
type Lexer interface {

	// Lex returns a token stream. The tokens may be lexed in a lazy fashion or
	// an immediate fashion; if it is immediate, errors will be returned at that
	// point. If it is lazy, then error token productions will be returned to
	// the callers of the returned TokenStream at the point where the error
	// occured.
	Lex(input io.Reader) (types.TokenStream, error)

	// RegisterClass registers a token class for use in some state of the Lexer.
	// Token classes must be registered before they can be used.
	RegisterClass(cl types.TokenClass, forState string)

	// AddPattern adds a new pattern for the lexer to recognize.
	AddPattern(pat string, action lex.Action, forState string, priority int) error

	// FakeLexemeProducer returns a map of token IDs to functions that will produce
	// a lexable value for that ID. As some token classes may have multiple ways of
	// lexing depending on the state, either state must be selected or combine must
	// be set to true.
	//
	// If combine is true, then state is ignored and all states' regexes for that ID
	// are combined into a single function that will alternate between them. If
	// combine is false, then state must be set and only the regexes for that state
	// are used to produce a lexable value.
	//
	// This can be useful for testing but may not produce useful values for all
	// token classes, especially those that have particularly complicated lexing
	// rules. If a caller finds that one of the functions in the map produced by
	// FakeLexemeProducer does not produce a lexable value, then it can be replaced
	// manually by replacing that entry in the map with a custom function.
	FakeLexemeProducer(combine bool, state string) map[string]func() string

	// SetStartingState sets the initial state of the lexer. If not set, the
	// starting state will be the default state.
	SetStartingState(s string)

	// StartingState returns the initial state of the lexer. If one wasn't set, this
	// will be the default state, "".
	StartingState() string

	// RegisterTokenListener provides a function to call whenever a new token is
	// lexed. It can be used for debug purposes.
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

	// DFAString returns a string representation of the DFA for this parser, if one
	// so exists. Will return the empty string if the parser is not of the type
	// to have a DFA.
	DFAString() string

	// Grammar returns the grammar that this parser can parse.
	Grammar() grammar.Grammar
}

// SDTS is a series of syntax-directed translations bound to syntactic rules of
// a grammar. It is used for evaluation of a parse tree into an intermediate
// representation, or for direct execution.
//
// Strictly speaking, this is closer to an Attribute grammar.
type SDTS interface {
	// SetHooks sets the hook table for mapping SDTS hook names as used in a
	// call to BindSynthesizedAttribute or BindInheritedAttribute to their
	// actual implementations.
	//
	// Because the map from strings to function pointers, this hook map must be
	// set at least once before the SDTS is used. It is recommended to set it
	// every time the SDTS is loaded as soon as it is loaded.
	//
	// Calling it multiple times will add to the existing hook table, not
	// replace it entirely. If there are any duplicate hook names, the last one
	// set will be the one that is used.
	SetHooks(hooks trans.HookMap)

	// BindInheritedAttribute creates a new SDTS binding for setting the value
	// of an inherited attribute with name attrName. The production that the
	// inherited attribute is set on is specified with forProd, which must have
	// its Type set to something other than RelHead (inherited attributes can be
	// set only on production symbols).
	//
	// The binding applies only on nodes in the parse tree created by parsing
	// the grammar rule productions with head symbol head and production symbols
	// prod.
	//
	// The AttributeSetter bound to hook is called when the inherited value
	// attrName is to be set, in order to calculate the new value. Attribute
	// values to pass in as arguments are specified by passing references to the
	// node and attribute name whose value to retrieve in the withArgs slice.
	// Explicitlygiving the referenced attributes in this fashion makes it easy
	// to determine the dependency graph for later execution.
	BindInheritedAttribute(head string, prod []string, attrName string, hook string, withArgs []trans.AttrRef, forProd trans.NodeRelation) error

	// BindSynthesizedAttribute creates a new SDTS binding for setting the value
	// of a synthesized attribute with name attrName. The attribute is set on
	// the symbol at the head of the rule that the binding is being created for.
	//
	// The binding applies only on nodes in the parse tree created by parsing
	// the grammar rule productions with head symbol head and production symbols
	// prod.
	//
	// The AttributeSetter bound to hook is called when the synthesized value
	// attrName is to be set, in order to calculate the new value. Attribute
	// values to pass in as arguments are specified by passing references to the
	// node and attribute name whose value to retrieve in the withArgs slice.
	// Explicitly giving the referenced attributes in this fashion makes it easy
	// to determine the dependency graph for later execution.
	BindSynthesizedAttribute(head string, prod []string, attrName string, hook string, withArgs []trans.AttrRef) error

	// SetNoFlow sets a binding to be explicitly allowed to not be required to
	// flow up to a particular parent. This will prevent it from causing an
	// error if it results in a disconnected dependency graph if the node of
	// that binding has the given parent.
	//
	// - forProd is only used if synth is false. It specifies the production
	// that the binding to match must apply to.
	// - which is the index of the binding to set it on, if multiple match the
	// prior criteria. Set to -1 or less to set it on all matching bindings.
	// - ifParent is the symbol that the parent of the node must be for no flow
	// to be considered acceptable.
	SetNoFlow(synth bool, head string, prod []string, attrName string, forProd trans.NodeRelation, which int, ifParent string) error

	// Bindings returns all bindings defined to apply when at a node in a parse
	// tree created by the rule production with head as its head symbol and prod
	// as its produced symbols. They will be returned in the order they were
	// defined.
	Bindings(head string, prod []string) []trans.SDDBinding

	// Bindings returns all bindings defined to apply when at a node in a parse
	// tree created by the rule production with head as its head symbol and prod
	// as its produced symbols, and when setting the attribute referred to by
	// dest. They will be returned in the order they were defined.
	BindingsFor(head string, prod []string, dest trans.AttrRef) []trans.SDDBinding

	// Evaluate takes a parse tree and executes the semantic actions defined as
	// SDDBindings for a node for each node in the tree and on completion,
	// returns the requested attributes values from the root node. Execution
	// order is automatically determined by taking the dependency graph of the
	// SDTS; cycles are not supported. Do note that this does not require the
	// SDTS to be S-attributed or L-attributed, only that it not have cycles in
	// its value dependency graph.
	//
	// Warn errors are provided in the slice of error and can be populated
	// regardless of whether the final (actual) error is non-nil.
	Evaluate(tree types.ParseTree, attributes ...string) (vals []interface{}, warns []error, err error)

	// Validate checks whether this SDTS is valid for the given grammar. It will
	// create a simulated parse tree that contains a node for every rule of the
	// given grammar and will attempt to evaluate it, returning an error if
	// there is any issue running the bindings.
	//
	// fakeValProducer should be a map of token class IDs to functions that can
	// produce fake values for the given token class. This is used to simulate
	// actual lexemes in the parse tree. If not provided, entirely contrived
	// values will be used, which may not behave as expected with the SDTS. To
	// get one that will use the configured regexes of tokens used for lexing,
	// call FakeLexemeProducer on a Lexer.
	Validate(grammar grammar.Grammar, attribute string, debug trans.ValidationOptions, fakeValProducer ...map[string]func() string) (warns []string, err error)
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
// * SLR(1)
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

// GetParserFromDisk retrieves the parser from a binary file on disk.
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
	data := rezi.EncString(p.Type().String())
	data = append(data, rezi.EncBinary(p)...)
	return data
}

// DecodeParserBytes takes bytes and returns the Parser encoded within it.
func DecodeParserBytes(data []byte) (p Parser, err error) {
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

// NewSDTS returns a new Syntax-Directed Translation Scheme.
func NewSDTS() SDTS {
	return trans.NewSDTS()
}

// Frontend is a complete input-to-intermediate representation compiler
// front-end.
type Frontend[E any] struct {
	Lexer       Lexer
	Parser      Parser
	SDT         SDTS
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
func (fe *Frontend[E]) AnalyzeString(s string) (ir E, pt *types.ParseTree, err error) {
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
func (fe *Frontend[E]) Analyze(r io.Reader) (ir E, pt *types.ParseTree, err error) {
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
