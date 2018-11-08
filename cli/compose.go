package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"

	"github.com/govm-project/govm/vm"
)

func compose() cli.Command {
	command := cli.Command{
		Name:    "compose",
		Aliases: []string{"co"},
		Usage:   "Deploy Vms from yaml templates",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "f",
				Value: "",
				Usage: "template file",
			},
		},
		Action: func(c *cli.Context) error {
			var template string
			var vmTemplate vm.ComposeTemplate
			if c.String("f") != "" {
				template = c.String("f")
				template, err := filepath.Abs(template)
				if err != nil {
					fmt.Printf("Unable to determine template file location: %v\n", err)
					os.Exit(1)
				}
				// Test if the template file exists
				templateStat, err := os.Stat(template)
				if err != nil {
					return fmt.Errorf("file %v does not exist", template)
				}

				// Test if the image is valid or has a valid path
				mode := templateStat.Mode()
				if !mode.IsRegular() {
					return fmt.Errorf("%v is not a regular file", template)
				}
			} else {
				fmt.Println("Missing template")
				os.Exit(1)
			}
			templateFile, _ := ioutil.ReadFile(template)
			err := yaml.Unmarshal(templateFile, &vmTemplate)
			if err != nil {
				fmt.Printf("yaml file error: %v\n", err)
				os.Exit(1)
			}

			vmTemplate = NewVMTemplate(&vmTemplate)
			for _, vm := range vmTemplate.VMs {
				vm.Launch()
			}
			return nil
		},
	}
	return command
}
