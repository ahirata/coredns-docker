package docker

import (
	"errors"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/docker/docker/client"

	"github.com/caddyserver/caddy"
)

// init registers this plugin.
func init() { plugin.Register("docker", setup) }

// setup is the function that gets called when the config parser see the token "docker". Setup is responsible
// for parsing any extra options the docker plugin may have. The first token this function sees is "docker".
func setup(c *caddy.Controller) error {
	c.Next() // Ignore "docker" and give us the next token.

	domains := make([]string, len(c.ServerBlockKeys))
	for i, key := range c.ServerBlockKeys {
		domains[i] = plugin.Host(key).Normalize()
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return plugin.Error("docker", errors.New("could not get a docker client"))
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &Docker{Domains: domains, Cli: cli, Next: next}
	})

	// All OK, return a nil error.
	return nil
}
