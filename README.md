# s-lang

This repository contains a minimal linux x86 compiler, which generates assembly language for `amd64`.

The generated code contains no external dependencies, so they are static and do not depend upon libC, etc.

With our bundled "runtime functions" the generated binaries start at approximately 8k.

* Written in Golang for portability, although the generated code is AMD64-specific.
* We have a real lexer, and parser, and internally generate an AST.
  * The AST is walked to generate an assembly representation of the program.
* We can automatically invokes the external `as` and `ld` binaries to compile and link if desired.

I was inspired by a simple compiler I saw recently:

* https://github.com/ismail0098-lang/Y-/tree/main

In terms of features:

* Single-pass compiler, which generates an assembly output for programs.
* Parsing using recursive descent with precedence layers:
  * Maths operations: `+`, `-`, `*`, `/`
  * Comparison operations: `<`, `<=`, `==`, `!=`, `>`, `>=`,
  * Logical operations: `&&` and `||`.
* Strings are interned:
  * So you can call "print("Steve")" 100 times and still see the text "Steve" in the binary only once.
* The ability to include inline assembly via `inline { .. }`.
* Loops via `while` (with support for `break` and `continue`).
* Conditional support with `if` with `else` branch too.

Anti-features, or limitations:

* The language is built around integers, with additional support for printing strings.
* There are no floating-point operations, no typed operations.

That said the code is clean, readable, and it could be updated to work with floating-point reasonably easily.



## Example Programs

See [examples/](examples/) for "real" programs.  A couple of highlights:

* [examples/factorial.in](examples/factorial.in) - Calculate factorials 1-20.
* [examples/fibonacci.in](examples/fibonacci.in) - Calculate fibonacci sequence, using recursion.
* [examples/fizzbuzz.in](examples/fizzbuzz.in) - Calculate fizzbuzz 0-100.
* [examples/functions.in](examples/functions.in) - Demonstrates user-defined functions.

Syntax is covered pretty well in our "misc example" file:

* [examples/example.in](examples/example.in) - Misc. Examples.



## Syntax

The following is a tour of our language:

    # Comments are prefixed with "#".

    # Set a variable and print it.
    let a = 3;
    print( a );

    # A newline will be added if you use "println"
    println(a);

    # simple loops with "while"
    let x = 10;

    # Looping on a variable is the same as "while ( x > 0 ) .."
    while(x) {
       print("The value in my loop is ", x, "\n");
       let x = x - 1;
    }

    inline {
       # Inline assembly here
       mov rax, 32
       call print_number
    }

    # Conditional expressions are present
    if (x >= 3) {
      print("x >= 3\n");
    } else {
      print("x is not >= 3\n");
    }

    # Printing of integer and string literal works too.
    print( "steve", " ", 21);

    # Exit with the given status
    return(1 + 2 * 3);

Trailing semicolons are mandatory (because that simplifies the parser. Sorry!)



## Grammar

```
program         ::= statements

statements      ::= { statement }

statement       ::= ";"
                  | "function" IDENT block
                  | "let" IDENT "=" expression
                  | "if" "(" expression ")" block [ else block ]
                  | "inline" "{" LITERAL "}"
                  | "while" "(" expression ")" block
                  | "print" "(" exprList ")"
                  | "println" "(" exprList ")"
                  | "return" "(" expression ")"

block           ::= "{" statements "}"

exprList        ::= expression { "," expression }

expression      ::= logicalOr

logicalOr       ::= logicalAnd
                    { "||" logicalAnd }

logicalAnd      ::= equality
                    { "&&" equality }

equality        ::= comparison
                    { ( "==" | "!=" ) comparison }

comparison      ::= addSub
                    { ( "<" | "<=" | ">" | ">=" ) addSub }

addSub          ::= mulDiv
                    { ( "+" | "-" ) mulDiv }

mulDiv          ::= primary
                    { ( "*" | "/" ) primary }

primary         ::= NUMBER
                  | STRING
                  | IDENT
                  | FUNCTION( exprList )
                  | "(" expression ")"
```



## Usage

Once built (and optionally installed) the `s-lang` binary may be used to
generate, compile, or inspect the output of various stages via a number of
sub-commands.

Here we see the four sub-commands that you might choose to use, though in
practice only the last two are expected to be used regularly.


### lex

This is an internal command to show what the lexer makes of a given input file:

     s-lang lex examples/example.in


### parse

This is an internal command to show what the parser makes of a given input file:

     s-lang parse examples/example.in


### generate

This is one of the main commands, and generates an assembly language version of the input file:

     s-lang generate [-output out.s] examples/example.in

You could assemble that output, and link it, like so:

     as -msyntax=intel -mnaked-reg out.s -o out.o
     ld -s -o out out.o

Then run:

    ./out

Though the `compile` sub-command does that for you.


### compile

This performs the same generation as in the `generate` sub-command, but also runs the assembler and linker for you:

     s-lang compile [-output a.out] examples/example.in

Typically you'd run something like this to generate and execute in one go:

     s-lang compile examples/example.in && ./a.out



## STDLIB

_Standard library_ is a grandiose term for the simple library routines we embed, but we have a couple of reusable functions within the generated assembly:

* newline
  * Prints a newline.
  * Invoked if you call `println`, which terminates output with a newline.  `print` trusts you to add `\n` if you want a newline.
* print_number
  * Assumes the value in the RAX register is a decimal integer, and prints it.
* print_string
  * Assumes RSI holds the address of the string, and RDX holds the length.
* exit_with_status
  * Assumes the value in the RAS register is the desired exit-code and terminates execution with that value.



## History

There is a simple perl-based prototype, beneath [prototype/](prototype/), which I hacked up to see if this would be a project that was within my means.

It parses via regexp which is terrible, but also good enough to show that things could work in a predictable fashion.



## Future Additions

Possible future improvements and additions, to be added slowly if ever.

* [x] negative numbers may be parsed and print'd
  * Implemented in [#14](https://github.com/skx/s-lang/pull/14)
* [x] allow assignment of strings to variables.
  * Implemented in [#16](https://github.com/skx/s-lang/pull/16)
* [x] user-defined functions (e.g. min/max/abs/etc.)
  * Implemented in [#18](https://github.com/skx/s-lang/pull/18)
* [x] user-defined functions can `return` values.
  * Implemented in [#19](https://github.com/skx/s-lang/pull/19)
* [x] user-defined functions can access (local) variables.
  * Implemented in [#20](https://github.com/skx/s-lang/pull/20)
* [x] arguments to user-defined functions.
  * Implemented in [#20](https://github.com/skx/s-lang/pull/20)
* [x] Implement `else` support for our `if` statements.
  * Implemented in [#20](https://github.com/skx/s-lang/pull/20)
* [x] Implement `break` and `continue` within a `while` statement.
  * Implemented in [#30](https://github.com/skx/s-lang/pull/30)
* [x] Constant folding - probably in a new pass after the parser.
  * Implemented in [#28](https://github.com/skx/s-lang/pull/28)
* [ ] Read `as` manual to see if there is support for dead-code elimination.
  * https://www.gnu.org/software/binutils/
* [ ] add types to our variables
* [ ] floating point numbers
* [ ] allow *x to get the address of x, for working with strings
