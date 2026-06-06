// Package parser parses input programs, via our lexer, and allows AST nodes
// to be generated.
//
// Our compiler will walk the generated AST nodes to generate an assembly
// language representation of the givne input function.
//
// Our parser is a simple Pratt-based implementation.
package parser

import (
	"fmt"
	"github.com/skx/s-lang/lexer"
)

const (
	PREC_LOWEST = iota

	PREC_OR
	PREC_AND

	PREC_EQUALITY
	PREC_COMPARISON

	PREC_ADD

	PREC_MUL

	PREC_POWER

	PREC_POSTFIX
)

// precedence is the function that determines the precedence of the given
// token-type, and is the key to the Pratt parser.
func precedence(t lexer.TokenType) int {
	switch t {

	case lexer.OR:
		return PREC_OR

	case lexer.AND:
		return PREC_AND

	case lexer.EQUALS, lexer.NOTEQUALS:
		return PREC_EQUALITY

	case lexer.LT,
		lexer.LTEQUALS,
		lexer.GT,
		lexer.GTEQUALS:
		return PREC_COMPARISON

	case lexer.PLUS,
		lexer.MINUS:
		return PREC_ADD

	case lexer.MULTIPLY,
		lexer.DIVIDE,
		lexer.MODULUS:
		return PREC_MUL

	case lexer.POWER:
		return PREC_POWER

	case lexer.LPAREN,
		lexer.LINDEX,
		lexer.PLUSPLUS,
		lexer.MINUSMINUS:
		return PREC_POSTFIX
	}

	return PREC_LOWEST
}

// isRightAssociative is a helper to determine if the given token-type
// is right-associative.
func isRightAssociative(t lexer.TokenType) bool {
	switch t {
	case lexer.POWER:
		return true
	}

	return false
}

// Parser object
type Parser struct {
	// l is our lexer
	l *lexer.Lexer

	// curToken holds the current token from our lexer.
	curToken *lexer.Token
}

// New returns our new parser-object.
func New(program string) *Parser {

	return &Parser{l: lexer.NewLexer(program)}
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
	return p.parsePratt(0)
}

// parsePratt implements the core of our parsing algorithm.
func (p *Parser) parsePratt(minPrec int) (Expr, error) {

	left, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	for {

		tok := p.l.Peek()
		prec := precedence(tok.Type)

		if prec <= minPrec {
			break
		}

		switch tok.Type {

		//
		// postfix operators
		//

		// [
		case lexer.LINDEX:

			p.l.Next()

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

		// (
		case lexer.LPAREN:

			p.l.Next()

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

			v, ok := left.(*VariableExpr)
			if !ok {
				return nil, fmt.Errorf("cannot call non-function")
			}

			left = &FunctionCallExpr{
				Name:      v.Name,
				Arguments: args,
			}

		// x++
		case lexer.PLUSPLUS:

			p.l.Next()

			left = &PostfixExpr{
				Expr: left,
				Op:   lexer.PLUSPLUS,
			}

		// x--
		case lexer.MINUSMINUS:

			p.l.Next()

			left = &PostfixExpr{
				Expr: left,
				Op:   lexer.MINUSMINUS,
			}

		//
		// infix operators
		//
		default:

			p.l.Next()

			rhsPrec := prec

			if !isRightAssociative(tok.Type) {
				rhsPrec = prec
			} else {
				rhsPrec = prec - 1
			}

			right, err := p.parsePratt(rhsPrec)
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

	return left, nil
}

// parseAtom parses literals - or grouped expressions.
func (p *Parser) parseAtom() (Expr, error) {
	tok := p.l.Next()

	switch tok.Type {

	// float
	case lexer.FLOAT:
		return &FloatLiteral{
			Value: tok.Value.(float64),
		}, nil

	// variable
	case lexer.IDENT:
		return &VariableExpr{
			Name: tok.Value.(string),
		}, nil

	// integer
	case lexer.INTEGER:
		return &IntegerLiteral{
			Value: int64(tok.Value.(float64)),
		}, nil

	// ( ... )
	case lexer.LPAREN:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		if p.l.Next().Type != lexer.RPAREN {
			return nil, fmt.Errorf("missing )")
		}

		return expr, nil

	// "string"
	case lexer.STRING:
		return &StringLiteral{
			Value: tok.Value.(string),
		}, nil

	}

	return nil, fmt.Errorf("unexpected token in parseAtom %v", tok)
}

// parseStatements is used to parse a collection of statements.
//
// It is extracted into a function because our WHILE and IF blocks
// will themselves contain statements.
func (p *Parser) parseStatements() ([]Statement, error) {
	res := []Statement{}

	// Setup the token
	p.curToken = p.l.Next()

	// continue until we hit the end of the file, or the block.
	// (blocks are parsed via recursion for things like "while" and "if")
	for p.curToken.Type != lexer.RBRACE && p.curToken.Type != lexer.EOF {

		switch p.curToken.Type {

		case lexer.BREAK:
			res = append(res, &Break{})

		case lexer.CONTINUE:
			res = append(res, &Continue{})

		case lexer.DATA:
			res = append(res, &Data{Text: p.curToken.Value.(string)})

		case lexer.FLOAT:
			return res, fmt.Errorf("bare literal is illegal: %s", p.curToken.String())

		case lexer.FUNCTION:
			name := p.l.Next()
			if name.Type != lexer.IDENT {
				return res, fmt.Errorf("function names must be identifiers")
			}
			start := p.l.Next()
			if start.Type != lexer.LPAREN {
				return res, fmt.Errorf("missing '(' after function name %s", name)
			}

			// If we've found parameters with default values then
			// all remaining parameters must have them.
			//
			// i.e. this is fine
			//    function foo( name = "Steve", address = "Secret")
			//
			// But this is not
			//
			//    function bar( name = "steve", age )
			//
			// Record here if we found a default
			defValue := false

			// collect parameters
			params := []*FunctionParameter{}

			for {
				tok := p.l.Next()
				if tok.Type == lexer.EOF {
					return res, fmt.Errorf("unexpected EOF in function definition")
				}

				// )?  Then we're at the end
				if tok.Type == lexer.RPAREN {
					break
				}
				// skip the comments
				if tok.Type == lexer.COMMA {
					continue
				}
				if tok.Type == lexer.IDENT {
					name := tok.Value.(string)

					// If we see "=" then we're looking at a default parameter
					if p.l.Peek().Type == lexer.ASSIGN {

						// We've seen a default value
						defValue = true

						p.l.Next()

						// Save the default value.
						val, err := p.parseExpr()
						if err != nil {
							return res, err
						}
						params = append(params, &FunctionParameter{
							Name:    name,
							Default: val})
						continue
					}
					if defValue {
						return nil, fmt.Errorf("function %s has parameter without default value after previously seen a default", name)

					}
					params = append(params, &FunctionParameter{
						Name: name})

				} else {
					return res, fmt.Errorf("function arguments must be identifiers")
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

		case lexer.IDENT:

			name := p.curToken.Value.(string)

			if p.l.Peek().Type == lexer.PLUSPLUS {
				// consume '++'
				p.l.Next()

				expr := &PostfixExpr{
					Expr: &VariableExpr{Name: name},
					Op:   lexer.PLUSPLUS,
				}
				res = append(res, expr)
			} else if p.l.Peek().Type == lexer.MINUSMINUS {

				// consume '--'
				p.l.Next()

				expr := &PostfixExpr{
					Expr: &VariableExpr{Name: name},
					Op:   lexer.MINUSMINUS,
				}
				res = append(res, expr)
			} else if p.l.Peek().Type == lexer.LINDEX {

				// index
				// consume '['
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

					if t.Type == lexer.EOF {
						return res, fmt.Errorf("unexpected EOF in function call")
					}

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

		case lexer.INTEGER:
			return res, fmt.Errorf("bare literal is illegal: %s", p.curToken.String())

		case lexer.LET:
			left, err := p.parseExpr()
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

		case lexer.PRAGMA:
			k := p.l.Next()
			if k.Type != lexer.IDENT {
				return res, fmt.Errorf("pragma key must be an ident")
			}
			v := p.l.Next()
			if v.Type != lexer.IDENT {
				return res, fmt.Errorf("pragma value must be an ident")
			}
			res = append(res, &Pragma{
				Key:   k.Value.(string),
				Value: v.Value.(string),
			})
		case lexer.RETURN:
			var expr Expr
			var err error
			start := p.l.Next()
			if start.Type == lexer.SEMICOLON {
				// "return;" with no value
			} else {
				if start.Type != lexer.LPAREN {
					return res, fmt.Errorf("missing '(' after return")
				}
				expr, err = p.parseExpr()
				if err != nil {
					return nil, err
				}

				end := p.l.Next()
				if end.Type != lexer.RPAREN {
					return res, fmt.Errorf("missing ')' after return value")
				}
			}
			res = append(res, &Return{Expression: expr})

		case lexer.SEMICOLON:
			// NOP

		case lexer.SWITCH:

			// look for the expression
			expr, err := p.parseExpr()
			if err != nil {
				return res, err
			}

			// switch statement
			swtch := &Switch{Value: expr}

			start := p.l.Next()
			if start.Type != lexer.LBRACE {
				return res, fmt.Errorf("missing '{' after switch")
			}

			// Process the block which we think will contain
			// various case-statements
			for {
				tok := p.l.Next()
				if tok.Type == lexer.EOF {
					return res, fmt.Errorf("unexpected EOF in switch statement")
				}
				if tok.Type == lexer.RBRACE {
					break
				}

				tmp := &Case{}

				// Default will be handled specially
				if tok.Type == lexer.DEFAULT {

					// We have a default-case here.
					tmp.Default = true

				} else if tok.Type == lexer.CASE {

					// Here we allow "case default" even though
					// most people would prefer to write "default".
					if tok.Type == lexer.DEFAULT {
						tmp.Default = true
					} else {

						// parse the match-expression.
						left, err := p.parseExpr()
						if err != nil {
							return res, err
						}

						tmp.Expression = left
					}
				} else {
					// error - unexpected token
					return res, fmt.Errorf("expected case|default, got %s", tok)
				}

				tok = p.l.Next()
				if tok.Type != lexer.LBRACE {
					return res, fmt.Errorf("missing '{' after case")
				}

				// parse the block
				stmts, err := p.parseStatements()
				if err != nil {
					return res, err
				}
				tmp.Statements = stmts

				// save the choice away
				swtch.Choices = append(swtch.Choices, tmp)

			}

			// More than one default is a bug
			count := 0
			for _, c := range swtch.Choices {
				if c.Default {
					count++
				}
			}
			if count > 1 {
				return res, fmt.Errorf("A switch-statement should only have one default block")
			}

			res = append(res, swtch)
		case lexer.STRING:
			return res, fmt.Errorf("bare literal is illegal: %s", p.curToken.String())

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

		default:
			return res, fmt.Errorf("unknown token type in parseStatements: %v", p.curToken)
		}

		// repeat
		p.curToken = p.l.Next()
	}
	return res, nil
}
