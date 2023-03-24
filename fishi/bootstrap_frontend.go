package fishi

import (
	"github.com/dekarrin/ictiobus"
)

// Frontend returns the created frontend for the fishi langauge. If any of the
// properties are non-nil, that will be used as that component of the parser;
// otherwise, the bootstrap versions of that component will be used (and
// possibly built from scratch). Up to one Lexer, up to one Parser, and up to
// one SDTS are allowed to be provided; doing so will replace the bootstrap
// version of that component with the provided one.
func Frontend(useComp ...interface{}) ictiobus.Frontend[AST] {
	var providedLx ictiobus.Lexer
	var providedParser ictiobus.Parser
	var providedSDTS ictiobus.SDTS

	// go through useComp and check if we have a Lexer, Parser, or SDTS. If any
	// other type is received, panic. If more than one of the same type is
	// received, panic. If the type is nil, do nothing.
	for _, comp := range useComp {
		switch comp.(type) {
		case ictiobus.Lexer:
			if providedLx != nil {
				panic("more than one lexer provided")
			}
			providedLx = comp.(ictiobus.Lexer)
		case ictiobus.Parser:
			if providedParser != nil {
				panic("more than one parser provided")
			}
			providedParser = comp.(ictiobus.Parser)
		case ictiobus.SDTS:
			if providedSDTS != nil {
				panic("more than one SDTS provided")
			}
			providedSDTS = comp.(ictiobus.SDTS)
		case nil:
			// do nothing
		default:
			panic("invalid type provided")
		}
	}

	fe := ictiobus.Frontend[AST]{
		Language:    "fishi",
		Version:     "1.0.0-bootstrap",
		IRAttribute: "ast",
	}

	if providedLx != nil {
		fe.Lexer = providedLx
	} else {
		fe.Lexer = CreateBootstrapLexer()
	}

	if providedParser != nil {
		fe.Parser = providedParser
	} else {
		// TODO: fishi should run analaysis on a grammar and parser when it is
		// building the frontend and also report ambigs.
		fe.Parser, _ = CreateBootstrapParser()
	}

	if providedSDTS != nil {
		fe.SDT = providedSDTS
	} else {
		fe.SDT = CreateBootstrapSDTS()
	}

	return fe
}