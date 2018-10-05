package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
	log "github.com/sirupsen/logrus"

	vmLauncher "github.com/govm-project/govm/vm"
)

//NewVMTemplate creates a new VMTemplate object
func NewVMTemplate(c *vmLauncher.ComposeTemplate) vmLauncher.ComposeTemplate {
	var newVMTemplate vmLauncher.ComposeTemplate

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	for _, vm := range c.VMs {
		// If no working directory is specified in the compose file,
		// use defaults.
		if vm.Workdir == "" {
			vm.Workdir = getWorkDir()
		}
		// Check if any flavor is provided

		if vm.Flavor != "" {
			vm.Size = vmLauncher.GetVMSizeFromFlavor(vm.Flavor)
		}
		newVMTemplate.VMs = append(newVMTemplate.VMs,
			vmLauncher.CreateVM(
				vm.Name,
				vm.ParentImage,
				vm.Workdir,
				vm.SSHKey,
				vm.UserData,
				vmLauncher.NewVMSize(vm.Size.CPUModel,
					vm.Size.Sockets,
					vm.Size.Cpus,
					vm.Size.Cores,
					vm.Size.Threads,
					vm.Size.RAM,
				),
				vm.Cloud,
				vm.Efi,
				vm.NetOpts,
				vm.Shares,
			),
		)
	}

	/* Docker network definitions */
	if len(c.Networks) > 0 {
		for _, net := range c.Networks {
			_, err := cli.NetworkCreate(ctx, net.Name, types.NetworkCreate{
				CheckDuplicate: true,
				IPAM: &network.IPAM{
					Config: []network.IPAMConfig{
						{Subnet: net.Subnet},
					},
				},
				Options: map[string]string{
					"com.docker.network.bridge.enable_icc":           "true",
					"com.docker.network.bridge.enable_ip_masquerade": "true",
					"com.docker.network.bridge.host_binding_ipv4":    "0.0.0.0",
					"com.docker.network.bridge.name":                 net.Name,
					"com.docker.network.driver.mtu":                  "1500",
				},
			})
			if err != nil {
				glog.Info(err)
			}
			glog.Flush()
		}
	}

	return newVMTemplate
}

func getHomeDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		log.Warn("Unable to determine $HOME")
		log.Error("Please specify -workdir and -pubkey")
	}
	return home
}

func getWorkDir() string {

	homeDir := getHomeDir()
	workDir := strings.Replace(VMLauncherWorkdir, "$HOME", homeDir, 1)
	workDir, _ = filepath.Abs(workDir)
	_, err := os.Stat(workDir)
	if err != nil {
		log.WithField("workdir", workDir).Warn(
			"Work Directory does not exist")

		log.WithField("workdir", workDir+"/data").Info(
			"Creating workdir")
		err = os.MkdirAll(workDir+"/data", 0755) // nolint: gas
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Creating %s", workDir+"/images")
		err = os.Mkdir(workDir+"/images", 0755) // nolint: gas
		if err != nil {
			log.Fatal(err)
		}
	}

	return workDir
}
