package main

import (
	"context"

	"clrgitlab.intel.com/clr-cloud/vmgo/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
)

type VmgoTemplate struct {
	VMs      []Vmgo           `yaml:"vms"`
	Networks []docker.Network `yaml:"networks"`
}

func NewVmgoTemplate(c *VmgoTemplate) VmgoTemplate {
	var newVmgoTemplate VmgoTemplate

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	/* VMs definitions */
	for _, vm := range c.VMs {
		newVmgoTemplate.VMs = append(newVmgoTemplate.VMs,
			NewVmgo(
				vm.Name,
				vm.ParentImage,
				NewVMSize(vm.Size.CpuModel,
					vm.Size.Sockets,
					vm.Size.Cpus,
					vm.Size.Cores,
					vm.Size.Threads,
					vm.Size.Ram,
				),
				vm.Cloud,
				vm.Efi,
				vm.Workdir,
				vm.SSHKey,
				vm.UserData,
				vm.NetOpts,
			),
		)
	}

	/* Docker network definitions */
	if len(c.Networks) > 0 {
		for _, net := range c.Networks {
			//err = docker.VerifyNetwork(ctx, cli, net.Name)

			if 1 != 1 {
				glog.Info(err)
			} else {
				_, err := cli.NetworkCreate(ctx, net.Name, types.NetworkCreate{
					CheckDuplicate: true,
					IPAM: &network.IPAM{
						Config: []network.IPAMConfig{
							network.IPAMConfig{Subnet: net.Subnet},
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
	}

	return newVmgoTemplate
}
