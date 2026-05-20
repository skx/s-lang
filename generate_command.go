package main

import (
	"flag"
	"fmt"
	"os"

	"s-lang/compiler"
)

// Structure for our options and state.
type generateCommand struct {
	// stdlibCheck specifies whether we should do
	// compile-time type checking of the standard
	// library calls.
	stdlibCheck bool

	// output will be the file to generate
	output string
}

// Arguments adds per-command args to the object.
func (g *generateCommand) Arguments(f *flag.FlagSet) {
	f.StringVar(&g.output, "output", "", "File to write our generated assembly to, STDOUT if not specified.")
	f.BoolVar(&g.stdlibCheck, "check-stdlib", true, "Should we run compile-time type checking on the standard library.")
}

// Info returns the name of this subcommand.
func (g *generateCommand) Info() (string, string) {
	return "generate", `Generate an assembly version of the given program.

Details:

This command allows you to generate an assembly version of the given program.

The assembly file is standalone and may then be compiled via 'as' and linked
with 'ld' as usual.  The built-in 'compile' sub-command will run the assembler
and linker for you.

Example:

    $ s-lang generate example.in
    $ s-lang generate -output example.s example.in
`
}

// processFile takes the given source file and generates the appropriate assembly
// language from its contents, after parsing.
func (g *generateCommand) processFile(path string) error {

	// Read the file-contents
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// If a file was provided, use it instead
	if g.output != "" {
		var file *os.File
		file, err = os.Create(g.output)
		if err != nil {
			return err
		}
		defer file.Close()
		output = file
	}

	// Create a compiler object
	c, err2 := compiler.New(
		compiler.WithSource(string(data)),
		compiler.WithCompileChecking(g.stdlibCheck),
	)
	if err2 != nil {
		return err2
	}

	txt := ""
	txt, err = c.Compile()
	if err != nil {
		return err
	}

	// Write the text to the output file/handle.
	fmt.Fprintf(output, "%s", txt)

	return nil
}

// Execute is invoked if the user specifies `generate` as the subcommand.
func (g *generateCommand) Execute(args []string) int {

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: s-lang generate [options] source.in\n")
		return 1
	}

	err := g.processFile(args[0])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return 1
	}
	return 0
}
