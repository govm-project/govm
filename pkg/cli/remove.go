package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/govm-project/govm/engines/docker"
	cli "gopkg.in/urfave/cli.v2"

	log "github.com/sirupsen/logrus"
)

var removeCommand = cli.Command{
	Name:    "remove",
	Aliases: []string{"delete", "rm", "del"},
	Usage:   "Remove VMs",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "Remove all VMs",
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() <= 0 && !c.Bool("all") {
			err := errors.New("missing VM name")
			fmt.Println(err)
			fmt.Printf("USAGE:\n govm remove [command options] [name]\n")
			os.Exit(1)
		}

		namespace := c.String("namespace")
		engine := docker.Engine{}
		engine.Init()

		names := []string{}
		if all := c.Bool("all"); all {
			instances, err := engine.List(namespace, all)
			if err != nil {
				log.Fatalf("Error when listing current GoVM instances: %v", err)
			}
			for _, instance := range instances {
				names = append(names, instance.Name)
			}
		} else {
			names = append(names, c.Args().First())
		}

		for _, name := range names {
			err := engine.Delete(namespace, name)
			if err != nil {
				log.Fatalf("Error when removing the VM %v: %v", name, err)
			}

			log.Printf("GoVM Instance %v has been successfully removed", name)
		}
		return nil
	},
}
