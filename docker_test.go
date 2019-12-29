package docker

import (
	"context"
	"fmt"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"github.com/miekg/dns"
)

type MockCli struct {
	client.ContainerAPIClient
}

func (cli MockCli) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	settings := network.EndpointSettings{IPAddress: "172.0.0.3", GlobalIPv6Address: "2001:db8::3"}
	networks := make(map[string]*network.EndpointSettings)
	networks["some-network"] = &settings
	networkSettings := types.SummaryNetworkSettings{Networks: networks}
	container := types.Container{Names: []string{"/some-container"}, NetworkSettings: &networkSettings}

	return []types.Container{container}, nil
}

func TestExample(t *testing.T) {
	tests := []struct {
		domain          string
		questionHost    string
		questionType    uint16
		questionTypeStr string
		expectedIP      string
		expectedError   bool
	}{
		{".", "some-container.", dns.TypeA, "A", "172.0.0.3", false},
		{"domain.", "some-container.domain.", dns.TypeA, "A", "172.0.0.3", false},
		{"domain.", "some-container.domain.", dns.TypeAAAA, "AAAA", "2001:db8::3", false},
	}

	for _, example := range tests {
		cli := MockCli{}
		dockerPlugin := Docker{Domains: []string{example.domain}, Cli: cli}
		ctx := context.TODO()

		query := new(dns.Msg)
		query.SetQuestion(example.questionHost, example.questionType)
		recorder := dnstest.NewRecorder(&test.ResponseWriter{})
		dockerPlugin.ServeDNS(ctx, recorder, query)

		expected := fmt.Sprintf("%s	50	IN	%s	%s", example.questionHost, example.questionTypeStr, example.expectedIP)
		if record := recorder.Msg.Answer[0].String(); record != expected {
			t.Errorf("Failed, got %s", record)
		}
	}
}
