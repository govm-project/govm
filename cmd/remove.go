package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/govm-project/govm/engines/docker"
	"github.com/govm-project/govm/internal"

	"github.com/codegangsta/cli"
	log "github.com/sirupsen/logrus"
)

func remove() cli.Command {
	defaultNamespace, err := internal.DefaultNamespace()
	if err != nil {
		log.Fatalf("get default namespace: %v", err)
	}
	command := cli.Command{
		Name:    "remove",
		Aliases: []string{"d"},
		Usage:   "Remove vms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "Remove all vms",
			},
			cli.StringFlag{
				Name:  "namespace",
				Value: defaultNamespace,
				Usage: "list VMs from this namespace",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() <= 0 {
				err := errors.New("missing VM name")
				fmt.Println(err)
				fmt.Printf("USAGE:\n govm remove [command options] [name]\n")
				os.Exit(1)
			}

			namespace := c.String("namespace")
			name := c.Args().First()

			engine := docker.Engine{}
			engine.Init()
			err := engine.Delete(namespace, name)
			if err != nil {
				log.Fatalf("Error when removing the VM %v: %v", name, err)
			}

			log.Printf("GoVM Instance %v has been successfully removed", name)

			return nil
		},
	}
	return command
}
