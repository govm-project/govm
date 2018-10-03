package cli

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// TODO: Reduce cyclomatic complexity
func connect() cli.Command { // nolint: gocyclo
	command := cli.Command{
		Name:    "connect",
		Aliases: []string{"conn"},
		Usage:   "Get a shell from a vm",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "user",
				Value: "",
				Usage: "ssh login user",
			},
			cli.StringFlag{
				Name:  "key",
				Value: "",
				Usage: "private key path (default: ~/.ssh/id_rsa)",
			},
		},
		Action: func(c *cli.Context) error {
			var name, loginUser, key string
			var vmID int
			nameFound := false
			nargs := c.NArg()
			switch {
			case nargs == 1:
				// Parse flags
				if c.String("user") != "" {
					loginUser = c.String("user")
				} else {
					usr, _ := user.Current()
					loginUser = usr.Name
				}

				if c.String("key") != "" {
					key, _ = filepath.Abs(c.String("key"))
				} else {
					usr, err := user.Current()
					if err != nil {
						log.Fatal(err)
					}

					key = usr.HomeDir + "/.ssh/id_rsa"
				}
				name = c.Args().First()
				cli, err := client.NewEnvClient()
				if err != nil {
					panic(err)
				}
				listArgs := filters.NewArgs()
				listArgs.Add("ancestor", VMLauncherContainerImage)
				containers, err := cli.ContainerList(context.Background(),
					types.ContainerListOptions{
						Quiet:   false,
						Size:    false,
						All:     true,
						Latest:  false,
						Since:   "",
						Before:  "",
						Limit:   0,
						Filters: listArgs,
					})
				if err != nil {
					panic(err)
				}
				for id, container := range containers {
					if container.Names[0][1:] == name {
						nameFound = true
						vmID = id
					}
				}
				if !nameFound {
					fmt.Printf("Unable to find a running vm with name: %s", name)
					os.Exit(1)
				} else {
					vmIP := containers[vmID].NetworkSettings.Networks["bridge"].IPAddress
					getNewSSHConn(loginUser, vmIP, key)
				}

			case nargs == 0:
				fmt.Println("No name provided as argument.")
				os.Exit(1)

			case nargs > 1:
				fmt.Println("Only one argument is allowed")
				os.Exit(1)
			}
			return nil
		},
	}
	return command
}
