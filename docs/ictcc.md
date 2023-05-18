# ictcc Manual

`ictcc` is the "Ictiobus Compiler-Compiler"; it is a command line tool that
reads a specification for a language and produces Go code for a compiler
frontend that can accept input written in that language.

In the general case, one or more files written in the [FISHI](fishi-usage.md)
language are read into ictcc, their contents are concatenated together, and they
are together interpreted as a single language specification. Once a
specification is successfully read, further actions as specified by CLI flags
will be done; by default a lexer, parser, and translation scheme is generated
for it, and output as part of a Go package which unifies them all and provides
access to their functionality via a single `Frontend()` function. This function
will return a type (a `github.com/dekarrin/ictiobus.Frontend`) that has an
`Analyze(io.Reader)` function that parses input in the specified language to its
configured intermediate value (IR).

Depending on the complexity of the language and how it is specified, the IR may
itself be a final value with no further calculation needed; for example, a
language that defines mathematical expressions could evaluate them on the fly
during the translation phase and make the IR be the value of the entire
expression. Some languages may instead work better by making their IR an
abstract syntax tree or some other representation that maintains the structure
of the original source code. Whatever is decided on to be the IR as determined
from the FISHI spec will be used as the return value of a `Frontend`'s `Analyze`
function.

## Requirements

Ictiobus does not have any additional requirements to generate Go source on any
system it is built for. But for some features, it requires a local Go build
environment to be available for use, and in particular the `go` command. Both
the language simulation feature and diagnostics binary creation feature require
this.

This is needed due to Go's lack of support for dynamically loaded libraries; at
this time, because Go cannot be instructed to load an external library without
it being available at the time it is built, the only way to run code that
depends on dynamic external code (such as would be specified with the --hooks
flag) is to copy all of the external code into a new project and compile it,
which ictcc does automatically when simulating language input or generating a
diagnostics binary.

## Reading Input

All input to ictcc is provided as files formatted as "FISHI markdown files"; the
specifics of this format are laid out in the [FISHI manual](fishi-usage.md), but
in general a FISHI markdown file is simply a markdown file that contains
codeblocks labeled with `fishi` immediately after the opening triple-tick.

When ictcc reads a FISHI markdown file, it ignores all content except for the
code contained in those special FISHI codeblocks. These are processed in the
order they are in the file to make up a complete spec.

To process all FISHI codeblocks in a file, simply pass the name of the file to
ictcc:

```
ictcc fishi-spec.md
```

Multiple files can be specified by giving multiple filenames:

```
ictcc spec1.md spec2.md spec3.md
```

They will each be individually parsed and if successful, their results will be
combined into a single spec. If reading any of the files fails, ictcc will
consider the entire spec malformed, and while it will attempt to read any
remaining files to report any issues with them, it will immediately exit after
that without attempting any further actions. Note that this has a slight caveat
for the exit code of ictcc; it will reflect the success of the last file
processed. That is, in the above example, if reading spec2.md were to fail due
to there being a syntax error in spec2.md, and if reading spec3.md were to
succeed, then ictcc would exit with code 0 after printing an error message for
spec2.md. This is not desirable behavior and will likely be patched in the
future; for now, if the exit code is relied on for reading multiple files, it is
recommended to manually cat them into a single file just before processing.

Input can be read from stdin by giving the filename "-":

```
cat spec.md | ictcc -
```

If stdin and files are specified to be read from, reading stdin follows the same
ordering rules as any other file; that is, any files given before it are read
first, then stdin is read, then any files given after it are read.

For example, in the following invocation, first spec1.md is read, then spec2.md,
then stdin-spec.md via stdin, then spec4.md:

```
cat stdin-spec.md | ictcc spec1.md spec2.md - spec4.md
```

ictcc also supports directly giving input with the --command/-C argument,
although note that the input must be markdown-formatted, which may be cumbersome
to produce. However, assuming properly formatted input is provided in it, the
FISHI markdown in the -C will be executed:

```
./ictcc -C "$(printf '```fishi\n%%%%tokens\n\d+  %%token int\n%%%%grammar\n{S} = {S} int | int\n```\n')"
```

The -C flag can be useful for reading in a file using subshell redirection (if
supported by the shell):

```
./ictcc -C "$(<some-spec.md)"
```

### Spec Restrictions

Some restrictions may not be immediately apparent when using FISHI to define
specs. These are listed in full here.

#### S-Attributed SDTS Required

For syntax-directed translation schemes, at this time Ictiobus supports only
S-attributed attribute grammars: only synthetic attributes are supported, not
inherited attributes. That means that, in the `%%actions` section of a FISHI
spec, attempting to assign the results of a hook to an attribute of any node
beside the head symbol's node (`{^}.something`) is not supported. The inclusion
of such an attribute will result in an error such as "ATTRIBUTE-BEING-ASSIGNED
is an inherited attribute (not an attribute of {^})".

To make this slightly confusing, there appears to be source code within Ictiobus
that supports inherited attributes. This was originally planned for, and may
eventually be re-added, but at this time it is not well-tested and may lead to
extremely bizarre edge cases. To force ictcc to enable this as an experimental
feature, pass --exp inherited-attributes to ictcc.

## Output Control

Ictiobus outputs several messages as it processes specs, giving its progress and
noting warnings that occur during generation. This output can be controlled
using CLI flags.

Quiet mode is enabled by passing the -q/--quiet to ictcc. This will disable all
progress and supplementary output messages. This has no effect on error message
output or warning message output; additionally, output that is specifically
requested via CLI args (such as a spec listing requested with -s) is output
regardless of whether quiet mode is enabled.

Warnings encountered during frontend generation are printed to stderr by
default. There are several categories of warnings, and each may be suppressed or
promoted to a fatal error. To suppress a warning, use the -S/--suppress flag and
give the type of warning to suppress. To promote a warning to a fatal error, use
the -F/--fatal flag and give the type of warning to promote. Multiple -S and -F
flags may be given in a single invocation of ictcc. If both -S and -F are given
for the same warning, -F takes precedence.

Besides a specific warning type, "all" may be given as the argument to a -F or
-S to specify that all warnings regardless of their type should be treated as
fatal or suppressed. The following warning types are available to be
suppressed/fatalized:

* `dupe-human`    - issued when there are multiple different human names defined
                    for the same token in a spec.
* `missing-human` - issued when there is no human name defined for some token in
                    a spec.
* `priority`      - issued when a spec marks a lexer action explicitly as having
                    priority 0, the default priority, therefore having no
                    effect.
* `unused`        - issued when a token defined in a spec is never used in any
                    rule of the context-free grammar as a terminal symbol.
* `ambig`         - issued when a grammar results in a parser with an ambiguous
                    parsing decision (LR conflict) for some rule of the grammar.
* `validation`    - issued when a warnable condition occurs during frontend
                    validation.
* `import`        - issued when the correct import for generated code cannot be
                    inferred due to outputting the Go package into a directory
                    not within a Go module, GOPATH, or GOROOT.
* `val-args`      - issued when validation cannot be performed due to a missing
                    --hook or --ir flag.

The prefix for all generated code can be set using the --prefix flag. Note that
this does not also change where generated diagnostics binaries are placed. In
general, unless doing work on ictcc itself, it makes more sense to directly
control the output destination with the --dest flag.

To disable source code generation entirely, use the -n/--no-gen flag. Note that
this only applies to the Go source code generated as output from ictcc; it will
not stop generation used for the diagnostics binary if --diag is enabled, nor
will it stop generation of the simulation binary if valid flags for it are
provided (--ir, --hooks, --hooks-table) and --sim-off is not passed to ictcc.

## Generated Code

Unless the -n/--no-gen flag is specified, invoking ictcc will produce a compiler
frontend as its main artifact. This consists of several Go source code files
that are placed in one or more Go packages rooted in a single directory. This
package provides several functions to give access to the entire frontend and the
components of it. Most users of ictiobus will simply call the `Frontend()`
function, which combines all the components into a single frontend type.

The directory that the generated files are placed in can
be set with --dest; by default it will be a directory called 'fe' in the current
working directory. The name of the Go packages can also be altered with the
--pkg flag; by default, the primary frontend package name will be "fe". There
will also be a sub-package located inside that contains all of the generated
token classes. This package will be named the same as the frontend package, but
suffixed with "token" (so "fetoken" by default).

The default directory tree of output files will look something like the
following:

    (directory ictcc was invoked from)
    |-- fe
    |   |-- fetoken
    |   |   \-- tokens.ict.go
    |   |
    |   |-- frontend.ict.go
    |   |-- lexer.ict.go
    |   |-- parser.cff
    |   |-- parser.ict.go
    |   \-- sdts.ict.go
    |
    \...(any other pre-existing files/dirs)

All generated Go source file names end in .ict.go, and should be relatively
human-readable. Files that end in .cff are binary data files in "compiled FISHI
format" - these hold a component of the frontend that is pre-compiled and
encoded in an internal binary format called 'REZI' and are not human-readable.

The output Go package contains several functions that return components of the
generated frontend:

* `fe.Lexer(bool)` returns the generated lexer. If true is given, the lexer will
perform *lazy* lexing, that is, when a token is requested of it, it will only
consume enough input to return the next token, continuing only once another
token is requested.
* `fe.Parser()` returns the generated parser.
* `fe.Grammar()` returns the context-free grammar that the frontend was
generated for.
* `fe.SDTS()` returns the generated syntax-directed translation scheme.
* `fe.Frontend(trans.HookMap, *FrontendOptions)` (or
`fe.Frontend[E](trans.HookMap, *FrontendOptions)`, depending on how it was
generated), returns the complete frontend immediately ready to be used. It
unifies all components into a single type, plugs in the interface to the
implementations of the translation scheme's hook functions, sets any options
requested, and returns the ready-to-use Frontend object.

Inside of the frontend package, there will be the tokens package. This contains
all of the token classes for the language as well as a single `ByID()` function
which will return the token class that was created for the given text token
class ID. The ID will match the name of the token class given in the FISHI spec
the frontend was generated from. The tokens package is separate from the main
one for historical reasons and may eventually be remerged back into the frontend
package.

The frontend package's `Frontend()` function has a signature that varies
depending on how ictcc was invoked. By default, since the returned type of
Frontend() is type-parameterized with respect to the type of the IR it returns,
the function will itself be type-parameterized and require the caller to know
the IR type and pass it in with the call to Frontend. On the other hand, if this
type is passed to ictcc with the --ir parameter, then it will have enough
information to fill the parameter in itself, and the generated `Frontend()` will
not require the caller to provide it at runtime.

Regardless of whether it was generated with type parameterization, `Frontend()`
requires two arguments. The first is a table that maps the names of hooks used
in the syntax-directed translation to their implementations. Here is an example
of the definition of such a table, referring to unexported functions in the same
package it is in:

```go
var (
	HooksTable = trans.HookMap{
		"int":          hookInt,
		"identity":     hookIdentity,
		"add":          hookAdd,
		"mult":         hookMult,
		"lookup_value": hookLookupValue,
	}
)

// an example function to show the signature of hook implementations
func hookInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    str, _ := args[0].(string)
    return strconv.Atoi(str)
}

// ... more hook functions would be below or elsewhere in the package
```

The second argument to `Frontend()` is a `FrontendOptions` pointer. This is an
options object that contains all options to control the behavior of the
frontend; this includes enabling debugging information, lazy lexer selection,
and any other features that can be tweaked. It can be set to `nil` to use the
default options (no lexer debug output, no parser debug output, lazy lexer
enabled).

Once `Frontend()` is successfully called, it will return a
`github.com/dekarrin/ictiobus.Frontend`. This `Frontend` struct is ready for
immediate use with no further configuration; all of that was handled by the call
to `Frontend()`. It can immediately be used to analyze input and parse it into
an IR value by calling `Frontend.Analyze(io.Reader)`. This will return three
items: the IR value, the parse tree from the parsing stage, and an error. Note
that even if the error is non-nil, the parse tree may itself still be valid, if
for instance the parse succeeded but the syntax-directed translation had an
error. The parse tree will be non-nil whenever it is valid.

The returned `Frontend` has a few properties that refer to the language it was
built for. `Frontend.Language` is set from the value of -l/--lang-name passed to
ictcc at generation time. `Frontend.Version` is similarly set from the value of
-v/--lang-ver.

Both `Frontend.Analyze` and `Frontend.AnalyzeString` will return in-depth syntax
errors with detailed information if it encounters an issue while lexing,
parsing, or translating the input source text. The returned error in that case
will be of type `*github.com/dekarrin/ictiobus/syntaxerr.Error`, and it can be
used with a checked cast.

The following is a complete example of a main function that obtains a frontend
from the generated package and then uses it to try and parse input:

```go
func main() {

    // hookspkg is a user-defined package where a hooks table is defined.
    hooksTable := hookspkg.HooksTable

    // fe is the name of the ictcc-generated package
    var feOpts *fe.FrontendOptions

    // Obtain a frontend that was set up to parse and interpret input to an
    // int IR value.
    scriptEngine := fe.Frontend[int](hooksTable, feOpts)

    // alternatively, if we had used --ir int in ictcc while generating this,
    // then Frontend() will not be parameterized:
    // scriptEngine := fe.Frontend(hooksTable, feOpts)

    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "ERROR: need to provide a file to execute\n")
        return
    }

    f, err := os.Open(os.Args[1])
    if err != nil {
        fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
        return
    }
    defer f.Close()

    // 2nd return value is the parse tree, which we don't really care about
    // right here.
    value, _, err := scriptEngine.Analyze(f)
    if err != nil {
        if syntaxErr, ok := err.(*syntaxerr.Error); ok {
            // if it's an ictiobus syntax error, display the detailed message
            // to the user:
            fmt.Fprintf(os.Stderr, "ERROR: %s\n", syntaxErr.FullMessage())
            return
        } else {
            panic(err)
        }
    }

    fmt.Printf("Result from file: %v\n", value)
}
```

## Parser Algorithm Selection

Ictiobus is capable of producing several different types of parsers, each with a
different algorithm for construction, parsing itself, or both. Each algorithm
has its own restrictions for what types of grammars it can be used with as well
as the size in memory of the parser itself. Additionally, some algorithms may
result in a parser with different worst-case parsing performance than other
algorithms, although at this time all algorithms supported by ictcc run in O(n).

A parsing algorithm may be manually selected by users of ictcc by passing in the
appropriate CLI flag. By default, if no parser algorithm is specified, ictcc
will attempt to automatically select one; each will be tried in the order listed
below, which is in general from most restrictive in which grammars they can
accept to least, and from smallest footprint in memory to largest, until one is
found that can parse the grammar in the spec.

* LL(k), selected with --ll. The *L*eft-to-right, *L*eftmost derivation parser
is a top-down parsing algorithm that is relatively restrictive in the grammars
it is able to parse. It is known to result in small parsers that have a fairly
fast construction time.
* SLR(k), selected with --slr. Also known as the simple LR(k) parser. The
*S*imple *L*eft-to-right, *R*ightmost derivation (in reverse) parser builds a
DFA from sets of LR items of a grammar and uses that to determine actions to
take when parsing. It is known to result in parsers that are often larger than
LL parsers and are much slower to construct, but can accept many more languages
than them.
* LALR(k), selected with --lalr. The *L*ook-*A*head *L*eft-to-right, *R*ightmost
derivation (in reverse) parser also builds a DFA from sets of LR items of a
grammar, but it uses a more complex construction than SLR parsers. It accepts
almost as many languages as a CLR parser, and often has significantly less of a
memory footprint due to a merging algorithm it applies during DFA construction.
* CLR(k), selected with --clr. Also known as the canonical LR(k) parser. The
*C*anonical *L*eft-to-right, *R*ightmost derivation (in reverse) parser uses the
same algorithm as LALR to build the initial DFA, but does not do the merging
afterwords that LALR does. As a result, the CLR parser can accept the most
languages of all algorithms listed here, but takes up the most space in memory.

Many parsing algorithms have a 'k' in their names; this stands for the number of
lookahead tokens from input that it uses to decide how to parse it. At the time
of this writing, ictcc can only produce parsers whose k = 1. For futureproofing
purposes, it is guaranteed that if ictcc ever becomes capable of higher values
of k, it will always select the lowest one required to build a parser for that
algorithm.

Automatic selection of the parsing algorithm may be slow; because of theoretical
restrictions, the problem of whether a particular type of parser that accepts a
particular grammar can be constructed can often only be answered by fully
constructing the parser and then testing it for validity. This means that, for
instance, if a grammar is parsable by a CLR parser, but not by an LL, SLR, or
LALR parser, ictcc would first try to construct every other type of parser as it
goes down the list before finally arriving at the CLR parser. Because of this,
it may be desirable to allow automatic algorithm selection to run only when
making changes to the grammar, and once ictcc finds the one that works, it can
be manually selected for future executions.

## Ambiguity Resolution

Some of the parsers available from Ictiobus are able to apply certain rules to
resolve ambiguities in the grammar. Any that are LR (CLR(k), SLR(k), LALR(k))
can handle certain ambiguous cases.

There are two types of parsing conflicts that can occur for LR-parsers as a
result of grammar ambiguity: shift-reduce, and reduce-reduce. Explaining these
in depth is beyond the scope of this manual; consider consulting relevant
literatures for LR-parsers.

If a shift-reduce conflict occurs, by default LR parsers generated by ictcc will
select the shift action, reading ahead additional input. If a reduce-reduce
conflict occurs, there are underlying issues with the grammar and it must be
re-written to address this before it can be accepted by that parser.

Shift-Reduce conflict resolution can be disabled by passing the --no-ambig flag
to ictcc. This will make any ambiguity in the grammar result in ictcc exiting
with an error regardless of which parsing algorithm is selected.

Sometimes a conflict occurs due to the type of the LR parser itself; if this is
the case, it's possible that selecting a different LR parser may resolve the
issue. Otherwise, the ambiguity in the grammar which caused the issue will need
to be resolved by hand.

## Debugging Specs

When creating a new programming language, a variety of issues can be
encountered, such as inadvertently creating ambiguous grammars, strange DFAs,
and unexpected parsing actions, to name a few. To aid with the debugging of
language specifications, ictiobus provides several tools.

As part of the process of creating an LR parser, a deterministic finite
automaton is constructed from the input language's grammar. Some of the errors
LR parsers report will refer to the states of this DFA. The -D/--dfa flag to
ictcc will make it output the DFA for examination. Output will include each
state of the DFA, with its name, the LR items that the state is associated with,
and a list of transitions from that state to other states. While somewhat
complicated to look over, it can be helpful to trace a parser's path through a
DFA to see where things have gone wrong.

Not every type of parser ictiobus supports uses a DFA, but all of them use a
parsing table. This table informs the parser what action it should take based on
the next token of input it sees; for LL(k) parsers, this is which grammar rule
to select, and for LR parsers, this is whether to shift, reduce to some symbol,
accept the input string, or error. This table can be printed by passing ictcc
the -T/--parse-table flag.

Both the DFA and the parse table output will have symbols that were not directly
defined by the language spec (and in fact, are reserved and forbidden from being
used as grammar symbols):

* `$`   - The end-of-input token/terminal symbol. Lexers in ictiobus return this
          token when at the end of an input string. This as a symbol in debug
          output should be considered as "the end of input has been reached".
* `ε`   - The empty string epsilon symbol. This indicates either the empty
          string in the general case, or more specifically in debug output, an
          epsilon production of a grammar rule.
* `*-P` - The augmented start symbol (`*` will be replaced by the actual name of
          the starting rule of the grammar the parser is for). LR parsers detect
          the end of input by building DFAs from an augmented version of the
          original input grammar; this is done by defining a new start symbol
          which derives the original start symbol, whose rule remains in the
          grammar.

Besides outputting information on parser construction, ictcc can output the spec
itself as it sees it, at various stages. The -P/--preproc flag will cause it to
output the spec source code for input files after it has completed preprocessing
steps, such as removing comments and normalizing lines of input. To see the spec
as a listing of definitions after ictcc has finished reading it but before it is
used for code generation, use the -s/--spec flag.

If you are attempting to read a spec, and you receive and error such as
"<SOME-ATTRIBUTE-REF> is an inherited attribute (not an attribute of {^})", you
may be attempting to load a spec with a non S-attributed SDTS. This is not
supported at this time; to fix this, adjust your spec's `%%action` section to
only assign results of hooks to the head symbol, "{^}". If translation steps
in the SDTS require the use of inherited attributes, you may need to move such
processing outside of the SDTS and use the SDTS to build up a different IR that
retains the information necessary to do this (such as an abstract syntax tree).

Beyond the above output options and tips, Ictiobus provides two methods for
validating specs that can be useful: language simulation and diagnostic binary
generation. These both allow the built frontend to be quickly and easily tested
on input completely independently of other Go code.

### Language Simulation

When a new frontend is generated from a spec, Ictiobus can attempt to
automatically validate that it works properly for possible inputs. It does this
by taking the grammar and using it to derive a series of possible valid inputs
and then feeding this into different stages of the frontend. Ictiobus will
generate at least one input for every possible derivation in the grammar.

This "language simulation" primarily validates that the translation scheme given
in the spec will be able to handle all types of input. Of course, it cannot
necessarily predict all incorrect behavior of the translation hooks, but it will
be able to at minimum detect whether a hook results in a value that is never
used, or whether a translation scheme results in an impossible evaluation loop.

This validation works by taking a generated frontend and using it to generate a
new binary that includes both the frontend and the implementation of all hooks
as specified by the user, then using that binary to run the simulated input.
This binary is known as the "simulation binary". Full binary generation is
needed because using the frontend requires hook implementations, which are in
code external to ictiobus and known only at ictcc runtime. Due to Go having no
cross-platform mechanism for dynamically loading external libraries, ictcc opts
to simply copy the hooks implementation wholesale and build it into a single
binary. This same method is also used for generating a diagnostics binary; they
are closely related. When language simulation completes, the simulation binary
is discarded.

Language simulation is enabled by default, but requires some CLI flags to be set
before it can proceed. Failure to set these flags will return a warning message
unless simulation is explicitly disabled with the --sim-off flag:

* The --ir flag sets the type of the intermediate representation that the
frontend will return, which is required so that the simulation binary can invoke
the generated `Frontend()` function correctly. To specify a type available
without importing, it can be set to that type directly, such as "int". For a
type that requires an import, its package must be included and it must be fully
qualified with the package import path, such as "*crypto/x509/pkix.Name", or
"github.com/dekarrin/ictiobus/fishi/syntax.AST" for example. The IR type may be
a pointer or slice type; simply include the relevant symbols before the name.
* The --hooks flag gives the path to a directory of a Go package containing a
hooks table (of type `trans.HookMap`). This hooks table is obtained by searching
for an exported var at package scope named `HooksTable`, but this name can be
changed by setting --hooks-table.
* The --hooks-table flag is optional and only needed if the package in --hooks
names its hooks table variable something besides `HooksTable`. If so, the
--hooks-table flag is how ictcc is informed of the name.

In addition to the above flags, other flags are available that control the
behavior of language simulation. As mentioned previously, --sim-off disables
language simulation entirely even when --ir and --hooks are set.

Additional diagnostic output for simulations can be selected with the
--sim-trees and --sim-graphs flags. If simulation finds an error in the SDTS
caused by one of the derived simulated parse trees that were fed to it, setting
--sim-trees will make ictcc print out the tree in full. The --sim-graphs flag
will cause any problematic dependency graphs generated by applying the SDTS to
the simulated input to be printed out.

Because error output from language simulation can often include multiple errors
with the same root cause, and because diagnostic output can be quite long,
options are provided to control which of the simulation errors are displayed.
The --sim-first-err flag makes it so only the first error from simulation is
printed, should there be multiple. This combines nicely with --sim-skip-errs,
which can be set to the number of errors to skip in output before the first
error is printed. So to for example see only the 3rd error in the output, one
could set --sim-skip-errs 2 --sim-first-err.

### Diagnostic Binary

During execution of ictcc, it can create a binary that includes the generated
frontend, the implementation of SDTS hooks, and a wrapper main function for
reading input that contains text in the language the frontend was generated for.
This is known as a "diagnostic binary", and it is enabled by specifying a path
to output the binary to with the -d/--diag flag.

This binary is generated in an almost identical fashion as a language simulation
binary, and as a result requires the same flags:

* --ir to set the type of the intermediate representation that the frontend will
return when given an input language. If it's a basic type, such as "int", then
it can be set to that directly. If it's a type that has to be imported, it must
include the package the type is in, fully qualified by its import path, such as
"github.com/dekarrin/neatlang/syntax.AST" for a type called 'AST' in a package
called "syntax" imported with "github.com/dekarrin/neatlang/syntax".
* --hooks to set the path to a directory containing a Go package containing the
hooks table to use to implement SDTS hooks (of type `trans.HookMap`). The hooks
table is located in the package by looking for an exported var named
`HooksTable`, but the name of the var to be used can be changed by specifying
the actual name with --hooks-table.

Once a diagnostic binary is produced, it can be executed on input written in the
source language in the same ways FISHI specs can be passed to ictcc. The
following examples assume a diagnostic bin named "diagbin" was produced during a
prior run of ictcc; in reality it will be named whatever was given as the
filename in the path used for --diag.

The diagnostic binary can read input files:

```shell
./diagbin input-file.txt

=== Analysis of input-file.txt ===
28
```

It can read from multiple input files:

```shell
./diagbin input-file1.txt input-file2.txt

=== Analysis of input-file1.txt ===
28
=== Analysis of input-file2.txt ===
413
```

Code can be directly interpreted with -C:

```shell
./diagbin -C "2+3"

5
```

Code can be read from stdin by specifying filename "-":

```shell
echo "2+3" | ./diagbin -

5
```

More fine-grained examination of the parsing process is also possible. The
-t/--tree flag will cause the parse tree(s) created from the input(s) to be
printed to stdout before they are sent to the SDTS phase for translation. A
detailed log of the tokens found by the lexer can be printed by enabling lexer
debug mode with the -l/--debug-lexer flag. The parser supports a similar output
mode, although it tends to be a bit more verbose than the lexer's; this is
enabled with the -p/--debug-parser flag.

By default, a diagnostics binary expects to receive UTF-8 encoded text that is
accepted by the grammar. If certain preprocessing steps generally are done to
input text to convert it from a typical format to text acceptable by the
grammar, (such as comment stripping, decoding source from wrapper formats,
etc.), the diagnostics binary can be configured to handle this. To do this,
ictcc needs to be given a path to a package that provides an io.Reader which can
handle decoding, formatting, and any other pre-processing that must be done.
This is done by specifying a path to a Go package with the -f/--diag-format-pkg
flag. This package must contain a function that accepts an io.Reader and returns
an io.Reader that returns only source code. By default, this function must have
a signature of "NewCodeReader(io.Reader) (io.Reader, error)", but the name of
the function used can be changed with the --c/--diag-format-call flag. The first
returned value of the function does not necessarily need to directly be
"io.Reader", but it must be a type that implements it.

If preprocessing is enabled using the above process, the diagnostics binary will
have an additional flag, the -P/--preproc flag, which will cause input sent to
it to be printed out after preprocessing has been applied to it but before it is
sent to the lexer.

The diagnostics binary supports an alternative execution mode where instead of
reading input, it runs language input simulation. This is the same simulation
that is normally automatically executed by ictcc. It is enabled by passing the
--sim flag to the diagnostics binary, and besides that supports all of the same
flags starting with --sim- that ictcc does, and they affect input simulation the
same way that they do when passed to ictcc; refer to the Language Simulation
section of this manual for more information.

If the diagnostics binary is to be used only for validating whether input can be
parsed, its -q/--quiet flag can be used to enable quiet mode. This will suppress
all non-error output, including outputting the IR value, and a successful parse
can be checked for by examining the diagnostics binary's exit code.

The diagnostics binary supports warning suppression and fatalization using the
same warning types and CLI options as ictcc. See the Output Control section of
this manual for more information.

## Development on Ictcc

The ictcc binary contains several flags that can assist with development on
ictcc itself and the FISHI language, which is self-hosted. In the first place,
as a general rule, always enable simulation. If any build of the FISHI frontend
is not simulated, it could mask major issues.

The ictcc command supports the same --debug-lexer and --debug-parser flags as
diagnostics binaries; while these aren't particularly useful for end-user usage
of ictcc, they can be helpful in debugging changes to FISHI. The ictcc command
also supports the -t/--tree flag to output the parse trees of the input files.

The AST of FISHI files read can be output by using the -a/--ast flags. This is
the intermediate value used for the FISHI self-hosted frontend, and it may be
helpful when trying to debug issues where the FISHI parses just fine, but the
post-frontend conversion to a `fishi.Spec` does not properly function.

The --prefix flag specifies the path prefix of any output files; this includes
binaries, source generation dirs (for creating binaries), and the output
generated Go code. Combining this with --preserve-bin-source, which will cause
ictcc to keep Go source files generated only to create a simulation or
diagnostics binary, can help to check binary code when it fails to build due to
issues with the generated code.

Templates for generating Go source files are located in fishi/templates. These
templates are brought into the ictcc binary by embedding their contents entirely
within variables; this means that compiled versions of ictcc have their
templates built in to the ictcc binary. To use a different template instead
(such as a current version of a template file, checked out in source), the path
to the new template file can be passed to ictcc using one of the following
flags: --tmpl-main to give a template for the main file used in generated
binaries, --tmpl-frontend to give a template for frontend.ict.go, --tmpl-lexer
to give a template for lexer.ict.go, --tmpl-parser to give a template for
parser.ict.go, --tmpl-sdts to give a template for sdts.ict.go, and --tmpl-tokens
to give a template for tokens.ict.go.

To see filled templates during codegen but before they are sent to gofmt for
formatting (where Go syntax errors will be detected if they are present), use
the --debug-templates flag.

For many of the generated binaries that need to refer to ictiobus code, ictcc
will pull from the current published latest release version of ictiobus. This is
correct for general use, but when developing ictcc one often needs it to pull
code from the current cloned repo on disk. Use the --dev to mark the current
working directory as the location to pull ictiobus code from for these purposes.
This will have the effect of making ictcc cease to function if it is called from
any directory besides one containing the ictiobus module.

## Command Reference

```
ictcc produces compiler frontends written in Go from frontend specifications
written in FISHI.

Usage:

	ictcc [flags] FILE ...

Ictcc reads in the provided FISHI code, either from a file specified as its
args, from CLI flag -C, from stdin by specifying file "-", or some combination
of the above. All FISHI read is combined into a single spec, which is then used
to generate a compiler frontend that is output as Go code.

All input must be UTF-8 encoded markdown-formatted text that contains code
blocks marked with the label `fishi`; only those codeblocks are read for FISHI
source code. The contents of all such codeblocks for an input are concatenated
together to form the "FISHI part" of an input. This concatenated series of FISHI
statements then has comment stripping and line normalization applied to it
before it is parsed into an AST.

When all inputs have been successfully parsed, their ASTs are joined into a
single one by concatenation in the order the inputs they were parsed from were
given, and that AST is then interpreted into a language spec.

This language spec is then used to create a lexer, parser, and then translation
scheme for the language described in the spec. The parser algorithm will be the
one specified by CLI flags; otherwise, the most restrictive one supported that
can handle the grammar is used.

If the --ir and --hooks options are provided, the generated frontend is then
validated by building it into a simulation binary which then simulates language
input against the frontend, covering every possible production in the grammar.
Any issues found at this stage are output; otherwise, the binary and its sources
are deleted.

The Go code for the generated frontend is then placed in a local directory;
"./fe" by default, which can be changed with --dest. The name of the package it
is placed in, "fe" by default, can be changed with the --pkg flag. The language
metadata, retrievable from the generated frontend, can be set by using the -l
and -v flags.

If an error occurs while parsing any of the FISHI, ictcc will still try to parse
any remaining input files for error reporting purposes, but will ultimately fail
to produce generated code. All files must contain parsable FISHI.

Flags:

    -a, --ast
        Print the AST of successfully read FISHI files to stdout.

    -c, --diag-format-call NAME
        Call the function called NAME in the package given by --diag-format-pkg
        when obtaining a code io.Reader in a generated diagnostics binary. This
        is "NewCodeReader" by default. --diag-format-call has no effect unless
        --diag-format-pkg and --diag are also set.

    --clr
        Generate a Canonical LR(k) parser. Mutually exclusive with --ll, --slr,
        and --lalr.

    -C, --command CODE
        Read the FISHI markdown document in CODE before any other input is read.

    -d, --diag FILE
        Generate a diagnostics binary from the spec and output it to the path
        FILE. This binary will contain a self-contained version of the generated
        frontend and can be used to validate it by attempting to use it to parse
        input files in the language the frontend was generated for. This flag
        requires the --ir and --hooks flags to also be set. By default the
        generated binary will not do any preprocessing of input files; to enable
        it, use the --diag-format-pkg flag.

    --debug-lexer
        Print each token as it lexed from FISHI input.

    --debug-parser
        Print each step the parser takes as it parsers FISHI input.

    --debug-templates
        Dump templates after they are filled for codegen but before they are
        formatted by gofmt, along with line numbers for easy reference.

    --dest PATH
        Place the generated Go files in a package rooted at PATH. The default
        value is "./fe".

    --dev
        Enable the use of and reference to ictiobus code located in the current
        working directory as it is currently written as opposed to using the
        latest release version of ictiobus.

    -D, --dfa
        Print a detailed representation of the DFA that is constructed for the
        generated parser to stdout.

    --exp FEATURE
        Enable experimental or untested feature FEATURE. The allowed values for
        FEATURE are as follows for this version of ictcc: "inherited-attributes"
        and "all".

    -f, --diag-format-pkg PATH
        Enable format reading in generated diagnostic binary specified with
        --diag by using the io.Reader provided by the Go package located at
        PATH. This package must provide a function that matches the signature
        "NewCodeReader(io.Reader) (io.Reader, error)", though the returned type
        can be any type that implements io.Reader. The name of that function can
        be selected with --diag-format-call. --diag-format-pkg has no effect
        unless --diag is also set.

    -F, --fatal WARNTYPE
        Treat WARNTYPE warnings as fatal. If the specified type of warning is
        encountered, ictcc will output it as though it were an error and
        immediately halt. Valid values for WARNTYPE are "dupe-human",
        "missing-human", "priority", "unused", "ambig", "validation", "import",
        "val-args", "exp-inherited-attributes", and "all". This flag may be
        specified multiple times and in conjunction with -S flags; if both -F
        and -S are specified for a warning, -F takes precedence.

    --hooks PATH
        Retrieve the hooks table binding translation scheme hooks to their
        implementations from the Go package located in the directory specified
        by PATH. The package must contain an exported var named "HooksTable" of
        type trans.HookMap. The name of the var searched for can be set with
        --hooks-table if needed.

    --hooks-table NAME
        Set the name of the exported hooks table variable in the Go package
        located at the path specified by --hooks. NAME must be the name of an
        exported var of type trans.HookMap. The default value is "HooksTable".

    --ir TYPE
        Set the type of the IR returned by the generated frontend to TYPE. TYPE
        must be either an unqualified basic type (such as "int" or "float32"),
        or, if using requires importing a package, the import path of the
        package, followed by a dot, followed by the name of the type (such as
        "github.com/dekarrin/ictiobus/fishi/syntax.AST", or
        "*crypto/x509/pkix.Name"). Packages with a different name than the last
        component of their import path are not supported at this time. Pointer
        types and slice types are both supported; map types are not. If --ir is
        provided, the Frontend() function in the generated Go package will be
        prefilled with this type, making it so callers of Frontend() do not need
        to supply it at runtime.

    -l, --lang NAME
        Set the language name in the metadata of the generated frontend to NAME.
        The default value is "Unspecified".

    --lalr
        Generate an LALR(k) parser. Mutually exclusive with --ll, --slr, and
        --clr.

    --ll
        Generate an LL(k) parser. Mutually exclusive with --lalr, --slr, and
        --clr.

    -n, --no-gen
        Do not output a Go package with source code files that contain the
        generated frontend. If no other options that would cause spec processing
        are provided, this will cause ictcc to stop after the spec has been
        read.

    --no-ambig
        Disallow generation for specs that define an ambiguous context-free
        grammar.

    --pkg NAME
        Set the name of the package the generated Go source files will be placed
        in. The default value is "fe". 

    --prefix PATH
        Prefix the path of all generated source files with PATH. This includes
        source files used as part of creating binaries as well as the output
        directory specified by --dest. This does not affect the location of the
        diagnostics binary specified with --diag.

    --preserve-bin-source
        Do not delete source files that are generated in the process of
        producing a binary (simulation or diagnostics), even if the binary is
        successfully built.

    -P, --preproc
        Show input FISHI after preprocessing is executed on it; this will be the
        FISHI that is directly provided to the lexer after it is gathered from
        codeblocks in the input markdown document.

    -q, --quiet
        Enable quiet mode; do not output progress or supplemantary messages.
        Output specifically requested via other flags or caused by warnings or
        errors is not affected by this flag.

    -s, --spec
        Print a formatted listing of the complete spec out once it is read from
        FISHI input files.

    --sim-first-err
        Print only the first error returned from language input simulation,
        after any that are skipped by --sim-skip-errs.

    --sim-graphs
        Print the full dependency graph info for any issue found during language
        input simulation that involves translation scheme dependency graphs.

    --sim-off
        Disable language input simulation, even if --ir and --hooks flags are
        provided.

    --sim-skip-errs N
        Skip outputting the first N errors encountered during language input
        simulation. Note that simulation errors will still cause ictcc to halt
        generation even if their output is suppressed.

    --sim-trees
        Print the parse trees of any inputs found to cause issues during
        language input simulation.

    --slr
        Generate a Simple LR(k) parser. Mutually exclusive with --ll, --lalr,
        and --clr.

    -S, --suppress WARNTYPE
        Suppress the output of WARNTYPE warnings. If the specified type of
        warning is encountered, ictcc will ignore it. Valid values for WARNTYPE
        are the same as for --fatal. This flag may be specified multiple times
        and in conjunction with -F flags; if both -F and -S are specified for a
        warning, -F takes precedence.

    -t, --tree
        Print the parse tree of successfully parsed FISHI files to stdout.

    --tmpl-frontend FILE
        Use the contents of FILE as the template to generate frontend.ict.go
        with during codegen.

    --tmpl-lexer FILE
        Use the contents of FILE as the template to generate lexer.ict.go with
        during codegen.

    --tmpl-main FILE
        Use the contents of FILE as the template to generate main.go with during
        codegen for binaries.
    
    --tmpl-parser FILE
        Use the contents of FILE as the template to generate parser.ict.go with
        during codegen.

    --tmpl-sdts FILE
        Use the contents of FILE as the template to generate sdts.ict.go with
        during codegen.

    --tmpl-tokens FILE
        Use the contents of FILE as the template to generate tokens.ict.go with
        during codegen.

    -T, --parse-table
        Print the parse table of the parser generated from the spec to stdout.

    -v, --lang-ver VERSION
        Set the language version in the metadata of the generated frontend to
        VERSION. The default value is "v0.0".

    --version
        Print the current version of ictcc and then exit.
```
