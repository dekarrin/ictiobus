Using FISHI
===========

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

## The Preproccessor

WIP - this section probably should be in the ictcc manual and summarized more
succinctly here or not at all. Preprocessing is outside the domain of FISHI
itself.

This section describes the FISHI preprocessor built into ictcc. It isn't
necessarily required to read to understand FISHI itself, and is included mainly
as a reference.

When FISHI is read by ictcc, before it is interpreted by the FISHI frontend, a
preprocessing step is run on input. Since error reporting is handled by the
frontend, which only knows about the preprocessed version of source code, this
means that syntax errors will refer to that modified version instead of directly
to the code that was input. As the preprocessor performs relatively benign
changes, errors are usually easily understandable, but if syntax error output
is confusing, use the -P flag with ictcc to see the exact source code after it
has been preprocessed.

Preprocessing performs a few different functions:

* It pulls all FISHI code out `fishi` code blocks and combines them into a
single FISHI document.
* It normalizes all lines of FISHI to have the line ending `\n`.
* It removes all comments that start with a single "#" up until the end of the
line.
* It converts all double "##" sequences into literal "#" characters.

## Specifying the Lexer with Tokens

The first stage of an Ictiobus frontend is lexing (sometimes referred to as
scanning). This is where code, input as a stream of UTF-8 encoded characters,
is scanned for for recognizable symbols, which are passed to the parsing stage
for further processing. These symbols are called *tokens* - the 'words' of the
input language. Each token has a type, called the *token class*, that is later
used by the parser.

This stage is specified in FISHI in `%%tokens` sections. Each entry in this
section begins with a regular expression pattern that tells the lexer how to
find groups of text, and gives one or more actions the lexer should perform.

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
    ^^^ - this is the pattern

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

### Lexing A New Token

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
note on how it is calculated, followed by "priority"

### Complete Example For FISHI Math

## Specifying the Parser with Grammar

### Symbols

### The Epsilon Production

### Complete Example For FISHI Math

## Specifying the Translation Scheme with Actions

(give typical action)

### Associated Symbol

### AttrRefs

### Shortcuts

### Hooks

### Synthesized vs Inherited Attributes

### Complete Example For FISHI Math


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