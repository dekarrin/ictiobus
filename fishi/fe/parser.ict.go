package fe

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --lalr --ir github.com/dekarrin/ictiobus/fishi/syntax.AST --dest fishi/fe -l FISHI -v 1.0.0 --hooks fishi/syntax fishi.md --dev
*/

import (
	_ "embed"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/parse"

	"github.com/dekarrin/ictiobus/fishi/fe/fetoken"
)

var (
	//go:embed parser.cff
	parserData []byte
)

// Grammar returns the grammar accepted by the generated ictiobus parser for
// FISHI. This grammar will also be included with with the parser itself,
// but it is included here as well for convenience.
func Grammar() grammar.CFG {
	g := grammar.CFG{}

	g.AddTerm(fetoken.TCAlt.ID(), fetoken.TCAlt)
	g.AddTerm(fetoken.TCAttrRef.ID(), fetoken.TCAttrRef)
	g.AddTerm(fetoken.TCDirDiscard.ID(), fetoken.TCDirDiscard)
	g.AddTerm(fetoken.TCDirHook.ID(), fetoken.TCDirHook)
	g.AddTerm(fetoken.TCDirHuman.ID(), fetoken.TCDirHuman)
	g.AddTerm(fetoken.TCDirIndex.ID(), fetoken.TCDirIndex)
	g.AddTerm(fetoken.TCDirPriority.ID(), fetoken.TCDirPriority)
	g.AddTerm(fetoken.TCDirProd.ID(), fetoken.TCDirProd)
	g.AddTerm(fetoken.TCDirSet.ID(), fetoken.TCDirSet)
	g.AddTerm(fetoken.TCDirShift.ID(), fetoken.TCDirShift)
	g.AddTerm(fetoken.TCDirState.ID(), fetoken.TCDirState)
	g.AddTerm(fetoken.TCDirSymbol.ID(), fetoken.TCDirSymbol)
	g.AddTerm(fetoken.TCDirToken.ID(), fetoken.TCDirToken)
	g.AddTerm(fetoken.TCDirWith.ID(), fetoken.TCDirWith)
	g.AddTerm(fetoken.TCEpsilon.ID(), fetoken.TCEpsilon)
	g.AddTerm(fetoken.TCEq.ID(), fetoken.TCEq)
	g.AddTerm(fetoken.TCEscseq.ID(), fetoken.TCEscseq)
	g.AddTerm(fetoken.TCFreeformText.ID(), fetoken.TCFreeformText)
	g.AddTerm(fetoken.TCHdrActions.ID(), fetoken.TCHdrActions)
	g.AddTerm(fetoken.TCHdrGrammar.ID(), fetoken.TCHdrGrammar)
	g.AddTerm(fetoken.TCHdrTokens.ID(), fetoken.TCHdrTokens)
	g.AddTerm(fetoken.TCId.ID(), fetoken.TCId)
	g.AddTerm(fetoken.TCInt.ID(), fetoken.TCInt)
	g.AddTerm(fetoken.TCNlEscseq.ID(), fetoken.TCNlEscseq)
	g.AddTerm(fetoken.TCNlFreeformText.ID(), fetoken.TCNlFreeformText)
	g.AddTerm(fetoken.TCNlNonterm.ID(), fetoken.TCNlNonterm)
	g.AddTerm(fetoken.TCNonterm.ID(), fetoken.TCNonterm)
	g.AddTerm(fetoken.TCTerm.ID(), fetoken.TCTerm)

	g.AddRule("FISHISPEC", []string{"BLOCKS"})

	g.AddRule("BLOCKS", []string{"BLOCKS", "BLOCK"})
	g.AddRule("BLOCKS", []string{"BLOCK"})

	g.AddRule("BLOCK", []string{"GBLOCK"})
	g.AddRule("BLOCK", []string{"TBLOCK"})
	g.AddRule("BLOCK", []string{"ABLOCK"})

	g.AddRule("ABLOCK", []string{"hdr-actions", "ACONTENT"})

	g.AddRule("ACONTENT", []string{"SYM-ACTIONS-LIST", "ASTATE-SET-LIST"})
	g.AddRule("ACONTENT", []string{"SYM-ACTIONS-LIST"})
	g.AddRule("ACONTENT", []string{"ASTATE-SET-LIST"})

	g.AddRule("ASTATE-SET-LIST", []string{"ASTATE-SET-LIST", "ASTATE-SET"})
	g.AddRule("ASTATE-SET-LIST", []string{"ASTATE-SET"})

	g.AddRule("ASTATE-SET", []string{"STATE-INS", "SYM-ACTIONS-LIST"})

	g.AddRule("SYM-ACTIONS-LIST", []string{"SYM-ACTIONS-LIST", "SYM-ACTIONS"})
	g.AddRule("SYM-ACTIONS-LIST", []string{"SYM-ACTIONS"})

	g.AddRule("SYM-ACTIONS", []string{"dir-symbol", "nonterm", "PROD-ACTION-LIST"})

	g.AddRule("PROD-ACTION-LIST", []string{"PROD-ACTION-LIST", "PROD-ACTION"})
	g.AddRule("PROD-ACTION-LIST", []string{"PROD-ACTION"})

	g.AddRule("PROD-ACTION", []string{"PROD-SPEC", "SEM-ACTION-LIST"})

	g.AddRule("SEM-ACTION-LIST", []string{"SEM-ACTION-LIST", "SEM-ACTION"})
	g.AddRule("SEM-ACTION-LIST", []string{"SEM-ACTION"})

	g.AddRule("SEM-ACTION", []string{"dir-set", "attr-ref", "dir-hook", "id"})
	g.AddRule("SEM-ACTION", []string{"dir-set", "attr-ref", "dir-hook", "id", "WITH"})

	g.AddRule("WITH", []string{"dir-with", "ATTR-REF-LIST"})

	g.AddRule("ATTR-REF-LIST", []string{"ATTR-REF-LIST", "attr-ref"})
	g.AddRule("ATTR-REF-LIST", []string{"attr-ref"})

	g.AddRule("PROD-SPEC", []string{"dir-prod", "PROD-ADDR"})
	g.AddRule("PROD-SPEC", []string{"dir-prod"})

	g.AddRule("PROD-ADDR", []string{"dir-index", "int"})
	g.AddRule("PROD-ADDR", []string{"APRODUCTION"})

	g.AddRule("APRODUCTION", []string{"ASYM-LIST"})
	g.AddRule("APRODUCTION", []string{"epsilon"})

	g.AddRule("ASYM-LIST", []string{"ASYM-LIST", "ASYM"})
	g.AddRule("ASYM-LIST", []string{"ASYM"})

	g.AddRule("ASYM", []string{"nonterm"})
	g.AddRule("ASYM", []string{"term"})
	g.AddRule("ASYM", []string{"int"})
	g.AddRule("ASYM", []string{"id"})

	g.AddRule("TBLOCK", []string{"hdr-tokens", "TCONTENT"})

	g.AddRule("TCONTENT", []string{"TENTRY-LIST", "TSTATE-SET-LIST"})
	g.AddRule("TCONTENT", []string{"TENTRY-LIST"})
	g.AddRule("TCONTENT", []string{"TSTATE-SET-LIST"})

	g.AddRule("TSTATE-SET-LIST", []string{"TSTATE-SET-LIST", "TSTATE-SET"})
	g.AddRule("TSTATE-SET-LIST", []string{"TSTATE-SET"})

	g.AddRule("TSTATE-SET", []string{"STATE-INS", "TENTRY-LIST"})

	g.AddRule("TENTRY-LIST", []string{"TENTRY-LIST", "TENTRY"})
	g.AddRule("TENTRY-LIST", []string{"TENTRY"})

	g.AddRule("TENTRY", []string{"PATTERN", "TOPTION-LIST"})

	g.AddRule("TOPTION-LIST", []string{"TOPTION-LIST", "TOPTION"})
	g.AddRule("TOPTION-LIST", []string{"TOPTION"})

	g.AddRule("TOPTION", []string{"DISCARD"})
	g.AddRule("TOPTION", []string{"STATESHIFT"})
	g.AddRule("TOPTION", []string{"TOKEN"})
	g.AddRule("TOPTION", []string{"HUMAN"})
	g.AddRule("TOPTION", []string{"PRIORITY"})

	g.AddRule("DISCARD", []string{"dir-discard"})

	g.AddRule("STATESHIFT", []string{"dir-shift", "TEXT"})

	g.AddRule("TOKEN", []string{"dir-token", "TEXT"})

	g.AddRule("HUMAN", []string{"dir-human", "TEXT"})

	g.AddRule("PRIORITY", []string{"dir-priority", "TEXT"})

	g.AddRule("PATTERN", []string{"TEXT"})

	g.AddRule("GBLOCK", []string{"hdr-grammar", "GCONTENT"})

	g.AddRule("GCONTENT", []string{"GRULE-LIST", "GSTATE-SET-LIST"})
	g.AddRule("GCONTENT", []string{"GRULE-LIST"})
	g.AddRule("GCONTENT", []string{"GSTATE-SET-LIST"})

	g.AddRule("GSTATE-SET-LIST", []string{"GSTATE-SET-LIST", "GSTATE-SET"})
	g.AddRule("GSTATE-SET-LIST", []string{"GSTATE-SET"})

	g.AddRule("GSTATE-SET", []string{"STATE-INS", "GRULE-LIST"})

	g.AddRule("GRULE-LIST", []string{"GRULE-LIST", "GRULE"})
	g.AddRule("GRULE-LIST", []string{"GRULE"})

	g.AddRule("GRULE", []string{"nl-nonterm", "eq", "ALTERNATIONS"})

	g.AddRule("ALTERNATIONS", []string{"GPRODUCTION"})
	g.AddRule("ALTERNATIONS", []string{"ALTERNATIONS", "alt", "GPRODUCTION"})

	g.AddRule("GPRODUCTION", []string{"GSYM-LIST"})
	g.AddRule("GPRODUCTION", []string{"epsilon"})

	g.AddRule("GSYM-LIST", []string{"GSYM-LIST", "GSYM"})
	g.AddRule("GSYM-LIST", []string{"GSYM"})

	g.AddRule("GSYM", []string{"nonterm"})
	g.AddRule("GSYM", []string{"term"})

	g.AddRule("STATE-INS", []string{"dir-state", "ID-EXPR"})

	g.AddRule("ID-EXPR", []string{"id"})
	g.AddRule("ID-EXPR", []string{"term"})

	g.AddRule("TEXT", []string{"NL-TEXT-ELEM", "TEXT-ELEM-LIST"})
	g.AddRule("TEXT", []string{"TEXT-ELEM-LIST"})
	g.AddRule("TEXT", []string{"NL-TEXT-ELEM"})

	g.AddRule("TEXT-ELEM-LIST", []string{"TEXT-ELEM-LIST", "TEXT-ELEM"})
	g.AddRule("TEXT-ELEM-LIST", []string{"TEXT-ELEM"})

	g.AddRule("NL-TEXT-ELEM", []string{"nl-escseq"})
	g.AddRule("NL-TEXT-ELEM", []string{"nl-freeform-text"})

	g.AddRule("TEXT-ELEM", []string{"escseq"})
	g.AddRule("TEXT-ELEM", []string{"freeform-text"})

	return g
}

// Parser returns the generated ictiobus Parser for FISHI.
func Parser() parse.Parser {
	p, err := parse.DecodeBytes(parserData)
	if err != nil {
		panic("corrupted parser.cff file: " + err.Error())
	}

	return p
}
