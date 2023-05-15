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

Users of ictcc are not advised to attempt this, but the -C flag can be useful
for reading in a file using subshell redirection, if desired:

```
./ictcc -C "$(<some-spec.md)"
```

## Compiler Algorithm Selection

## Using The Generated Frontend

## Command Reference
ictcc produces compiler frontends written in Go for languages specified with the
FISHI specification langauge.

Usage:

	ictcc [flags] file1.md file2.md ...

