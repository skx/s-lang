package main

import (
	"fmt"
	"os"
	"s-lang/lexer"
	"s-lang/parser"
	"strings"

	"github.com/skx/subcommands"
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

// printStmt prints out a single statement, but might call itself
// recursively to handle while-loops and conditionals.
func (p *parseCommand) printStmt(st parser.Statement) error {
	switch stmt := st.(type) {
	case *parser.Break:
		fmt.Fprintf(output, "\tBREAK;\n")
	case *parser.Continue:
		fmt.Fprintf(output, "\tCONTINUE;\n")
	case *parser.Function:
		fmt.Fprintf(output, "FUNCTION Definition %s(%v);\n", stmt.Name, stmt.Parameters)
	case *parser.FunctionCallExpr:
		s := []string{}
		for _, a := range stmt.Arguments {
			s = append(s, a.String())
		}
		fmt.Fprintf(output, "%s(%s);\n", stmt.Name, strings.Join(s, ","))
	case *parser.IntegerLiteral:
		fmt.Fprintf(output, "Integer Literal %d;\n", stmt.Value)
	case *parser.IndexAssign:
		fmt.Fprintf(output, "%s[%s] = %s\n", stmt.Left, stmt.Index, stmt.Expression)
	case *parser.FloatLiteral:
		fmt.Fprintf(output, "Float Literal %f;\n", stmt.Value)
	case *parser.Let:
		fmt.Fprintf(output, "LET %s = %v;\n", stmt.Left.String(), stmt.Expression)
	case *parser.Inline:
		fmt.Fprintf(output, "INLINE {%s}", stmt.Text)
	case *parser.Return:
		fmt.Fprintf(output, "RETURN(%v)\n", stmt.Expression)
	case *parser.StringLiteral:
		fmt.Fprintf(output, "String Literal %s;\n", stmt.Value)
	case *parser.VariableExpr:
		fmt.Fprintf(output, "%s\n", stmt.Name)
	case *parser.While:
		fmt.Fprintf(output, "while(%v) {\n", stmt.Expression)
		for _, x := range stmt.Statements {
			fmt.Fprintf(output, "\t")
			err := p.printStmt(x)
			if err != nil {
				return err
			}
		}
		fmt.Fprintf(output, "}\n")
	case *parser.If:
		fmt.Fprintf(output, "if(%V) { \n", stmt.Expression)
		for _, x := range stmt.True {
			fmt.Fprintf(output, "\t")
			err := p.printStmt(x)
			if err != nil {
				return err
			}
		}
		fmt.Fprintf(output, "}\n")

		if len(stmt.False) > 0 {
			fmt.Fprintf(output, "else {\n")
			for _, x := range stmt.False {
				fmt.Fprintf(output, "\t")
				err := p.printStmt(x)
				if err != nil {
					return err
				}
			}
			fmt.Fprintf(output, "}\n")
		}

	default:
		return fmt.Errorf("unknown item at printStmt at %T %v", stmt, stmt)
	}
	return nil
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
		err := p.printStmt(stmt)
		if err != nil {
			return err
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
