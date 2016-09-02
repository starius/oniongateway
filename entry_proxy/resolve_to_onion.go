package main

import (
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strings"

	"github.com/coocood/freecache"
	"github.com/miekg/dns"
)

type NsResolver interface {
	LookupNS(string) ([]string, []uint32, error)
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

func (r RealNsResolver) LookupNS(hostname string) (
	names []string,
	ttls []uint32,
	err error,
) {
	parentDomain, err := getParentDomain(hostname)
	if err != nil {
		err = fmt.Errorf("Can't get %s's parent domain: %s", hostname, err)
		return
	}
	nss, err := net.LookupNS(parentDomain)
	if err != nil {
		err = fmt.Errorf("Unable to get %s's NS: %s", parentDomain, err)
		return
	}
	if len(nss) == 0 {
		err = fmt.Errorf("There is no NS records for %s", parentDomain)
		return
	}
	parentNs := nss[rand.Intn(len(nss))].Host
	parentServer := net.JoinHostPort(parentNs, "53")
	message := new(dns.Msg)
	message.SetQuestion(dns.Fqdn(hostname), dns.TypeNS)
	message.RecursionDesired = false
	in, err := dns.Exchange(message, parentServer)
	if err != nil {
		err = fmt.Errorf(
			"Unable to get %s's NS from %s: %s",
			hostname,
			parentServer,
			err,
		)
		return
	}
	for _, record := range in.Ns {
		if nsRecord, ok := record.(*dns.NS); ok {
			names = append(names, nsRecord.Ns)
			ttls = append(ttls, nsRecord.Hdr.Ttl)
		}
	}
	return
}

type HostToOnionResolver struct {
	regex      *regexp.Regexp
	nsResolver NsResolver
	cache      *freecache.Cache
}

func NewHostToOnionResolver(cacheSize int) HostToOnionResolver {
	var err error
	o := HostToOnionResolver{
		nsResolver: RealNsResolver{},
		cache:      freecache.NewCache(cacheSize),
	}
	o.regex, err = regexp.Compile("[a-z0-9]{16}.onion")
	if err != nil {
		panic("wtf: failed to compile regex")
	}
	return o
}

func (o *HostToOnionResolver) ResolveToOnion(hostname string) (onion string, err error) {
	key := []byte(hostname)
	if cachedOnion, err := o.cache.Get(key); err == nil {
		return string(cachedOnion), nil
	}
	nss, ttls, err := o.nsResolver.LookupNS(hostname)
	if err != nil {
		return
	}
	if len(nss) == 0 {
		err = fmt.Errorf("No NS records for %s", hostname)
		return
	}
	for i, ns := range nss {
		match := o.regex.FindString(ns)
		if match != "" {
			value := []byte(match)
			ttl := int(ttls[i])
			_ = o.cache.Set(key, value, ttl)
			return match, nil
		}
	}
	return "", fmt.Errorf("No suitable NS records for %s", hostname)
}
