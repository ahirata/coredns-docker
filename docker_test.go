package docker

import (
	"errors"
	"io"
	"testing"

	"github.com/docker/docker/api/types/events"
	"github.com/miekg/dns"
)

func TestContainerListFailure(t *testing.T) {
	cli := ConnFailureCli{}
	_, err := NewDockerDNS([]string{}, cli)
	if err == nil {
		t.Errorf("Failed, expected err ")
	}
}

func TestContainers(t *testing.T) {
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

	for _, example := range tests {
		if records := dockerDNS.GetRecords()[example.questionHost]; len(records) == 0 && !example.expectedError {
			t.Errorf("Failed, container [%v] not found in %v", example, dockerDNS.GetRecords())
		} else {
			found := false
			for _, record := range records {
				if record.Header().Rrtype == example.questionType {
					found = true
					break
				}
			}
			if !found && !example.expectedError {
				t.Errorf("Failed, container [%v] not found in %v", example, records)
			}
		}
	}
}

func TestContainerUpdates(t *testing.T) {
	messages := make(chan events.Message)
	errs := make(chan error, 1)

	cli := WorkingCli{messages: messages, errs: errs}
	NewDockerDNS([]string{"domain."}, cli)

	messages <- events.Message{Type: "network", Action: "connect", Actor: events.Actor{Attributes: map[string]string{"container": "some-container-6"}}}
	messages <- events.Message{Type: "network", Action: "disconnect", Actor: events.Actor{Attributes: map[string]string{"name": "some-container-6"}}}
	messages <- events.Message{Type: "network", Action: "disconnect", Actor: events.Actor{Attributes: map[string]string{"name": "some-container-8"}}}
	messages <- events.Message{Type: "container", Action: "rename", Actor: events.Actor{Attributes: map[string]string{"name": "some-container-6", "oldName": "/some-old-container"}}}
	errs <- errors.New("Unexpected error")
	errs <- io.EOF
}
