package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func list() cli.Command {
	command := cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List vms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "List all images",
			},
		},
		Action: func(c *cli.Context) error {
			var containerIP string

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
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w, "ID\t IP\t VNC_URL\t NAME")
			for _, container := range containers {
				for _, net := range container.NetworkSettings.Networks {
					containerIP = net.IPAddress
					break
				}
				fmt.Fprintln(w, container.ID[:10]+
					"\t "+containerIP+
					"\t http://localhost:"+container.Labels["websockifyPort"]+
					"\t "+container.Names[0][1:])
			}

			err = w.Flush()
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}
	return command
}
