// Package check provides some limited support for type-checking arguments
// to known standard-library routines.
//
// It can also be used to validate the number of arguments passed to user-defined
// functions.
package check

import (
	"fmt"
)

// Type is used to determine the type a compiled expression returns
type Type int

const (
	_ = iota
	INTEGER
	FLOAT
	NUMBER
	STRING
	UNKNOWN
)

// Types holds our state.
type Types struct {
	// known holds details about known functions, and their arguments.
	//
	// User-define functions can be added.
	known map[string][]Type
}

// New is our constructor.
func New() *Types {
	m := make(map[string][]Type)
	return &Types{
		known: m,
	}
}

// RegisterStdLib adds the calling types for known standard
// library functions.
//
// This is not enabled by default, because we need to be able
// to disable this checking to test our run-time type checking.
func (tc *Types) RegisterStdLib() {

	tc.known["argv"] = []Type{INTEGER}
	tc.known["exit"] = []Type{INTEGER}
	tc.known["fclose"] = []Type{INTEGER}
	tc.known["filesize"] = []Type{STRING}
	tc.known["float2int"] = []Type{FLOAT}
	tc.known["fopen"] = []Type{STRING, STRING}
	tc.known["getenv"] = []Type{STRING}
	tc.known["int2float"] = []Type{INTEGER}
	tc.known["malloc"] = []Type{INTEGER}
	tc.known["memlen"] = []Type{STRING}
	tc.known["mmap"] = []Type{INTEGER}
	tc.known["panic"] = []Type{STRING}
	tc.known["putc"] = []Type{INTEGER}
	tc.known["rand"] = []Type{INTEGER}
	tc.known["readfile"] = []Type{STRING}
	tc.known["sleep"] = []Type{NUMBER}
	tc.known["sqrt"] = []Type{NUMBER}
	tc.known["str2float"] = []Type{STRING}
	tc.known["str2int"] = []Type{STRING}
	tc.known["strcat"] = []Type{STRING, STRING}
	tc.known["strcmp"] = []Type{STRING, STRING}
	tc.known["strdup"] = []Type{STRING}
	tc.known["strlen"] = []Type{STRING}

	// ignored standard library functons:
	//
	//   type: Any type is valid.
	//  print: variadic arguments of any type.
	// printf: variadic arguments of any type.

}

// Type2String converts the given Type to a string description.
func (tc *Types) Type2String(in Type) string {
	switch in {
	case INTEGER:
		return "integer"
	case FLOAT:
		return "float"
	case STRING:
		return "string"
	case UNKNOWN:
		return "unknown"
	case NUMBER:
		return "number"
	default:
		return "CANT HAPPEN"
	}
}

// AddUserFunction adds argument information for functions the user defined,
// we don't do type checking on those, but we can test that the number
// of arguments meets expectations.
func (tc *Types) AddUserFunction(name string, argCount int) {
	t := make([]Type, argCount)
	for i := range t {
		t[i] = UNKNOWN
	}

	tc.known[name] = t
}

// Check is called to see if the given argument types and counts
// are valid for the known standard library function, or registered
// user-function, as added via AddUserFunction.
func (tc *Types) Check(name string, supplied []Type) error {

	// Is this a check of a known function?
	known, ok := tc.known[name]
	if !ok {
		// Unknown function so we'll let it pass
		return nil
	}

	if len(supplied) != len(known) {
		return fmt.Errorf("argument lengths differ for function %s: expected %d, got %d",
			name, len(known), len(supplied))
	}

	// for each expected type see if we have a match
	for i, s := range known {

		// What we received
		got := supplied[i]

		switch s {

		case INTEGER:
			if got != INTEGER && got != UNKNOWN {
				return fmt.Errorf("type mismatch for %s argument %d %s != %s", name, i+1, tc.Type2String(s), tc.Type2String(got))
			}
		case FLOAT:
			if got != FLOAT && got != UNKNOWN {
				return fmt.Errorf("type mismatch for %s argument %d %s != %s", name, i+1, tc.Type2String(s), tc.Type2String(got))
			}
		case NUMBER:
			if got != INTEGER && got != FLOAT && got != UNKNOWN {
				return fmt.Errorf("type mismatch for %s argument %d %s != %s", name, i+1, tc.Type2String(s), tc.Type2String(got))
			}

		case STRING:
			if got != STRING && got != UNKNOWN {
				return fmt.Errorf("type mismatch for %s argument %d %s != %s", name, i+1, tc.Type2String(s), tc.Type2String(got))
			}
		case UNKNOWN:
			// This is for user-functions, and is okay.
		default:
			return fmt.Errorf("type mismatch for %s argument %d %s != %s", name, i+1, tc.Type2String(s), tc.Type2String(got))

		}
	}

	// All okay
	return nil

}
