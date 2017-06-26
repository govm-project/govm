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

	"clrgitlab.intel.com/clr-cloud/vmgo/docker"
)

var vncPort string

type NetworkingOptions struct {
	IP    string `yaml"ip"`
	MAC   string `yaml:"mac"`
	NetID string `yaml:"net-id"`
}

type Vmgo struct {
	Name        string            `yaml:"name"`
	ParentImage string            `yaml:"image"`
	Size        HostOpts          `yaml:"size"`
	Cloud       bool              `yaml:"cloud"`
	Efi         bool              `yaml:"efi"`
	Workdir     string            `yaml:"workdir"`
	SSHKey      string            `yaml:"sshkey"`
	UserData    string            `yaml:"user-data"`
	NetOpts     NetworkingOptions `yaml:"networking"`

	containerID      string
	generateUserData bool
}

func NewVmgo(name, parentImage string, size HostOpts, cloud, efi bool, workdir string, publicKey string, userData string, netOpts NetworkingOptions) Vmgo {
	var vmgo Vmgo
	var err error

	if parentImage == "" {
		fmt.Println("Missing --image argument")
		os.Exit(1)
	}
	vmgo.ParentImage, err = filepath.Abs(parentImage)
	if err != nil {
		fmt.Printf("Unable to determine image location: %v\n", err)
		os.Exit(1)
	}
	err = saneImage(parentImage)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	/* Optional Flags */
	if name != "" {
		vmgo.Name = name
	} else {
		vmgo.Name = namesgenerator.GetRandomName(0)
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
		vmgo.Workdir = workdir
	} else {
		vmgo.Workdir = wdir
	}

	// Check if user data is provided
	if userData != "" {
		absUserData, err := filepath.Abs(userData)
		if err != nil {
			fmt.Printf("Unable to determine %s user data file location: %v\n", vmgo, err)
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
			if validShebang == true {
				vmgo.generateUserData = true
				vmgo.UserData = userData
			} else {
				fmt.Println("Unable to determine the user data content")
				os.Exit(1)
			}

		} else {
			vmgo.UserData = absUserData
		}
	}

	// Check if any flavor is provided
	if size != "" {
		vmgo.Size = getFlavor(string(size))
	} else {
		vmgo.Size = getFlavor("")
	}

	// Check if efi flag is provided
	if efi != false {
		vmgo.Efi = efi
	}

	// Check if cloud flag is provided
	if cloud != false {
		vmgo.Cloud = cloud

	}

	if publicKey != "" {
		key, err := ioutil.ReadFile(publicKey)
		if err != nil {
			log.Fatal(err)
		}
		vmgo.SSHKey = string(key)
	} else {
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			log.Fatal(err)
		}
		vmgo.SSHKey = string(key)
	}

	if netOpts != (NetworkingOptions{}) {
		vmgo.NetOpts.IP = netOpts.IP
		vmgo.NetOpts.MAC = netOpts.MAC
		vmgo.NetOpts.NetID = netOpts.NetID
	}

	return vmgo
}

func (vmgo *Vmgo) ShowInfo() {
	ctx := context.Background()

	// Create the Docker API client
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containerInfo, _ := cli.ContainerInspect(ctx, vmgo.containerID)
	fmt.Printf("[%s]\nIP Address: %s\n", containerInfo.Name[1:], containerInfo.NetworkSettings.DefaultNetworkSettings.IPAddress)

}

func (vmgo *Vmgo) setVNC(vmgoName string, port string) error {
	ctx := context.Background()

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	err = docker.PullImage(ctx, cli, "vmgo/novnc-server")
	if err != nil {
		return err
	}

	err = docker.ContainerSearch(ctx, cli, "vmgo-novnc")
	if err != nil {
		mountBinds := []string{
			fmt.Sprintf("%v/data:/vmgo", vmgo.Workdir)}

		containerConfig := &container.Config{
			Image:  "vmgo/novnc-server",
			Cmd:    nil,
			Env:    nil,
			Labels: nil,
		}

		hostConfig := &container.HostConfig{
			Privileged:      true,
			PublishAllPorts: true,
			NetworkMode:     "host",
			Binds:           mountBinds,
		}
		_, err := docker.Run(ctx, cli, containerConfig, hostConfig, &network.NetworkingConfig{}, "vmgo-novnc")
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	execCmd := []string{"/noVNC/utils/websockify/run",
		"--web", "/noVNC",
		"--unix-target", "/vmgo/" + vmgoName + "/vnc",
		port}

	execConfig := types.ExecConfig{
		Detach: true,
		Cmd:    execCmd,
	}

	err = docker.Exec(ctx, cli, "vmgo-novnc", execConfig)

	return err
}

func (vmgo *Vmgo) Launch() {
	ctx := context.Background()

	// Create the data dir
	vmDataDirectory := vmgo.Workdir + "/data/" + vmgo.Name
	err := os.MkdirAll(vmDataDirectory, 0740)
	if err != nil {
		fmt.Printf("Unable to create: %s", vmDataDirectory)
		os.Exit(1)
	}

	// Create the metadata file
	vmMetaData := ConfigDriveMetaData{
		"vmgo",
		vmgo.Name,
		"0",
		vmgo.Name,
		map[string]string{},
		map[string]string{
			"mykey": vmgo.SSHKey,
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
	if vmgo.generateUserData == true {
		err = ioutil.WriteFile(vmDataDirectory+"/user_data", []byte(vmgo.UserData), 0664)
		if err != nil {
			log.Fatal(err)
		}
		vmgo.UserData = vmDataDirectory + "/user_data"
	}

	// Default Enviroment Variables
	env := []string{
		"AUTO_ATTACH=yes",
		"DEBUG=yes",
		fmt.Sprintf("KVM_CPU_OPTS=%v", vmgo.Size),
	}
	if host_dns {
		env = append(env, "ENABLE_DHCP=no")
	}

	/* QEMU ARGUMENTS PASSED TO THE CONTAINER */
	qemuParams := []string{
		"-vnc unix:/data/vnc",
	}
	if vmgo.Efi {
		qemuParams = append(qemuParams, "-bios /OVMF.fd ")
	}
	if vmgo.Cloud {
		env = append(env, "CLOUD=yes")
		env = append(env, "CLOUD_INIT_OPTS=-drive file=/data/seed.iso,if=virtio,format=raw ")
	}

	// Default Mount binds
	defaultMountBinds := []string{
		fmt.Sprintf("%v:/image/image", vmgo.ParentImage),
		fmt.Sprintf("%v:/data", vmDataDirectory),
		fmt.Sprintf("%v:/cloud-init/openstack/latest/meta_data.json", vmDataDirectory+"/meta_data.json"),
	}

	if vmgo.UserData != "" {
		defaultMountBinds = append(defaultMountBinds, fmt.Sprintf("%s:/cloud-init/openstack/latest/user_data", vmgo.UserData))
	}

	// Create the Docker API client
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	err = docker.PullImage(ctx, cli, "vmgo/vmgo")
	if err != nil {
		panic(err)
	}

	// Get an available port for VNC
	vncPort = strconv.Itoa(findPort())

	// Create the Container
	containerConfig := &container.Config{
		Image: "vmgo/vmgo",
		Cmd:   qemuParams,
		Env:   env,
		Labels: map[string]string{
			"websockifyPort": vncPort,
			"dataDir":        vmDataDirectory,
		},
	}

	hostConfig := &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: true,
		Binds:           defaultMountBinds,
	}

	if vmgo.NetOpts.IP != "" {
		containerConfig.Labels["ip"] = vmgo.NetOpts.IP
	}

	networkConfig := &network.NetworkingConfig{
		map[string]*network.EndpointSettings{
			vmgo.NetOpts.NetID: &network.EndpointSettings{
				IPAddress:  vmgo.NetOpts.IP,
				MacAddress: vmgo.NetOpts.MAC,
				NetworkID:  vmgo.NetOpts.NetID,
			},
		},
	}

	// TODO - Proper error handling
	vmgo.containerID, err = docker.Run(ctx, cli, containerConfig, hostConfig, networkConfig, vmgo.Name)
	if err != nil {
		log.Fatal(err)
	}

	vmgo.setVNC(vmgo.Name, vncPort)
}
