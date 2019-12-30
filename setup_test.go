package docker

import (
	"errors"
	"github.com/coredns/coredns/core/dnsserver"
	"testing"

	"github.com/caddyserver/caddy"
)

func TestDockerInit(t *testing.T) {
	if _, err := f(); err != nil {
		t.Errorf("Unxpected error %v", err)
	}
}

func TestDockerInitFailureSetup(t *testing.T) {
	f = func() (DockerCli, error) {
		return nil, errors.New("docker client init failed")
	}

	c := caddy.NewTestController("dns", "")
	if _ = setup(c); len(dnsserver.GetConfig(c).Plugin) != 0 {
		t.Errorf("Expected error")
	}
}

func TestDockerConnectFailureSetup(t *testing.T) {
	f = func() (DockerCli, error) {
		return ConnFailureCli{}, nil
	}

	c := caddy.NewTestController("dns", "")
	if _ = setup(c); len(dnsserver.GetConfig(c).Plugin) != 0 {
		t.Errorf("Expected error")
	}
}

func TestSetup(t *testing.T) {
	f = func() (DockerCli, error) {
		return WorkingCli{}, nil
	}

	tests := []struct {
		body          string
		expectedError bool
	}{
		{"docker", false},
	}

	for _, test := range tests {
		c := caddy.NewTestController("dns", test.body)
		c.ServerBlockKeys = []string{"domain.com.:8053", "dynamic.domain.com.:8053"}
		if err := setup(c); len(dnsserver.GetConfig(c).Plugin) < 1 {
			t.Errorf("Unexpected errors: %v", err)
		} else if h := dnsserver.GetConfig(c).Plugin[0](nil); h == nil {
			t.Errorf("Unexpected errors")
		}
	}
}
