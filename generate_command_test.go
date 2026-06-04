package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
)

// TestGenerateInfo just calls the Info() method for coverage
func TestGenerateInfo(t *testing.T) {

	x := generateCommand{}
	_, _ = x.Info()
}

// TestGeneration tries to construct a program, and examine the generated
// assembly language for it.
func TestGeneration(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	src := `
exit(3);
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// cleanup once done
	defer os.Remove(f.Name())

	x := generateCommand{}
	err = x.processFile(f.Name())
	if err != nil {
		t.Fatalf("error compiling program %s", err)
	}

	out := buff.String()
	if !strings.Contains(out, "rax, 3") {
		t.Fatalf("failed to find expected content in compiled code.")
	}
}

// TestGenerateOutput tries to construct a program, and examine the generated
// assembly language for it.
func TestGenerateOutput(t *testing.T) {

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	src := `
exit(3);
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// Create a temporary file for a fake output
	o, err2 := os.CreateTemp("", "sample")
	if err2 != nil {
		t.Fatalf("error making temporary file")
	}

	// delete the file
	os.Remove(o.Name())

	// cleanup once done
	defer os.Remove(f.Name())

	// Setup the output to go to our file
	x := generateCommand{output: o.Name()}
	err = x.processFile(f.Name())
	if err != nil {
		t.Fatalf("error compiling program %s", err)
	}

	// Read the output
	data, err3 := os.ReadFile(o.Name())
	if err3 != nil {
		t.Fatalf("error reading generated source %s", err)
	}

	if !strings.Contains(string(data), "rax, 3") {
		t.Fatalf("failed to find expected content in compiled code.")
	}
	os.Remove(o.Name())
	os.Remove(f.Name())
}

// TestGenerateBroken confirms a broken program gets an error
func TestGenerateBroken(t *testing.T) {

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
while( 1 ) { break; }
break;
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// cleanup once done
	defer os.Remove(f.Name())

	x := generateCommand{}
	err = x.processFile(f.Name())
	if err == nil {
		t.Fatalf("expected error with program, got none.")
	}

	// Remove the file and try again
	os.Remove(f.Name())

	// This will fail as the file is unreadable / non-existent.
	x = generateCommand{}
	err = x.processFile(f.Name())
	if err == nil {
		t.Fatalf("expected error with program, got none.")
	}
}

// TestGenerateFailureToWrite tests what happens when we cannot
// write to an output file - because it is a directory.
func TestGenerateFailureToWrite(t *testing.T) {

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}
	defer os.Remove(f.Name())

	src := `
exit(3);
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// Create a temporary directory, which will
	// prevent a file from being created for output.
	dir := t.TempDir()

	// Tell the output to go to a file, which will
	// fail, because we've just created a directory
	// there.
	x := generateCommand{output: dir}
	err = x.processFile(f.Name())
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}

// TestGenerateDriver tests we can call the "Execute" method,
// as our CLI would generate.
func TestGenerateDriver(t *testing.T) {
	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	src := `
# This is fine
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	// cleanup once done
	defer os.Remove(f.Name())

	g := &generateCommand{}
	g.Execute([]string{f.Name()})
	g.Execute([]string{})
	g.Execute([]string{"/file/not/found"})

	// remove old file
	os.Remove(f.Name())

	// create a new one
	f, err = os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	src = `
# This is fine.
exit(1);
`
	// The program we'll compile
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	g.Execute([]string{f.Name()})

	// Test flags
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	g.Arguments(flags)

}
