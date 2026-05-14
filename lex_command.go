package main

import (
	"fmt"
	"os"

	"github.com/skx/subcommands"
	"s-lang/lexer"
)

// Structure for our options and state.
type lexCommand struct {
	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info returns the name of this subcommand.
func (l *lexCommand) Info() (string, string) {
	return "lex", `Show the lexer output of the given program.

Details:

This command allows you to see what the lexer produces for the
given input file.

Example:

    $ s-lang lex example.in
`
}

func (l *lexCommand) lexFile(path string) error {

	// Read the file-contents
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lex := lexer.NewLexer(string(data))

	tok := lex.Next()
	for tok.Type != lexer.EOF {
		_, err := fmt.Fprintf(output, "%v\n", tok)
		if err != nil {
			return err
		}
		tok = lex.Next()
	}
	return nil
}

// Execute is invoked if the user specifies `lex` as the subcommand.
func (l *lexCommand) Execute(args []string) int {

	// For each argument ..
	for _, arg := range args {

		err := l.lexFile(arg)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return 1
		}
	}

	return 0
}
