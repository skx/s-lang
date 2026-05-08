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
func (p *Parser) parseExpr() Expr {
	return p.parseAddSub()
}

func (p *Parser) parseAddSub() Expr {
	left := p.parseMulDiv()

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.PLUS &&
			tok.Type != lexer.MINUS {
			break
		}

		p.l.Next()

		right := p.parseMulDiv()

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Type,
			Right: right,
		}
	}

	return left
}

func (p *Parser) parseMulDiv() Expr {
	left := p.parsePrimary()

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.MULTIPLY &&
			tok.Type != lexer.DIVIDE {
			break
		}

		p.l.Next()

		right := p.parsePrimary()

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Type,
			Right: right,
		}
	}

	return left
}

func (p *Parser) parsePrimary() Expr {
	tok := p.l.Next()

	switch tok.Type {

	case lexer.NUMBER:
		return &NumberExpr{
			Value: int64(tok.Value.(float64)),
		}

	case lexer.IDENT:
		return &VariableExpr{
			Name: tok.Value.(string),
		}

	case lexer.LPAREN:
		expr := p.parseExpr()

		if p.l.Next().Type != lexer.RPAREN {
			panic("missing )")
		}

		return expr
	}

	panic("unexpected token")
}

func (p *Parser) parseComparison() Expr {
	left := p.parseAddSub()

	for {
		tok := p.l.Peek()

		switch tok.Type {

		case lexer.LT,
			lexer.LT_EQUALS,
			lexer.GT,
			lexer.GT_EQUALS:

			p.l.Next()

			right := p.parseAddSub()

			left = &BinaryExpr{
				Left:  left,
				Op:    tok.Type,
				Right: right,
			}

		default:
			return left
		}
	}
}
func (p *Parser) parseEquality() Expr {
	left := p.parseComparison()

	for {
		tok := p.l.Peek()

		switch tok.Type {

		case lexer.EQUALS,
			lexer.NOT_EQUALS:

			p.l.Next()

			right := p.parseComparison()

			left = &BinaryExpr{
				Left:  left,
				Op:    tok.Type,
				Right: right,
			}

		default:
			return left
		}
	}
}

func (p *Parser) parseLogicalOr() Expr {
	left := p.parseLogicalAnd()

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.OR {
			return left
		}

		p.l.Next()

		right := p.parseLogicalAnd()

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Type,
			Right: right,
		}
	}
}

func (p *Parser) parseLogicalAnd() Expr {
	left := p.parseEquality()

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.AND {
			return left
		}

		p.l.Next()

		right := p.parseEquality()

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Type,
			Right: right,
		}
	}
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

			expr := p.parseEquality()
			start = p.l.Next()
			if start.Type != lexer.RPAREN {
				return res, fmt.Errorf("missing ')' after if")
			}

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
			res = append(res, &If{Expression: expr, Statements: stmts})

		case lexer.LET:
			name := p.l.Next()
			eq := p.l.Next()

			if eq.Type != lexer.ASSIGN {
				return res, fmt.Errorf("missing '=' after LET")
			}
			vals := p.parseExpr()
			res = append(res,
				&LetStatement{Name: name.Value.(string), Expression: vals})

		case lexer.WHILE:
			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after while")
			}
			val := p.parseEquality()
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
			res = append(res, &While{Expression: val, Statements: stmts})

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
