package fishi

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/dekarrin/ictiobus/parse"
	"github.com/dekarrin/ictiobus/trans"
)

// File cgstructs.go contains structs used as part of code generation.

// CodegenOptions is options used for code generation.
type CodegenOptions struct {
	// DumpPreFormat will dump the generated code before it is formatted.
	DumpPreFormat bool

	// TemplateFiles is a map of template names to a path to a custom template
	// file for that template. If entries are detected under the key of one of
	// the ComponentX constants, the path in it is parsed as a template file and
	// used for outputting the generated code for that component instead of the
	// default embedded template.
	TemplateFiles map[string]string

	// IRType is the fully-qualified type of the intermediate representation in
	// the frontend. This is used to make the Frontend function return a
	// specific type instead of requiring an explicit type instantiation when
	// called.
	IRType string

	// PreserveBinarySource is whether to keep the source files for any
	// generated binary after the binary has been successfully
	// compiled/executed. Normally, these files are removed, but preserving them
	// allows for diagnostics on the generated source.
	PreserveBinarySource bool
}

// GeneratedCodeInfo contains information about the generated code.
type GeneratedCodeInfo struct {
	// MainFile is the path to the main executable file, relative to Path.
	MainFile string

	// Path is the location of the root of the generated code.
	Path string
}

// mainBinaryParams is paramters for generating a main.go file for a binary.
// Unless otherwise specified, all fields are required.
type mainBinaryParams struct {
	// Parser is the parser to use for the generated compiler.
	Parser parse.Parser

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

	// LocalIctiobusSource is used to specify a local path to ictiobus to use
	// instead of the currently published latest version. This is useful for
	// debugging while developing ictiobus itself.
	LocalIctiobusSource string
}

// DiagBinParams are parameters for the generation of diagnostic binaries.
type DiagBinParams struct {
	// Parser is the built parser of the frontend to be validated.
	Parser parse.Parser

	// HooksPkgDir is the path to the directory containing the hooks package.
	HooksPkgDir string

	// HooksExpr is the expression to use to get the hooks map. This can be a
	// function call, constant name, or var name.
	HooksExpr string

	// PathPrefix is a prefix to apply to the paths of generated source files.
	// If empty, the current directory will be used.
	PathPrefix string

	// Opts are options for code generation. This must be set and its IRType
	// field is required to be set, but all other fields within it are optional.
	Opts CodegenOptions

	// LocalIctiobusSource is used to specify a local path to ictiobus to use
	// instead of the currently published latest version. This is useful for
	// debugging while developing ictiobus itself.
	LocalIctiobusSource string

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

	// BinPath is the path to the binary to create.
	BinPath string
}

// SimulatedInputParams are parameters for simulating input on a generated
// parser.
type SimulatedInputParams struct {
	// Parser is the built parser of the frontend to be validated.
	Parser parse.Parser

	// HooksPkgDir is the path to the directory containing the hooks package.
	HooksPkgDir string

	// HooksExpr is the expression to use to get the hooks map. This can be a
	// function call, constant name, or var name.
	HooksExpr string

	// PathPrefix is a prefix to apply to the paths of generated source files.
	// If empty, the current directory will be used.
	PathPrefix string

	// LocalIctiobusSource is used to specify a local path to ictiobus to use
	// instead of the currently published latest version. This is useful for
	// debugging while developing ictiobus itself.
	LocalIctiobusSource string

	// Opts are options for code generation. This must be set and its IRType
	// field is required to be set, but all other fields within it are optional.
	Opts CodegenOptions

	// ValidationOpts are options for executing the validation itself. This can
	// be nil and if so will be treated as an empty struct.
	ValidationOpts *trans.ValidationOptions

	// WarningHandler is the current warning handler and is queried to see which
	// warning fatal/suppression options should be passed to the simulation
	// binary.
	WarningHandler *WarnHandler

	// QuietMode is whether quiet mode should be enabled in the simulation
	// execution.
	QuietMode bool
}

type codegenTemplate struct {
	tmpl    *template.Template
	outFile string
}

// codegen data for template fill of main.go
type cgMainData struct {
	BinPkg            string
	BinName           string
	Version           string
	Lang              string
	HooksPkg          string
	HooksTableExpr    string
	ImportFormatPkg   bool
	TokenPkgName      string
	FrontendPkgImport string
	FormatPkg         string
	FormatCall        string
	FrontendPkg       string
	IRTypePackage     string
	IRType            string
	IRIsBuiltInType   bool
	IncludeSimulation bool
}

// codegenData for template fill.
type cgData struct {
	FrontendPackage   string
	Lang              string
	Version           string
	IRAttribute       string
	IRType            string
	IRPackage         string
	TokenPkgName      string
	FrontendPkgImport string
	Command           string
	CommandArgs       string
	Classes           []cgClass
	Patterns          cgPatterns
	Rules             []cgRule
	Bindings          []cgBinding
}

func (cgd cgData) String() string {
	var sb strings.Builder

	sb.WriteString("codegenData{\n")
	sb.WriteString(fmt.Sprintf("  FrontendPackage:   %q\n", cgd.FrontendPackage))
	sb.WriteString(fmt.Sprintf("  Lang:              %q\n", cgd.Lang))
	sb.WriteString(fmt.Sprintf("  Version:           %q\n", cgd.Version))
	sb.WriteString(fmt.Sprintf("  IRAttribute:       %q\n", cgd.IRAttribute))
	sb.WriteString(fmt.Sprintf("  IRType:            %q\n", cgd.IRType))
	sb.WriteString(fmt.Sprintf("  IRPackage:         %q\n", cgd.IRPackage))
	sb.WriteString(fmt.Sprintf("  TokenPkgName:      %q\n", cgd.TokenPkgName))
	sb.WriteString(fmt.Sprintf("  FrontendPkgImport: %q\n", cgd.FrontendPkgImport))
	sb.WriteString(fmt.Sprintf("  Command:           %q\n", cgd.Command))
	sb.WriteString(fmt.Sprintf("  CommandArgs:       %q\n", cgd.CommandArgs))

	// classes
	sb.WriteString("  Classes:           [")
	if len(cgd.Classes) < 1 {
		sb.WriteString("]\n")
	} else {
		sb.WriteString("\n")
		for i := range cgd.Classes {
			sb.WriteString(fmt.Sprintf("    %s", cgd.Classes[i].String()))
			if i+1 < len(cgd.Classes) {
				sb.WriteRune(',')
			}
			sb.WriteRune('\n')
		}
		sb.WriteString("  ]\n")
	}

	// patterns
	sb.WriteString("  Patterns:          {\n")
	sb.WriteString("    (default state): {\n")
	for i := range cgd.Patterns.DefaultState.Classes {
		sb.WriteString(fmt.Sprintf("      %s\n", cgd.Patterns.DefaultState.Classes[i].String()))
	}
	for i := range cgd.Patterns.DefaultState.Entries {
		sb.WriteString(fmt.Sprintf("      %s\n", cgd.Patterns.DefaultState.Entries[i].String()))
	}
	sb.WriteString("    }\n")
	for i := range cgd.Patterns.NonDefaultStates {
		st := cgd.Patterns.NonDefaultStates[i]
		sb.WriteString(fmt.Sprintf("    %q: {\n", st.State))
		for j := range st.Classes {
			sb.WriteString(fmt.Sprintf("      %s\n", st.Classes[j].String()))
		}
		for j := range st.Entries {
			sb.WriteString(fmt.Sprintf("      %s\n", st.Entries[j].String()))
		}
		sb.WriteString("    }\n")
	}
	sb.WriteString("  }\n")

	// rules

	sb.WriteString("  Rules:             [")
	if len(cgd.Rules) < 1 {
		sb.WriteString("]\n")
	} else {
		sb.WriteString("\n")
		for i := range cgd.Rules {
			sb.WriteString(fmt.Sprintf("    %s", cgd.Rules[i].String()))
			sb.WriteRune('\n')
		}
		sb.WriteString("  ]\n")
	}

	// bindings
	sb.WriteString("  Bindings:          [")
	if len(cgd.Bindings) < 1 {
		sb.WriteString("]\n")
	} else {
		sb.WriteString("\n")
		for i := range cgd.Bindings {
			bind := cgd.Bindings[i]
			for j := range bind.Productions {
				sb.WriteString(fmt.Sprintf("    %s: %s", bind.Head, bind.Productions[j].String()))
				sb.WriteRune('\n')
			}
		}
		sb.WriteString("  ]\n")
	}

	sb.WriteRune('}')

	return sb.String()
}

type cgPatterns struct {
	DefaultState     cgStatePatterns
	NonDefaultStates []cgStatePatterns
}

type cgStatePatterns struct {
	State   string
	Classes []cgClass
	Entries []cgPatternEntry
}

type cgPatternEntry struct {
	Regex    string
	Action   string
	Priority int
}

func (cgpe cgPatternEntry) String() string {
	return fmt.Sprintf("<P%d /%s/ %s>", cgpe.Priority, cgpe.Regex, cgpe.Action)
}

type cgBinding struct {
	Head        string
	Productions []cgSDTSProd
}

func (cgb cgBinding) String() string {
	s := fmt.Sprintf("<%s [", cgb.Head)
	for i := range cgb.Productions {
		s += cgb.Productions[i].String()
		if i+1 < len(cgb.Productions) {
			s += ", "
		}
	}
	s += "]>"
	return s
}

type cgSDTSProd struct {
	Symbols     []string
	Attribute   string
	Hook        string
	Args        []cgArg
	Synthetic   bool
	ForRelType  string
	ForRelIndex int
}

func (cgsp cgSDTSProd) String() string {
	prodStr := strings.Join(cgsp.Symbols, ", ")
	if prodStr == "" {
		prodStr = "ε"
	}

	arStr := fmt.Sprintf("{%s$%d}.%s", cgsp.ForRelType, cgsp.ForRelIndex, cgsp.Attribute)

	attrType := "S"
	if !cgsp.Synthetic {
		attrType = "I"
	}

	hookCall := cgsp.Hook + "("
	for i := range cgsp.Args {
		hookCall += cgsp.Args[i].String()
		if i+1 < len(cgsp.Args) {
			hookCall += ", "
		}
	}
	hookCall += ")"

	return fmt.Sprintf("<%s [%s]: %s = %s>", attrType, prodStr, arStr, hookCall)
}

type cgArg struct {
	RelType   string
	RelIndex  int
	Attribute string
}

func (cga cgArg) String() string {
	return fmt.Sprintf("{%s$%d}.%s", cga.RelType, cga.RelIndex, cga.Attribute)
}

type cgRule struct {
	Head        string
	Productions []cgGramProd
}

func (cgr cgRule) String() string {
	var sb strings.Builder

	sb.WriteString(cgr.Head)
	sb.WriteString(" -> ")

	if len(cgr.Productions) == 0 {
		sb.WriteString(" (no productions defined)")
		return sb.String()
	}

	for i := range cgr.Productions {
		sb.WriteString(cgr.Productions[i].String())
		if i+1 < len(cgr.Productions) {
			sb.WriteString(" | ")
		}
	}

	return sb.String()
}

type cgGramProd struct {
	Symbols []string
}

func (cggp cgGramProd) String() string {
	prodStr := strings.Join(cggp.Symbols, ", ")
	if prodStr == "" {
		prodStr = "ε"
	}
	return "[" + prodStr + "]"
}

type cgClass struct {
	Name  string
	ID    string
	Human string
}

func (cgc cgClass) String() string {
	return fmt.Sprintf("(%s %s %q)", cgc.Name, cgc.ID, cgc.Human)
}
