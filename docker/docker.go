package docker

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// PullImage pulls image from docker registry
func PullImage(ctx context.Context, cli *client.Client, imageName string) error {
	_, err := cli.ImagePull(ctx, "alpine", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	return err
}

// Exec executes commands inside a running container
func Exec(ctx context.Context, cli *client.Client, containerName string,
	execConfig types.ExecConfig) error {

	resp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)

	if err != nil {
		return err
	}
	err = cli.ContainerExecStart(ctx, resp.ID, types.ExecStartCheck{Detach: true, Tty: false})
	return err
}

// Run starts a new docker container
func Run(ctx context.Context, cli *client.Client, containerConfig *container.Config,
	hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, name string) (string, error) {

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, name)
	if err != nil {
		return "", err
	}

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	return resp.ID, err
}

// ContainerSearch searches a container from the running docker containers
func ContainerSearch(ctx context.Context, cli *client.Client, name string) error {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	err = errors.New("NotFound")

	for _, container := range containers {
		if container.Names[0] == fmt.Sprintf("/%v", name) {
			return nil
		}
	}
	return err
}
