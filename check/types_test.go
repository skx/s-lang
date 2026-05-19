package check

import (
	"testing"
)

// TestTrivial tests trivially that we can check known functions
func TestTrivial(t *testing.T) {
	x := New()

	// malloc should be known
	found, ok := x.known["malloc"]
	if !ok {
		t.Fatalf("failed to find 'malloc'")
	}
	if len(found) != 1 {
		t.Fatalf("malloc() should have one argument")
	}
}

// TestTypeStrings checks we can return our constants to strings
func TestTypeStrings(t *testing.T) {
	x := New()

	if x.Type2String(INTEGER) != "integer" {
		t.Fatalf("converting value failed")
	}
	if x.Type2String(STRING) != "string" {
		t.Fatalf("converting value failed")
	}
	if x.Type2String(FLOAT) != "float" {
		t.Fatalf("converting value failed")
	}
	if x.Type2String(UNKNOWN) != "unknown" {
		t.Fatalf("converting value failed")
	}
	if x.Type2String(UNKNOWN+1) != "CANT HAPPEN" {
		t.Fatalf("converting value failed")
	}

}

// TestUserFunctions tests registering a user-function
func TestUserFunctions(t *testing.T) {

	x := New()

	// function "foo" is not known by default
	_, ok := x.known["foo"]
	if ok {
		t.Fatalf("unexpectedly found 'foo'")
	}

	if x.Check("foo", []Type{}) != nil {
		t.Fatalf("unknown function should be fine")
	}

	// Add a user function "foo" which takes two arguments
	x.AddUserFunction("foo", 2)

	// Now we should find it
	_, ok = x.known["foo"]
	if !ok {
		t.Fatalf("failed to find 'foo', after registration")
	}

	// zero, one, and three arguments are bad
	if x.Check("foo", []Type{}) == nil {
		t.Fatalf("unexpected success with zero args")
	}
	if x.Check("foo", []Type{INTEGER}) == nil {
		t.Fatalf("unexpected success with one arg")
	}
	if x.Check("foo", []Type{INTEGER, STRING, STRING}) == nil {
		t.Fatalf("unexpected success with three args")
	}

	// but two arguments will succeed, regardless of their type
	if x.Check("foo", []Type{STRING, STRING}) != nil {
		t.Fatalf("unexpected failure with two args")
	}
	if x.Check("foo", []Type{INTEGER, INTEGER}) != nil {
		t.Fatalf("unexpected failure with two args")
	}
	if x.Check("foo", []Type{UNKNOWN, UNKNOWN}) != nil {
		t.Fatalf("unexpected failure with two args")
	}
	if x.Check("foo", []Type{STRING, INTEGER}) != nil {
		t.Fatalf("unexpected failure with two args")
	}
}
