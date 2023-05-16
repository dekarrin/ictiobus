# `ictcc` Manual
`ictcc` is the "Ictiobus Compiler-Compiler"; it is a command line tool that
reads a specification for a language and produces Go code for a compiler
frontend that can accept input written in that language.

In the general case, one or more files written in the [FISHI](fishi-usage.md)
language are read into ictcc, their contents are concatenated together, and they
are together interpreted as a single language specification. Once a
specification is successfully read, further actions as specifed by CLI flags
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
expression. Some languages may instead prefer to make their IR be an abstract
syntax tree or some other representation that maintains the structure of the
original source code. Whatever is decided on to be the IR as determined from the
FISHI spec will be used as the return value of a `Frontend`'s `Analyze`
function.

## Requirements

Ictiobus does not have any additional requirements to generate Go source on any
system it is built for. But for some features, it requires a local Go build
environment to be available for use, and in particular the `go` command. Both
the language simulation feature and diagnostics binary creation feature require
this.

This is needed due to Go's lack of support for dynamically loaded libraries; at
this time, because Go cannot be instructed to load an external library wihtout
it being available at the time it is built, the only way to run code that
depends on dynamic external code (such as would be specified with the --hooks
flag) is to copy all of the external code into a new project, and compiles that,
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
spec2.md. This is not desireable behavior and will likely be patched in the
future; for now, if the exit code is relied on for reading multiple files, it is
recommended to manually cat them into a single file just before processing.

Input can be read from stdin by giving the filename "-":

```
cat spec.md | ictcc -
```

If stdin and files are specified to be read from, reading stdin follows the same
ordering rules as any other file; that is, any files given before it are read
first, then stdin is read, then any file given after it are read.

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

The -C flag can be useful for reading in a file using subshell redirection, if
supported by the shell:

```
./ictcc -C "$(<some-spec.md)"
```

## Output Control

Note on suppression

Note on warning promotion

Note on quiet mode.

Note on prefix switch.

## Generated Code

Unless the -n/--no-gen flag is specified, invoking ictcc will produce a compiler
frontend as its main artifact. This consists of several Go source code files
that are placed in one or more Go packages rooted in a single directory. This
package provides several functions to give access to the entire frontend, or to
the components of it. Most users of ictiobus will simply call the `Frontend()`
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
format" - these hold ia component of the frontend that is pre-compiled and
encoded in an internal binary format called 'REZI' and are not human-readable.

The output Go pacakge contains several functions that return components of the
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

Tokens package summary.

Frontend function signature

Function hooks

Frontend options

Using the Frontend

- note on Langauge and Version being basically set by generation -l and -v

Catching Syntax Errors


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
almost as many langauges as a CLR parser, and often has significantly less of a
memory footprint due to a merging algorithm it applies during DFA construction.
* CLR(k), selected with --clr. Also known as the canonical LR(k) parser. The
*C*anonical *L*eft-to-right, *R*ightmost derivation (in reverse) parser uses the
same algorithm as LALR to build the initial DFA, but does not do the merging
afterwords that LALR does. As a result, the CLR parser can accept the most
languages of all algorithms listed here, but takes up the most space in memory.

Many parsing algorithms have a 'k' in their names; this stands for the number of
lookahead tokens from input that it uses to decide how to parse it. At the time
of this writing, ictcc can only produce parsers whose k = 1. For futureproofing
purposes, it is gauranteed that if ictcc ever becomes capable of higher values
of k, it will always select the lowest one required to build a parser for that
algorithm.

Automatic selection of the parsing algorithm may be slow; because of theoretical
restrictions, the problem of whether a paraticular type of parser that accepts
a particular grammar can be constructed can often only be answered by fully
constructing the parser and then testing it for validity. This means that, for
instance, if a grammar is parsable by a CLR parser, but not by an LL, SLR, or
LALR parser, ictcc would first try to construct every other type of parser as it
goes down the list before finally arriving at the CLR parser. Because of this,
it may be desirable to allow automatic algorithm selection to run only when
making changes to the grammar, and once ictcc finds the one that works, manually
setting it for future executions.

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

## Language Simulation

Ictcc uses sim to verify generated. note go is invoked, and requires use of IR
value and hooks value. Also mention --hooks-table

Description of process of derivation. Include limitation wrt to the tree
over-generation

Description of using derived tree on grammar.

Description of using flags to control behavior, and in-depth... debug section
will refer to here.


## Debugging Specs

-T/--parse-table
-D/--dfa
-p/-preproc
-s/--spec output the spec


### Diagnostic Binary

Note go is invoked, and requires ir and hooks. Also mention --hooks-table.

Basic use, for input reading. -C more useful here.

Debugging options for showing the lexer and parser as they go.

Debugging options for seeing parse tree.

Enabling pre-format of input code.

Turning on simulation mode, and using it. Refer to lang sim section for more in
depth info.

## Development on Ictcc

Note flags: --debug-lexer, --debug-parser.

-a/--ast output spec ast
-t/--tree output spec parse tree

--preserve-bin-source.

--debug-templates

--tmpl-* (main, parser, lexer, sdts, tokens, frontend)

## Command Reference
ictcc produces compiler frontends written in Go for languages specified with the
FISHI specification langauge.

Usage:

	ictcc [flags] file1.md file2.md ...

