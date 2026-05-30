# examples

Compile all of the examples by running `make`.

If you'd like to see the generated assembly for each example as well as produce the binaries then run `make ASM=1`.



## Contents

* [brainfuck.in](brainfuck.in) - Brainfuck interpreter.
  * Reads the command-line arguments to decide what to do.
    * Either run one of the embedded examples.
    * Or executes the program in the path you specify.
* [cat.in](cat.in)
  * A simple `cat` program which echos input.
* [example.in](example.in)
  * Misc. examples.
* [factorial.in](factorial.in)
  * Calculate factorials.
* [fibonacci.in](fibonacci.in)
  * Calculate the fibonacci sequence, recursively.
* [fizzbuzz.in](fizzbuzz.in)
  * The standard test.
* [functions.in](functions.in)
  * Demonstrate user-defined functions, and inline-assembly.
* [life.in](life.in)
  * Conway's Game of Life.
  * Randomly populate 20% of the arena, and evolve until bored!
* [num2hex.in](num2hex.in)
  * Convert a (decimal) number to hex, and return that result.
  * e.g. "255" -> "`0xFF`".
* [primes.in](primes.in)
  * Calculate the first 100 prime numbers.
* [string.in](string.in)
  * Demonstrate iterating over strings, setting their contents, etc.
  * Also shows some conversion and comparison results.
* [types.in](types.in)
  * Demonstration of our types.
* [while.in](while.in)
  * Demonstrate using (nested) while-loops.

Trivial ones:

* [empty.in](empty.in) - Literally an empty source file.
* [return.in](return.in) - Return a given status-code.
