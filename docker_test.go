package docker

import (
	"context"
	"fmt"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"

	"github.com/miekg/dns"
)

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

func TestContainerListFailure(t *testing.T) {
	cli := ConnFailureCli{}
	dockerPlugin := docker{domains: []string{}, cli: cli}
	ctx := context.TODO()

	query := new(dns.Msg)
	query.SetQuestion("some-host.", dns.TypeA)
	recorder := dnstest.NewRecorder(&test.ResponseWriter{})
	_, err := dockerPlugin.ServeDNS(ctx, recorder, query)
	if err == nil {
		t.Errorf("Failed, expected err ")
	}
}

func TestContainers(t *testing.T) {
	tests := []struct {
		domains         []string
		questionHost    string
		questionType    uint16
		questionTypeStr string
		expectedIP      string
		expectedError   bool
	}{
		{[]string{"."}, "some-container-4.", dns.TypeA, "A", "172.0.0.3", false},
		{[]string{"."}, "some-container-4.", dns.TypeAAAA, "AAAA", "", true},
		{[]string{"domain."}, "some-container-4.otherdomain.", dns.TypeA, "A", "172.0.0.3", true},
		{[]string{"domain."}, "some-container-4.domain.", dns.TypeA, "A", "172.0.0.3", false},
		{[]string{"domain."}, "some-container-6.domain.", dns.TypeAAAA, "AAAA", "2001:db8::3", false},
		{[]string{"domain1.", "domain2."}, "some-container-6.domain2.", dns.TypeAAAA, "AAAA", "2001:db8::3", false},
	}

	for _, example := range tests {
		cli := WorkingCli{}
		dockerPlugin := docker{domains: example.domains, cli: cli}
		ctx := context.TODO()

		query := new(dns.Msg)
		query.SetQuestion(example.questionHost, example.questionType)
		recorder := dnstest.NewRecorder(&test.ResponseWriter{})
		dockerPlugin.ServeDNS(ctx, recorder, query)

		expected := fmt.Sprintf("%s	0	IN	%s	%s", example.questionHost, example.questionTypeStr, example.expectedIP)
		if example.expectedError {
			if recorder.Msg != nil {
				t.Errorf("Failed, got %v", recorder.Msg)
			}
		} else if record := recorder.Msg.Answer[0].String(); record != expected {
			t.Errorf("Failed [%s], got %s", expected, record)
		}
	}
}
