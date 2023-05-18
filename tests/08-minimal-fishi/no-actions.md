Simple Markdown file that contains a minimal set of FISHI instructions not
intended to be executed. This one is missing any action definitions.

### FISHI

```fishi
%%tokens

\d    %token int     %human integer

\s+   %discard

%%grammar
{INTSEQ} = int {INTSEQ} | int

```