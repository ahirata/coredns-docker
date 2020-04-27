package docker

import (
	"context"
	"io"
	"net"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin/pkg/dnsutil"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"

	"github.com/miekg/dns"
)

type DockerDNS struct {
	cli     DockerCli
	domains []string
	records map[string][]dns.RR
}

func NewDockerDNS(domains []string, cli DockerCli) (*DockerDNS, error) {
	dockerDNS := DockerDNS{domains: domains, cli: cli}
	if err := dockerDNS.initializeRecords(); err != nil {
		return nil, err
	}

	go dockerDNS.eventListener()

	return &dockerDNS, nil
}

func (d *DockerDNS) GetRecords() map[string][]dns.RR {
	return d.records
}

func (d *DockerDNS) eventListener() {
	eventArgs := filters.NewArgs()
	eventArgs.Add("type", "network")
	eventArgs.Add("event", "connect")
	eventArgs.Add("event", "disconnect")

	eventArgs.Add("type", "container")
	eventArgs.Add("event", "rename")

	eventOptions := types.EventsOptions{Filters: eventArgs}
	messages, errs := d.cli.Events(context.Background(), eventOptions)

	for {
		select {
		case err := <-errs:
			if err != nil {
				if err == io.EOF {
					logger.Info("Docker channel EOF.", err)
				} else {
					logger.Error("Error while reading docker events channel", err)
				}
			} else {
				logger.Debug("Seems we don't have a connection with docker events.")
				messages, errs = d.retryConnection(eventOptions)
			}
		case msg := <-messages:
			logger.Debug(msg)
			d.handleMessage(msg)
		}
	}
}

func (d *DockerDNS) retryConnection(eventOptions types.EventsOptions) (<-chan events.Message, <-chan error) {
	if len(d.records) > 0 {
		d.records = make(map[string][]dns.RR)
	}
	for {
		time.Sleep(1 * time.Second)
		logger.Info("Trying to get new event channel... ")
		messages, errs := d.cli.Events(context.Background(), eventOptions)
		select {
		case m := <-errs:
			logger.Errorf("%v\n", m)
		case <-time.After(1 * time.Second):
			logger.Debug("Nothing came on the error channel. Initializing records since everything seems fine")
			d.initializeRecords()
			return messages, errs
		}
	}
}

func (d *DockerDNS) handleMessage(msg events.Message) {
	container, err := d.containerInspect(msg)
	if err != nil {
		logger.Errorf("Error while inspecting the container %v, %v", msg, err)
		return
	}

	if msg.Action == "connect" {
		d.addContainer(container.Name, container.NetworkSettings.Networks)
	} else if msg.Action == "disconnect" {
		d.removeContainer(container.Name)
	} else if msg.Action == "rename" {
		d.renameContainer(msg.Actor.Attributes["oldName"], msg.Actor.Attributes["name"], container.NetworkSettings.Networks)
	}
}

func (d *DockerDNS) containerInspect(msg events.Message) (types.ContainerJSON, error) {
	var containerID string
	if msg.Action == "connect" {
		containerID = msg.Actor.Attributes["container"]
	} else if msg.Action == "disconnect" || msg.Action == "rename" {
		containerID = msg.Actor.Attributes["name"]
	}

	listArgs := filters.NewArgs()
	listArgs.Add("id", containerID)
	return d.cli.ContainerInspect(context.Background(), containerID)
}

func (d *DockerDNS) renameContainer(oldName string, newName string, endpointSettings map[string]*network.EndpointSettings) {
	for _, domain := range d.domains {
		delete(d.records, dnsutil.Join(strings.Split(oldName, "/")[1], domain))

		newFqdn := dnsutil.Join(newName, domain)
		d.recordsForFQDN(newFqdn, endpointSettings)
	}
}

func (d *DockerDNS) addContainer(name string, endpointSettings map[string]*network.EndpointSettings) {
	for _, domain := range d.domains {
		fqdn := dnsutil.Join(strings.Split(name, "/")[1], domain)
		d.recordsForFQDN(fqdn, endpointSettings)
	}
}

func (d *DockerDNS) removeContainer(name string) {
	for _, domain := range d.domains {
		delete(d.records, dnsutil.Join(strings.Split(name, "/")[1], domain))
	}
}

func (d *DockerDNS) initializeRecords() error {
	containers, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return err
	}

	d.records = make(map[string][]dns.RR)

	for _, container := range containers {
		for _, name := range container.Names {
			d.addContainer(name, container.NetworkSettings.Networks)
		}
	}
	return nil
}

func (d *DockerDNS) recordsForFQDN(fqdn string, endpointSettings map[string]*network.EndpointSettings) {
	generators := []func(string, *network.EndpointSettings) dns.RR{a, aaaa}
	var records []dns.RR
	for _, nt := range endpointSettings {
		for _, generateRecord := range generators {
			if r := generateRecord(fqdn, nt); r != nil {
				records = append(records, r)
			}
		}
	}
	d.records[fqdn] = records
}

func a(name string, nt *network.EndpointSettings) dns.RR {
	return &dns.A{
		A:   net.ParseIP(nt.IPAddress),
		Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
	}
}

func aaaa(name string, nt *network.EndpointSettings) dns.RR {
	if nt.GlobalIPv6Address == "" {
		return nil
	}
	return &dns.AAAA{
		AAAA: net.ParseIP(nt.GlobalIPv6Address),
		Hdr:  dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 0},
	}
}
