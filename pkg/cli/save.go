package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/govm-project/govm/engines/docker"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

// nolint: gochecknoglobals
var saveCommand = cli.Command{
	Name:    "save",
	Aliases: []string{"sv"},
	Usage:   "Save a GoVM Instance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "hot",
			Value: "",
			Usage: "Save vm without stopping container",
		},
		&cli.StringFlag{
			Name:  "cold",
			Value: "",
			Usage: "Save vm stopping the container",
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() <= 0 {
			err := errors.New("missing GoVM Instance name")
			fmt.Println(err)
			fmt.Printf("USAGE:\n govm save [command options] [name]\n")
			os.Exit(1)
		}

		namespace := c.String("namespace")
		name := c.Args().First()

		engine := docker.Engine{}
		engine.Init()
		err := engine.Save(namespace, name)
		if err != nil {
			log.Fatalf("Error when saving the GoVM Instance %v: %v", name, err)
		}

		log.Printf("GoVM Instance %v has been successfully saved", name)

		return nil
	},
}