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

	x := parseCommand{}
	err = x.parseFile(f.Name())
	if err == nil {
		t.Fatalf("expected error parsing program, got none.")
	}

	// Remove the file and try again
	os.Remove(f.Name())

	// This will fail as the file is unreadable / non-existent.
	x = parseCommand{}
	err = x.parseFile(f.Name())
	if err == nil {
		t.Fatalf("expected error parsing program, got none.")
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
