package types

import "fmt"

type ParserType string

const (
	ParserLL1   ParserType = "LL(1)"
	ParserSLR1  ParserType = "SLR(1)"
	ParserCLR1  ParserType = "CLR(1)"
	ParserLALR1 ParserType = "LALR(1)"
)

func (pt ParserType) String() string {
	return string(pt)
}

func ParseParserType(s string) (ParserType, error) {
	switch s {
	case ParserLL1.String():
		return ParserLL1, nil
	case ParserSLR1.String():
		return ParserSLR1, nil
	case ParserCLR1.String():
		return ParserCLR1, nil
	case ParserLALR1.String():
		return ParserLALR1, nil
	default:
		return ParserLL1, fmt.Errorf("not a valid ParserType: %q", s)
	}
}
