package fishi

import "github.com/dekarrin/ictiobus"

// File cgstructs.go contains structs used as part of code generation.

// MainBinaryParams is paramters for generating a main.go file for a binary.
// Unless otherwise specified, all fields are required.
type MainBinaryParams struct {
	// Parser is the parser to use for the generated compiler.
	Parser ictiobus.Parser

	// HooksPkgDir is the path to the directory containing the hooks package.
	HooksPkgDir string

	// HooksExpr is the expression to use to get the hooks map. This can be a
	// function call, constant name, or var name.
	HooksExpr string

	// FormatPkgDir is the path to the directory containing the format package.
	// It is completely optional; if not set, the generated main will not
	// contain any pre-formatting code and will assume files are directly ready
	// to be fed into the frontend. Must be set if FormatCall is set.
	FormatPkgDir string

	// FormatCall is the name of a function within the package specified by
	// FormatPkgDir that gets an io.Reader that will run any required
	// pre-formatting on an input io.Reader to get code that can be analyzed by
	// the frontend. Is is optional; if not set, the generated main will not
	// contain any pre-formatting code and will assume files are directly ready
	// to be fed into the frontend. Must be set if FormatPkgDir is set.
	FormatCall string

	// FrontendPkgName is the name of the package to place generated frontend
	// code in.
	FrontendPkgName string

	// GenPath is the path to a directory to generate code in. If it does not
	// exist, it will be created. If it does exist, any existing files in it
	// will be removed will be emptied before code is generated.
	GenPath string

	// BinName is the name of the binary being generated. This will be used
	// within code for showing help output and other messages.
	BinName string

	// Opts are options for code generation. This must be set and its IRType
	// field is required to be set, but all other fields within it are optional.
	Opts CodegenOptions
}
