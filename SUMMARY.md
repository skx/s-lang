# Language Summary

`s-lang` is a small scripting language which compiles to Linux/amd64 assembly language.



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

There is no dynamic typing so this program is invalid:

```text
a = "17.4";
b = 3.5
print( a + b );
```



## Type Encoding

As noted we support three different variable types (integer, float, and string/pointer).  We use the lower two bits of values to store their types:

* integers have their lower two bits set to `00`
* pointers have their lower two bits set to `01`.
* floats are allocated on the heap, and the pointer has the lower two bits set to `10`.

There is space left for one more type, if the lower two bits are `11` which might be used in the future.


### Type Encoding Examples

You should be able to work this out from the "Type Encoding" section above, but here are examples of getting values for each of our types:

Getting an integer from RAX:

        sar rax, 2  # Shift right, removing lower two bits.

Getting a float from RAX into XMM0 where it can be operated upon:

        and rax, -4       # Clear the type bits
        movsd xmm0, [rax] # Load the heap-allocated float.

Getting a string/pointer from RAX:

        mov rdi, rax  # Get the string
        and rdi, -4   # Remove typing bits

Returning a number from a function:

        mov rax, 42   # Load the value
        sal rax, 2    # Shift left, so the lower two bits are now 00

Returning a float from a function:

        call alloc8         # allocate 8-byte boxed float
        movsd [rax], xmm0   # store XMM0 in that new pointer.
        or rax, 2           # tag pointer as float (10)

Returning a (static) string from a function:

        mov rax, offset str_ptr  # Load the string
        or rax, 1                # Mark the type

Finally here's how to do type-checking of the parameter in RAX:

        mov rcx, rax
        and rcx, 3

        cmp rcx, 0
        je print_integer

        cmp rcx, 1
        je print_string

        cmp rcx, 2
        je print_float

        cmp rcx, 3
        je print_reserved
        ret

**NOTE**  Our `alloc8` and `malloc` functions will do their own error-checking, if allocation fails they will print a message and terminate execution.  That means there is no need to check the result of calls to allocation routines.



## Variables

Variables are created and assigned like so:

```text
let x = value;
```

If the variable is new it is created, otherwise the existing value in the current scope, or higher scope, is updated.  Examples:

```text
let count = 10;
let count = count + 1;
count++;
```

LET is optional so these work in the same way:

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

The return value may be ignored, and it is valid to have a function body end without a "return" statement.



## Function Call ABI

We define a simple ABI for function invocation:

* All function parameters are passed on the stack.
* The _number_ of parameters is passed in the RAX register.

Do remember that variables have types!  So for example if you wanted to print the integer 17 using inline assembly you would run:

```text
mov rax, 17  # store value in register.
sal rax, 2   # Shift to ensure the bottom two bits are "00".
push rax     # parameters are passed on the stack

mov rax, 1   # one argument is being passed
call print   # call the stdlib function


xor rax, rax   # Newline function takes no arguments
call newline   # Call it
```



## Scoping

New scopes are created by:

* functions
* "if" blocks
* "while" blocks
*" case" blocks inside switch statements

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

Truthiness is supported integers and floats are always true unless their value is zero:

```text
if (x) {
}

while (x) {
}
```

"while" loops may have their execution flow changed with:

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
<
<=
>
>=
```

Equality:

The equality and inequality operators are unlike the other comparison operators, as they may be applied to strings too.  You may compare integers with integers, floats with floats, or strings with strings.

Mixed-type equality/inequality tests are limited to integer/floats.  (So it is valid to compare `1.5 == 1`, or `3 != 1.5`.  But it is illegal to compare `"Steve" == 1.2`.)

```text
==
!=
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

As we have no dynamic typing the comparisons, prefix, suffix, and arithmetic operations may generally only be applied to floats and integers.



## Booleans

The two words `false` and `true` are recognized, but they are converted into integers as part of the lexing process.

```text
print( true );   # prints 1
print( false );  # prints 0
```

So we do not have true support for booleans.



## Characters

Character literals use single quotes:

```text
'A'
'+'
'\n'
```

A character literal is converted to an integer byte value (0-255) as the input program is lexed, and before it is compiled.   Escape codes are recognized, as they are in strings, only for the following cases:

* `\\` - A single slash
* `\n` - A newline
* `\r` - A linefeed
* `\t` - A tab



## Strings

String literals use double quotes:

```text
"hello"
"world\n"
```

The escape characters recognized within character literals are supported, and additionally `\"` allows including an inline quote:

```text
let str = "My name is \"Bob\".";
```



## Arrays and Memory

Pointers returned from `malloc()` or `mmap()` may be indexed:

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

After this, indexed accesses use 16-bit elements instead of bytes.  Valid sizes are:

* size8 - Byte access
* size16 - Word access
* size32 - Double-word access
* size64 - Quad-word access.
* Other values are illegal.



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

It is also possible to allocate executable memory with `mmap`, insert code into it, and call it:

```text
// allocate
let a = mmap(10);

// populate with assembly
//   mov rax, 32
//   ret
a[0] = 72;
a[1] = 199;
a[2] = 192
a[3] = 32
a[4] = 0
a[5] = 0
a[6] = 0
a[7] = 195

// call it
print(call(a), "\n");
```

Similarly you may take the address of a function, and call it:

```text
function foo() {
  print("Called!\n");
}


// Takes the address of the function, and calls it.
call(foo);
```

It is possible to get the _address_ of a pointer via `addr(ptr);` but there are few times when that would be necessary.



## Comments

We have two forms of comments:

* Single-line comments are prefixed with either `#` or `//`.
* Multi-line comments are sandwiched between `/*` and `*/`.

Here you can see both types demonstrated:

```text
/*
 * This function does stuff.
 */
function life() {
   return 42;     // The only answer.
}
```



## Standard Library Functions

Here is a brief list of standard library functions, if the name matches a C-language function assume it operates in a similar way.

* `addr(PTR)`
  * Return the address of a pointer returned by `mmap(N)` or `malloc(N)`.
  * Necessary if you write a JIT, but not otherwise.
* `argc()`
  * Return the count of supplied command-line arguments, as an integer.
* `argv(N)`
  * Return the Nth command-line argument, as a string.
* `call(N|POINTER)`
  * Call the given address.  Expected to be used for jumptables, etc.
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
  * Note memory is not executable; use mmap() if you need that.
  * **NOTE**: We have no corresponding `free`.
* `memlen(PTR|STR)`
  * Return the length of the given string/pointer-allocation as an integer.
* `mmap(N)`
  * Allocate N bytes of readable/writable/executable memory via MMAP.
* `newline()`
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
  * You can use `if ( str == str ) { ..` to directly test for equality now, but this was not previously possible.  Similarly `!=` works for inequality testing.
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
