package config

import (
	"net"
	"testing"
)

func TestGatewayHexToIP(t *testing.T) {
	t.Parallel()
	ip, err := gatewayHexToIP("010011AC")
	if err != nil {
		t.Fatal(err)
	}
	if !ip.Equal(net.IPv4(172, 17, 0, 1)) {
		t.Fatalf("got %s", ip)
	}
}

func TestDedupeHosts(t *testing.T) {
	t.Parallel()
	got := dedupeHosts([]string{"a", "b", "a", "", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
}
