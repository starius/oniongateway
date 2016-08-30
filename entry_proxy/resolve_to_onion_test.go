package main

import (
	"errors"
	"testing"
)

type EmptyMockNsResolver struct{}

func (o EmptyMockNsResolver) LookupNS(hostname string) ([]string, error) {
	return []string{}, nil
}

func TestEmptyMockNsResolver(t *testing.T) {
	resolver := NewHostToOnionResolver()
	resolver.nsResolver = EmptyMockNsResolver{}
	_, err := resolver.ResolveToOnion("example.com")
	if err == nil {
		t.Fatal("Empty NS resolver works, but it must not")
	}
}

type NoOnionsMockNsResolver struct{}

func (o NoOnionsMockNsResolver) LookupNS(hostname string) ([]string, error) {
	return []string{"ns1.example.com", "ns2.eample.com"}, nil
}

func TestNoOnionsMockNsResolver(t *testing.T) {
	resolver := NewHostToOnionResolver()
	resolver.nsResolver = NoOnionsMockNsResolver{}
	_, err := resolver.ResolveToOnion("example.com")
	if err == nil {
		t.Fatal("No-onions NS resolver works, but it must not")
	}
}

type ThrowingMockNsResolver struct{}

func (o ThrowingMockNsResolver) LookupNS(hostname string) ([]string, error) {
	return []string{}, errors.New("I always throw")
}

func TestThrowingMockNsResolver(t *testing.T) {
	resolver := NewHostToOnionResolver()
	resolver.nsResolver = ThrowingMockNsResolver{}
	_, err := resolver.ResolveToOnion("example.com")
	if err == nil {
		t.Fatal("Throwing NS resolver works, but it must not")
	}
}
