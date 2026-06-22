# s-lang

This repository contains a compiler for a minimal programming language, targeting linux/amd64 systems.  It seems that languages are traditionally named after their creators, or have a single-letter name.  This has both of course - on the basis that it is being developed as a learning exercise I have no expectation that anyone other than myself will ever use it I can do that!

The generated code contains no external dependencies, so when compiled static binaries are produced which do not depend upon libC, etc.   The standard library routines which are not used may be removed by the linker, reducing size, and generated binaries start at around 2k in size.

* Written in Golang for portability.
  * Although the generated code is obviously Linux/AMD64-specific.
* We have a real lexer, and parser.
  * Internally an AST is constructed which is then walked to generate the assembly language representation of the program.
* The compiler may automatically invoke the external `as` and `ld` binaries to compile and link if desired.

In terms of features:

* Single-pass compiler which generates an assembly output for programs.
* Parsing uses a Pratt-style parser with precedence layers:
  * Maths operations: `+`, `-`, `*`, `/`, `%`, and `^`.
  * Comparison operations: `<`, `<=`, `>`, `>=`, for integers, floats, and mixtures of the two.
  * Equality operations `==`, `!=` work for integers, floats, and even strings.
    * You can compare integer with integer, float with float, string with string, and mixed operations where you compare either a float or an integer.
    * You cannot compare a string with anything other than a string.
  * Logical operations: `&&` and `||`.
  * Bit operations `<<`, `>>`, `|` and `&`.
  * Postfix decrement/increment support for variables (`i++;`, or `index--;` for example).
  * Unary operations `-`, `+`, and `!`.
* Support for integers, floats, and strings.
  * String literals are interned.
    * So you can call "`print("Steve");`" 100 times and still see the text "Steve" in the binary only once.
* There is support for string/memory indexing.
  * This is byte-based by default, but you can use a `pragma` to treat memory as arrays of 8, 16, 32, or 64 bit entries.  This is the closest we come to types.
  * Bounds checking is enabled and enforced, and pograms terminated with a readable message (`array index out of bounds: x[2] at line 3`)
* The ability to include inline assembly via `inline { .. }`.
  * `inline` statements are generated as they are encountered.
  * If you want to add new sections then use a `data { ..  }`-block, that is guaranteed to be inserted at the end of the assembly-generation.  So you can add "`.section blah .. ..`" without fear of breaking things.
* Looping is available with either the C-style `for`loop, or the `while` statement, including standard support for `break` and `continue`.
* Conditional support with `if` with `else` branch too.
* We support `switch` statements, albeit only with integer/character literals for the `case` matches.
  * `default` is supported too, of course.
* Some support for catching signals via "magic functions":
  * `sigint()` is called, if defined, when SIGINT is received (i.e. Ctrl-C is pressed).
    * We use this in [examples/life.in](examples/life.in) to clear the screen, and restore the cursor.
  * `sigfpe()` is called, if defined, when SIGFPE is received.
    * This is the floating-point exception generated upon division by zero.
* Cleanup via `at_exit()` which is called, if defined, when the program terminates, either due to signals being caught, or at an ordinary exit.
* User defined functions, with default values.
* File input and output via `fopen`, `fread`, `fwrite` and `fclose`.

Anti-features, or limitations:

* The language is built around numbers (integers&floats), and strings.
  * We have no support for arrays, hashes, or structures.
  * That said you can _fake_ arrays via indexing into characters of strings, or `malloc`'d areas of memory.
    * You can see that done in this [test/jumptable1.in](test/jumptable1.in) where we use it to implement a simple dynamic dispatch routine.
    * [test/jumptable2.in](test/jumptable2.in) is cleaner approach where we directly get/set 64-bit values.
* There are only a few functions in the standard library.
* There is no general purpose support for types "u8 x = 8", "u16 y = 16384", etc.
  * We can allocate memory with `malloc()` and index it with "m[n]" - by default the memory will be treated as arrays of bytes.
    * `pragma` can be used to reinterpret a pointer as an arrays of 8, 16, 32, or 64-byte entries
    * (See [test/jumptable2.in](test/jumptable2.in) for an example of that.)
* We do not support closures, or nested functions.
  * Though at a push you could simulate things, as in [examples/adder.in](examples/adder.in).

That said the code is clean, commented/documented, and contains a fair number of test-cases (both of the internal golang packages, and the compilation and execution of programs).

The intent behind this project was to learn, and increase my knowledge of low-level stuff.  So everything here works, and everything is commented, but this is not a production-grade general-purpose language by any means.


### Alternatives

Because the type-system here is so constrained and baked into every aspect of the compiler, the standard-library, and the interface between the two, I decided to step back.

I created a simple lisp compiler where I could worry less about parsing, syntax, and standard-libraries.  Instead focus on typing for lists, lambdas, & etc:

* https://github.com/skx/slisp

A different kind of learning experience; although I still manage to avoid a strongly-typed language!



## Example Programs

There's a good [SUMMARY.md](SUMMARY.md) of the syntax, and implementation details, which documents all supported features and syntax, as well as a collection of [examples/](examples/) showing _real_ programs.

A couple of highlights from the examples:

* [examples/brainfuck.in](examples/brainfuck.in) - Naive brainfuck interpreter.
  * Contains three hardcoded programs inline:
    * The classic "Hello World" program.
    * A simple "cat", which copies STDIN to STDOUT.
    * The impressive mandelbrot generation program.
  * If executed with the path to a file containing a brainfuck program it will read and execute that instead of any of the inline programs.
  * This interpreter takes **approximately 2 minutes** to render the mandelbrot example program.
* [examples/jit.in](examples/jit.in) - JIT brainfuck Compiler
  * This runs the same programs as the previous version, and also has the ability to load other programs if you specify their path.
  * This example _compiles_ the specified brainfuck programs to x86 assembly, before executing it.
  * This runs the mandelbrot generation example in **approximately three seconds**.
* [examples/life.in](examples/life.in)
  * Conway's Game of Life.
  * Randomly populate 20% of the arena, and evolves it until Ctrl-C is pressed.
* Math examples:
  * [examples/factorial.in](examples/factorial.in) - Calculate factorials 1-20.
  * [examples/fibonacci.in](examples/fibonacci.in) - Calculate fibonacci sequence, using recursion.
  * [examples/fizzbuzz.in](examples/fizzbuzz.in) - Calculate fizzbuzz 0-100.
  * [examples/primes.in](examples/primes.in) - Calculate first 100 prime numbers.
  * [examples/num2hex.in](examples/num2hex.in) - Convert a decimal number to a hex string.
* Misc examples
  * [examples/example.in](examples/example.in) - Misc. examples to demonstrate some of the facilities.
  * [examples/sort.in](examples/sort.in) - Sort arrays of characters, or characters in strings.



## Syntax

The following is a tour of our language, again check [SUMMARY.md](SUMMARY.md) for a concrete list of examples, syntax, and caveats along with implementation notes:

    # Comments are prefixed with "#" and last until the end of the line.
    // You can use C-style comments if you prefer.

    # Set a variable and print it.
    let a = 3;
    print( a );
    printf("A = %d\n", a);

    /* Multi-
       line
       comments
       work
       too
     */

    # Indexing works - and is bound-checked at run-time.
    # LET is not required to declare/update a variable.
    s = "Steve";
    print(s[0],"\n");

    # Updating too - again bounds checking is enforced.
    s[1] = 42;
    print(s,"\n");

    # Printing a newline is common.
    newline();

    # simple loops with "while"
    let x = 10;

    # Looping on a variable is the same as "while ( x > 0 ) .."
    while(x) {
       print("The value in my loop is ", x, "\n");
       printf("Again, using printf: %d\n", x);
       x--;
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
    # (Our print function is variadic.)
    print( "steve", " ", 21);

    # Exit with the given status (7)
    exit(1 + 2 * 3);

Trailing semicolons are mandatory, and brackets around the `if`/`while` tests, because that simplifies the parser!

There is an emacs lisp mode, [s-lang.el](s-lang.el) providing syntax highlighting for our language, although there are no other features beyond that.



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

This performs the same generation as in the `generate` sub-command, but also runs the assembler and linker for you (the linking step is pretty aggressive, we remove unused sections, and strip):

     s-lang compile [-output a.out] examples/example.in

Typically you'd run something like this to generate and execute in one go:

     s-lang compile examples/example.in && ./a.out

(Or use the `execute` sub-command to create a binary and run it in one step.)


### execute

This performs the same generation as in the `compile` sub-command, but also runs the resulting binary for you:

     s-lang execute [-output a.out] examples/example.in


### version

Report the version number of this binary, using the `git`-information that `go` embeds within generated binaries.



## STDLIB

We embed a small number of functions within the generated programs, our so-called "standard library".  These are functions which seemed to be useful enough to include globally, and each function that accepts arguments has type-checking, both at compile-time and run-time.

* `addr(PTR)`
  * Return the address of a pointer returned by `malloc(N)`.
  * Necessary if you write a JIT, but not otherwise.
* `argc()`
  * Return the count of supplied command-line arguments.
* `argv(N)`
  * Return the Nth command-line argument, as a string.
* `call(N|POINTER)`
  * Call the given address.  Expected to be used for jumptables, etc.
* `exit(N)`
  * Terminate execution with the given exit-code.
* `fclose(N)`
  * Close the file handle.
* `filesize(STR)`
  * Return the size of the given file.
* `float2int(F)`
  * Convert the given floating-point number to an integer.
* `fopen(PATH,MODE)`
  * Open a file by path, and return the corresponding file handle.
* `fread(HANDLE)`
  * Read and return the complete contents from the given file handle.
* `free(PTR)`
  * free/unmap pointers returned from `malloc(N)`.  After this accesses to the (now-freed) pointer will trigger our sig_segv handler, and terminate program execution.
* `fwrite(HANDLE, PTR, LEN)`
  * Write the given data to the open file handle.
* `getc()`
  * Read a single character from STDIN, returns 0 on EOF.
* `getenv(STR)`
  * Return the contents of the environmental variable with the given name.
* `int2float(N)`
  * Convert the given integer to a floating point.
* `malloc(N)`
  * Allocate N bytes of memory.
  * Internally we use the `MMAP` syscall, and ensure that memory requested is readable, writable, and executable.
* `memlen(PTR|STR)`
  * Return the length of the given string/pointer-allocation as an integer.
* `newline()`
  * Print a newline to STDOUT.
* `panic(STR)`
  * Print the given message, and exit.
  * Do not invoke `at_exit()`.
* `print(...,...,...)`
  * This function is variadic; it will accept any number of arguments of any type.
  * Print each argument in turn.
* `printf(fmt,...,...,...)`
  * This function is variadic; it will accept any number of arguments of any type.
  * Print each argument according to the specified format string.
* `putc(N)`
  * Print the ASCII character corresponding to the given integer to STDOUT, i.e `putc(42);` will print `*`.
  * Calls to this could be replaced with `printf("%c", x);`
* `rand(N)`
  * Return a random number between 0-(N-1).
* `sleep(N|F)`
  * Sleep for the given duration, integer or float.
* `sqrt(N|F)`
  * Calculate the square root of the given integer/float.
  * Always returns a floating-point result.
* `strcat(STR, STR)`
  * Combine the two strings, and return the new string.
* `strcmp(STR, STR)`
  * Compare two strings for equality, return `0` if equal.
  * You can use `if ( str == str ) { ..` to directly test for equality now, but this was not previously possible.  Similarly `!=` works for inequality testing.
* `strdup(STR)`
  * Allocate a copy of the given string, and return it.
* `strlen(STR)`
  * Return the length of the given string.
* `str2int(STR)`
  * Convert a string into an integer.
* `str2float(STR)`
  * Convert a string into a floating-point number.

You can find the implementation of our standard library routines beneath the [compiler/templates/stdlib](compiler/templates/stdlib) directory.


### Adding to the standard library

If you wish to add a new function which will be available to all compiled programs you need to add it to a new file beneath `compiler/templates/stdlib`, then rebuild the compiler.  Template files beneath `compiler/templates` are embedded within the compiler - there is need to configure load-paths, or similar.

As an example if you wished to define the new function `foo`:

* Create `compiler/templates/stdlib/foo.tmpl`
* Inside there define a new function, with the label `foo:`.
* Rebuild the compiler (with "`go build .`")

Once that is done your prgrams can immmediately call it:

    let a = foo();

This works because internally a call to `foo( [args] )` is converted into a call to the assembly-language function named `foo`.  (i.e. Defined with the label `foo:`).  You can use `inline` to define/call such a function manually if you wish, providing you follow our function ABI.

It is assumed your function will check the types of any arguments it receives, but you can add an entry to the type-checking package, described later, if you wish to add some additional compile-time type checking.



## Type Checking

There are two forms of possible type-checking:

* Type checking at compilation time.
* Type checking at run time.

At compilation time we can detect invalid argument counts for standard library functions _and_ user-defined functions.  For example this is caught:

     function foo( x ) {  print( "I got: ", x , "\n" );
     foo();  # ERROR - Expected one argument, received zero.

For checking actual types at compile time we're limited, we can detect this error:

     strlen(3);  # Wrong type, expected string but got int

However this is permitted:

     let a = 3;
     print(strlen(a))  # Type information didn't survive the assignment

**NOTE**: Compile-time type checking of standard-library functions requires an explicit definition within our [check/](check) package.  If you add a new function please do add an entry there.

Run-time checking of types is deferred to our standard library routines, and they _should_ all check their argument types are valid before they execute their jobs.  They will return an error string "strlen: expected STRING", or similar, instead of their normal result.



## Optimizing Generated Binary Size

When the `compile` sub-command is executed we run generate `NAME.s` from `NAME.in`, and then:

      as -msyntax=intel -mnaked-reg NAME.s -o NAME.o
      ld --gc-sections -s -no-pie -z noseparate-code -o NAME NAME.o

The linker command here shrinks our binaries significantly, if you don't want/need that size-saving then a more typical usage will suffice:

      ld -o NAME NAME.o

The difference in the two approaches can be seen by our brainfuck example:

* Just using `ld -o brainfuck brainfuck.o`:
  * `-rwxr-xr-x 1 skx skx 38648 May 24 13:27 brainfuck`
* With our longer `ld` usage:
  * `-rwxr-xr-x 1 skx skx 18208 May 24 13:28 brainfuck`

We went from 25k to 18k, which is a good saving.

You can see some tips on debugging with `gdb` in [DEBUGGING.md](DEBUGGING.md), if you want to ease your debugging it is recommended you assemble and link yourself, that way there will be debug information available to you.



## Development & Testing

Development is nearing completion now.  There are a few small and obvious things to add, but at the same time the scripting language itself is pretty complete, the standard library is complex enough to write real programs, and I suspect my urge to add new things will diminish over time.

* If any compiled program terminates with a segfault at runtime that's a bug I will definitely fix.

Updates _should_ be contributed by pull-requests which address open issues, but sometimes I'm less strict with myself than I should be.

I've written test-cases covering most of the implementation, which you can run in the standard manner:

```
$ go test ./...
ok      s-lang	0.005s
ok      s-lang/compiler	0.009s
ok      s-lang/lexer	0.006s
ok      s-lang/parser	0.003s
```

There is also support for the fuzz-testing that golang provides, you can run five minutes of fuzz-testing by executing the following (remove the `-fuzztime=300s` to run _forever_, and remove `-parallel=1` to run more than a single instance at a time):

```
go test -fuzztime=300s -parallel=1 -fuzz=FuzzProject -v
```

In addition to the golang tests, and fuzzer, we have some functional test-cases beneath `test/`, there is a trivial driver which executes each of the sample programs, and compares the output produced to known-good results:

```
$ cd test && make
expr.in
 compiling expr.in to expr
 executing expr > expr.out
 comparing expr.out to expr.expected
 cleanup
...
```

Running `make test` should run both of those things.



## Lessons Learned

* Not having (strong) types meant it was easy to get started, but in the long-run it made several things very limited.
  * Storing type-data in the lower bits was a decision which worked, but it becomes pervasive.
* I should have started with a Pratt-style parser, rather than recursive descent.
  * It's much easier to extend and manage precedence levels this way.
* It's surprising how small a standard-library you actually need to write complex, or "useful", programs.
* Implementing a real AST was very useful, and a good decision.
  * I had been tempted to start with a BASIC-style language and skip the use of an AST.
  * You can see hints of that in my use of "`LET x = y`".
* Just because it is assembly doesn't mean it is fast.
  * The mandelbrot example takes two minutes to run, which is 1:58 too slow.
* Using a real ABI would make functions faster and easier to manage.
* Using a flexible register allocation would also have made things faster, but it was easy to just assume values were stored in RAX, etc.
* There's a lot of value to be obtained by writing functional tests, not just implementation tests.
  * These genuinely caught regressions.
  * Especially when I updated parts of the standard-library and forgot where parts call each other.
* Putting functions in separate sections allows the linker to optimize very effectively.
* It wasn't immediately obvious that I didn't need to write code for **both** `==` (equality-checking) and the reverse (`!=`/inequality checking).
  * Instead I needed to only write one routine and invert the result for the other operation.
  * Similarly ">" is the same as "not <=", and "<" is the same as "not >=".
* Adding bounds-checks for arrays was pretty simple and definitely worthwhile.
  * But adding line-number, and details of the array and index involved, made it much more useful.
