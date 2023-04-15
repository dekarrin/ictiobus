# FISHI v1.0 Language Specification

This file describes and specifies the FISHI language, which is the language used
to specify languages that ictiobus is to create compilers for. So, this file is
both the manual for use and the self-description used by Ictiobus to generate
its own frontend.

FISHI stands for the  Frontend Instruction Specification for (self-)Hosting
Ictiobus. The name is a little contrived but that was the best way to get it to
spell out something ocean-related, which was considered of paramount importance
by the initial creators of this language, and also maintains consistency with
the naming of `ictiobus` after the buffalofish.

This specification is for version 1.0 of FISHI.

## Version History
* 2/18/23 - (v1.0, first draft) - The initial version of this document, started
by Jello! Glubglu8glub 38D. This was an initial version that changed somewhat
rapidly as the language developed, and was never really 'completed' in the sense
of being published at a specific version.
* 3/27/23 - (v1.0, second draft) - The second version of this document which was
produced after a bootstrap frontend was created manually. This frontend was able
to read in Fishi code and display the AST of it. Therefore, this document was
updated to reflect that frontend, adjusting both it and the frontend as
necessary to ensure it could continue to read it.

## Overview

FISHI is embedded in Markdown files to allow natural language descriptions of
the languages they describe as well as particular FISHI blocks themselves. FISHI
is read from code blocks (blocks delimited with the triple-tick) whose syntax
tag is set to `fishi`, as in the following example:

    ```fishi
    (FISHI directives would go here)
    ```

Multiple `fishi` code blocks are concatenated together and the resulting block
of FISHI source code is then compiled to a compiler for the language it
describes.

### Escape Sequences

To escape something that would otherwise have special meaning in FISHI, use the
escape sequence directly before it, `%!`. This does not work in all situations,
and is exclusively used in blocks of free-form text to do things such as mark a
percent as a literal one and escape the end-of-line.

If a terminal name would confict with a built-in FISHI operator, a different
terminal name that does not conflict should be used instead of trying to escape
it.

The backslash was avoided as the escape character due to its common use in
regular expressions, which are used heavily in FISHI lexical specifications.
This allows the backslash to be used in regular expressions without having to
escape it every time.

### Case-Sensitivity

With few exceptions, most things in FISHI are case-sensitive. However,
non-terminals will be converted to all uppercase and terminal names will be
converted to all lowercase in the built grammar.

### Basic Layout

This section needs to be expanded; for now, see the below section for an example
spec written in FISHI, which describes the FISHI language.

## Specification

This section contains the formal specification for the FISHI language. It can be
parsed by `ictcc` to create a new FISHI frontend, although do note that a very
particular invocation is required to completely replace the existing one.

### Parser
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

{WITH}             =  dir-with {ATTR-REF-LIST}

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
{PRIORITY}         =  dir-priority {TEXT}

{PATTERN}          =  {TEXT}

# grammar branch:

{GBLOCK}           =  hdr-grammar {GCONTENT}

{GCONTENT}         =  {GRULE-LIST} {GSTATE-SET-LIST}
                   |  {GRULE-LIST}
                   |  {GSTATE-SET-LIST}

{GSTATE-SET-LIST}  =  {GSTATE-SET-LIST} {GSTATE-SET} | {GSTATE-SET}

{GSTATE-SET}       =  {STATE-INS} {GRULE-LIST}

{GRULE-LIST}       =  {GRULE-LIST} {GRULE} | {GRULE}

{GRULE}            =  nl-nonterm eq {ALTERNATIONS}

{ALTERNATIONS}     =  {GPRODUCTION}
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

### Lexer
The following gives the lexical specification for the FISHI language.

For all states:

```fishi
%%tokens

%!%!.                                  %token escseq
%human Escape Sequence

%!%%!%[Tt][Oo][Kk][Ee][Nn][Ss]         %token hdr-tokens
%human %!%%!%tokens header             %stateshift TOKENS

%!%%!%[Gg][Rr][Aa][Mm][Mm][Aa][Rr]     %token hdr-grammar
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

%!%[Dd][Ii][Ss][Cc][Aa][Rr][Dd]                   %token dir-discard
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

(?:{(?:&|\.)(?:[0-9]+)?}|{[0-9]+}|{\^}|{[A-Za-z][^{}]*}|[^\s{}]+)\.[\$A-Za-z][\$A-Za-z0-9_]*
%token attr-ref    %human attribute reference literal

[0-9]+
%token int         %human integer literal

{[A-Za-z][^}]*}
%token nonterm   # human already defined so should be able to skip it

%!%[Ss][Tt][Aa][Tt][Ee]
%token dir-state   %stateshift STATE-A

%!%[Ss][Yy][Mm][Bb][Oo][Ll]
%token dir-symbol  %human %!%symbol directive

(?:->|%!%[Pp][Rr][Oo][Dd])
%token dir-prod    %human %!%prod directive '->'

(?:\(|%!%[Ww][Ii][Tt][Hh])
%token dir-with    %human %!%with directive '('

(?:=|%!%[Hh][Oo][Oo][Kk])
%token dir-hook    %human %!%hook directive '='

\)                 %discard
\(\)               %discard
,                  %discard

(?::|%!%[Ss][Ee][Tt])
%token dir-set     %human %!%set directive ':'

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

### Syntax-Directed Translation Scheme
The following gives the Syntax-directed translations for the FISHI language.

```fishi
%%actions

%symbol {FISHISPEC} ->: {^}.ast = make_fishispec({BLOCKS}.value)

%symbol
{BLOCKS} -> {BLOCKS} {BLOCK}
: {^}.value = block_list_append({BLOCKS}.value, {BLOCK}.ast)

         -> %index 1
: {^}.value = block_list_start({BLOCK}.ast)


# TODO: add %prod %all selection.
%symbol {BLOCK}
->: {^}.ast = ident({0}.ast)
->: {^}.ast = ident({0}.ast)
->: {^}.ast = ident({0}.ast)


%symbol {ABLOCK}
->: {^}.ast = make_ablock({1}.ast)

%symbol {TBLOCK}
->: {^}.ast = make_tblock({1}.ast)

%symbol {GBLOCK}
->: {^}.ast = make_gblock({1}.ast)

%symbol {TCONTENT}
->: {^}.ast = tokens_content_blocks_prepend(
                        {TSTATE-SET-LIST}.value,
                        {TENTRY-LIST}.value
                     )
->: {^}.ast = tokens_content_blocks_start_entry_list({0}.value)
->: {^}.ast = ident({0}.value)

%symbol {ACONTENT}
->: {^}.ast = actions_content_blocks_prepend(
                        {ASTATE-SET-LIST}.value,
                        {SYM-ACTIONS-LIST}.value
                     )
->: {^}.ast = actions_content_blocks_start_sym_actions({0}.value)
->: {^}.ast = ident({0}.value)

%symbol {GCONTENT}
->: {^}.ast = grammar_content_blocks_prepend(
                        {GSTATE-SET-LIST}.value
                        {GRULE-LIST}.value
                     )
->: {^}.ast = grammar_content_blocks_start_rule_list({0}.value)
->: {^}.ast = ident({0}.value)


%symbol {GSTATE-SET}
->: {^}.value = make_grammar_content_node(
                        {STATE-INS}.state
                        {GRULE-LIST}.value
                     )

%symbol {ASTATE-SET}
->: {^}.value = make_actions_content_node(
                        {STATE-INS}.state
                        {SYM-ACTIONS-LIST}.value
                     )

%symbol {TSTATE-SET}
->: {^}.value = make_tokens_content_node(
                        {STATE-INS}.state
                        {TENTRY-LIST}.value
                       )

%symbol {PROD-ACTION-LIST}
->: {^}.value = prod_action_list_append(
                        {PROD-ACTION-LIST}.value
                        {PROD-ACTION}.value
                       )
->: {^}.value = prod_action_list_start({PROD-ACTION}.value)


%symbol {ATTR-REF-LIST}
->: {^}.value = attr_ref_list_append({0}.value, {1}.$text, {1}.$ft)
->: {^}.value = attr_ref_list_start({0}.$text, {0}.$ft)

%symbol {WITH}
->: {^}.value = ident({1}.value)

%symbol {SEM-ACTION}
->: {^}.value = make_semantic_action({1}.$text, {1}.$ft, {3}.$text, {3}.$ft)
->: {^}.value = make_semantic_action({1}.$text, {1}.$ft, {3}.$text, {3}.$ft, {4}.value)

%symbol {SEM-ACTION-LIST}
->: {^}.value = semantic_action_list_append({0}.value, {1}.value)
->: {^}.value = semantic_action_list_start({0}.value)

%symbol {ASYM}
->: {^}.value = get_nonterminal({0}.$text)
->: {^}.value = get_terminal({0}.$text)
->: {^}.value = get_int({0}.$text)
->: {^}.value = ident({0}.$text)

%symbol {ASYM-LIST}
->: {^}.value = string_list_append({0}.value, {1}.value)
->: {^}.value = string_list_start({0}.value)

%symbol {APRODUCTION}
->: {^}.value = ident({0}.value)
->: {^}.value = epsilon_string_list

%symbol {PROD-ADDR}
->: {^}.value = make_prod_specifier_index({1}.$text)
->: {^}.value = make_prod_specifier_literal({0}.value)

%symbol {PROD-SPEC}
->: {^}.value = ident({1}.value)
->: {^}.value = make_prod_specifier_next()

%symbol {PROD-ACTION}
->: {^}.value = make_prod_action({0}.value, {1}.value)

%symbol {SYM-ACTIONS}
->: {^}.value = make_symbol_actions({1}.$text, {1}.$ft, {2}.value)

%symbol {SYM-ACTIONS-LIST}
->: {^}.value = symbol_actions_list_append({0}.value, {1}.value)
->: {^}.value = symbol_actions_list_start({0}.value)

%symbol {ASTATE-SET-LIST}
->: {^}.value = actions_state_block_list_append({0}.value, {1}.value)
->: {^}.value = actions_state_block_list_start({0}.value)

%symbol {TSTATE-SET-LIST}
->: {^}.value = tokens_state_block_list_append({0}.value, {1}.value)
->: {^}.value = tokens_state_block_list_start({0}.value)

%symbol {GSTATE-SET-LIST}
->: {^}.value = grammar_state_block_list_append({0}.value, {1}.value)
->: {^}.value = grammar_state_block_list_start({0}.value)

%symbol {GRULE-LIST}
->: {^}.value = rule_list_append({0}.value, {1}.value)
->: {^}.value = rule_list_start({0}.value)

%symbol {TENTRY-LIST}
->: {^}.value = entry_list_append({0}.value, {1}.value)
->: {^}.value = entry_list_start({0}.value)

%symbol {TENTRY}
->: {^}.value = make_token_entry({0}.value, {1}.value)

%symbol {GRULE}
->: {^}.value = make_rule({0}.$text, {2}.value)

%symbol {ALTERNATIONS}
->: {^}.value = string_list_list_start({0}.value)
->: {^}.value = string_list_list_append({0}.value, {2}.value)

%symbol {GPRODUCTION}
->: {^}.value = ident({0}.value)
->: {^}.value = epsilon_string_list()

%symbol {GSYM-LIST}
->: {^}.value = string_list_append({0}.value, {1}.value)
->: {^}.value = string_list_start({0}.value)

%symbol {PRIORITY}   ->: {^}.value = trim_string({1}.value)
%symbol {HUMAN}      ->: {^}.value = trim_string({1}.value)
%symbol {TOKEN}      ->: {^}.value = trim_string({1}.value)
%symbol {STATESHIFT} ->: {^}.value = trim_string({1}.value)

%symbol {TOPTION}
->: {^}.value = make_discard_option()
->: {^}.value = make_stateshift_option({0}.value)
->: {^}.value = make_token_option({0}.value)
->: {^}.value = make_human_option({0}.value)
->: {^}.value = make_priority_option({0}.value)

%symbol {TOPTION-LIST}
->: {^}.value = token_opt_list_append({0}.value, {1}.value)
->: {^}.value = token_opt_list_start({0}.value)

%symbol {PATTERN}
->: {^}.value = trim_string({0}.value)

%symbol {GSYM}
->: {^}.value = get_nonterminal({0}.$text)
->: {^}.value = get_terminal({0}.$text)

%symbol {STATE-INS}
->: {^}.state = make_state_ins({1}.value, {1}.$ft)

%symbol {ID-EXPR}
->: {^}.value = ident({0}.$text)
->: {^}.value = ident({0}.$text)

%symbol {TEXT}
->: {^}.value = append_strings_trimmed({0}.value, {1}.value)
->: {^}.value = ident({0}.value)
->: {^}.value = ident({0}.value)

%symbol {TEXT-ELEM-LIST}
->: {^}.value = append_strings({0}.value, {1}.value)
->: {^}.value = ident({0}.value)

%symbol {NL-TEXT-ELEM}
->: {^}.value = interpret_escape({0}.$text)
->: {^}.value = ident({0}.$text)

%symbol {TEXT-ELEM}
->: {^}.value = interpret_escape({0}.$text)
->: {^}.value = ident({0}.$text)
```