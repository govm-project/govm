package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/govm-project/govm/engines/docker"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/govm-project/govm/utils"
	log "github.com/sirupsen/logrus"
)

func list() cli.Command {
	defaultNamespace, err := utils.DefaultNamespace()
	if err != nil {
		log.Fatalf("get default namespace: %v", err)
	}
	command := cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List vms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name: "all",
				// Usage: "List all images",
				Usage: "list VMs from all namespaces",
			},
			cli.StringFlag{
				Name:  "namespace",
				Value: defaultNamespace,
				Usage: "list VMs from this namespace",
			},
		},
		Action: func(c *cli.Context) error {
			var containerIP string

			// TODO: Replace it when docker engine is properly addressed
			docker.SetAPIVersion()
			cli, err := client.NewClientWithOpts(client.FromEnv)
			if err != nil {
				panic(err)
			}
			namespace := c.String("namespace")

			listArgs := filters.NewArgs()
			listArgs.Add("ancestor", VMLauncherContainerImage)
			if !c.Bool("all") {
				listArgs.Add("label", "namespace="+namespace)
			}

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
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			if c.Bool("all") {
				fmt.Fprintln(w, "ID\tIP\tVNC URL\tNAME\tNAMESPACE")
				for _, container := range containers {
					for _, net := range container.NetworkSettings.Networks {
						containerIP = net.IPAddress
						break
					}
					fmt.Fprintf(w, "%s\t%s\thttp://localhost:%s\t%s\t%s\n",
						container.ID[:10], containerIP,
						container.Labels["websockifyPort"],
						container.Labels["vmName"],
						container.Labels["namespace"])
				}
			} else {
				fmt.Fprintln(w, "ID\tIP\tVNC URL\tNAME")
				for _, container := range containers {
					for _, net := range container.NetworkSettings.Networks {
						containerIP = net.IPAddress
						break
					}
					fmt.Fprintln(w, container.ID[:10]+
						"\t "+containerIP+
						"\t http://localhost:"+container.Labels["websockifyPort"]+
						"\t "+container.Labels["vmName"])
				}
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
