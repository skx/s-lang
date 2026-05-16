package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

// Structure for our options and state.
type executeCommand struct {
	// The binary to generate, and later run.
	output string
}

// Arguments adds per-command args to the object.
func (e *executeCommand) Arguments(f *flag.FlagSet) {
	f.StringVar(&e.output, "output", "a.out", "Where to write the generated binary")
}

// Info returns the name of this subcommand.
func (e *executeCommand) Info() (string, string) {
	return "execute", `Execute the given source program.

Details:

This command calls the "generate" sub-command to write a compiled
version of the given source file to disk, then executes it.

Example:

    $ s-lang execute example/example.in
`
}

func (e *executeCommand) processFile(path string) error {

	// Use our generate Command as a helper
	g := &compileCommand{output: e.output}
	err := g.processFile(path)
	if err != nil {
		return err
	}

	// compile
	run := exec.Command("./a.out")
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr

	err = run.Run()
	if err != nil {
		return (fmt.Errorf("error launching binary: %s", err))
	}

	return nil
}

// Execute is invoked if the user specifies `execute` as the subcommand.
func (e *executeCommand) Execute(args []string) int {

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: s-lang execute source.in\n")
		return 1
	}

	err := e.processFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		return 1
	}

	return 0
}
