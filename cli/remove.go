package cli

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func remove() cli.Command {
	command := cli.Command{
		Name:    "remove",
		Aliases: []string{"d"},
		Usage:   "Remove vms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "Remove all vms",
			},
		},
		Action: func(c *cli.Context) error {
			var name string

			if c.NArg() <= 0 {

				/* Mandatory argument */

				// VM name argument
				err := errors.New("missing VM name")
				fmt.Println(err)
				fmt.Printf("USAGE:\n govm remove [command options] [name]\n")
				os.Exit(1)
			}
			name = c.Args().First()
			ctx := context.Background()
			cli, err := client.NewEnvClient()
			if err != nil {
				panic(err)
			}
			containerJSON, err := cli.ContainerInspect(ctx, name)
			if err != nil {
				log.Fatal(err)
			}

			containerDataPath := containerJSON.Config.Labels["dataDir"]
			pid, err := ioutil.ReadFile(containerDataPath + "/websockifyPid")
			if err == nil {
				websockifyPid, _ := strconv.Atoi(string(pid))
				websockifyProcess, err := os.FindProcess(websockifyPid)
				if err != nil {
					log.Fatal(err)
				}

				err = websockifyProcess.Kill()
				if err != nil {
					// TODO: change it to warning once log package is changed
					log.Println(err)
				}
			}

			err = cli.ContainerRemove(ctx, name,
				types.ContainerRemoveOptions{
					RemoveVolumes: false,
					RemoveLinks:   false,
					Force:         true,
				})
			if err != nil {
				log.Fatal(err)
			}

			err = os.RemoveAll(containerDataPath)
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}
	return command
}
