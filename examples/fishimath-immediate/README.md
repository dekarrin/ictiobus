# Example: FISHIMath With AST IR

This example shows an implementation of the FISHIMath example language. It
includes a compiler frontend specification in `fm-ast.md`. It can be tested by
running the diagnostics binary generated from the spec.

The specified compiler frontend immediately evaluates input and returns the
result of each statement as its intermediate representation.

## Building

Execute the `build-diag-fm.sh` script located in example. This will generate an
Ictiobus diagnostics binary called `diag-fm`.

## Running

Once the example is built, the `diag-fm` binary can be executed to evaluate
FISHIMath statements. You can give a filename that has FM code in it to execute:

    ./diag-fm eights.fm

Or you can use -h to see additional options:

    ./diag-fm -h
    
More info on Ictiobus diagnostics binaries and detailed information on their use
can be found in the [ictcc manual](../../docs/ictcc.md).
