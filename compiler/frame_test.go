package compiler

import (
	"testing"
)

func TestGlobal(t *testing.T) {

	// Create a new scope
	parent := NewScope(nil)

	// now a child
	child := NewScope(parent)

	// There should be no global variables
	if len(child.GetAllGlobals()) != 0 {
		t.Fatalf("unexpected global count")
	}

	// Define a global
	g := &GlobalVariable{
		Name:  "Me",
		Label: "Steve",
	}

	// defining it should be fine
	err := child.DefineGlobalVariable(g)
	if err != nil {
		t.Fatalf("error defining global")
	}

	// but repeats should fail
	err = child.DefineGlobalVariable(g)
	if err == nil {
		t.Fatalf("expected error defining duplicate variable")
	}
	// regardless of scope
	err = parent.DefineGlobalVariable(g)
	if err == nil {
		t.Fatalf("expected error defining duplicate variable")
	}

	// Now we should have only one result
	a := parent.GetAllGlobals()
	b := child.GetAllGlobals()
	if len(a) != len(b) {
		t.Fatalf("count mismatch")
	}
	if len(a) != 1 {
		t.Fatalf("only one variable should exist")
	}
}

func TestLocal(t *testing.T) {

	// Create a new scope
	parent := NewScope(nil)

	// define a variable
	_, err := parent.DefineArgument("foo", 8)
	if err != nil {
		t.Fatalf("expected no error, got one %s", err)
	}

	// repeat the definition, which should fail
	_, err = parent.DefineArgument("foo", 8)
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}

	_, err = parent.DefineLocal("steve")
	if err != nil {
		t.Fatalf("expected no error, but got one %s", err)
	}

	// second time should fail
	_, err = parent.DefineLocal("steve")
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}
}
