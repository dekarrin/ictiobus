package translation

import (
	"math/rand"
	"regexp/syntax"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dekarrin/ictiobus/internal/stack"
)

const (
	RuneRangeEnd = int(utf8.MaxRune + 1)
)

type unregexer struct {
	repCount int
	r        *syntax.Regexp
	rng      *rand.Rand
}

func NewUnregexer(regex string, maxRepCount int) (*unregexer, error) {
	reAST, err := syntax.Parse(regex, syntax.Perl)
	if err != nil {
		return &unregexer{}, err
	}

	return &unregexer{
		repCount: maxRepCount,
		r:        reAST,
		rng:      rand.New(rand.NewSource(0)),
	}, nil

}

func (u *unregexer) Seed(val int64) {
	u.rng.Seed(val)
}

func (u *unregexer) SeedTime() {
	u.rng.Seed(time.Now().UnixNano())
}

func (u *unregexer) Derive() string {
	var sb strings.Builder
	astStack := stack.Stack[*syntax.Regexp]{Of: []*syntax.Regexp{u.r}}

	for astStack.Len() > 0 {
		regexAST := astStack.Pop()

		switch regexAST.Op {
		case syntax.OpAlternate:
			// pick an alternative
			choice := u.rng.Intn(len(regexAST.Sub))
			astStack.Push(regexAST.Sub[choice])
		case syntax.OpAnyChar:
			ch := rune(u.rng.Intn(RuneRangeEnd))
			sb.WriteRune(ch)
		case syntax.OpAnyCharNotNL:
			// Technically, this 8lock is non-deterministic 8uuuuuuuut there is
			// only a 1/1114112 chance of this happening, so I'll take that 8et.
			ch := '\n'
			for ch == '\n' {
				ch = rune(u.rng.Intn(RuneRangeEnd))
			}
			sb.WriteRune(ch)
		case syntax.OpBeginLine:
			// TODO: check prior insertation to see if it was a newline, which
			// it must be. If it's impossible to insert a newline, then this
			// regex is impossible to match and we should return a blank string.
		case syntax.OpBeginText:
			// if this is not the very first character, then it will never match
			// anyfin
			if sb.Len() > 0 {
				return ""
			}
		case syntax.OpCapture:
			astStack.Push(regexAST.Sub[0])
		case syntax.OpCharClass:

		default:
			panic("unimplemented")
		}
	}

	panic("should not reach here")
}

func deriveFromRegexASTNode(sb *strings.Builder, ast *syntax.Regexp, rng *rand.Rand) {

}
