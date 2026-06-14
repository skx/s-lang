//go:build linux

package main

import (
	"bytes"
	"flag"
	"os"
	"testing"
)

// TestExecuteInfo just calls the Info() method for coverage
func TestExecuteInfo(t *testing.T) {

	x := executeCommand{}
	_, _ = x.Info()
}

// TestExecuteDriver tests we can call the "Execute" method,
// as our CLI would generate.
func TestExecuteDriver(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	c := &executeCommand{}
	c.Execute([]string{"/file/not/found"})
	c.Execute([]string{"/file/not/found", "/too/many/args"})

	// Test flags
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	c.Arguments(flags)

}

func TestExecute(t *testing.T) {
	// Create a temporary file for a fake source
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		t.Fatalf("error making temporary file")
	}

	// cleanup once done
	defer os.Remove(f.Name())

	// Program
	src := `
print("OK\n");
`
	_, err = f.Write([]byte(src))
	if err != nil {
		t.Fatalf("error writing %s", err)
	}

	x := executeCommand{}
	x.output = "a.out"
	x.verbose = true

	err = x.processFile(f.Name())
	if err != nil {
		t.Fatalf("unexpected error processing %s", err)
	}

	x.Execute([]string{f.Name()})
}
