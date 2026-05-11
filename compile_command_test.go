package main

import (
	"bytes"
	"flag"
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
