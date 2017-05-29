package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

var Port string

type GoVM struct {
	Name        string
	Size        HostOpts
	ParentImage string
	Cloud       bool
	Efi         bool
	Workdir     string

	containerID string
}

func NewGoVM(name, parentImage string, size HostOpts, cloud, efi bool, workdir string) GoVM {
	var govm GoVM
	govm.Name = name
	govm.Size = size
	govm.ParentImage = parentImage
	govm.Cloud = cloud
	govm.Efi = efi
	govm.Workdir = workdir

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

func (govm *GoVM) setVNC(dataDir string, port string) {
	vncFile := dataDir + "/vnc"
	exists := false

	/* Allow others to be able to read the vnc socket */
	for exists == false {
		if _, err := os.Stat(vncFile); !os.IsNotExist(err) {
			chmodString := fmt.Sprintf("chmod o+rw %s", vncFile)
			chmodCommand := strings.Split(chmodString, " ")
			err := exec.Command("sudo", chmodCommand...).Run()
			if err != nil {

				log.Fatal("Unable to chmod: ", err)
			}
			exists = true
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}

	/* Launch a websockify process */
	cmd := exec.Command("websockify", fmt.Sprintf("--unix-target=%s", vncFile), "0.0.0.0:"+port)
	websockifyPidFile, err := os.Create(dataDir + "/websockifyPid")
	if err != nil {
		log.Fatal("Failed to create websockify pid file ", err)
		log.Fatal("Failed to write websockify pid number: ", err)
	}
	defer websockifyPidFile.Close()

	err = cmd.Start()
	if err != nil {
		log.Fatal("Failed to run websockify: ", err)
	}

	websockifyPid := strconv.Itoa(cmd.Process.Pid)
	_, err = websockifyPidFile.WriteString(websockifyPid)
	if err != nil {
		log.Fatal(err)
	}
	websockifyPidFile.Sync()
}

func (govm *GoVM) Launch() {
	ctx := context.Background()

	// Create the data dir
	vmDataDirectory := govm.Workdir + "/data/" + govm.Name + "/" + uuid.New().String()
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
			"mykey": SSHKey,
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
	if efi {
		qemuParams = append(qemuParams, "-bios /OVMF.fd ")
	}
	if cloud {
		env = append(env, "CLOUD=yes")
		env = append(env, "CLOUD_INIT_OPTS=-drive file=/data/seed.iso,if=virtio,format=raw ")
	}

	// Default Mount binds
	defaultMountBinds := []string{
		fmt.Sprintf("%v:/image/image", govm.ParentImage),
		fmt.Sprintf("%v:/data", vmDataDirectory),
		fmt.Sprintf("%v:/cloud-init/openstack/latest/meta_data.json", vmDataDirectory+"/meta_data.json"),
	}

	if userData != "" {
		defaultMountBinds = append(defaultMountBinds, fmt.Sprintf("%s:/cloud-init/openstack/latest/user_data", userData))
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

	if VNC {
		/* Get an available port */
		Port = strconv.Itoa(findPort())
	}

	// Create the Container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "verbacious/govm",
		Cmd:   qemuParams,
		//Cmd:          []string{"top"},
		Env: env,
		Labels: map[string]string{
			"websockifyPort": Port,
			"dataDir":        vmDataDirectory,
		},
		//ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: true,
		Binds:           defaultMountBinds,
		//PortBindings:    ports,
	}, nil, name)
	if err != nil {
		panic(err)
	}

	govm.containerID = resp.ID

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	if verbose {
		fmt.Println(qemuParams)
	}

	if VNC {
		govm.setVNC(vmDataDirectory, Port)

	} else {
		fmt.Println("no vnc")
	}
}
