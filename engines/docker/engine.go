package docker

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/govm-project/govm/internal"
	"github.com/govm-project/govm/vm"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
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

	return e.docker.Stop(container.ID,"")
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
