// Package parser parses input programs, via our lexer, and allows AST nodes
// to be generated - these can then be walked and turned into assembly.
package parser

import (
	"fmt"
	"s-lang/lexer"
)

// Parser object
type Parser struct {
	// l is our lexer
	l *lexer.Lexer

	// curToken holds the current token from our lexer.
	curToken *lexer.Token
}

// New returns our new parser-object.
func New(l *lexer.Lexer) *Parser {

	// Create the parser, and prime the pump
	p := &Parser{l: l}

	// All done
	return p

}

// ParseProgram used to parse the whole program
func (p *Parser) ParseProgram() (*Program, error) {
	program := &Program{}
	var err error
	program.Statements, err = p.parseStatements()
	return program, err
}

// parseExpr is called to parse "LET X = ...." - where we need to handle
// several cases in the "...." section:
//
// Integer Literal
// Register
// Simple expression
//
// Stop at ";" or the end of the input if one is missing.
func (p *Parser) parseExpr() []*lexer.Token {
	res := []*lexer.Token{}

	x := p.l.Next()
	for x.Type != lexer.SEMICOLON && x.Type != lexer.EOF {
		res = append(res, x)
		x = p.l.Next()
	}

	return res
}

// parseConditional is designed to parse the test used in
// an if-statement.
func (p *Parser) parseConditional() []*lexer.Token {
	res := []*lexer.Token{}

	x := p.l.Next()
	for x.Type != lexer.RPAREN && x.Type != lexer.EOF {
		res = append(res, x)
		x = p.l.Next()
	}

	return res
}

// parseStatements is used to parse a collection of statements.
//
// It is extracted into a function because our WHILE and IF blocks
// will themselves contain statements - when they exist.
func (p *Parser) parseStatements() ([]Statement, error) {
	res := []Statement{}

	// Setup the token
	p.curToken = p.l.Next()

	// continue until we hit the end of the file, or the block.
	// (blocks are parsed via recursion for things like "while" and "if")
	for p.curToken.Type != lexer.RBRACE && p.curToken.Type != lexer.EOF {

		switch p.curToken.Type {

		case lexer.SEMICOLON:
			// NOP

		case lexer.IF:
			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after if")
			}

			cnd := p.parseConditional()

			end := p.l.Next()
			if end.Type != lexer.LBRACE {
				return res, fmt.Errorf("missing '{' after if")
			}

			// Now parse the block
			// that will terminate on "}"
			stmts, err := p.parseStatements()
			if err != nil {
				return res, err
			}
			res = append(res, &If{Condition: cnd, Statements: stmts})

		case lexer.LET:
			name := p.l.Next()
			eq := p.l.Next()

			if eq.Type != lexer.ASSIGN {
				return res, fmt.Errorf("missing '=' after LET")
			}
			vals := p.parseExpr()
			if len(vals) != 1 && len(vals) != 3 {
				return res, fmt.Errorf("invalid expression: %v", vals)
			}

			res = append(res,
				&LetStatement{Name: name.Value.(string), Expression: vals})

		case lexer.WHILE:
			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after while")
			}
			val := p.l.Next()
			end := p.l.Next()
			if end.Type != lexer.RPAREN {
				return res, fmt.Errorf("missing ')' after while")
			}
			end = p.l.Next()
			if end.Type != lexer.LBRACE {
				return res, fmt.Errorf("missing '{' after while")
			}

			// Now parse the block
			// that will terminate on "}"
			stmts, err := p.parseStatements()
			if err != nil {
				return res, err
			}
			res = append(res, &While{Value: val, Statements: stmts})

		case lexer.PRINT, lexer.PRINTLN:
			calledAs := p.curToken.Type
			start := p.l.Next()

			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after print")
			}
			tmp := []*lexer.Token{}
			val := p.l.Next()
			for val.Type != lexer.RPAREN && val.Type != lexer.EOF {
				if val.Type != lexer.COMMA {
					tmp = append(tmp, val)
				}
				val = p.l.Next()
			}
			if val.Type != lexer.RPAREN {
				return res, fmt.Errorf("missing ')' after print value")
			}

			res = append(res, &Print{Values: tmp, NewLine: (calledAs == lexer.PRINTLN)})

		case lexer.RETURN:
			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after return")
			}
			val := p.l.Next()
			end := p.l.Next()
			if end.Type != lexer.RPAREN {
				return res, fmt.Errorf("missing ')' after return value")
			}

			res = append(res, &Return{Value: val})

		default:
			return res, fmt.Errorf("uknown token type %v", p.curToken)
		}

		// repeat
		p.curToken = p.l.Next()
	}
	return res, nil
}
