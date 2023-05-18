Simple Markdown file that contains a minimal set of FISHI instructions not
intended to be executed. It has a variety of comment and hash character usages
which is what it was created to test.

### FISHI

```fishi
%%tokens

\## #                     %token +         %human plus sign '+'
%token ##   %human comment start

#
# above comment should match search regex even though its the only thing on the
# line

# comment here - ignore whitespace
\s+  #all this is garbage:%token *         %human multiplication sign '*'
%discard # rest of line should be fine

%%grammar
{S} = ## {S} | ##
```