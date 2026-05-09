package parser

import (
	"fmt"
	"s-lang/lexer"
)

// Statement represents a single statement.
type Statement interface {
}

// Program represents a complete program.
//
// Programs contain Statements.
type Program struct {
	// Statements is the set of statements which the program is comprised of.
	Statements []Statement
}

// Expr is the interface for expression (maths) parsing.
type Expr interface {
	expr()

	// String returns the value of a given expression.
	String() string
}

// IntegerExpr holds a reference to a function to call
type FunctionCallExpr struct {
	// Name is the name of the function we should call
	Name string
}

// String returns the value of the given expression.
func (f *FunctionCallExpr) String() string {
	return fmt.Sprintf("%s();", f.Name)
}
func (FunctionCallExpr) expr() {}

// IntegerExpr holds a literal integer, positive or negative.
type IntegerExpr struct {
	Value int64
}

// String returns the value of the given expression.
func (n *IntegerExpr) String() string {
	return fmt.Sprintf("%d", n.Value)
}
func (IntegerExpr) expr() {}

// StringExpr holds a literal String.
type StringExpr struct {
	Value string
}

// String returns the value of the given expression.
func (s *StringExpr) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}
func (StringExpr) expr() {}

// VariableExpr holds a variable reference.
type VariableExpr struct {
	Name string
}

// String returns the value of the given expression.
func (v *VariableExpr) String() string {
	return (v.Name)
}
func (VariableExpr) expr() {}

// BinaryExpr holds a binop ("a + b", or similar).
type BinaryExpr struct {
	Left  Expr
	Op    lexer.TokenType
	Right Expr
}

// String returns the value of the given expression.
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s",
		b.Left.String(),
		b.Op,
		b.Right.String())
}

func (BinaryExpr) expr() {}

// Inline holds inline assembly the user wants to add to the program
type Inline struct {
	// Text is the raw text to insert into our generated source
	Text string
}

// Function holds details about (user-defined) functions.
type Function struct {
	Name       string
	Statements []Statement
}

// Let holds a let-statemnt.  Shocking.
type Let struct {
	// Name is the name of the variable to which we're assigning
	Name string

	// Expression is the thing to set the name to.
	Expression Expr
}

// Print represents a call to our stdlib "print(...)" function.
// This is also called by the "println(...)" function parser too.
type Print struct {
	// Values holds a list of IDENTIFIER, INTEGERLITERAL, and STRINGLITERAL
	// which should be printed
	Values []Expr

	// Show a newline afterward?
	NewLine bool
}

// Return terminates the running program, with a specified exit-code.
type Return struct {
	// Expression is the return code we'll have
	Expression Expr
}

// While is our looping operation which currently allows a block of code to
// be repeated while a variable contains a non-zero value.
type While struct {
	// Expression is the expression we evaluate each time through the loop
	Expression Expr

	// Statements are the things we execute while the condition is true
	Statements []Statement
}

// If is our conditional operation - note this is not yet implemented, and
// when it is we don't have an Else facility in mind.
type If struct {
	// Expression is the expression we test before processing the statements
	// within the block.
	Expression Expr

	// Statements are the things we execute if the condition is true
	Statements []Statement
}
