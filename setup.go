package docker

import (
	"errors"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/docker/docker/client"

	"github.com/caddyserver/caddy"
)

func init() { plugin.Register("docker", setup) }

func setup(c *caddy.Controller) error {
	c.Next()

	domains := make([]string, len(c.ServerBlockKeys))
	for i, key := range c.ServerBlockKeys {
		domains[i] = plugin.Host(key).Normalize()
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return plugin.Error("docker", errors.New("could not get a docker client"))
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &docker{domains: domains, cli: cli, next: next}
	})

	return nil
}
