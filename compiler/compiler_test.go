package compiler

import (
	"strings"
	"testing"
)

// TestSanity just tests compiling a couple of simple programs,
// to ensure things look somewhat sane.
func TestSanity(t *testing.T) {

	// Empty program
	c := New("")
	txt, err := c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if !strings.Contains(txt, "rax") {
		t.Fatalf("suspicious output")
	}

	// Simple program
	c = New(`
print("Hello, world!\n");

let a = 3;
print(a);
`)
	txt, err = c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling basic program: %s", err)
	}
	if !strings.Contains(txt, "rax") {
		t.Fatalf("suspicious output")
	}
}

// TestBroken tests a couple of programs which are broken
func TestBroken(t *testing.T) {

	// "return" cannot handle strings
	c := New(`return "Steve";`)
	_, err := c.Compile()
	if err == nil {
		t.Fatalf("expected error, got none.")
	}

	// "if" doesn't like strings
	c = New(`if ( "Steve" ) { print( 1 ); } `)
	_, err = c.Compile()
	if err == nil {
		t.Fatalf("expected error, got none.")
	}
}
