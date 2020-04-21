package docker

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/network"

	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type WorkingCli struct {
	client.ContainerAPIClient
	client.SystemAPIClient
	messages <-chan events.Message
	errs     <-chan error
}

type EventCli struct {
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

func ipv4Container() types.Container {
	settings := network.EndpointSettings{IPAddress: "172.0.0.3"}
	networks := make(map[string]*network.EndpointSettings)
	networks["some-network"] = &settings
	networkSettings := types.SummaryNetworkSettings{Networks: networks}
	return types.Container{Names: []string{"/some-container-4"}, NetworkSettings: &networkSettings}
}

func ipv6Container() types.Container {
	settings := network.EndpointSettings{GlobalIPv6Address: "2001:db8::3"}
	networks := make(map[string]*network.EndpointSettings)
	networks["some-network"] = &settings
	networkSettings := types.SummaryNetworkSettings{Networks: networks}
	return types.Container{Names: []string{"/some-container-6"}, NetworkSettings: &networkSettings}
}

func ipv4containerJSON() types.ContainerJSON {
	settings := network.EndpointSettings{IPAddress: "172.0.0.3"}
	networks := make(map[string]*network.EndpointSettings)
	networks["some-network"] = &settings
	networkSettings := types.NetworkSettings{Networks: networks}
	return types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{Name: "/some-container-4"}, NetworkSettings: &networkSettings}
}

func ipv6containerJSON() types.ContainerJSON {
	settings := network.EndpointSettings{GlobalIPv6Address: "2001:db8::3"}
	networks := make(map[string]*network.EndpointSettings)
	networks["some-network"] = &settings
	networkSettings := types.NetworkSettings{Networks: networks}
	return types.ContainerJSON{ContainerJSONBase: &types.ContainerJSONBase{Name: "/some-container-6"}, NetworkSettings: &networkSettings}
}

func (cli WorkingCli) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return []types.Container{ipv4Container(), ipv6Container()}, nil
}

func (cli WorkingCli) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if containerID == "some-container-4" {
		return ipv4containerJSON(), nil
	}
	if containerID == "some-container-6" {
		return ipv6containerJSON(), nil
	}
	return types.ContainerJSON{}, errors.New("connection failure")
}

func (cli WorkingCli) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	return cli.messages, cli.errs
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
