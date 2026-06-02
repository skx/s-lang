package main

import (
	"github.com/skx/s-lang/compiler"
	"strings"
	"testing"
)

// FuzzProject runs the fuzz-testing against our parser and compiler.
//
// We mostly catch errors with the lexer and parser here, the compiler itself
// will generate and return text for the AST we produce.
func FuzzProject(f *testing.F) {

	// Known errors we might see
	known := []string{
		"bare literal is illegal",
		"unhandled token in",                    // Input is just A
		"unknown token type in parseStatements", // blah
		"unexpected token in parsePrimary",      // foo(123
		"unterminated character literal",        // '
		"undefined variable",                    // easy
		"missing ]",                             // a[1
		"missing )",                             // a[(00
		"cannot call non-function",              // A(0(0)
		"symbol already defined",                // duplicate variable
		"'(' after if",                          // IF ..
		"')' after if",                          //
		"'}' after if",                          //
		"'{' after if",                          //
		"'}' after else",                        // IF .. ELSE
		"'{' after else",                        //
		"'=' after LET",                         //
		"missing '(' after while",               // WHILE
		"missing ')' after while",               //
		"missing '{' after while",               //
		"missing '}' after while",               //
		"missing '(' after return",              // RETURN
		"missing ')' after return",              //
		"unexpected EOF",                        // FUNCTIOn
		"missing '(' after function",            //
		"missing ')' after function",            //
		"missing '}' after function",            //
		"missing '{' after function",            //
		"function names must be identifiers",    //
		"parameter without default value after previously seen a default",
		"argument lengths differ for function", // unCtion A(A){A()
		"cannot assign",                        // let 0=0': cannot assign to *parser.IntegerLiteral
		"only permits a numerical expression",  // if/while
	}

	//
	// Some examples to seed the fuzz corpus
	//
	testcases := []string{
		// simple maths
		"print( 3 + 3 );",
		"print( 3 / 3 );",
		"print( 1 + 2 * 3 );",

		// function call
		"newline();",

		// assignment
		"let a = 3;",
		"let a = \"steve\";",

		// varargs
		`print( "test", "me", 3, 3.2, "\n"`,

		// if
		"if ( 1 ) { print(2); } else { print (3); }",

		// while
		"while ( 1 ) { print(2); }",
		"let a = 3; while ( a > 0 ) { let a = a - 1; }",

		// function
		`function foo() { print("steve\n"); }`,
		`function foo(age, name="Steve") { print(name); }  foo(3)`,

		// return
		"return(3);",
		"let a = 32; return( a ) ; ",
	}

	//
	// Seed the fuzzer with our samples
	//
	for _, tc := range testcases {
		f.Add([]byte(tc))
	}

	//
	// Run the fuzzer.
	//
	f.Fuzz(func(t *testing.T, input []byte) {
		falsePositive := false

		//
		// create a compiler object with our source
		//
		c, err := compiler.New(compiler.WithSource(string(input)))
		if err != nil {
			t.Fatalf("failed to create compiler with source %v %s", input, err)
		}

		//
		// Try to generate some assembly - this runs the lexer, the parser, and the generator.
		//
		// We mostly expect errors from the lexer and parser, the generator is going to
		// produce output if it has some valid AST.
		//
		_, err = c.Compile()
		if err != nil {

			//
			// We got an error, was it a false-positive?
			//
			for _, ignored := range known {
				if strings.Contains(err.Error(), ignored) {
					falsePositive = true
				}
			}

			//
			// If it wasn't a false positive we want to see what
			// was produced and mark it as a failure.
			//
			if !falsePositive {
				t.Fatalf("error running input: '%s': %v", input, err)
			}
		}
	})
}
