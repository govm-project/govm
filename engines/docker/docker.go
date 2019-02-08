package docker

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Docker stands as a docker service client.
type Docker struct {
	ctx context.Context
	*client.Client
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
	out, err := d.ImagePull(d.ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	_, err = io.Copy(ioutil.Discard, out)
	return err
}

// Exec executes commands inside a running container
func (d *Docker) Exec(containerName string, execConfig types.ExecConfig) error {

	resp, err := d.ContainerExecCreate(d.ctx, containerName, execConfig)
	if err != nil {
		return err
	}

	err = d.ContainerExecStart(d.ctx, resp.ID,
		types.ExecStartCheck{Detach: true, Tty: false})
	return err
}

// Create creates a new docker container
func (d *Docker) Create(containerConfig *container.Config, hostConfig *container.HostConfig,
	networkConfig *network.NetworkingConfig, name string) (string, error) {

	if !d.ImageExists(containerConfig.Image) {
		log.Printf("Pulling %v image", containerConfig.Image)
		err := d.PullImage(containerConfig.Image)
		if err != nil {
			return "", err
		}
	}
	resp, err := d.ContainerCreate(d.ctx, containerConfig, hostConfig,
		networkConfig, name)

	return resp.ID, err
}

// Start starts a previously created container.
func (d *Docker) Start(id, name string) error {

	if id == "" {
		container, err := d.Search(name)
		if err != nil {
			return err
		}
		id = container.ID
	}

	return d.ContainerStart(d.ctx, id, types.ContainerStartOptions{})
}

// Search searches a container from the running docker containers
func (d *Docker) Search(name string) (types.Container, error) {
	containers, err := d.ContainerList(d.ctx, types.ContainerListOptions{})
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

	images, err := d.ImageList(d.ctx,
		types.ImageListOptions{
			All:     false,
			Filters: filters.NewArgs(filters.Arg("reference", "govm/govm")),
		},
	)
	if err != nil || len(images) == 0 {
		return false
	}
	return true
}

// Inspect inspects and return details about an specific container.
func (d *Docker) Inspect(ID string) (types.ContainerJSON, error) {
	return d.ContainerInspect(d.ctx, ID)
}

// List lists all docker-based VM instances that mee the passed filters
func (d *Docker) List(args filters.Args) ([]types.Container, error) {
	return d.ContainerList(context.Background(),
		types.ContainerListOptions{
			Quiet:   false,
			Size:    false,
			All:     true,
			Latest:  false,
			Since:   "",
			Before:  "",
			Limit:   0,
			Filters: args,
		})
}

// Remove wraps the ContainerRemove functionality.
func (d *Docker) Remove(id string) error {
	return d.ContainerRemove(context.Background(), id,
		types.ContainerRemoveOptions{
			RemoveVolumes: false,
			RemoveLinks:   false,
			Force:         true,
		})
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
