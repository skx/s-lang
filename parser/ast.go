// ast.go - holds our AST definitions.

package parser

import (
	"fmt"
	"github.com/skx/s-lang/lexer"
)

// Program represents a complete program, which consists of an arbitrary number of statements.
type Program struct {
	// Statements is the set of statements which the program is comprised of.
	Statements []Statement
}

// Statement represents a single statement.
type Statement interface {
}

// Expr is the interface for expression (maths) parsing.
type Expr interface {
	// expr is used to enforce types on our expressions.
	expr()

	// String returns the value of a given expression.
	String() string
}

////
//// Now the more concrete statements
////
//// Ordered alphabetically.
////

// Break holds a break-statement, these are only valid within a while loop.
type Break struct {
	// Empty
}

// Case handles a specific case within a switch statement.
//
// We support a default-handler if no other case matches, but it is an error to
// have more than one default case.
type Case struct {

	// Default branch?
	Default bool

	// Expression is the thing we match
	Expression Expr

	// Statements holds the code to execute if there is a match.
	Statements []Statement
}

// Continue holds a continue-statement, these are only valid within a while loop.
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

// For handles our for-loop
type For struct {

	// Init is the initialization expression.
	Init Statement

	// Cmp is the comparison expression which tests to see
	// if we should execute the body.
	Cmp Expr

	// Post is the expression we add to the end of the body.
	Post Statement

	// Statements are the body of the expression
	Statements []Statement
}

// Function holds details about (user-defined) functions.
type Function struct {
	// Name is the name of the function that is being defined
	Name string

	// Parameters holds the parameters the function accepts
	Parameters []*FunctionParameter

	// Statements are the body of the function
	Statements []Statement
}

// FunctionParameter is a structure to hold details of the
// parameters a function accepts.  It is used to allow an
// optional default to be specified for each parameter.
type FunctionParameter struct {
	// Name is the name of the parameter
	Name string

	// Default, if set, is the value of the parameter
	Default Expr
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

// Pragma holds an arbitrary key=value entry.
//
// We use this only to modify the "size" of index-operations against strings/memory
// at the moment.  But it is possible we might do more in the future.
type Pragma struct {

	// Key contains the key of the pragma entry.
	Key string

	// Value is the value stored with the specified key.
	Value string
}

// Return terminates execution of a function.
//
// It is special because there might be an expression, or there might not be.
// This means "return;" is valid, as is "return(3);" - if a value must be returned
// it will require the brackets and trailing ";".
// at the point the function was called from.
type Return struct {

	// Expression is used for the return value, if specified.
	Expression Expr
}

// Switch handles a switch statement
type Switch struct {

	// Value is the thing that is evaluated to determine
	// which block should be executed.
	Value Expr

	// The branches we handle
	Choices []*Case
}

// While is our looping operation which allows a block of code to
// be repeated while a condition evaluates to a true-value.
type While struct {
	// Expression is the expression we evaluate each time through the loop.
	Expression Expr

	// Statements are the things we execute while the condition is true.
	Statements []Statement
}

////
//// Now the more concrete expression-types.
////
//// Ordered alphabetically.
////

// BinaryExpr holds a binary expression, i.e. two arguments with
// an operation applied to them.
type BinaryExpr struct {
	// Left represents the left-side of the expression.
	Left Expr

	// Right holds the right-side of the expression.
	Right Expr

	// Op holds the operation to be carried out upon the two
	// operands (+, -, etc)
	Op lexer.TokenType
}

// String returns the value of the given expression.
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("%s %s %s",
		b.Left.String(),
		b.Op,
		b.Right.String())
}

func (BinaryExpr) expr() {}

// FloatLiteral holds a literal float, positive or negative.
type FloatLiteral struct {
	Value float64
}

// String returns the value of the given expression.
func (n *FloatLiteral) String() string {
	return fmt.Sprintf("%f", n.Value)
}
func (FloatLiteral) expr() {}

// FunctionCallExpr represents a function-call..
type FunctionCallExpr struct {
	// Name is the name of the function we should call.
	Name string

	// Arguments are what we'll setup for the function
	Arguments []Expr
}

// String returns the value of the given expression.
func (f *FunctionCallExpr) String() string {
	return fmt.Sprintf("%s(%s)", f.Name, fmt.Sprintf("%v", f.Arguments))
}
func (FunctionCallExpr) expr() {}

// IndexAssign holds an indexing operation: expr[index] = val
type IndexAssign struct {
	// Name is the thing we're indexing.
	Left Expr

	// Index is the index we're going to update.
	Index Expr

	// Expression is evaluated to set the value.
	Expression Expr

	// The line number where this appears
	Line int
}

// String converts this structure to a string.
func (i *IndexAssign) String() string {
	return fmt.Sprintf("%s[%s] = %s  # Line %d",
		i.Left.String(),
		i.Index.String(),
		i.Expression.String(),
		i.Line)
}

func (IndexAssign) expr() {}

// IndexExpr holds an indexing operation: expr[index]
type IndexExpr struct {
	// Left is the thing we're indexing
	Left Expr

	// Index is the index we're going to retrieve.
	Index Expr

	// Line is the line number in the source where the statement was found
	Line int
}

// String converts this structure to a string.
func (i *IndexExpr) String() string {
	return fmt.Sprintf("%s[%s]   # Line %d",
		i.Left.String(),
		i.Index.String(),
		i.Line)
}

func (IndexExpr) expr() {}

// IntegerLiteral holds a literal integer, positive or negative.
type IntegerLiteral struct {
	Value int64
}

// String returns the value of the given expression.
func (n *IntegerLiteral) String() string {
	return fmt.Sprintf("%d", n.Value)
}
func (IntegerLiteral) expr() {}

// PrefixExpr holds prefix expressions.
//
// We already cope with negative numbers, so we don't have a unary-minus,
// our only prefix expression is ! for the logical not.
type PrefixExpr struct {
	Expr Expr
	Op   lexer.TokenType
}

// String returns the value of the given expression.
func (p *PrefixExpr) String() string {
	return fmt.Sprintf("%s%s",
		p.Op,
		p.Expr.String())
}

func (PrefixExpr) expr() {}

// PostfixExpr holds a postfix expression,
type PostfixExpr struct {
	// Expr holds the value we're applying the operation to.
	Expr Expr

	// Op holds the operation we're applying (++, --, etc).
	Op lexer.TokenType
}

// String returns the value of the given expression.
func (p *PostfixExpr) String() string {
	return fmt.Sprintf("%s %s",
		p.Expr.String(),
		p.Op)
}

func (PostfixExpr) expr() {}

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
