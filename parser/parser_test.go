package parser

import (
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
                p := New(tt.input)
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
                p := New(tt.input)
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
                p := New(tt.input)
                _, err := p.ParseProgram()
                if err == nil {
                        t.Fatalf("expected err parsing program, but got none: %s", tt.input)
                }
        }
}


func TestAdditionalValidStatements(t *testing.T) {
        tests := []struct {
                input string
        }{
                // return without value
                {"return;"},

                // prefix operators
                {"let a = !b;"},
                {"let a = -1;"},
                {"let a = +1;"},

                // float literals
                {"let a = 3.14;"},

                // modulus / power
                {"let a = 10 % 3;"},
                {"let a = 2 ^ 8;"},

                // postfix statements
                {"a++;"},
                {"a--;"},

                // postfix expressions
                {"let a = b++;"},
                {"let a = b--;"},
                {"a = b++;"},
                {"a = b--;"},

                // index assignment
                {"a[0] = 1;"},
                {"a[x + 1] = 3;"},

                // index expression
                {"let a = b[0];"},
                {"let a = b[x + 1];"},

                // function defaults
                {"function foo(a = 1) { return(a); }"},
                {"function foo(a = 1, b = 2) { return(a + b); }"},

                // pragma
                {"pragma foo bar;"},

                // switch
                {"switch a { case 1 { return(1); } }"},
                {"switch a { case 1 { return(1); } default { return(0); } }"},
        }

        for _, tt := range tests {
                p := New(tt.input)
                _, err := p.ParseProgram()
                if err != nil {
                        t.Fatalf("unexpected err parsing program %q: %v", tt.input, err)
                }
        }
}

func TestAdditionalErrors(t *testing.T) {
        tests := []struct {
                input string
        }{
                // pragma validation
                {"pragma foo;"},
                {"pragma;"},

                // assignment
                { "let a = 3 *;" },
                { "a = 3 *;" },
                { "a = foo[ *;" },
                { "a = foo[ 0 ;" },

                // if
                { "if ( true ) { switch a { default { return(0); } default { return(1); } }} "},
                { "if ( true ) { } else { switch a { default { return(0); } default { return(1); } }} "},

                // default parameter ordering
                {"function foo(a = 1, b) { return(a); }"},
                { "foo( foo[ 0  );" },


                // switch validation
                {"switch a {"},
                {"switch a { case 1 return(1); }"},
                {"switch a * { case 1 return(1); }"},
                {"switch a { default { return(0); } default { return(1); } }"},
                {"switch a { while( true ) { break; } }"},

                // functions need idents
                {"function 3() {print(3);}"},
                {"function test(\"non-ident\") {print(3);}"},
                {"foo(3, 4, \"steve\""},
                {"function foo(x) { switch x { case default { return; } default { return; } } }"},

                // EOF on argument-parsing
                {"function 3(a, b"},

                // bare literals are illegal statements
                {"3;"},
                {"3.14;"},
                {"\"hello\";"},

                // malformed index assignment
                {"a[0 = 1;"},
                {"a[0] ;"},
                {"a[ c * ) ] ;"},

                // malformed postfix usage
                {"++a;"},

                // malformed function definitions
                {"function foo(a = ) { return(1); }"},
        }

        for _, tt := range tests {
                p := New(tt.input)
                _, err := p.ParseProgram()
                if err == nil {
                        t.Fatalf("expected error parsing %q", tt.input)
                }
        }
}

func TestSwitchParses(t *testing.T) {
        p := New(`
                switch a {
                        case 1 { return(1); }
                        case 2 { return(2); }
                        default { return(0); }
                }
        `)

        program, err := p.ParseProgram()
        if err != nil {
                t.Fatalf("parse failed: %v", err)
        }

        if len(program.Statements) != 1 {
                t.Fatalf("expected 1 statement, got %d", len(program.Statements))
        }

        if _, ok := program.Statements[0].(*Switch); !ok {
                t.Fatalf("expected Switch statement")
        }
}

func TestPragmaParses(t *testing.T) {
        p := New("pragma optimize speed;")

        program, err := p.ParseProgram()
        if err != nil {
                t.Fatalf("parse failed: %v", err)
        }

        if len(program.Statements) != 1 {
                t.Fatalf("expected 1 statement")
        }

        if _, ok := program.Statements[0].(*Pragma); !ok {
                t.Fatalf("expected Pragma statement")
        }
}
