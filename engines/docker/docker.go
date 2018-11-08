package docker

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// Docker stands as a docker service client.
type Docker struct {
	ctx    context.Context
	client *client.Client
}

// NewDockerClient returns a new Docker service client.
func NewDockerClient() *Docker {
	cli, err := client.NewEnvClient()
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
