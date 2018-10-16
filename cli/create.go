package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/govm-project/govm/types"
	"github.com/govm-project/govm/utils"
	"github.com/govm-project/govm/vm"
	log "github.com/sirupsen/logrus"
)

func create() cli.Command {
	defaultNamespace, err := utils.DefaultNamespace()
	if err != nil {
		log.Fatalf("get default namespace: %v", err)
	}
	command := cli.Command{
		Name:      "create",
		Aliases:   []string{"c"},
		Usage:     "Create a new vm",
		ArgsUsage: "name",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "image",
				Value: "",
				Usage: "Path to image",
			},
			cli.StringFlag{
				Name:  "user-data",
				Value: "",
				Usage: "Path to user data file",
			},
			cli.BoolFlag{
				Name:  "efi",
				Usage: "Use efi bootloader",
			},
			cli.BoolFlag{
				Name:  "cloud",
				Usage: "Create config-drive for cloud-images",
			},
			cli.StringFlag{
				Name:  "flavor",
				Usage: "VM specs descriptor",
			},
			cli.StringFlag{
				Name:  "key",
				Value: "",
				Usage: "SSH key to be included in a cloud image.",
			},
			cli.StringFlag{
				Name:  "name",
				Value: "",
				Usage: "vm name",
			},
			cli.StringFlag{
				Name:  "namespace",
				Value: defaultNamespace,
				Usage: "vm namespace",
			},
			cli.StringFlag{
				Name:  "cpumodel",
				Value: "",
				Usage: "Model of the virtual cpu. See: qemu-system-x86_64 -cpu help",
			},
			cli.IntFlag{
				Name:  "sockets",
				Value: 1,
				Usage: "Number of sockets.",
			},
			cli.IntFlag{
				Name:  "cpus",
				Value: 1,
				Usage: "Number of cpus",
			},
			cli.IntFlag{
				Name:  "cores",
				Value: 2,
				Usage: "Number of cores",
			},
			cli.IntFlag{
				Name:  "threads",
				Value: 2,
				Usage: "Number of threads",
			},
			cli.IntFlag{
				Name:  "ram",
				Value: 1024,
				Usage: "Allocated RAM",
			},
			cli.BoolFlag{
				Name:  "debug",
				Usage: "Debug mode",
			},
			cli.StringSliceFlag{
				Name:  "share",
				Usage: "Share directories. e.g. --share /host/path:/guest/path",
			},
		},
		Action: func(c *cli.Context) error {
			var parentImage string
			var flavor types.VMSize

			if c.Bool("debug") {
				log.SetLevel(log.DebugLevel)
			}

			if c.String("image") == "" {
				fmt.Println("Missing --image argument")
				os.Exit(1)
			}
			parentImage, err := filepath.Abs(c.String("image"))
			if err != nil {
				fmt.Printf("Unable to determine image location: %v\n", err)
				os.Exit(1)
			}
			err = vm.SaneImage(parentImage)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			// Check if any flavor is provided
			if c.String("flavor") != "" {
				flavor = vm.GetVMSizeFromFlavor(c.String("flavor"))
			} else {
				flavor = vm.NewVMSize(
					c.String("cpumodel"),
					c.Int("sockets"),
					c.Int("cpus"),
					c.Int("cores"),
					c.Int("threads"),
					c.Int("ram"),
				)
			}

			// Check if there are any shares and validate the format.
			// They must be separated by the ":" characted as docker does
			if len(c.StringSlice("share")) > 0 {
				for _, dir := range c.StringSlice("share") {
					share := strings.Split(dir, ":")
					if len(share) != 2 {
						log.Fatal("Wrong share format: " + dir +
							"\nUsage: --share /host/path:/guest/path")

					}

				}
			}

			workDir := c.String("workdir")
			if workDir == "" {
				workDir = getWorkDir()
			}
			newVM := vm.CreateVM(
				c.String("name"),
				c.String("namespace"),
				parentImage,
				workDir,
				c.String("key"),
				c.String("user-data"),
				flavor,
				c.Bool("cloud"),
				c.Bool("efi"),
				types.NetworkingOptions{},
				c.StringSlice("share"),
			)
			newVM.Launch()
			newVM.ShowInfo()
			return nil
		},
	}
	return command
}
