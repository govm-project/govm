package cli

import (
	"fmt"

	"github.com/codegangsta/cli"
)

// Revision is used to print the commit hash when using the --version flag.
// This variable must be modified with ldflag when building.
var Revision string

//Init initializes the CLI app
func Init() *cli.App {

	// Modify default binary's version string
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version=%s revision=%s\n", c.App.Version, Revision)
	}
	vmCLI := cli.NewApp()
	vmCLI.Name = "govm"
	vmCLI.Usage = "VMs as you go"
	/* Global flags */
	vmCLI.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "workdir",
			Value: "",
			Usage: "Alternate working directory. Default: ~/vms",
		},
	}

	// sub-commands
	vmCLI.Commands = []cli.Command{
		create(),
		remove(),
		list(),
		compose(),
		connect(),
	}

	return vmCLI
}
