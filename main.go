package main

import (
	"fmt"
	"io"
	"os"

	"github.com/skx/subcommands"
)

// output is used as the output handle for all our
// commands.  This is so that their output may be
// captured by our test-cases
var output io.Writer = os.Stdout

// Recovery is good
func recoverPanic() {
	if os.Getenv("DEBUG") != "" {
		return
	}

	if r := recover(); r != nil {
		fmt.Printf("recovered from panic while running %v\n%s\n", os.Args, r)
		fmt.Printf("To see the panic run 'export DEBUG=on' and repeat.\n")
	}
}

// Register the subcommands, and run the one the user chose.
func main() {

	//
	// Catch errors
	//
	defer recoverPanic()

	subcommands.Register(&lexCommand{})
	subcommands.Register(&parseCommand{})
	subcommands.Register(&generateCommand{})
	subcommands.Register(&compileCommand{})
	subcommands.Register(&executeCommand{})
	subcommands.Register(&versionCommand{})

	//
	// Execute the one the user chose.
	//
	os.Exit(subcommands.Execute())
}
