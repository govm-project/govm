package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	yaml "gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

/* global variables */
var efi bool
var cloud bool
var hostDNS bool
var keyPath string
var wdir string

//ConfigDriveMetaData stands for cloud images config drive
type ConfigDriveMetaData struct {
	AvailabilityZone string            `json:"availavility_zone"`
	Hostname         string            `json:"hostname"`
	LaunchIndex      string            `json:"launch_index"`
	Name             string            `json:"name"`
	Meta             map[string]string `json:"meta"`
	PublicKeys       map[string]string `json:"public_keys"`
	UUID             string            `json:"uuid"`
}

// helper function to find a tcp port
// TODO: Find a better way to do this
func findPort() int {
	address, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}

	listen, err := net.ListenTCP("tcp", address)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = listen.Close()
		// TODO: log error in case close statement fails
	}()
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
	wdir = strings.Replace(VMLauncherWorkdir, "$HOME", home, 1)
	keyPath = strings.Replace(SSHPublicKeyFile, "$HOME", home, 1)

	// Check sane working directory
	wdir, _ = filepath.Abs(wdir)
	_, err := os.Stat(wdir)
	if err != nil {
		fmt.Printf(" %v does not exists\n", wdir)

		fmt.Printf("Creating %s", wdir+"/data")
		err = os.MkdirAll(wdir+"/data", 0755)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Creating %s", wdir+"/images")
		err = os.Mkdir(wdir+"/images", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	vm := vmInit()
	err = vm.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

/* Define the vm cli app */
func vmInit() *cli.App {
	vmCLI := cli.NewApp()
	vmCLI.Name = "govm"
	vmCLI.Usage = "VMs as you go"
	/* Global flags */
	vmCLI.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "workdir",
			Value: "",
			Usage: "Alternate working directory. Default: ~/vms",
		},
	}

	// sub-commands
	vmCLI.Commands = []cli.Command{
		create(),
		delete(),
		list(),
		compose(),
		connect(),
	}
	return vmCLI
}

func create() cli.Command {
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
		},
		Action: func(c *cli.Context) error {
			var parentImage string
			var flavor VMSize
			if c.String("image") == "" {
				fmt.Println("Missing --image argument")
				os.Exit(1)
			}
			parentImage, err := filepath.Abs(c.String("image"))
			if err != nil {
				fmt.Printf("Unable to determine image location: %v\n", err)
				os.Exit(1)
			}
			err = saneImage(parentImage)
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			// Check if any flavor is provided
			if c.String("flavor") != "" {
				flavor = GetVMSizeFromFlavor(c.String("flavor"))
			} else {
				flavor = NewVMSize(
					c.String("cpumodel"),
					c.Int("sockets"),
					c.Int("cpus"),
					c.Int("cores"),
					c.Int("threads"),
					c.Int("ram"),
				)
			}

			vm := NewVM(
				c.String("name"),
				parentImage,
				flavor,
				c.Bool("cloud"),
				c.Bool("efi"),
				c.String("workdir"),
				c.String("key"),
				c.String("user-data"),
				NetworkingOptions{})
			vm.Launch()
			vm.ShowInfo()
			return nil
		},
	}
	return command
}

func delete() cli.Command {
	command := cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "Delete vms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "Delete all vms",
			},
		},
		Action: func(c *cli.Context) error {
			var name string

			if c.NArg() <= 0 {

				/* Mandatory argument */

				// VM name argument
				err := errors.New("missing VM name")
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

				err = websockifyProcess.Kill()
				if err != nil {
					// TODO: change it to warning once log package is changed
					log.Println(err)
				}
			}

			err = cli.ContainerRemove(ctx, name,
				types.ContainerRemoveOptions{
					RemoveVolumes: false,
					RemoveLinks:   false,
					Force:         true,
				})
			if err != nil {
				log.Fatal(err)
			}

			err = os.RemoveAll(containerDataPath)
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}
	return command
}

func list() cli.Command {
	command := cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List vms",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all",
				Usage: "List all images",
			},
		},
		Action: func(c *cli.Context) error {
			var containerIP string

			cli, err := client.NewEnvClient()
			if err != nil {
				panic(err)
			}
			listArgs := filters.NewArgs()
			listArgs.Add("ancestor", VMLauncherContainerImage)
			containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
				Quiet:   false,
				Size:    false,
				All:     true,
				Latest:  false,
				Since:   "",
				Before:  "",
				Limit:   0,
				Filters: listArgs,
			})
			if err != nil {
				panic(err)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w, "ID\t IP\t VNC_URL\t NAME")
			for _, container := range containers {
				for _, net := range container.NetworkSettings.Networks {
					containerIP = net.IPAddress
					break
				}
				fmt.Fprintln(w, container.ID[:10]+
					"\t "+containerIP+
					"\t http://localhost:"+container.Labels["websockifyPort"]+
					"\t "+container.Names[0][1:])
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
			var vmTemplate VMTemplate
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

func connect() cli.Command {
	command := cli.Command{
		Name:    "connect",
		Aliases: []string{"conn"},
		Usage:   "Get a shell from a vm",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "user",
				Value: "",
				Usage: "ssh login user",
			},
			cli.StringFlag{
				Name:  "key",
				Value: "",
				Usage: "private key path (default: ~/.ssh/id_rsa)",
			},
		},
		Action: func(c *cli.Context) error {
			var name, loginUser, key string
			var vmID int
			var nameFound bool = false
			nargs := c.NArg()
			switch {
			case nargs == 1:
				// Parse flags
				if c.String("user") != "" {
					loginUser = c.String("user")
				} else {
					usr, _ := user.Current()
					loginUser = usr.Name
				}

				if c.String("key") != "" {
					key, _ = filepath.Abs(c.String("key"))
				} else {
					usr, err := user.Current()
					if err != nil {
						log.Fatal(err)
					}

					key = usr.HomeDir + "/.ssh/id_rsa"
				}
				name = c.Args().First()
				cli, err := client.NewEnvClient()
				if err != nil {
					panic(err)
				}
				listArgs := filters.NewArgs()
				listArgs.Add("ancestor", VMLauncherContainerImage)
				containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
					Quiet:   false,
					Size:    false,
					All:     true,
					Latest:  false,
					Since:   "",
					Before:  "",
					Limit:   0,
					Filters: listArgs,
				})
				if err != nil {
					panic(err)
				}
				for id, container := range containers {
					if container.Names[0][1:] == name {
						nameFound = true
						vmID = id
					}
				}
				if nameFound != true {
					fmt.Printf("Unable to find a running vm with name: %s", name)
					os.Exit(1)
				} else {
					vmIP := containers[vmID].NetworkSettings.Networks["bridge"].IPAddress
					getNewSSHConn(loginUser, vmIP, key)
				}

			case nargs == 0:
				fmt.Println("No name provided as argument.")
				os.Exit(1)

			case nargs > 1:
				fmt.Println("Only one argument is allowed")
				os.Exit(1)
			}
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
		Usage:   "Global vm configuration",
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
