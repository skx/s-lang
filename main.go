package main

import (
	"fmt"
	"os"

	"github.com/skx/subcommands"
)

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

	//
	// Execute the one the user chose.
	//
	os.Exit(subcommands.Execute())
}
