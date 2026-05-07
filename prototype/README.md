# Protype

This is a prototype of the compiler, written in hacky perl

## Purpose

Given the following program, generate assembly, compile it, and link it into `out`:

  
    let c = 3
    let a = 17
    let b = 4

    print( b );
    newline();

    print(a);
    newline();

    return c;

## Usage

To run the compiler run `make`.

To cleanup the generated binary and object files run `make clean`
