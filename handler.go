package docker

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type DockerPlugin struct {
	next      plugin.Handler
	dockerDNS *DockerDNS
}

func (d *DockerPlugin) Name() string { return "docker" }

// ServeDNS implements the plugin.Handler interface.
func (d *DockerPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	answers := d.findAnswer(state.Name(), state.QType())
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers

	if len(answers) == 0 {
		m.Rcode = dns.RcodeNameError
	}

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

func (d *DockerPlugin) findAnswer(query string, qType uint16) []dns.RR {
	var r []dns.RR
	for _, record := range d.dockerDNS.GetRecords()[query] {
		if record.Header().Rrtype == qType {
			r = append(r, record)
		}
	}
	return r
}
