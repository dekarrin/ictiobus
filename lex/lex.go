package lex

import (
	"fmt"
	"io"
	"regexp"

	"github.com/dekarrin/ictiobus/internal/unregex"
	"github.com/dekarrin/ictiobus/types"
)

// TODO: src is useless as its in pat.
type patAct struct {
	priority int
	src      string
	pat      *regexp.Regexp
	act      Action
}

type lexerTemplate struct {
	lazy bool

	patterns   map[string][]patAct
	startState string

	// classes by ID by state
	classes map[string]map[string]types.TokenClass

	listener func(types.Token)
}

func NewLexer(lazy bool) *lexerTemplate {
	return &lexerTemplate{
		lazy:       lazy,
		patterns:   map[string][]patAct{},
		startState: "",
		classes:    map[string]map[string]types.TokenClass{},
	}
}

func (lx *lexerTemplate) Lex(input io.Reader) (types.TokenStream, error) {
	if lx.lazy {
		return lx.LazyLex(input)
	} else {
		return lx.ImmediatelyLex(input)
	}
}

func (lx *lexerTemplate) SetStartingState(s string) {
	lx.startState = s
}

func (lx *lexerTemplate) StartingState() string {
	return lx.startState
}

func (lx *lexerTemplate) RegisterTokenListener(fn func(t types.Token)) {
	lx.listener = fn
}

// AddClass adds the given token class to the lexer. This will mark that token
// class as a lexable token class, and make it available for use in the Action
// of an AddPattern.
//
// If the given token class's ID() returns a string matching one already added,
// the provided one will replace the existing one.
func (lx *lexerTemplate) RegisterClass(cl types.TokenClass, forState string) {
	stateClasses, ok := lx.classes[forState]
	if !ok {
		stateClasses = map[string]types.TokenClass{}
	}

	stateClasses[cl.ID()] = cl
	lx.classes[forState] = stateClasses
}

// GetPattern returns the pattern that will lex to the given token class. If no
// pattern lexes to the given token class, an empty string is returned.
// TODO: probably drop thie function, replaced it almost entirely with
// FakeLexemeProducer.
func (lx *lexerTemplate) GetPattern(cl types.TokenClass, forState string) string {
	statePatterns, ok := lx.patterns[forState]
	if !ok {
		return ""
	}

	for i := range statePatterns {
		pt := statePatterns[i]
		if (pt.act.Type == ActionScan || pt.act.Type == ActionScanAndState) && pt.act.ClassID == cl.ID() {
			return pt.src
		}
	}

	return ""
}

// Priority can be 0 for "in order added"
func (lx *lexerTemplate) AddPattern(pat string, action Action, forState string, priority int) error {
	statePatterns, ok := lx.patterns[forState]
	if !ok {
		statePatterns = make([]patAct, 0)
	}
	stateClasses, ok := lx.classes[forState]
	if !ok {
		stateClasses = map[string]types.TokenClass{}
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
			return fmt.Errorf("%q is not a defined token class on this lexer; add it with AddClass first", id)
		}
	}
	if action.Type == ActionState || action.Type == ActionScanAndState {
		if action.State == "" {
			return fmt.Errorf("action includes state shift but does not define state to shift to (cannot shift to empty state)")
		}
	}

	record := patAct{
		priority: priority,
		src:      pat,
		pat:      compiled,
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
				ur, ok := unregexers[pat.src]
				if !ok {
					var err error
					ur, err = unregex.New(pat.src)
					if err != nil {
						// should never happen
						panic(fmt.Sprintf("creating unregex for class %s (%q) failed: %v", pat.act.ClassID, pat.src, err))
					}
					ur.Seed(0)
					ur.AnyCharsMax = 0x04ff
					unregexers[pat.src] = ur
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
