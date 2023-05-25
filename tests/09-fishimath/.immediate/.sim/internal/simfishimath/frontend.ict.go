// Package simfishimath contains the frontend for analyzing FISHIMath
// code. The function [Frontend] is the primary entrypoint for callers, and will
// return a full FISHIMath frontend ready for immediate use.
package simfishimath

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --clr --ir []github.com/dekarrin/fishimath/fmhooks.FMValue -l FISHIMath -v 1.0 -d /home/dekarrin/projects/ictiobus/tests/09-fishimath/fmc-eval --hooks fmhooks -S all --dev -nq /home/dekarrin/projects/ictiobus/tests/09-fishimath/fm-eval.md --sim-graphs --sim-trees
*/

import (
	"fmt"
	"os"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/lex"
	"github.com/dekarrin/ictiobus/trans"

	"github.com/dekarrin/ictiobus/langexec/fishimath/internal/fmhooks"
)

// FrontendOptions allows options to be set on the compiler frontend returned by
// [Frontend]. It allows setting of debug flags and other optional
// functionality.
type FrontendOptions struct {

	// LexerEager is whether the Lexer should immediately read all input the
	// first time it is called. The default is lazy lexing, where the minimum
	// number of tokens are read when required by the parser.
	LexerEager bool

	// LexerTrace is whether to add tracing functionality to the lexer. This
	// will cause the tokens to be printed to stderr as they are lexed. Note
	// that with LexerEager set, this implies that they will all be lexed and
	// therefore printed before any parsing occurs.
	LexerTrace bool

	// ParserTrace is whether to add tracing functionality to the parser. This
	// will cause parsing events to be printed to stderr as they occur. This
	// includes operations such as token or symbol stack manipulation, and for
	// LR parsers, shifts and reduces.
	ParserTrace bool
}

// Frontend returns the complete compiled frontend for the FISHIMath langauge.
// The hooks map must be provided as it is the interface between the translation
// scheme in the frontend and the external code executed in the backend. The
// opts parameter allows options to be set on the frontend for debugging and
// other purposes. If opts is nil, it is treated as an empty FrontendOptions.
func Frontend(hooks trans.HookMap, opts *FrontendOptions) ictiobus.Frontend[[]fmhooks.FMValue] {
	if opts == nil {
		opts = &FrontendOptions{}
	}

	fe := ictiobus.Frontend[[]fmhooks.FMValue]{

		Language:    "FISHIMath",
		Version:     "1.0",
		IRAttribute: "ir",
		Lexer:       Lexer(!opts.LexerEager),
		Parser:      Parser(),
		SDTS:        SDTS(),
	}

	// Add traces if requested

	if opts.LexerTrace {
		fe.Lexer.RegisterTokenListener(func(t lex.Token) {
			fmt.Fprintf(os.Stderr, "Token: %s\n", t)
		})
	}

	if opts.ParserTrace {
		fe.Parser.RegisterTraceListener(func(s string) {
			fmt.Fprintf(os.Stderr, "Parser: %s\n", s)
		})
	}

	// Set the hooks
	fe.SDTS.SetHooks(hooks)

	return fe
}