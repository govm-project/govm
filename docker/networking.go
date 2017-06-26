package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
)

type Network struct {
	Name   string `yaml:"name"`
	Subnet string `yaml:"subnet"`
}

func VerifyNetwork(ctx context.Context, cli *client.Client, networkName string) error {
	filters := filters.NewArgs()
	filters.Add("name", networkName)
	_, err := cli.NetworkList(ctx, types.NetworkListOptions{
		filters,
	})
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}
