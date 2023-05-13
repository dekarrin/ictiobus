package lex

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// regex for newline/whitespace prefix
	atLeastOneNewlineWSPrefixRegex = regexp.MustCompile(`^\s*\n\s*\S`)
)

type lazyTokenStream struct {
	// buffered reader that can run regex and retrieve results
	r *regexReader

	// cur state
	state string

	// track these for placement in tokens, for later error reporting
	curLine     int
	curPos      int
	curFullLine string

	// set to true when the lazyLex has reached end of input, causing all
	// subsequent calls to Next() to return a Token with class
	// types.TokenEndOfText and all subsequent calls to HasNext() to return
	// false.
	done bool

	// panic mode is entered when no lexeme is found; the next call to Next()
	// will begin discarding characters until a valid one is found
	panicMode bool

	// classes mapping
	classes map[string]map[string]TokenClass

	// split actions from regexes to match indexes to capturing groups
	actions map[string][]Action

	// one regex per state. each regex will be constructed by taking all regex
	// for a state and placing them in capturing groups separated by alternation
	// operators.
	patterns map[string]*regexp.Regexp

	// listener is called whenever a token is produced
	listener func(Token)
}

// LazyLex returns a token stream that reads the provided input only as much as
// it needs to to, preferring to stop once it has lexed a token until the next
// call to Next(). If any lexing errors occur, they will be returned as an Error
// token from the stream's Next() method.
func (lx *lexerTemplate) LazyLex(input io.Reader) (TokenStream, error) {
	active := &lazyTokenStream{
		r:        newRegexReader(input),
		patterns: make(map[string]*regexp.Regexp),
		classes:  make(map[string]map[string]TokenClass),
		actions:  make(map[string][]Action),
		state:    lx.StartingState(),
		listener: lx.listener,
	}

	// move all patterns into "super pattern"; one per state. and separate the
	// actions into their own data structure
	for k := range lx.patterns {
		var statePats []patAct
		if k != "" {
			defPats, ok := lx.patterns[""]
			if ok {
				statePats = defPats
			}
		}
		statePats = append(statePats, lx.patterns[k]...)

		// sort by priority
		priorityPats := [][]patAct{}
		for i := range statePats {
			p := statePats[i]
			if p.priority <= 0 {
				p.priority = 0
			}

			for len(priorityPats) < (p.priority + 1) {
				priorityPats = append(priorityPats, []patAct{})
			}

			priorityList := priorityPats[p.priority]
			priorityList = append(priorityList, p)
			priorityPats[p.priority] = priorityList
		}
		// and then put back in statePats
		//
		// (but 0 is actually the LOWEST priority; other than that, all others
		// are simply in numerical order)
		statePats = make([]patAct, 0)
		for i := range priorityPats {
			if i == 0 {
				continue
			}
			statePats = append(statePats, priorityPats[i]...)
		}
		statePats = append(statePats, priorityPats[0]...)

		var superRegex strings.Builder
		superRegex.WriteString("^(?:")
		lazyActs := make([]Action, len(statePats))

		for i := range statePats {
			act := statePats[i].act
			src := statePats[i].rx.String()
			superRegex.WriteString("(" + src + ")")
			if i+1 < len(statePats) {
				superRegex.WriteRune('|')
			}
			lazyActs[i] = act
		}

		superRegex.WriteRune(')')

		compiled, err := regexp.Compile(superRegex.String())
		if err != nil {
			// should never happen
			return nil, fmt.Errorf("composing token regexes: %w", err)
		}

		active.patterns[k] = compiled
		active.actions[k] = lazyActs
	}

	// move over classes too (although they might not be needed)
	for k := range lx.classes {
		stateClasses := map[string]TokenClass{}
		if k != "" {
			defClasses, ok := lx.classes[""]
			if ok {
				for j := range defClasses {
					stateClasses[j] = defClasses[j]
				}
			}
		}
		for j := range lx.classes[k] {
			stateClasses[j] = lx.classes[k][j]
		}

		active.classes[k] = stateClasses
	}

	// set current line and pos
	active.curLine = 1
	active.curPos = 1
	active.curFullLine = readLineWithoutAdvancing(active.r)

	return active, nil
}

// Next returns the next token in the stream and advances the stream by one
// token. If at the end of the stream, this will return a token whose Class()
// is types.TokenEndOfText. If an error in lexing occurs, it will return a token
// whose Class() is types.TokenError and whose lexeme is a message explaining
// the error.
func (lx *lazyTokenStream) Next() Token {
	if lx.done {
		return lx.makeEOTToken()
	}

	// the rule that you get all default states along with whatever state
	pat := lx.patterns[lx.state]
	stateActions := lx.actions[lx.state]
	stateClasses := lx.classes[lx.state]

	var matches []string
	var readError error
	for {
		// retrieve the current matches, discarding runes until we find a match
		// if in panic mode.

		if lx.panicMode {
			for lx.panicMode {
				// track the rune we are dropping to add to source text context
				// tracking
				var ch rune
				ch, _, readError = lx.r.ReadRune()

				if readError != nil {
					return lx.tokenForIOError(readError)
				}

				if ch == '\n' {
					lx.curLine++
					lx.curPos = 0
					lx.curFullLine = readLineWithoutAdvancing(lx.r)
				}
				lx.curPos++

				matches, readError = lx.r.SearchAndAdvance(pat)
				if readError != nil {
					return lx.tokenForIOError(readError)
				}

				if len(matches) > 0 {
					// we found something. exit panic mode and continue
					lx.panicMode = false
				}
			}
		} else {
			matches, readError = lx.r.SearchAndAdvance(pat)
			if readError != nil {
				return lx.tokenForIOError(readError)
			}

			if len(matches) < 1 {
				// no match at start of reader. return an error token and enter
				// panic mode
				lx.panicMode = true
				return lx.makeErrorTokenf("unknown input")
			}
		}

		actionIdx, lexeme := lx.selectMatch(matches)

		// update source text context tracking BEFORE creating token in case
		// we need to update it for a token that starts with a newline
		var numNewLines int
		var leadingLineChars string
		curLine := lx.curLine
		curPos := lx.curPos
		curFullLine := lx.curFullLine
		for _, ch := range lexeme {
			if ch == '\n' {
				curLine++
				curPos = 0
				numNewLines++
				leadingLineChars = ""
			} else if numNewLines > 0 {
				leadingLineChars += string(ch)
			}
			curPos++
		}
		if numNewLines > 0 {
			curFullLine = leadingLineChars + readLineWithoutAdvancing(lx.r)
		}

		action := stateActions[actionIdx]
		var tok Token
		var retToken bool

		// if lexeme has a prefix consisting of only whitespace with at least
		// one newline, and lexeme contains at least one non-whitespace rune,
		// then the source line info should be updated to point to the first
		// non-whitespace rune for the creation of the token with that lexeme to
		// aid in error reporting.
		if (action.Type == ActionScan || action.Type == ActionScanAndState) && atLeastOneNewlineWSPrefixRegex.MatchString(lexeme) {
			// find the point where the first non-whitespace rune is

			// Feels like we could have 8een doing this a8ove while looping over
			// the string; this means that this will 8e the third time we're
			// doing it!
			var wsScanNumNewLines int
			var wsScanLeadingLineChars string
			var hitNonSpace bool
			var hitNLAfterNonSpace bool
			for _, ch := range lexeme {
				if ch == '\n' {
					if !hitNonSpace {
						lx.curLine++
						lx.curPos = 0
						wsScanLeadingLineChars = ""
					} else {
						hitNLAfterNonSpace = true
					}
					wsScanNumNewLines++
				} else if !unicode.IsSpace(ch) {
					hitNonSpace = true
					if wsScanNumNewLines > 0 && !hitNLAfterNonSpace {
						wsScanLeadingLineChars += string(ch)
					}
				} else if wsScanNumNewLines > 0 && !hitNLAfterNonSpace {
					wsScanLeadingLineChars += string(ch)
				}
				if !hitNonSpace {
					lx.curPos++
				}
			}
			if !hitNLAfterNonSpace {
				wsScanLeadingLineChars += readLineWithoutAdvancing(lx.r)
			}
			lx.curFullLine = wsScanLeadingLineChars
		}

		switch action.Type {
		case ActionNone:
			// discard the lexeme (do nothing), then keep lexing
		case ActionScan:
			// return the token
			class := stateClasses[action.ClassID]
			tok = lx.makeToken(class, lexeme)
			retToken = true
		case ActionState:
			// modify state, then keep lexing
			newState := action.State
			lx.state = newState
		case ActionScanAndState:
			// modify state, then return the token

			// doing token creation first in case a state shift alters what is
			// in the token
			class := stateClasses[action.ClassID]
			tok = lx.makeToken(class, lexeme)
			retToken = true

			newState := action.State
			lx.state = newState
		}

		// update source text context tracking
		lx.curLine = curLine
		lx.curPos = curPos
		lx.curFullLine = curFullLine

		// return token if we do that now
		if retToken {
			if lx.listener != nil {
				lx.listener(tok)
			}
			return tok
		}
	}
}

// Peek returns the next token in the stream without advancing the stream.
func (lx *lazyTokenStream) Peek() Token {
	// preserve all parts of the lexer that might change during a call to Next()
	// so we can restore it afterward
	lx.r.Mark("peek")
	oldState := lx.state
	oldFullLine := lx.curFullLine
	oldLine := lx.curLine
	oldPos := lx.curPos
	oldDone := lx.done
	oldPanic := lx.panicMode

	// run lexing as normal:
	tok := lx.Next()

	// restore original data
	lx.r.Restore("peek")
	lx.state = oldState
	lx.curFullLine = oldFullLine
	lx.curLine = oldLine
	lx.curPos = oldPos
	lx.done = oldDone
	lx.panicMode = oldPanic

	// and finally, return the token
	return tok
}

// HasNext returns whether the stream has any additional tokens.
func (lx *lazyTokenStream) HasNext() bool {
	return !lx.done
}

func (lx *lazyTokenStream) makeToken(class TokenClass, lexeme string) Token {
	return lexerToken{
		class:   class,
		line:    lx.curFullLine,
		linePos: lx.curPos,
		lineNum: lx.curLine,
		lexed:   lexeme,
	}
}

func (lx *lazyTokenStream) makeEOTToken() Token {
	return lx.makeToken(TokenEndOfText, "")
}

func (lx *lazyTokenStream) makeErrorTokenf(formatMsg string, args ...any) Token {
	msg := fmt.Sprintf(formatMsg, args...)
	return lx.makeToken(TokenError, msg)
}

// token for read error takes the given error returned from an I/O operation,
// sets state on lx based on whether the error is io.EOF or some other error,
// then returns a token appropriate for the error, either one of class
// types.TokenEndOfText for io.EOF or types.TokenError for all other errors.
func (lx *lazyTokenStream) tokenForIOError(err error) Token {
	lx.done = true

	if err == io.EOF {
		lx.panicMode = false
		return lx.makeEOTToken()
	}
	return lx.makeErrorTokenf("I/O error: %s", err.Error())
}

// select match from slice of all regex matches. If there is exactly 1 match,
// return that. assumes that the first element of candidates is a 'full match'
// and therefore useless, and that blank entries in subsequent indexes indicates
// non-match.
//
// Returns the index of the action associated with the match, and the match
// itself.
func (lx *lazyTokenStream) selectMatch(candidates []string) (int, string) {
	// we now have our list of matches. which sub-expression(s) matched?
	// (and consider a blank match to be 'no match' at this time)

	// toss them all into a 'sparse array' at their index-1 so they have
	// direct correspondance to the index of the action they imply.
	subExprMatches := map[int]string{}
	for i := 1; i < len(candidates); i++ {
		if candidates[i] != "" {
			subExprMatches[i-1] = candidates[i]
		}
	}

	// do we have a conflict between two lexemes? if so, do gnu lex style
	// resolution: prefer the longer one, and if all are equal, prefer the
	// one first defined.
	if len(subExprMatches) > 1 {
		// find the longest length
		var longest int
		for i := range subExprMatches {
			m := subExprMatches[i]
			runeCount := utf8.RuneCountInString(m)
			if runeCount > longest {
				longest = runeCount
			}
		}

		// eliminate all but the longest length one(s)
		keep := map[int]string{}
		for i := range subExprMatches {
			m := subExprMatches[i]
			runeCount := utf8.RuneCountInString(m)
			if runeCount == longest {
				keep[i] = m
			}
		}
		subExprMatches = keep

		// do we still have multiple matches? if so, take the first one
		// defined (with the lowest index)
		if len(subExprMatches) > 1 {

			// need to scan for lowest index because iteration order is not
			// guaranteed
			lowestIndex := math.MaxInt
			for i := range subExprMatches {
				if i < lowestIndex {
					lowestIndex = i
				}
			}

			// just grab that one and put it into a new map
			keep := map[int]string{
				lowestIndex: subExprMatches[lowestIndex],
			}
			subExprMatches = keep
		}
	}

	// we now have exactly one candidate match in our map, so iteration will
	// give us this value

	var matchIndex int
	var matchText string
	for i := range subExprMatches {
		matchIndex = i
		matchText = subExprMatches[i]
		break
	}

	return matchIndex, matchText
}
