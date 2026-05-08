// Package lexer contains a simple lexer, which consumes text from our
// input language and returns a series of tokens.
package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// These constants are used to describe the type of token which has been lexed.
const (
	// Basic token-types
	EOF    = "EOF"
	ERROR  = "ERROR"
	IDENT  = "IDENT"
	NUMBER = "NUMBER"
	STRING = "STRING"

	// Statements
	IF      = "IF"
	LET     = "LET"
	PRINT   = "PRINT"
	PRINTLN = "PRINTLN"
	RETURN  = "RETURN"
	WHILE   = "WHILE"

	// Specials
	ASSIGN    = "="
	COMMA     = ","
	SEMICOLON = ";"

	// Paren
	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// Operations
	PLUS     = "+"
	MINUS    = "-"
	MULTIPLY = "*"
	DIVIDE   = "/"

	// Comparisons
	LT         = "<"
	LT_EQUALS  = "<="
	GT         = ">"
	GT_EQUALS  = ">="
	EQUALS     = "=="
	NOT_EQUALS = "!="
)

// TokenType is the type of our tokens.
type TokenType string

// Token holds a lexed token from our input.
type Token struct {

	// The type of the token.
	Type TokenType

	// The value of the token.
	//
	// If the type of the token is NUMBER then this
	// will be stored as a float64.  Otherwise the
	// value will be a string representation of the token.
	//
	Value interface{}
}

func (t Token) String() string {
	if t.Type == NUMBER {
		return fmt.Sprintf("Token{Type:%s Value:%s}", t.Type, fmt.Sprintf("%d", int64(t.Value.(float64))))
	}
	return fmt.Sprintf("Token{Type:%s Value:%s}", t.Type, t.Value)
}

// Lexer holds our lexer state.
type Lexer struct {

	// input is the string we're lexing.
	input string

	// position is the current position within the input-string.
	position int

	// simple map of single-character tokens to their type
	known map[string]string

	// peek is used for peeking
	peek *Token
}

// NewLexer creates a new lexer, for the given input.
func NewLexer(input string) *Lexer {

	// Create the lexer object.
	l := &Lexer{input: input, peek: nil}

	// Populate the simple token-types in a map for later use.
	//
	// Note that we don't have "=", "<", ">", etc here because they
	// might be part of a multi-character token (i.e. ">=").
	l.known = make(map[string]string)
	l.known["*"] = MULTIPLY
	l.known["+"] = PLUS
	l.known["-"] = MINUS
	l.known["/"] = DIVIDE
	l.known["("] = LPAREN
	l.known[")"] = RPAREN
	l.known["{"] = LBRACE
	l.known["}"] = RBRACE
	l.known[";"] = SEMICOLON
	l.known[","] = COMMA

	return l
}

// Peek returns the upcoming token.
func (l *Lexer) Peek() *Token {

	if l.peek == nil {
		l.peek = l.Next()
	}
	return l.peek
}

func (l *Lexer) peekChar() string {
	val := ""
	if l.position < len(l.input) {

		val = string(l.input[l.position])
	}
	return val
}

// Next returns the next token from our input stream.
//
// This is pretty naive lexer, however it is sufficient to
// recognize numbers, identifiers, and our small set of
// operators.
func (l *Lexer) Next() *Token {

	if l.peek != nil {
		x := l.peek
		l.peek = nil
		return x
	}

	// Loop until we've exhausted our input.
	for l.position < len(l.input) {

		// Get the next character
		char := string(l.input[l.position])

		// Is this a known character/token?
		t, ok := l.known[char]
		if ok {
			// skip the character, and return the token
			l.position++
			return &Token{Value: char, Type: TokenType(t)}
		}

		// If we reach here it is something more complex.
		switch char {

		// Look for the more annoying cases
		case "<":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: LT_EQUALS, Value: "<="}
			}
			return &Token{Type: LT, Value: "<"}
		case ">":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: GT_EQUALS, Value: ">="}
			}
			return &Token{Type: GT, Value: ">"}
		case "!":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: NOT_EQUALS, Value: "!="}
			}
			return &Token{Type: ERROR, Value: "invalid character '!'"}
		case "=":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: EQUALS, Value: "=="}
			}
			return &Token{Type: ASSIGN, Value: "="}

		// Skip whitespace
		case " ", "\n", "\r", "\t", ";":
			l.position++
			continue

		case "#":
			// skip the comment
			l.position++

			// skip everything until the end of the line
			for l.position < len(l.input) {
				c := l.input[l.position]
				if c == '\n' {
					break
				}
				l.position++
			}
			continue

		case "\"":
			// skip the comment
			l.position++

			str := ""

			for l.position < len(l.input) {
				// get the character
				c := l.input[l.position]
				l.position++

				// end of the string?  We're done
				// and we've already bumped ot the next
				// character so all is okay.
				if c == '"' {
					return &Token{Value: str, Type: STRING}

				}
				str += string(c)
			}
			return &Token{Value: "unterminated string", Type: ERROR}

			// Is it a potential number?
		case "-", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ".":

			//
			// Loop for more digits
			//

			// Starting offset of our number
			start := l.position

			// ending offset of our number.
			end := l.position

			// keep walking forward, minding we don't wander
			// out of our input.
			for end < len(l.input) {

				if !l.isNumberComponent(l.input[end], end == start) {
					break
				}
				end++
			}

			l.position = end

			// Here we have the number
			token := l.input[start:end]

			// too many periods?
			bits := strings.Split(token, ".")
			if len(bits) > 2 {
				return &Token{Type: ERROR, Value: fmt.Sprintf("too many periods in '%s'", token)}
			}

			// Convert to float64
			number, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return &Token{Value: fmt.Sprintf("failed to parse number: %s", err.Error()), Type: ERROR}
			}

			return &Token{Value: number, Type: NUMBER}
		}

		//
		// We'll assume we have an identifier at this point.
		//

		// Starting offset of our ident
		start := l.position

		// ending offset of our ident.
		end := l.position

		// keep walking forward, minding we don't wander
		// out of our input.
		for end < len(l.input) {

			// Build up identifiers from any permitted
			// character.
			//
			// We allow unicode "letters" only.
			if l.isIdentifierCharacter(l.input[end]) {
				end++
			} else {
				break
			}
		}

		// Change the position to be after the end of the identifier
		// we found - if we didn't find one then that results in no
		// change.
		l.position = end

		// Now record the text of the token (i.e. identifier).
		token := l.input[start:end]

		//
		// In a real language/lexer we might have a lot of keywords/reserved-words to handle.
		//
		// We only need handle a few.
		//
		if strings.ToLower(token) == "if" {
			return &Token{Value: "if", Type: IF}
		}
		if strings.ToLower(token) == "let" {
			return &Token{Value: "let", Type: LET}
		}
		if strings.ToLower(token) == "return" {
			return &Token{Value: "return", Type: RETURN}
		}
		if strings.ToLower(token) == "print" {
			return &Token{Value: "print", Type: PRINT}
		}
		if strings.ToLower(token) == "println" {
			return &Token{Value: "println", Type: PRINTLN}
		}
		if strings.ToLower(token) == "while" {
			return &Token{Value: "while", Type: WHILE}
		}

		//
		// So we handled the easy cases, and then defaulted
		// to looking for our only supported identifier.
		//
		// If we failed to find one that means that we've got
		// to skip the unknown character - to avoid an infinite
		// loop.
		//
		// We'll skip over the character, and return the error.
		//
		if token == "" {
			l.position++
			return &Token{Value: fmt.Sprintf("unknown character %c", l.input[end]), Type: ERROR}
		}

		//
		// We found a non-empty identifier, which
		// wasn't converted into a `let` keyword.
		//
		// Return it.
		//
		return &Token{Value: token, Type: IDENT}

	}

	//
	// If we get here then we've walked past the end of
	// our input-string.
	//
	return &Token{Value: "", Type: EOF}
}

// isIdentifierCharacter tests whether the given character is
// valid for use in an identifier.
func (l *Lexer) isIdentifierCharacter(d byte) bool {

	return (unicode.IsLetter(rune(d)))
}

// isNumberComponent looks for characters that can make up integers/floats
//
// We handle the first-character specially, which is why that's an argument
func (l *Lexer) isNumberComponent(d byte, first bool) bool {

	// digits
	if unicode.IsDigit(rune(d)) {
		return true
	}

	// floating-point numbers require the use of "."
	if d == '.' {
		return true
	}

	// negative sign can only occur at the start of the input
	if d == '-' && first {
		return true
	}

	// No, this is not part of a number.
	return false
}
