package docker

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/govm-project/govm/internal"
	"github.com/govm-project/govm/vm"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	log "github.com/sirupsen/logrus"
)

// Engine stands an entry point for the docker container services
// nolint: typecheck
type Engine struct {
	docker *Docker
}

// Init initializes a the Engine's docker client
// nolint: typecheck
func (e *Engine) Init() {
	e.docker = NewDockerClient()
}

// Create create a new instance
func (e Engine) Create(spec vm.Instance) (id string, err error) { // nolint: funlen
	vmDataDirectory := spec.Workdir + "/data/" + spec.Name
	// Default Environment Variables
	env := []string{
		"AUTO_ATTACH=yes",
		"DEBUG=yes",
		fmt.Sprintf(vm.KVMCPUOpts,
			spec.Size.CPUModel,
			spec.Size.Sockets,
			spec.Size.Cpus,
			spec.Size.Cores,
			spec.Size.Threads,
			(spec.Size.Sockets * spec.Size.Cores * spec.Size.Threads),
			spec.Size.RAM,
		),
		fmt.Sprintf("COW_SIZE=%d", spec.Size.DISK),
	}
	env = append(env, spec.ContainerEnvVars...)

	qemuParams := []string{
		"-vnc unix:/data/vnc",
	}
	if spec.Efi {
		qemuParams = append(qemuParams, "-bios /OVMF.fd ")
	}

	if spec.Cloud {
		env = append(env, "CLOUD=yes")
		env = append(env, vm.CloudInitOpts)
	}

	// Default Mount binds
	defaultMountBinds := []string{
		fmt.Sprintf(vm.ImageMount, spec.ParentImage),
		fmt.Sprintf(vm.DataMount, vmDataDirectory),
		fmt.Sprintf(vm.MetadataMount, vmDataDirectory, vm.MedatataFile),
	}
	// Append shares to defaultMountBinds if any.
	// Append guest directory/ies to env (container's environment).
	if len(spec.Shares) > 0 {
		defaultMountBinds = append(defaultMountBinds, spec.Shares...)

		var sharedDirs string
		for _, share := range spec.Shares {
			// Get the guest directory part from the string and pass
			// it to the container environment. The "startvm" script
			// will handle these shared directories.
			sharedDirs += strings.Split(share, ":")[1] + " "
		}

		env = append(env, "SHARED_DIRS="+sharedDirs)
	}

	if spec.UserData != "" {
		defaultMountBinds = append(defaultMountBinds,
			fmt.Sprintf(vm.UserDataMount, spec.UserData))
	}

	// Get an available port for VNC
	vncPort := strconv.Itoa(internal.FindAvailablePort())

	// Create the Container
	containerConfig := &container.Config{
		Image:    VMLauncherContainerImage,
		Hostname: spec.Name,
		Cmd:      qemuParams,
		Env:      env,
		Labels: map[string]string{
			"websockifyPort": vncPort,
			"dataDir":        vmDataDirectory,
			"namespace":      spec.Namespace,
			"govmType":       "instance",
			"vmName":         spec.Name,
		},
	}

	hostConfig := &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: true,
		Binds:           defaultMountBinds,
		DNS:             spec.NetOpts.DNS,
		RestartPolicy:   container.RestartPolicy{Name: "always"},
	}

	if spec.NetOpts.IP != "" {
		containerConfig.Labels["ip"] = spec.NetOpts.IP
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			spec.NetOpts.NetID: {
				IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: spec.NetOpts.IP},
				MacAddress: spec.NetOpts.MAC,
				NetworkID:  spec.NetOpts.NetID,
				IPAddress:  spec.NetOpts.IP,
			},
		},
	}

	containerName := internal.GenerateContainerName(spec.Namespace, spec.Name)
	if _, err = e.docker.Search(containerName); err != nil && err.Error() != "NotFound" {
		return id, err
	}

	id, err = e.docker.Create(containerConfig, hostConfig, networkConfig, containerName)

	return id, err
}

// Start starts a Docker container-based VM instance
func (e Engine) Start(namespace, id string) error {
	container, err := e.docker.Inspect(id)
	if err != nil {
		fullName := internal.GenerateContainerName(namespace, id)
		container, err = e.docker.Inspect(fullName)

		if err != nil {
			return err
		}
	}

	return e.docker.Start(container.ID, "")
}

// Stop stops a Docker container-based VM instance
func (e Engine) Stop(namespace, id string) error {
	container, err := e.docker.Inspect(id)
	if err != nil {
		fullName := internal.GenerateContainerName(namespace, id)
		container, err = e.docker.Inspect(fullName)

		if err != nil {
			return err
		}
	}

	return e.docker.Stop(container.ID, "")
}

// Save a Docker container-based VM instance
func (e Engine) Save(namespace, id, outputFile string, stopVM bool) error {
	fullName := internal.GenerateContainerName(namespace, id)
	containerObj, err := e.docker.Inspect(fullName)
	if err != nil {
		return err
	}

	// Save base image fist into data dir
	execCmd := []string{}
	execConfig := types.ExecConfig{
		AttachStdout: true,
		Detach:       false,
		Cmd:          execCmd,
	}
	execConfig.Cmd = strings.Split("cp /image/image /data/base_image", " ")
	err = e.docker.Exec(fullName, execConfig)
	if err != nil {
		log.Printf("Couldn't save base image from [%v]", fullName)
	}

	// Stop VM
	if stopVM {
		err = e.Stop(namespace, id)
		if err != nil {
			log.Printf("Couldn't stop the container [%v]", containerObj.ID)
			return err
		}
	}

	// Save VM
	/// Create a Backup Container
	baseContainerName := strings.Replace(containerObj.Name, "/", "", -1)
	backupContainerName := fmt.Sprintf("%v-backup", baseContainerName)
	_, err = e.docker.Inspect(backupContainerName)
	if err != nil {
		containerConfig := &container.Config{
			Image:    "govm/qemu",
			Hostname: "backuptest",
			Cmd:      []string{"sh", "-c", "while true ; do sleep 1; done"},
			Labels:   map[string]string{},
		}

		dataDir := containerObj.Config.Labels["dataDir"]
		currentDir, _ := os.Getwd()
		mountBinds := []string{
			fmt.Sprintf(vm.DataMount, dataDir),
			fmt.Sprintf(vm.DataMount, currentDir) + "-out",
		}
		hostConfig := &container.HostConfig{
			Privileged:      true,
			PublishAllPorts: true,
			Binds:           mountBinds,
		}
		networkConfig := &network.NetworkingConfig{}
		backupContainerID, err := e.docker.Create(containerConfig, hostConfig, networkConfig, backupContainerName)
		if err != nil {
			log.Printf("Backup Container Error: %v", err)
		}
		err = e.Start(namespace, backupContainerID)
		if err != nil {
			log.Printf("Backup Container Starting failed: %v", err)
		}
	}
	// Exec qemu backup commands
	cmds := []string{
		"rm /tmp/*",
		"cp /data/base_image /data-out/",
		"cp /data/cow_image.qcow2 /data-out/head.qcow2",
		"qemu-img rebase -f qcow2 -F qcow2 -p -u -b /data-out/base_image /data-out/head.qcow2",
		"qemu-img commit -p /data-out/head.qcow2",
		fmt.Sprintf("mv /data-out/base_image /data-out/%v", outputFile),
		"rm /data-out/head.qcow2",
	}

	for _, cmd := range cmds {
		log.Println(cmd)
		execConfig.Cmd = strings.Split(cmd, " ")
		err = e.docker.Exec(backupContainerName, execConfig)
		if err != nil {
			log.Printf("Couldn't execute backup commands on [%v]", backupContainerName)
		}
	}

	// Start VM
	if stopVM {
		err = e.Start(namespace, id)
		if err != nil {
			log.Printf("Couldn't start the container [%v]", containerObj.Name)
			return err
		}
	}

	// Remove Backup Container
	govmID := strings.SplitN(backupContainerName, ".", 3)[2]
	return e.Delete(namespace, govmID)
}

// List lists  all the Docker container-based VM instances
// nolint: typecheck
func (e Engine) List(namespace string, all bool) ([]vm.Instance, error) {
	listArgs := filters.NewArgs()
	instances := []vm.Instance{}

	listArgs.Add("label", "namespace="+namespace)

	if all {
		listArgs.Add("label", "govmType=instance")
	}

	containers, err := e.docker.List(listArgs)
	if err != nil {
		return instances, err
	}

	for _, container := range containers {
		containerIP := ""
		for _, net := range container.NetworkSettings.Networks {
			containerIP = net.IPAddress
			break
		}

		vncPort, _ := strconv.ParseInt(container.Labels["websockifyPort"], 10, 32)

		instances = append(instances, vm.Instance{
			ID:        container.ID[:10],
			Name:      container.Labels["vmName"],
			Namespace: namespace,
			VNCPort:   vncPort,
			NetOpts:   vm.NetworkingOptions{IP: containerIP},
		},
		)
	}

	return instances, err
}

// Delete deletes an Instance of GoVM
func (e Engine) Delete(namespace, id string) error {
	container, err := e.docker.Inspect(id)
	if err != nil {
		fullName := internal.GenerateContainerName(namespace, id)
		container, err = e.docker.Inspect(fullName)

		if err != nil {
			return err
		}
	}

	dataPath := container.Config.Labels["dataDir"]
	defer os.RemoveAll(dataPath)

	pid, err := ioutil.ReadFile(dataPath + "/websockifyPid")
	if err == nil {
		websockifyPid, _ := strconv.Atoi(string(pid))

		websockifyProcess, err := os.FindProcess(websockifyPid)
		if err != nil {
			return err
		}

		err = websockifyProcess.Kill()
		if err != nil {
			return err
		}
	}

	return e.docker.Remove(container.ID)
}

// nolint:godox
// TODO: Figure how to set VNC on new GoVM instaces
// This shoud run after a GoVM instance start.
//func (e Engine) setVNC(vncID string, port string) error {
//
//	vncContainer, err := e.docker.Search(VNCServerContainerName)
//	if err != nil {
//		log.WithFields(log.Fields{
//			"error":     err.Error(),
//			"container": VNCContainerImage,
//		}).Warn("VNC container was not found")
//
//		mountBinds := []string{
//			fmt.Sprintf("%v/data:/vm", internal.GetDefaultWorkDir())}
//
//		containerConfig := &container.Config{
//			Image:    VNCContainerImage,
//			Hostname: VNCServerContainerName,
//			Cmd:      nil,
//			Env:      nil,
//			Labels:   nil,
//		}
//
//		hostConfig := &container.HostConfig{
//			Privileged:      true,
//			PublishAllPorts: true,
//			NetworkMode:     "host",
//			Binds:           mountBinds,
//		}
//		_, err := e.docker.Create(containerConfig, hostConfig,
//			&network.NetworkingConfig{}, VNCServerContainerName)
//		if err != nil {
//			return err
//		}
//	}
//
//	execCmd := []string{"/noVNC/utils/websockify/run",
//		"--web", "/noVNC",
//		"--unix-target", "/vm/" + vncID + "/vnc",
//		port}
//
//	execConfig := types.ExecConfig{
//		Detach: true,
//		Cmd:    execCmd,
//	}
//
//	err = e.docker.Exec(vncContainer.Names[0], execConfig)
//
//	return err
//}
