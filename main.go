package main

import (
	"flag"
	"fmt"
	"os/exec"
	"path/filepath"
)

type HostOpts string

const (
	LargeVmx    HostOpts = "--enable-kvm -m 8192 -smp cpus=4,cores=2,threads=2 -cpu host"
	MediumVmx   HostOpts = "--enable-kvm -m 4096 -smp cpus=4,cores=2,threads=2 -cpu host"
	SmallVmx    HostOpts = "--enable-kvm -m 2048 -smp cpus=4,cores=2,threads=2 -cpu host"
	LargeNoVmx  HostOpts = "-cpu Haswell -m 8192"
	MediumNoVmx HostOpts = "-cpu Haswell -m 4096"
	SmallNoVmx  HostOpts = "-cpu Haswell -m 2048"
)

type VMSyze string

const (
	LargeVM  VMSyze = "largeVM"
	MediumVM VMSyze = "mediumVM"
	SmallVM  VMSyze = "smallVM"
)

var imageFile string
var name string
var small bool
var large bool
var size VMSyze = MediumVM
var efi bool

func vmxSupport() bool {
	err := exec.Command("grep", "-qw", "vmx", "/proc/cpuinfo").Run()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true

}

func getHostOpts(s VMSyze) (opts HostOpts) {
	opts = MediumNoVmx
	if vmxSupport() {
		switch s {
		case LargeVM:
			opts = LargeVmx
		case SmallVM:
			opts = SmallVmx
		case MediumVM:
			opts = MediumVmx
		}
	} else {
		switch s {
		case LargeVM:
			opts = LargeNoVmx
		case SmallVM:
			opts = SmallNoVmx
		}
	}
	return opts
}

func startVM() {
	command := fmt.Sprintf("docker run --name %v -td --privileged ", name)
	command += fmt.Sprintf("-v %v:/image/image -e AUTO_ATTACH=yes ", imageFile)
	command += fmt.Sprintf("-v /var/lib/vms/%v:/var/lib/vms/%v ", name, name)
	command += fmt.Sprintf("obedmr/govm -vnc unix:/var/lib/vms/%v/vnc ", name)
	if efi {
		command += "-bios /OVMF.fd "
	}
	command += string(getHostOpts(size))
	fmt.Println(command)
}

func init() {
	flag.StringVar(&imageFile, "image", "image.qcow2", "qcow2 image file path")
	flag.StringVar(&name, "name", "", "VM's name")
	flag.BoolVar(&small, "small", false, "Small VM flavor (2G ram, cpus=4,cores=2,threads=2)")
	flag.BoolVar(&large, "large", false, "Small VM flavor (8G ram, cpus=8,cores=4,threads=4)")
	flag.BoolVar(&efi, "efi", false, "EFI-enabled vm (Optional)")

}

func main() {
	flag.Parse()
	imageFile, _ = filepath.Abs(imageFile)
	startVM()

}
