package fmfront

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --ir github.com/dekarrin/ictfishimath_ast/fmhooks.AST -l FISHIMath -v 1.0 -d /home/dekarrin/projects/ictiobus/examples/fishimath-ast/diag-fm --hooks fmhooks --dest /home/dekarrin/projects/ictiobus/examples/fishimath-ast/fmfront --pkg fmfront --dev /home/dekarrin/projects/ictiobus/examples/fishimath-ast/fm-ast.md
*/

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"

	"github.com/dekarrin/ictfishimath_ast/fmfront/fmfronttoken"
)

// Lexer returns the generated ictiobus Lexer for FISHIMath.
func Lexer(lazy bool) lex.Lexer {
	var lx lex.Lexer
	if lazy {
		lx = ictiobus.NewLazyLexer()
	} else {
		lx = ictiobus.NewLexer()
	}

	// default state, shared by all
	lx.RegisterClass(fmfronttoken.TCAsterisk, "")
	lx.RegisterClass(fmfronttoken.TCSolidus, "")
	lx.RegisterClass(fmfronttoken.TCHyphenMinus, "")
	lx.RegisterClass(fmfronttoken.TCPlusSign, "")
	lx.RegisterClass(fmfronttoken.TCFishtail, "")
	lx.RegisterClass(fmfronttoken.TCFishhead, "")
	lx.RegisterClass(fmfronttoken.TCShark, "")
	lx.RegisterClass(fmfronttoken.TCTentacle, "")
	lx.RegisterClass(fmfronttoken.TCId, "")
	lx.RegisterClass(fmfronttoken.TCFloat, "")
	lx.RegisterClass(fmfronttoken.TCInt, "")

	lx.AddPattern(`\s+`, lex.Discard(), "", 0)
	lx.AddPattern(`\*`, lex.LexAs(fmfronttoken.TCAsterisk.ID()), "", 0)
	lx.AddPattern(`/`, lex.LexAs(fmfronttoken.TCSolidus.ID()), "", 0)
	lx.AddPattern(`-`, lex.LexAs(fmfronttoken.TCHyphenMinus.ID()), "", 0)
	lx.AddPattern(`\+`, lex.LexAs(fmfronttoken.TCPlusSign.ID()), "", 0)
	lx.AddPattern(`>\{`, lex.LexAs(fmfronttoken.TCFishtail.ID()), "", 0)
	lx.AddPattern(`'\}`, lex.LexAs(fmfronttoken.TCFishhead.ID()), "", 0)
	lx.AddPattern(`<o\^><`, lex.LexAs(fmfronttoken.TCShark.ID()), "", 0)
	lx.AddPattern(`=o`, lex.LexAs(fmfronttoken.TCTentacle.ID()), "", 0)
	lx.AddPattern(`[A-Za-z_][A-Za-z0-9_]*`, lex.LexAs(fmfronttoken.TCId.ID()), "", 0)
	lx.AddPattern(`[0-9]*\.[0-9]+`, lex.LexAs(fmfronttoken.TCFloat.ID()), "", 0)
	lx.AddPattern(`[0-9]+`, lex.LexAs(fmfronttoken.TCInt.ID()), "", 0)

	return lx
}
