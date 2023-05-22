Simple Markdown file that contains a FISHI spec for an addition and
multiplication expression language that is LL(1).

This file is suitable as-is to load as a FISHI spec with `ictcc -qns`. Note that
`-n`/`--no-gen` must be specified as there are additional options that must be
set in order to actually produce a frontend.

### Tokens

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
```

### Grammar

The expression grammar is extremely simple and can be used with any LR parser
as-is.

This defines precedence of operations via production rules. Parnthetical
grouping has the highest precedence, followed by multiplication, followed by
addition.

```fishi
%%grammar

{E}  = {T} {EP}
{EP} = + {T} {EP} | {}
{T}  = {F} {TP}
{TP} = * {F} {TP} | {}
{F}  = id | int | lp {E} rp
```

### Translation Actions

This section defines the actions to take. Each hook function will require an
entry of that name in the HooksTable it declares.

This particular scheme simply provides a value for the entire expression by
evaluating it.

```fishi
%%actions

%symbol {E}
-> {T} {EP}    : {^}.value = eval_chain({0}.value, {1}.op_data)

%symbol {EP}
-> {}          : {^}.op_data = empty_chain()
-> + {T} {EP}  : {^}.op_data = add_chain({1}.value, {2}.op_data)

%symbol {T}
-> {F} {TP}    : {^}.value = eval_chain({0}.value, {1}.op_data)

%symbol {TP}
-> {}          : {^}.op_data = empty_chain()
-> * {F} {TP}  : {^}.op_data = mult_chain({1}.value, {2}.op_data)

%symbol {F}
-> lp {E} rp : {^}.value = identity({1}.value)
-> id        : {^}.value = lookup_value({0}.$text)
-> int       : {^}.value = int({0}.$text)
```