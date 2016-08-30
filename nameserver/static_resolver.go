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
		// FIXME code repeat with etcd_resolver.go
		nameservers := []string{"example.com."} // FIXME store in etcd
		result := make([]string, len(nameservers))
		for i := 0; i < len(nameservers); i++ {
			result[i] = onion + nameservers[i]
		}
		return result, nil
	} else {
		return nil, fmt.Errorf("Unknown question type: %d", qtype)
	}
	n := len(proxies)
	if n == 0 {
		return nil, fmt.Errorf("No proxies for question of type %d", qtype)
	}
	k := r.AnswerCount
	if n < k {
		k = n
	}
	var result []string
	for _, i := range randsample.Sample(n, k) {
		address := proxies[i]
		result = append(result, address)
	}
	return result, nil
}
