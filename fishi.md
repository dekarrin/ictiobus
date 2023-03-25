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

```

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

\n\s*%!%!.                                         %token line-start-escseq
%human escape sequence on next line

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

[^\S\n]+        %discard




```

For grammar state:

```fishi
```

For actions state:
```fishi

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
```
