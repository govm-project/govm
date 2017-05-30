package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/moby/moby/pkg/namesgenerator"
)

/* cli argument variables */
var flavor HostOpts
var efi bool
var cloud bool
var image string
var cowImage string
var name string
var host_dns bool

var verbose bool
var userData string

var wdir string
var keyPath string
var ParentImage string

var SSHKey string

const (
	WORKDIR   = "$HOME/govm"
	SSHPUBKEY = "$HOME/.ssh/id_rsa.pub"
	IMAGE     = "$PWD/image.qcow2"
)

type ConfigDriveMetaData struct {
	AvailabilityZone string            `json:"availavility_zone"`
	Hostname         string            `json:"hostname"`
	LaunchIndex      string            `json:"launch_index"`
	Name             string            `json:"name"`
	Meta             map[string]string `json:"meta"`
	PublicKeys       map[string]string `json:"public_keys"`
	UUID             string            `json:"uuid"`
}

/* helper function to find a tcp port */
func findPort() int {
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}

	listen, err := net.ListenTCP("tcp", address)
	if err != nil {
		panic(err)
	}
	defer listen.Close()
	return listen.Addr().(*net.TCPAddr).Port
}

func saneImage(path string) error {

	// Test if the image file exists
	imgArg, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("Image %v does not exist", path)
	}

	// Test if the image is valid or has a valid path
	mode := imgArg.Mode()
	if !mode.IsRegular() {
		return fmt.Errorf("%v is not a regular file", path)
	}
	return nil
}

func prepare() error {
	return nil
}

func main() {

	/* Check environment */
	home := os.Getenv("HOME")
	if home == "" {
		fmt.Printf("\nUnable to determine $HOME\n")
		fmt.Printf("Please specify -workdir and -pubkey\n")
		os.Exit(1)
	}
	wdir = strings.Replace(WORKDIR, "$HOME", home, 1)
	keyPath = strings.Replace(SSHPUBKEY, "$HOME", home, 1)

	// Check sane working directory
	wdir, _ = filepath.Abs(wdir)
	_, err := os.Stat(wdir)
	if err != nil {
		fmt.Printf(" %v does not exists\n", wdir)
		fmt.Printf("Run the setup.sh first or try:\n\n\tmkdir -p %s\n", wdir)
		os.Exit(1)
	}

	govm := govmInit()
	govm.Run(os.Args)
}

/* Define the govm cli app */
func govmInit() *cli.App {
	govmcli := cli.NewApp()
	govmcli.Name = "govm"
	govmcli.Usage = "Virtual Machines on top of Docker containers"
	/* Global flags */
	govmcli.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "workdir",
			Value: "",
			Usage: "Alternate working directory. Default: ~/govm",
		},
	}

	/* govm commands */
	govmcli.Commands = []cli.Command{
		create(),
		delete(),
		list(),
	}
	return govmcli
}

/* COMMANDS */
func create() cli.Command {
	command := cli.Command{
		Name:      "create",
		Aliases:   []string{"c"},
		Usage:     "Create a new govm",
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
				Name:  "size",
				Usage: "Custom VM specs. --size <cores>,<threads>,<ram>",
			},
			cli.StringFlag{
				Name:  "key",
				Value: "",
				Usage: "SSH key to be included in a cloud image.",
			},
			cli.StringFlag{
				Name:  "name",
				Value: "",
				Usage: "govm name",
			},
		},
		Action: func(c *cli.Context) error {
			/* Mandatory Flags */
			if c.String("image") == "" {
				fmt.Println("Missing --image argument")
				os.Exit(1)
			}
			ParentImage, err := filepath.Abs(c.String("image"))
			if err != nil {
				fmt.Printf("Unable to determine image location: %v\n", err)
				os.Exit(1)
			}
			err = saneImage(ParentImage)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			/* Optional Flags */
			if c.String("name") != "" {
				name = c.String("name")
			} else {
				name = namesgenerator.GetRandomName(0)
			}

			ctx := context.Background()
			cli, err := client.NewEnvClient()
			if err != nil {
				panic(err)
			}
			_, err = cli.ContainerInspect(ctx, name)
			if err == nil {
				log.Fatal("There is an existing container with the same name")
			}

			// Check if user data is provided
			if c.String("user-data") != "" {
				userData, _ = filepath.Abs(c.String("user-data"))
			}

			// Check if any flavor is provided
			if c.String("flavor") != "" {
				flavor = getFlavor(c.String("flavor"))
			} else if c.String("size") != "" {
				flavor = getCustomFlavor(c.String("size"))
			} else {
				flavor = getFlavor("")
			}

			// Check if efi flag is provided
			if c.Bool("efi") != false {
				efi = c.Bool("efi")
			}

			// Check if cloud flag is provided
			if c.Bool("cloud") != false {
				cloud = c.Bool("cloud")

			}

			if c.String("key") != "" {
				key, err := ioutil.ReadFile(c.String("key"))
				if err != nil {
					log.Fatal(err)
				}
				SSHKey = string(key)
			} else {
				key, err := ioutil.ReadFile(keyPath)
				if err != nil {
					log.Fatal(err)
				}
				SSHKey = string(key)
			}

			govm := NewGoVM(name, ParentImage, flavor, cloud, efi, wdir)
			govm.Launch()
			govm.ShowInfo()
			return nil
		},
	}
	return command
}

func delete() cli.Command {
	command := cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "Delete govms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "Delete all govms",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() <= 0 {

				/* Mandatory argument */

				// VM name argument
				err := errors.New("Missing VM name.\n")
				fmt.Println(err)
				fmt.Printf("USAGE:\n govm delete [command options] [name]\n")
				os.Exit(1)
			}
			name = c.Args().First()
			ctx := context.Background()
			cli, err := client.NewEnvClient()
			if err != nil {
				panic(err)
			}
			containerJSON, err := cli.ContainerInspect(ctx, name)
			if err != nil {
				log.Fatal(err)
			}

			containerDataPath := containerJSON.Config.Labels["dataDir"]
			pid, err := ioutil.ReadFile(containerDataPath + "/websockifyPid")
			if err == nil {
				websockifyPid, _ := strconv.Atoi(string(pid))
				websockifyProcess, err := os.FindProcess(websockifyPid)
				if err != nil {
					log.Fatal(err)
				}
				websockifyProcess.Kill()
			}

			err = cli.ContainerRemove(ctx, name, types.ContainerRemoveOptions{false, false, true})
			if err != nil {
				log.Fatal(err)
			}
			os.RemoveAll(containerDataPath)

			return nil
		},
	}
	return command
}

func list() cli.Command {
	command := cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List govms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "List all images",
			},
		},
		Action: func(c *cli.Context) error {
			//if c.NArg() > 0 {}
			cli, err := client.NewEnvClient()
			if err != nil {
				panic(err)
			}
			listArgs := filters.NewArgs()
			listArgs.Add("ancestor", "verbacious/govm")
			containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
				false,
				false,
				true,
				false,
				"",
				"",
				0,
				listArgs,
			})
			if err != nil {
				panic(err)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w, "ID\tIP\tVNC_PORT\tNAME")
			for _, container := range containers {
				fmt.Fprintln(w, container.ID[:10]+
					"\t"+container.NetworkSettings.Networks["bridge"].IPAddress+
					"\t"+container.Labels["websockifyPort"]+
					"\t"+container.Names[0][1:])
			}
			w.Flush()

			return nil
		},
	}
	return command
}

/* WIP
func config() cli.Command {
	command := cli.Command{
		Name:    "config",
		Aliases: []string{"conf"},
		Usage:   "Global govm configuration",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "websockify",
				Usage: "Enable websockify",
			},
		},
	}
	return command
}
*/
