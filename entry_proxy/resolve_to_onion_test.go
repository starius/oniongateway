package main

import (
	"errors"
	"testing"
	"time"
)

type EmptyMockNsResolver struct{}

func (o EmptyMockNsResolver) LookupNS(hostname string) (
	names []string,
	ttls []uint32,
	err error,
) {
	return []string{}, []uint32{}, nil
}

func TestEmptyMockNsResolver(t *testing.T) {
	resolver := NewHostToOnionResolver(1000)
	resolver.nsResolver = EmptyMockNsResolver{}
	_, err := resolver.ResolveToOnion("example.com")
	if err == nil {
		t.Fatal("Empty NS resolver works, but it must not")
	}
}

type NoOnionsMockNsResolver struct{}

func (o NoOnionsMockNsResolver) LookupNS(hostname string) (
	names []string,
	ttls []uint32,
	err error,
) {
	names = []string{"ns1.example.com", "ns2.eample.com"}
	ttls = []uint32{100, 100}
	return
}

func TestNoOnionsMockNsResolver(t *testing.T) {
	resolver := NewHostToOnionResolver(1000)
	resolver.nsResolver = NoOnionsMockNsResolver{}
	_, err := resolver.ResolveToOnion("example.com")
	if err == nil {
		t.Fatal("No-onions NS resolver works, but it must not")
	}
}

type ThrowingMockNsResolver struct{}

func (o ThrowingMockNsResolver) LookupNS(hostname string) (
	names []string,
	ttls []uint32,
	err error,
) {
	return []string{}, []uint32{}, errors.New("I always throw")
}

func TestThrowingMockNsResolver(t *testing.T) {
	resolver := NewHostToOnionResolver(1000)
	resolver.nsResolver = ThrowingMockNsResolver{}
	_, err := resolver.ResolveToOnion("example.com")
	if err == nil {
		t.Fatal("Throwing NS resolver works, but it must not")
	}
}

type CountingMockNsResolver struct {
	count *int
}

func (o CountingMockNsResolver) LookupNS(hostname string) (
	names []string,
	ttls []uint32,
	err error,
) {
	(*o.count)++
	names = []string{"t3mny6lhnyku4wrd.onion.ns.com"}
	ttls = []uint32{1}
	return
}

func TestCache(t *testing.T) {
	resolver := NewHostToOnionResolver(1000)
	var count int
	resolver.nsResolver = CountingMockNsResolver{
		count: &count,
	}
	_, err := resolver.ResolveToOnion("example.com")
	if err != nil {
		t.Fatalf("Unexpected error in CountingMockNsResolver: %s", err)
	}
	_, err = resolver.ResolveToOnion("example.com")
	if err != nil {
		t.Fatalf("Unexpected error in CountingMockNsResolver: %s", err)
	}
	if count != 1 {
		t.Fatalf("Number of lookups is %d, not 1", count)
	}
}

func TestExpiredCache(t *testing.T) {
	resolver := NewHostToOnionResolver(1000)
	var count int
	resolver.nsResolver = CountingMockNsResolver{
		count: &count,
	}
	_, err := resolver.ResolveToOnion("example.com")
	if err != nil {
		t.Fatalf("Unexpected error in CountingMockNsResolver: %s", err)
	}
	time.Sleep(2 * time.Second)
	_, err = resolver.ResolveToOnion("example.com")
	if err != nil {
		t.Fatalf("Unexpected error in CountingMockNsResolver: %s", err)
	}
	if count != 2 {
		t.Fatalf("Number of lookups is %d, not 2", count)
	}
}
