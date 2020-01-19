package docker

import (
	"errors"
	"github.com/coredns/coredns/core/dnsserver"
	"testing"

	"github.com/caddyserver/caddy"
)

func TestDockerInit(t *testing.T) {
	if _, err := newDockerCli(); err != nil {
		t.Errorf("Unxpected error %v", err)
	}
}

func TestDockerInitFailureSetup(t *testing.T) {
	newDockerCli = func() (DockerCli, error) {
		return nil, errors.New("docker client init failed")
	}

	c := caddy.NewTestController("dns", "")
	if err := setup(c); err == nil || len(dnsserver.GetConfig(c).Plugin) != 0 {
		t.Errorf("Expected error")
	}
}

func TestDockerConnectFailureSetup(t *testing.T) {
	newDockerCli = func() (DockerCli, error) {
		return ConnFailureCli{}, nil
	}

	c := caddy.NewTestController("dns", "")
	if err := setup(c); err != nil || len(dnsserver.GetConfig(c).Plugin) != 0 {
		t.Errorf("Expected error")
	}
}

func TestSetup(t *testing.T) {
	newDockerCli = func() (DockerCli, error) {
		return WorkingCli{}, nil
	}

	c := caddy.NewTestController("dns", "docker")
	c.ServerBlockKeys = []string{"domain.com.:8053", "dynamic.domain.com.:8053"}
	if err := setup(c); err != nil {
		t.Errorf("Unexpected errors: %v", err)
	}

	plugin := dnsserver.GetConfig(c).Plugin
	if len(plugin) != 1 || plugin[0](nil) == nil {
		t.Errorf("Unexpected errors")
	}
}
