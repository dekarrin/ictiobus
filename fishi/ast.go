package fishi

import (
	"fmt"
	"strings"

	"github.com/dekarrin/ictiobus/fishi/ir"
)

// This file contains a copy of ir.AST for convenience so callers may directly
// get an AST from fishi and to have an interface that doesn't have to pass into
// the inconveniently named 'ir' package.

// It's horri88le engineering, and it duplic8s code, so we might need to remove
// this l8er.
// TODO: proper review! What if we made the *AST* be here, but renamed everyfin
// in ir to no longer use 'AST'-prefixed names for them? We could also stop
// using ir.AST entirely and replace all uses with just returning []ir.ASTBlock
// (or []ir.Block, as it would likely be called). Hmm. Maybe it should be named
// 'syntax' to match the convention established by regexp/syntax? But it really
// isn't syntax itself, it's semantic translations. Otoh, semantic translation
// into an AST really kind of *is* a representation of the syntax. Glub 38T. Not
// shore, review in future glub.

// AST is the abstract syntax tree of a fishi spec.
type AST struct {
	Nodes []ir.ASTBlock
}

func (ast AST) String() string {
	var sb strings.Builder

	sb.WriteRune('<')
	if len(ast.Nodes) > 0 {
		sb.WriteRune('\n')
		for i := range ast.Nodes {
			n := ast.Nodes[i]
			switch n.Type() {
			case ir.BlockTypeError:
				sb.WriteString("  <ERR>\n")
			case ir.BlockTypeGrammar:
				gram := n.Grammar()
				sb.WriteString("  <GRAMMAR:\n")
				for j := range gram.Content {
					cont := gram.Content[j]
					if cont.State != "" {
						sb.WriteString("    <RULE-SET FOR STATE " + fmt.Sprintf("%q\n", cont.State))
					} else {
						sb.WriteString("    <RULE-SET FOR ALL STATES\n")
					}
					for k := range cont.Rules {
						r := cont.Rules[k]
						sb.WriteString("      * " + r.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case ir.BlockTypeTokens:
				toks := n.Tokens()
				sb.WriteString("  <TOKENS:\n")
				for j := range toks.Content {
					cont := toks.Content[j]
					if cont.State != "" {
						sb.WriteString("    <ENTRY-SET FOR STATE " + fmt.Sprintf("%q\n", cont.State))
					} else {
						sb.WriteString("    <ENTRY-SET FOR ALL STATES\n")
					}
					for k := range cont.Entries {
						entry := cont.Entries[k]
						sb.WriteString("      * " + entry.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			case ir.BlockTypeActions:
				acts := n.Actions()
				sb.WriteString("  <ACTIONS:\n")
				for j := range acts.Content {
					cont := acts.Content[j]
					if cont.State != "" {
						sb.WriteString("    <ACTION-SET FOR STATE " + fmt.Sprintf("%q\n", cont.State))
					} else {
						sb.WriteString("    <ACTION-SET FOR ALL STATES\n")
					}
					for k := range cont.Actions {
						action := cont.Actions[k]
						sb.WriteString("      * " + action.String() + "\n")
					}
					sb.WriteString("    >\n")
				}
				sb.WriteString("  >\n")
			}
		}
	}
	sb.WriteRune('>')

	return sb.String()
}
