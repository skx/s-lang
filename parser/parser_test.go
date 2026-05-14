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
		_, ok := stmt.(*Let)
		if !ok {
			t.Fatalf("unexpected statement type")
		}
	}
}

// TestValid runs some programs which should be valid, and confirms no
// errors are returned
func TestValid(t *testing.T) {
	tests := []struct {
		input string
	}{

		{"break; "},
		{"continue; "},

		{"if(1) { return( 1 ) ; } else { return (3); }"},
		{"if(a) { return( 1 ) ; }"},
		{"if(a < b) { return( 1 ) ; }"},
		{"function test() { return ( 1 ); } ; return( test() );"},
		{"function test() { return ( 1 ); } ; test();"},
		{"function test(a, b, c) { return ( a + b + c ); } ; return( test(1, 2, 3) );"},
		{"function test(a, b, c) { return ( a + b + c ); } ; test(1, 2, 3) ;"},
		{"inline{ }"},
		{"data{ }"},

		{"let a = 3; a ;"},
		{"let a = 3 * 3;"},

		{"return(1);"},
		{"return(a);"},
		{"return(1 + 2 * 3);"},

		{"while(1) { print(3) };"},
		{"let a = 10; while(a) { let a = a - 1; println( a ); };"},
	}
	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := New(l)
		_, err := p.ParseProgram()
		if err != nil {
			t.Fatalf("unexpected err parsing program: %s %s", tt.input, err)
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

		{"function foo { return 3; } "},
		{"function test() ; return ( 1 );"},

		{"if ( a ) "},
		{"if ( a * ) { return 1: } "},
		{"if ( a * ) { return 1: } else ; "},
		{"if ( a ) { return \"steve\"; }"},
		{"if ( a  "},
		{"if  a  "},

		{"if ( a  ) { return(1); } else "},

		{"data {"},
		{"inline {"},

		{"let a = ( 3 + 3"},
		{"let a "},

		{"while "},
		{"while ("},
		{"while ( 3 * 3 *"},
		{"while ( 3 * 3  "},
		{"while ( 3 * 3 ) print "},
		{"while ( 3 * 3 ) { return \"steve\"; } "},
	}
	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := New(l)
		_, err := p.ParseProgram()
		if err == nil {
			t.Fatalf("expected err parsing program, but got none: %s", tt.input)
		}
	}
}
