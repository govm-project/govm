package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

const (
	WORKDIR   = "$HOME/govm"
	SSHPUBKEY = "$HOME/.ssh/id_rsa.pub"
	IMAGE     = "$PWD/image.qcow2"
)

type VMSyze string

const (
	LargeVM  VMSyze = "largeVM"
	MediumVM VMSyze = "mediumVM"
	SmallVM  VMSyze = "smallVM"
	TinyVM   VMSyze = "tinyVM"
)

type MetaData struct {
	AvailabilityZone string            `json:"availability_zone"`
	Hostname         string            `json:"hostname"`
	LaunchIndex      int               `json:"launch_index"`
	Name             string            `json:"name"`
	Meta             map[string]string `json:"meta"`
	PublicKey        map[string]string `json:"public_keys"`
	UUID             string            `json:"uuid"`
}

var image string
var cowImage string
var name string
var small bool
var large bool
var tiny bool
var size VMSyze = MediumVM
var efi bool
var cidata string // Cloud init iso for running cloud images.
var cloud bool
var workdir string
var vmDir string
var verbose bool
var resize int
var userData string
var sshKeyPath string

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
	// Docker arguments
	command := fmt.Sprintf("run --name %v -td --privileged ", name)
	command += fmt.Sprintf("-v %v:%v ", image, image)
	command += fmt.Sprintf("-v %v/%v.img:/image/image -e AUTO_ATTACH=yes ", vmDir, name)
	command += fmt.Sprintf("-v %v:%v ", vmDir, vmDir)
	if cloud {
		command += fmt.Sprintf("-v %v:/cidata.iso ", cidata)
	}

	// Qemu arguments, passed to the container.
	command += fmt.Sprintf("verbacious/govm -vnc unix:%v/vnc ", vmDir)
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
	if verbose {
		fmt.Println(splitted_command)
	}
	err := exec.Command("docker", splitted_command...).Run()
	if err != nil {
		fmt.Printf("Error on docker command: %v", err)
		os.Exit(1)
	}

}

func showInfo() {
	vmIPArgs := fmt.Sprintf("inspect -f {{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}} %v", name)
	svmIPArgs := strings.Split(vmIPArgs, " ")
	vmIPCommand, err := exec.Command("docker", svmIPArgs...).Output()
	if err != nil {
		fmt.Println("Error running docker inspect command")
		fmt.Println(err)
	}
	vmIP := string(vmIPCommand)
	fmt.Println("[" + name + "]" + " info:\n" + "IP " + vmIP)
}

func genCiData(mdfile string, name string, hostname string, pk string) {
	// Customize meta_data.json file
	md_file, err := ioutil.ReadFile(mdfile)
	if err != nil {
		fmt.Printf("File error, reading the meta_data.json file: %v\n", err)
		os.Exit(1)
	}
	var md MetaData
	err = json.Unmarshal(md_file, &md)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	md.Hostname = hostname
	md.Name = name
	md.PublicKey["mykey"] = pk
	m, err := json.Marshal(&md)
	err = ioutil.WriteFile(mdfile, m, 0666)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}

	// genCiData takes the directory cloud-init to generate a cloud-init iso.
	command_args := fmt.Sprintf("--output %v/cidata.iso -volid config-2 -joliet -rock %v/cloud-init", vmDir, vmDir)
	sc := strings.Split(command_args, " ")
	err = exec.Command("genisoimage", sc...).Run()
	if err != nil {
		fmt.Println("Failed to create cidata.iso")
		fmt.Println(err)
		os.Exit(1)
	}
}

func resizeImage() {
	if verbose {
		fmt.Printf("resizing %v +%vG\n", image, resize)
	}
	command_args := fmt.Sprintf("resize %v +%vG", image, resize)
	sc := strings.Split(command_args, " ")
	err := exec.Command("qemu-img", sc...).Run()
	if err != nil {
		fmt.Println("Failed to resize image")
		fmt.Println(err)
		os.Exit(1)
	}
}

func saneImage(path string) error {

	// Test if the image file exists
	imgArg, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%v does not exist", path)
	}

	// Test if the image is valid or has a valid path
	mode := imgArg.Mode()
	if !mode.IsRegular() {
		return fmt.Errorf("%v is not a regular file", path)
	}
	return nil
}

func prepare() error {
	return nil
}

func init() {
	flag.StringVar(&workdir, "workdir", WORKDIR, "govm working directory")
	flag.StringVar(&sshKeyPath, "pubkey", SSHPUBKEY, "Public SSH key for VM management")
	flag.StringVar(&image, "image", "image.qcow2", "qcow2 image file path")
	flag.StringVar(&name, "name", "", "VM's name")
	flag.BoolVar(&tiny, "tiny", false, "Tiny VM flavor (512MB ram, cpus=1,cores=1,threads=1)")
	flag.BoolVar(&small, "small", false, "Small VM flavor (2G ram, cpus=4,cores=2,threads=2)")
	flag.BoolVar(&large, "large", false, "Small VM flavor (8G ram, cpus=8,cores=4,threads=4)")
	flag.BoolVar(&efi, "efi", false, "EFI-enabled vm (Optional)")
	flag.BoolVar(&cloud, "cloud", false, "Cloud VM (Optional)")
	flag.BoolVar(&verbose, "v", false, "Enable verbosity")
	flag.IntVar(&resize, "resize", 0, "Resize value in GB (Only for QCOW Images).")
	flag.StringVar(&userData, "user-data", "", "User-provided user_data file")
}

func main() {

	flag.Parse()

	home := os.Getenv("HOME")
	if home == "" {
		fmt.Printf("\nUnable to determine $HOME\n")
		fmt.Printf("Please specify -workdir and -pubkey\n")
		flag.Usage()
		os.Exit(1)
	}
	wdir := strings.Replace(WORKDIR, "$HOME", home, 1)
	keyPath := strings.Replace(SSHPUBKEY, "$HOME", home, 1)

	image, err := filepath.Abs(image)
	if err != nil {
		fmt.Printf("Unable to determine image location: %v\n", err)
		os.Exit(1)
	}

	err = saneImage(image)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Check sane working directory
	wdir, _ = filepath.Abs(wdir)
	_, err = os.Stat(wdir)
	if err != nil {
		fmt.Printf(" %v does not exists\n", wdir)
		fmt.Printf("Run the setup.sh first or try:\n\n\tmkdir -p %s\n", wdir)
		os.Exit(1)
	}

	// Check existing container's name
	dockerName := exec.Command("docker", "inspect", name).Run()
	if dockerName == nil {
		fmt.Printf("There is a %s container already running", name)
		os.Exit(1)
	}

	// Resize Image
	if resize != 0 {
		resizeImage()
	}

	// Perform a copy-on-write
	// First create the VM data directory
	UUIDcmd, err := exec.Command("uuidgen").Output()
	vmUUID := strings.TrimSpace(string(UUIDcmd))
	vmDir = wdir + "/data" + "/" + name + "/" + vmUUID
	_, err = os.Stat(vmDir)
	if err != nil {
		err = os.MkdirAll(vmDir, 0777)
		if err != nil {
			fmt.Println("Unable to create the hold dir for the cow image")
			fmt.Println(err)
			os.Exit(1)
		}
	}
	// Create copy-on-write image
	cowArgs := fmt.Sprintf("create -f qcow2 -o backing_file=%v temp.img", image)
	splittedCowArgs := strings.Split(cowArgs, " ")
	if verbose {
		fmt.Println(splittedCowArgs)
	}
	err = exec.Command("qemu-img", splittedCowArgs...).Run()
	if err != nil {
		fmt.Println("Unable to create the cow image")
		fmt.Println(err)
		os.Exit(1)
	}
	cowImage = vmDir + "/" + name + ".img"
	output, err = exec.Command("mv", "temp.img", cowImage).Run()

	// Handle cloud argument argument
	if cloud {
		// Copy cloud-init directory to vmDir
		cpArgs := fmt.Sprintf("-r %v/cloud-init %v", wdir, vmDir)
		splittedCpArgs := strings.Split(cpArgs, " ")
		output, err := exec.Command("cp", splittedCpArgs...).CombinedOutput()
		if err != nil {
			fmt.Println("Unable to copy cloud-init directory")
			fmt.Println(err)
			os.Exit(1)
		}
		// Check if a user_data file is provided
		if userData != "" {
			userData, _ = filepath.Abs(userData)
			// Copy user-provided user data to the VM dir
			cpArgs = fmt.Sprintf("-f %v %v/cloud-init/openstack/latest/user_data", userData, vmDir)
			splittedCpArgs = strings.Split(cpArgs, " ")
			err = exec.Command("cp", splittedCpArgs...).Run()
			if err != nil {
				fmt.Println("Unable to copy the user_data file")
				fmt.Println(err)
				os.Exit(1)
			}
		}
		// Generate a cloud-init iso
		cidata, _ = filepath.Abs(vmDir + "/cidata.iso")
		hostname := name
		publicKey, err := ioutil.ReadFile(keyPath)
		if err != nil {
			fmt.Printf("File error: %v", err)
			os.Exit(1)
		}
		genCiData(vmDir+"/cloud-init/openstack/latest/meta_data.json", name, hostname, string(publicKey))
	}
	startVM()
	showInfo()
}
