package homedial

import (
	"testing"
)

func TestMapAddress(t *testing.T) {
	for _, test := range []struct {
		net, addr, addrWant string
	}{{
		net:      "tcp",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "178.128.130.77:443",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "178.128.130.77:443",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io:80",
		addrWant: "178.128.130.77:80",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io.:80",
		addrWant: "178.128.130.77:80",
	}, {
		net:      "tcp6",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "fabrics.homedrive.io:443",
	}, {
		net:      "udp",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "fabrics.homedrive.io:443",
	}} {
		got := mapAddress(test.net, test.addr)
		if got != test.addrWant {
			t.Errorf(
				"map net=%s addr=%q, got %q, want %q",
				test.net, test.addr, test.addrWant, got,
			)
		}
	}
}
