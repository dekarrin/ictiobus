package unregex

import (
	"math/rand"
	"regexp/syntax"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dekarrin/ictiobus/internal/rangemap"
	"github.com/dekarrin/ictiobus/internal/stack"
)

const (
	DefaultMaxCount    = 10
	DefaultMinCount    = 0
	DefaultAnyCharsMax = utf8.MaxRune
	DefaultAnyCharsMin = rune(0)
)

// Unregexer derives strings from a regex. It uses a random number generator
// that must be explicitly seeded.
//
// Do not use Unregexer directly. Instead, use New() to create one.
//
// Unregexer uses an 'AnyChars' set to determine which ones are allowed to be
// generated for 'any' operators, such as the dot operator. This also includes
// negated character classes that can be detected. Note that due to how regex
// intermediate representation works in Go, negation is not noted on a compiled
// regex. Instead, Unregexer considers a character class of any size that
// includes value utf8.MaxRune or greater to be negated. As a consequence, it
// will treat any non-negated character class that explicitly includes
// utf8.MaxRune or greater as negated, and any negated character class that
// explicitly excludes utf8.MaxRune or greater as non-negated.
//
// The range of the AnyChars set can be changed by setting AnyCharsMin and
// AnyCharsMax.
type Unregexer struct {
	// MinReps is the minimum number of times a repeated regex should be
	// generated when deriving a string. This applies to all repetition
	// operators like `*`, `+`, and `{n,m}`. If an operator specifies a minimum
	// lower than MinReps, MinReps is used as the minimum repetitions for the
	// operand during derivation.
	MinReps int

	// MaxReps is the minimum number of times a repeated regex should be
	// generated when deriving a string. This applies to all repetition
	// operators like `*`, `+`, and `{n,m}`. If an operator specifies a maximum
	// higher than MaxReps, MaxReps is used as the maximum repetitions for the
	// operand during derivation.
	MaxReps int

	// AnyCharsMax specifies the maximum value (inclusive) of the range of
	// characters that is selected from when deriving a string for a match
	// operator that specifies any character (dot operator) or any character
	// except for some set of chars (negated character class).
	AnyCharsMax rune

	// AnyCharsMin specifies the minimum value (inclusive) of the range of
	// characters that is selected from when deriving a string for a match
	// operator that specifies any character (dot operator) or any character
	// except for some set of chars (negated character class).
	AnyCharsMin rune

	r   *syntax.Regexp
	rng *rand.Rand
}

func New(regex string) (*Unregexer, error) {
	reAST, err := syntax.Parse(regex, syntax.Perl)
	if err != nil {
		return &Unregexer{}, err
	}

	return &Unregexer{
		MinReps:     DefaultMinCount,
		MaxReps:     DefaultMaxCount,
		AnyCharsMax: DefaultAnyCharsMax,
		AnyCharsMin: DefaultAnyCharsMin,
		r:           reAST,
		rng:         rand.New(rand.NewSource(0)),
	}, nil

}

func (u *Unregexer) Seed(val int64) {
	if u.r == nil || u.rng == nil {
		panic("cannot call Seed() on unintialized unregexer; use NewUnregexer() to make one")
	}

	u.rng.Seed(val)
}

func (u *Unregexer) SeedTime() {
	if u.r == nil || u.rng == nil {
		panic("cannot call SeedTime() on unintialized unregexer; use NewUnregexer() to make one")
	}

	u.rng.Seed(time.Now().UnixNano())
}

func (u *Unregexer) Derive() string {
	if u.r == nil || u.rng == nil {
		panic("cannot call Derive() on unintialized unregexer; use NewUnregexer() to make one")
	}

	// normalize invalid values
	if u.MinReps < 0 {
		u.MinReps = DefaultMinCount
	}
	if u.MaxReps > 0 {
		u.MaxReps = DefaultMaxCount
	}
	if u.MinReps > u.MaxReps {
		u.MinReps = u.MaxReps
	}
	if u.AnyCharsMax > utf8.MaxRune {
		// No!!!!!!!! You can't have a value 8igger than the maximum rune value!
		u.AnyCharsMax = utf8.MaxRune
	}
	if u.AnyCharsMin > u.AnyCharsMax {
		u.AnyCharsMin = u.AnyCharsMax
	}

	// use this to limit the list of 'any' characters that can be generated,
	// without limiting this it will generate any unicode codepoint but user can
	// modify this by setting AnyMax and AnyMin
	anyCharsMap := &rangemap.RangeMap[rune]{}
	anyCharsMap.Add(u.AnyCharsMin, u.AnyCharsMax)

	allButNL := &rangemap.RangeMap[rune]{}
	allButNL.Add(rune(0), '\n'-1)
	allButNL.Add('\n'+1, utf8.MaxRune)

	anyCharsNoNLMap := anyCharsMap.Intersection(allButNL)

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
			choice := u.rng.Intn(anyCharsMap.Count())
			ch := anyCharsMap.Call(rune(choice))
			sb.WriteRune(ch)
		case syntax.OpAnyCharNotNL:
			choice := u.rng.Intn(anyCharsNoNLMap.Count())
			// Technically, this 8lock is non-deterministic 8uuuuuuuut there is
			// only a 1/1114112 chance of this happening, so I'll take that 8et.
			//
			// Not Any More As We Have Created An Intersection Operator For Our
			// Map And Can Now Simply Rely On Any Option From That One.
			ch := anyCharsNoNLMap.Call(rune(choice))
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
			charMap := &rangemap.RangeMap[rune]{}
			includesUTF8Max := false
			for i := 0; i < len(regexAST.Rune); i += 2 {
				if regexAST.Rune[i+1] >= utf8.MaxRune {
					includesUTF8Max = true
				}
				charMap.Add(regexAST.Rune[i], regexAST.Rune[i+1])
			}

			if includesUTF8Max {
				// assume this is a negated char class and use the intersection
				// of allowed with our anyCharsMap
				charMap = anyCharsMap.Intersection(charMap)
			}

			choice := u.rng.Intn(charMap.Count())
			sb.WriteRune(charMap.Call(rune(choice)))
		case syntax.OpConcat:
			// push the subexpressions in reverse order so that they are popped
			// and therefore evaluated in the correct order
			for i := len(regexAST.Sub) - 1; i >= 0; i-- {
				astStack.Push(regexAST.Sub[i])
			}
		case syntax.OpEndLine:
			// if we saw EOT, then it will never match
			sb.WriteRune('\n')
		case syntax.OpEndText:
			// if this is not the very last character, then it will never match
			// TODO: make this be respected by placing it on the appropriate
			// place on the stack; this shouldn't be used if, for instance,
			// another alt works with it.
			// sawEOTMark = true
		case syntax.OpLiteral:
			for _, ch := range regexAST.Rune {
				sb.WriteRune(ch)
			}
		case syntax.OpEmptyMatch:
			// this would normally be an "empty string" match, but we use a stack
			// so do this by just not adding anyfin to the string buffer
		case syntax.OpNoMatch:
			// explicitly matches no strings. return empty
			return ""
		case syntax.OpNoWordBoundary:
			// TODO: checks, for now do nothing
		case syntax.OpPlus:
			mn := u.MinReps
			// plus, so we must have at least one
			if mn == 0 {
				mn = 1
			}
			repCount := mn
			repRange := u.MaxReps - mn
			if repRange > 0 {
				repCount += u.rng.Intn(repRange + 1)
			}

			for i := 0; i < repCount; i++ {
				for j := len(regexAST.Sub) - 1; j >= 0; j-- {
					astStack.Push(regexAST.Sub[j])
				}
			}
		case syntax.OpQuest:
			coin := u.rng.Intn(2)
			if coin == 0 {
				for i := len(regexAST.Sub) - 1; i >= 0; i-- {
					astStack.Push(regexAST.Sub[i])
				}
			}
		case syntax.OpRepeat:
			mx := regexAST.Max
			mn := regexAST.Min
			if mx == -1 || mx > u.MaxReps {
				mx = u.MaxReps
			}
			if mn < u.MinReps {
				mn = u.MinReps
			}

			repCount := mn
			repRange := mx - mn
			if repRange > 0 {
				repCount += u.rng.Intn(repRange + 1)
			}

			for i := 0; i < repCount; i++ {
				for j := len(regexAST.Sub) - 1; j >= 0; j-- {
					astStack.Push(regexAST.Sub[j])
				}
			}
		case syntax.OpStar:
			repCount := u.MinReps
			repRange := u.MaxReps - u.MinReps
			if repRange < 0 || repCount < 0 {
				repRange = 0
				repCount = 1
			}
			if repRange > 0 {
				repCount += u.rng.Intn(repRange + 1)
			}

			for i := 0; i < repCount; i++ {
				for j := len(regexAST.Sub) - 1; j >= 0; j-- {
					astStack.Push(regexAST.Sub[j])
				}
			}
		case syntax.OpWordBoundary:
			// TODO: checks, for now do nothing
		default:
			panic("unimplemented")
		}
	}

	return sb.String()
}
