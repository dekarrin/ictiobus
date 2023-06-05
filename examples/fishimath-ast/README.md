# Example: FISHIMath With AST IR

This example shows an implementation of the FISHIMath example language. It
includes both a compiler frontend specification in `fm-ast.md` and code for an
executable (called `fmi` for "FISHIMath Interpreter") that uses that frontend.

The specified compiler frontend creates an abstract syntax tree from parsing
FISHIMath code as its intermediate representation. The main `fmi` executable
passes input code to the frontend, then evaluates the expression contained
within the AST.

## Building

Execute the `build-fmi.sh` script located in example. This will generate an
Ictiobus frontend from `fm-ast.md` and place it in a package called `fmfront` in
a directory of the same name, and then build the `fmi` binary that uses that
frontend. Along the way, it will also end up building an Ictiobus diagnostics
binary called `diag-fm`.

## Running

Once the example is built, the `fmi` binary can be executed to evaluate
FISHIMath statements. It can read files containing FISHIMath; give it the
name(s) of the file(s) as arguments to do this:

    ./fmi eights.fm

You can also give FM code directly with one or more -c flags:

    ./fmi -c '8 + 2   <o^><'

Or you can enter a very limited REPL mode by giving no -c flags or file
arguments:

    ./fmi

    FISHIMath REPL
    (end lines with <o^>< to evaluate; type \q to quit)
    ---------------------------------------------------
    ==> 

Besides that, you can specify the initial value of variables for FM code with
the -s flag:

    ./fmi -s myNum=100 -c 'myNum * 2  <o^><' 

    200

The diagnostics binary `diag-fm` can also be executed to test the stages of the
output frontend. More info on Ictiobus diagnostics binaries and detailed
information on their use can be found in the
[ictcc manual](../../docs/ictcc.md).
