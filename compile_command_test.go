//go:build linux

package main

import (
	"bytes"
	"flag"
	"os"
	"testing"
)

// TestCompileInfo just calls the Info() method for coverage
func TestCompileInfo(t *testing.T) {

	x := compileCommand{}
	_, _ = x.Info()
}

// TestCompileDriver tests we can call the "Execute" method,
// as our CLI would generate.
func TestCompileDriver(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	c := &compileCommand{}
	c.Execute([]string{"/file/not/found"})
	c.Execute([]string{"/file/not/found", "/too/many/args"})

	// Test flags
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	c.Arguments(flags)

}

func TestCompile(t *testing.T) {
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

	x := compileCommand{}
	x.output = "s.out"
	x.verbose = true

	err = x.processFile(f.Name())
	if err != nil {
		t.Fatalf("unexpected error processing %s", err)
	}

	x.Execute([]string{f.Name()})
}
