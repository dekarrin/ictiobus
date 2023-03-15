package unregex

import (
	"log"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testSeed  = int64(8335)
	testCount = 100
)

func Test_Digit_Class(t *testing.T) {
	// setup
	const regex = `\d`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_Number(t *testing.T) {
	// setup
	const regex = `\d+`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_Literal(t *testing.T) {
	// setup
	const regex = `a`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_CharClass_Single(t *testing.T) {
	// setup
	const regex = `[a]`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_CharClass_MultipleExplicit(t *testing.T) {
	// setup
	const regex = `[abc]`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_CharClass_SingleRange(t *testing.T) {
	// setup
	const regex = `[A-Z]`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_CharClass_MultipleRanges(t *testing.T) {
	// setup
	const regex = `[A-Za-z0-9]`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_CharClass_Mixed(t *testing.T) {
	// setup
	const regex = `[aA-Z_012345678#$]`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
	assert.True(false)
}

func Test_CharClass_Single_Negated(t *testing.T) {
	// setup
	const regex = `[^a]`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)
	// make the unregexer not generate chars in the astral planes or specials from basic multilingual plane
	un.AnyCharsMax = 0xffef

	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}

func Test_SomeLongRegex_Emails(t *testing.T) {
	// please do not actually use this to parse or validate emails; the real
	// regex is far more complicated and there are built in lib functions to do
	// this.

	// setup
	const regex = `^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`

	assert := assert.New(t)
	un, err := New(regex)
	if err != nil {
		t.Fatal(err)
	}
	un.Seed(testSeed)

	matcher := regexp.MustCompile(regex)

	for i := 0; i < testCount; i++ {
		// execute
		str := un.Derive()

		// verify
		if !assert.True(matcher.MatchString(str), "iteration %d: regex '%s' does not match derived %q", i, regex, str) {
			return
		}

		log.Printf("iteration %d: %q", i, str)
	}
}
