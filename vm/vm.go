package vm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/govm-project/govm/internal"

	log "github.com/sirupsen/logrus"
)

// Size specifies all custome VM size fields
type Size struct {
	CPUModel string `yaml:"cpu-model"`
	Sockets  int    `yaml:"sockets"`
	Cpus     int    `yaml:"cpus"`
	Cores    int    `yaml:"cores"`
	Threads  int    `yaml:"threads"`
	RAM      int    `yaml:"ram"`
	DISK     int    `yaml:"disk"`
}

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

//NetworkingOptions specifies network details for new VM
type NetworkingOptions struct {
	IP    string   `yaml:"ip"`
	MAC   string   `yaml:"mac"`
	NetID string   `yaml:"net-id"`
	DNS   []string `yaml:"dns"`
}

// ComposeConfig defines a VMs orchestration template
type ComposeConfig struct {
	VMs       []Instance `yaml:"vms"`
	Namespace string     `yaml:"namespace"`
}

//Instance contains all VM's attributes
type Instance struct {
	ID               string            `yaml:"id"`
	Name             string            `yaml:"name"`
	Namespace        string            `yaml:"namespace"`
	ParentImage      string            `yaml:"image"`
	Flavor           string            `yaml:"flavor"`
	Size             Size              `yaml:"size"`
	Workdir          string            `yaml:"workdir"`
	SSHPublicKeyFile string            `yaml:"sshkey"`
	UserData         string            `yaml:"user-data"`
	Cloud            bool              `yaml:"cloud"`
	Efi              bool              `yaml:"efi"`
	VNCPort          int64             `yaml:"vnc-port"`
	NetOpts          NetworkingOptions `yaml:"network"`
	Shares           []string          `yaml:"shares"`
	ContainerEnvVars []string          `yaml:"ContainerEnvVars"`
}

// Check validates and fixes VMs values
// nolint: gocyclo, funlen, gocognit
func (ins *Instance) Check() (err error) {
	ins.ParentImage, err = internal.CheckFilePath(ins.ParentImage)
	if err != nil {
		return
	}

	if ins.Name == "" {
		ins.Name = strings.Replace(internal.RandomName(), "_", "-", 1)
	}

	vmDataDirectory := ins.Workdir + "/data/" + ins.Name
	err = os.MkdirAll(vmDataDirectory, 0740) // nolint: gas
	if err != nil {
		err = fmt.Errorf("unable to create: %s", vmDataDirectory)
		return
	}

	// Check if user data is provided
	if ins.UserData != "" {
		// By default handle UserData as user_data file content.
		userData := []byte(ins.UserData)
		// But if UserData is a name of an existing file, take
		// contents from that file.
		ins.UserData, err = internal.CheckFilePath(ins.UserData)
		if err == nil {
			userData, _ = os.ReadFile(ins.UserData)

		}
		// Look for a script verifying the shebang
		var validShebang bool

		validShebangs := []string{
			"#cloud-config",
			"#!/bin/sh",
			"#!/bin/bash",
			"#!/usr/bin/env python",
		}

		_, shebang, _ := bufio.ScanLines(userData, true)

		for _, sb := range validShebangs {
			if string(shebang) == sb {
				validShebang = true
			}
		}

		if !validShebang {
			err = fmt.Errorf("unable to determine the user data content")
			return err
		}

		err = ioutil.WriteFile(vmDataDirectory+"/user_data",
			userData, 0664)
		if err != nil {
			log.Fatal(err)
		}

		ins.UserData = vmDataDirectory + "/user_data"
	}

	if ins.Size == (Size{}) {
		ins.Size = GetSizeFromFlavor(ins.Flavor)
	}

	if ins.SSHPublicKeyFile != "" {
		ins.SSHPublicKeyFile, err = internal.CheckFilePath(ins.SSHPublicKeyFile)
		if err != nil {
			return
		}
	} else {
		homeDir := internal.GetUserHomePath()
		ins.SSHPublicKeyFile = fmt.Sprintf("%v/.ssh/%v", homeDir, DefaultSSHPublicKeyFile)
	}

	key, err := ioutil.ReadFile(ins.SSHPublicKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	ins.SSHPublicKeyFile = string(key)

	// Check if there are any VM Shares (shared directories) and validate them
	if len(ins.Shares) > 0 {
		for _, dir := range ins.Shares {
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
	}

	if ins.NetOpts.NetID == "" {
		ins.NetOpts.NetID = "bridge"
		ins.NetOpts.IP = ""
	}

	// Create the metadata file
	metaData := ConfigDriveMetaData{
		AvailabilityZone: "vm",
		Hostname:         ins.Name,
		LaunchIndex:      "0",
		Name:             ins.Name,
		Meta:             map[string]string{},
		PublicKeys: map[string]string{
			"mykey": ins.SSHPublicKeyFile,
		},
		UUID: "0",
	}

	metaDataJSON, err := json.Marshal(metaData)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(vmDataDirectory+"/meta_data.json", metaDataJSON, 0664)
	if err != nil {
		return
	}

	// Create the user_data file

	return nil
}

//NewSize creates a new VMSize specification
func NewSize(model string, sockets, cpus, cores, threads, ram, disk int) Size {
	var vmSize Size

	if model != "" {
		vmSize.CPUModel = model
	} else if vmxSupport() {
		vmSize.CPUModel = "host"
	}

	if sockets != 0 {
		vmSize.Sockets = sockets
	} else {
		vmSize.Sockets = 1
	}

	if cpus != 0 {
		vmSize.Cpus = cpus
	} else {
		vmSize.Cpus = 1
	}

	if cores != 0 {
		vmSize.Cores = cores
	} else {
		vmSize.Cores = 2
	}

	if threads != 0 {
		vmSize.Threads = threads
	} else {
		vmSize.Threads = 2
	}

	if ram != 0 {
		vmSize.RAM = ram
	} else {
		vmSize.RAM = 4096
	}

	if disk != 0 {
		vmSize.DISK = disk
	} else {
		vmSize.DISK = DiskDefaultSizeGB
	}

	return vmSize
}

//GetSizeFromFlavor gets default set of values from a given flavor
func GetSizeFromFlavor(flavor string) (size Size) {
	var cpuModel string

	if vmxSupport() {
		cpuModel = "host"
	} else {
		cpuModel = "haswell"
	}

	switch flavor {
	case "micro":
		size = NewSize(cpuModel, 1, 1, 1, 1, 512, DiskDefaultSizeGB)
	case "tiny":
		size = NewSize(cpuModel, 1, 1, 1, 1, 1024, 2*DiskDefaultSizeGB)
	case "small":
		size = NewSize(cpuModel, 1, 1, 2, 1, 2048, 4*DiskDefaultSizeGB)
	case "medium":
		size = NewSize(cpuModel, 2, 2, 2, 2, 4096, 8*DiskDefaultSizeGB)
	case "large":
		size = NewSize(cpuModel, 1, 4, 4, 4, 8192, 16*DiskDefaultSizeGB)
	default:
		size = NewSize(cpuModel, 2, 2, 2, 2, 4096, DiskDefaultSizeGB)
	}

	return
}

func vmxSupport() bool {
	err := exec.Command("grep", "-qw", "vmx", "/proc/cpuinfo").Run() // nolint: gas
	if err != nil {
		log.Error(err)
		return false
	}

	return true
}
