package main

import (
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strings"

	"github.com/miekg/dns"
)

type NsResolver interface {
	LookupNS(string) ([]string, error)
}

type RealNsResolver struct{}

func getParentDomain(domain string) (string, error) {
	parts := strings.SplitN(domain, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("Failed to split %s to subdomains", domain)
	}
	if parts[1] == "" {
		return "", fmt.Errorf("%s is TLD", domain)
	}
	return parts[1], nil
}

func (r RealNsResolver) LookupNS(hostname string) ([]string, error) {
	parentDomain, err := getParentDomain(hostname)
	if err != nil {
		return nil, fmt.Errorf("Can't get %s's parent domain: %s", hostname, err)
	}
	nss, err := net.LookupNS(parentDomain)
	if err != nil {
		return nil, fmt.Errorf("Unable to get %s's NS: %s", parentDomain, err)
	}
	if len(nss) == 0 {
		return nil, fmt.Errorf("There is no NS records for %s", parentDomain)
	}
	parentNs := nss[rand.Intn(len(nss))].Host
	parentServer := net.JoinHostPort(parentNs, "53")
	message := new(dns.Msg)
	message.SetQuestion(dns.Fqdn(hostname), dns.TypeNS)
	message.RecursionDesired = false
	in, err := dns.Exchange(message, parentServer)
	if err != nil {
		return nil, fmt.Errorf(
			"Unable to get %s's NS from %s: %s",
			hostname,
			parentServer,
			err,
		)
	}
	var result []string
	for _, record := range in.Ns {
		if nsRecord, ok := record.(*dns.NS); ok {
			result = append(result, nsRecord.Ns)
		}
	}
	return result, nil
}

type HostToOnionResolver struct {
	regex       *regexp.Regexp
	nsResolver NsResolver
}

func NewHostToOnionResolver() HostToOnionResolver {
	var err error
	o := HostToOnionResolver{
		nsResolver: RealNsResolver{},
	}
	o.regex, err = regexp.Compile("[a-z0-9]{16}.onion")
	if err != nil {
		panic("wtf: failed to compile regex")
	}
	return o
}

func (o *HostToOnionResolver) ResolveToOnion(hostname string) (onion string, err error) {
	nss, err := o.nsResolver.LookupNS(hostname)
	if err != nil {
		return
	}
	if len(nss) == 0 {
		err = fmt.Errorf("No NS records for %s", hostname)
		return
	}
	for _, ns := range nss {
		match := o.regex.FindString(ns)
		if match != "" {
			return match, nil
		}
	}
	return "", fmt.Errorf("No suitable NS records for %s", hostname)
}
