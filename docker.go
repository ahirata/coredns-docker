package docker

import (
	"context"
	"io"
	"net"
	"strings"

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
	ips     map[string]dns.RR
}

func NewDockerDNS(domains []string, cli DockerCli) (*DockerDNS, error) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	dockerDNS := DockerDNS{domains: domains, cli: cli}
	dockerDNS.initializeRecords(containers)
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
			if err != nil && err != io.EOF {
				logger.Println("Error while reading docker events channel.", err)
			}
		case msg := <-messages:
			d.handleMessage(msg)
		}
	}
}

func (d *DockerDNS) handleMessage(msg events.Message) {
	container, err := d.containerInspect(msg)
	if err != nil {
		logger.Printf("Error while inspecting the container %v, %v", msg, err)
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

func (d *DockerDNS) initializeRecords(containers []types.Container) {
	d.records = make(map[string][]dns.RR)
	d.ips = make(map[string]dns.RR)

	for _, container := range containers {
		for _, name := range container.Names {
			d.addContainer(name, container.NetworkSettings.Networks)
		}
	}
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
