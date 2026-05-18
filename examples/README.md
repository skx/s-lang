# examples

Compile all of the examples by running `make`.

If you'd like to see the generated assembly for each example as well as produce the binaries then run `make ASM=1`.



## Contents

* [brainfuck.in](brainfuck.in) - Brainfuck interpreter.
  * Runs the classic "Hello World" example program.
  * Fails to run the mandelbrot example, but I've not looked into why yet.
* [cat.in](cat.in) - Simple `cat` which echos input.
* [example.in](example.in)
  * Misc. examples.
* [factorial.in](factorial.in)
  * Calculate factories.
* [fibonacci.in](fibonacci.in)
  * Calculate the fibonacci sequence, recursively.
* [fizzbuzz.in](fizzbuzz.in)
  * The standard test.
* [functions.in](functions.in)
  * Demonstrate user-defined functions, and inline-assembly.
* [string.in](string.in)
  * Demonstrate iterating over strings, setting their contents, etc.
* [string-dump.in](string-dump.in)
  * Processing each character of a string, using the index-operator.
* [types.in](types.in)
  * Demonstration of our types
* [while.in](while.in)
  * Demonstrate using (nested) while-loops.

Trivial ones:

* [empty.in](empty.in) - Literally an empty source file.
* [return.in](return.in) - Return a given status-code.
