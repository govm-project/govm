package cli

import (
	"github.com/codegangsta/cli"
)

func Init() *cli.App {
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
