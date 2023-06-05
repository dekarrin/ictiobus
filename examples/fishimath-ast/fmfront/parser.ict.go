package fmfront

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr --ir github.com/dekarrin/ictfishimath_ast/fmhooks.AST -l FISHIMath -v 1.0 -d /home/dekarrin/projects/ictiobus/examples/fishimath-ast/diag-fm --hooks fmhooks --dest /home/dekarrin/projects/ictiobus/examples/fishimath-ast/fmfront --pkg fmfront --dev /home/dekarrin/projects/ictiobus/examples/fishimath-ast/fm-ast.md
*/

import (
	_ "embed"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/parse"

	"github.com/dekarrin/ictfishimath_ast/fmfront/fmfronttoken"
)

var (
	//go:embed parser.cff
	parserData []byte
)

// Grammar returns the grammar accepted by the generated ictiobus parser for
// FISHIMath. This grammar will also be included with with the parser itself,
// but it is included here as well for convenience.
func Grammar() grammar.CFG {
	g := grammar.CFG{}

	g.AddTerm(fmfronttoken.TCAsterisk.ID(), fmfronttoken.TCAsterisk)
	g.AddTerm(fmfronttoken.TCPlusSign.ID(), fmfronttoken.TCPlusSign)
	g.AddTerm(fmfronttoken.TCHyphenMinus.ID(), fmfronttoken.TCHyphenMinus)
	g.AddTerm(fmfronttoken.TCSolidus.ID(), fmfronttoken.TCSolidus)
	g.AddTerm(fmfronttoken.TCFishhead.ID(), fmfronttoken.TCFishhead)
	g.AddTerm(fmfronttoken.TCFishtail.ID(), fmfronttoken.TCFishtail)
	g.AddTerm(fmfronttoken.TCFloat.ID(), fmfronttoken.TCFloat)
	g.AddTerm(fmfronttoken.TCId.ID(), fmfronttoken.TCId)
	g.AddTerm(fmfronttoken.TCInt.ID(), fmfronttoken.TCInt)
	g.AddTerm(fmfronttoken.TCShark.ID(), fmfronttoken.TCShark)
	g.AddTerm(fmfronttoken.TCTentacle.ID(), fmfronttoken.TCTentacle)

	g.AddRule("FISHIMATH", []string{"STATEMENTS"})

	g.AddRule("STATEMENTS", []string{"STMT", "STATEMENTS"})
	g.AddRule("STATEMENTS", []string{"STMT"})

	g.AddRule("STMT", []string{"EXPR", "shark"})

	g.AddRule("EXPR", []string{"id", "tentacle", "EXPR"})
	g.AddRule("EXPR", []string{"SUM"})

	g.AddRule("SUM", []string{"PRODUCT", "+", "EXPR"})
	g.AddRule("SUM", []string{"PRODUCT", "-", "EXPR"})
	g.AddRule("SUM", []string{"PRODUCT"})

	g.AddRule("PRODUCT", []string{"TERM", "*", "PRODUCT"})
	g.AddRule("PRODUCT", []string{"TERM", "/", "PRODUCT"})
	g.AddRule("PRODUCT", []string{"TERM"})

	g.AddRule("TERM", []string{"fishtail", "EXPR", "fishhead"})
	g.AddRule("TERM", []string{"int"})
	g.AddRule("TERM", []string{"float"})
	g.AddRule("TERM", []string{"id"})

	return g
}

// Parser returns the generated ictiobus Parser for FISHIMath.
func Parser() parse.Parser {
	p, err := parse.DecodeBytes(parserData)
	if err != nil {
		panic("corrupted parser.cff file: " + err.Error())
	}

	return p
}