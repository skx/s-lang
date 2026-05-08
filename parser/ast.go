package parser

import (
	"s-lang/lexer"
)

// Statement represents a single statement.
type Statement interface {
}

// Program represents a complete program.
type Program struct {
	// Statements is the set of statements which the program is comprised of.
	Statements []Statement
}

// Expr is the interface for expression (maths) parsing.
type Expr interface {
	expr()
}

// NumberExpr holds a literal number
type NumberExpr struct {
	Value int64
}

func (NumberExpr) expr() {}

// VariableExpr holds a variable reference.
type VariableExpr struct {
	Name string
}

func (VariableExpr) expr() {}

// BinaryExpr holds a binop ("a + b", or similar).
type BinaryExpr struct {
	Left  Expr
	Op    lexer.TokenType
	Right Expr
}

func (BinaryExpr) expr() {}

// LetStatement holds a let-statemnt.  Shocking.
type LetStatement struct {
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
	Values []*lexer.Token

	// Show a newline afterward?
	NewLine bool
}

// Return terminates the running program, with a specified exit-code.
type Return struct {
	// Value to return is IDENTIFIER, or INTEGERLITERAL.
	Value *lexer.Token
}

// While is our looping operation which currently allows a block of code to
// be repeated while a variable contains a non-zero value.
type While struct {
	// Value is the register we test against.
	// TODO: Parse this in the same we will for IF.
	Value *lexer.Token

	// Statements are the things we execute while the condition is true
	Statements []Statement
}

// If is our conditional operation - note this is not yet implemented, and
// when it is we don't have an Else facility in mind.
type If struct {
	// Condition is the thing we use to decide if to execute the block
	// It will either be a single token like "b", or a simple three-value alternative [a <= b]
	// TODO: This should be shared with WHILE.
	Condition []*lexer.Token

	// Statements are the things we execute if the condition is true
	Statements []Statement
}
