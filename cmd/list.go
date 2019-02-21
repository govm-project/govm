package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/govm-project/govm/engines/docker"
	"github.com/govm-project/govm/internal"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func list() cli.Command {
	defaultNamespace, err := internal.DefaultNamespace()
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

			engine := docker.Engine{}
			engine.Init()

			instances, err := engine.List(c.String("namespace"), c.Bool("all"))
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tName\tNamespace\tIP\tVNC")
			for _, instance := range instances {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					instance.ID,
					instance.Name,
					instance.Namespace,
					instance.NetOpts.IP,
					fmt.Sprintf("http://localhost:%v", instance.VNCPort),
				)
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
