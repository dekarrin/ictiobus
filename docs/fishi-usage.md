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

### Complete Example For FISHI Math

WIP

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
using special BNF-ish syntax 

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
    {PRODUCT} * {TERM}            | {PRODUCT}  =  {PROUDCT} * {TERM}
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

### Complete Example For FISHI Math

WIP

## Specifying the Translation Scheme with Actions

The third and final stage of an Ictiobus frontend is translation. This is where
a *syntax-directed translation scheme* (or "SDTS") is applied to nodes of the
parse tree created in the parsing phase to produce the final result, the
*intermediate representation* (or "IR") of the input. This IR is then returned
to the user of the frontend as the return value of `Frontend.Analyze()`.

### FISHI Actions Quick Reference

WIP 

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
    -> {SUM} + {TERM}: {^}.val = add({0}.val, {1}.val)
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
absolutely no semantic meaning anywhere and is completely ignored.

WIP NEED TO FIND PLACE TO MENTION S-ATTRIBUTED ONLY

### The Actions Entry

Entries in a FISHI Actions section begin with an associated symbol for the
entry. This is specified with the `%symbol` directive followed by the
non-terminal symbol at the head of the grammar rule that attributes are going to
be defined for.

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.value, {1}.value)
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

    %symbol {SUM} -> {SUM} + {TERM}   : {^}.val = add({0}.value, {1}.value)
                  ^^^^^^^^^^^^^^^^^   ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        Production action set start   SDD

There can be multiple production action sets for a single head symbol, each of
which gives SDDs for the production they specify:

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.val, {1}.val)
                  -> {TERM}         : {^}.val = identity({0}.val)

Each SDD for the production starts with a `:` sign (or the outdated `%set`
keyword). SDDs use a somewhat complex syntax that will be covered shortly in the
"SDDs" section below. There can be multiple SDDs per production. If there are,
they will be applied in the order they are defined or the order needed to
satisfy attribute calculations that depend on them.

    %%actions

    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.val, {1}.val)
                                    : {^}.val = set_plus_op()
                  -> {TERM}         : {^}.val = identity({0}.val)

Because whitespace is ignored in `%%actions` sections, they can be formatted in
any way that the user finds readable.

    # this:
    %symbol {SUM} -> {SUM} + {TERM} : {^}.val = add({0}.val, {1}.val)
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

Each 

### AttrRefs
### Hooks

### Synthesized vs Inherited Attributes

### Abstract Syntax Trees

This IR value can be anything that you wish it to be. For simpler languages,
such as FISHIMath, it can be immediately calculated by evaluating mathematical
options to build up a final result. For others, a special representation of the
input code called an *abstract syntax tree* might be built up. An abstract
syntax tree difers from a parse tree by abstracting the grammatical details
of how the constructs within the code were arranged into a more natural
representation that reflects the logical structures of the language, such as
control structures. It is usually suitable for further evaluation by an
interpretion engine.

```go
if x >= 8 && y < 2 {
    fmt.Printf("It's true!\n")
}
```

The above block of Go syntax might be parsed into the following tree:

    IF-BLOCK
       |
    
    if COND



### Complete Example For FISHI Math