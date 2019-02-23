package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/govm-project/govm/engines/docker"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

var startCommand = cli.Command{
	Name:    "start",
	Aliases: []string{"up", "s"},
	Usage:   "Start a GoVM Instance",
	Flags:   []cli.Flag{},
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
