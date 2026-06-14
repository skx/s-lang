package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/skx/subcommands"
)

// Structure for our options and state.
type versionCommand struct {
	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info returns the name of this subcommand.
func (v *versionCommand) Info() (string, string) {
	return "version", `Show the version of this compiler.

Details:

This command outputs the git version information which was
produced when this binary was compiled.

Example:

    $ s-lang version
`
}

// Execute is invoked if the user specifies `version` as the subcommand.
func (v *versionCommand) Execute(args []string) int {

	version := "unknown"

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				version = setting.Value
			}
		}
	}

	// print the name and version
	fmt.Printf("%s - %s\n", filepath.Base(os.Args[0]), version)

	// Now any other information we found.
	if ok {
		for _, setting := range info.Settings {
			if strings.Contains(setting.Key, "vcs") {
				fmt.Printf("\t%s\t%s\n", setting.Key, setting.Value)
			}
		}
	}

	return 0
}
