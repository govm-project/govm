package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type HostOpts string

// HostOpts define the specs for the VMs
const (
	LargeVmx  HostOpts = "--enable-kvm -m 8192 -smp cpus=4,cores=2,threads=2 -cpu host"
	MediumVmx HostOpts = "--enable-kvm -m 4096 -smp cpus=4,cores=2,threads=2 -cpu host"
	SmallVmx  HostOpts = "--enable-kvm -m 2048 -smp cpus=4,cores=2,threads=2 -cpu host"
	TinyVmx   HostOpts = "--enable-kvm -m 512 -smp cpus=1,cores=1,threads=1 -cpu host"

	LargeNoVmx  HostOpts = "-cpu Haswell -m 8192"
	MediumNoVmx HostOpts = "-cpu Haswell -m 4096"
	SmallNoVmx  HostOpts = "-cpu Haswell -m 2048"
	TinyNoVmx   HostOpts = "-cpu Haswell -m 512"
)

type VMSize string

const (
	LargeVM  VMSize = "largeVM"
	MediumVM VMSize = "mediumVM"
	SmallVM  VMSize = "smallVM"
	TinyVM   VMSize = "tinyVM"
)

func vmxSupport() bool {
	err := exec.Command("grep", "-qw", "vmx", "/proc/cpuinfo").Run()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func getHostOpts(s VMSize) (opts HostOpts) {
	opts = MediumNoVmx
	if vmxSupport() {
		switch s {
		case LargeVM:
			opts = LargeVmx
		case SmallVM:
			opts = SmallVmx
		case MediumVM:
			opts = MediumVmx
		case TinyVM:
			opts = TinyVmx
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

func getFlavor(flavor string) HostOpts {
	var size HostOpts

	switch string(flavor) {
	case "tiny":
		size = getHostOpts(TinyVM)
	case "small":
		size = getHostOpts(SmallVM)
	case "medium":
		size = getHostOpts(MediumVM)
	case "large":
		size = getHostOpts(LargeVM)
	default:
		size = getHostOpts(MediumVM)
	}
	return size
}

func getCustomFlavor(flavor string) HostOpts {
	var opts string
	var cflavor []string

	cflavor = strings.Split(flavor, ",")
	if vmxSupport() {
		opts = fmt.Sprintf("--enable-kvm -smp cpus=%s,cores=%s,threads=%s -cpu host -m %s", cflavor[0], cflavor[0], cflavor[1], cflavor[2])
	} else {
		opts = fmt.Sprintf("-cpu Haswell -m %v", opts[2])
	}
	return HostOpts(opts)
}
