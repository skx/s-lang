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

// parseExpr is called to parse expressions - be they in IF, LET, or WHILE.
func (p *Parser) parseExpr() (Expr, error) {
	return p.parseLogicalOr()
}

func (p *Parser) parseAddSub() (Expr, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.PLUS &&
			tok.Type != lexer.MINUS {
			break
		}

		p.l.Next()

		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Type,
			Right: right,
		}
	}

	return left, nil
}

func (p *Parser) parseMulDiv() (Expr, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.MULTIPLY &&
			tok.Type != lexer.DIVIDE {
			break
		}

		p.l.Next()

		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Type,
			Right: right,
		}
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Expr, error) {
	tok := p.l.Next()

	switch tok.Type {

	case lexer.NUMBER:
		return &NumberExpr{
			Value: int64(tok.Value.(float64)),
		}, nil

	case lexer.IDENT:
		return &VariableExpr{
			Name: tok.Value.(string),
		}, nil

	case lexer.STRING:
		return &StringExpr{
			Value: tok.Value.(string),
		}, nil

	case lexer.LPAREN:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		if p.l.Next().Type != lexer.RPAREN {
			return expr, fmt.Errorf("missing )")
		}

		return expr, nil
	}

	return nil, fmt.Errorf("unexpected token %v", tok)
}

func (p *Parser) parseComparison() (Expr, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.l.Peek()

		switch tok.Type {

		case lexer.LT,
			lexer.LT_EQUALS,
			lexer.GT,
			lexer.GT_EQUALS:

			p.l.Next()

			right, err := p.parseAddSub()
			if err != nil {
				return nil, err
			}

			left = &BinaryExpr{
				Left:  left,
				Op:    tok.Type,
				Right: right,
			}

		default:
			return left, nil
		}
	}
}

func (p *Parser) parseEquality() (Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.l.Peek()

		switch tok.Type {

		case lexer.EQUALS,
			lexer.NOT_EQUALS:

			p.l.Next()

			right, err := p.parseComparison()
			if err != nil {
				return nil, err
			}

			left = &BinaryExpr{
				Left:  left,
				Op:    tok.Type,
				Right: right,
			}

		default:
			return left, nil
		}
	}
}

func (p *Parser) parseLogicalOr() (Expr, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.OR {
			return left, nil
		}

		p.l.Next()

		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}

		left = &BinaryExpr{
			Left:  left,
			Op:    tok.Type,
			Right: right,
		}
	}
}

func (p *Parser) parseLogicalAnd() (Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.l.Peek()

		if tok.Type != lexer.AND {
			return left, nil
		}

		p.l.Next()

		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}

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

			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

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

		case lexer.INLINE:
			res = append(res, &Inline{Text: p.curToken.Value.(string)})

		case lexer.LET:
			name := p.l.Next()
			eq := p.l.Next()

			if eq.Type != lexer.ASSIGN {
				return res, fmt.Errorf("missing '=' after LET")
			}
			vals, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			res = append(res,
				&LetStatement{Name: name.Value.(string), Expression: vals})

		case lexer.WHILE:
			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after while")
			}
			val, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
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

			var x []Expr

			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			x = append(x, expr)

			for {
				tok := p.l.Peek()
				if tok.Type == lexer.RPAREN {
					p.l.Next()
					break
				}
				if tok.Type == lexer.COMMA {
					p.l.Next()
					expr, err = p.parseExpr()
					if err != nil {
						return nil, err
					}

					x = append(x, expr)
				}
			}

			res = append(res, &Print{Values: x, NewLine: (calledAs == lexer.PRINTLN)})

		case lexer.RETURN:
			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after return")
			}
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			end := p.l.Next()
			if end.Type != lexer.RPAREN {
				return res, fmt.Errorf("missing ')' after return value")
			}

			res = append(res, &Return{Expression: expr})

		default:
			return res, fmt.Errorf("uknown token type %v", p.curToken)
		}

		// repeat
		p.curToken = p.l.Next()
	}
	return res, nil
}
