# Test

This directory contains a series of smple test programs, which allow compling fixed programs and ensuring that their output matches expectations.

There is a `Makefile` supplied which will run all tests, when invoked via `make`:

* For each file `foo.in`
  * We compile the program, and execute it whilst capturing output in `foo.out`
  * We compare the generated output (`foo.out`) with the expected output (`foo.expected`)
    * If they differ that's a regression, and a failure.

Adding new test-cases should be intuitive.
