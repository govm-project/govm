package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/govm-project/govm/engines/docker"
	"github.com/govm-project/govm/internal"
	"github.com/govm-project/govm/vm"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

func compose() cli.Command {
	command := cli.Command{
		Name:    "compose",
		Aliases: []string{"co"},
		Usage:   "Deploy VMs from a compose config file",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "f",
				Value: "",
				Usage: "compose config file",
			},
		},
		Action: func(c *cli.Context) (err error) {
			var composeConfig vm.ComposeConfig
			composeFilePath := c.String("f")
			if composeFilePath != "" {
				composeFilePath, err = internal.CheckFilePath(composeFilePath)
				if err != nil {
					return
				}
			} else {
				fmt.Println("Missing compose file")
				os.Exit(1)
			}

			composeFile, _ := ioutil.ReadFile(composeFilePath)
			err = yaml.Unmarshal(composeFile, &composeConfig)
			if err != nil {
				fmt.Printf("yaml file error: %v\n", err)
				os.Exit(1)
			}

			engine := docker.Engine{}
			engine.Init()
			for _, vm := range composeConfig.VMs {
				if vm.Workdir == "" {
					vm.Workdir = internal.GetDefaultWorkDir()
				}

				if vm.Namespace == "" {
					defaultNamespace, err := internal.DefaultNamespace()
					if err != nil {
						log.Fatalf("get default namespace: %v", err)
					}
					vm.Namespace = defaultNamespace
				}

				if err := vm.Check(); err != nil {
					log.Fatalf("Error on VM Instance pre-check: %v", err)
				}
				id, err := engine.Create(vm)
				if err != nil {
					log.Fatalf("Error when creating the new VM: %v", err)
				}
				err = engine.Start(vm.Namespace, id)
				if err != nil {
					log.Fatalf("Error when starting the new VM: %v", err)
				}

				log.Printf("GoVM Instance %v has been successfully created", vm.Name)
			}
			return nil
		},
	}
	return command
}
