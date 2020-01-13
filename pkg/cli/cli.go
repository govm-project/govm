package cli

import (
	"github.com/govm-project/govm/pkg/homedir"
	"github.com/govm-project/govm/pkg/nameutil"
	cli "github.com/urfave/cli/v2"
)

// New creates a new command line GoVM application
func New() (*cli.App, error) {
	defaultNamespace, err := nameutil.DefaultNamespace()
	if err != nil {
		return nil, err
	}

	return &cli.App{
		Name:  "govm",
		Usage: "VMs as you go",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "namespace",
				Aliases: []string{"ns", "N"},
				Value:   defaultNamespace,
				Usage:   "operate on VMs from this namespace",
			},
			&cli.StringFlag{
				Name:  "workdir",
				Value: homedir.ExpandPath(VMLauncherWorkdir),
				Usage: "alternative working directory",
			},
		},
		Commands: []*cli.Command{
			&createCommand,
			&listCommand,
			&removeCommand,
			&startCommand,
			&composeCommand,
			&sshCommand,
			&stopCommand,
		},
	}, nil
}
