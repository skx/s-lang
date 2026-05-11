package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

// Structure for our options and state.
type compileCommand struct {

	// output will be the file to generate
	output string
}

// Arguments adds per-command args to the object.
func (c *compileCommand) Arguments(f *flag.FlagSet) {
	f.StringVar(&c.output, "output", "a.out", "Where to write the generated binary")
}

// Info returns the name of this subcommand.
func (c *compileCommand) Info() (string, string) {
	return "compile", `Compile the given source program.

Details:

This command allows you to generate an compiled version of the
given source file.  A temporary file will be created to store
the source code, then this will be used by the compiler and
removed after processing.

This sub-command automates the compilation and linking of the
output produced by the 'generate' sub-command.

Example:

    $ s-lang compile example.in
    $ s-lang compile -output example example.in
`
}

func (c *compileCommand) processFile(path string) error {
	f, err := os.CreateTemp("", "sample")
	if err != nil {
		return err
	}

	// cleanup once done
	defer os.Remove(f.Name())

	// Use our generate Command as a helper
	g := &generateCommand{output: f.Name()}
	err = g.processFile(path)
	if err != nil {
		return err
	}

	// compile
	ass := exec.Command("as", "-msyntax=intel", "-mnaked-reg", f.Name(), "-o", f.Name()+".o")
	ass.Stdout = os.Stdout
	ass.Stderr = os.Stderr

	err = ass.Run()
	if err != nil {
		return (fmt.Errorf("error launching assembler: %s", err))
	}

	// remove the generated object file once complete
	defer os.Remove(f.Name() + ".o")

	// link
	ld := exec.Command("ld", "-s", "-o", c.output, f.Name()+".o")
	ld.Stdout = os.Stdout
	ld.Stderr = os.Stderr

	err = ld.Run()
	if err != nil {
		return (fmt.Errorf("error launching linker: %s", err))
	}

	return nil
}

// Execute is invoked if the user specifies `compile` as the subcommand.
func (c *compileCommand) Execute(args []string) int {

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: s-lang compile [options] source.in\n")
		return 1
	}

	err := c.processFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		return 1
	}

	return 0
}
