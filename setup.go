package docker

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/caddyserver/caddy"
)

var logger = clog.NewWithPlugin("docker")

func init() { plugin.Register("docker", setup) }

type DockerCli interface {
	client.ContainerAPIClient
	client.SystemAPIClient
}

var newDockerCli = func() (DockerCli, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func setup(c *caddy.Controller) error {
	c.Next()

	domains := make([]string, len(c.ServerBlockKeys))
	for i, key := range c.ServerBlockKeys {
		domains[i] = plugin.Host(key).Normalize()
	}

	cli, err := newDockerCli()
	if err != nil {
		return plugin.Error("docker", err)
	}

	ctx := context.Background()
	if _, err := cli.Ping(ctx); err != nil {
		logger.Errorf("Disabling plugin due to errors. %v", err)
	} else {
		dockerDNS, _ := NewDockerDNS(domains, cli)
		dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
			return &DockerPlugin{next: next, dockerDNS: dockerDNS}
		})
	}

	return nil
}
