package docker

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/docker/docker/api/types/events"
	"github.com/miekg/dns"
)

func TestPluginName(t *testing.T) {
	dockerPlugin := DockerPlugin{next: nil, dockerDNS: nil}

	if pluginName := dockerPlugin.Name(); pluginName != "docker" {
		t.Errorf("Failed, got %v", pluginName)
	}
}

func TestHandler(t *testing.T) {
	tests := []struct {
		questionHost  string
		questionType  uint16
		expectedError bool
	}{
		{"some-container-4.", dns.TypeA, true},
		{"some-container-4.otherdomain.", dns.TypeA, true},
		{"some-container-4.domain.", dns.TypeA, false},
		{"some-container-6.", dns.TypeAAAA, true},
		{"some-container-6.otherdomain.", dns.TypeAAAA, true},
		{"some-container-6.domain.", dns.TypeAAAA, false},
	}

	cli := WorkingCli{messages: make(chan events.Message), errs: make(chan error, 1)}
	dockerDNS, _ := NewDockerDNS([]string{"domain."}, cli)

	dockerPlugin := DockerPlugin{next: nil, dockerDNS: dockerDNS}

	for _, example := range tests {
		ctx := context.TODO()

		query := new(dns.Msg)
		query.SetQuestion(example.questionHost, example.questionType)
		recorder := dnstest.NewRecorder(&test.ResponseWriter{})
		dockerPlugin.ServeDNS(ctx, recorder, query)

		if recorder.Msg == nil && !example.expectedError {
			t.Errorf("Failed, got %v", recorder.Msg)
		}
	}
}
