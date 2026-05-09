package main

import (
	"fmt"
	"os"

	"github.com/skx/subcommands"
	"s-lang/lexer"
	"s-lang/parser"
)

// Structure for our options and state.
type parseCommand struct {
	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info returns the name of this subcommand.
func (p *parseCommand) Info() (string, string) {
	return "parse", `Show the parser output for the given program.

Details:

This command allows you to see what the parser produces for the
given input file.

Example:

    $ s-lang parse example.in
`
}

func (p *parseCommand) printStmt(st parser.Statement) {
	switch stmt := st.(type) {
	case *parser.Function:
		fmt.Printf("FUNCTION Definition %s();\n", stmt.Name)
	case *parser.Let:
		fmt.Printf("LET %s = %v;\n", stmt.Name, stmt.Expression)
	case *parser.Inline:
		fmt.Printf("INLINE {%s}", stmt.Text)
	case *parser.Print:
		if stmt.NewLine {
			fmt.Printf("PRINTLN(")
		} else {
			fmt.Printf("PRINT(")
		}
		for _, x := range stmt.Values {
			fmt.Printf("%v, ", x)
		}
		fmt.Printf(")\n")

	case *parser.Return:
		fmt.Printf("RETURN(%v)\n", stmt.Expression)
	case *parser.While:
		fmt.Printf("while(%v) {\n", stmt.Expression)
		for _, x := range stmt.Statements {
			fmt.Printf("\t")
			p.printStmt(x)
		}
		fmt.Printf("}\n")
	case *parser.If:
		fmt.Printf("if(%V) { \n", stmt.Expression)
		for _, x := range stmt.Statements {
			fmt.Printf("\t")
			p.printStmt(x)
		}
		fmt.Printf("}\n")
	default:
		fmt.Printf("Uknown %T %v\n", stmt, stmt)
	}
}

func (p *parseCommand) parseFile(path string) error {

	// Read the file-contents
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lex := lexer.NewLexer(string(data))

	parse := parser.New(lex)
	program, err := parse.ParseProgram()
	if err != nil {
		return err
	}

	for _, stmt := range program.Statements {
		p.printStmt(stmt)
	}
	return nil
}

// Execute is invoked if the user specifies `lex` as the subcommand.
func (p *parseCommand) Execute(args []string) int {

	// For each argument ..
	for _, arg := range args {

		err := p.parseFile(arg)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return 1
		}
	}

	return 0
}
