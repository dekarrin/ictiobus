# Frontend Instruction Specification for self-Hosting Ictiobus (FISHI), v1.0
This is a grammar for the Frontend Instruction Specification for
(self-)Hosting Ictiobus (FISHI). It is for version 1.0 and this version was
started on 2/18/23 by Jello! Glubglu8glub 38D

## Escape Sequence
To escape something that would otherwise have special meaning in FISHI, use the
escape sequence directly before it, `%!`.

## Format Use
Languages that describe themselves in the FISHI language are taken from
definitions described with FISHI for the frontend of Ictiobus and used to
produce a compiler frontend.

These definitions are to be embedded in Markdown-formatted text in special code
blocks delimited with the triple-tick that are marked with the special syntax
tag `fishi`, as in the following:

    ```fishi
    (FISHI directives would go here)
    ```

Multiple consecutive `fishi` code blocks in the same file are appended together
to create the full source that is parsed.

## Parser
This is the context-free grammar for FISHI, glub.

```fishi
%%grammar

{FISHISPEC}        =  {BLOCKS}

{BLOCKS}           =  {BLOCKS} {BLOCK} | {BLOCK}

{BLOCK}            =  {GBLOCK} | {TBLOCK} | {ABLOCK}

# actions branch:

{ABLOCK}           =  hdr-actions {ACONTENT}

{ACONTENT}         =  {SYM-ACTIONS-LIST} {ASTATE-SET-LIST}
                   |  {SYM-ACTIONS-LIST}
                   |  {ASTATE-SET-LIST}

{ASTATE-SET-LIST}  =  {ASTATE-SET-LIST} {ASTATE-SET} | {ASTATE-SET}

{ASTATE-SET}       =  {STATE-INS} {SYM-ACTIONS-LIST}

{SYM-ACTIONS-LIST} =  {SYM-ACTIONS-LIST} {SYM-ACTIONS} | {SYM-ACTIONS}

{SYM-ACTIONS}      =  dir-symbol nonterm {PROD-ACTION-LIST}

{PROD-ACTION-LIST} =  {PROD-ACTION-LIST} {PROD-ACTION} | {PROD-ACTION}

{PROD-ACTION}      =  {PROD-SPEC} {SEM-ACTION-LIST}

{SEM-ACTION-LIST}  =  {SEM-ACTION-LIST} {SEM-ACTION} | {SEM-ACTION}

{SEM-ACTION}       =  dir-set attr-ref dir-hook id
                   |  dir-set attr-ref dir-hook id {WITH}

{WITH}             =  dir-with ATTR-REF-LIST

{ATTR-REF-LIST}    =  {ATTR-REF-LIST} attr-ref
                   |  attr-ref

{PROD-SPEC}        =  dir-prod {PROD-ADDR}
                   |  dir-prod

{PROD-ADDR}        =  dir-index int
                   |  {APRODUCTION}

{APRODUCTION}      =  {ASYM-LIST} | epsilon

{ASYM-LIST}        =  {ASYM-LIST} {ASYM} | {ASYM}

{ASYM}             =  nonterm | term | int | id

# tokens branch:

{TBLOCK}           = hdr-tokens {TCONTENT}

{TCONTENT}         =  {TENTRY-LIST} {TSTATE-SET-LIST}
                   |  {TENTRY-LIST}
                   |  {TSTATE-SET-LIST}

{TSTATE-SET-LIST}  =  {TSTATE-SET-LIST} {TSTATE-SET} | {TSTATE-SET}

{TSTATE-SET}       =  {STATE-INS} {TENTRY-LIST}

{TENTRY-LIST}      =  {TENTRY-LIST} {TENTRY} | {TENTRY}

{TENTRY}           =  {PATTERN} {TOPTION-LIST}

{TOPTION-LIST}     =  {TOPTION-LIST} {TOPTION} | {TOPTION}

{TOPTION}          =  {DISCARD} | {STATESHIFT} | {TOKEN} | {HUMAN} | {PRIORITY}
{DISCARD}          =  dir-discard
{STATESHIFT}       =  dir-shift {TEXT}
{TOKEN}            =  dir-token {TEXT}
{HUMAN}            =  dir-human {TEXT}
{PRIORITY}         =  dir-priority {TEXT}  # TODO: shouldn't this be int?

{PATTERN}          =  {TEXT}

# grammar branch:

{GBLOCK}           =  hdr-grammar {GCONTENT}

{GCONTENT}         =  {GRULE-LIST} {GSTATE-SET-LIST}
                   |  {GRULE-LIST}
                   |  {GSTATE-SET-LIST}

{GSTATE-SET-LIST}  =  {GSTATE-SET-LIST} {GSTATE-SET} | {GSTATE-SET}

{GSTATE-SET}       =  {STATE-INS} {GRULE-LIST}

{GRULE-LIST}       =  {GRULE-LIST} {GRULE} | {GRULE}

{GRULE}            =  nl-nonterm eq {ATERNATIONS}

{ALTERNATIONS}     =  {PRODUCTION}
                   |  {ALTERNATIONS} alt {GPRODUCTION}

{GPRODUCTION}      =  {GSYM-LIST} | epsilon

{GSYM-LIST}        =  {GSYM-LIST} {GSYM} | {GSYM}

{GSYM}             =  nonterm | term

# state instruction def

{STATE-INS}        =  dir-state {ID-EXPR}
{ID-EXPR}          =  id | term

# text block glueing and ensuring it only goes until end of line unless escaped:

{TEXT}             =  {NL-TEXT-ELEM} {TEXT-ELEM-LIST}
                   |  {TEXT-ELEM-LIST}
                   |  {NL-TEXT-ELEM}
{TEXT-ELEM-LIST}   =  {TEXT-ELEM-LIST} {TEXT-ELEM} | {TEXT-ELEM}

{NL-TEXT-ELEM}     =  nl-escseq | nl-freeform-text
{TEXT-ELEM}        =  escseq    | freeform-text

```

## Lexer
The following gives the lexical specification for the FISHI language.

For all states:

```fishi
%%tokens

%!%!.                                  %token escseq
%human Escape Sequence

%!%%!%[Tt][Oo][Kk][Ee][Nn][Ss]         %token hdr-tokens
%human %!%%!%tokens header             %stateshift TOKENS

%!%%!%[Gg][Rr][Aa][Mm][Mm][Ee][Rr]     %token hdr-grammar
%human %!%%!%grammar header            %stateshift GRAMMAR

%!%%!%[Aa][Cc][Tt][Ii][Oo][Nn][Ss]     %token hdr-actions 
%human %!%%!%actions header            %stateshift ACTIONS
```

For tokens state:

```fishi
%state TOKENS

\n\s*%!%!.                                         %token nl-escseq
%human escape sequence after this line

%!%[Ss][Tt][Aa][Tt][Ee]                            %token dir-state
%human %!%state directive                          %stateshift STATE-T

%!%[Ss][Tt][Aa][Tt][Ee][Ss][Hh][Ii][Ff][Tt]        %token dir-shift
%human %!%stateshift directive                     %priority 1

%!%[Hh][Uu][Mm][Aa][Nn]                            %token dir-human
%human %!%human directive

%!%[Tt][Oo][Kk][Ee][Nn]                            %token dir-token
%human %!%token directive

%!%![Dd][Ii][Ss][Cc][Aa][Rr][Dd]                   %token dir-discard
%human %!%discard directive

%!%[Pp][Rr][Ii][Oo][Rr][Ii][Tt][Yy]                %token dir-priority
%human %!%priority directive

[^\S\n]+                                           %discard

\n\s*[^%!%\s]+[^%!%\n]*                            %token nl-freeform-text
%human freeform text after this line

\n                                                 %discard

[^%!%\s]+[^%!%\n]*                                 %token freeform-text
%human freeform text
```

For grammar state:

```fishi
%state GRAMMAR

%!%[Ss][Tt][Aa][Tt][Ee]  %token dir-state  %stateshift STATE-G
%human %!%state directive     

[^\S\n]+                 %discard

\n\s*{[A-Za-z][^}]*}     %token nl-nonterm
%human non-terminal symbol literal after this line

\n                       %discard
\|                       %token alt     %human alternations bar '|'
{}                       %token epsilon %human epsilon production '{}'
{[A-Za-z][^}]*}          %token nonterm %human non-terminal symbol literal
[^=\s]\S*|\S\S+          %token term    %human terminal symbol literal
=                        %token eq      %human rule production operator '='
```

For actions state:
```fishi
%state ACTIONS

\s+                      %discard

(?:{\^}|{\*}|{[A-Za-z][^{}]*}|[^\s{}]+)(?:\$\d+)?\.[\$A-Za-z][$A-Za-z0-9_-]*
%token attr-ref    %human attribute reference literal

[0-9]+
%token int         %human integer literal

{[A-Za-z][^}]*}
%token nonterm   # human already defined so should be able to skip it

%!%[Ss][Tt][Aa][Tt][Ee]
%token dir-state   %stateshift STATE-A

%!%[Ss][Yy][Mm][Bb][Oo][Ll]
%token dir-symbol  %human %!%symbol directive

%!%[Pp][Rr][Oo][Dd]
%token dir-prod    %human %!%prod directive

(?:\(|%!%[Ww][Ii][Tt][Hh])
%token dir-with    %human %!%with directive '('

(?:=|%!%[Hh][Oo][Oo][Kk])
%token dir-hook    %human %!%hook directive '='

\)                 %discard
,                  %discard

%!%[Ss][Ee][Tt]
%token dir-set     %human %!%set directive

%!%[Ii][Nn][Dd][Ee][Xx]
%token dir-index   %human %!%index directive

[A-Za-z][A-Za-z0-9_-]*
%token id          %human identifier

{}
%token epsilon

\S+
%token term
```

Because we don't have a state stack yet:

```fishi
%state STATE-T
\s+        %discard
[A-Za-z][A-Za-z0-9_-]*      %token id     %stateshift TOKENS

%state STATE-A
\s+        %discard
[A-Za-z][A-Za-z0-9_-]*      %token id     %stateshift ACTIONS

%state STATE-G
\s+        %discard
[A-Za-z][A-Za-z0-9_-]*      %token id     %stateshift GRAMMAR
```

## Translation Scheme
The following gives the Syntax-directed translations for the FISHI language.

```fishi
%%actions

%symbol {FISHISPEC}
%prod %set      {^}.ast = make_fishispec({BLOCKS}.value)

%symbol {BLOCKS}
%prod {BLOCKS} {BLOCK}
%set {^}.value = block_list_append({BLOCKS}$1.value, {BLOCK}.ast)

%prod %index 1
%set {^}.value = block_list_start({BLOCK}$1.value)

# TODO: add %prod %all selection.
%symbol {BLOCK}
%prod %set {^}.ast = ident({*}.ast)
%prod %set {^}.ast = ident({*}.ast)
%prod %set {^}.ast = ident({*}.ast)


%symbol {ABLOCK}
%prod %set {^}.ast = make_ablock({*}$1.ast)

%symbol {TBLOCK}
%prod %set {^}.ast = make_tblock({*}$1.ast

%symbol {GBLOCK}
%prod %set {^}.ast = make_gblock({*}$1.ast)

%symbol {TCONTENT}
%prod %set {^}.ast = tokens_content_blocks_start_entry_list({*}.value)
%prod %set {^}.ast = ident({*}.value)
%prod %set {^}.ast = tokens_content_blocks_prepend(
                        {TSTATE-SET-LIST}.value,
                        {TENTRY-LIST}.value
                     )

%symbol {ACONTENT}
%prod %set {^}.ast = actions_content_blocks_start_sym_actions({*}.value)
%prod %set {^}.ast = ident({*}.value)
%prod %set {^}.ast = actions_content_blocks_prepend(
                        {ASTATE-SET-LIST}.value,
                        {SYM-ACTIONS-LIST}.value
                     )

%symbol {GCONTENT}
%prod %set {^}.ast = grammar_content_blocks_start_rule_list({*}.value)
%prod %set {^}.ast = ident({*}.value)
%prod %set {^}.ast = grammar_content_blocks_prepend(
                        {GSTATE-SET-LIST}.value
                        {GRULE-LIST}.value
                     )


%symbol {GSTATE-SET}
%prod %set {^}.value = make_grammar_content_node(
                        {STATE-INS}.state
                        {GRULE-LIST}.value
                     )

%symbol {ASTATE-SET}
%prod %set {^}.value = make_actions_content_node(
                        {STATE-INS}.state
                        {SYM-ACTIONS-LIST}.value
                     )

%symbol {TSTATE-SET}
%prod %set {^}.value = make_tokens_content_node(
                        {STATE-INS}.state
                        {TENTRY-LIST}.value
                       )

%symbol {PROD-ACTION-LIST}
%prod %set {^}.value = prod_actions_list_append(
                        {PROD-ACTION-LIST}.value
                        {PROD-ACTION}.value
                       )
%prod %set {^}.value = prod_action_list_start({PROD-ACTION}.value)


%symbol {ATTR-REF-LIST}
%prod %set {^}.value = attr_ref_list_append({*}$0.value, {*}$1.$text)
%prod %set {^}.value = attr_ref_list_start({*}.value)

%symbol {WITH}
%prod %set {^}.value = ident({*}$1.value)

%symbol {SEM-ACTION}
%prod %set {^}.value = make_semantic_action({*}$1.$text, {*}$3.$text)
%prod %set {^}.value = make_semantic_action
                        ({*}$1.$text, {*}$3.$text, {*}$4.value)

```