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
var saveCommand = cli.Command{
	Name:    "save",
	Aliases: []string{"snapshot"},
	Usage:   "Saves a GoVM Instance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "out",
			Value: "backup.img",
			Usage: "Path to backup file",
		},
		&cli.BoolFlag{
			Name:  "stopvm",
			Value: false,
			Usage: "Stop the VM during snapshot",
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() <= 0 {
			err := errors.New("missing GoVM Instance name")
			fmt.Println(err)
			fmt.Printf("USAGE:\n govm stop [command options] [name]\n")
			os.Exit(1)
		}

		namespace := ctx.String("namespace")
		name := ctx.Args().First()
		backupFile := ctx.String("out")
		stopVM := ctx.Bool("stopvm")

		engine := docker.Engine{}
		engine.Init()
		err := engine.Save(namespace, name, backupFile, stopVM)
		if err != nil {
			log.Fatalf("Error when saving the GoVM Instance %v: %v", name, err)
		}

		log.Printf("GoVM Instance %v has been successfully stopped", name)

		return nil
	},
}
