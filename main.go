package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type HostOpts string

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

type VMSyze string

const (
	LargeVM  VMSyze = "largeVM"
	MediumVM VMSyze = "mediumVM"
	SmallVM  VMSyze = "smallVM"
	TinyVM   VMSyze = "tinyVM"
)

var imageFile string
var cowImage string
var name string
var small bool
var large bool
var tiny bool
var size VMSyze = MediumVM
var efi bool
var cidata string // Cloud init iso for running cloud images.
var cloud bool

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

func startVM() {
	fmt.Println("Starting VM")
	// Docker arguments
	command := fmt.Sprintf("run --name %v -td --privileged ", name)
	command += fmt.Sprintf("-v %v:%v ", imageFile, imageFile)
	command += fmt.Sprintf("-v %v:/image/image -e AUTO_ATTACH=yes ", cowImage)
	command += fmt.Sprintf("-v /var/lib/vms/%v:/var/lib/vms/%v ", name, name)
	if cloud {
		command += fmt.Sprintf("-v %v:/cidata.iso ", cidata)
	}

	// Qemu arguments, passed to the container.
	command += fmt.Sprintf("obedmr/govm -vnc unix:/var/lib/vms/%v/vnc ", name)
	if efi {
		command += "-bios /OVMF.fd "
	}
	if cloud {
		command += "-drive file=/cidata.iso,if=virtio "
	}
	if tiny {
		size = TinyVM
		command += string(getHostOpts(size))
	} else {
		command += string(getHostOpts(size))
	}

	// Split the string command for it's execution. See os/exec.
	splitted_command := strings.Split(command, " ")
	fmt.Println(splitted_command)
	err := exec.Command("docker", splitted_command...).Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func genCiData() {
	// genCiData takes the directory cloud-init to generate a cloud-init iso.
	command_args := "--output cidata.iso -volid config-2 -joliet -rock cloud-init"
	sc := strings.Split(command_args, " ")
	err := exec.Command("genisoimage", sc...).Run()
	if err != nil {
		fmt.Println("Failed to create cidata.iso")
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&imageFile, "image", "image.qcow2", "qcow2 image file path")
	flag.StringVar(&name, "name", "", "VM's name")
	flag.BoolVar(&tiny, "tiny", false, "Tiny VM flavor (512MB ram, cpus=1,cores=1,threads=1)")
	flag.BoolVar(&small, "small", false, "Small VM flavor (2G ram, cpus=4,cores=2,threads=2)")
	flag.BoolVar(&large, "large", false, "Small VM flavor (8G ram, cpus=8,cores=4,threads=4)")
	flag.BoolVar(&efi, "efi", false, "EFI-enabled vm (Optional)")
	flag.BoolVar(&cloud, "cloud", false, "Cloud VM (Optional)")

}

func main() {
	flag.Parse()
	imageFile, _ = filepath.Abs(imageFile)
	// Test if the image file exists
	imgArg, err := os.Stat(imageFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Test if the image is valid or has a valid path
	mode := imgArg.Mode()
	if !mode.IsRegular() {
		fmt.Println("The image file provided is not valid")
		os.Exit(1)
	}

	// Test the working directory
	currentUser, _ := user.Current()
	workdir := currentUser.HomeDir + "/govm"
	_, err = os.Stat(workdir)
	if err != nil {
		fmt.Println("~/govm does not exists. Trying to create it.")
		// Try to create the dir
		err = os.Mkdir(workdir, 0777)
		if err != nil {
			os.Exit(1)
		} else {
			fmt.Println("Created " + workdir)
		}

	}

	// Handle cloud argument argument
	if cloud {
		cidata, _ = filepath.Abs("cidata.iso")
		genCiData()
	}

	// Perform a copy-on-write
	// First create the holding dir
	holdDir := workdir + "/" + name
	_, err = os.Stat(holdDir)
	if err != nil {
		err = os.Mkdir(holdDir, 0777)
		if err != nil {
			fmt.Println("Unable to create the hold dir for the cow image")
			fmt.Println(err)
		}
	}
	// COW
	cowArgs := fmt.Sprintf("create -f qcow2 -o backing_file=%v temp.img", imageFile)
	splittedCowArgs := strings.Split(cowArgs, " ")
	fmt.Println(splittedCowArgs)
	err = exec.Command("qemu-img", splittedCowArgs...).Run()
	if err != nil {
		fmt.Println("Unable to create the cow image")
		fmt.Println(err)
		os.Exit(1)
	}
	cowImage = holdDir + "/" + name + "01.img"
	err = exec.Command("mv", "temp.img", cowImage).Run()
	fmt.Println(imageFile)

	startVM()

}
