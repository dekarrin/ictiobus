Simple Markdown file that contains a FISHI spec for an addition and
multiplication expression language.

This file is suitable as-is to load as a FISHI spec with `ictcc -qns`. Note that
`-n`/`--no-gen` must be specified as there are additional options that must be
set in order to actually produce a frontend.

### Tokens

The simple expression language has:

* Plus signs, made up of a single `+`.
* Multiplication signs, made up of a single `*`.
* The parentheses characters `(` and `)` for grouping.
* Identifiers, which are made of the characters `A`-`Z`, `a`-`z`, `0`-`9`, and
`_`, but must not start with a digit.
* Integers, which are a sequence of digits.

Additionally, all other whitespace is discarded.

```fishi
%%tokens

\+			              %token +         %human plus sign '+'
\*                        %token *         %human multiplication sign '*'
\(                        %token lp        %human left parenthesis '('
\)                        %token rp        %human right parenthesis ')'
\d+                       %token int       %human integer
[A-Za-z_][A-Za-z_0-9]*    %token id        %human identifier

# ignore whitespace
\s+                       %discard
```

### Grammar

The expression grammar is extremely simple and can be used with any LR parser
as-is.

This defines precedence of operations via production rules. Parnthetical
grouping has the highest precedence, followed by multiplication, followed by
addition.

```fishi
%%grammar

{S} = {S} + {E} | {E}
{E} = {E} * {F} | {F}
{F} = lp {S} rp | id | int
```

### Translation Actions

This section defines the actions to take. Each hook function will require an
entry of that name in the HooksTable it declares.

This particular scheme simply provides a value for the entire expression by
evaluating it.

```fishi
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