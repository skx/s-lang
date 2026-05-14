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

// printStmt prints out a single statement, but might call itself
// recursively to handle while-loops and conditionals.
func (p *parseCommand) printStmt(st parser.Statement) error {
	switch stmt := st.(type) {
	case *parser.Break:
		_, err := fmt.Fprintf(output, "\tBREAK;\n")
		if err != nil {
			return err
		}

	case *parser.Continue:
		_, err := fmt.Fprintf(output, "\tCONTINUE;\n")
		if err != nil {
			return err
		}

	case *parser.Function:
		_, err := fmt.Fprintf(output, "FUNCTION Definition %s(%v);\n", stmt.Name, stmt.Parameters)
		if err != nil {
			return err
		}

	case *parser.Let:
		_, err := fmt.Fprintf(output, "LET %s = %v;\n", stmt.Name, stmt.Expression)
		if err != nil {
			return err
		}

	case *parser.Inline:
		_, err := fmt.Fprintf(output, "INLINE {%s}", stmt.Text)
		if err != nil {
			return err
		}
	case *parser.Print:
		if stmt.NewLine {
			_, err := fmt.Fprintf(output, "PRINTLN(")
			if err != nil {
				return err
			}

		} else {
			_, err := fmt.Fprintf(output, "PRINT(")
			if err != nil {
				return err
			}

		}
		for _, x := range stmt.Values {
			_, err := fmt.Fprintf(output, "%v, ", x)
			if err != nil {
				return err
			}

		}
		_, err := fmt.Fprintf(output, ")\n")
		if err != nil {
			return err
		}

	case *parser.Return:
		_, err := fmt.Fprintf(output, "RETURN(%v)\n", stmt.Expression)
		if err != nil {
			return err
		}

	case *parser.While:
		_, err := fmt.Fprintf(output, "while(%v) {\n", stmt.Expression)
		if err != nil {
			return err
		}
		for _, x := range stmt.Statements {
			_, err := fmt.Fprintf(output, "\t")
			if err != nil {
				return err
			}
			err = p.printStmt(x)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(output, "}\n")
		if err != nil {
			return err
		}

	case *parser.If:
		_, err := fmt.Fprintf(output, "if(%V) { \n", stmt.Expression)
		if err != nil {
			return err
		}
		for _, x := range stmt.True {
			_, err = fmt.Fprintf(output, "\t")
			if err != nil {
				return err
			}
			err = p.printStmt(x)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(output, "}\n")
		if err != nil {
			return err
		}

		if len(stmt.False) > 0 {
			_, err = fmt.Fprintf(output, "else {\n")
			if err != nil {
				return err
			}

			for _, x := range stmt.False {
				_, err = fmt.Fprintf(output, "\t")
				if err != nil {
					return err
				}
				err = p.printStmt(x)
				if err != nil {
					return err
				}
			}
			_, err = fmt.Fprintf(output, "}\n")
			if err != nil {
				return err
			}
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
