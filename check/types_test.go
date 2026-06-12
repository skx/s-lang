package check

import (
	"testing"
)

// TestTrivial tests trivially that we can check known functions
func TestTrivial(t *testing.T) {
	x := New()
	x.RegisterStdLib()

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

// TestStdlibTypes does some minimal checking of stdlib types - this caught the issue
// with FLOAT handling actually comparing against an integer.
func TestStdlibTypes(t *testing.T) {
	x := New()
	x.RegisterStdLib()

	// check each registered type works as expected when given an argument
	// of the same type.
	for fun, arg := range x.known {

		switch arg[0] {
		case NUMBER:
			// If arg is a number try with both int and float.
			//
			// This works because we have zero functions that take a number
			// as a second argument - functions either take a single number-arg,
			// or arguments of other types.
			if x.Check(fun, []Type{INTEGER}) != nil {
				t.Fatalf("unexpected result for function %s with integer argument", fun)
			}
			if x.Check(fun, []Type{FLOAT}) != nil {
				t.Fatalf("unexpected result for function %s with float argument", fun)
			}

		default:
			// Otherwise just use the same argument type
			if x.Check(fun, arg) != nil {
				t.Fatalf("unexpected result for function %s", fun)
			}
		}
	}
}
