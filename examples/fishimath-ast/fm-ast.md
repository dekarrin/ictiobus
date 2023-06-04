# FISHIMath - Immediate Evaluation

This specifies the language "FISHIMath" using a translation scheme that builds
an AST and returns that. Use with the hooks package located in the same
directory as this file.

## Spec

### Tokens

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
[0-9]*\.[0-9]+           %token float        %human floating-point literal
[0-9]+                   %token int          %human integer literal
```

### Grammar

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

### Actions

```fishi
%%actions

# This actions section creates an abstract syntax tree and returns it as the
# IR. No evaluation of the statements is performed.

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