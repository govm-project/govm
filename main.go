package main

import (
	"fmt"
	"os"

	"github.com/govm-project/govm/pkg/cli"
	clilib "gopkg.in/urfave/cli.v2"
)

var version = "undefined"

func main() {
	clilib.VersionPrinter = func(c *clilib.Context) {
		fmt.Fprintf(c.App.Writer, "%s version %s", c.App.Name, version)
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
