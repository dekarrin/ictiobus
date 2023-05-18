Using FISHI
###########

The Frontend Instruction Specification for Languages Hosted on Ictiobus (or just
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

## The Three Phases Of A Compiler Frontend

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

The phases of a frontend:
```
              I N P U T
          (text in language)

             "(3+8) * 2"
             
                  |
                  V
             [L E X E R]
                  |
                  V
               (tokens)

  (lparen "("), (int "3"), (plus "+"),
  (int "8"), (rparen ")"), (times "*"),
              (int "2")
                 
                  |
                  V
            [P A R S E R]
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
[T R A N S L A T I O N   S C H E M E]
                  |
                  V
    (intermediate representation)

                  22

             O U T P U T
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
    # FISHI source code goes here, commented with the hash character
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



## Specifying the Lexer with Tokens

### The Pattern

### Lexing A New Token
%tok
%hum
### Discarding Input
%discard

### Pattern Priority
note on how it is calculated, followed by "priority"

## Specifying the Parser with Grammar

### Symbols

### The Epsilon Production

## Specifying the Translation Scheme with Actions

(give typical action)

### Associated Symbol

### AttrRefs

### Shortcuts

### Hooks

### Synthesized vs Inherited Attributes


##### (old content below this)
*NOTE: this document is being kept for historical purproses, and some content
may be correct, but it was the first attempt to standardize the fishi language
and is heavily out of date. Refer to fishi.md instead of this file for an
example; the correct parts of this file will eventually be worked into the manual
for FISHI.*

This is a complete example of a language specified in the special ictiobus
format. It is a markdown based format that uses specially named sections to
allow both specificiation of a language and freeform text to co-exist.

It will process the sections in order, and parse only those sections in code
fences with the `ictio` syntax specifier.

Specifying Terminals
--------------------
One of the first steps in any language specification is to give the patterns
that the lexer needs to identify tokens in source text. Ictiobus allows this to
be done programmatically if desired, or in a markdown file code block that
starts with the line `%%tokens`.

```ictio
# Comments can be included in any line in an ictio block with the '#' character.
# All content from the first # encountered until the end of the line are
# considered a comment.
#
# If a literal '#' is needed at any point in the ictio source, it can be escaped
# with a backslash.

%%tokens

# if a %state X sub-section is found in a %%token section, then all entries will
# be considered to apply only to state X of the lexer. The first state given
# is considered to be the starting state; otherwise, %start Y can be given at
# any point in a %%token section (but before the first %state) to say that the
# lexer should begin in that state.
#
# if no states are explicitly given, all rules are assumed to apply to the same,
# default lexer state.
#
# State names are case-insensitive.

%state normal

[A-Za-z_][A-Za-z0-9_]+    %token identifier    %human "identifier sequence"
"(?:\\\\|\\"|[^"])+"      %token str           %human "string"
\d+                       %token int           %human "integer"
\+                        %token +             %human "'+'"
\*                        %token *             %human "'*'"
\(                        %token (             %human "'('"
\)                        %token )             %human "')'"
<                         %token <             %human "'<'"    %stateshift angled

# this state shift sequence is very contrived and could easily be avoided by
# simply not state-shifting on <, but it makes for a good example and test.

%state angled

# multiple %human definitions are allowed if the others are in another state
# we'll use this state to create some of the same tokens with slightly different
# human-readable names to reflect that they were in another state, but by using
# the same token class names, we can treat them all the same regardless of which
# state they were produced in in the cfg.

>                         %token >             %human "'>'"    %stateshift normal
"(?:\\\\|\\"|[^"])+"      %token str           %human "angly string"
\d+                       %token int           %human "angly int"
,                         %token ,             %human ","
```

All lexer specifications begin with a pattern in regular expression RE2 syntax.
As the lexer uses the built-in RE2 engine in Go to analyze input and identify
tokens and state shifts, these must follow the allowed syntax given in the
[RE2 Specification](https://github.com/google/re2/wiki/Syntax).

As lookaheads and lookbehinds are not supported at this time, a single capture
group may be given within the regex that gives the lexeme of interest; whatever
is captured in this group will be what becomes the body of a parsed token. Note
that if this method is used, when advancing token input, ONLY the characters
that come before the subexpression and those captured in the subexpression
itself will be considered 'processed'; any that match after the capturing group
will be processed during the next attempt to read a token.

A subexpression is optional; if one is not provided, the entire pattern is
considered to be of interest. If there is more than 1 subexpression in a
pattern, it is an error, and the specification will not be accepted.
Non-capturing groups are not subject to this limitation and may be given as many
times as desired.

For example, if the input being processed is `int3alpha double`, then the
pattern `[A-Za-z]+(\d)alpha` would match the `int3alpha` and the `3` would be
set as the lexeme parsed for the associated token. The `[A-Za-z]+` in the
pattern that comes directly before the capturing group would match the `int`,
and the `alpha` part of the pattern would match the literal `alpha` in the
input. `int` comes before the capture group that received `3`, and so the input
would be advanced past `int3` and the next lexical analysis pass would attempt
to match against `alpha double`.

Patterns can embed previously-defined patterns within them. Just wrap the ID of
the pattern within '{' and '}' in the regex. The ID is automatically generated
for each pattern as simply the order that it occurs within an ictiobus
specification, starting with 1. So the first pattern is {1}, the second is {2},
etc. To use a literal '{' or '}' within a pattern, escape them with a backslash.

After the pattern, a sequence of lexer directives are given that tell the
lexer what to do on matching the pattern.

They each begin with a keyword starting with %, and can be specified in any
order after the pattern. To use a % literally in a pattern, escape it with a
backslash.

The most common directive is `%token`, which tells the lexer to take the input
that the pattern matched (or the input that is in the capturing group in the
regex, if one is specified), and create a new token. The name of the class of
that token is given after `%token`; it can contain any characters (but a %
must be escaped) but is case-insensitive; conventionally, they are given as
lower-case, as internally they will be converted to lowercase.

* `%token CLASS` - Specifies to take the matched input and use it to create a
new token of the given class. `CLASS` may be any characters but note that they
are case-insenseitve and will be represented internally as lower-case. The same
token class may be given for different patterns; this means that a match against
any of those patterns will result in a token of that class.
* `%human HUMAN-READABLE NAME` - Specifies that the token class given should
have the human-readable name given between the two quotes. This is generally
used for reporting errors related to that token. This human readable name is
associated with the token named by the `%token` directive on the same pattern;
if no `%token` is given, it is an error. The `%human` part of the token is
shared by all uses of that token class; this means that `%human` may only be
specified one time for a given `%token` class. It is an error to have multiple
`%human` directives across patterns with the same token class name. If desired
this can be handled entirely in a `%tokendef class "human"` directive outside
of any pattern.
* `%changestate STATE` - Specifies that the lexer should swap state to STATE
and then continue lexing. If in the same rule as a `%token`, the token will be
lexed in the prior state before swapping to the new one.

Specifying Grammar
------------------
The grammar of the language accepted by the parsing phase is given in a
`%%grammar` section in an `ictio` code block. In that section is a context-free
grammar that uses a "BNF-ish" notation. In this notation, a non-terminal is
defined by a symbol name on the left side enclosed in curly-braces `{` and `}`,
followed by space and the character `=`, followed by one or more productions
for the non-terminal separated by a `|` character. Non-terminal names must start
with a letter of the alphabet but besides that can contain any symbol and are
not case-sensitive. Internally they are represented as upper-case.

Terminals are specified by giving the name of the token class that makes the
terminal. This can be any `%token` class name defined in a `%%tokens` section.
To specify a token that is one of the reserved characters `{`, `}`, `=`, `|`,
use a backslash character to escape it.

The start production of the language is the first non-terminal that is defined
in a `%%grammar` section.

```ictio
%%grammar

{S}       =   {expr}
{expr}    =   {math}
{math}    =   {sum}
```

Alternations may be on same line, as in this two sequence:

```ictio
%%grammar

{sum}     =   {mult} + {sum} | {mult}
```

Alternations may also be split across lines:

```ictio
{mult}    =   {value} * {mult}
          | {parens}
```

Thhe indent level does not matter, nor does whitespace between symbols:

```ictio
    {parens}   =   ( {expr}) |{value}
```

The only exception to the whitespace meaning is that of a newline; if the first
symbol at the start of a line is a non-terminal, it is the start of a new
definition. If the first symbol is a `|` character, it is an alternation from
the prior rule.

The first symbol on a line must be one of those two; a non-terminal is not
allowed:

```
(example block, not marked with `ictio`, not parsed, 38O glub)

# not allowed:
{EXAMPLE} = {EXAMPLE}
            EXAMPLE_TERMINAL
```

Note that the symbol after is considered part of the alternation. To include the
empty string as part of the productions (the epsilon production), simply include
a bar at the end of a series of alternations with no productions. There are two
forms:


```
(example block, not marked with `ictio`, not parsed, 38O glub)

# empty string specified by having a `|` at the end of a line:
{EXAMPLE}    =   {EXAMPLE}  |

# empty string specified by having a `|` on a line with no other contents:
{EXAMPLE}    =   {EXAMPLE}
             |
```

```ictio

{value}          =   {primitive} | {list}
{list}           =   < {list-contents} >
{list-contents}  =   {expr} | {list-contents} , {expr}

{primitive}      =   STR |INT|ID
```

Specifying Semantic Actions
---------------------------
Semantic actions in Ictiobus are defined by the `%%actions` directive.

```ictio
%%actions

%symbol {primitive}
%prod STR
%action

prim$0.val = prim_str( STR.$text )

# auto-numbering of productions if not explicit
%action
prim$0.val = prim_int( INT.$text )

%prod %index 2  # or can be specified with prod index, 0-indexed

# %action is followed by an attr reference that is to be defined. It must be
# either the head of the production (to set a synthesized attribute value) or a
# non-terminal production in the body.
# 
%action prim$0.val

# hook must be an identifier [A-Za-z_][A-Za-z_0-9]*. It's the name of the func
# passed to 'register' of the Frontend.
%hook prim_id

# %with must be an attr reference. an attr reference starts with the name of the
# symbol in the production that has the attribute being referenced. If there is
# both a terminal and non-terminal defined with the same name, the terminal is
# selected by default. To specify a non-terminal specifically, wrap it in {}.
# The name may be abbreviated to the shortest disambigauting prefix of the thing
# it refers to. Within this section, both the `$` and the `.` characters have
# special meaning which can be avoided by escaping with backslash.
#
# Optionally, after either a terminal name or a non-terminal name, the `$` sign
# can be given, after which is an index. The index is a number that specifies
# the instance of this symbol in the production, numbered left to right and
# 0-indexed. The left-most instance of the production symbol is the left-hand
# of the production itself, so that is index 0.
#
# After the prior section comes a period character `.`, after which follows the
# name of the attribute, which must be identifier pattern [A-Za-z_][A-Za-z_0-9]*
# and is case-insensitive.

%with ID.$text

```