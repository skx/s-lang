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

    $ sysbox parse example.in
`
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
		switch stmt := stmt.(type) {
		case *parser.LetStatement:
			if len(stmt.Expression) == 1 {
				fmt.Printf("LET %s = %v;\n", stmt.Name, stmt.Expression[0])
			} else {
				fmt.Printf("LET %s = %v %v %v;\n", stmt.Name, stmt.Expression[0], stmt.Expression[1], stmt.Expression[2])
			}
		case *parser.Print:
			if stmt.NewLine {
				fmt.Printf("PRINTLN(%v)\n", stmt.Values)
			} else {
				fmt.Printf("PRINT(%v)\n", stmt.Values)
			}
		case *parser.Return:
			fmt.Printf("RETURN(%s)\n", stmt.Value)
		case *parser.While:
			fmt.Printf("while(%s) { .. }\n", stmt.Value)
		default:
			fmt.Printf("Uknown %T %v\n", stmt, stmt)
		}
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
