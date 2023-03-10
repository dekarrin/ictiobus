package parse

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/icterrors"
	"github.com/dekarrin/ictiobus/internal/decbin"
	"github.com/dekarrin/ictiobus/internal/stack"
	"github.com/dekarrin/ictiobus/types"
)

type ll1Parser struct {
	table grammar.LL1Table
	g     grammar.Grammar
	trace func(s string)
}

func (ll *ll1Parser) GetDFA() string {
	return ""
}

func (ll *ll1Parser) RegisterTraceListener(listener func(s string)) {
	ll.trace = listener
}

func (ll *ll1Parser) TableString() string {
	return ll.table.String()
}

func (ll *ll1Parser) MarshalBinary() ([]byte, error) {
	data := decbin.EncBinary(ll.table)
	data = append(data, decbin.EncBinary(ll.g)...)
	return data, nil
}

func (ll *ll1Parser) UnmarshalBinary(data []byte) error {
	n, err := decbin.DecBinary(data, &ll.table)
	if err != nil {
		return fmt.Errorf("table: %w", err)
	}
	data = data[n:]

	_, err = decbin.DecBinary(data, &ll.g)
	if err != nil {
		return fmt.Errorf("g: %w", err)
	}

	return nil
}

// EmptyLL1Parser returns a completely empty LL1 parser, unsuitable for use.
// Generally this should not be used directly except for internal purposes; use
// GenerateLL1Parser to generate one ready for use.
func EmptyLL1Parser() *ll1Parser {
	return &ll1Parser{}
}

// GenerateLL1Parser generates a parser for LL1 grammar g. The grammar must
// already be LL1 or convertible to an LL1 grammar.
//
// The returned parser parses the input using LL(k) parsing rules on the
// context-free Grammar g (k=1). The grammar must already be LL(1); it will not
// be forced to it.
func GenerateLL1Parser(g grammar.Grammar) (*ll1Parser, error) {
	M, err := g.LLParseTable()
	if err != nil {
		return &ll1Parser{}, err
	}
	return &ll1Parser{table: M, g: g.Copy()}, nil
}

func (ll1 *ll1Parser) Type() types.ParserType {
	return types.ParserLL1
}

func (ll1 ll1Parser) notifyPopped(s string) {
	if ll1.trace != nil {
		ll1.trace(fmt.Sprintf("popped %q", s))
	}
}

func (ll1 ll1Parser) notifyPushed(s string) {
	if ll1.trace != nil {
		ll1.trace(fmt.Sprintf("pushed %q", s))
	}
}

func (ll1 *ll1Parser) Parse(stream types.TokenStream) (types.ParseTree, error) {
	symStack := stack.Stack[string]{Of: []string{ll1.g.StartSymbol(), "$"}}
	next := stream.Peek()
	X := symStack.Peek()
	ll1.notifyPopped(X)
	pt := types.ParseTree{Value: ll1.g.StartSymbol()}
	ptStack := stack.Stack[*types.ParseTree]{Of: []*types.ParseTree{&pt}}

	node := ptStack.Peek()
	for X != "$" { /* stack is not empty */
		if strings.ToLower(X) == X {
			stream.Next()

			// is terminals
			t := ll1.g.Term(X)
			if next.Class().ID() == t.ID() {
				node.Terminal = true
				node.Source = next
				symStack.Pop()
				X = symStack.Peek()
				ll1.notifyPopped(X)
				ptStack.Pop()
				node = ptStack.Peek()
			} else {
				return pt, icterrors.NewSyntaxErrorFromToken(fmt.Sprintf("There should be a %s here, but it was %q!", t.Human(), next.Lexeme()), next)
			}

			next = stream.Peek()
		} else {
			nextProd := ll1.table.Get(X, ll1.g.TermFor(next.Class()))
			if nextProd.Equal(grammar.Error) {
				return pt, icterrors.NewSyntaxErrorFromToken(fmt.Sprintf("It doesn't make any sense to put a %q here!", next.Class().Human()), next)
			}

			symStack.Pop()
			ptStack.Pop()
			for i := len(nextProd) - 1; i >= 0; i-- {
				if nextProd[i] != grammar.Epsilon[0] {
					symStack.Push(nextProd[i])
					ll1.notifyPushed(nextProd[i])
				}

				child := &types.ParseTree{Value: nextProd[i]}
				if nextProd[i] == grammar.Epsilon[0] {
					child.Terminal = true
				}
				node.Children = append([]*types.ParseTree{child}, node.Children...)

				if nextProd[i] != grammar.Epsilon[0] {
					ptStack.Push(child)
				}
			}

			X = symStack.Peek()
			ll1.notifyPopped(X)

			// node stack will always be one smaller than symbol stack bc
			// glub, we dont put a node onto the stack for "$".
			if X != "$" {
				node = ptStack.Peek()
			}
		}
	}

	return pt, nil
}
