// Package lex provides lexing functionality for the ictiobus parser generator.
// It uses the regex provided by Go's built-in RE2 engine for matching on input
// and supports multiple states and state swapping, although does not retain any
// info about prior states.
//
// All lexers provided by this package support four different handlings of input
// pattern matching: lex the input and return a token of some class, change the
// lexer state to a new one, lex a token *and then* change the lexer state to a
// new one, or discard the matched text and continue from after it.
//
// Lexing is invoked by obtaining a [Lexer] and calling its Lex method. This
// will return a [TokenStream] that returns tokens lexed from input when its
// Next method is called. This TokenStream can be passed on to further stages of
// input analysis, such as a parser.
package lex

import (
	"fmt"
	"io"
	"regexp"

	"github.com/dekarrin/ictiobus/internal/unregex"
)

// A Lexer represents an in-progress or ready-built lexing engine ready for use.
// It can be stored as a byte representation and retrieved from bytes as well.
type Lexer interface {

	// Lex returns a token stream. The tokens may be lexed in a lazy fashion or
	// an immediate fashion; if it is immediate, errors will be returned at that
	// point. If it is lazy, then error token productions will be returned to
	// the callers of the returned TokenStream at the point where the error
	// occured.
	Lex(input io.Reader) (TokenStream, error)

	// RegisterClass registers a token class for use in some state of the Lexer.
	// Token classes must be registered before they can be used.
	RegisterClass(cl TokenClass, forState string)

	// AddPattern adds a new pattern for the lexer to recognize.
	AddPattern(pat string, action Action, forState string, priority int) error

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

	// RegisterTraceListener provides a function to call whenever a new token is
	// lexed. It can be used for debug purposes.
	RegisterTraceListener(func(t Token))
}

type patAct struct {
	priority int
	rx       *regexp.Regexp
	act      Action
}

type lexerTemplate struct {
	lazy bool

	patterns   map[string][]patAct
	startState string

	// classes by ID by state
	classes map[string]map[string]TokenClass

	listener func(Token)
}

// NewLexer creates a new Lexer that performs lexing in a lazy or immediate
// fashion as specified by lazy.
func NewLexer(lazy bool) Lexer {
	return &lexerTemplate{
		lazy:       lazy,
		patterns:   map[string][]patAct{},
		startState: "",
		classes:    map[string]map[string]TokenClass{},
	}
}

// Lex returns a stream of tokens lexed from the given input.
func (lx *lexerTemplate) Lex(input io.Reader) (TokenStream, error) {
	if lx.lazy {
		return lx.LazyLex(input)
	} else {
		return lx.ImmediatelyLex(input)
	}
}

// SetStartingState sets the initial state of the lexer. If not set, the
// starting state will be the default state.
func (lx *lexerTemplate) SetStartingState(s string) {
	lx.startState = s
}

// StartingState returns the initial state of the lexer. If one wasn't set, this
// will be the default state, "".
func (lx *lexerTemplate) StartingState() string {
	return lx.startState
}

// RegisterTraceListener provides a function to call whenever a new token is
// lexed. It can be used for debug purposes.
func (lx *lexerTemplate) RegisterTraceListener(fn func(t Token)) {
	lx.listener = fn
}

// RegisterClass adds the given token class to the lexer. This will mark that
// token class as a lexable token class, and make it available for use in the
// Action of an AddPattern.
//
// If the given token class's ID() returns a string matching one already added,
// the provided one will replace the existing one.
func (lx *lexerTemplate) RegisterClass(cl TokenClass, forState string) {
	stateClasses, ok := lx.classes[forState]
	if !ok {
		stateClasses = map[string]TokenClass{}
	}

	stateClasses[cl.ID()] = cl
	lx.classes[forState] = stateClasses
}

// AddPattern adds a new pattern and action to take to the lexer.
func (lx *lexerTemplate) AddPattern(pat string, action Action, forState string, priority int) error {
	statePatterns, ok := lx.patterns[forState]
	if !ok {
		statePatterns = make([]patAct, 0)
	}
	stateClasses, ok := lx.classes[forState]
	if !ok {
		stateClasses = map[string]TokenClass{}
	}

	compiled, err := regexp.Compile(pat)
	if err != nil {
		return fmt.Errorf("cannot compile regex: %w", err)
	}

	if action.Type == ActionScan || action.Type == ActionScanAndState {
		// check class exists
		id := action.ClassID
		_, ok := stateClasses[id]
		if !ok {
			return fmt.Errorf("%q is not a defined token class on this lexer; add it with RegisterClass first", id)
		}
	}
	if action.Type == ActionState || action.Type == ActionScanAndState {
		if action.State == "" {
			return fmt.Errorf("action includes state shift but does not define state to shift to (cannot shift to empty state)")
		}
	}

	record := patAct{
		priority: priority,
		rx:       compiled,
		act:      action,
	}
	statePatterns = append(statePatterns, record)

	lx.patterns[forState] = statePatterns
	// not modifying lx.classes so no need to set it again
	return nil
}

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
func (lx *lexerTemplate) FakeLexemeProducer(combine bool, state string) map[string]func() string {
	mapperFunc := map[string]func() string{}
	funcsForIDByState := map[string]map[string][]func() string{}
	unregexers := map[string]*unregex.Unregexer{}

	for st, patterns := range lx.patterns {
		funcsForID, ok := funcsForIDByState[st]
		if !ok {
			funcsForID = map[string][]func() string{}
		}
		for i := range patterns {
			pat := patterns[i]
			if pat.act.Type == ActionScan || pat.act.Type == ActionScanAndState {
				ur, ok := unregexers[pat.rx.String()]
				if !ok {
					var err error
					ur, err = unregex.New(pat.rx.String())
					if err != nil {
						// should never happen
						panic(fmt.Sprintf("creating unregex for class %s (%q) failed: %v", pat.act.ClassID, pat.rx.String(), err))
					}
					ur.Seed(0)
					ur.AnyCharsMax = 0x04ff
					unregexers[pat.rx.String()] = ur
				}

				idFuncs, ok := funcsForID[pat.act.ClassID]
				if !ok {
					idFuncs = []func() string{}
				}
				idFuncs = append(idFuncs, ur.Derive)

				funcsForID[pat.act.ClassID] = idFuncs
			}
		}

		funcsForIDByState[st] = funcsForID
	}

	includedFuncs := map[string][]func() string{}

	// now each ID that can be scanned should have one or more unregexer funcs
	// that can be used.
	if combine {
		for st := range funcsForIDByState {
			funcsByClassID := funcsForIDByState[st]
			for class := range funcsByClassID {
				stateIDFuncs := funcsByClassID[class]
				idFuncs, ok := includedFuncs[class]
				if !ok {
					idFuncs = []func() string{}
				}

				idFuncs = append(idFuncs, stateIDFuncs...)

				includedFuncs[class] = idFuncs
			}
		}
	} else {
		funcsByClassID := funcsForIDByState[state]
		for class := range funcsByClassID {
			stateIDFuncs := funcsByClassID[class]
			includedFuncs[class] = stateIDFuncs
		}
	}

	// combine all included functions into a single function for each ID
	for class := range includedFuncs {
		idFuncs := includedFuncs[class]
		if len(idFuncs) == 1 {
			// only one function, no need to get fancy
			mapperFunc[class] = idFuncs[0]
		} else {
			// multiple functions; close on a variable to track this
			var patternIndex int
			idClosure := func() string {
				toCall := idFuncs[patternIndex]
				patternIndex++
				patternIndex %= len(idFuncs)
				return toCall()
			}
			mapperFunc[class] = idClosure
		}
	}

	return mapperFunc
}

var eolRegex = regexp.MustCompile(`([^\n]*)(?:\n|$)`)

// scans through the reader to find the remainder of the current line and
// returns it
func readLineWithoutAdvancing(r *regexReader) string {
	r.Mark("line")
	matches, err := r.SearchAndAdvance(eolRegex)
	if err != nil {
		panic(fmt.Sprintf("trying to get rest of line: %s", err))
	}
	if len(matches) < 2 {
		panic("rest of line did not have subexpression")
	}
	line := matches[1]

	r.Restore("line")

	return line
}
