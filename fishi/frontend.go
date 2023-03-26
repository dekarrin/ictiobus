package fishi

import (
	"errors"
	"fmt"
	"os"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/translation"
)

// GetFrontend gets the frontend for the fishi compiler-compiler. If cffFile is
// provided, it is used to load the cached parser from disk. Otherwise, a new
// frontend is created.
func GetFrontend(opts Options) (ictiobus.Frontend[AST], error) {
	// check for preload
	var preloadedParser ictiobus.Parser
	if opts.ParserCFF != "" && opts.ReadCache {
		var err error
		preloadedParser, err = ictiobus.GetParserFromDisk(opts.ParserCFF)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				preloadedParser = nil
			} else {
				return ictiobus.Frontend[AST]{}, fmt.Errorf("loading cachefile %q: %w", opts.ParserCFF, err)
			}
		}
	}

	feOpts := FrontendOptions{
		LexerTrace:  opts.LexerTrace,
		ParserTrace: opts.ParserTrace,
	}

	fishiFront := Frontend(feOpts, preloadedParser)

	// check the parser encoding if we generated a new one:
	if preloadedParser == nil && opts.ParserCFF != "" && opts.WriteCache {
		err := ictiobus.SaveParserToDisk(fishiFront.Parser, opts.ParserCFF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writing parser to disk: %s\n", err.Error())
		} else {
			fmt.Printf("wrote parser to %q\n", opts.ParserCFF)
		}
	}

	// validate our SDTS if we were asked to
	if opts.SDTSValidate {
		valProd := fishiFront.Lexer.FakeLexemeProducer(true, "")

		di := translation.ValidationOptions{
			ParseTrees:    opts.SDTSValShowTrees,
			FullDepGraphs: opts.SDTSValShowGraphs,
			ShowAllErrors: opts.SDTSValAllTrees,
			SkipErrors:    opts.SDTSValSkipTrees,
		}

		sddErr := fishiFront.SDT.Validate(fishiFront.Parser.Grammar(), fishiFront.IRAttribute, di, valProd)
		if sddErr != nil {
			return ictiobus.Frontend[AST]{}, fmt.Errorf("sdd validation error: %w", sddErr)
		}
	}

	return fishiFront, nil
}
