package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// TestLexInfo just calls the Info() method for coverage
func TestLexInfo(t *testing.T) {

	x := lexCommand{}
	_, _ = x.Info()
}

// TestLex tries to construct a program, and examine the generated
// lexted output.
func TestLex(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	src := `
return(3);
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// cleanup once done
	defer os.Remove(f.Name())

	x := lexCommand{}
	err = x.lexFile(f.Name())
	if err != nil {
		t.Fatalf("error lexing program %s", err)
	}

	out := buff.String()
	if !strings.Contains(out, "RETURN") {
		t.Fatalf("failed to find expected content in lexer output.")
	}
}

// TestLexBroken tries to parse a missing file.
func TestLexBroken(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	// Remove the file
	os.Remove(f.Name())

	x := lexCommand{}
	err = x.lexFile(f.Name())
	if err == nil {
		t.Fatalf("expected error parsing missing file, got none.")
	}
}

// TestLexDriver tests we can call the "Execute" method,
// as our CLI would generate.
func TestLexDriver(t *testing.T) {
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

	l := &lexCommand{}
	l.Execute([]string{f.Name()})
	l.Execute([]string{})
	l.Execute([]string{"/file/not/found"})

}
