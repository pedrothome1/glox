# Book challenges

## Chapter 1

1. There are at least six domain-specific languages used in the [little system 
I cobbled together](https://github.com/munificent/craftinginterpreters) to write and publish this book. 
What are they? *They are: HTML, CSS, SCSS, Markdown, Makefile and Shell Script.*

2. *Skipped*

3. *Skipped*


## Chapter 2

1. Pick an open source implementation of a language you like. 
Download the source code and poke around in it. 
Try to find the code that implements the scanner and parser.
Are they handwritten, or generated using tools like Lex and Yacc? 
(.l or .y files usually imply the latter.)
*I picked Python. 
[Scanner](https://github.com/python/cpython/blob/649768fb6781ba810df44017fee1975a11d65e2f/Parser/tokenizer.c) 
and [Parser](https://github.com/python/cpython/blob/649768fb6781ba810df44017fee1975a11d65e2f/Parser/parser.c).*

2. Just-in-time compilation tends to be the fastest way 
to implement dynamically typed languages, but not all of them use it. 
What reasons are there to not JIT? *Between other things, according to ChatGPT:
Slower startup time, security concerns, portability, warm-up time. Myself, I
think complexity of implementation is an important consideration as well.*

3. Most Lisp implementations that compile to C also contain 
an interpreter that lets them execute Lisp code on the fly as well. Why?
*I guess because, having the translation to C, it's straightforward to add
a REPL as well to ease testing and toying around with the language.*


## Chapter 3

1. *Skipped*

2. *Skipped*

3. Lox is a pretty tiny language. What features do you think 
it is missing that would make it annoying to use for real
programs? (Aside from the standard library, of course.)
*Modularization features. Importing/Exporting declarations.*

## Chapter 4

1. *Skipped*

2. *Skipped*

3. Our scanner here, like most, discards comments and whitespace 
since those aren’t needed by the parser. 
Why might you want to write a scanner that does not discard those? 
What would it be useful for?
*It would be useful for writing a formatter tool, use comments as code metadata
via reflection, and maybe other things.*

4. Add support to Lox’s scanner for C-style `/* ... */` block comments. 
Make sure to handle newlines in them. Consider allowing them to nest. 
Is adding support for nesting more work than you expected? Why?
*The testing code is located at `challenges/chapter4_4/`. Adding support for
nesting brings the necessity to ensure the balance between `/*` and `*/`.
I've used a stack for that.*















