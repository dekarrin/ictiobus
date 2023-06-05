package fmfront

/*
File automatically generated by the ictiobus compiler. DO NOT EDIT. This was
created by invoking ictiobus with the following command:

    ictcc --slr --ir github.com/dekarrin/ictfishimath_ast/fmhooks.AST -l FISHIMath -v 1.0 -d /home/dekarrin/projects/ictiobus/examples/fishimath-ast/diag-fm --hooks fmhooks --dest /home/dekarrin/projects/ictiobus/examples/fishimath-ast/fmfront --pkg fmfront --dev /home/dekarrin/projects/ictiobus/examples/fishimath-ast/fm-ast.md
*/

import (
	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/trans"

	"fmt"
	"strings"
)

// SDTS returns the generated ictiobus syntax-directed translation scheme for
// FISHIMath.
func SDTS() trans.SDTS {
	sdts := ictiobus.NewSDTS()

	sdtsBindTCFishimath(sdts)
	sdtsBindTCStatements(sdts)
	sdtsBindTCStmt(sdts)
	sdtsBindTCExpr(sdts)
	sdtsBindTCSum(sdts)
	sdtsBindTCProduct(sdts)
	sdtsBindTCTerm(sdts)

	return sdts
}

func sdtsBindTCFishimath(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"FISHIMATH", []string{"STATEMENTS"},
		"ir",
		"ast",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 0}, Name: "stmt_nodes"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"STATEMENTS"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "FISHIMATH", prodStr, err.Error()))
	}
}

func sdtsBindTCStatements(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"STATEMENTS", []string{"STMT", "STATEMENTS"},
		"stmt_nodes",
		"node_slice_prepend",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "stmt_nodes"},
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"STMT", "STATEMENTS"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "STATEMENTS", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"STATEMENTS", []string{"STMT"},
		"stmt_nodes",
		"node_slice_start",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"STMT"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "STATEMENTS", prodStr, err.Error()))
	}
}

func sdtsBindTCStmt(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"STMT", []string{"EXPR", "shark"},
		"node",
		"identity",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"EXPR", "shark"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "STMT", prodStr, err.Error()))
	}
}

func sdtsBindTCExpr(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"EXPR", []string{"id", "tentacle", "EXPR"},
		"node",
		"assignment_node",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 2}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"id", "tentacle", "EXPR"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "EXPR", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"EXPR", []string{"SUM"},
		"node",
		"identity",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"SUM"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "EXPR", prodStr, err.Error()))
	}
}

func sdtsBindTCSum(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"SUM", []string{"PRODUCT", "+", "EXPR"},
		"node",
		"binary_node_add",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 0}, Name: "node"},
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 1}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"PRODUCT", "+", "EXPR"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "SUM", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"SUM", []string{"PRODUCT", "-", "EXPR"},
		"node",
		"binary_node_sub",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 0}, Name: "node"},
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 1}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"PRODUCT", "-", "EXPR"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "SUM", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"SUM", []string{"PRODUCT"},
		"node",
		"identity",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"PRODUCT"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "SUM", prodStr, err.Error()))
	}
}

func sdtsBindTCProduct(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"PRODUCT", []string{"TERM", "*", "PRODUCT"},
		"node",
		"binary_node_mult",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 0}, Name: "node"},
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 1}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"TERM", "*", "PRODUCT"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "PRODUCT", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"PRODUCT", []string{"TERM", "/", "PRODUCT"},
		"node",
		"binary_node_div",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 0}, Name: "node"},
			{Rel: trans.NodeRelation{Type: trans.RelNonTerminal, Index: 1}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"TERM", "/", "PRODUCT"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "PRODUCT", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"PRODUCT", []string{"TERM"},
		"node",
		"identity",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"TERM"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "PRODUCT", prodStr, err.Error()))
	}
}

func sdtsBindTCTerm(sdts trans.SDTS) {
	var err error
	err = sdts.Bind(
		"TERM", []string{"fishtail", "EXPR", "fishhead"},
		"node",
		"group_node",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 1}, Name: "node"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"fishtail", "EXPR", "fishhead"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "TERM", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"TERM", []string{"int"},
		"node",
		"lit_node_int",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"int"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "TERM", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"TERM", []string{"float"},
		"node",
		"lit_node_float",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"float"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "TERM", prodStr, err.Error()))
	}

	err = sdts.Bind(
		"TERM", []string{"id"},
		"node",
		"var_node",
		[]trans.AttrRef{
			{Rel: trans.NodeRelation{Type: trans.RelSymbol, Index: 0}, Name: "$text"},
		},
	)
	if err != nil {
		prodStr := strings.Join([]string{"id"}, " ")
		panic(fmt.Sprintf("binding %s -> [%s]: %s", "TERM", prodStr, err.Error()))
	}
}