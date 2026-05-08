# v

I was inspired by a simple compiler I saw recently:

* https://github.com/ismail0098-lang/Y-/tree/main

This is my own take on such a minimalist system:

* Simple compiler written in Golang.
* Generates an AST internally.
* Uses that to generate assembly language.
  * This includes a couple of utility functions.
* Invokes the external `as` and `ld` binaries to compile and link.



## Example Programs

See [examples/](examples/) for "real" programs:

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

* Here you can guess there are 26 variables ("a"-"z"), which are assigned to via `let`.
  * e.g. `let a = 1 + 2 * 3`.
* You can see printing in three forms:
  * `print(n)` shows the contents of register `n`.
  * `print(31)` prints the integer literal `31`.
  * `print("Steve")` prints the string literal `Steve`.
  * You can use `println` instead to add a trailing newline.
  * Add multiple comma-separated arguments to print multiple expressions.
* The exit-code of the generated binary is set via `return(x);`.
  * Where `x` is a variable name, an integer literal, or the value of an expression.
* We have support for both `while` and `if`.
  * Both of these allow simple tests to be made such as `>=`, `<`, `a`, etc.

Using the ability to decrease a variable (`let i = i - 1`) we can also write a loop:

    let x = 3;

    while ( x ) {
        println( x );
        let x = x - 1;
    }

The same comparison support is present for our `if` statements:

    if (a == 3 ) { ... }
    if (a != b ) { ... }
    if (a <= b ) { ... }
    if (a < b ) { ... }
    if (a > b ) { ... }
    if (a >= b ) { ... }

We have support for logical and (`&&`) and or (`||`) too.



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
