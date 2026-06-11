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
	EOF     = "EOF"
	ERROR   = "ERROR"
	IDENT   = "IDENT"
	INTEGER = "INTEGER"
	FLOAT   = "FLOAT"
	STRING  = "STRING"

	// Statements
	BREAK    = "BREAK"
	CASE     = "CASE"
	CONTINUE = "CONTINUE"
	DATA     = "DATA"
	DEFAULT  = "DEFAULT"
	ELSE     = "ELSE"
	FUNCTION = "FUNCTION"
	IF       = "IF"
	INLINE   = "INLINE"
	LET      = "LET"
	PRAGMA   = "PRAGMA"
	RETURN   = "RETURN"
	SWITCH   = "SWITCH"
	WHILE    = "WHILE"

	// Specials
	ASSIGN    = "="
	COMMA     = ","
	SEMICOLON = ";"

	// Paren
	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	LINDEX = "["
	RINDEX = "]"

	// Operations
	DIVIDE   = "/"
	EXCLAIM  = "!"
	MINUS    = "-"
	MODULUS  = "%"
	MULTIPLY = "*"
	PLUS     = "+"
	POWER    = "^"

	// Comparisons
	AND       = "&&"
	OR        = "||"
	LT        = "<"
	LTEQUALS  = "<="
	GT        = ">"
	GTEQUALS  = ">="
	EQUALS    = "=="
	NOTEQUALS = "!="

	// postfix
	PLUSPLUS   = "++"
	MINUSMINUS = "--"
)

// TokenType is the type of our tokens.
type TokenType string

// Token holds a lexed token from our input.
type Token struct {

	// Line is the source-line upon which the token appears
	Line int

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

// String returns a human-readable version of our token.
func (t Token) String() string {
	if t.Type == INTEGER {
		return fmt.Sprintf("Token{Type:%s Value:%s}", t.Type, fmt.Sprintf("%d", int64(t.Value.(float64))))
	}
	if t.Type == FLOAT {
		return fmt.Sprintf("Token{Type:%s Value:%s}", t.Type, fmt.Sprintf("%f", t.Value.(float64)))
	}
	return fmt.Sprintf("Token{Type:%s Value:%s}", t.Type, t.Value)
}

// Lexer holds our lexer state.
type Lexer struct {

	// input is the string we're lexing.
	input string

	// line is our current line-number.
	line int

	// position is the current position within the input-string.
	position int

	// simple map of single-character tokens to their type
	known map[string]string

	// keywords holds known keywords
	keywords map[string]bool

	// peek is used for peeking
	peek *Token
}

// NewLexer creates a new lexer, for the given input.
func NewLexer(input string) *Lexer {

	// Create the lexer object.
	l := &Lexer{input: input, peek: nil, line: 1}

	// Populate the simple token-types in a map for later use.
	//
	// Note that we don't have "=", "<", ">", etc here because they
	// might be part of a multi-character token (i.e. ">=").
	//
	// We also don't have "-" because we need to parse numbers and that
	// might be present as the leading character.
	//
	l.known = make(map[string]string)
	l.known["%"] = MODULUS
	l.known["("] = LPAREN
	l.known[")"] = RPAREN
	l.known["*"] = MULTIPLY
	l.known[","] = COMMA
	l.known["/"] = DIVIDE
	l.known[";"] = SEMICOLON
	l.known["["] = LINDEX
	l.known["]"] = RINDEX
	l.known["^"] = POWER
	l.known["{"] = LBRACE
	l.known["}"] = RBRACE

	l.keywords = make(map[string]bool)
	l.keywords["break"] = true
	l.keywords["case"] = true
	l.keywords["continue"] = true
	l.keywords["default"] = true
	l.keywords["else"] = true
	l.keywords["function"] = true
	l.keywords["if"] = true
	l.keywords["let"] = true
	l.keywords["pragma"] = true
	l.keywords["return"] = true
	l.keywords["switch"] = true
	l.keywords["while"] = true
	return l
}

// Peek returns the upcoming token.
func (l *Lexer) Peek() *Token {

	if l.peek == nil {
		l.peek = l.Next()
	}
	return l.peek
}

// peekChar looks one character forward within our input.
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

		// bump the line count
		if char == "\n" {
			l.line++
		}

		// Is this a known character/token?
		t, ok := l.known[char]
		if ok {
			// skip the character, and return the token
			l.position++
			return &Token{Value: char, Type: TokenType(t), Line: l.line}
		}

		// If we reach here it is something more complex.
		switch char {

		// Look for the more annoying cases
		case "'":
			// skip over the character literal
			l.position++

			var chr rune

			// get the first character
			if l.position < len(l.input) {
				chr = rune(l.input[l.position])
				l.position++

				// If it's an escape read the next character
				switch chr {
				case '\\':
					if l.position < len(l.input) {
						c := l.input[l.position]
						l.position++
						switch c {
						case 'n':
							chr = '\n'
						case 'r':
							chr = '\r'
						case 't':
							chr = '\t'
						case '\\':
							chr = '\\'
						default:
							return &Token{Type: ERROR, Value: fmt.Sprintf("unrecognized escape character in character literal %c", c), Line: l.line}
						}
					} else {
						return &Token{Type: ERROR, Value: fmt.Sprintf("unterminated character literal, got %s", l.peekChar()), Line: l.line}
					}

				default:
					break
				}
			} else {
				return &Token{Type: ERROR, Value: fmt.Sprintf("unterminated character literal, got %s", l.peekChar()), Line: l.line}
			}

			if l.peekChar() != "'" {
				return &Token{Type: ERROR, Value: fmt.Sprintf("unterminated character literal, got %s", l.peekChar()), Line: l.line}
			}
			l.position++
			return &Token{Value: float64(chr), Type: INTEGER, Line: l.line}

		case "+":
			l.position++
			if l.peekChar() == "+" {
				l.position++
				return &Token{Type: PLUSPLUS, Value: "++", Line: l.line}
			}
			return &Token{Type: PLUS, Value: "+", Line: l.line}

		case "<":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: LTEQUALS, Value: "<=", Line: l.line}
			}
			return &Token{Type: LT, Value: "<", Line: l.line}
		case ">":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: GTEQUALS, Value: ">=", Line: l.line}
			}
			return &Token{Type: GT, Value: ">", Line: l.line}
		case "!":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: NOTEQUALS, Value: "!=", Line: l.line}
			}
			return &Token{Type: EXCLAIM, Value: "!", Line: l.line}
		case "&":
			l.position++
			if l.peekChar() == "&" {
				l.position++
				return &Token{Type: AND, Value: "&&", Line: l.line}
			}
			return &Token{Type: ERROR, Value: "invalid character '&'", Line: l.line}
		case "|":
			l.position++
			if l.peekChar() == "|" {
				l.position++
				return &Token{Type: OR, Value: "||", Line: l.line}
			}
			return &Token{Type: ERROR, Value: "invalid character '|'", Line: l.line}

		case "=":
			l.position++
			if l.peekChar() == "=" {
				l.position++
				return &Token{Type: EQUALS, Value: "==", Line: l.line}
			}
			return &Token{Type: ASSIGN, Value: "=", Line: l.line}

		case "-":
			l.position++
			if l.peekChar() == "-" {
				l.position++
				return &Token{Type: MINUSMINUS, Value: "--", Line: l.line}
			}
			return &Token{Value: "-", Type: MINUS, Line: l.line}

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
			// skip the opening quote
			l.position++

			str := ""

			for l.position < len(l.input) {
				c := l.input[l.position]
				l.position++
				if c == '\n' {
					l.line++
				}

				// Handle escapes
				if c == '\\' {
					// Unterminated escape
					if l.position >= len(l.input) {
						return &Token{Value: "unterminated escape", Type: ERROR, Line: l.line}
					}

					next := l.input[l.position]
					l.position++

					switch next {
					case '"':
						str += `"`
					case '\\':
						str += `\`
					case 'n':
						str += "\n"
					case 'r':
						str += "\r"
					case 't':
						str += "\t"
					default:
						// Unknown escape: keep literally
						str += string(next)
					}

					continue
				}

				// End of string
				if c == '"' {
					return &Token{Value: str, Type: STRING, Line: l.line}
				}

				str += string(c)
			}

			return &Token{Value: "unterminated string", Type: ERROR, Line: l.line}

		// Is it a potential number?
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ".":

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

				if !l.isNumberComponent(l.input[end]) {
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
				return &Token{Type: ERROR, Value: fmt.Sprintf("too many periods in '%s'", token), Line: l.line}
			}

			// Convert to float64
			number, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return &Token{Value: fmt.Sprintf("failed to parse number: %s", err.Error()), Type: ERROR, Line: l.line}
			}

			// Is this an int?  Or a float?
			if strings.Contains(token, ".") {
				return &Token{Value: number, Type: FLOAT, Line: l.line}
			}
			return &Token{Value: number, Type: INTEGER, Line: l.line}
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
		if strings.ToLower(token) == "inline" || strings.ToLower(token) == "data" {
			// We want to be able to handle input of
			// the form :
			//
			//   inline {
			//     blah
			//   }
			//
			// If we return a token of "inline" then
			// rely on the rest of the tokens to get the
			// body things will be broken, because
			// we'll swallow whitespace, etc, etc.
			//
			// So we have to return it here.
			txt := ""

			// skip over the opening "{"
			for l.position < len(l.input) {
				c := l.input[l.position]
				l.position++

				if c == '\n' {
					l.line++
				}

				if c == '{' {
					break
				}
			}

			// Build up all characters
			for l.position < len(l.input) {
				// get the character
				c := l.input[l.position]
				l.position++

				if c == '\n' {
					l.line++
				}

				// end of the string?  We're done
				// and we've already bumped ot the next
				// character so all is okay.
				if c == '}' {
					if strings.ToLower(token) == "inline" {
						return &Token{Value: txt, Type: INLINE, Line: l.line}
					}
					if strings.ToLower(token) == "data" {
						return &Token{Value: txt, Type: DATA, Line: l.line}
					}
				}
				txt += string(c)
			}
			return &Token{Value: fmt.Sprintf("unterminated %s", strings.ToLower(token)), Type: ERROR, Line: l.line}
		}

		// Special case true/false.  We could handle
		// them as keywords, but this approach feels
		// fine.
		if strings.ToLower(token) == "true" {
			return &Token{Value: float64(1), Type: INTEGER, Line: l.line}
		}
		if strings.ToLower(token) == "false" {
			return &Token{Value: float64(0), Type: INTEGER, Line: l.line}
		}

		//
		// Should we convert the token from an IDENT into a known
		// keyword?  If so do it.
		//
		_, ok = l.keywords[strings.ToLower(token)]
		if ok {
			return &Token{Value: strings.ToLower(token), Type: TokenType(strings.ToUpper(token)), Line: l.line}
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
			return &Token{Value: fmt.Sprintf("unknown character %c", l.input[end]), Type: ERROR, Line: l.line}
		}

		return &Token{Value: token, Type: IDENT, Line: l.line}

	}

	//
	// If we get here then we've walked past the end of
	// our input-string.
	//
	return &Token{Value: "", Type: EOF, Line: l.line}
}

// isIdentifierCharacter tests whether the given character is
// valid for use in an identifier.
func (l *Lexer) isIdentifierCharacter(d byte) bool {

	// letters
	if unicode.IsLetter(rune(d)) {
		return true
	}

	// digits
	if unicode.IsDigit(rune(d)) {
		return true
	}

	// underscore is useful
	if d == '_' {
		return true
	}

	return false
}

// isNumberComponent looks for characters that can make up integers/floats.
func (l *Lexer) isNumberComponent(d byte) bool {

	// digits
	if unicode.IsDigit(rune(d)) {
		return true
	}

	// floating-point numbers require the use of "."
	if d == '.' {
		return true
	}

	// No, this is not part of a number.
	return false
}
