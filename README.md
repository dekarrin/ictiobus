# ictiobus
Lexer/parser generator in pure Go. Generates parsers written in Go, exclusively.

Ictiobus is intended to be implementations of the techniques given in the
textbook "Compilers: Principles, Techniques, and Tools", by Aho, Lam, Sethi, and
Ullman (otherwise known as the "Purple Dragon Book"). It is first and foremost
an experimental learning system and secondarily the parser generator used as the
parser for a scripting language in the tunaquest text adventure engine.

In general, you probably want to use something modern and well-tested for
generating parsers. Specifically which one depends on your language, but Bison,
ANTLR, or PLY are a few good ones that support outputting to widely-used
languages. If you are interested in poking at the fairly messy internals of a
learning project, by all means take a look at `ictiobus`.

## Building
To build the `ictcc` command, execute the `build.sh` script from bash or
compatible shell.

```shell
./build.sh
```

## Running

Ictiobus is invoked by running the `ictcc` command on markdown files. The files
will be scanned for code fences with a langauge tag of `fishi`. FISHI is the
Frontend Instruction Specification for self-Hosting Ictiobus and is the primary
way that languages are specified to build compilers for.

These files are read in and then a compiler is produced. At least in theory. For
now, it only prints an AST out.

```shell
./ictcc test/test.md
```
