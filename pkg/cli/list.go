package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/govm-project/govm/engines/docker"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

var listCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"ls"},
	Usage:   "List VMs",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "list VMs from all namespaces",
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
