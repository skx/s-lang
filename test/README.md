# Test

This directory contains a series of simple test programs, along with their expected output.

Running the tests is a simple matter of compiling each program in turn, executing it to record the output which was generated, and then comparing that with the expected result.

There is a `Makefile` supplied which will run all tests, when invoked via `make`:

* For each file `foo.in`
  * We compile the program, and execute it whilst capturing output in `foo.out`
  * We compare the generated output (`foo.out`) with the expected output (`foo.expected`)
    * If they differ that's a regression, and a failure.

Adding new test-cases should be intuitive.
