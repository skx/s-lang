package parser

import (
	"fmt"
	"github.com/skx/s-lang/lexer"
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

// FunctionCallExpr holds a reference to a function to call.
type FunctionCallExpr struct {
	// Name is the name of the function we should call.
	Name string

	// Arguments are what we'll setup for the function
	Arguments []Expr
}

// IndexExpr holds an indexing operation: expr[index]
type IndexExpr struct {
	// Left is a reference to the thing we're indexing
	Left Expr

	// Index is the index we're going to retrieve.
	Index Expr
}

// String converts this structure to a string.
func (i *IndexExpr) String() string {
	return fmt.Sprintf("%s[%s]",
		i.Left.String(),
		i.Index.String())
}

func (IndexExpr) expr() {}

// IndexAssign holds an indexing operation: expr[index] = val
type IndexAssign struct {
	// Name is the name of the thing we're indexing.
	Left Expr

	// Index is the index we're going to update.
	Index Expr

	// Expression is evaluated to set the value.
	Expression Expr
}

// String converts this structure to a string.
func (i *IndexAssign) String() string {
	return fmt.Sprintf("%s[%s] = %s",
		i.Left.String(),
		i.Index.String(),
		i.Expression.String())
}

func (IndexAssign) expr() {}

// String returns the value of the given expression.
func (f *FunctionCallExpr) String() string {
	return fmt.Sprintf("%s(%s)", f.Name, fmt.Sprintf("%v", f.Arguments))
}
func (FunctionCallExpr) expr() {}

// IntegerLiteral holds a literal integer, positive or negative.
type IntegerLiteral struct {
	Value int64
}

// String returns the value of the given expression.
func (n *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", n.Value)
}
func (IntegerLiteral) expr() {}

// FloatLiteral holds a literal float, positive or negative.
type FloatLiteral struct {
	Value float64
}

// String returns the value of the given expression.
func (n *FloatLiteral) String() string {
	return fmt.Sprintf("%f", n.Value)
}
func (FloatLiteral) expr() {}

// StringLiteral holds a literal String.
type StringLiteral struct {
	Value string
}

// String returns the value of the given expression.
func (s *StringLiteral) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}
func (StringLiteral) expr() {}

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

// Break holds a break-statement, only valid within a while loop
type Break struct {
	// Empty
}

// Continue holds a continue-statement, only valid within a while loop
type Continue struct {
	// Empty
}

// Data holds inline assembly the user wants to add to the program.
// Data is like Inline, but guaranteed to be added at the end of the
// generated source.
type Data struct {
	// Text is the raw text to insert into our generated source
	Text string
}

// Function holds details about (user-defined) functions.
type Function struct {
	// Name is the name of the function that is being defined
	Name string

	// Parameters holds the parameters the function accepts
	Parameters []*lexer.Token

	// Statements are the body of the function
	Statements []Statement
}

// If is our conditional operation.
type If struct {
	// Expression is the expression we test before processing the statements
	// within the block.
	Expression Expr

	// True are the things we execute if the condition is true.
	True []Statement

	// False are the statements we execute if the condition is not true.
	False []Statement
}

// Inline holds inline assembly the user wants to add to the program.
// Inline data is embedded whenever it is seen, unlike Data which is
// added to the end of the program.
type Inline struct {
	// Text is the raw text to insert into our generated source
	Text string
}

// Let holds a let-statemnt.  Shocking.
type Let struct {
	// Left is the thing we'll assign to
	Left Expr

	// Expression is the thing to set the name to.
	Expression Expr
}

// Return behaves differently depending on the scope.
//
// In global scope it terminates the running program, with a specified exit-code.
// Inside a function it breaks out of that function, allowing execution to continue
// at the point the function was called from.
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
