package trans

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/grammar"
	"github.com/dekarrin/ictiobus/internal/textfmt"
)

// AttrRef contains no uncomparable attributes and can be assigned/copied
// directly.
type AttrRef struct {
	Relation NodeRelation
	Name     string
}

// String returns the string representation of an AttrRef.
func (ar AttrRef) String() string {
	return fmt.Sprintf("%s[%q]", ar.Relation.String(), ar.Name)
}

// ResolveSymbol finds the name of the symbol being referred to in a grammar
// production rule. Head is the head symbol of the rule, prod is the production
// of that rule.
//
// If the AttrRef does not refer to any symbol in the rule, a blank string and a
// non-nil error is returned.
func (ar AttrRef) ResolveSymbol(g grammar.Grammar, head string, prod grammar.Production) (string, error) {
	switch ar.Relation.Type {
	case RelHead:
		return head, nil
	case RelNonTerminal:
		ntIndex := -1
		for i := range prod {
			if g.IsNonTerminal(prod[i]) {
				ntIndex++
				if ntIndex == ar.Relation.Index {
					return prod[i], nil
				}
			}
		}
		return "", fmt.Errorf("no %d%s nonterminal in rule production", ar.Relation.Index, textfmt.OrdinalSuf(ar.Relation.Index))
	case RelTerminal:
		termIndex := -1
		for i := range prod {
			if g.IsTerminal(prod[i]) {
				termIndex++
				if termIndex == ar.Relation.Index {
					return prod[i], nil
				}
			}
		}
		return "", fmt.Errorf("no %d%s terminal in rule production", ar.Relation.Index, textfmt.OrdinalSuf(ar.Relation.Index))
	case RelSymbol:
		if ar.Relation.Index >= len(prod) {
			return "", fmt.Errorf("no %d%s symbol in rule production", ar.Relation.Index, textfmt.OrdinalSuf(ar.Relation.Index))
		}
		return prod[ar.Relation.Index], nil
	}

	return "", fmt.Errorf("invalid Relation.Type in AttrRef: %v", ar.Relation.Type)
}

// NodeRelationType is the type of a NodeRelation.
type NodeRelationType int

const (
	RelHead NodeRelationType = iota
	RelTerminal
	RelNonTerminal
	RelSymbol
)

// GoString returns the go string representation of a NodeRelationType.
func (nrt NodeRelationType) GoString() string {
	switch nrt {
	case RelHead:
		return "RelHead"
	case RelTerminal:
		return "RelTerminal"
	case RelNonTerminal:
		return "RelNonTerminal"
	case RelSymbol:
		return "RelSymbol"
	default:
		return fmt.Sprintf("NodeRelationType(%d)", int(nrt))
	}
}

// String returns the string representation of a NodeRelationType.
func (nrt NodeRelationType) String() string {
	if nrt == RelHead {
		return "head symbol"
	} else if nrt == RelTerminal {
		return "terminal symbol"
	} else if nrt == RelNonTerminal {
		return "non-terminal symbol"
	} else if nrt == RelSymbol {
		return "symbol"
	} else {
		return fmt.Sprintf("NodeRelationType<%d>", int(nrt))
	}
}

// NodeRelation is a relation to a symbol in a node of an annotated parse tree.
// It is either the head symbol of the node itself, or one of the symbols in
// the production.
type NodeRelation struct {
	// Type is the type of the relation.
	Type NodeRelationType

	// Index specifies which of the nodes of the given type that the relation
	// points to. If it is RelHead, this will be 0.
	Index int
}

// String returns the string representation of a NodeRelation.
func (nr NodeRelation) String() string {
	if nr.Type == RelHead {
		return nr.Type.String()
	}

	humanIndex := nr.Index + 1
	return fmt.Sprintf("%d%s %s", humanIndex, textfmt.OrdinalSuf(humanIndex), nr.Type.String())
}

// ValidFor returns whether the given node relation refers to a valid and
// existing node when applied to a node in parse tree that is the result of
// parsing production head -> production.
func (nr NodeRelation) ValidFor(head string, prod []string) bool {
	// Refering to the head is refering to the node itself, so is always valid.
	if nr.Type == RelHead {
		return true
	} else if nr.Type == RelSymbol {
		return nr.Index < len(prod) && nr.Index >= 0
	} else if nr.Type == RelTerminal {
		searchTermIdx := nr.Index

		// find the nth terminal
		curTermIdx := -1
		foundIdx := -1
		for i := range prod {
			sym := prod[i]

			if strings.ToLower(sym) != sym {
				continue
			} else {
				curTermIdx++
				if curTermIdx == searchTermIdx {
					foundIdx = i
					break
				}
			}
		}
		return foundIdx != -1
	} else if nr.Type == RelNonTerminal {
		searchNonTermIdx := nr.Index

		// find the nth non-terminal
		curNonTermIdx := -1
		foundIdx := -1
		for i := range prod {
			sym := prod[i]

			if strings.ToLower(sym) != sym {
				continue
			} else {
				curNonTermIdx++
				if curNonTermIdx == searchNonTermIdx {
					foundIdx = i
					break
				}
			}
		}
		return foundIdx != -1
	} else {
		return false
	}
}
