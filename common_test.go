package docker

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type WorkingCli struct {
	client.ContainerAPIClient
	client.SystemAPIClient
}

type InitFailureCli struct {
	client.ContainerAPIClient
	client.SystemAPIClient
}

type ConnFailureCli struct {
	client.ContainerAPIClient
	client.SystemAPIClient
}

func (cli WorkingCli) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return []types.Container{ipv4Container(), ipv6Container()}, nil
}

func (cli WorkingCli) Ping(ctx context.Context) (types.Ping, error) {
	return types.Ping{}, nil
}

func (cli ConnFailureCli) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return nil, errors.New("connection failure")
}

func (cli ConnFailureCli) Ping(ctx context.Context) (types.Ping, error) {
	var ping types.Ping
	return ping, errors.New("Failed")
}
