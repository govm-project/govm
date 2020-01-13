package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/govm-project/govm/engines/docker"
	"github.com/govm-project/govm/internal"
	"github.com/govm-project/govm/vm"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// nolint: gochecknoglobals
var createCommand = cli.Command{
	Name:      "create",
	Aliases:   []string{"c"},
	Usage:     "Create a new VM",
	ArgsUsage: "name",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "image",
			Value: "",
			Usage: "Path to image",
		},
		&cli.StringFlag{
			Name:  "user-data",
			Value: "",
			Usage: "Path to user data file",
		},
		&cli.BoolFlag{
			Name:  "efi",
			Usage: "Use efi bootloader",
		},
		&cli.BoolFlag{
			Name:  "cloud",
			Usage: "Create config-drive for cloud-images",
		},
		&cli.StringFlag{
			Name:  "flavor",
			Usage: "VM specs descriptor",
		},
		&cli.StringFlag{
			Name:  "key",
			Usage: "SSH key to be included in a cloud image.",
		},
		&cli.StringFlag{
			Name:  "name",
			Value: "",
			Usage: "vm name",
		},
		&cli.StringFlag{
			Name:  "cpumodel",
			Value: "",
			Usage: "Model of the virtual cpu. See: qemu-system-x86_64 -cpu help",
		},
		&cli.IntFlag{
			Name:  "sockets",
			Value: 1,
			Usage: "Number of sockets.",
		},
		&cli.IntFlag{
			Name:  "cpus",
			Value: 1,
			Usage: "Number of cpus",
		},
		&cli.IntFlag{
			Name:  "cores",
			Value: 2,
			Usage: "Number of cores",
		},
		&cli.IntFlag{
			Name:  "threads",
			Value: 2,
			Usage: "Number of threads",
		},
		&cli.IntFlag{
			Name:  "ram",
			Value: 1024,
			Usage: "Allocated RAM in MB",
		},
		&cli.IntFlag{
			Name:  "disk",
			Value: 50,
			Usage: "Root disk size in GB",
		},
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug mode",
		},
		&cli.StringSliceFlag{
			Name:  "share",
			Usage: "Share directories. e.g. --share /host/path:/guest/path",
		},
		&cli.StringSliceFlag{
			Name:  "container-env",
			Usage: "Environment variable. e.g. --container-env http_proxy=$http_proxy",
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		if ctx.String("image") == "" {
			fmt.Println("Missing --image argument")
			os.Exit(1)
		}

		// Check if any flavor is provided
		var size vm.Size
		if ctx.String("flavor") != "" {
			size = vm.GetSizeFromFlavor(ctx.String("flavor"))
		} else {
			size = vm.NewSize(
				ctx.String("cpumodel"),
				ctx.Int("sockets"),
				ctx.Int("cpus"),
				ctx.Int("cores"),
				ctx.Int("threads"),
				ctx.Int("ram"),
				ctx.Int("disk"),
			)
		}

		// Check if there are any shares and validate the format.
		// They must be separated by the ":" characted as docker does
		if len(ctx.StringSlice("share")) > 0 {
			for _, dir := range ctx.StringSlice("share") {
				share := strings.Split(dir, ":")
				if len(share) != 2 {
					log.Fatal("Wrong share format: " + dir +
						"\nUsage: --share /host/path:/guest/path")

				}

			}
		}

		workDir := ctx.String("workdir")
		if workDir == "" {
			workDir = internal.GetDefaultWorkDir()
		}
		newVM := vm.Instance{
			Name:             ctx.String("name"),
			Namespace:        ctx.String("namespace"),
			ParentImage:      ctx.String("image"),
			Workdir:          workDir,
			SSHPublicKeyFile: ctx.String("key"),
			UserData:         ctx.String("user-data"),
			Size:             size,
			Cloud:            ctx.Bool("cloud"),
			Efi:              ctx.Bool("efi"),
			NetOpts:          vm.NetworkingOptions{},
			Shares:           ctx.StringSlice("share"),
			ContainerEnvVars: ctx.StringSlice("container-env"),
		}

		if err := newVM.Check(); err != nil {
			log.Fatalf("Error on VM Instance pre-check: %v", err)
		}

		engine := docker.Engine{}
		engine.Init()
		id, err := engine.Create(newVM)
		if err != nil {
			log.Fatalf("Error when creating the new VM: %v", err)
		}
		err = engine.Start(newVM.Namespace, id)
		if err != nil {
			log.Fatalf("Error when starting the new VM: %v", err)
		}

		log.Printf("GoVM Instance %v has been successfully created", newVM.Name)

		return nil
	},
}
