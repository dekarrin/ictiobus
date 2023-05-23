package simfishimath

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --clr --ir []github.com/dekarrin/fishimath/fmhooks.FMValue -l FISHIMath -v 1.0 -d /home/dekarrin/projects/ictiobus/tests/09-fishimath/fmc-eval --hooks fmhooks -S all --dev -nq /home/dekarrin/projects/ictiobus/tests/09-fishimath/fm-eval.md --sim-graphs --sim-trees
*/

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"

	"github.com/dekarrin/ictiobus/langexec/fishimath/internal/simfishimath/simfishimathtoken"
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
	lx.RegisterClass(simfishimathtoken.TCAsterisk, "")
	lx.RegisterClass(simfishimathtoken.TCSolidus, "")
	lx.RegisterClass(simfishimathtoken.TC, "")
	lx.RegisterClass(simfishimathtoken.TCPlusSign, "")
	lx.RegisterClass(simfishimathtoken.TCFishtail, "")
	lx.RegisterClass(simfishimathtoken.TCFishhead, "")
	lx.RegisterClass(simfishimathtoken.TCShark, "")
	lx.RegisterClass(simfishimathtoken.TCTentacle, "")
	lx.RegisterClass(simfishimathtoken.TCId, "")
	lx.RegisterClass(simfishimathtoken.TCFloat, "")
	lx.RegisterClass(simfishimathtoken.TCInt, "")

	lx.AddPattern(`\s+`, lex.Discard(), "", 0)
	lx.AddPattern(`\*`, lex.LexAs(simfishimathtoken.TCAsterisk.ID()), "", 0)
	lx.AddPattern(`/`, lex.LexAs(simfishimathtoken.TCSolidus.ID()), "", 0)
	lx.AddPattern(`-`, lex.LexAs(simfishimathtoken.TC.ID()), "", 0)
	lx.AddPattern(`\+`, lex.LexAs(simfishimathtoken.TCPlusSign.ID()), "", 0)
	lx.AddPattern(`>\{`, lex.LexAs(simfishimathtoken.TCFishtail.ID()), "", 0)
	lx.AddPattern(`'\}`, lex.LexAs(simfishimathtoken.TCFishhead.ID()), "", 0)
	lx.AddPattern(`<o\^><`, lex.LexAs(simfishimathtoken.TCShark.ID()), "", 0)
	lx.AddPattern(`=o`, lex.LexAs(simfishimathtoken.TCTentacle.ID()), "", 0)
	lx.AddPattern(`[A-Za-z_][A-Za-z0-9_]*`, lex.LexAs(simfishimathtoken.TCId.ID()), "", 0)
	lx.AddPattern(`[0-9]*.[0-9]+`, lex.LexAs(simfishimathtoken.TCFloat.ID()), "", 0)
	lx.AddPattern(`[0-9]+`, lex.LexAs(simfishimathtoken.TCInt.ID()), "", 0)

	return lx
}
