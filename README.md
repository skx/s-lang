# s-lang

This repository contains a compiler for a minimal programming language, targeting linux/amd64 systems.

The generated code contains no external dependencies, so when compiled they are static binaries and do not depend upon libC, etc.   The standard library routines which are not used may be removed by the linker, reducing size, and generated binaries start around 4k.

* Written in Golang for portability, although the generated code is obviously Linux/AMD64-specific.
* We have a real lexer, and parser, and internally generate an AST.
  * The AST is then walked to generate the assembly language representation of the program.
* We can automatically invoke the external `as` and `ld` binaries to compile and link if desired.

In terms of features:

* Single-pass compiler, which generates an assembly output for programs.
* Parsing using recursive descent with precedence layers:
  * Maths operations: `+`, `-`, `*`, `/`
  * Comparison operations: `<`, `<=`, `==`, `!=`, `>`, `>=`,
  * Logical operations: `&&` and `||`.
* Support for integers, floats, and strings.
  * Float and string literals are interned.
  * So you can call "print("Steve")" 100 times and still see the text "Steve" in the binary only once.
* The ability to include inline assembly via `inline { .. }`.
  * `inline` statements are generated inline with the current code-position.
  * If you want to add new sections then use a `data { ..  }`-block, that is guaranteed to be inserted at the end of the file.  So you can add "`.section blah .. ..`" without fear of breaking things.
* Loops via `while` (with support for `break` and `continue`).
* Conditional support with `if` with `else` branch too.

Anti-features, or limitations:

* The language is built around integers, and strings.
  * Float support is rudimentary and mostly limited to setting values, constant expressions and printing their values.
  * There are only a few functions in the standard library.

That said the code is clean, readable, and it could be updated to work with floating-point reasonably easily.



## Example Programs

See [examples/](examples/) for "real" programs.  A couple of highlights:

* [examples/brainfuck.in](examples/brainfuck.in) - Brainfuck interpreter.
  * Runs the classic "Hello World" example program.
* [examples/factorial.in](examples/factorial.in) - Calculate factorials 1-20.
* [examples/fibonacci.in](examples/fibonacci.in) - Calculate fibonacci sequence, using recursion.
* [examples/fizzbuzz.in](examples/fizzbuzz.in) - Calculate fizzbuzz 0-100.
* [examples/functions.in](examples/functions.in) - Demonstrates user-defined functions.
* [examples/types.in](examples/types.in) - Demonstrates getting variable types.

Syntax is covered pretty well in our "misc example" file:

* [examples/example.in](examples/example.in) - Misc. Examples.



## Syntax

The following is a tour of our language:

    # Comments are prefixed with "#".

    # Set a variable and print it.
    let a = 3;
    print( a );

    # Printing a newline is common.
    newline();

    # simple loops with "while"
    let x = 10;

    # Looping on a variable is the same as "while ( x > 0 ) .."
    while(x) {
       print("The value in my loop is ", x, "\n");
       let x = x - 1;
    }

    inline {
       # Inline assembly here
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

Here we see the available sub-commands that you might choose to use, though in
practice only the last three are useful for users.


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

Confirm the assembly is sane:

     objdump -M intel -S out

Then run it:

    ./out

The `compile` sub-command automates the process of generating source, compiling it, and linking it to produce a final binary.


### compile

This performs the same generation as in the `generate` sub-command, but also runs the assembler and linker for you:

     s-lang compile [-output a.out] examples/example.in

Typically you'd run something like this to generate and execute in one go:

     s-lang compile examples/example.in && ./a.out

(Or use the `execute` sub-command to create a binary and run it in one step.)


### execute

This performs the same generation as in the `compile` sub-command, but also runs the resulting binary for you:

     s-lang execute [-output a.out] examples/example.in



## STDLIB

_Standard library_ is a grandiose term for the simple library routines we embed, but we have implemented several functions:

* `exit`
  * Assumes the value in the RAX register is the desired exit-code and terminates execution with that value.
* `getc`
  * Get a single byte from STDIN, `let in = getc();`.
* `newline`
  * Prints a newline.
* `print`
  * Determine the type of the given variable, and print it appropriately.
* `putc`
  * Print the ASCII character corresponding to the given integer, i.e `putc(42);` will print `*`.
* `strcmp`
  * Compare two strings for equality, return `0` if equal.
* `strlen`
  * Return the length of the given string.

You can see our standard library routines beneath the [compiler/templates/stdlib](compiler/templates/stdlib) directory.

You don't need to do anything special to add new standard library functions, if you were to add a new standard-library function beneath `compiler/templates/stdlib` it would become immediately available for calling:

* Define a new template function "`foo: .. ret;`", within `compiler/templates/stdlib/foo.tmpl`.
* Your code can immediately call it `let a = foo(17);`

This is because internally a call to `foo( [args] )` is converted into a call to the assembly-language function named `foo`.  (i.e. Defined with the label `foo:`).  You can use `inline` to define/call such a function manually if you wish, providing you follow our ABI.



## Types

We used heap-allocated boxed pointers for floats, and ints, and static string pointers for strng values.

To identify what kind of pointer we have we use the lower two bits:

* decimal 00 binary `00` -> integer
* decimal 01 binary `01` -> pointer/string
* decimal 02 binary `10` -> float
* decimal 03 binary `11` -> reserved

TLDR; We allocate memory for integers and floats, strings are just pointers to static defitions within the `.data` section, the bottom two bits of the pointers identify the type.



## ABI

We define a simple ABI for function invocation:

* All function parameters are passed on the stack.
* The _number_ of parameters is passed in the RAX register.

You can see how this is handled in [our standard-library functions](compiler/templates/stdlib) for reference, but do remember that variables have types.   As an example if you want to print the integer 17  using inline assembly you would run:

     inline {
        call alloc8              # allocate 8-byte boxed integer
        mov qword ptr [rax], 17  # store payload
        or rax, 0                # the pointer in RAX is an integer

        push rax     # parameters are passed on the stack
        mov rax, 1   # one argument is being passed

        call print   # call the stdlib function
     }



## Testing / Development

Testing is largely done interactively, but there are golang tests for all the internal packages and code, with pretty high/good coverage:

```
$ cover ./...
ok      s-lang	0.004s	coverage: 75.2% of statements
ok      s-lang/compiler	(cached)	coverage: 75.5% of statements
ok      s-lang/lexer	(cached)	coverage: 93.9% of statements
ok      s-lang/parser	(cached)	coverage: 93.0% of statements
```

Run the tests as you usually would:

```
$ go test ./...
ok      s-lang	0.005s
ok      s-lang/compiler	0.009s
ok      s-lang/lexer	0.006s
ok      s-lang/parser	0.003s
```

For _real_ testing compile all the examples and run them:

```
$ cd examples && make
$ ./factorial
$ ./fibonacci
$ ./functions
$ ./while
..
```



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
* [x] Read [the `as` manual](https://www.gnu.org/software/binutils/) to see if there is support for dead-code elimination.
  * There is support for removing unused sections inside the `ld` linker.
  * See [#39](https://github.com/skx/s-lang/issues/39) for details.
* [x] add types to our variables
  * Implemented in [#31](https://github.com/skx/s-lang/pull/31)
* [x] string comparison should work
* [x] floating point numbers
* [ ] allow *x to get the address of x, for working with strings
