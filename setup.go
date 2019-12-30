package docker

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/caddyserver/caddy"
)

func init() { plugin.Register("docker", setup) }

type DockerCli interface {
	client.ContainerAPIClient
	client.SystemAPIClient
}

// exposed for testing
var f = func() (DockerCli, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func setup(c *caddy.Controller) error {
	var log = clog.NewWithPlugin("docker")

	c.Next()

	domains := make([]string, len(c.ServerBlockKeys))
	for i, key := range c.ServerBlockKeys {
		domains[i] = plugin.Host(key).Normalize()
	}
	cli, err := f()
	if err != nil {
		return plugin.Error("docker", err)
	}

	ctx := context.Background()
	if _, err := cli.Ping(ctx); err != nil {
		log.Error("Disabling plugin due to errors. ", err)
	} else {
		dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
			return &docker{domains: domains, cli: cli, next: next}
		})
	}

	return nil
}
