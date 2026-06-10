package compiler

import (
	"fmt"
)

// Symbol is an interface to a symbol.
type Symbol interface {

	// SymbolName returns the name of the symbol
	SymbolName() string
}

// GlobalVariable holds global variables.
type GlobalVariable struct {
	// Name is the name of the variable.
	Name string

	Label string
}

// SymbolName returns the name of the symbol, and is part of the Symbol
// interface.
func (g *GlobalVariable) SymbolName() string { return g.Name }

// FunctionVariable holds details of scoped/functional variables.
type FunctionVariable struct {
	// Name is the name of the variable
	Name string

	// Offset, relative to RBP, is where the variable is stored.
	Offset int
}

// SymbolName returns the name of the symbol, and is part of the Symbol
// interface.
func (f *FunctionVariable) SymbolName() string { return f.Name }

// Scope stores stack-frames, and allows us to create new frames for
// functions.
type Scope struct {
	// Parent holds a reference to a possible parent frame,
	// and allows variables in lower scopes to access parent ones.
	Parent *Scope

	// Symbols holds the symbols in this scope.
	Symbols map[string]Symbol

	// nextLocalOffset starts at -8 and grows downward
	nextLocalOffset int
}

// NewScope defines a new scope, with reference to an optional
// parent.
func NewScope(parent *Scope) *Scope {
	s := &Scope{
		Parent:          parent,
		Symbols:         make(map[string]Symbol),
		nextLocalOffset: -8,
	}

	if parent != nil {
		s.nextLocalOffset = parent.nextLocalOffset
	}

	return s
}

// Define defines a new symbol within the current scope,
// if this already exists it is denied.
func (s *Scope) Define(sym Symbol) error {
	name := sym.SymbolName()

	if _, exists := s.Symbols[name]; exists {
		return fmt.Errorf("symbol already defined: %s", name)
	}

	s.Symbols[name] = sym
	return nil
}

// Lookup returns a pre-existing symbol, if it exists.
// Higher scopes are consulted if necessary.
func (s *Scope) Lookup(name string) (Symbol, bool) {
	cur := s

	for cur != nil {
		if sym, ok := cur.Symbols[name]; ok {
			return sym, true
		}
		cur = cur.Parent
	}

	return nil, false
}

// DefineArgument defines a new argument for a function.
func (s *Scope) DefineArgument(name string, argIndex int) (*FunctionVariable, error) {

	// SystemV:
	//
	// [rbp+8]   return address
	// [rbp+16]  arg0
	// [rbp+24]  arg1
	//
	offset := 16 + (argIndex * 8)

	v := &FunctionVariable{
		Name:   name,
		Offset: offset,
	}

	if err := s.Define(v); err != nil {
		return nil, err
	}

	return v, nil
}

// DefineLocal defines a local symbol, creating a suitable
// offset for it.
func (s *Scope) DefineLocal(name string) (*FunctionVariable, error) {
	v := &FunctionVariable{
		Name:   name,
		Offset: s.nextLocalOffset,
	}

	s.nextLocalOffset -= 8

	if err := s.Define(v); err != nil {
		return nil, err
	}

	return v, nil
}

// DefineGlobalVariable declares a variable in the topmost scope.
//
// We use this specifically to define global-variables.  If we were to define
// a global variable inside a child-scope it would go .. out of scope .. when
// that frame was cleaned up and removed.  Breaking the very idea of a global
// variable.
func (s *Scope) DefineGlobalVariable(sym Symbol) error {
	cur := s

	// get the topmost scope
	for cur.Parent != nil {
		cur = cur.Parent
	}

	// ensure it doesn't already exist
	name := sym.SymbolName()
	if _, exists := cur.Symbols[name]; exists {
		return fmt.Errorf("symbol already defined: %s", name)
	}

	// define
	cur.Symbols[name] = sym
	return nil
}

// GetAllGlobals is a helper method to get all known global
// variables.  We use this to populate our rendered assembly
// language template.
func (s *Scope) GetAllGlobals() []*GlobalVariable {

	res := []*GlobalVariable{}

	// Get all globals
	for _, ent := range s.Symbols {
		switch s := ent.(type) {
		case *GlobalVariable:
			res = append(res, s)
		}
	}

	// Add any the parent knows too.  Recursively.
	if s.Parent != nil {
		res = append(res, s.Parent.GetAllGlobals()...)
	}

	return res
}
