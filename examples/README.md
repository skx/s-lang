# examples

Compile all of the examples by running `make`.

If you'd like to see the generated assembly for each example as well as produce the binaries then run `make ASM=1`.



## Contents

* [brainfuck.in](brainfuck.in) - Brainfuck interpreter.
  * Reads the command-line arguments to decide what to do.
    * Either run one of the embedded examples.
    * Or executes the program in the path you specify.
  * This is an interpreter, there's also `jit.in` listed lower down which is a JIT compiler.
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
* [jit.in](jit.in) - JIT compiler for Brainfuck
  * Executes the same programs as the `brainfuck.in` example, but significantly faster.
* [life.in](life.in)
  * Conway's Game of Life.
  * Randomly populate 20% of the arena, and evolve until bored!
* [num2hex.in](num2hex.in)
  * Convert a (decimal) number to hex, and return that result.
  * e.g. "255" -> "`0xFF`".
* [pf.in](pf.in)
  * Calculate prime factors of the given number.
* [primes.in](primes.in)
  * Calculate the first 100 prime numbers.
* [sort.in](sort.in)
  * Sorting the values of an array.
  * And characters in a string.
* [string.in](string.in)
  * Demonstrate iterating over strings, setting their contents, etc.
  * Also shows some conversion and comparison results.
* [types.in](types.in)
  * Demonstration of our types.
* [while.in](while.in)
  * Demonstrate using (nested) while-loops.

Trivial ones:

* [empty.in](empty.in) - Literally an empty source file.
* [ext.in](exit.in) - Exit with the given status-code.
