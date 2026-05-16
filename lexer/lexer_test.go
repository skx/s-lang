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
		{IF, "if"},
		{COMMA, ","},
		{WHILE, "while"},
		{RETURN, "return"},
		{MULTIPLY, "*"},
		{ASSIGN, "="},
		{INTEGER, "3"},
		{PLUS, "+"},
		{INTEGER, "4"},
		{MULTIPLY, "*"},
		{INTEGER, "5"},
		{MINUS, "-"},
		{INTEGER, "1"},
		{DIVIDE, "/"},
		{INTEGER, "2"},
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

	l := NewLexer("LEt if, WHILE      retURN * = 3 + 4 * 5 - 1 / 2 == <= >= != < > && || ! & |")

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
		{INTEGER, "1"},
		{SEMICOLON, ";"},

		{LPAREN, "("},
		{IDENT, "b"},
		{MINUS, "-"},
		{IDENT, "b"},
		{RPAREN, ")"},
		{SEMICOLON, ";"},

		{LET, "let"},
		{IDENT, "C"},
		{ASSIGN, "="},
		{FLOAT, "3.1"},
		{SEMICOLON, ";"},

		{EOF, ""},
	}

	l := NewLexer("LeT b = 1; ( b -b); LET C = 3.1;")

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

func TestData(t *testing.T) {
	lexer := NewLexer(`data { `)
	out := lexer.Next()
	if out.Type != ERROR {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(out.Value.(string), "unterminated") {
		t.Fatalf("got error, but wrong one")
	}

	lexer = NewLexer(`data { # test "steve } `)
	out = lexer.Next()
	if out.Type != DATA {
		t.Fatalf("expected string, got %v", out)
	}
	if !strings.Contains(out.Value.(string), "# test \"steve") {
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

func TestPeek(t *testing.T) {
	lexer := NewLexer(`let a = 3;`)

	first := lexer.Peek()
	if first.Type != LET {
		t.Fatalf("peek gave the wrong result")
	}

	// Peek should always match what Next returns
	i := 0
	for i < 20 {
		peek := lexer.Peek()
		real := lexer.Next()
		if peek.Type != real.Type {
			t.Fatalf("type mismatch")
		}
		if peek.Value != real.Value {
			t.Fatalf("value mismatch")
		}
		if peek.String() != real.String() {
			t.Fatalf("String() mismatch")
		}
		i++
	}
}

// TestIndex tests a simple string index
func TestIndex(t *testing.T) {
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LET, "let"},
		{IDENT, "b"},
		{ASSIGN, "="},
		{STRING, "steve"},
		{LINDEX, "["},
		{INTEGER, "3"},
		{RINDEX, "]"},
		{SEMICOLON, ";"},

		{EOF, ""},
	}

	l := NewLexer("LeT b = \"steve\"[3];")

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
