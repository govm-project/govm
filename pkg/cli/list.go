package cli

import (
	"fmt"
	"os"

	"github.com/govm-project/govm/engines/docker"

	"github.com/intel/tfortools"
	cli "github.com/urfave/cli/v2"
)

// nolint: gochecknoglobals
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
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "string containing the template code to execute",
		},
	},
	Action: func(c *cli.Context) error {

		engine := docker.Engine{}
		engine.Init()

		result, err := engine.List(c.String("namespace"), c.Bool("all"))
		if err != nil {
			return err
		}

		// Required due a limitation on intel/tfortools for embedded structures
		// https://github.com/intel/tfortools/issues/18
		type outInstance struct {
			ID        string
			Name      string
			Namespace string
			IP        string
		}

		instances := []outInstance{}
		for _, elem := range result {
			instances = append(instances, outInstance{elem.ID, elem.Name, elem.Namespace, elem.NetOpts.IP})
		}

		format := c.String("format")
		if format == "" {
			format = `{{table .}}`
		}

		err = tfortools.OutputToTemplate(os.Stdout, "format", format, instances, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, tfortools.GenerateUsageDecorated("format", instances, nil))
			return fmt.Errorf("unable to execute template : %v", err)
		}

		return nil
	},
}
