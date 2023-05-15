# Ictiobus

![Tests Status Badge](https://github.com/dekarrin/ictiobus/actions/workflows/tests.yml/badge.svg?branch=main&event=push)
[![Go Reference](https://pkg.go.dev/badge/github.com/dekarrin/ictiobus.svg)](https://pkg.go.dev/github.com/dekarrin/ictiobus)

Lexer/parser/translator generator in pure Go. Generates compiler frontends
written in and accessible via Go, exclusively.

Ictiobus is intended to have implementations of the techniques given in the
textbook "Compilers: Principles, Techniques, and Tools", by Aho, Lam, Sethi, and
Ullman (otherwise known as the "Purple Dragon Book"). It is first and foremost
an experimental learning system and secondarily the parser generator used as the
parser for a scripting language in the tunaquest text adventure engine.

The name `ictiobus` comes from the Latin name for the buffalo fish. The buffalo
(as in the non-fish kind) is related to the bison, and `bison` is a popular
parser-generator. That, combined with the fact that the name `ictiobus` doesn't
seem likely to be confused with other tools, is why the name was chosen.

## Using Ictiobus

Ictiobus is used to generate parsers and other tools for languages. This allows
new languages to be created for use within programs written in Go; it lets you
create new scripting languages! In theory, its output could be plugged into a
modern compiler middle end or back end, but this is not the intent; its main
purposes is to help you design and use your own scripting languages.

### Overview: Creating Scripting Languages

Building and using a scripting language with Ictiobus consists of a few steps.
First, you'll need to create a specification for the language. This includes a
context-free grammar that describes the language, the regular expression
patterns that are to be lexed as tokens, and any steps needed to translate
parsed constructs from your program into something your code can analyze, such
as an abstract syntax tree. This specification is then laid out in a markdown
file in the FISHI specification language, which is explained in the
[FISHI manual](docs/fishi-usage.md) and in the [FISHI spec](docs/fishi.md)
itself.

This file is then read by the `ictcc` command, which compiles a compiler
frontend for the language it describes. It outputs Go code with the token
definitions, generated lexer, generated parser, and generated syntax-directed
translation scheme built from the spec. The package it outputs will include a
`Frontend()` function which can be called by outside code to get a Frontend;
from there, `Analyze(io.Reader)` can be called on readers containingg the
scripting language to parse them into the intermediate representation.

For more information on using `ictcc`, invoke it with `-h` to see the help,
check the Go docs page for the ictcc command, or see the
[ictcc manual](docs/ictcc.md).

### Installation
Ictiobus parsers are generated by running the `ictcc` command on markdown files
that contain specially-formatted codeblocks that give a specification for a
programming language. For information on that format, called FISHI, see the
[FISHI manual](docs/fishi-usage.md).

To install `ictcc` on your system, either grab one of the distributions from
the [Releases page](https://github.com/dekarrin/ictiobus/releases/) of the
Ictiobus repository, or run Go install:

```shell
go install github.com/dekarrin/ictiobus/cmd/ictcc@latest
```

### Creating The FISHI Spec
Once `ictcc` is installed on your system, it's time to put together a FISHI spec
that describes the language you want to create. This section gives a brief
overview on the topic; for an in-depth description of the FISHI language and
using it to define specs, see the [FISHI manual](docs/fishi-usage.md).

A FISHI spec is defined in a markdown file with codeblocks containing FISHI
marked with the label "fishi". These are the only parts of the document that
will be read by the parser generator tool, `ictcc`. Within these blocks, you'll
need at least one of each of the three parts of a FISHI spec:

* A `%%tokens` block, which gives definitions of the tokens in your language and
defines the text patterns that the lexer should use to find them in source text.
* A `%%grammar` block, which defines the context-free grammar that the parser
will use to generate a parse tree from input tokens.
* An `%%actions` block, which gives the actions for a syntax-directed
translation to take to convert a parse tree into a final value that you will
receive when calling `Analyze(io.Reader)` on input written in the new language.

Your spec might look like the following example.

neatlang-spec.md:

    # NeatLang Specification

    This is a cool new language made with ictiobus! It's called NeatLang and it
    does simple math.

    ```fishi
    %%tokens
    
    \+                        %token +         %human plus sign '+'
    \*                        %token *         %human multiplication sign '*'
    \(                        %token lp        %human left parenthesis '('
    \)                        %token rp        %human right parenthesis ')'
    \d+                       %token int       %human integer
    [A-Za-z_][A-Za-z_0-9]*    %token id        %human identifier
    
    # ignore whitespace
    \s+                       %discard
    
    
    %%grammar
    
    {SUM}       =   {SUM} + {PRODUCT}  | {PRODUCT}
    {PRODUCT}   =   {PRODUCT} * {TERM} | {TERM}
    {TERM}      =   lp {S} rp | id | int
    
    
    %%actions
    
    %symbol {S}
    -> {S} + {E} : {^}.value = add({0}.value, {2}.value)
    -> {E}       : {^}.value = identity({0}.value)
    
    %symbol {E}
    -> {E} * {F} : {^}.value = mult({0}.value, {2}.value)
    -> {F}       : {^}.value = identity({0}.value)
    
    %symbol {F}
    -> lp {S} rp : {^}.value = identity({1}.value)
    -> id        : {^}.value = lookup_value({0}.$text)
    -> int       : {^}.value = int({0}.$text)
    ```

For the most part, FISHI is self-contained and builds up definitions used in one
section from FISHI code in a prior section; `%%tokens` sections define token
classes to be used as terminal symbols in a grammar, `%%grammar` sections define
non-terminal symbols by using the defined terminal symbols, and `%%actions`
sections will refer to both non-terminal and terminal symbols defined in the
prior section.

However, the `%%actions` section needs a bit more extra tooling before it can
be used. It refers to *hook functions* by names in its syntax-directed
definitions: `add`, `identity`, `mult`, `lookup_value`, and `int`. These
functions must be implemented in Go code and provided to ictiobus by creating a
`github.com/dekarrin/ictiobus/trans.HookMap` called `HookTable` that maps the
names used in the FISHI spec to their implementation functions. Later, when
`ictcc` is called, it will be informed of this code's location using CLI
arguments.

The hook functions for the above spec might look something like the following.

neatlanghooks/neatlanghooks.go:

```go
package neatlanghooks

import (
    "strconv"

    "github.com/dekarrin/ictiobus/trans"
)

var (
	HooksTable = trans.HookMap{
		"int":          hookInt,
		"identity":     hookIdentity,
		"add":          hookAdd,
		"mult":         hookMult,
		"lookup_value": hookLookupValue,
	}
)

func hookInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    str, _ := args[0].(string)
    return strconv.Atoi(str)
}

func hookIdentity(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    return args[0], nil
}

func hookAdd(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    num1, _ := args[0].(int)
    num2, _ := args[1].(int)
    return num1 + num2, nil
}

func hookMult(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    num1, _ := args[0].(int)
    num2, _ := args[1].(int)
    return num1 * num2, nil
}

func hookLookupValue(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    varName, _ := args[0].(string)

    // in a real program, we might use a lookup table to find real value of the
    // variable, as this function name implies. For now, to keep things simple
    // we'll just use the length of the variable name as its value.
    return len(varName), nil
}
```

Note that a spec is not necessarily required to have Go implementations
associated with the hook functions before `ictcc` can be run on it at all. Some
modes of operation only require looking at the FISHI spec, and some don't even
require that an `%%actions` section exist. But Go hook implementations are
required to use automatic SDTS validation when generating a new compiler
frontend, an important step that can greatly reduce testing time. Additionally,
the Go hook implementations are required at run-time to be provided by callers
who wish to obtain a Frontend, even if they are not provided when calling
`ictcc`, so it's a good idea to start thinking about them early in the process.

### Running `ictcc`
The `ictcc` command takes FISHI spec files and performs analysis on them to
ultimately produce a complete frontend for the language in the spec. The files
will be scanned for code fences with a language tag of `fishi`. FISHI is the
Frontend Instruction Specification for Hosting in Ictiobus and is how languages
to build compilers for is are specified.

These files are read in and then a compiler frontend is produced. With no other
options, ictcc will read in the spec file(s), attempt to create the most
restrictive parser possible by trying them in order of most to least
restrictive, and if successful, outputs generated Go code containing the
Frontend into a package called `fe` by default.

This section contains a brief overview of some of the most common usages of the
ictcc command. For a complete guide and reference to using ictcc, be shore to
check the [ictcc manual](docs/ictcc.md), glub.

For instance, to build a compiler frontend for a new language called 'NeatLang'
specified by a FISHI file called 'neatlang-spec.md' while specifying the name
and version of the language:

```shell
$ ictcc neatlang-spec.md --lang NeatLang --lang-ver 1.0
```

Then, ictcc will get to work reading the spec and producing a frontend for it:

```
Reading FISHI input file...
Generating language spec from FISHI...
Creating most restrictive parser from spec...
Successfully generated SLR(1) parser from grammar
WARN: skipping SDTS validation due to missing --ir parameter
Generating compiler frontend in ./fe...
```

#### SDTS Validation

SDTS (syntax-directed translation scheme) validation is an important step for
preparing a frontend that can automatically validate that the provided
translation scheme can handle all possible combinations of grammar constructs in
the language. If you have the Go implementations of the hook functions (see
above section on creating a FISHI spec for an example of one), it is recommended
you provide it to ictcc so the validation process can execute.

You will need to provide two flags to do this. The `--ir` flag specifies the
type that the Frontend will return from parsing the input; SDTS validation
requires knowing the type in order to test the value returned from the SDTS.

But `--ir` by itself isn't quite enough for validation to run; it also needs the
`--hooks` flag. This flag tells ictiobus the path on disk to a Go package that
contains the special hooks table variable which binds names used in the
`%%actions` section of the FISHI spec to their implementations. Validation needs
this by necessity; it cannot evaluate the translation scheme on a parse tree
without knowing which functions to call.

For instance, for the above NeatLang spec, we could give SDTS hook
implementations in a Go package called `neatlanghooks` in the current directory.
The neatlanghooks package would define functions needed and provide bindings of
hook implementations to names in an exported variable, by default called
`HooksTable`. When calling ictcc, we specify the path to `neatlanghooks` and
that our SDTS will produce an int.

```shell
$ ictcc neatlang-spec.md --ir int --hooks neatlanghooks
```

```
Reading FISHI input file...
Generating language spec from FISHI...
Creating most restrictive parser from spec...
Successfully generated SLR(1) parser from grammar
Generating parser simulation binary in .sim...
Simulation completed with no errors
Generating compiler frontend in ./fe...
```

This time `ictcc` has enough information to perform full validation, and does
so with language input simulation. Luckily, we have everything defined nicely in
NeatLang, so the simulation finds no errors. Great! This means that our NeatLang
translation scheme can handle any input the its grammar allows us to throw at it
without crashing.

#### Diagnostics Binary

In addition to outputting the generated Go code, ictcc supports building a
'diagnostics binary' for testing the frontend from the command line. This is a
self-contained executable that can read code in the target language, parse it
for an intermediate value, and return the string representation of that value.

Creating the diagnostics binary requires the same CLI flags as enabling SDTS
validation, `--ir` and `--hooks`. `--ir` gives the type that the compiler
returns as its ultimate value, and `--hooks` gives the path to a Go package
which binds implementations of hooks to their names. See the above section for a
more detailed explanation of the two flags.

The diagnostics binary is created by specifying `--diag` (or just `-d`) and
giving a path for the new binary. For example, if we wanted to create a
diagnostics binary called `nli` (for NeatLang Interpreter, glub) for the
NeatLang language used as an example in this doc, we would do the following:

```shell
$ ictcc neatlang-spec.md --ir int --hooks ./neatlanghooks --diag nli -n
```

(The `-n` flag instructs ictcc not to output the Go code as it normally does. We
used it above because right now we only care about the diagnostics binary, not
the generated code).

Now after ictcc reads the spec file, it will create a new `nli` program:

```
Reading FISHI input file...
Generating language spec from FISHI...
Creating most restrictive parser from spec...
Successfully generated SLR(1) parser from grammar
Generating parser simulation binary in .sim...
Simulation completed with no errors
Format preprocessing disabled in diagnostics bin; set -f to enable
Generating diagnostics binary code in .gen...
Built diagnostics binary 'nli'
(frontend code output skipped due to flags)
```

And that nli program can be used to parse programs written in NeatLang and
interpret them into their value:

```shell
$ echo '5 + 2' > input.txt
$ ./nli input.txt
=== Analysis of input.txt ===
7
```

It can even execute code directly with the -C flag:

```
$ ./nli -C '8 + 2 * 6'
20

$ ./nil -C '(8 + 2) * 6'   # test parentheses grouping
60
```

The diagnostics binary is a powerful tool for testing. For more information on
the diagnostics binary, including how to make it perform preprocessing on input
before executing it or how to manually run SDTS validation, see the appropriate
section in the [ictcc manual](docs/ictcc.md).

## Development
If you're developing on ictiobus, you must have at least Go 1.19 in order to
support generics. Ictiobus relies somewhat heavily on their existence.

### Building

To create a new version of the ictcc command, the primary entrypoint into
frontend generation, call `scripts/build.sh` from bash or a compatible shell:

```bash
$ scripts/build.sh
```

### Scripts
The scripts directory contains several shell scripts useful for development.

* `scripts/all-tests.sh` - Executes all unit tests and assuming they pass, all
integration tests as well.
* `scripts/build.sh` - Builds the `ictcc` command in the root of the repo.
* `scripts/create-dists.sh` - Creates distribution tarballs of `ictcc`. By
default, it will build for Darwin (Mac), Windows, and Linux, all amd64. Two
tarballs per build are created; one denoted as 'latest' and the other denoted as
the actual version of the build. Version info is automatically obtained by
running the currently-built `ictcc` with `--version`.
* `scripts/gendev.sh` - Generates a frontend for the FISHI language itself,
using the FISHI spec in docs/fishi.md. This script will build a new diagnostics
binary called `fishic` in the current directory, and output generated Go code
to a directory called `.testout` in the repo root for examination by the
developer and to avoid replacing the actual live FISHI frontend. Does not
automatically build a new `ictcc` binary to use for the generation; it will use
whatever ictcc is in the repo root, and will fail if one has not yet been built.
* `scripts/genfrontend.sh` - Generates a new live frontend for the FISHI spec
language and replaces the one currently used in ictcc with the new one. This
script *will* build a new version of `ictcc` automatically by calling
`scripts/build.sh`.

