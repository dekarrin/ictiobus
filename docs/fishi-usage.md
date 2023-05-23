Using FISHI
===========

The Frontend Instruction Specification for Hosted languages on Ictiobus (the
"FISHI language" or "FISHI" for short), is the special language used to define
specifications for languages that Ictiobus can then generate parsers for! Well,
compiler frontends, really, which include the lexer, parser, and translation
scheme.

To use Ictiobus to make your own programming or scripting language, you first
define a spec for it in the FISHI langauge, then run the command ictcc on the
spec and poof! You'll have a whole glubbin langauge parser (compiler frontend)
ready to use in the Go package you tell ictcc to output it to. All your code
needs to do is call the package's `Frontend()` function.

But building a spec is non-trivial! This document is intended as a somewhat
user-friendly manual explaining how to build a complete FISHI spec using an
example language. If you're looking for an example of a FISHI spec, you're in
luck, because FISHI is self-hosted: there's a
[FISHI spec for FISHI itself](./fishi.md), and it's used to generate the FISHI
parser that ictcc uses. Glub.

## FISHIMath

FISHIMath is an example language that this manual will use to demonstrate
complete examples at the ends of important sections. It's a math expressions
language that has special little symbols for some things, where whitespace does
not matter.

Statements in FISHIMath are ended with "the statement shark" (`<o^><`), which is
hungry, but not so hungry it wants to eat more than one statement! Exactly one
is the perfect amount. Oh, but you don't want to pollute the oceans, so make
sure that every statement you put into FISHIMath has a statement shark to clean
it up for you.

```
8 + 2    <o^><
8 * 2.2  <o^><

2 + 6    /* invalid! no shark to eat the statement */
```

FISHIMath has support for variables! They do not need to be declared, they just
need to be used to exist. Variables with no value assigned will be assumed to
have the value 0. Assignment to a new variable is done by using the "value
tentacle" (`=o`) to make the variable on the left "reach out" and grab the value
on the right.

```
x =o 2       <o^><
```

Grouping parentheses are implemented not by the traditional `(` and `)`, but by
the "fish-tail" `>{` and "fish-head" `'}`.

```
>{8+2'} * 3    <o^><
```

Other than that, FISHIMath supports integers and decimal values (as IEEE-754
single-precision floating point values), and the operators `+`, `-`, `/`, and
`*`.

## Note On The Preproccessor

When FISHI is read by ictcc, before it is interpreted by the FISHI frontend, a
preprocessing step is run on input. Since error reporting is handled by the
frontend, which only knows about the preprocessed version of source code, this
means that syntax errors will refer to that modified version instead of directly
to the code that was input. As the preprocessor performs relatively benign
changes, errors are usually easily understandable, but if syntax error output
is confusing, use the -P flag with ictcc to see the exact source code after it
has been preprocessed.

## The Three Phases Of An Ictiobus Compiler Frontend

So, the term "parser" is often used in two distinct ways. Outside of the context
of compilers and translators, it's often used to refer to an entire analysis
system that can on its own take in input and produce a value from it, either the
result of execution or some representation of the input that can be further
processed. Here, "parsing" refers to the entire process of reading in input,
scanning it for recognizable symbols ("tokens"), interpreting the tokens
according to some format, and then using that interpretation to produce the
final value.

But often in the context of compilers and translators, and absolutely within the
context of FISHI and Ictiobus, "parsing" refers to just one of three main phases
in a complete analysis of input. This complete analysis is known as the
*frontend* of a compiler or interpreter, and the final result value from it is
called the intermediate representation of the input, or IR for short. When input
is analyzed by a frontend, it goes through all three phases in its quest to
become an IR!

Diagram of the phases of a frontend:
```
             | I N P U T |
             \-----------/
                   |
                   V
           (text in language)

              "(3+8) * 2"
             
                   |
                   V
               /-------\
              /         \
             | L E X E R |
              \         /
               \-------/
                   |
                   V
                (tokens)

   (lparen "("), (int "3"), (plus "+"),
   (int "8"), (rparen ")"), (times "*"),
               (int "2")
                 
                   |
                   V
              /---------\
             /           \
            | P A R S E R |
             \           /
              \---------/
                   |
                   V
              (parse tree)

                  sum
                   |
                product
                /    |  \ 
             term   "*"  term
            /  |  \       |
         "("  sum  ")"   "2"
             / |  \
         sum  "+"  product
          |           |
       product      term
          |           |
        term         "8"
          |
         "3"

                   |
                   V
  /---------------------------------\
 /                                   \
| T R A N S L A T I O N   S C H E M E |
 \                                   /
  \---------------------------------/
                   |
                   V
     (intermediate representation)
 
                   22
                   
                   |
                   V
             /-----------\
            | O U T P U T |
```

FISHI specs have 3 different types of "sections" in them that correspond to each
of the three phases: the `%%tokens` sections which give all types of token
classes and the patterns to lex them, the `%%grammar` sections which define the
context-free grammar for the language, and the `%%actions` sections which define
actions to apply in the syntax-directed translation scheme.

## FISHI File Structure

A FISHI file is really easy to write. It's markdown! And for the most part, it's
totally ignored by ictcc. The only place that it looks for code is in special
code blocks (the triple ticks) that have the label `fishi`.

    # Example Markdown File

    This is some markdown! The `ictcc` program won't parse this part at all, and
    it will be rendered as normal markdown.

    You can have code blocks of any kind:

    ```go
    func Hello(name string) string
    ```

    ## FISHI Code Blocks

    ...But only code blocks labeled with "fishi" will be interpreted as FISHI:

    ```fishi
    # FISHI source code goes here
    ```

If there's multiple FISHI code blocks in the same file, they will be read and
interpreted in sequence. So for instance, the following example:

    ## Our Tokens

    These are the lexed tokens:

    ```fishi
    %%tokens

    \d+            %token int
    [A-Za-z_0-9]   %token identifier
    ```

    ## Discarded Patterns

    These patterns are discarded by the lexer:

    ```fishi
    \s+            %discard
    ```

Will be interpreted as exactly equivalent to this:

    ## Lexer

    This is the spec for the lexing phase:

    ```fishi
    %%tokens

    \d+            %token int
    [A-Za-z_0-9]   %token identifier
    \s+            %discard
    ```

Embedding FISHI within markdown like this means that FISHI code can live right
next to code that explains it in rich text. Of course, FISHI also supports
comments within FISHI code blocks; the '#' goes from the initial '#' to the end
of the line.

Every FISHI file contains three types of "sections", each of which is started
with a special header. It is completely acceptable to have multiple of the same
type of section, and it's even okay to interleave them!

So, this:

    ```fishi
    %%tokens

    \d+            %token int
    ```

    And some other text

    ```fishi
    %%grammar

    {S} = {S} {E} | {E}
    ```

    ```fishi
    %%tokens

    [A-Za-z_0-9]   %token identifier
    ```

Is interpreted the same as this:

    ```fishi
    %%tokens

    \d+            %token int
    ```

    And some other text

    ```fishi
    %%tokens

    [A-Za-z_0-9]   %token identifier
    ```

    ```fishi
    %%grammar

    {S} = {S} {E} | {E}
    ```

Which, since any FISHI without a section header will be interpreted as a
continuation of the section from before, is interpreted the same as this:

    ```fishi
    %%tokens

    \d+            %token int
    ```

    And some other text

    ```fishi
    [A-Za-z_0-9]   %token identifier
    ```

    ```fishi
    %%grammar

    {S} = {S} {E} | {E}
    ```

### Comments

FISHI comments start with a '#' and go until the end of the line. If you need to
put a literal '#' in FISHI, put in "##":

    ```fishi
    %%tokens
    
    # This entire line is a comment. The below line is a literal single '#':
    ##    %token hash-sign    %human hash/pound sign
    
    \d+   %token int          %human integer   # comments can start anywhere on a line
    ```

*Note: Converting the doubled '#' back into a single one is handled by the FISHI
preprocessor, not the frontend. As a result, any syntax errors reported on lines
that contain a ## will instead show it as a single one, and the position of the
error on the line will be similarly affected. This may be fixed in a future
version of FISHI.*

### Headers And Directives

There are two main types of special keywords in FISHI specs. The first is
"headers". They are used at the top of a *section* (described below) and give
the type of information that will be in it. All headers start with two percent
signs (`%%`). The headers in FISHI are `%%tokens`, `%%grammar`, and `%%actions`.

The other type of keyword is the directive. They mark special information and
options within a spec. The exact types of directives allowed depends on the
section, but all of them start with a single percent sign (`%`). The directives
in FISHI are `%token`, `%stateshift`, `%human`, `%priority`, `%state`,
`%discard`, `%symbol`, `%prod`, `%with`, `%hook`, `%set`, and `%index`.

### Sections

FISHI statements are organized into three different types of *sections*. Each
section has a header that gives the type of the section, followed by statements
that have different structure based on the section they are in.

Tokens sections have the header `%%tokens` and specify patterns for the lexer to
use to find tokens in input text, as well as patterns that should be ignored.

    ## Example Tokens section:

    ```fishi
    %%tokens

    \d+                       %token int         %human integer literal
    \+                        %token +           %human plus sign
    [A-Za-z_][A-Za-z0-9_]*    %token id          %human identifier

    \s+                       %discard
    ```

Grammar sections have the header `%%grammar` and specify the context-free
grammar for the language, using a syntax similar to BNF. Any token defined in a
Tokens section can be used as a terminal symbol.

    ## Example Grammar section:

    ```fishi
    %%grammar
    
    {SUM}   =  {SUM} + {TERM}
            |  {TERM}

    {TERM}  =  id | int
    ```

Actions sections have the header `%%actions` and specify the actions that the
frontend should take during the syntax-directed translation to produce an
intermediate representation. Any grammar rule defined in a Grammar section can
have a syntax-directed definition (action) specified for it.

    ## Example Actions section:

    ```fishi
    %%actions

    %symbol {TERM}
    -> id:             {^}.value = lookup_value({0}.$text)
    -> int:            {^}.value = int({0}.$text)

    %symbol {SUM}
    -> {SUM} + {TERM}: {^}.value = add({0}.value, {2}.value)
    -> {TERM}:         {^}.value = identity({0}.value)
    ```

Sections in FISHI do not need to be in any particular order, and you can define
as many of the same type of section as you'd like. For instance, you can have a
Grammar section, followed by a Tokens section, followed by an Actions section,
followed by another Tokens section. The order of the types of sections doesn't
matter; all of the statements in the Tokens section are read in the order they
appear in a FISHI spec.

## Specifying the Lexer with Tokens

The first stage of an Ictiobus frontend is lexing (sometimes referred to as
scanning). This is where code, input as a stream of UTF-8 encoded characters,
is scanned for for recognizable symbols, which are passed to the parsing stage
for further processing. These symbols are called *tokens* - the 'words' of the
input language. Each token has a type, called the *token class*, that is later
used by the parser.

### FISHI Tokens Quick Reference

This section lays out how FISHI specifies the lexer for those already familiar
with the terminology of lexical specifications. If it doesn't make any sense to
you, skip this section for a guide to how the ictiobus lexer works using FISHI
for its examples.

FISHI specifies a lexer in `%%tokens` sections. Each `%%tokens` section has a
series of entries. A series of entries may be preceeded by a `%state STATENAME`
directive in it, which makes those entries be applied only when the lexer is in
state `STATENAME`; otherwise, they will be defined for the default state and all
other states as well.

Each entry starts with a regular expression, which must be the first thing on a
line. The regular expression is given using the RE2 syntax used by Go's `regexp`
package. Any leading and trailing whitespace is trimmed from the pattern;
literal whitespace chars that are the first or last characters in the pattern
need to be represented as a character class but are fine within it.

The pattern is followed by one or more directives that say how to handle matches
of that regular expression. All matches will advance the lexer by the matched
text, regardless of what action it takes for it. The following directives can be
specified for a pattern:

* `%token CLASS`       - Lex a token of class `CLASS` whose text is the match.
                       `CLASS` must not have whitespace in it and can have any
                       characters, but any letters should be lower-case to
                       distinguish from non-terminals in later stages of the
                       frontend.

* `%human HUMAN-NAME`  - Use the the human-readable name `HUMAN-NAME` to refer
                       to the associated token class in error and diagnostic
                       output. Must be used in an entry that also has a
                       `%token CLASS`; cannot be used on its own. `HUMAN-NAME`
                       can have any characters in it at all. Applies to all
                       cases where `CLASS` tokens are lexed, even if only
                       defined for one. If multiple distinct human names are
                       given to the same class, the last one defined is used.

* `%discard`           - Take no action with the matched text and continue
                       scanning. Mutually exclusive with `%token` and
                       `%stateshift`.

* `%priority PRIORITY` - Treat the pattern as having the given priority number,
                       with bigger numbers being higher priority. All patterns
                       are priority `0` by default. Can be used with any other
                       directive in this list. If two patterns are of the same
                       priority and both match, if one results in matching more
                       source text, the lexer will use that pattern, otherwise
                       the one defined first in the spec is used.

* `%stateshift STATE`  - Exit the current state and enter state `STATE`. All
                       knowledge of the prior state is discarded. The only way
                       to return to it is with another `%stateshift` defined for
                       that state (or one defined in the default state). It is
                       not possible to return to the default state after leaving
                       it. If both `%stateshift` and `%token` are declared for a
                       matched pattern, the lexer will first lex the token in
                       the current state, then shift to the new state.

Lexers provided by Ictiobus have a simple state mechanism. By default, they will
only use the default state and are effectively stateless. State functionality
is invoked by using the `%stateshift` and/or `%state` directive. A lexer in a
given state will use all specifications defined for that state as well as all of
those defined for the default state. A stateful lexer does not use a stack for
its states and retains no information of the prior states it was in when it
swaps states.

The `%` character has special meaning within FISHI and in particular in
`%%tokens` sections it is used to detect the end of patterns and arguments to
directives. Literal `%` characters used in these contexts in `%%tokens` sections
must be escaped by putting the escape sequence `%!` in front of them.

The newline character `\n` has special meaning within FISHI and in particular in
`%%tokens` sections it is used to detect the end of patterns and arguments to
directives. Literal `\n` characters used in these contexts in `%%tokens`
sections msut be escaped by putting the escape sequence `%!` in front of them.
Note that due to line-ending normalization, the line ending sequence `\r\n` will
be interpreted as `\n` in a FISHI spec, and only needs to have the `%!` in front
of the `\r` in the original source.

### Token Specifications

The lexing stage is specified in FISHI in `%%tokens` sections. Each entry in
this section begins with a regular expression pattern that tells the lexer how
to find groups of text, and gives one or more actions the lexer should perform.

    %%tokens

    \d+       %token int     %human integer literal
    "         %token dquote  %human interpreted string    %stateshift istring
    '[^']*'   %token string  %human string literal
    \+        %token +       %human plus sign

    \s+       %discard     # ignore whitespace

The ordering of the options doesn't matter; in the above, we could have put the
`%human` directives before the `%token` directives and the result would be the
same.

Additionally, the formatting is fairly freeform. The only major restrictions are
that each entry must start with a pattern as the first thing on a line, and that
arguments to directives must be on the same line as the directive. Any
directives which follow a pattern, regardless of whether they are on the same
line, are considered part of the same entry.

    %%tokens

    \d+    %token int    %human integer literal

    # the above is exactly equivalent to the following:

    \d+
    %token int
    %human integer literal

The lexer provided by ictiobus is *stateful*; it can change which patterns it
recognizes based on the state it is currently in, and can be configured to
change states whenever it encounters particular patterns. For more on this, see
the "Using Lexer States" section below.

### The Pattern

    \d+       %token int     %human integer literal
     ^-- this is the pattern

Each entry begins with a pattern. This is a regular expression given in the same
[RE2 syntax](https://github.com/google/re2/wiki/Syntax) that is accepted by the
`regexp` standard library in Go, and it has the same restrictions.

The pattern needs to be the first thing on a line (though there can be
whitespace before it) and cannot begin or end with a whitespace character. If a
literal space is needed to match at the start or end, use a character class
containing only the space, or escape it with `%!`:

    [ ]+     %token space-seq

    # the above is the same as:

    %! +     %token space-seq

The percent sign '%' is used to mark directives that come after the pattern. If
a literal percent sign is in the pattern, it must be escaped with `%!`:

    \d{1,3}%!%   %token percentage

The above pattern would be interpreted as `\d{1,3}%`.

As can be seen from above examples, the backslash character `\` does *not* need
to be escaped. This is relatively uncommon compared to other programming
languages; because FISHI uses backslashes so frequently in its lexer patterns,
it was decided to make the backslash be treated as a literal character and to
instead use an alternative sequence for escaping (`%!`).

The pattern will be applied to input text at the current position of the lexer.
Note that the lexer is, in general, not aware of the start or end of line with
respect to the patterns it uses; this means that `^` will *always* match (the
current position of the lexer at any given time is considered the start of the
string), and `$` will *never* match except at the very end of all input text. So
these cannot be used to distinguish patterns, or at least, not as they would
typically be used in regular expressions.

### Lexing Tokens

New tokens in FISHI are declared by using the `%token` directive after a
pattern, followed by the name of a token class. The lexer will lex the text
matched by the patten as a token of the given class.

    # declare a token class named 'int' found by matching against a sequence of
    # one or more digits:

    \d+     %token int

Token classes do not need to be pre-declared; using any name after `%token`
counts as declaring it. Additionally, the same token class can be used multiple
times to give multiple possible patterns to match against.

    # define the int token as lexed by matching against a sequence of digits,
    # but also by matching against the case-insensitve words "one", "two",
    # "three", "four", and "five":

    \d+                    %token int
    [Oo][Nn][Ee]           %token int
    [Tt][Ww][Oo]           %token int
    [Tt][Hh][Rr][Ee][Ee]   %token int
    [Ff][Oo][Uu][Rr]       %token int
    [Ff][Ii][Vv][Ee]       %token int

The token class is used in other sections to refer to a token as a *terminal
symbol* (more on that later), so it has particular formatting requirements and
it is *case-insensitive* (convention in FISHI is to use all lower-case letters
for the declaration, because when referring to them in other sections they must
be lower-case). Class names must be one single word with no whitespace in it or
else it would be impossible to use later in grammar sections; however, virtually
every other character is allowed, including symbols and non-word characters.

    # lex a token with class '+' whenever the lexer matches a literal '+'
    # character:
    \+        %token +

There are some soft limitations on the token class names; for instance, using
"(" as a token class is perfectly acceptable in Tokens and Grammar sections of
the spec, but it will not be able to be used to identify a production in Actions
sections because the "(" character has additional meaning there. To get around
this, you would have to give the token class a different name.

Token classes can be given a human-readable name that is used in syntax errors
and other diagnostic output. If one isn't given, then the name of the token
class will be used. To specify a new one, use the `%human` directive followed by
its human-readable name on the same line. This can have any character in it,
including spaces.

    # give the parentheses nice name for error output:

    \(    %token lp    %human left parenthesis ('(')
    \)    %token rp    %human right parenthesis (')')

You can use a percent sign or newline character in human names, but they have to
be escaped with the "%!" sequence.

    [Hh][Ii][Gg][Hh][%!%]    %token high-percent    %human keyword 'high %!%'
    [Ll][Oo][Ww][%!%]        %token low-percent     %human keyword 'low %!%'

If there are multiple uses of the same token class, only one of them needs to
have a %human directive on it. The same one will be used for that class no
matter how it was lexed.

    \d+                    %token int    %human integer literal
    [Oo][Nn][Ee]           %token int
    [Tt][Ww][Oo]           %token int

    # it's also fine to use the same human readable name for multiple of the
    # same type

    \d+                    %token int    %human integer literal
    [Oo][Nn][Ee]           %token int    %human integer literal
    [Tt][Ww][Oo]           %token int    %human integer literal

If there are multiple different human-readable names for the same class, then
the last one given will be used for the class.

    # don't use different human definitions for the same token! The below
    # example results in any 'int' being called 'number two' because it was the
    # last human name given; that's just confusing, glub

    \d+                    %token int    %human integers
    [Oo][Nn][Ee]           %token int    %human number one
    [Tt][Ww][Oo]           %token int    %human number two

### Discarding Input

Sometimes, you will want the lexer skip over certain kinds of text. The most
common use of this tends to be for ignoring whitespace, but it can be used to
ignore anything you can write a regex for.

The `%discard` directive is used after a pattern to inform the lexer to take the
matching text and do nothing with it. The match is discarded as if it weren't
there and the lexer continues scanning input after that match.

    # ignore whitespace
    \s+         %discard

This can't be used with the `%token` directive; it wouldn't make much sense to
both discard the text *and* lex it as a token.

### Using Lexer States

The lexer provided by ictiobus is capable of using states to switch between what
patterns it recognizes. This lets you swap modes for lexical specifications
which cannot be described succinctly (or at all) using just regular expressions.
Ideally, the lexer stage does not use states and instead uses carefully-crafted
patterns, but the world is not always ideal, and while the general advice is to
avoid using a stateful parser as much as possible, sometimes one must resort to
having them.

The lexer has a very simple model for states - it is in one at any given time
and retains no knowledge of any previous states. When the lexer begins, it will
be in the *default state* which has no name, and will only match patterns
defined for no state in particular. When it shifts to a new state, it begins
using the patterns defined for that state **in addition to the default
patterns**, as opposed to instead of them. The next time it is instructed to
shift states, it will begin using the patterns for the new state in addition to
the default ones, and stop using the patterns for the state it just exited.

Unless declared as being for a particular state, all patterns in the lexer
specification will be for the default state. Patterns for a particular state can
be defined by first using a `%state` directive, followed by the patterns to use
in that state. `%state` only needs to be declared once; all following patterns
(until the end of the Tokens section or until another `%state` directive) will
be used only for that state.

    # a contrived example that shows the use of states to enable different
    # patterns:

    %%tokens

    # patterns for the default state, which will be used by ALL states:
    
    \s+                     %discard
    [Mm][Oo][Dd][Ee]1       %stateshift MODE1
    [Mm][Oo][Dd][Ee]2       %stateshift MODE2

    %state MODE1

    "[^"]*"                 %token dstr      %human double-quoted string literal
    '[^']*'                 %token sstr      %human single-quoted string literal

    %state MODE2

    \d+                     %token int       %human integer literal
    [A-Za-z_][A-Za-z0-9_]   %token id        %human identifier

To instruct the lexer to change states when it matches a pattern, use the
`%stateshift` directive after that pattern, and give the state it should shift
to after it.

    [Mm][Oo][Dd][Ee]2       %stateshift MODE2

You can use `%stateshift` with `%token` directives; in this case, the lexer will
lex the matched text as the given token class, and then immediately swap to the
provided state.

### Pattern Priority

If two patterns match the same input, the Ictiobus lexer will prefer patterns
that consume more input over those that consume less.

    # for the below patterns, the input "obj->" would be lexed as an obj-ptr
    # token:

    obj         %token obj-bare
    obj->       %token obj-ptr

If two patterns would match the same length, the lexer will prefer the one that
is declared first.

    # for the below patterns, "cat" would be lexed as a mover token:
    ca[rt]       %token mover
    c[ao]t       %token sleeper

### Complete Tokens Example For FISHIMath

Here's a Tokens section used to implement FISHIMath. Some of the patterns need
to be escaped, due to being special characters in RE2 syntax.

    ```fishi
    %%tokens
    
    \s+                      %discard
    
    \*                       %token *            %human multiplication sign "*"
    /                        %token /            %human division sign "/"
    -                        %token -            %human minus sign "-"
    \+                       %token +            %human plus sign "+"
    >\{                      %token fishtail     %human fish-tail ">{"
    '\}                      %token fishhead     %human fish-head "'}"
    <o\^><                   %token shark        %human statement shark "<o^><"
    =o                       %token tentacle     %human value tentacle "=o"
    [A-Za-z_][A-Za-z0-9_]*   %token id           %human identifier
    [0-9]*.[0-9]+            %token float        %human floating-point literal
    [0-9]+                   %token int          %human integer literal
    ```

## Specifying the Parser with Grammar

The second stage of an Ictiobus frontend is parsing. In this stage, tokens read
from the lexer in the first stage are grouped into increasingly larger syntactic
constructs. This is organized into a *parse tree*, which describes the structure
of the input code.

There are many different algorithms that can be used to construct a parser. The
ictcc command allows you to select one with the `--ll`, `--slr`, `--clr`, or
`--lalr` options. If you're not sure which one would be best, ictcc can select
one automatically. See the [ictcc Manual](./ictcc.md) for more info on parsing
algorithms.

### FISHI Grammar Quick Reference

This section lays out how FISHI specifies the parser for those already familiar
with the terminology of context-free grammars. If it doesn't make any sense to
you, skip this section to see how regular grammars work using FISHI syntax for
its examples.

A FISHI spec specifies the parsing phase with a context-free grammar declared
using special BNF-ish syntax in `%%grammar` sections. By convention there are
spaces between symbols, but this is not required unless it would otherwise make
a terminal ambiguous with another terminal or with the symbol `=` or `|`.

    ```
    %%grammar

    {S}   =   {E} + {S} | {E}
    {E}   =   {F} * {E} | {F}
    
    {F}   =   ( {S} )
           |  id
           |  int
    ```

The head symbol of the first rule defined in a Grammar section of a spec is the
start symbol of the grammar.

Non-terminals are specified by wrapping their name in curly braces `{` and `}`.
Their name must start with an upper-case letter A-Z but beyond that they may
contain any characters besides `}`.

Terminals are specified by giving their name. All arguments to `%token`
directives in Tokens section in the same spec are valid terminal symbols. Their
names should be lower-case if they contain letters. A terminal symbol name
cannot start with `{` and end with `}` due to ambiguity with non-terminals that
would cause.

FISHI uses the equals sign `=` to indicate the head symbol derives the
production(s) on the right. Alternative productions for the same rule are
separated by `|`. The alternative may be listed on the next line, but if this is
the case then the `|` must be on the next line as well; lines cannot be ended
with a `|`.

The epsilon production is specified with an empty pair of curly braces, `{}`.
These braces must be together as a pair; they cannot be separated by whitespace.

### The Context-Free Grammar

All parsing algorithms build a parser by using a *context-free grammar* (CFG),
which is a special, rigorous definition of a language. A language itself in this
sense is a strictly defined concept, but the important thing for FISHI is that
it can be expressed using a CFG, and programming languages are almost always
this kind of language.

The Context-Free Grammar for the parser is specified in FISHI in `%%grammar`
sections. It's declared in `a special format that somewhat resembles BNF, but
with some modifications.

The CFG is made up of several rules; each rule gives a sequence of symbols that
the head symbol (the symbol on the left of the rules) can be turned into. This
is called *deriving* the symbols on the right hand of the rule. Spacing between
the symbols is ignored in FISHI, this will be ignored.

This rule says that the non-terminal symbol called SUM can derive the string
made up of the *terminals* (symbols with no rule in the gramamr where it is the
head symbol) "int", "+", then "int".

    %%grammar

    {SUM}   =   int + int

### Symbols In Grammars

Every symbol in a grammar rule is either a *terminal symbol* or a *non-terminal
symbol*. These are also often referred to as simply "terminals" and
"non-terminals" respectively. A terminal symbol has no rule in the grammar where
it is the head symbol, whereas a non-terminal symbol has at least one rule.
Terminals are so-called because they are the termination of derivation; once you
get a terminal, you cannot derive it into anything else.

Terminals in FISHI CFGs must be token classes that are defined in a Tokens
section somewhere else in the spec. They do not necessarily need to be
lower-case in FISHI, but internally within Ictiobus they will always be
represented as a lower-case string, so by convention any letters in them in
FISHI CFGs are lower-case.

Non-Terminals (symbols with at least one rule in the gramamr where it is the
head symbol) in FISHI are surrounded with curly braces `{` and `}` to
distinguish them from terminals. Their name must begin with an upper-case
letter, but after that they can be any character (except `}` of course, as that
would close the non-terminal braces).

Non-terminals can be in a *production* (the string on the right-hand side that
can be derived from the head symbol):

    {SUM}   =   {TERM} + {TERM}

### Multiple Productions

A non-terminal can have more than one string it can derive to. This is expressed
in FISHI by putting a vertical bar `|` between the strings, either on the same
line as both it and the alternative or on a new line. When deriving the head
symbol, any of them may be selected. These are called the productions (or,
sometimes, the *alternatives*) of the head symbol.

    %%grammar

    # This rule states that non-terminal TERM can derive a string consisting of
    # a single int, or a single id.

    {TERM}   =   int | id

    {EXPR}   =   {SUM}
             |   {PRODUCT}
             |   {TERM}
    
    # this version is not allowed, because every line must start with a head
    # symbol for a rule or the alternations bar '|' to give an alternation for
    # the current rule; you can't "continue" the alternation on the next line
    #
    # NOT ALLOWED:
    #
    # {TERM} = int |
    #          id

### Derivation With A CFG

The idea for using a grammar is that you start with the first non-terminal
defined, derive a string of symbols from it. Then take that string, and for each
symbol that is a non-terminal, derive a string for *that* symbol and replace it
with the string, and successively continue replacing non-terminals with a string
they could derive until only terminal symbols remain. Any time there are
multiple alternatives to select from, you take your pick as to which one to
derive. Any string of terminals that your able to derive this way is said to be
"in" the grammar, and thus is a valid string in the language it describes.

For example, given the following CFG:

    %%grammar

    {EXPR}     =  {EXPR} + {PRODUCT} | {PRODUCT}
    {PRODUCT}  =  {PRODUCT} * {TERM} | {TERM}
    {TERM}     =  ( {EXPR} ) | int | id

You would start with the non-terminal `EXPR`. From there, you might perform the
following set of derivations:

    STRING                        | GRAMMAR RULE USED
    ----------------------------------------------------------------
    {EXRP}                        | Start Symbol
    {PRODUCT}                     | {EXPR}     =  {PRODUCT}
    {PRODUCT} * {TERM}            | {PRODUCT}  =  {PRODUCT} * {TERM}
    {PRODUCT} * int               | {TERM}     =  int
    {TERM} * int                  | {PRODUCT}  =  {TERM}
    ( {EXPR} ) * int              | {TERM}     =  ( {EXPR} )
    ( {EXPR} + {PRODUCT} ) * int  | {EXPR}     =  {EXPR} + {PRODUCT}
    ( {EXPR} + {TERM} ) * int     | {EXPR}     =  {TERM}
    ( {EXPR} + id ) * int         | {TERM}     =  id
    ( {PRODUCT} + id ) * int      | {EXPR}     =  {PRODUCT}
    ( {TERM} + id ) * int         | {PRODUCT}  =  {TERM}
    ( int + id ) * int            | {TERM}     =  int


You might have noticed that these derivations look an awful lot like
definitions as well; one way to think of the grammar is a series of saying what
a symbol is made up of, then describing THOSE structures, until you get to only
the basic tokens that the lexer can make.

### The Epsilon Production

It's possible to have a rule that has the head symbol derive an empty string.
This kind of production is known as the *epsilon production*, because in
theoretical literature, it is often represented by a lower-case greek letter
epsilon (`ε`). Epsilon is not an easy-to-type character on most keyboards, so
while Ictiobus will use it in certain outputs, in FISHI it is represented by a
production having only an empty set of braces. When performing a derivation with
the epsilon production, the head symbol is replaced with the empty string (i.e.
nothing at all), which effectively is removing it from the final string.

    %%grammar

    {FUNC-CALL}   =    identifier ( )
                  |    identifier ( {ARG} {NEXT-ARGS} )

    {NEXT-ARGS}   =    , {ARG} {NEXT-ARGS} | {}

    {ARG}         =    int | string | float | identifier

The above rules can derive a string representing a function call of any length.
The first production for `FUNC-CALL` specifies that a function call may be an
`identifier` followed by a pair of parentheses `(` and `)` with nothing in
between them for zero arguments. The other production is invoked for one or more
arguments. It states a function call may be an `identifier` followed by a left
parenthesis `(`, then an `ARG` (which itself is an `int`, `string`, `float`, or
`identifier`), a `NEXT-ARGS`, then finally the closing right parnethesis `)`.
`NEXT-ARGS` has a *recursive* production (one that derives at least one copy of
the head symbol), and this production can be repeatedly invoked to build up more
and more comma-separated `ARG` symbols until the desired amount is reached. The
remaining `NEXT-ARGS` is then eliminated by using the epsilon production to
derive an empty string for it.

For example, to derive a string that represents a function call with 3
arguments using the above grammar, the following derivation could be performed:

    STRING                                            | GRAMMAR RULE USED
    ---------------------------------------------------------------------------------------------
    {FUNC-CALL}                                       | Start Symbol
    identifier ( {ARG} {NEXT-ARGS} )                  | {FUNC} = identifier ( {ARG} {NEXT-ARGS} )
    identifier ( {ARG} , {ARG} {NEXT-ARGS} )          | {NEXT-ARGS} = , {ARG} {NEXT-ARGS}
    identifier ( {ARG} , {ARG} , {ARG} {NEXT-ARGS} )  | {NEXT-ARGS} = , {ARG} {NEXT-ARGS}
    identifier ( {ARG} , {ARG} , {ARG} )              | {NEXT-ARGS} = {}
    identifier ( int , {ARG} , {ARG} )                | {ARG} = int
    identifier ( int , identifier , {ARG} )           | {ARG} = identifier
    identifier ( int , identifier , string )          | {ARG} = string

This trick can be useful, but sometimes the epsilon production can cause certain
issues with parser generation. Often the epsilon production can be eliminated
from a grammar simply by rewriting the rules. For example, the above grammar
could be rewritten without epsilon productions as follows:

    %%grammar

    {FUNC-CALL}   =    identifier ( )
                  |    identifier ( {NEXT-ARG} )

    {NEXT-ARG}    =    {ARG}
                  |    {ARG} , {NEXT-ARG}

    {ARG}         =    int | string | float | identifier

The good news is that there are well-defined techniques for eliminating epsilon
productions from a grammar that always work. Describing them is beyond the scope
of this guide, but they can be easily found looking up the relevant literature.

### Complete Grammar Example For FISHIMath

This example defines the context-free grammar for FISHIMath, using the tokens
defined in the previous example for FISHIMath in the Tokens section of this
manual.


    ```fishi
    %%grammar


    {FISHIMATH}   =   {STATEMENTS}

    {STATEMENTS}  =   {STMT} {STATEMENTS} | {STMT}

    {STMT}        =   {EXPR} shark

    {EXPR}        =   id tentacle {EXPR} | {SUM}

    {SUM}         =   {PRODUCT} + {EXPR}
                  |   {PRODUCT} - {EXPR}
                  |   {PRODUCT}

    {PRODUCT}     =   {TERM} * {PRODUCT}
                  |   {TERM} / {PRODUCT}
                  |   {TERM}

    {TERM}        =   fishtail {EXPR} fishhead
                  |   int | float | id
    ```

## Specifying the Translation Scheme with Actions

The third and final stage of an Ictiobus frontend is translation. This is where
a *syntax-directed translation scheme* (or "SDTS") is applied to nodes of the
parse tree created in the parsing phase to produce the final result, the
*intermediate representation* (or "IR") of the input. This IR is then returned
to the user of the frontend as the return value of `Frontend.Analyze()`.

### FISHI Actions Quick Reference


The `%%actions` section contains the syntax-directed translation scheme,
specified with a series of entries. Each entry gives syntax-directed definitions
for attributes in a conceptual attribute grammar defined for the language.

An entry takes the following form:

```
   /-- %symbol directive
   |
   |           /-- head symbol
/-----\ /--------------\
%symbol {A-NON-TERMINAL}

                                         /-- production action set
                                         |
/----------------------------------------------------------------------------------\

-> {PROD-SYM1} prod-sym-2 {PROD-SYM3} ... : {^}.value = hook_func({PROD-SYM3}.value)
                                          : {^}.value = hook_func()

   \------------------------------------/ \----------------------------------------/
           \-- production selector                   \-- semantic actions
```

Each entry begins with `%symbol` followed by the non-terminal symbol that is at
the head of the grammar rule that SDDs are being given for. This is then
followed by one or more production action sets.

A production action set begins with the symbol `->` followed by an optional
selector for a production of the head symbol. The selectors are as follows:

* `-> SYMBOLS-IN-PRODUCTION-EXACTLY-AS-IN-THE-GRAMMAR-RULE` - directly specifies
the production it corresponds to.
* `-> %index PRODUCTION-INDEX` - specifies the Nth production, where
`PRODUCTION-INDEX` is N-1.
* `->` (no selector given) - specifies the production with index one higher than
the one selected by the previous production action set (or index 0 if it is the
first set).

A production action set's selector is followed by one or more semantic actions
(or SDDs as the term is used in this manual). A semantic action begins with the
`:` symbol, then an AttrRef giving the attribute to be set, an equals sign
(`=`), and then the name of a hook function to call to calculate the value. The
hook function can have one or more arguments specified by AttrRefs which can be
comma-separated for readability, or it can have zero args.

An AttrRef takes the following form:

```
 |-- node-reference
/-\

{^}.value

    \___/
      |-- name
```

The `node-reference` part refers to a node relative to the current node the SDD
is being defined for/will run on, and the the `name` part is the name of an
attribute on that node.

The `node-reference` part of an AttrRef before the `.` can take one of the
following forms:

* `{^}`          - The head symbol/current node. Only valid for LHS of a
                   semantic action due to requirement that STDS be implementable
                   with an S-attributed grammar.
* `{.}`          - The child node corresponding to the first terminal symbol in
                   the production.
* `{.N}`         - The child node corresponding to the Nth terminal symbol in
                   the production (0-indexed).
* `{&}`          - The child node corresponding to the first non-terminal symbol
                   in the production.
* `{&N}`         - The child node corresponding to the Nth non-terminal symbol
                   in the production (0-indexed).
* `{NON-TERM}`   - The child node corresponding to the first occurance of
                   non-terminal symbol `NON-TERM` in the production.
* `{NON-TERM$N}` - The child node corresponding to the Nth occurance of
                   non-terminal symbol `NON-TERM` in the production (0-indexed).
* `term`         - The child node corresponding to the first occurance of
                   terminal symbol `term` in the production. If used as an arg
                   to a hook function, must be separated from prior symbol by
                   whitespace.
* `term$N`       - The child node corresponding to the Nth occurance of terminal
                   symbol `term` in the production. If used as an arg to a hook
                   function, must be separated from prior symbol by whitespace.

### Syntax-Directed Translation Schemes

In a syntax-directed translation scheme, the intermediate representation is
built up from the parse tree, which will have a node for every grammar construct
that would be invoked to derive the parsed string. Each node is visited and has
a named value set on it (called an *attribute*) by taking the value of zero or
more other attributes already set on the node or its children and applying an
operation on them to produce a new value.

Some nodes have attributes automatically set on them before any evaluation is
performed, such as the `$text` attribute of nodes representing terminal symbols,
which is the text that was lexed for the token. These are used as the starting
values for a translation scheme. All such built-in attributes have a name that
starts with a `$`.

The IR translation process is started by first setting attributes that require
no additional attributes or only built-in attributes in order to calculate their
value. Next, it goes one level up, calculating the value of attributes on parent
nodes that rely on the recently calculated ones. This is repeated until all
defined attributes are set on the root node of the parse tree. One of these
attributes is designated as the IR attribute and once it is set, its value is
returned from the frontend.

The rules used to define new attributes in an SDTS, known as *syntax-directed
definitions* (or SDDs), are often expressed using an *attribute grammar*. This
has any SDDs for setting new values on attributes in a parse tree node written
next to the grammar rule that must be invoked to create that node. The SDDs are
often written in syntax that can be directly executed by the language the parser
is being generated in.

```
An example attribute grammar that you might see in literature, which specifies
how to build up an IR consisting of the result of evaluating the mathematical
value of expressions.

SUM   ->  SUM + TERM    { SUM.val = S.val + E.val }
SUM   ->  TERM          { SUM.val = E.val }
TERM  ->  int           { TERM.val = int.$lexed }
```

Although many parser generators have a specification syntax similar to this,
FISHI does not take this exact approach. Instead, it splits the specification
for the translation stage into a separate section from the grammar, the
`%%actions` section. This keeps the grammar clear of hard-to-read SDDs and opens
the option of performing *no* translation of the parse tree, instead returning
it as the IR (not directly supported at this time).

The above example could be represented by the following FISHI:

    ```fishi
    %%grammar

    {SUM}   =  {SUM} + {TERM} | {TERM}
    {TERM}  =  int



    %%actions

    %symbol {SUM}
    -> {SUM} + {TERM}: {^}.val = add({0}.val, {2}.val)
    -> {TERM}:         {^}.val = identity({0}.val)

    %symbol {TERM}
    -> int:            {^}.val = int({0}.$text)
    ```

The syntax for SDDs in FISHI is significantly more complex than the theoretical
one. This is partially due to decoupling Go from its own syntax so that
executions of Go needed to produce a value can be handled by separate libraries.
This part of FISHI is the newest and is the most likely to have its syntax
updated.

Unlike other sections in FISHI, whitespace in `%%actions` sections has
absolutely no semantic meaning anywhere and is completely ignored. The only
place where one might encounter something contrary to this is the `()` symbol in
certain places, described below.

### The Actions Entry

Entries in a FISHI Actions section begin with an associated symbol for the
entry. This is specified with the `%symbol` directive followed by the
non-terminal symbol at the head of the grammar rule that attributes are going to
be defined for.

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.value, {2}.value)
    ^^^^^^^^^^^^^
    Associated Symbol

For instance, the above FISHI starts an Actions entry for parse tree nodes
created by invoking grammar rules for derivations of `SUM`.

It is acceptable to have multiple entries for the same symbol; they will be
applied in the order that they are defined.

The associated symbol is followed by one or more production action sets, each of
which contains a "selector" for a production of the head symbol and one or more
SDDs for setting an attribute on nodes created by deriving that production. Each
production action set starts with a `->` symbol (or the outdated `%prod`
keyword) followed by the selector.

    %%actions

    %symbol {SUM} -> {SUM} + {TERM}   : {^}.val = add({0}.value, {2}.value)
                  ^^^^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        Production action set start   SDD

There can be multiple production action sets for a single head symbol, each of
which gives SDDs for the production they specify:

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.val, {2}.val)
                  -> {TERM}         : {^}.val = identity({0}.val)

Each SDD for the production starts with a `:` sign (or the outdated `%set`
keyword). SDDs use a somewhat complex syntax that will be covered shortly in the
"SDDs" section below. There can be multiple SDDs per production. If there are,
they will be applied in the order they are defined or the order needed to
satisfy attribute calculations that depend on them.

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.val, {2}.val)
                                    : {^}.val = set_plus_op()
                  -> {TERM}         : {^}.val = identity({0}.val)

Because whitespace is ignored in `%%actions` sections, they can be formatted in
any way that the user finds readable.

    # this:
    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.val, {2}.val)
                  -> {TERM}         : {^}.val = identity({0}.val)

    # is the same as this:
    %symbol {SUM}
    ->{SUM} + {TERM}:   {^}.val = add({0}.val, {1}.val)
    ->{TERM}:           {^}.val = identity({0}.val)

### Shortcuts For Productions

Though typing out every production for a rule in full results in code that
clearly gives the production that the SDDs are associated with, it can be
cumbersome, especially for large grammars. FISHI allows the production to be
specified by its index within all productions of the head symbol as defined in
the grammar with the `%index` directive (the first production has index 0, the
second has index 1, etc).

    %%grammar

    #assuming the rule for SUM has the following order for its productions...
    {SUM}    =    {SUM} + {TERM} | {TERM}


    %actions

    # ...then this:
    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.val, {1}.val)
                  -> {TERM}         : {^}.val = identity({0}.val)

    # is the same as this:
    %symbol {SUM} -> %index 0       : {^}.val = add({0}.val, {1}.val)
                  -> %index 1       : {^}.val = identity({0}.val)

If the indexes are in order, they can be omitted entirely. A production action
set with no production specified will be for the production whose index is one
higher than the previous production, or 0 if it is the first action set.

    # so this would be the same as the above examples:
    %symbol {SUM} -> : {^}.val = add({0}.val, {1}.val)
                  -> : {^}.val = identity({0}.val)

But do note that the implicit production selector is fragile to re-ordering of
the grammar, so it might be best to hold off on using until the grammar is in a
relatively stable state.

### SDDs

As mentioned previously, each production action set contains one or more
syntax-directed definitions (SDDs). Each SDD starts with the symbol `:` (or the
outdated keyword `%set`) and then specifies the attribute that is to be defined
with an AttrRef (see below section "AttrRefs"), which at this time
must always be an attribute on the root node `{^}`. This is followed by an
equals sign `=` (or the outdated keyword `%hook`), and then the name of a *hook
function* to execute whose return value is set as the value of the attribute. If
the hook takes zero arguments, it can have a `()` after its name (but they
*must* be together with no space between them).

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} :  {^}.node_type  =  make_type_sum()
                                    |  \___________/  |  \_____________/
                                    |        |        |         \- hook function
                                    |        |        |
                                    |        |        \- `=` sign
                                    |        |
                                    |        \- attribute to set
                                    |
                                    \- SDD start sign `:`

The actual definition of what the hook function does is not defined in a FISHI
spec; instead, it names the hook functions and leaves their implementations as a
detail that the creator of the language must provide when retrieving the
generated frontend. This is passed in as a special
`github.com/dekarrin/ictiobus/trans.HookMap` type variable that maps the names
of the hook function to functions that are executed for them. For instance, the
above hook `make_type_sum` might be implmented by users of the frontend
providing the following hook table at runtime:

```go

import (
    "github.com/dekarrin/ictiobus/trans"
)

var (
	HooksTable = trans.HookMap{
		"make_type_sum":     hookImplMakeTypeSum,
        // ... could also be more bindings of hook names to implementations
	}
)

// an example function to show the signature of hook implementations
func hookImplMakeTypeSum(info trans.SetterInfo, args []interface{}) (interface{}, error) {
    typeInfoMap := map[string]string{
        "type": "sum",
    }

    return typeInfoMap, nil
}
```

For more information on defining hook implementations, see the "Generated Code"
section of the [ictcc Manual](./ictcc.md), or for a complete working example of
hook implementations, see the main Readme file at the root of this project.

Hook functions may take on one or more arguments by specifying them after the
opening parenthesis `(` (or after the outdated keyword `%with`). They can be
comma-separated, and each one must also be an AttrRef. At this time, arguments
to hooks may only refer to nodes for production symbols, and not the head symbol
node (see the AttrRef section below).

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} :  {^}.value = add({0}.value, {2}.value)

The values of the referenced attributes will be passed to the hook implemenation
at runtime in its second parameter as a slice of `interface{}` objects. When
there are no arguments, the hook implementation is passed an empty slice.

The closing parenthesis at the end of a list of arguments to the hooks is pure
syntactic sugar and is ignored by the parser, as are the commas between
arguments, but it is recommended they be used for readability purposes. The
empty parentheses pair `()` for a hook function with no arguments is also
syntactic sugar and may be omitted without changing the meaning. It is
considered a single symbol, and splitting the parentheses apart will result in
their being interpreted as a `(` (start of arguments list) followed by a `)`
(discarded and ignored), which will likely result in errors regarding an AttrRef
being expected as opposed to whatever comes next.

    # These two are equivalent:

    %symbol {SUM} -> {SUM} + {TERM} :  {^}.node_type  =  make_type_sum()
    %symbol {SUM} -> {SUM} + {TERM} :  {^}.node_type  =  make_type_sum

    # but this is actually *not* the same, and will break:
    %symbol {SUM} -> {SUM} + {TERM} :  {^}.node_type  =  make_type_sum( )

While it is a peculiarity of FISHI that `()` is treated separately from `(` and
`)` (and in the future, this might not be the case!), this is done to
distinguish the syntactic sugar of indicating `empty arguments` from the
syntatic sugar of starting the arg list with a single `(`. It is the way it is
for historical reasons and future versions of Ictiobus may change this.

### AttrRefs

The Actions sections of FISHI represent references to attributes using a
syntactic construct called the *AttrRef* (short for attribute reference). This
consists first of a reference to a particular node in the parse tree relative to
the node for the head symbol of the action it is being defined for. This
reference can specify nodes by the name of their symbol, what type of symbol it
is, and other criteria. The exact syntax for the node-reference portion of an
AttrRef is different based on the type of criteria it uses, but most are wrapped
in curly braces `{` and `}`.

The node-reference portion is followed by a `.`, and then the name of an
attribute on the referred-to node. This attribute name is made up of the
characters `A-Z`, `a-z`, `_`, and `$`, and must start with either a letter or
`$`. Attribute names that start with `$` are reserved for internal use. When
defining a new attribute, starting with `$` is not allowed.

    %%actions

    %symbol {SUM}
    -> {SUM} + {TERM}:  {^}.value   =   add({0}.value, {2}.value)

In the above SDD, `{^}.value` refers to the attribute named `value` on the node
in the parse tree whose symbol is the head node (aka the node that the SDD is
defined for). If `value` didn't exist on a node before this action is invoked on
it, this will create it; otherwise, it will be updated. `{0}.value` refers to an
attribute named `value` on the child node corresponding to the 0th (i.e. first)
symbol in the production (so the node for the `SUM` non-terminal), and
`{2}.value` likewise refers to an attribute named `value` on the third symbol in
the production, `TERM`.

#### Types Of Node References

The node-reference part of an AttrRef can vary in appearance based on how it
specifies the node. As mentioned before, `{^}` refers to the head symbol of the
associated grammar rule:

    %%actions

    %symbol {SUM}
    -> {SUM} + {TERM}:  {^}.value   =   add({0}.value, {2}.value)
                        \_______/
                            |- attribute named 'value' on the node for the head symbol

The node-reference can refer to the child node of the current one that
corresponds to the Nth symbol in the associated production by putting its
0-based index in between curly braces. This is by-far the shortest form of node
reference for a production symbol (although it can be difficult to read if the
production specifier does not explicitly give the production):

    %%actions

    %symbol {SUM}
    -> {SUM} + {TERM}:  {^}.value   =   add({0}.value, {2}.value)

In the above example, the `{0}` refers to the `{SUM}` in the production, because
it is the first symbol (index 0). The `{2}` refers to the `{TERM}`, because it
is the third symbol (index 2).

The node-reference can refer to an instance of a particular non-terminal symbol
in the production by instead giving the name of the symbol in braces, similar to
how non-terminal symbols are defined:

    # the above could have been:
    -> {SUM} + {TERM}:  {^}.value   =   add({SUM}.value, {TERM}.value)

If there's more than one occurance of that non-terminal in the production, the
index of the occurance (starting at 0) can be given by putting a dollar sign
followed by the index inside of the braces. If it's not given, it's assumed to
be 0 (the first occurance):

    %%grammar

    {SUM}  =  {EXPR} + {EXPR}


    %%actions

    %symbol {SUM}
    -> {EXPR} + {EXPR} : {^}.value = add({EXPR$0}.value, {EXPR$1}.value)

    # the above uses the first instance of EXPR in the production (`$0`) as the
    # first arg to add(), and the second instance of EXPR in the production
    # (`$1`) as the second arg.

The occurance must be explicitly given if the name of the non-terminal itself
contains a dollar so that the non-terminal name is correctly parsed:

    -> {BIG$SYMBOL} + {TERM} : {^}.value = add({BIG$SYMBOL$0}.value, {TERM}.value)

An instance of a particular terminal symbol can be specified in a very similar
way as non-terminals, just without the braces:

    %%actions

    %symbol {TERM}
    ->  int :       {^}.value = int_value( int.$text)

Note that due to limitations of the FISHI parser, if you're using an attr ref
that specifies a terminal by name, you need to separate it by space from
whatever came before it, otherwise the entire prior thing could be detected as
an "unexpected attribute reference literal".

Just like with non-terminals, the above `int.$text` specifies the attribute
named `$text` (a built-in one, due to the leading `$`) defined on the child node
corresponding to the terminal `int`.

This follows the same rules for specifying a particular index as non-terminals;
to refer to occurances of a non-terminal symbol that are not the first, the
index of the occurance must be specified with a dollar sign after the name of
the terminal:

    %%actions

    %symbol {INT-SUM}
    -> int + int:      {^}.value = add( int$0.$text, int$1.$text)

The occurance must be explicitly given if the name of the terminal itself
contains a dollar so that the non-terminal name is correctly parsed:

    -> $int$  : {^}.value = add( $int$$0.$text)

Besides occurances of specific terminals and non-terminals, an AttrRef may refer
to *any* terminal or non-terminal by using `{.}` or `{&}` respectively. `{.}`
and `{&}` alone specify the first non-terminal or terminal:

    -> {SUM} int:  {^}.value   =   add({.}.$text, {&}.value)

They can also specify the Nth terminal or non-terminal by giving the index of
the terminal after the `.`/`&`, still within braces:

    -> int + string:   {^}.value = concat({.2}.$text, {.0}.$text)

    -> {TERM} + {SUM}: {^}.value = add({&1}.value, {&0}.value)

### Built-In Attributes

Usually, if an attribute is being used as an argument to a hook, it must be
defined by another SDD elsewhere on the grammar rule for that symbol. But some
attributes are automatically added to a parse tree during initial creation of
the annotated parse tree. All of these attributes will have a name that starts
with a dollar sign `$`.

* `$text` - Defined for terminal symbol nodes only. This is the text that was
lexed for the token and it will always be of type `string`.
* `$id`   - Defined for all nodes. This is a unique ID of type `uint64`. Every
node will have an `$id` attribute defined.
* `$ft`   - Defined for all nodes except terminal nodes representing the epsilon
production. This is the "First Token" of the node; for terminal nodes, this is
the token that was lexed for it, for non-terminal nodes, this is the first token
that is part of the grammar construct. Note that the token info for the node a
hook is producing a value for is always passed to the implementation as its
first argument; `$ft` can be used to get the first-tokens of symbols from a
production.

### Synthesized vs Inherited Attributes

Within Ictiobus (and sometimes in FISHI) you may come across the concept of
*synthesized* attributes. A synthesized attribute is one that is set on the same
node as its SDD is called on. Every attribute described in this document so far
has been synthesized; in FISHI this is done by assigning to an attribute of the
head symbol with `{^}.someAttributeName = ...`. For synthesized attributes,
arguments to the hook can only be from nodes that are children of the node that
the SDD is called on (i.e. symbols from the production). Again, this is the only
example shown in this document.

There is also the idea of *inherited* attributes. Inherited attributes would be
created by setting the value of an attribute in a child node of the one the SDD
is running on. They can use the values of attributes on sibling nodes and the
head note as arguments to the hook.

A translation scheme defined by a series of SDDs on grammar rules (known
formally as an *attribute grammar*) that only define synthesized attributes is
known as S-attributed. Despite having some code for handling inherited
attributes, Ictiobus does not officially support this and it is poorly tested.
They should in general not be used.

### The IR Attribute

When building up the translation scheme, the first attribute defined for the
starting symbol of the gramamr is set as the attribute that the IR is taken from
once SDTS evaluation completes.

    %%grammar

    {COMPLETE-PROGRAM}  =   {STMT-LIST}
    {STMT-LIST}         =   {STMT-LIST} {STMT} | {STMT}
    {STMT}              =   {FUNC-CALL} | {EXPR}

    # ... a lot more grammar rules


    %%actions

    %symbol {COMPLETE-PROGRAM}
    -> {STMT-LIST} :  {^}.ast = make_list_node({0}.ast)

In the above example, the attribute `ast` on the root `COMPLETE-PROGRAM` node of
parse trees would be used to retrieve the IR, because it is the first one
defined for the grammar start symbol (`COMPLETE-PROGRAM`).


### Creation Of Abstract Syntax Trees

This IR value can be anything that you wish it to be. For simpler languages,
such as FISHIMath, it can be immediately calculated by evaluating mathematical
options to build up a final result. For others, a special representation of the
input code called an *abstract syntax tree* (AST) might be built up. An abstract
syntax tree difers from a parse tree by abstracting the grammatical details
of how the constructs within the code were arranged into a more natural
representation that reflects the logical structures of the language, such as
control structures. It is suitable for further evaluation by the caller of the
frontend.

```go
if x >= 8 && y < 2 {
    fmt.Printf("It's true!\n")
}
```

The above block of Go syntax might be parsed into something that resembles the
following tree:

    [IF-BLOCK]
     |-- (kw-if "if")
     |-- [COND]
     |    \-- [EXPR]
     |         |-- [EXPR]
     |         |    |-- [EXPR]
     |         |    |    \-- (identifier "x")
     |         |    |
     |         |    |-- (binary-op ">=")
     |         |    |
     |         |    \-- [EXPR]
     |         |         \-- (int "8")
     |         |
     |         |-- (binary-op "&&")
     |         |              
     |         \-- [EXPR]
     |              |-- [EXPR]
     |              |    \-- (identifier "y")
     |              |
     |              |-- (binary-op "<")
     |              |
     |              \-- [EXPR]
     |                   \-- (int "2")
     |
     \-- [BODY]
          |-- (block-open "{")
          |-- [STATEMENTS]
          |    \-- [STATEMENT]
          |         \-- [FUNC-CALL]
          |               |-- (identifier "fmt")
          |               |-- (period ".")
          |               |-- (identifier "Printf")
          |               |-- (lparen "("))
          |               |-- [ARGS]
          |               |    \-- [ARG]
          |               |         \-- (dquote-string ""It's true!\n"")
          |               |
          |               \-- (rparen ")")
          \-- (block-close "}")

That's ludicrously messy, and would be difficult to directly analyze. By
defining a series of actions in the SDTS which convert each item to some sort of
AST node, passing the values along as neede, a tree that representes the same
syntax but in a more abstract fashion can be created:

    [if-statement]
     |-- condition: [binary expression (op="&&")]
     |               |-- left:  [binary-expression (op=">=")]
     |               |           |-- left:  [variable (name="x")]
     |               |           \-- right: [literal (type=int, value=8)]
     |               \-- right: [binary-expression (op="<")]
     |                           |-- left:  [variable (name="y")]
     |                           \-- right: [literal (type=int, value=2)]
     \-- body: [statement-list]
                     \-- stmts[0]: [func-call (pkg="fmt", name="Printf")]
                                     \-- args[0]: [literal (type=string, value="It's true!\n")]

Much better! And way easier to analyze! For langauges that cannot be immediately
evaluated via the SDTS, an AST is a reasonable IR.

### Complete Actions Example For FISHIMath

This section builds up a translation scheme from the grammar and tokens defined
in previous FISHIMath sections. It gives it in two forms: one that performs
immediate evaluation and returns the result as the IR, and one that creates an
AST of the input and uses that as the IR.

#### Immediate Evaluation Version

This example performs immediate evaluation. Its IR is a slice of numbers, where
the item with index N is the value of evaluating the Nth statment. If you want
to try it out, be sure and use the Go code in Appendix A to provide the
HooksTable.

    ```fishi
    %%actions

    # This actions section creates an IR that is the result of evaluating the
    # expressions. It will return a slice of FMValue, one FMValue for each
    # statement in the input.


    # {FISHIMATH}.ir is called that for readability purposes; there's nothing
    # special about the attribute name "ir". What makes this the final IR value
    # returned from the frontend is that it is the first attribute in an SDD
    # defined for the grammar start symbol {FISHIMATH}.

    %symbol {FISHIMATH}
    -> {STATEMENTS}:              {^}.ir = identity({&0}.result)

    %symbol {STATEMENTS}
    -> {STMT} {STATEMENTS}:       {^}.result = num_slice_prepend({1}.result, {0}.value)
    -> {STMT}:                    {^}.result = num_slice_start({0}.value)

    %symbol {STMT}
    -> {EXPR} shark:              {^}.value = identity({EXPR}.value)

    %symbol {EXPR}
    -> id tentacle {EXPR}:        {^}.value = write_var( id.$text, {EXPR}.value)
    -> {SUM}:                     {^}.value = identity({SUM}.value)

    %symbol {SUM}
    -> {PRODUCT} + {EXPR}:        {^}.value = add({&0}.value, {&1}.value)
    -> {PRODUCT} - {EXPR}:        {^}.value = subtract({&0}.value, {&1}.value)
    -> {PRODUCT}:                 {^}.value = identity({PRODUCT}.value)

    %symbol {PRODUCT}
    -> {TERM} * {PRODUCT}:        {^}.value = multiply({&0}.value, {&1}.value)
    -> {TERM} / {PRODUCT}:        {^}.value = divide({&0}.value, {&1}.value)
    -> {TERM}:                    {^}.value = identity({TERM}.value)

    %symbol {TERM}
    -> fishtail {EXPR} fishhead:  {^}.value = identity({EXPR}.value)
    -> int:                       {^}.value = int({0}.$text)
    -> float:                     {^}.value = float({0}.$text)
    -> id:                        {^}.value = read_var( id.$text)
    ```

#### AST Creation Version

This example performs creation of an abstract syntax tree, deferring actual
evaluation to the caller of the frontend. Its IR is an AST successively built up
out of each of the grammar constructs. If you want to try it out, be sure and
use the Go code in Appendix B to provide the HooksTable.

    ```fishi
    %%actions

    # This actions section creates an abstract syntax tree and returns it as the
    # IR.


    # {FISHIMATH}.ir is called that for readability purposes; there's nothing
    # special about the attribute name "ir". What makes this the final IR value
    # returned from the frontend is that it is the first attribute in an SDD
    # defined for the grammar start symbol {FISHIMATH}.

    %symbol {FISHIMATH}
    -> {STATEMENTS}:              {^}.ir = ast({&0}.stmt_nodes)

    %symbol {STATEMENTS}
    -> {STMT} {STATEMENTS}:       {^}.stmt_nodes = node_slice_prepend({1}.stmt_nodes, {0}.node)
    -> {STMT}:                    {^}.stmt_nodes = node_slice_start({0}.node)

    %symbol {STMT}
    -> {EXPR} shark:              {^}.node = identity({EXPR}.node)

    %symbol {EXPR}
    -> id tentacle {EXPR}:        {^}.node = assignment_node( id.$text, {EXPR}.node)
    -> {SUM}:                     {^}.node = identity({SUM}.node)

    %symbol {SUM}
    -> {PRODUCT} + {EXPR}:        {^}.node = binary_node_add({&0}.node, {&1}.node)
    -> {PRODUCT} - {EXPR}:        {^}.node = binary_node_sub({&0}.node, {&1}.node)
    -> {PRODUCT}:                 {^}.node = identity({PRODUCT}.node)

    %symbol {PRODUCT}
    -> {TERM} * {PRODUCT}:        {^}.node = binary_node_mult({&0}.node, {&1}.node)
    -> {TERM} / {PRODUCT}:        {^}.node = binary_node_div({&0}.node, {&1}.node)
    -> {TERM}:                    {^}.node = identity({TERM}.node)

    %symbol {TERM}
    -> fishtail {EXPR} fishhead:  {^}.node = group_node({EXPR}.node)
    -> int:                       {^}.node = lit_node_int({0}.$text)
    -> float:                     {^}.node = lit_node_float({0}.$text)
    -> id:                        {^}.node = var_node( id.$text)
    ```

## Appendix A: FISHIMath Immediate Eval HooksMap

This section contains a Go package that provides a HooksTable suitable for use
with the Actions section given as an example in the "Immediate Evaluation
Version" section of the FISHIMath example for Actions.

To use it, place the below code into a Go package called "fmhooks". Then,
prepare the FISHI spec using the examples in this document. Next, use ictcc to
build the FISHIMath frontend, and give the arguments `--ir
[]import/path/to/fmhooks.FMValue` and `--hooks filesystem/path/to/fmhooks`. This
will allow you to do full validation testing and create a diagnostics binary
with `-d` that has a fully-featured interpretation engine.


```go
package fmhooks

import (
    "fmt"
    "strings"
    "strconv"
    "math"

    "github.com/dekarrin/ictiobus/trans"
)

var (
    HooksTable = trans.HookMap{
        "identity":             hookIdentity,
        "int":                  hookInt,
        "float":                hookFloat,
        "multiply":             hookMultiply,
        "divide":               hookDivide,
        "add":                  hookAdd,
        "subtract":             hookSubtract,
        "read_var":             hookReadVar,
        "write_var":            hookWriteVar,
        "num_slice_start":      hookNumSliceStart,
        "num_slice_prepend":    hookNumSlicePrepend,
    }
)

var (
    symbolTable = map[string]FMValue{}
)

func hookIdentity(_ trans.SetterInfo, args []interface{}) (interface{}, error) { return args[0], nil }

func hookFloat(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    literalText, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    f64Val, err := strconv.ParseFloat(literalText, 32)
    if err != nil {
        return nil, err
    }
    fVal := float32(f64Val)

    return FMFloat(fVal), nil
}

func hookInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    literalText, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    iVal, err := strconv.Atoi(literalText)
    if err != nil {
        return nil, err
    }

    return FMInt(iVal), nil
}

func hookMultiply(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    v1, v2, err := getBinaryArgsCoerced(args)
    if err != nil {
        return nil, err
    }

    if v1.IsFloat {
        return FMFloat(v1.Float() * v2.Float()), nil
    }
    return FMInt(v1.Int() * v2.Int()), nil
}

func hookDivide(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    v1, v2, err := getBinaryArgsCoerced(args)
    if err != nil {
        return nil, err
    }

    if v1.IsFloat {
        return FMFloat(v1.Float() / v2.Float()), nil
    }
    return FMInt(v1.Int() / v2.Int()), nil
}

func hookAdd(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    v1, v2, err := getBinaryArgsCoerced(args)
    if err != nil {
        return nil, err
    }

    if v1.IsFloat {
        return FMFloat(v1.Float() + v2.Float()), nil
    }
    return FMInt(v1.Int() + v2.Int()), nil
}

func hookSubtract(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    v1, v2, err := getBinaryArgsCoerced(args)
    if err != nil {
        return nil, err
    }

    if v1.IsFloat {
        return FMFloat(v1.Float() - v2.Float()), nil
    }
    return FMInt(v1.Int() - v2.Int()), nil
}

func hookReadVar(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    varName, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    varVal := symbolTable[varName]

    return varVal, nil
}

func hookWriteVar(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    varName, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    varVal, ok := args[1].(FMValue)
    if !ok {
        return nil, fmt.Errorf("arg 2 is not an FMValue")
    }

    symbolTable[varName] = varVal

    return varVal, nil
}

func hookNumSliceStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    v, ok := args[0].(FMValue)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not an FMValue")
    }

    return []FMValue{v}, nil
}

func hookNumSlicePrepend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    vSlice, ok := args[0].([]FMValue)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not an []FMValue")
    }

    v, ok := args[2].(FMValue)
    if !ok {
        return nil, fmt.Errorf("arg 2 is not an FMValue")
    }

    vSlice = append([]FMValue{v}, vSlice...)

    return vSlice, nil
}

func getBinaryArgsCoerced(args []interface{}) (left, right FMValue, err error) {
    v1, ok := args[0].(FMValue)
    if !ok {
        return left, right, fmt.Errorf("arg 1 is not an FMValue")
    }

    v2, ok := args[1].(FMValue)
    if !ok {
        return left, right, fmt.Errorf("arg 2 is not an FMValue")
    }

    // if one is a float, they are now both floats
    if v1.IsFloat && !v2.IsFloat {
        v2 = FMFloat(v2.Float())
    } else if v2.IsFloat && !v1.IsFloat {
        v1 = FMFloat(v1.Float())
    }

    return v1, v2, nil
}

// FMValue is a calculated result from FISHIMath. It holds either a float32 or
// int and is convertible to either. The type of value it holds is querable with
// IsFloat. Int() or Float() can be called on it to get the value as that type.
type FMValue struct {
    IsFloat bool
    i int
    f float32
}

// FMFloat creates a new FMValue that holds a float32 value.
func FMFloat(v float32) FMValue {
    return FMValue{IsFloat: true, f: v}
}

// FMInt creates a new FMValue that holds an int value.
func FMInt(v int) FMValue {
    return FMValue{i: v}
}

// Int returns the value of v as an int, converting if necessary from a float.
func (v FMValue) Int() int {
    if v.IsFloat {
        return int(math.Round(float64(v.f)))
    }
    return v.i
}

// Float returns the value of v as a float32, converting if necessary from an
// int.
func (v FMValue) Float() float32 {
    if !v.IsFloat {
        return float32(v.i)
    }
    return v.f
}

// String returns the string representation of an FMValue.
func (v FMValue) String() string {
    if v.IsFloat {
        str := fmt.Sprintf("%.7f", v.f)
        // remove extra 0's...
        str = strings.TrimRight(str, "0")
        // ...but there should be at least one 0 if nothing else
        if strings.HasSuffix(str, ".") {
            str = str + "0"
        }
        return str
    }
    return fmt.Sprintf("%d", v.i)
}
```

## Appendix B: FISHIMath AST HooksMap

This section contains a Go package that provides a HooksTable suitable for use
with the Actions section given as an example in the "AST Creation Version"
section of the FISHIMath example for Actions.

To use it, place the below code into a Go package called "fmhooks". Then,
prepare the FISHI spec using the examples in this document. Next, use ictcc to
build the FISHIMath frontend, and give the arguments `--ir
import/path/to/fmhooks.AST` and `--hooks filesystem/path/to/fmhooks`. This
will allow you to do full validation testing and create a diagnostics binary
with `-d` that will produce ASTs of input FISHIMath code.



```go
package fmhooks

import (
    "fmt"
    "strings"
    "strconv"
    "math"

    "github.com/dekarrin/ictiobus/trans"
    "github.com/dekarrin/ictiobus/lex"
)

var (
    HooksTable = trans.HookMap{
        "identity":             hookIdentity,
        "var_node":             hookVarNode,
        "assignment_node":      hookAssignmentNode,
        "lit_node_float":       hookLitNodeFloat,
        "lit_node_int":         hookLitNodeInt,
        "group_node":           hookGroupNode,
        "binary_node_mult":     hookFnForBinaryNode(Multiply),
        "binary_node_div":      hookFnForBinaryNode(Divide),
        "binary_node_add":      hookFnForBinaryNode(Add),
        "binary_node_sub":      hookFnForBinaryNode(Subtract),
        "node_slice_start":     hookNodeSliceStart,
        "node_slice_prepend":   hookNodeSlicePrepend,
        "ast":                  hookAST,
    }
)

func hookIdentity(_ trans.SetterInfo, args []interface{}) (interface{}, error) { return args[0], nil }

func hookAST(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    nodeSlice, ok := args[0].([]Node)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a []Node")
    }

    return AST{Statements: nodeSlice}, nil
}

func hookVarNode(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    varName, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    return VariableNode{Name: varName}, nil
}

func hookAssignmentNode(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    varName, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    varVal, ok := args[1].(Node)
    if !ok {
        return nil, fmt.Errorf("arg 2 is not a Node")
    }

    return AssignmentNode{Name: varName, Expr: varVal}, nil
}

func hookLitNodeFloat(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    strVal, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    f64Val, err := strconv.ParseFloat(strVal, 32)
    if err != nil {
        return nil, err
    }

    return LiteralNode{Value: FMValue{vType: Float, f: float32(f64Val)}}, nil
}

func hookLitNodeInt(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    strVal, ok := args[0].(string)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a string")
    }

    iVal, err := strconv.Atoi(strVal)
    if err != nil {
        return nil, err
    }

    return LiteralNode{Value: FMValue{vType: Int, i: iVal}}, nil
}

func hookGroupNode(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    exprNode, ok := args[0].(Node)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a Node")
    }

    return GroupNode{Expr: exprNode}
}

func hookFnForBinaryNode(op Operation) trans.Hook {
    fn := func(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
        left, ok := args[0].(Node)
        if !ok {
            return nil, fmt.Errorf("arg 1 is not a Node")
        }

        right, ok := args[1].(Node)
        if !ok {
            return nil, fmt.Errorf("arg 2 is not a Node")
        }

        return BinaryOpNode{
            Left: left,
            Right: right,
            Op: op,
        }
    }

    return fn
}

func hookNodeSliceStart(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    node, ok := args[0].(Node)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a Node")
    }

    return []Node{node}, nil
}

func hookNodeSlicePrepend(_ trans.SetterInfo, args []interface{}) (interface{}, error) {
    nodeSlice, ok := args[0].([]Node)
    if !ok {
        return nil, fmt.Errorf("arg 1 is not a []Node")
    }

    node, ok := args[1].(Node)
    if !ok {
        return nil, fmt.Errorf("arg 2 is not a Node")
    }

    nodeSlice = append([]Node{node}, nodeSlice...)

    return nodeSlice, nil
}

// AST is an abstract syntax tree containing a complete representation of input
// written in FISHIMath.
type AST struct {
    Statements []Node
}

// String returns a pretty-print representation of the AST, with depth in the
// tree indicated by indent.
func (ast AST) String() string {
    if len(ast.Statements) < 1 {
        return "AST<>"
    }

    var sb strings.Builder
    sb.WriteString("AST<\n")

    labelFmt := " STMT #%0*d: "
    largestDigitCount := len(fmt.Sprintf("%d", len(ast.Statements)))

    for i := range ast.Statements {
        label := fmt.Sprintf(labelFmt, largestDigitCount, i+1)
        stmtStr := spaceIndentNewlines(ast.Statements[i].String(), len(label))
        sb.WriteString(label)
        sb.WriteString(stmtStr)
        sb.WriteRune('\n')
    }

    sb.WriteRune('>')
    return sb.String()
}

// FMString returns a string of FISHIMath code that if parsed, would result in
// an equivalent AST. Each statement is put on its own line, but the last line
// will not end in \n.
func (ast AST) FMString() string {
    if len(ast.Statements) < 1 {
        return ""
    }

    var sb strings.Builder

    for i, stmt := range ast.Statements {
        sb.WriteString(stmt.FMString())
        sb.WriteString(" <o^><")

        if i + 1 < len(ast.Statements) {
            sb.WriteRune('\n')
        }
    }

    return sb.String()
}

// NodeType is the type of a node in the AST.
type NodeType int

const (
    // Literal is a numerical value literal in FISHIMath, such as 3 or 7.284.
    // FISHIMath only supports float32 and int literals, so it will represent
    // one of those.
    Literal NodeType = iota

    // Variable is a variable used in a non-assignment context (i.e. one whose
    // value is being *used*, not set).
    Variable

    // BinaryOp is an operation consisting of two operands and the operation
    // performed on them.
    BinaryOp

    // Assignment is an assingment of a value to a variable in FISHIMath using
    // the value tentacle.
    Assignment

    // Group is an expression grouped with the fishtail and fishhead symbols.
    Group
)

// Operation is a type of operation being performed.
type Operation int

const (
    NoOp Operation = iota
    Add
    Subtract
    Multiply
    Divide
)

// Symbol returns the FISHIMath syntax string that represents the operation. If
// there isn't one for op, "" is returned.
func (op Operation) Symbol() string {
    if op == Add {
        return "+"
    } else if op == Subtract {
        return "-"
    } else if op == Multiply {
        return "*"
    } else if op == Divide {
        return "/"
    }

    return ""
}

func (op Operation) String() string {
    if op == Add {
        return "addition"
    } else if op == Subtract {
        return "subtraction"
    } else if op == Multiply {
        return "multiplication"
    } else if op == Divide {
        return "division"
    }

    return fmt.Sprintf("operation(code=%d)", int(op))
}

// Node is a node of the AST. It can be converted to the actual type it is by
// calling the appropriate function.
type Node interface {
    // Type returns the type of thing this node is. Whether or not the other
    // functions return valid values depends on the type of AST Node returned
    // here.
    Type() NodeType

    // AsLiteral returns this Node as a LiteralNode. Panics if Type() is not
    // Literal.
    AsLiteral() LiteralNode

    // AsVariable returns this Node as a VariableNode. Panics if Type() is not
    // Variable.
    AsVariable() VariableNode

    // AsBinaryOperation returns this Node as a BinaryOpNode. Panics if Type()
    // is not BinaryOp.
    AsBinaryOp() BinaryOpNode

    // AsAssignment returns this Node as an AssignmentNode. Panics if Type() is
    // not Assignment.
    AsAssignment() AssignmentNode

    // AsGroup returns this Node as an GroupNode. Panics if Type() is not Group.
    AsGroup() GroupNode

    // FMString converts this Node into FISHIMath code that would produce an
    // equivalent Node.
    FMString() string

    // String returns a human-readable string representation of this Node, which
    // will vary based on what Type() is.
    String() string
}

// LiteralNode is an AST node representing a numerical constant used in
// FISHIMath.
type LiteralNode struct {
    Value FMValue
}

func (n LiteralNode) Type() NodeType { return Literal }
func (n LiteralNode) AsLiteral() LiteralNode { return n }
func (n LiteralNode) AsVariable() VariableNode { panic("Type() is not Variable") }
func (n LiteralNode) AsBinaryOp() BinaryOpNode { panic("Type() is not BinaryOp") }
func (n LiteralNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n LiteralNode) AsGroup() GroupNode { panic("Type() is not Group") }

func (n LiteralNode) FMString() string {
    return n.Value.String()
}

func (n LiteralNode) String() string {
    return fmt.Sprintf("[LITERAL value=%v]", n.Value)
}

// VariableNode is an AST node representing the use of a variable's value in
// FISHIMath. It does *not* represent assignment to a variable, as that is done
// with an AssignmentNode.
type VariableNode struct {
    Name string
}

func (n VariableNode) Type() NodeType { return Variable }
func (n VariableNode) AsLiteral() LiteralNode { panic("Type() is not Literal") }
func (n VariableNode) AsVariable() VariableNode { return n }
func (n VariableNode) AsBinaryOp() BinaryOpNode { panic("Type() is not BinaryOp") }
func (n VariableNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n VariableNode) AsGroup() GroupNode { panic("Type() is not Group") }

func (n VariableNode) FMString() string {
    return n.Name
}

func (n VariableNode) String() string {
    return fmt.Sprintf("[VARIABLE name=%v]", n.Name)
}

// BinaryOpNode is an AST node representing a binary operation in FISHIMath. It
// has a left operand, a right operand, and an operation to perform on them.
type BinaryOpNode struct {
    Left    Node
    Right   Node
    Op      Operation
}

func (n BinaryOpNode) Type() NodeType { return BinaryOp }
func (n BinaryOpNode) AsLiteral() LiteralNode { panic("Type() is not Literal") }
func (n BinaryOpNode) AsVariable() VariableNode { panic("Type() is not Variable") }
func (n BinaryOpNode) AsBinaryOp() BinaryOpNode { return n }
func (n BinaryOpNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n BinaryOpNode) AsGroup() GroupNode { panic("Type() is not Group") }

func (n BinaryOpNode) FMString() string {
    return fmt.Sprintf("%s %s %s", n.Left.FMString(), n.Op.Symbol(), n.Right.FMString())
}

func (n BinaryOpNode) String() string {
    const (
        leftStart =  " left:  "
        rightStart = " right: "
    )

    leftStr := spaceIndentNewlines(n.Left.String(), len(leftStart))
    rightStr := spaceIndentNewlines(n.Right.String(), len(rightStart))

    return fmt.Sprintf("[BINARY_OPERATION type=%v\n%s%s\n%s%s\n]", n.Op.String(), leftStart, leftStr, rightStart, rightStr)
}

// AssignmentNode is an AST node representing the assignment of an expression to
// a variable in FISHIMath. Name is the name of the variable, Expr is the
// expression being assigned to it.
type AssignmentNode struct {
    Name string
    Expr Node
}

func (n AssignmentNode) Type() NodeType { return Assignment }
func (n AssignmentNode) AsLiteral() LiteralNode { panic("Type() is not Literal") }
func (n AssignmentNode) AsVariable() VariableNode { panic("Type() is not Variable") }
func (n AssignmentNode) AsBinaryOp() BinaryOpNode { panic("Type() is not BinaryOp") }
func (n AssignmentNode) AsAssignment() AssignmentNode { return n }
func (n AssignmentNode) AsGroup() GroupNode { panic("Type() is not Group") }

func (n AssignmentNode) FMString() string {
    return fmt.Sprintf("%s =o %s", n.Name, n.Expr.FMString())
}

func (n AssignmentNode) String() string {
    const (
        exprStart =  " expr:  "
    )

    exprStr := spaceIndentNewlines(n.Expr.String(), len(exprStart))

    return fmt.Sprintf("[ASSIGNMENT name=%q\n%s%s\n]", n.Name, exprStart, exprStr)
}

// GroupNode is an AST node representing an expression grouped by the fishtail
// and fishhead symbols in FISHIMath. Expr is the expression in the group.
type GroupNode struct {
    Expr Node
}

func (n GroupNode) Type() NodeType { return Assignment }
func (n GroupNode) AsLiteral() LiteralNode { panic("Type() is not Literal") }
func (n GroupNode) AsVariable() VariableNode { panic("Type() is not Variable") }
func (n GroupNode) AsBinaryOp() BinaryOpNode { panic("Type() is not BinaryOp") }
func (n GroupNode) AsAssignment() AssignmentNode { panic("Type() is not Assignment") }
func (n GroupNode) AsGroup() GroupNode { return n }

func (n GroupNode) FMString() string {
    return fmt.Sprintf(">{ %s '}", n.Expr.FMString())
}

func (n GroupNode) String() string {
    const (
        exprStart =  " expr:  "
    )

    exprStr := spaceIndentNewlines(n.Expr.String(), len(exprStart))

    return fmt.Sprintf("[GROUP\n%s%s\n]", n.Name, exprStart, exprStr)
}

func spaceIndentNewlines(str string, amount int) string {
    if strings.Contains(str, "\n") {
        // need to pad every newline
        pad := " "
        for len(pad) < amount {
            pad += " "
        }
        str = strings.ReplaceAll(str, "\n", "\n" + pad)
    }
    return str
}

// ValueType is the type of a value in FISHIMath. Only Float and Int are
// supported.
type ValueType int

const (
    // Int is an integer of at least 32-bits.
    Int ValueType = iota

    // Float is a floating point number represented as an IEEE-754 single
    // precision (32-bit) float.
    Float
)

// FMValue is a typed value used in FISHIMath. The type of value it holds is
// querable with Type(). Int() or Float() can be called on it to get the value
// as that type, otherwise Interface() can be called to return the exact value
// as whatever type it is.
type FMValue struct {
    vType ValueType
    i int
    f float32
}

// Int returns the value of v as an int, converting if Type() is not Int.
func (v FMValue) Int() int {
    if v.vType == Float {
        return int(math.Round(float64(v.f)))
    }
    return v.i
}

// Float returns the value of v as a float32, converting if Type() is not Float.
func (v FMValue) Float() float32 {
    if !v.IsFloat {
        return float32(v.i)
    }
    return v.f
}

// Interface returns the value held within this as its native Go type.
func (v FMValue) Interface() interface{} {
    if v.vType == Float {
        return v.f
    }
    return v.i
}

// Type returns the type of this value.
func (v FMValue) Type() ValueType {
    return v.vType
}

// String returns the string representation of an FMValue.
func (v FMValue) String() string {
    if v.vType == Float {
        str := fmt.Sprintf("%.7f", v.f)
        // remove extra 0's...
        str = strings.TrimRight(str, "0")
        // ...but there should be at least one 0 if nothing else
        if strings.HasSuffix(str, ".") {
            str = str + "0"
        }
        return str
    }
    return fmt.Sprintf("%d", v.i)
}
```