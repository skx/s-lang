package compiler

import (
	"strings"
	"testing"
)

// TestSanity just tests compiling a couple of simple programs,
// to ensure things look somewhat sane.
func TestSanity(t *testing.T) {

	// Empty program
	c, err := New()
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err2 := c.Compile()
	if err2 != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err2)
	}
	if !strings.Contains(txt, "rax") {
		t.Fatalf("suspicious output")
	}

	// Simple program
	c, err = New(WithSource(`
print("Hello, world!\n");

let a = 3;
print(a);
`))
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}

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
	c, err := New(WithSource(`return "Steve";`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	_, err = c.Compile()
	if err == nil {
		t.Fatalf("expected error, got none.")
	}

	// "if" doesn't like strings
	c, err = New(WithSource(`if ( "Steve" ) { print( 1 ); } `))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	_, err = c.Compile()
	if err == nil {
		t.Fatalf("expected error, got none.")
	}

	// nested functions are illegal
	c, err = New(WithSource(`
function foo() {
   function bar() {
   }
}
`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	_, err = c.Compile()
	if err == nil {
		t.Fatalf("expected error, got none.")
	}
}

// TestConstantFolding attempts to ensure that constant folding works.
func TestConstantFolding(t *testing.T) {

	// Simple program
	c, err := New(WithSource(`
# 7 is the magic number
exit( 1 + 2 * 3);
`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err := c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if !strings.Contains(txt, "rax, 7") {
		t.Fatalf("suspicious output")
	}

	// Now do it again, but this time disable constant
	// folding
	// Simple program
	c, err = New(WithSource(`
# 7 is the magic number
exit( 1 + 2 * 3);
`), WithConstantFolding(false))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err = c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if strings.Contains(txt, "rax, 7") {
		t.Fatalf("suspicious output - looks like we've got a constant")
	}
	if !strings.Contains(txt, "rax, 3") {
		t.Fatalf("suspicious output - missing the literal 3")
	}
	if !strings.Contains(txt, "rax, 2") {
		t.Fatalf("suspicious output - missing the literal 2")
	}
}

// TestConstantIf attempts to ensure that constant IF tests avoid
// generating all the code they might need to.
func TestConstantIf(t *testing.T) {

	// Simple program
	c, err := New(WithSource(`
function true() { }
function false() { }

if ( 1 + 2 * 3 ) {
   true();
} else {
   false();
}
`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err := c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if strings.Contains(txt, "call false") {
		t.Fatalf("suspicious output")
	}
	// 7 -> 28 with our typing.
	if strings.Contains(txt, ", 28") {
		t.Fatalf("suspicious output")
	}

	// Same again, but this time the code will only contain the
	// FALSE block.
	c, err = New(WithSource(`
function true() { }
function false() { }

if ( 0 ) {
   true();
} else {
   false();
}
`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err = c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if strings.Contains(txt, "call true") {
		t.Fatalf("suspicious output")
	}
	if !strings.Contains(txt, "call false") {
		t.Fatalf("suspicious output")
	}

	//
	// Now disable the folding, which will disable
	// the optimization.
	//
	// We should see both branches are present.
	//
	c, err = New(WithSource(`
function true() { }
function false() { }

if ( 1 + 3 ) {
   true();
} else {
   false();
}
`), WithConstantFolding(false))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err = c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if !strings.Contains(txt, "call true") {
		t.Fatalf("suspicious output")
	}
	if !strings.Contains(txt, "call false") {
		t.Fatalf("suspicious output")
	}
}

// TestConstantWhile attempts to ensure that constant WHILE tests
// are optimized correctly.

func TestConstantWhile(t *testing.T) {

	// Simple program
	c, err := New(WithSource(`
function bogus() {  # can't happen
}

while( 0 ) {
   bogus();
}
while( 1 - 1 ) {
   bogus();
}
`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err := c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if strings.Contains(txt, "call bogus") {
		t.Fatalf("suspicious output")
	}

	// Same again, but this time the code will always run
	c, err = New(WithSource(`
function valid() { print("always\n"); }

while( 1 ) {
   valid();
}
while( 1.5 ) {
   valid();
}
`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err = c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if !strings.Contains(txt, "call valid") {
		t.Fatalf("suspicious output")
	}

	//
	// Now disable the folding, which will disable
	// the optimization.
	//
	// We should see both branches are present.
	//
	c, err = New(WithSource(`
function bogus() { }

while( 1 ) {
  bogus();
  break;
}
while( 1.0 ) {
  bogus();
  break;
}
`), WithConstantFolding(false))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err = c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling empty program: %s", err)
	}
	if !strings.Contains(txt, "call bogus") {
		t.Fatalf("suspicious output")
	}
}

// TestAll tests generation of code for "all things"
func TestAll(t *testing.T) {

	// Simple program
	c, err := New(WithSource(`
function test (n) {
  let n = 10 - 4;
  return( n ) ;
}
function testing () {
   return( 1 + ( 4 / 2) );
}
let a = test(2);
let a = a + 2;

while( a ) {
  let a = a - 1;
}
if ( a ) {
  print("non-zero\n");
}
let s = "Steve";
print( s );
print( test(3));
inline { }
`))
	if err != nil {
		t.Fatalf("failed to create compiler")
	}
	txt, err := c.Compile()
	if err != nil {
		t.Fatalf("unexpected error compiling program: %s", err)
	}
	if !strings.Contains(txt, "rax, 6") {
		t.Fatalf("suspicious output")
	}

}
