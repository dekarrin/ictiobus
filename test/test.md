Simple Markdown file that contains only our test input.

```fishi
%%actions
	
						%symbol
	
	
						{hey}
						%prod  %index 8
	
					%action {thing}.thing %hook thing
						%prod {}
	
					%action {thing}.thing %hook thing
						%prod {test} this {THING}
	
						%action {thing}.thing %hook thing
					%prod {ye} + {A}
	
					%action {thing}.thing %hook thing
	
							%symbol {yo}%prod + {EAT} ext
	
					%action {thing}.thing %hook thing
					%%tokens
					[somefin]
	
					%stateshift   someState
	
			%%tokens
	
			%!%[more]%!%bluggleb*shi{2,4}   %stateshift glub
			%token lovely %human Something for this
	
				%%tokens
	
					glub  %discard
	
	
					[some]{FREEFORM}idk[^bullshit]text\*
					%discard
	
					%!%[more]%!%bluggleb*shi{2,4}   %stateshift glub
				%token lovely %human Something nice
					%priority 1
	
				%state this
	
				[yo] %discard
	
				%%grammar
				%state glub
				{RULE} =   {SOMEBULLSHIT}
	
							%%grammar
							{RULE}=                           {WOAH} | n
							{R2}				= =+  {DAMN} cool | okaythen + 2 | {}
											 | {SOMEFIN ELSE}
	
							%state someState
	
							{ANOTHER}=		{HMM}
	
	
	
	
				%%actions
	
				%symbol {text-element}
				%prod FREEFORM_TEXT
				%action {text-element}.str
				%hook identity  %with FREEFORM_TEXT.$text
	
				%prod ESCSEQ
				%action {text-element}.str
				%hook unescape  %with ESCSEQ.$test
	
	
				%symbol {OTHER}
				%prod EHHH
				%action {OTHER}.str
				%hook identity  %with FREEFORM_TEXT.$text
	
				%prod ESCSEQ
				%action {text-element}$12.str
				%hook unescape  %with ESCSEQ.$test
	
				%state someGoodState
	
				%symbol {text-element}
				%prod FREEFORM_TEXT
				%action {text-element}.str
				%hook identity  %with FREEFORM_TEXT.$text
	
				%prod ESCSEQ
				%action {text-element}.str
				%hook unescape  %with ESCSEQ.$test
	
```
