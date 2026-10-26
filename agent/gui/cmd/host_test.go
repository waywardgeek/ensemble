package main

import (
	"net"
	"net/url"
	"reflect"
	"testing"
)

func TestListenURLs(test *testing.T) {
	for _, entry := range []struct {
		host string
		want string
	}{
		{"127.0.0.1", "http://127.0.0.1:8088"},
		{"127.0.0.2", "http://127.0.0.2:8088"},
		{"::1", "http://[::1]:8088"},
		{"2001:db8::1", "http://[2001:db8::1]:8088"},
		{"localhost", "http://localhost:8088"},
	} {
		test.Run(entry.host, func(test *testing.T) {
			if got := listenURLs(entry.host, 8088); !reflect.DeepEqual(got, []string{entry.want}) {
				test.Errorf("URLs=%v, want %s", got, entry.want)
			}
		})
	}
	for _, host := range []string{"0.0.0.0", "::"} {
		test.Run(host, func(test *testing.T) {
			addresses := listenURLs(host, 8088)
			if len(addresses) == 0 {
				test.Fatal("wildcard listener has no usable URL")
			}
			for index, address := range addresses {
				parsed, err := url.Parse(address)
				if err != nil {
					test.Fatal(err)
				}
				ip := net.ParseIP(parsed.Hostname())
				if ip == nil || ip.IsUnspecified() || (ip.To4() == nil) != (net.ParseIP(host).To4() == nil) || (ip.To4() == nil && ip.IsLinkLocalUnicast()) || parsed.Port() != "8088" {
					test.Errorf("unusable URL %q for %s", address, host)
				}
				if index == 0 && !ip.IsLoopback() {
					test.Error("wildcard listener omitted its local URL")
				}
			}
		})
	}
}

func TestAdvertisableIP(test *testing.T) {
	for _, entry := range []struct {
		address  string
		listener string
		want     bool
	}{
		{"192.0.2.1", "0.0.0.0", true},
		{"169.254.1.1", "0.0.0.0", true},
		{"2001:db8::1", "::", true},
		{"fd00::1", "::", true},
		{"fe80::1", "::", false},
		{"::1", "::", false},
		{"::", "::", false},
		{"192.0.2.1", "::", false},
		{"2001:db8::1", "0.0.0.0", false},
		{"invalid", "::", false},
	} {
		test.Run(entry.address+"/"+entry.listener, func(test *testing.T) {
			if got := advertisableIP(net.ParseIP(entry.address), net.ParseIP(entry.listener)); got != entry.want {
				test.Errorf("advertisable=%t, want %t", got, entry.want)
			}
		})
	}
}
