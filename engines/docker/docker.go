package docker

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Docker stands as a docker service client.
type Docker struct {
	ctx    context.Context
	client *client.Client
}

// NewDockerClient returns a new Docker service client.
func NewDockerClient() *Docker {

	SetAPIVersion()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	return &Docker{context.Background(), cli}
}

// PullImage pulls image from docker registry
func (d *Docker) PullImage(image string) error {
	_, err := d.client.ImagePull(d.ctx, image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	return err
}

// Exec executes commands inside a running container
func (d *Docker) Exec(containerName string, execConfig types.ExecConfig) error {

	resp, err := d.client.ContainerExecCreate(d.ctx, containerName, execConfig)
	if err != nil {
		return err
	}

	err = d.client.ContainerExecStart(d.ctx, resp.ID,
		types.ExecStartCheck{Detach: true, Tty: false})
	return err
}

// Create creates a new docker container
func (d *Docker) Create(containerConfig *container.Config, hostConfig *container.HostConfig,
	networkConfig *network.NetworkingConfig, name string) (string, error) {

	resp, err := d.client.ContainerCreate(d.ctx, containerConfig, hostConfig,
		networkConfig, name)
	if err != nil {
		return "", err
	}

	err = d.client.ContainerStart(d.ctx, resp.ID, types.ContainerStartOptions{})
	return resp.ID, err
}

// Search searches a container from the running docker containers
func (d *Docker) Search(name string) (types.Container, error) {
	containers, err := d.client.ContainerList(d.ctx, types.ContainerListOptions{})
	if err != nil {
		return types.Container{}, err
	}

	err = errors.New("NotFound")

	for _, container := range containers {
		if container.Names[0] == fmt.Sprintf("/%v", name) {
			return container, nil
		}
	}
	return types.Container{}, err
}

// ImageExists verifies that an image exists in the local docker registry
func (d *Docker) ImageExists(name string) bool {
	images, err := d.client.ImageSearch(d.ctx, name, types.ImageSearchOptions{Limit: 1})
	if err != nil || len(images) == 0 {
		fmt.Printf("ERROR: %v\n", err)
		return false
	}
	return true
}

// Start starts a previously created container.
func (d *Docker) Start(name string) error {

	container, err := d.Search(name)
	if err != nil {
		return err
	}

	return d.client.ContainerStart(d.ctx, container.ID, types.ContainerStartOptions{})

}

// Inspect inspects and return details about an specific container.
func (d *Docker) Inspect(ID string) (types.ContainerJSON, error) {
	return d.client.ContainerInspect(d.ctx, ID)
}

// SetAPIVersion gets local docker server API version.
// TODO: Investigate how we can replace the exec.Command approach
func SetAPIVersion() {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.APIVersion}}")
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error getting Docker Server API version: %v", err)
	}
	apiVersion := strings.TrimSpace(string(cmdOutput.Bytes()))
	_ = os.Setenv("DOCKER_API_VERSION", apiVersion)
}
