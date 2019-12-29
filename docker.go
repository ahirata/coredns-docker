package docker

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/dnsutil"
	"github.com/coredns/coredns/request"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"github.com/miekg/dns"
)

type docker struct {
	next    plugin.Handler
	cli     client.ContainerAPIClient
	domains []string
}

func (d *docker) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	containers, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return dns.RcodeServerFailure, errors.New("failure to list containers")
	}

	state := request.Request{W: w, Req: r}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	switch state.QType() {
	case dns.TypeA:
		m.Answer = d.generateAnswers(state.Name(), containers, d.a)
	case dns.TypeAAAA:
		m.Answer = d.generateAnswers(state.Name(), containers, d.aaaa)
	}
	println("meh")

	if len(m.Answer) == 0 {
		return plugin.NextOrFailure(d.Name(), d.next, ctx, w, r)
	}

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

func (d *docker) a(name string, nt *network.EndpointSettings) dns.RR {
	ip := net.ParseIP(nt.IPAddress)
	r := new(dns.A)
	r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 50}
	r.A = ip
	return r
}

func (d *docker) aaaa(name string, nt *network.EndpointSettings) dns.RR {
	if nt.GlobalIPv6Address == "" {
		return nil
	}
	ip := net.ParseIP(nt.GlobalIPv6Address)
	r := new(dns.AAAA)
	r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 50}
	r.AAAA = ip
	return r
}

func (d *docker) generateAnswers(query string, containers []types.Container, generateRecord func(string, *network.EndpointSettings) dns.RR) []dns.RR {
	var answers []dns.RR
	for _, container := range containers {
		for _, name := range container.Names {
			if fqdn := d.toFQDN(name); fqdn == query {
				for _, nt := range container.NetworkSettings.Networks {
					if r := generateRecord(query, nt); r != nil {
						answers = append(answers, r)
					}
				}
				return answers
			}
		}
	}
	return answers
}

func (d *docker) Name() string { return "docker" }

func (d *docker) toFQDN(name string) string {
	return dnsutil.Join(strings.Split(name, "/")[1], d.domains[0])
}
