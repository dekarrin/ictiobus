package fishi

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dekarrin/ictiobus"
	"github.com/dekarrin/ictiobus/fishi/fe"
	"github.com/dekarrin/ictiobus/fishi/format"
	"github.com/dekarrin/ictiobus/fishi/syntax"
	"github.com/dekarrin/ictiobus/types"
)

const (
	// DiagGenerationDir is the name of the directory in which source code for
	// the diagnostic binary is placed when it is generated, prior to
	// compilation.
	DiagGenerationDir = ".gen"

	// SimGenerationDir is the name of the directory in which source code for
	// the simulation binary is placed when it is generated, prior to
	// compilation.
	SimGenerationDir = ".sim"
)

// This type alias is dum8!!!!!!!! Just make the callers import syntax, it's not
// like it's unexported ::::/
type AST = syntax.AST

type Results struct {
	AST  *AST
	Tree *types.ParseTree
}

type Options struct {
	ParserCFF   string
	ReadCache   bool
	WriteCache  bool
	LexerTrace  bool
	ParserTrace bool
}

// ValidateSimulatedInput generates a lightweight compiler with the spec'd
// frontend in a special directory (".sim" in the local directory or in the path
// specified by pathPrefix, if set) and then runs SDTS validation on a variety
// of parse tree inputs designed to cover all the productions of the grammar at
// least once.
//
// If running validation with the test compiler succeeds, it and the directory
// it was generated in are deleted. If it fails, the directory is left in place
// for inspection.
//
// IRType is required to be set in cgOpts.
//
// valOpts is not required to be set, and if nil will be treated as if it were
// set to an empty struct.
//
// No binary is generated as part of this, but source is which is then executed.
// If PreserveBinarySource is set in cgOpts, the source will be left in the
// .sim directory.
//
// localSource is an optional path to a local copy of ictiobus to use instead of
// the currently published latest version. This is useful for debugging while
// developing ictiobus itself.
func ValidateSimulatedInput(spec Spec, md SpecMetadata, params SimulatedInputParams) error {
	pkgName := "sim" + strings.ToLower(md.Language)

	binName := safeTCIdentifierName(md.Language)
	binName = binName[2:] // remove initial "tc".
	binName = strings.ToLower(binName)
	binName = "test" + binName

	outDir := SimGenerationDir
	if params.PathPrefix != "" {
		outDir = filepath.Join(params.PathPrefix, outDir)
	}

	// not setting the format package and call here because we don't need
	// preformatting to run verification simulation.
	genInfo, err := GenerateBinaryMainGo(spec, md, MainBinaryParams{
		Parser:              params.Parser,
		HooksPkgDir:         params.HooksPkgDir,
		HooksExpr:           params.HooksExpr,
		FrontendPkgName:     pkgName,
		GenPath:             outDir,
		BinName:             binName,
		Opts:                params.Opts,
		LocalIctiobusSource: params.LocalIctiobusSource,
	})
	if err != nil {
		return fmt.Errorf("generate test compiler: %w", err)
	}

	err = ExecuteTestCompiler(genInfo, params.ValidationOpts, params.WarningHandler)
	if err != nil {
		return fmt.Errorf("execute test compiler: %w", err)
	}

	if !params.Opts.PreserveBinarySource {
		// if we got here, no errors. delete the test compiler and its directory
		err = os.RemoveAll(genInfo.Path)
		if err != nil {
			return fmt.Errorf("remove test compiler: %w", err)
		}
	}

	return nil
}

func ParseMarkdownFile(filename string, opts Options) (Results, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Results{}, err
	}
	defer f.Close()

	return ParseMarkdown(f, opts)
}

func ParseMarkdown(r io.Reader, opts Options) (Results, error) {
	bufF := bufio.NewReader(r)
	r, err := format.NewCodeReader(bufF)
	if err != nil {
		return Results{}, err
	}

	res, err := Parse(r, opts)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Parse converts the fishi source code read from the given reader into an AST.
func Parse(r io.Reader, opts Options) (Results, error) {
	// get the frontend
	fishiFront, err := GetFrontend(opts)
	if err != nil {
		return Results{}, fmt.Errorf("could not get frontend: %w", err)
	}

	res := Results{}
	// now, try to make a parse tree
	ast, pt, err := fishiFront.Analyze(r)
	res.Tree = pt // need to do this before we return
	if err != nil {
		return res, err
	}
	res.AST = &ast

	return res, nil
}

// GetFrontend gets the frontend for the fishi compiler-compiler. If cffFile is
// provided, it is used to load the cached parser from disk. Otherwise, a new
// frontend is created.
func GetFrontend(opts Options) (ictiobus.Frontend[syntax.AST], error) {
	feOpts := fe.FrontendOptions{
		LexerTrace:  opts.LexerTrace,
		ParserTrace: opts.ParserTrace,
	}

	fishiFront := fe.Frontend(syntax.HooksTable, &feOpts)

	return fishiFront, nil
}

// ParseFQType parses a fully-qualified type name into its package and type
// along with the name of the package. Any number of leading [] and * are
// allowed, but map types are not supported, although types with an underlying
// map type are supported.
//
// Unqualified built-in types are allowed; they will be returned with pkgName
// 'builtin'.
//
// For example, ParseFQType("[]*github.com/ictiobus/fishi.Options") would return
// "github.com/ictiobus/fishi", "[]*fishi.Options", "fishi", "[]*Options", nil.
func ParseFQType(fqType string) (pkg, typeName, pkgName, bareType string, err error) {
	preType := ""
	for strings.HasPrefix(fqType, "[]") || strings.HasPrefix(fqType, "*") {
		if strings.HasPrefix(fqType, "[]") {
			preType += "[]"
			fqType = fqType[2:]
		} else {
			preType += "*"
			fqType = fqType[1:]
		}
	}
	typeParts := strings.Split(fqType, ".")
	if len(typeParts) < 2 {
		// assume they mean 'builtin'.
		typeParts = []string{"builtin", fqType}
	}
	fqPackage := strings.Join(typeParts[:len(typeParts)-1], ".")
	pkgParts := strings.Split(fqPackage, "/")
	pkgName = pkgParts[len(pkgParts)-1]
	irType := preType + pkgName + "." + typeParts[len(typeParts)-1]
	bareType = preType + typeParts[len(typeParts)-1]

	return fqPackage, irType, pkgName, bareType, nil
}

// MakeFQType makes a fully-qualified type name from the given package and type.
// The typeName may include the package name, but it is not required and will be
// ignored in favor of using pkg. Pkg must be the fully-qualified package name.
func MakeFQType(pkg, typeName string) string {
	preType := ""
	for strings.HasPrefix(typeName, "[]") || strings.HasPrefix(typeName, "*") {
		if strings.HasPrefix(typeName, "[]") {
			preType += "[]"
			typeName = typeName[2:]
		} else {
			preType += "*"
			typeName = typeName[1:]
		}
	}

	typeParts := strings.Split(typeName, ".")
	rawTypeName := typeParts[len(typeParts)-1]

	return preType + pkg + "." + rawTypeName
}
