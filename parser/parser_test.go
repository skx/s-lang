package parser

import (
	"s-lang/lexer"
	"testing"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"let x =5;"},
		{"let x =5 - 1;"},
		{"let x =5 + 2;"},
		{"let x =5 * 2;"},
		{"let x =5 > 2;"},
		{"let x =5 / 2;"},
		{"let x = ( 3 * 3 );"},
		{"let x = a + 3;"},
		{"let x = (a == 3);"},
		{"let x = (a == 3) && 1;"},
		{"let x = (a == 3) || 1;"},
		{"let x = \"steve\";"},
	}
	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := New(l)
		program, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("unexpected error parsing program")
		}
		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}
		stmt := program.Statements[0]
		_, ok := stmt.(*LetStatement)
		if !ok {
			t.Fatalf("unexpected statement type")
		}
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"return( \"steve );"},
		{"return( \"steve "},
		{"return"},
		{"return("},
		{"return(3"},
		{"let a = ( 3 + 3"},
		{"let a "},
	}
	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := New(l)
		_, err := p.ParseProgram()
		if err == nil {
			t.Fatalf("unexpected error parsing program")
		}
	}
}
