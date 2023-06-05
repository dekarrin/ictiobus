# FISHIMath - Immediate Evaluation

This specifies the language "FISHIMath" using a translation scheme that
evaluates end values as they are created. Use with the hooks package located in
the same directory as this file.

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

# This actions section creates an IR that is the result of evaluating the
# expressions. It will return a slice of FMValues, one FMValue for each
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