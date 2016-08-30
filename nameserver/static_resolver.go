package main

import (
	"fmt"

	"github.com/dgryski/go-randsample"
	"github.com/miekg/dns"
)

// StaticResolver resolves DNS requests from static in-memory config
type StaticResolver struct {
	IPv4Proxies  []string
	IPv6Proxies  []string
	Domain2Onion map[string]string
	Nameservers  []string
	AnswerCount  int
}

// Start does nothing
func (r *StaticResolver) Start() {
	if r.AnswerCount == 0 {
		panic("StaticResolver: set AnswerCount")
	}
}

// Resolve fetches result value for DNS request from memory
func (r *StaticResolver) Resolve(
	domain string,
	qtype, qclass uint16,
) (
	[]string,
	error,
) {
	var proxies []string
	if qtype == dns.TypeA {
		proxies = r.IPv4Proxies
	} else if qtype == dns.TypeAAAA {
		proxies = r.IPv6Proxies
	} else if qtype == dns.TypeNS {
		onion, ok := r.Domain2Onion[domain]
		if !ok {
			return nil, fmt.Errorf("NS request of unknown domain: %q", domain)
		}
		return r.MakeNameservers(onion)
	} else {
		return nil, fmt.Errorf("Unknown question type: %d", qtype)
	}
	n := len(proxies)
	if n == 0 {
		return nil, fmt.Errorf("No proxies for question of type %d", qtype)
	}
	k := r.roundAnswerCount(n)
	var result []string
	for _, i := range randsample.Sample(n, k) {
		address := proxies[i]
		result = append(result, address)
	}
	return result, nil
}

func (r *StaticResolver) roundAnswerCount(n int) int {
	if n < r.AnswerCount {
		return n
	}
	return r.AnswerCount
}

// MakeNameservers makes hostnames of nameservers for given onion (FQDN)
func (r *StaticResolver) MakeNameservers(onion string) ([]string, error) {
	n := len(r.Nameservers)
	if n == 0 {
		return nil, fmt.Errorf("No known nameservers")
	}
	k := r.roundAnswerCount(n)
	var result []string
	for _, i := range randsample.Sample(n, k) {
		ns := onion + r.Nameservers[i]
		result = append(result, ns)
	}
	return result, nil
}
