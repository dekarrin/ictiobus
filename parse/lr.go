package parse

import (
	"encoding"
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/box"
	"github.com/dekarrin/ictiobus/internal/rezi"
	"github.com/dekarrin/ictiobus/internal/textfmt"
	"github.com/dekarrin/ictiobus/lex"
)

// lrParseTable is a table of information passed to an LR parser. These will be
// generated from a grammar for the purposes of performing bottom-up parsing.
type lrParseTable interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	// Initial returns the initial state of the parse table, if that is
	// applicable for the table.
	Initial() string

	// Action returns the LR-parser action to perform given that the current
	// state is i and the next terminal input symbol seen is a.
	Action(state, symbol string) lrAction

	// Goto maps a state and a grammar symbol to some other state. It specifies
	// the state to transition to after reducing to a non-terminal symbol.
	Goto(state, symbol string) (string, error)

	// String prints a string representation of the table. If two LRParseTables
	// produce the same String() output, they are considered equal.
	String() string

	// DFAString returns the DFA simulated by the table. Some tables may in fact
	// be the DFA itself along with supplementary info.
	DFAString() string
}

type lrParser struct {
	table     lrParseTable
	parseType Algorithm
	gram      grammar.CFG
	trace     func(s string)
}

// Grammar returns the grammar that was used to generate the parser.
func (lr *lrParser) Grammar() grammar.CFG {
	return lr.gram
}

// DFAString returns a string representation. of the DFA that drives the LR
// parser.
func (lr *lrParser) DFAString() string {
	return lr.table.DFAString()
}

// RegisterTraceListener sets a function to be called with messages that
// indicate what action the parser is taking. It is useful for debug purposes.
func (lr *lrParser) RegisterTraceListener(listener func(s string)) {
	lr.trace = listener
}

// Type returns the type of the parser.
func (lr *lrParser) Type() Algorithm {
	return lr.parseType
}

// TableString returns the parser table as a string.
func (lr *lrParser) TableString() string {
	return lr.table.String()
}

// MarshalBinary converts lr into a slice of bytes that can be decoded with
// UnmarshalBinary.
func (lr *lrParser) MarshalBinary() ([]byte, error) {
	data := rezi.EncString(lr.parseType.String())
	data = append(data, rezi.EncBinary(lr.table)...)
	data = append(data, rezi.EncBinary(lr.gram)...)
	return data, nil
}

// UnmarshalBinary decodes a slice of bytes created by MarshalBinary into lr.
// All of lr's fields will be replaced by the fields decoded from data.
func (lr *lrParser) UnmarshalBinary(data []byte) error {
	var err error
	var n int

	var parseTypeName string
	parseTypeName, n, err = rezi.DecString(data)
	if err != nil {
		return fmt.Errorf("parseType: %w", err)
	}
	data = data[n:]
	lr.parseType, err = ParseAlgorithm(parseTypeName)
	if err != nil {
		return fmt.Errorf("parsing parseType: %w", err)
	}

	var tableVal lrParseTable
	switch lr.parseType {
	case CLR1:
		tableVal = &canonicalLR1Table{}
	case LALR1:
		tableVal = &lalr1Table{}
	case SLR1:
		tableVal = &slrTable{}
	default:
		return fmt.Errorf("unknown parse type: %s", lr.parseType.String())
	}

	n, err = rezi.DecBinary(data, tableVal)
	if err != nil {
		return fmt.Errorf("table: %w", err)
	}
	data = data[n:]
	lr.table = tableVal

	_, err = rezi.DecBinary(data, &lr.gram)
	if err != nil {
		return fmt.Errorf("gram: %w", err)
	}

	return nil
}

func (lr lrParser) notifyTraceFn(fn func() string) {
	if lr.trace != nil {
		lr.trace(fn())
	}
}

func (lr lrParser) notifyTrace(fmtStr string, args ...interface{}) {
	lr.notifyTraceFn(func() string { return fmt.Sprintf(fmtStr, args...) })
}

func (lr lrParser) notifyStatePush(s string) {
	lr.notifyTrace("states.push(): %s", s)
}

func (lr lrParser) notifyStatePop(s string) {
	if s == "" {
		lr.notifyTrace("states.pop()")
	} else {
		lr.notifyTrace("states.pop(): %s", s)
	}
}

func (lr lrParser) notifyAction(act lrAction) {
	lr.notifyTrace("Action: %s", act.Type.String())
}

func (lr lrParser) notifyNextToken(tok lex.Token) {
	lr.notifyTrace("Got next token: %s", tok.String())
}

func (lr lrParser) notifyTokenStack(st *box.Stack[lex.Token]) {
	stackElems := st.Elements()
	lr.notifyTraceFn(func() string {
		var lexStr strings.Builder
		var tokStr strings.Builder
		for i := range stackElems {
			tok := stackElems[(len(stackElems)-1)-i]
			lexStr.WriteRune('"')
			lexStr.WriteString(strings.ReplaceAll(tok.Lexeme(), "\n", "\\n"))
			lexStr.WriteRune('"')

			tokStr.WriteString(strings.ToUpper(tok.Class().ID()))

			if i+1 < len(stackElems) {
				lexStr.WriteString(", ")
				tokStr.WriteString(", ")
			}
		}
		if len(stackElems) < 1 {
			lexStr.WriteString("(empty)")
			tokStr.WriteString("(empty)")
		}

		str := fmt.Sprintf("Token stack (lexed): %s", lexStr.String())
		str += "\n"
		str += fmt.Sprintf("Token stack (ttype): %s", tokStr.String())

		return str
	})
}

// Parse parses the input stream with the internal LR parse table. If any syntax
// errors are encountered, an empty parse tree and a *types.SyntaxError is
// returned.
//
// This is an implementation of Algorithm 4.44, "LR-parsing algorithm", from
// the purple dragon book.
func (lr *lrParser) Parse(stream lex.TokenStream) (Tree, error) {
	stateStack := box.NewStack([]string{lr.table.Initial()})

	// we will use these to build our parse tree
	tokenBuffer := &box.Stack[lex.Token]{}
	subTreeRoots := &box.Stack[*Tree]{}

	// let a be the first symbol of w$;
	a := stream.Next()
	lr.notifyNextToken(a)

	for { /* repeat forever */
		// let s be the state on top of the stack;
		s := stateStack.Peek()

		lr.notifyTrace("NEXT PASS: state=%s, tok=%s", s, a.String())
		lr.notifyTokenStack(tokenBuffer)

		ACTION := lr.table.Action(s, a.Class().ID())
		lr.notifyAction(ACTION)

		switch ACTION.Type {
		case lrShift: // if ( ACTION[s, a] = shift t )
			// add token to our buffer
			tokenBuffer.Push(a)

			t := ACTION.State

			// push t onto the stack
			stateStack.Push(t)
			lr.notifyStatePush(t)

			// let a be the next input symbol
			a = stream.Next()
			lr.notifyNextToken(a)
		case lrReduce: // else if ( ACTION[s, a] = reduce A -> β )
			A := ACTION.Symbol
			beta := ACTION.Production
			prodStr := strings.ToLower(beta.String())
			if len(prodStr) == 0 {
				prodStr = grammar.Epsilon.String()
			}
			lr.notifyTrace("%s -> %s", strings.ToUpper(A), prodStr)

			// use the reduce to create a node in the parse tree
			node := &Tree{Value: A, Children: make([]*Tree, 0)}

			// SPECIAL CASE: if we just reduced an epsilon production, immediately
			// add the epsilon node to the new one
			if len(beta) == 0 {
				node.Children = append(node.Children, &Tree{
					Terminal: true,
				})
			}

			// we need to go from right to left of the production to pop things
			// from the stacks in the correct order
			for i := len(beta) - 1; i >= 0; i-- {
				sym := beta[i]
				if strings.ToLower(sym) == sym {
					// it is a terminal. read the source from the token buffer
					tok := tokenBuffer.Pop()
					subNode := &Tree{Terminal: true, Value: tok.Class().ID(), Source: tok}
					node.Children = append([]*Tree{subNode}, node.Children...)
				} else {
					// it is a non-terminal. it should be in our stack of
					// current tree roots.
					subNode := subTreeRoots.Pop()
					node.Children = append([]*Tree{subNode}, node.Children...)
				}
			}
			// remember it for next time
			subTreeRoots.Push(node)

			// pop |β| symbols off the stack;
			for i := 0; i < len(beta); i++ {
				stateStack.Pop()
				lr.notifyStatePop("")
			}

			// let state t now be on top of the stack
			t := stateStack.Peek()
			lr.notifyTrace("back to old state %s", t)

			// push GOTO[t, A] onto the stack
			toPush, err := lr.table.Goto(t, A)
			if err != nil {
				return Tree{}, lex.NewSyntaxErrorFromToken(fmt.Sprintf("LR parsing error; DFA has no valid transition from here on %q", A), a)
			}
			stateStack.Push(toPush)
			lr.notifyTrace("Transition %s =(%q)=> %s", t, strings.ToLower(A), toPush)
			lr.notifyStatePush(toPush)

			// output the production A -> β
			// (TODO: put it on the parse tree)
		case lrAccept: // else if ( ACTION[s, a] = accept )
			// parsing is done. there should be at least one item on the stack
			pt := subTreeRoots.Pop()
			return *pt, nil
		case lrError:

			// call error-recovery routine here when/if we add it in future
			// - So never, huh? -V
			// - nonononono glub, it's a feature req for later -D
			expMessage := lr.getExpectedString(s)

			// if it's an error token, then display that as a message
			if a.Class().ID() == lex.TokenError.ID() {
				return Tree{}, lex.NewSyntaxErrorFromToken(fmt.Sprintf("%s; %s", a.Lexeme(), expMessage), a)
			}
			return Tree{}, lex.NewSyntaxErrorFromToken(fmt.Sprintf("unexpected %s; %s", a.Class().Human(), expMessage), a)
		}
		lr.notifyTrace("-----------------")
	}
}

func (lr lrParser) getExpectedString(stateName string) string {
	expected := lr.findExpectedTokens(stateName)

	var sb strings.Builder

	sb.WriteString("expected ")

	commas := false
	finalOr := false

	if len(expected) > 1 {
		finalOr = true
		if len(expected) > 2 {
			commas = true
		}
	}

	var prevEndedWithSpace bool
	for i := range expected {
		t := expected[i]

		if i == 0 {
			sb.WriteString(textfmt.ArticleFor(t.Human(), false))
			sb.WriteRune(' ')
		}

		if finalOr && i+1 == len(expected) {
			if !prevEndedWithSpace {
				sb.WriteRune(' ')
			}
			sb.WriteString("or ")
		}

		sb.WriteString(t.Human())
		if commas && i+1 < len(expected) {
			sb.WriteString(", ")
			prevEndedWithSpace = true
		} else {
			prevEndedWithSpace = false
		}
	}

	return sb.String()
}

// findExpectedAt returns all token classes that are allowed/expected for
// the given state, that is, those symbols that result in a non-error entry.
func (lr lrParser) findExpectedTokens(stateName string) []lex.TokenClass {
	terms := lr.gram.Terminals()

	classes := make([]lex.TokenClass, 0)
	for i := range terms {
		t := lr.gram.Term(terms[i])
		act := lr.table.Action(stateName, t.ID())
		if act.Type != lrError {
			classes = append(classes, t)
		}
	}

	return classes
}
