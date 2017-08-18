package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/moby/moby/pkg/namesgenerator"

	"github.intel.com/clrgdc/govm/docker"
)

var vncPort string

//NetworkingOptions specifies network details for new VM
type NetworkingOptions struct {
	IP    string   `yaml:"ip"`
	MAC   string   `yaml:"mac"`
	NetID string   `yaml:"net-id"`
	DNS   []string `yaml:"dns"`
}

//VM contains all VM's attributes
type VM struct {
	Name        string            `yaml:"name"`
	ParentImage string            `yaml:"image"`
	Size        VMSize            `yaml:"size"`
	Cloud       bool              `yaml:"cloud"`
	Efi         bool              `yaml:"efi"`
	Workdir     string            `yaml:"workdir"`
	SSHKey      string            `yaml:"sshkey"`
	UserData    string            `yaml:"user-data"`
	NetOpts     NetworkingOptions `yaml:"networking"`

	containerID      string
	generateUserData bool
}

//NewVM creates a new VM object
func NewVM(name, parentImage, workdir, publicKey, userData string, size VMSize,
	cloud, efi bool, netOpts NetworkingOptions) VM {
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
	err = saneImage(parentImage)
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

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	_, err = cli.ContainerInspect(ctx, name)
	if err == nil {
		log.Fatal("There is an existing container with the same name")
	}
	// Check the workdir
	if workdir != "" {
		vm.Workdir = workdir
	} else {
		vm.Workdir = wdir
	}

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
	if size != (VMSize{}) {
		vm.Size = size
	} else {
		vm.Size = GetVMSizeFromFlavor("")
	}

	// Check if efi flag is provided
	if efi {
		vm.Efi = efi
	}

	// Check if cloud flag is provided
	if cloud {
		vm.Cloud = cloud
	}

	if publicKey != "" {
		key, err := ioutil.ReadFile(publicKey)
		if err != nil {
			log.Fatal(err)
		}
		vm.SSHKey = string(key)
	} else {
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			log.Fatal(err)
		}
		vm.SSHKey = string(key)
	}

	vm.NetOpts.IP = netOpts.IP
	vm.NetOpts.MAC = netOpts.MAC
	vm.NetOpts.NetID = netOpts.NetID
	vm.NetOpts.DNS = netOpts.DNS

	return vm
}

// ShowInfo shows VM's information
func (vm *VM) ShowInfo() {
	ctx := context.Background()

	// Create the Docker API client
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containerInfo, _ := cli.ContainerInspect(ctx, vm.containerID)
	fmt.Printf("[%s]\nIP Address: %s\n", containerInfo.Name[1:],
		containerInfo.NetworkSettings.DefaultNetworkSettings.IPAddress)

}

func (vm *VM) setVNC(vmName string, port string) error {
	ctx := context.Background()

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	err = docker.PullImage(ctx, cli, VNCContainerImage)
	if err != nil {
		return err
	}

	err = docker.ContainerSearch(ctx, cli, VNCContainerImage)
	if err != nil {
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
		_, err := docker.Run(ctx, cli, containerConfig, hostConfig,
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

	err = docker.Exec(ctx, cli, VNCServerContainerName, execConfig)

	return err
}

//Launch executes the required commands to launch a new VM
func (vm *VM) Launch() {
	ctx := context.Background()

	// Create the data dir
	vmDataDirectory := vm.Workdir + "/data/" + vm.Name
	err := os.MkdirAll(vmDataDirectory, 0740)
	if err != nil {
		fmt.Printf("Unable to create: %s", vmDataDirectory)
		os.Exit(1)
	}

	// Create the metadata file
	vmMetaData := ConfigDriveMetaData{
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

	// Default Enviroment Variables
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
	if hostDNS {
		env = append(env, "ENABLE_DHCP=no")
	}

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

	if vm.UserData != "" {
		defaultMountBinds = append(defaultMountBinds,
			fmt.Sprintf(UserDataMount, vm.UserData))
	}

	// Create the Docker API client
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	err = docker.PullImage(ctx, cli, VMLauncherContainerImage)
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
			vm.NetOpts.NetID: &network.EndpointSettings{
				IPAddress:  vm.NetOpts.IP,
				MacAddress: vm.NetOpts.MAC,
				NetworkID:  vm.NetOpts.NetID,
			},
		},
	}

	// TODO - Proper error handling
	vm.containerID, err = docker.Run(ctx, cli, containerConfig, hostConfig,
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
