package main

import (
	"fmt"
	"net"
	"regexp"
)

type NsResolver interface {
	LookupNS(string) ([]string, error)
}

type RealNsResolver struct{}

func (r RealNsResolver) LookupNS(hostname string) ([]string, error) {
	nss, err := net.LookupNS(hostname)
	if err != nil {
		return nil, fmt.Errorf("Unable to get %s's NS: %s", hostname, err)
	}
	result := make([]string, len(nss))
	for i := 0; i < len(nss); i++ {
		result[i] = nss[i].Host
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
