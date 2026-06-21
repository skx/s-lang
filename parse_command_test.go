package main

import (
	"bytes"
	"github.com/skx/s-lang/parser"
	"os"
	"strings"
	"testing"
)

// TestParseInfo just calls the Info() method for coverage
func TestParseInfo(t *testing.T) {

	x := parseCommand{}
	_, _ = x.Info()
}

// TestParse tries to construct a program, and examine the generated
// parsed output.
func TestParse(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	src := `
function steve() { return(17); }
let a = steve();
let b = 1 + 2 * 3;
print("OK\n", a, b, 3, steve());
println("hello, world!\n");
let a = 3;
while( a ) {
  let a = a - 1;
  continue;
}
while( 1 ) {
  break;
}
if ( 1 ) {
  print("one\n");
}
inline { }
return(3);

for( i = 0; i < 10 ; i++ ) { printf("%d\n", i); }
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// cleanup once done
	defer os.Remove(f.Name())

	x := parseCommand{}
	err = x.parseFile(f.Name())
	if err != nil {
		t.Fatalf("error parsing program %s", err)
	}

	out := buff.String()
	if !strings.Contains(out, "RETURN") {
		t.Fatalf("failed to find expected content in parser output.")
	}
}

// TestParseBroken confirms a broken program gets an error
func TestParseBroken(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	// cleanup once done
	defer os.Remove(f.Name())

	// The tests that each have a broken program
	tests := []string{
		`
for ( i = 0, i < 10, i++ ) {
  # This is an error
  return("Steve

`,
		`if ( true ) { return( "S `,
		`if ( true ) { return ; } else { return( "S `,
		`function foo() { return( "...`,
		`switch a { case 1 { return 1; } case 2 { return "S } } `,
	}

	for _, src := range tests {

		_, err = f.Write([]byte(src))
		if err != nil {
			t.Fatalf("error writing %s", err)
		}

		x := parseCommand{}
		err = x.parseFile(f.Name())
		if err == nil {
			t.Fatalf("expected error parsing program, got none.")
		}
	}

	// This will fail as the file is unreadable / non-existent.
	x := parseCommand{}
	err = x.parseFile(f.Name())
	if err == nil {
		t.Fatalf("expected error parsing program, got none.")
	}

}

// TestParseError ensures we can abort on unknown types
func TestParseError(t *testing.T) {

	// IntegerLiteral is not handled as a bare statement,
	// it is part of an assignment, binop, or similar.
	//
	// So finding it inside a function body is a problem
	fn := &parser.Function{
		Name: "hello",
		Statements: []parser.Statement{
			&parser.IntegerLiteral{},
		},
	}
	p := &parseCommand{}
	err := p.printStmt(fn)
	if err == nil {
		t.Fatalf("expected function error, got none")
	}

	// IntegerLiteral is not handled as a bare statement,
	// it is part of an assignment, binop, or similar.
	//
	// So similar story here, we have two if-cases to cover
	// one in the true-block and one in the false.
	if1 := &parser.If{
		Expression: &parser.IntegerLiteral{},
		True: []parser.Statement{
			&parser.IntegerLiteral{},
		},
	}
	if2 := &parser.If{
		Expression: &parser.IntegerLiteral{},
		False: []parser.Statement{
			&parser.IntegerLiteral{},
		},
	}
	err = p.printStmt(if1)
	if err == nil {
		t.Fatalf("expected if1 error, got none")
	}
	err = p.printStmt(if2)
	if err == nil {
		t.Fatalf("expected if2 error, got none")
	}

	// IntegerLiteral is not handled as a bare statement,
	// it is part of an assignment, binop, or similar.
	//
	// So similar story here, inside a case statement this
	// is an error.
	swtch := &parser.Switch{
		Value: &parser.IntegerLiteral{},
		Choices: []*parser.Case{
			&parser.Case{
				Expression: &parser.IntegerLiteral{},
				Statements: []parser.Statement{
					&parser.IntegerLiteral{},
				},
			},
		},
	}
	err = p.printStmt(swtch)
	if err == nil {
		t.Fatalf("expected switch error, got none")
	}

	// Final one
	//
	// IntegerLiteral is not handled as a bare statement,
	// it is part of an assignment, binop, or similar.
	//
	// So similar story here, we have to cover the While
	// case.
	whle := &parser.While{
		Expression: &parser.IntegerLiteral{},
		Statements: []parser.Statement{
			&parser.IntegerLiteral{},
		},
	}
	err = p.printStmt(whle)
	if err == nil {
		t.Fatalf("expected while error, got none")
	}

}

// TestParsePrintStmt tests printing bogus things, to ensure
// errors are caught
func TestParsePrintStmt(t *testing.T) {

	p := &parseCommand{}

	// nil statement
	err := p.printStmt(nil)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(err.Error(), "unknown item at printStmt") {
		t.Fatalf("got error, but the wrong one: %s", err)
	}

	// nil statement in while() body
	w := &parser.While{
		// Need an expression otherwise the printing will segfault
		Expression: &parser.IntegerLiteral{Value: 13},
		Statements: []parser.Statement{nil},
	}
	err = p.printStmt(w)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(err.Error(), "unknown item at printStmt") {
		t.Fatalf("got error, but the wrong one: %s", err)
	}

	// nil statement in if() body
	i := &parser.If{
		// Need an expression otherwise the printing will segfault
		Expression: &parser.IntegerLiteral{Value: 3},
		True:       []parser.Statement{nil},
	}
	err = p.printStmt(i)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(err.Error(), "unknown item at printStmt") {
		t.Fatalf("got error, but the wrong one: %s", err)
	}
}

// TestParseDriver tests we can call the "Execute" method,
// as our CLI would generate.
func TestParseDriver(t *testing.T) {
	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	src := `
# This is an error
return("Steve
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// cleanup once done
	defer os.Remove(f.Name())

	p := &parseCommand{}
	p.Execute([]string{f.Name()})
	p.Execute([]string{})
	p.Execute([]string{"/file/not/found"})

}

func TestParseAll(t *testing.T) {

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	// Simple program
	src := `
function test (n) {
  let n = 10 - 4;

  return( n ) ;
}
function defaults (a, b = "steve", c=442) {
}
function testing () {
   return( 1 + ( 4 / 2) );
}
let a = test(2);
let a = a + 2;

defaults("none");
defaults("none", "none");

while( a ) {
pragma a size8
  let a = a - 1;
a--;
a++;
a--;
}
if ( a ) {
  print("non-zero\n");
} else {
  print("Trouble at the mill\n");
}
let s = "Steve";
print(s[0], "\n");
s[0] = 42;
putc(s[0]);
while( 1 ) {
  break;
}
let a = 3;
while( a ) {
   a++;
   x = !a;
   y = -a;
   z = +a;
   continue;
   a--;
}

c = 7;
switch c {
  case 3 {  print("three\n");
  }
  case 0 {
	    print("zero\n");
  }
  default {
	    print("default\n");
  }
}
print( s );
print( test(3));
inline { }

# maths
a = 1;
b = 3;

# maths
print( a + b, "\n");
print( a - b, "\n");
print( a * b, "\n");
print( a / b, "\n");
print( a % b, "\n");
print( a ^ b, "\n");

# comparison
print( a <  b, "\n");
print( a <= b, "\n");
print( a >  b, "\n");
print( a >=  b, "\n");
print( a ==  b, "\n");
print( a !=  b, "\n");
print( a &&  b, "\n");
print( a ||  b, "\n");

# collapse constants
print( 1 + 1 , "\n" );
print( 1 - 1 , "\n" );
print( 1 / 1 , "\n" );
print( 1 * 1 , "\n" );
print( 1 % 1 , "\n" );
print( 1 ^ 1 , "\n" );

print( 1.0 + 1.0 , "\n" );
print( 1.0 - 1.0 , "\n" );
print( 1.0 / 1.0 , "\n" );
print( 1.0 * 1.0 , "\n" );

data {
}
inline {
}
function greet(name = "Steve") { print("Hello, ", name , "\n");}
greet();
greet("World");
greet(32.2);
pragma foo bar
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// cleanup once done
	defer os.Remove(f.Name())

	x := parseCommand{}
	err = x.parseFile(f.Name())
	if err == nil {
		t.Fatalf("expected error parsing program, got none.")
	}

	// Remove the file and try again
	os.Remove(f.Name())

}
