# s

This is my own take on a minimal linux x86 compiler:

* Simple compiler written in Golang.
* Generates an AST internally.
* Uses that to generate assembly language.
* Invokes the external `as` and `ld` binaries to compile and link if desired.

I was inspired by a simple compiler I saw recently:

* https://github.com/ismail0098-lang/Y-/tree/main



## Example Programs

See [examples/](examples/) for "real" programs.  A couple of highlights:

* [examples/factorial.in](examples/factorial.in) - Calculate factorials 1-20
* [examples/fizzbuzz.in](examples/fizzbuzz.in) - Calculate fizzbuzz 0-100

Syntax below, and some sample code here:

* [example.in](example.in) - Misc. Examples.
* [if.in](if.in) - Examples of comparisons.



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

    # Only looping on variable values.
    while(x) {
       print("The value in my loop is ", x, "\n");
       let x = x - 1;
    }

    inline {
       # Inline assembly here
       mov rax, 32
       call print_number
    }

    # Simple comparisons are present
    if (x >= 3 ) {
      print("x >= 3\n");
    }

    # Printing of integer and string literal works too.
    print( "steve", " ", 21);

    # Exit with the given status
    return(1 + 2 * 3);

> **NOTE**: Example files include [example.in](example.in) and [if.in](if.in).

Trailing semicolons are mandatory (because that simplifies the parser. Sorry!)



## Grammar

```
program         ::= statements

statements      ::= { statement }

statement       ::= ";"
                  | "function" IDENT block
                  | "let" IDENT "=" expression
                  | "if" "(" expression ")" block
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
                  | FUNCTION()
                  | "(" expression ")"
```



## Usage

Build, and optionally install, then use the `s-lang` binary with our [example.in](example.in) file:


### lex

This is an internal command to show what the lexer makes of a given input file:

     s-lang lex example.in


### parse

This is an internal command to show what the parser makes of a given input file:

     s-lang parse example.in


### generate

This is one of the main commands, and generates an assembly language version of the input file:

     s-lang generate example.in [-o out.s]


### compile

This performs the same generation as in the `generate` sub-command, but also runs the assembler and linker for you:

     s-lang compile example.in [-o a.out]

Typically you'd run something like:

     s-lang compile example.in && ./a.out



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

* [x] negative numbers (implemented in #14).
* [x] allow assignment of strings to variables (implemented in #16).
* [ ] add types to our variables
* [ ] floating point numbers
* [ ] allow *x to get the address of x, for working with strings
* [x] user-defined functions (e.g. min/max/abs/etc.)
* [ ] arguments to user-defined functions.
