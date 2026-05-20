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
	left, err := p.parsePostfix()
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

		right, err := p.parsePostfix()
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

func (p *Parser) parsePostfix() (Expr, error) {
	left, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	for {
		switch p.l.Peek().Type {

		// function call
		case lexer.LPAREN:
			p.l.Next() // consume '('

			var args []Expr

			for {
				t := p.l.Peek()

				if t.Type == lexer.RPAREN {
					p.l.Next()
					break
				}

				if t.Type == lexer.COMMA {
					p.l.Next()
					continue
				}

				expr, err := p.parseExpr()
				if err != nil {
					return nil, err
				}

				args = append(args, expr)
			}

			// only identifiers are callable right now
			v, ok := left.(*VariableExpr)
			if !ok {
				return nil, fmt.Errorf("cannot call non-function")
			}

			left = &FunctionCallExpr{
				Name:      v.Name,
				Arguments: args,
			}

		// index operation
		case lexer.LINDEX:
			p.l.Next() // consume '['

			index, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			if p.l.Next().Type != lexer.RINDEX {
				return nil, fmt.Errorf("missing ]")
			}

			left = &IndexExpr{
				Left:  left,
				Index: index,
			}

		default:
			return left, nil
		}
	}
}

func (p *Parser) parseAtom() (Expr, error) {
	tok := p.l.Next()

	switch tok.Type {

	case lexer.INTEGER:
		return &IntegerLiteral{
			Value: int64(tok.Value.(float64)),
		}, nil

	case lexer.FLOAT:
		return &FloatLiteral{
			Value: tok.Value.(float64),
		}, nil

	case lexer.IDENT:
		return &VariableExpr{
			Name: tok.Value.(string),
		}, nil

	case lexer.STRING:
		return &StringLiteral{
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

	return nil, fmt.Errorf("unexpected token in parsePrimary %v", tok)
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
			lexer.LTEQUALS,
			lexer.GT,
			lexer.GTEQUALS:

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
			lexer.NOTEQUALS:

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

		case lexer.INTEGER, lexer.FLOAT, lexer.STRING:
			return res, fmt.Errorf("bare literal is illegal: %s", p.curToken.String())
		case lexer.BREAK:
			res = append(res, &Break{})

		case lexer.CONTINUE:
			res = append(res, &Continue{})

		case lexer.DATA:
			res = append(res, &Data{Text: p.curToken.Value.(string)})

		case lexer.SEMICOLON:
			// NOP

		case lexer.FUNCTION:
			name := p.l.Next()

			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after function name %s", name)
			}

			// collect parameters
			params := []*lexer.Token{}

			for {
				tok := p.l.Next()
				// )?  Then we're at the end
				if tok.Type == lexer.RPAREN {
					break
				}
				// skip the comments
				if tok.Type == lexer.COMMA {
					continue
				}
				if tok.Type == lexer.IDENT {
					params = append(params, tok)
				}
			}

			end := p.l.Next()
			if end.Type != lexer.LBRACE {
				return res, fmt.Errorf("missing '{' after function definition")
			}

			// Now parse the block
			// that will terminate on "}"
			stmts, err := p.parseStatements()
			if err != nil {
				return res, err
			}
			res = append(res, &Function{Name: name.Value.(string), Parameters: params, Statements: stmts})

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

			var False []Statement

			// Is there a false block?
			tok := p.l.Peek()
			if tok.Type == lexer.ELSE {
				p.l.Next()

				end := p.l.Next()
				if end.Type != lexer.LBRACE {
					return res, fmt.Errorf("missing '{' after else")
				}

				False, err = p.parseStatements()
				if err != nil {
					return res, err
				}
			}

			res = append(res, &If{Expression: expr, True: stmts, False: False})

		case lexer.INLINE:
			res = append(res, &Inline{Text: p.curToken.Value.(string)})

		case lexer.LET:
			left, err := p.parsePostfix()
			if err != nil {
				return res, err
			}
			eq := p.l.Next()

			if eq.Type != lexer.ASSIGN {
				return res, fmt.Errorf("missing '=' after LET")
			}
			vals, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			res = append(res,
				&Let{Left: left, Expression: vals})

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

		case lexer.IDENT:

			name := p.curToken.Value.(string)

			// index
			if p.l.Peek().Type == lexer.LINDEX {

				// consume '('
				p.l.Next()

				index, err := p.parseExpr()
				if err != nil {
					return nil, err
				}

				if p.l.Next().Type != lexer.RINDEX {
					return nil, fmt.Errorf("missing ]")
				}

				// consume '='
				p.l.Next()

				vals, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				res = append(res,
					&IndexAssign{
						Left:       &VariableExpr{Name: name},
						Index:      index,
						Expression: vals,
					})

			} else if p.l.Peek().Type == lexer.LPAREN {

				// consume '('
				p.l.Next()

				var params []Expr

				for {
					t := p.l.Peek()

					if t.Type == lexer.RPAREN {
						p.l.Next()
						break
					}

					if t.Type == lexer.COMMA {
						p.l.Next()
						continue
					}

					expr, err := p.parseExpr()
					if err != nil {
						return nil, err
					}

					params = append(params, expr)
				}

				res = append(res, &FunctionCallExpr{
					Name:      name,
					Arguments: params,
				})
			} else if p.l.Peek().Type == lexer.ASSIGN {

				// consume '='
				p.l.Next()

				vals, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				res = append(res,
					&Let{Left: &VariableExpr{Name: name}, Expression: vals})

			} else {

				// plain variable
				res = append(res, &VariableExpr{
					Name: name,
				})
			}
		default:
			return res, fmt.Errorf("unknown token type in parseStatements: %v", p.curToken)
		}

		// repeat
		p.curToken = p.l.Next()
	}
	return res, nil
}
