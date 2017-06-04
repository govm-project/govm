package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var vncPort string

type GoVM struct {
	Name        string   `yaml:"name"`
	Size        HostOpts `yaml:"size"`
	ParentImage string   `yaml:"image"`
	Cloud       bool     `yaml:"cloud"`
	Efi         bool     `yaml:"efi"`
	Workdir     string   `yaml:"workdir"`
	SSHKey      string   `yaml:"sshkey"`
	UserData    string   `yaml:"user-data"`

	containerID      string
	generateUserData bool
}

func NewGoVM(name, parentImage string, size HostOpts, cloud, efi bool, workdir string, publicKey string, userData string) GoVM {
	var govm GoVM
	govm.Name = name
	govm.Size = size
	govm.ParentImage = parentImage
	govm.Cloud = cloud
	govm.Efi = efi
	govm.Workdir = workdir
	govm.SSHKey = publicKey
	govm.UserData = userData

	return govm
}

func (govm *GoVM) ShowInfo() {
	ctx := context.Background()

	// Create the Docker API client
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containerInfo, _ := cli.ContainerInspect(ctx, govm.containerID)
	fmt.Printf("[%s]\nIP Address: %s\n", containerInfo.Name[1:], containerInfo.NetworkSettings.DefaultNetworkSettings.IPAddress)

}

func (govm *GoVM) setVNC(govmName string, port string) {
}

func (govm *GoVM) Launch() {
	ctx := context.Background()

	// Create the data dir
	vmDataDirectory := govm.Workdir + "/data/" + govm.Name
	err := os.MkdirAll(vmDataDirectory, 0740)
	if err != nil {
		fmt.Printf("Unable to create: %s", vmDataDirectory)
		os.Exit(1)
	}

	// Create the metadata file
	vmMetaData := ConfigDriveMetaData{
		"govm",
		govm.Name,
		"0",
		govm.Name,
		map[string]string{},
		map[string]string{
			"mykey": govm.SSHKey,
		},
		"0",
	}

	vmMetaDataJSON, err := json.Marshal(vmMetaData)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(vmDataDirectory+"/meta_data.json", vmMetaDataJSON, 0664)
	if err != nil {
		log.Fatal(err)
	}

	if govm.generateUserData == true {
		// Dump user data into a file
		err = ioutil.WriteFile(vmDataDirectory+"/user_data", []byte(govm.UserData), 0664)
		if err != nil {
			log.Fatal(err)
		}
		govm.UserData = vmDataDirectory + "/user_data"
	}

	// Default Enviroment Variables
	env := []string{
		"AUTO_ATTACH=yes",
		"DEBUG=yes",
		fmt.Sprintf("KVM_CPU_OPTS=%v", govm.Size),
	}
	if host_dns {
		env = append(env, "ENABLE_DHCP=no")
	}

	/* QEMU ARGUMENTS PASSED TO THE CONTAINER */
	qemuParams := []string{
		"-vnc unix:/data/vnc",
	}
	if govm.Efi {
		qemuParams = append(qemuParams, "-bios /OVMF.fd ")
	}
	if govm.Cloud {
		env = append(env, "CLOUD=yes")
		env = append(env, "CLOUD_INIT_OPTS=-drive file=/data/seed.iso,if=virtio,format=raw ")
	}

	// Default Mount binds
	defaultMountBinds := []string{
		fmt.Sprintf("%v:/image/image", govm.ParentImage),
		fmt.Sprintf("%v:/data", vmDataDirectory),
		fmt.Sprintf("%v:/cloud-init/openstack/latest/meta_data.json", vmDataDirectory+"/meta_data.json"),
	}

	if govm.UserData != "" {
		defaultMountBinds = append(defaultMountBinds, fmt.Sprintf("%s:/cloud-init/openstack/latest/user_data", govm.UserData))
	}

	// Create the Docker API client
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// Get the verbacious/govm image

	//_, err = cli.ImagePull(ctx, "govm", types.ImagePullOptions{})
	//if err != nil {
	//	panic(err)
	//}

	/* WIP Exposed Ports
	// Default Ports
	var ports nat.PortMap
	var exposedPorts nat.PortSet
	vncPort := "5910"
	_, ports, _ = nat.ParsePortSpecs([]string{
		fmt.Sprintf(":%v:%v", vncPort, vncPort),
	})

	exposedPorts = map[nat.Port]struct{}{
	      "5910/tcp": {},
	}
	*/

	// Get an available port for VNC
	vncPort = strconv.Itoa(findPort())

	// Create the Container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "verbacious/govm",
		Cmd:   qemuParams,
		//Cmd:          []string{"top"},
		Env: env,
		Labels: map[string]string{
			"websockifyPort": vncPort,
			"dataDir":        vmDataDirectory,
		},
		//ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: true,
		Binds:           defaultMountBinds,
		//PortBindings:    ports,
	}, nil, govm.Name)
	if err != nil {
		panic(err)
	}

	govm.containerID = resp.ID

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	govm.setVNC(govm.Name, vncPort)
}
