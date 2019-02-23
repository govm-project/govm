package main

import (
	"fmt"
	"os"

	"github.com/govm-project/govm/pkg/cli"
	clilib "gopkg.in/urfave/cli.v2"
)

// Revision is used to print the commit hash when using the --version flag.
// This variable must be modified with ldflag when building.
var Revision string

func main() {
	// Custom version printer to show the Revision
	clilib.VersionPrinter = func(c *clilib.Context) {
		fmt.Fprintf(c.App.Writer, "%s version %s", c.App.Name, c.App.Version)
		if Revision != "" {
			fmt.Fprintf(c.App.Writer, " revision %s", Revision)
		}
		fmt.Fprintln(c.App.Writer)
	}

	cmd, err := cli.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = cmd.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
