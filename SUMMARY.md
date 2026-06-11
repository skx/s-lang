# Language Summary

This is a small scripting language which compiles to Linux/amd64 assembly language.



## Types

There are three runtime types:

* `integer`
* `float`
* `string/pointer`

Strings and pointers share the same underlying type. `malloc(size)` returns a
pointer, as do some of our standard-library routines such as "strdup", and
"strcat".

Integer operations remain integer-based, with the exception of division which
always returns a floating point number.   A float may be truncated to an
integer with the standard library routine "float2int".  A string may be
converted to a number with either "str2int" or "str2float":

```text
f = str2float("3.15");
i = float2int(f);
```



## Variables

Variables are created and assigned like so:

```text
let x = value;
```

If the variable is new it is created, otherwise the existing value in the current scope, or higher scope, is updated.

Examples:

```text
let count = 10;
let count = count + 1;
count++;
```

LET is optional so these are identical:

```text
count = 10;
count = count + 1;
count++;
```



## Functions

Functions are declared with:

```text
function name(arg1, arg2) {
}
```

Functions may have default arguments of basic types (integer, float, or string):

```text
function greet(name = "world") {
    print("Hello ", name, "\n");
}
```

Functions may return a value:

```text
return(value);
```

Or return nothing:

```text
return;
```

Function calls:

```text
greet();
greet("Steve");
```

The return value may be ignored.



## Scoping

New scopes are created by:

* functions
* if blocks
* while blocks
* case blocks inside switch statements

Inner scopes may access variables from outer scopes.



## Control Flow

Conditionals:

```text
if (condition) {
} else {
}
```

Loops:

```text
while (condition) {
}
```

Truthiness is supported:

```text
if (x) {
}

while (x) {
}
```

While loops may have their execution flow changed with:

```text
break;
continue;
```



## Switch Statements

Syntax:

```text
switch (value) {

    case 1 {
        ...
    }

    case 2 {
        ...
    }
    default {
        ...
    }
}
```

But note:

* no fall-through.
* each case body is its own scope.
* explicit break statements are not required.



## Expressions

Arithmetic:

```text
+
-
*
/
%  (modulus)
^  (power)
```

Comparison:

```text
==
!=
<
<=
>
>=
```

Logical:

```text
&&
||
```

Unary prefix functions:

```text
!
-
+  (nop)
```

Increment/decrement postfix expressions for variables:

```text
x++;
x--;
```



## Booleans

The two words `false` and `true` are recognized, but they are converted into integers as part of the lexing process.

```text
print( true );   # prints 1
print( false );  # prints 0
```



## Characters

Character literals use single quotes:

```text
'A'
'+'
'\n'
```

A character literal is converted to an integer byte value (0-255) as the input program is lexed, and before it is compiled.



## Strings

String literals use double quotes:

```text
"hello"
"world\n"
```



## Arrays and Memory

Pointers returned from `malloc()` may be indexed:

```text
let ram = malloc(4096);

ram[0] = 65;
print(ram[0]);
```

Indexing syntax:

```text
ptr[index]
ptr[index] = value
```

Pragmas may change element size:

```text
pragma table size16
```

After this, indexed accesses use 16-bit elements instead of bytes.



## Inline Assembly

Functions may contain inline assembly:

```text
function fast() {
    inline {
        mov rax, 42
        sar rax, 2
        ret
    }
}
```



## Comments

Comments begin with `#` and continue to end of line.



## Standard Library Functions

Here is a brief list of standard library functions, if the name matches a C-language function assume it operates in a similar way.

* `argc()`
  * Return the count of supplied command-line arguments, as an integer.
* `argv(N)`
  * Return the Nth command-line argument, as a string.
* `exit(N)`
  * Terminate execution with the given exit-code.
* `filesize(STR)`
  * Return the size of the given file as an integer.
* `float2int(F)`
  * Convert the given floating-point number to an integer.
* `getc()`
  * Read a single character from STDIN, returns 0 on EOF, otherwise an integer in the range 0-255.
* `getenv(STR)`
  * Return the contents of the environmental variable with the given name.
* `int2float(N)`
  * Convert the given integer to a floating point.
* `malloc(N)`
  * Allocate N bytes on the heap.
  * **NOTE**: We have no corresponding `free`.
* `memlen(PTR|STR)`
  * Return the length of the given string/pointer-allocation as an integer.
* `newline`
  * Print a newline to STDOUT.
* `panic(STR)`
  * Print the given message, and exit.
* `print(...,...,...)`
  * This function is variadic, it will accept any number of arguments of any type, print each argument in turn.
* `putc(N)`
  * Print the ASCII character corresponding to the given integer to STDOUT, i.e `putc(42);` will print `*`.
* `rand(N)`
  * Return a random number between 0-(N-1).
* `readfile(STR)`
  * Return the contents of the given file as a string.
* `sleep(N|F)`
  * Sleep for the given duration, integer or float.
* `sqrt(N|F)`
  * Calculate the square root of the given integer/float.
  * Always returns a floating-point result.
* `strcat(STR, STR)`
  * Concatenate the two strings, and return the new string result.
* `strcmp(STR, STR)`
  * Compare two strings for equality, return `0` if equal.
* `strdup(STR)`
  * Allocate a copy of the given string, and return it.
* `strlen(STR)`
  * Return the length of the given string as an integer.
* `str2int(STR)`
  * Convert a string into an integer.
* `str2float(STR)`
  * Convert a string into a floating-point number.



## Style Guidance

Prefer C-style formatting:

```text
function max(a, b) {

    if (b > a) {
        return(b);
    }

    return(a);
}
```

Braces are necessary for all control-flow blocks, it is an error to try to omit them as is possible in C for single-statement if-blocks.
