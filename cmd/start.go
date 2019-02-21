package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/govm-project/govm/engines/docker"
	"github.com/govm-project/govm/internal"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func start() cli.Command {
	defaultNamespace, err := internal.DefaultNamespace()
	if err != nil {
		log.Fatalf("get default namespace: %v", err)
	}
	command := cli.Command{
		Name:    "start",
		Aliases: []string{"d"},
		Usage:   "Start a GoVM Instance",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "namespace",
				Value: defaultNamespace,
				Usage: "Define Instance's namespace",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() <= 0 {
				err := errors.New("missing GoVM Instance name")
				fmt.Println(err)
				fmt.Printf("USAGE:\n govm start [command options] [name]\n")
				os.Exit(1)
			}

			namespace := c.String("namespace")
			name := c.Args().First()

			engine := docker.Engine{}
			engine.Init()
			err := engine.Start(namespace, name)
			if err != nil {
				log.Fatalf("Error when starting the GoVM Instance %v: %v", name, err)
			}

			log.Printf("GoVM Instance %v has been successfully started", name)

			return nil
		},
	}
	return command
}
