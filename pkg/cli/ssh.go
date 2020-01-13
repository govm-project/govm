package cli

import (
	"fmt"

	"github.com/govm-project/govm/engines/docker"
	"github.com/govm-project/govm/pkg/termutil"
	cli "github.com/urfave/cli/v2"
)

// nolint: gochecknoglobals
var sshCommand = cli.Command{
	Name:  "ssh",
	Usage: "ssh into a running VM",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "login as this username",
		},
		&cli.StringFlag{
			Name:    "key",
			Aliases: []string{"k"},
			Usage:   "ssh private key file",
			Value:   "~/.ssh/id_rsa",
		},
	},
	Action: func(c *cli.Context) error {
		if c.Args().Len() != 1 {
			return fmt.Errorf("VM name required")
		}
		name := c.Args().First()
		namespace := c.String("namespace")
		user := c.String("user")
		if user == "" {
			return fmt.Errorf("--user argument required")
		}
		key := c.String("key")
		term := termutil.StdTerminal()

		engine := docker.Engine{}
		engine.Init()

		return engine.SSHVM(namespace, name, user, key, term)
	},
}
