package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/govm-project/govm/engines/docker"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// nolint: gochecknoglobals
var stopCommand = cli.Command{
	Name:    "stop",
	Aliases: []string{"down", "d"},
	Usage:   "Stop a GoVM Instance",
	Flags:   []cli.Flag{},
	Action: func(c *cli.Context) error {
		if c.NArg() <= 0 {
			err := errors.New("missing GoVM Instance name")
			fmt.Println(err)
			fmt.Printf("USAGE:\n govm stop [command options] [name]\n")
			os.Exit(1)
		}

		namespace := c.String("namespace")
		name := c.Args().First()

		engine := docker.Engine{}
		engine.Init()
		err := engine.Stop(namespace, name)
		if err != nil {
			log.Fatalf("Error when stopping the GoVM Instance %v: %v", name, err)
		}

		log.Printf("GoVM Instance %v has been successfully stopped", name)

		return nil
	},
}
