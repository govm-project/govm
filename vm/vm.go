package vm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/moby/moby/pkg/namesgenerator"

	"github.com/govm-project/govm/engines/docker"
	vmTypes "github.com/govm-project/govm/types"
)

var vncPort string

//VM contains all VM's attributes
type VM struct {
	Name             string                    `yaml:"name"`
	ParentImage      string                    `yaml:"image"`
	Size             vmTypes.VMSize            `yaml:"size"`
	Workdir          string                    `yaml:"workdir"`
	SSHKey           string                    `yaml:"sshkey"`
	UserData         string                    `yaml:"user-data"`
	Cloud            bool                      `yaml:"cloud"`
	Efi              bool                      `yaml:"efi"`
	generateUserData bool                      `yaml:"userdata"`
	containerID      string                    `yaml:"container-id"`
	NetOpts          vmTypes.NetworkingOptions `yaml:"networking"`
	Shares           []string                  `yaml:"shares"`
}

// CreateVM creates a new VM object
// TODO: Reduce cyclomatic complexity
func CreateVM( // nolint: gocyclo
	name,
	parentImage,
	workdir,
	publicKey,
	userData string,
	size vmTypes.VMSize,
	cloud, efi bool,
	netOpts vmTypes.NetworkingOptions,
	shares []string) VM {

	var vm VM
	var err error

	if parentImage == "" {
		fmt.Println("Missing --image argument")
		os.Exit(1)
	}
	vm.ParentImage, err = filepath.Abs(parentImage)
	if err != nil {
		fmt.Printf("Unable to determine image location: %v\n", err)
		os.Exit(1)
	}
	err = SaneImage(parentImage)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Optional Flags
	if name != "" {
		vm.Name = name
	} else {
		vm.Name = namesgenerator.GetRandomName(0)
	}

	client := docker.NewDockerClient()
	if err != nil {
		panic(err)
	}
	_, err = client.Inspect(name)
	if err == nil {
		log.Fatal("There is an existing container with the same name")
	}
	vm.Workdir = workdir

	// Check if user data is provided
	if userData != "" {
		absUserData, err := filepath.Abs(userData)
		if err != nil {
			fmt.Printf("Unable to determine %v user data file location: %v\n", vm, err)
			os.Exit(1)
		}
		// Test if the template file exists
		_, err = os.Stat(absUserData)
		if err != nil {
			// Look for a script verifying the shebang
			var validShebang bool
			validShebangs := []string{
				"#cloud-config",
				"#!/bin/sh",
				"#!/bin/bash",
				"#!/usr/bin/env python",
			}
			_, shebang, _ := bufio.ScanLines([]byte(userData), true)
			for _, sb := range validShebangs {
				if string(shebang) == sb {
					validShebang = true
				}
			}
			if validShebang {
				vm.generateUserData = true
				vm.UserData = userData
			} else {
				fmt.Println("Unable to determine the user data content")
				os.Exit(1)
			}

		} else {
			vm.UserData = absUserData
		}
	}

	// Check if any flavor is provided
	if size != (vmTypes.VMSize{}) {
		vm.Size = size
	} else {
		vm.Size = GetVMSizeFromFlavor("")
	}

	vm.Efi = efi
	vm.Cloud = cloud

	if publicKey != "" {
		key, err := ioutil.ReadFile(publicKey)
		if err != nil {
			log.Fatal(err)
		}
		vm.SSHKey = string(key)
	} else {
		homeDir := getHomeDir()
		keyPath := strings.Replace(SSHPublicKeyFile, "$HOME", homeDir, 1)
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			log.Fatal(err)
		}
		vm.SSHKey = string(key)
	}

	// Check if there are any VM Shares (shared directories) and validate them
	// They must be separated by the ":" character as docker does.
	if len(shares) > 0 {
		for _, dir := range vm.Shares {
			share := strings.Split(dir, ":")
			// Validate if the host share exists
			stat, err := os.Stat(share[0])
			if err != nil {
				log.WithFields(log.Fields{
					"share": share[0],
				}).Fatal("Host directory does not exists.")
			}

			// Validate if it's a directory
			if !stat.IsDir() {
				log.WithFields(log.Fields{
					"share": share[0],
				}).Fatal("Host field is not a directory.")
			}

		}
		vm.Shares = shares
	}

	vm.NetOpts = netOpts
	if vm.NetOpts.NetID == "" {
		vm.NetOpts.NetID = "bridge"
		vm.NetOpts.IP = ""
	}

	return vm
}

// ShowInfo shows VM's information
func (vm *VM) ShowInfo() {

	client := docker.NewDockerClient()
	containerInfo, err := client.Inspect(vm.containerID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[%s]\nIP Address: %s\n", containerInfo.Name[1:],
		containerInfo.NetworkSettings.DefaultNetworkSettings.IPAddress)

}

func (vm *VM) setVNC(vmName string, port string) error {

	client := docker.NewDockerClient()
	err := client.PullImage(VNCContainerImage)
	if err != nil {
		return err
	}

	_, err = client.Search(VNCServerContainerName)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err.Error(),
			"container": VNCContainerImage,
		}).Warn("VNC container was not found")

		mountBinds := []string{
			fmt.Sprintf("%v/data:/vm", vm.Workdir)}

		containerConfig := &container.Config{
			Image:    VNCContainerImage,
			Hostname: vm.Name,
			Cmd:      nil,
			Env:      nil,
			Labels:   nil,
		}

		hostConfig := &container.HostConfig{
			Privileged:      true,
			PublishAllPorts: true,
			NetworkMode:     "host",
			Binds:           mountBinds,
		}
		_, err := client.Create(containerConfig, hostConfig,
			&network.NetworkingConfig{}, VNCServerContainerName)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	execCmd := []string{"/noVNC/utils/websockify/run",
		"--web", "/noVNC",
		"--unix-target", "/vm/" + vmName + "/vnc",
		port}

	execConfig := types.ExecConfig{
		Detach: true,
		Cmd:    execCmd,
	}

	err = client.Exec(VNCServerContainerName, execConfig)

	return err
}

// Launch executes the required commands to launch a new VM
// TODO: Reduce cyclomatic complexity
func (vm *VM) Launch() { // nolint: gocyclo
	// Create the data dir
	vmDataDirectory := vm.Workdir + "/data/" + vm.Name
	err := os.MkdirAll(vmDataDirectory, 0740) // nolint: gas
	if err != nil {
		fmt.Printf("Unable to create: %s", vmDataDirectory)
		os.Exit(1)
	}

	// Create the metadata file
	vmMetaData := vmTypes.ConfigDriveMetaData{
		"vm",
		vm.Name,
		"0",
		vm.Name,
		map[string]string{},
		map[string]string{
			"mykey": vm.SSHKey,
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

	// Create the user_data file
	if vm.generateUserData {
		err = ioutil.WriteFile(vmDataDirectory+"/user_data",
			[]byte(vm.UserData), 0664)
		if err != nil {
			log.Fatal(err)
		}
		vm.UserData = vmDataDirectory + "/user_data"
	}

	// Default Environment Variables
	env := []string{
		"AUTO_ATTACH=yes",
		"DEBUG=yes",
		fmt.Sprintf(KVMCPUOpts,
			vm.Size.CPUModel,
			vm.Size.Sockets,
			vm.Size.Cpus,
			vm.Size.Cores,
			vm.Size.Threads,
			(vm.Size.Sockets * vm.Size.Cores * vm.Size.Threads),
			vm.Size.RAM,
		),
	}

	//if hostDNS {
	//	env = append(env, "ENABLE_DHCP=no")
	//}

	/* QEMU ARGUMENTS PASSED TO THE CONTAINER */
	qemuParams := []string{
		"-vnc unix:/data/vnc",
	}
	if vm.Efi {
		qemuParams = append(qemuParams, "-bios /OVMF.fd ")
	}
	if vm.Cloud {
		env = append(env, "CLOUD=yes")
		env = append(env, CloudInitOpts)
	}

	// Default Mount binds
	defaultMountBinds := []string{
		fmt.Sprintf(ImageMount, vm.ParentImage),
		fmt.Sprintf(DataMount, vmDataDirectory),
		fmt.Sprintf(MetadataMount, vmDataDirectory, MedatataFile),
	}
	// Append shares to defaultMountBinds if any.
	// Append guest directory/ies to env (container's environment).
	if len(vm.Shares) > 0 {
		defaultMountBinds = append(defaultMountBinds, vm.Shares...)
		var sharedDirs string
		for _, share := range vm.Shares {
			// Get the guest directory part from the string and pass
			// it to the container environment. The "startvm" script
			// will handle these shared directories.
			sharedDirs += strings.Split(share, ":")[1] + " "
		}
		env = append(env, "SHARED_DIRS="+sharedDirs)
	}

	if vm.UserData != "" {
		defaultMountBinds = append(defaultMountBinds,
			fmt.Sprintf(UserDataMount, vm.UserData))
	}

	// Create the Docker API client
	client := docker.NewDockerClient()
	err = client.PullImage(VMLauncherContainerImage)
	if err != nil {
		panic(err)
	}

	// Get an available port for VNC
	vncPort = strconv.Itoa(findPort())

	// Create the Container
	containerConfig := &container.Config{
		Image:    VMLauncherContainerImage,
		Hostname: vm.Name,
		Cmd:      qemuParams,
		Env:      env,
		Labels: map[string]string{
			"websockifyPort": vncPort,
			"dataDir":        vmDataDirectory,
		},
	}

	hostConfig := &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: true,
		Binds:           defaultMountBinds,
		DNS:             vm.NetOpts.DNS,
	}

	if vm.NetOpts.IP != "" {
		containerConfig.Labels["ip"] = vm.NetOpts.IP
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			vm.NetOpts.NetID: {
				IPAddress:  vm.NetOpts.IP,
				MacAddress: vm.NetOpts.MAC,
				NetworkID:  vm.NetOpts.NetID,
			},
		},
	}

	// TODO - Proper error handling
	vm.containerID, err = client.Create(containerConfig, hostConfig,
		networkConfig, vm.Name)
	if err != nil {
		log.Fatal(err)
	}

	err = vm.setVNC(vm.Name, vncPort)
	if err != nil {
		// TODO: Change to warning when the log package is changed
		log.Println(err)
	}
}
