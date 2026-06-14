package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestVersionInfo just calls the Info() method for coverage
func TestVersionInfo(t *testing.T) {

	x := versionCommand{}
	_, _ = x.Info()
}

// TestCompileDriver tests we can call the "Execute" method,
// as our CLI would generate.
func TestVersionDriver(t *testing.T) {

	// Ensure the generation goes here.
	var buff bytes.Buffer
	output = &buff

	c := &versionCommand{}
	c.Execute([]string{})

	// Ensure we got some output
	str := buff.String()
	if !strings.Contains(str, "\n") {
		t.Fatalf("output didn't have newlines: %s", str)
	}
	if !strings.Contains(str, "s-lang") {
		t.Fatalf("output didn't have our binary name: %s", str)
	}
}
