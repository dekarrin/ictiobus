package parse

import (
	"testing"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/stretchr/testify/assert"
)

func Test_constructDFAForLALR1(t *testing.T) {
	testCases := []struct {
		name        string
		grammar     string
		expect      string
		expectStart string
	}{
		{
			name: "2-rule ex from https://www.cs.york.ac.uk/fp/lsa/lectures/lalr.pdf",
			grammar: `
				S -> C C ;
				C -> c C | d ;
			`,
			expect: `<START: "{C -> . c C, c, C -> . c C, d, C -> . d, c, C -> . d, d, S -> . C C, $, S-P -> . S, $}", STATES:
	(({C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d} [=(C)=> {C -> c C ., $, C -> c C ., c, C -> c C ., d}, =(c)=> {C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d}, =(d)=> {C -> d ., $, C -> d ., c, C -> d ., d}])),
	(({C -> . c C, $, C -> . d, $, S -> C . C, $} [=(C)=> {S -> C C ., $}, =(c)=> {C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d}, =(d)=> {C -> d ., $, C -> d ., c, C -> d ., d}])),
	(({C -> . c C, c, C -> . c C, d, C -> . d, c, C -> . d, d, S -> . C C, $, S-P -> . S, $} [=(C)=> {C -> . c C, $, C -> . d, $, S -> C . C, $}, =(S)=> {S-P -> S ., $}, =(c)=> {C -> . c C, $, C -> . c C, c, C -> . c C, d, C -> . d, $, C -> . d, c, C -> . d, d, C -> c . C, $, C -> c . C, c, C -> c . C, d}, =(d)=> {C -> d ., $, C -> d ., c, C -> d ., d}])),
	(({C -> c C ., $, C -> c C ., c, C -> c C ., d} [])),
	(({C -> d ., $, C -> d ., c, C -> d ., d} [])),
	(({S -> C C ., $} [])),
	(({S-P -> S ., $} []))
>`,
		},
		{
			name: "purple dragon 'efficient' LALR construction grammar",
			grammar: `
					S -> L = R | R ;
					L -> * R | id ;
					R -> L ;
			`,
			expect: `<START: "{L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, S -> . L = R, $, S -> . R, $, S-P -> . S, $}", STATES:
	(({L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =} [=(*)=> {L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =}, =(L)=> {R -> L ., $, R -> L ., =}, =(R)=> {L -> * R ., $, L -> * R ., =}, =(id)=> {L -> id ., $, L -> id ., =}])),
	(({L -> * R ., $, L -> * R ., =} [])),
	(({L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, S -> . L = R, $, S -> . R, $, S-P -> . S, $} [=(*)=> {L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =}, =(L)=> {R -> L ., $, S -> L . = R, $}, =(R)=> {S -> R ., $}, =(S)=> {S-P -> S ., $}, =(id)=> {L -> id ., $, L -> id ., =}])),
	(({L -> . * R, $, L -> . id, $, R -> . L, $, S -> L = . R, $} [=(*)=> {L -> * . R, $, L -> * . R, =, L -> . * R, $, L -> . * R, =, L -> . id, $, L -> . id, =, R -> . L, $, R -> . L, =}, =(L)=> {R -> L ., $, R -> L ., =}, =(R)=> {S -> L = R ., $}, =(id)=> {L -> id ., $, L -> id ., =}])),
	(({L -> id ., $, L -> id ., =} [])),
	(({R -> L ., $, R -> L ., =} [])),
	(({R -> L ., $, S -> L . = R, $} [=(=)=> {L -> . * R, $, L -> . id, $, R -> . L, $, S -> L = . R, $}])),
	(({S -> L = R ., $} [])),
	(({S -> R ., $} [])),
	(({S-P -> S ., $} []))
>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			assert := assert.New(t)
			g := grammar.MustParse(tc.grammar)

			// execute
			actual, err := constructDFAForLALR1(g)
			if !assert.NoError(err) {
				return
			}

			// assert
			assert.Equal(tc.expect, actual.String())
		})
	}

}
