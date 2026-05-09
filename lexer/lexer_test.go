package lexer

import (
	"fmt"
	"strings"
	"testing"
)

// Test basic invocation of our lexer.
func TestLexer(t *testing.T) {

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LET, "let"},
		{PRINT, "print"},
		{PRINTLN, "println"},
		{IF, "if"},
		{COMMA, ","},
		{WHILE, "while"},
		{RETURN, "return"},
		{MULTIPLY, "*"},
		{ASSIGN, "="},
		{NUMBER, "3"},
		{PLUS, "+"},
		{NUMBER, "4"},
		{MULTIPLY, "*"},
		{NUMBER, "5"},
		{MINUS, "-"},
		{NUMBER, "1"},
		{DIVIDE, "/"},
		{NUMBER, "2"},
		{EQUALS, "=="},
		{LT_EQUALS, "<="},
		{GT_EQUALS, ">="},
		{NOT_EQUALS, "!="},
		{LT, "<"},
		{GT, ">"},
		{AND, "&&"},
		{OR, "||"},
		{ERROR, "invalid character '!'"},
		{ERROR, "invalid character '&'"},
		{ERROR, "invalid character '|'"},
		{EOF, ""},
	}

	l := NewLexer("LEt PrinT printLn if, WHILE      retURN * = 3 + 4 * 5 - 1 / 2 == <= >= != < > && || ! & |")

	for i, tt := range tests {
		tok := l.Next()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if fmt.Sprintf("%v", tok.Value) != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Value)
		}
	}

}

// Test we can parse numbers correctly
func TestNumbers(t *testing.T) {

	//
	// We're going to create a number so big that it cannot
	// be parsed by strconv.ParseFloat.
	//
	// Maximum value.
	//
	fmax := 1.7976931348623157e+308

	// Now, as a string.
	fmaxStr := fmt.Sprintf("%f", fmax)

	// Add a prefix to make it too big.
	fmaxStr = "9999" + fmaxStr

	tests := []struct {
		input  string
		error  bool
		errMsg string
	}{
		{"-3", false, ""},
		{".1", false, ""},
		{".1.1", true, "too many"},
		{"$", true, "unknown character"},
		{fmaxStr, true, "failed to parse number"},
	}

	for n, test := range tests {

		l := NewLexer(test.input)

		// Loop over all tokens and see if we found an error
		err := ""

		tok := l.Next()
		for tok.Type != EOF {
			if tok.Type == ERROR {
				err = tok.Value.(string)
			}
			tok = l.Next()

		}

		if test.error {
			if err == "" {
				t.Fatalf("tests[%d] %s - expected error, got none", n, test.input)
			}
			if !strings.Contains(err, test.errMsg) {
				t.Fatalf("expected error to match '%s', but got '%s'", test.errMsg, err)
			}
		} else {
			if err != "" {
				t.Fatalf("tests[%d] %s - didn't expect error, got %s", n, test.input, err)
			}
		}
	}

}

// TestIssue15 confirms https://github.com/skx/sysbox/issues/15 is closed.
func TestIssue15(t *testing.T) {
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LET, "let"},
		{IDENT, "b"},
		{ASSIGN, "="},
		{NUMBER, "1"},
		{SEMICOLON, ";"},
		{LPAREN, "("},
		{IDENT, "b"},
		{MINUS, "-"},
		{IDENT, "b"},
		{RPAREN, ")"},
		{EOF, ""},
	}

	l := NewLexer("LeT b = 1; ( b -b)")

	for i, tt := range tests {
		tok := l.Next()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if fmt.Sprintf("%v", tok.Value) != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Value)
		}
	}
}

func TestNumeric(t *testing.T) {

	lexer := NewLexer("bogus stuff")

	ok := lexer.isNumberComponent('-', true)
	if !ok {
		t.Fatalf("leading '-' wasn't handled")
	}

	ok = lexer.isNumberComponent('-', false)
	if ok {
		t.Fatalf("'-' isn't valid unless at the start of a number")
	}
}

func TestString(t *testing.T) {
	lexer := NewLexer(`"bogus stuff`)
	out := lexer.Next()
	if out.Type != ERROR {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(out.Value.(string), "unterminated") {
		t.Fatalf("got error, but wrong one")
	}

	lexer = NewLexer(`"bogus stuff"`)
	out = lexer.Next()
	if out.Type != STRING {
		t.Fatalf("expected string, got %v", out)
	}
	if !strings.Contains(out.Value.(string), "bogus stuff") {
		t.Fatalf("got value, but wrong one")
	}
}

func TestInline(t *testing.T) {
	lexer := NewLexer(`inline { `)
	out := lexer.Next()
	if out.Type != ERROR {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(out.Value.(string), "unterminated") {
		t.Fatalf("got error, but wrong one")
	}

	lexer = NewLexer(`inline { # test "steve } `)
	out = lexer.Next()
	if out.Type != INLINE {
		t.Fatalf("expected string, got %v", out)
	}
	if !strings.Contains(out.Value.(string), "# test \"steve") {
		t.Fatalf("got value, but wrong one")
	}
}

func TestComment(t *testing.T) {
	lexer := NewLexer(`#bogus stuff`)
	out := lexer.Next()
	if out.Type != EOF {
		t.Fatalf("expected EOF, got none")
	}

	lexer = NewLexer(`#bogus stuff

`)
	out = lexer.Next()
	if out.Type != EOF {
		t.Fatalf("expected EOF, got none")
	}
}
