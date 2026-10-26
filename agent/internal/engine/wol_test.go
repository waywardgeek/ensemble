package engine

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
)

func TestWakePacket(t *testing.T) {
	mac, _ := net.ParseMAC("02:00:00:00:00:01")
	packet := wakePacket(mac)
	if len(packet) != 102 || !bytes.Equal(packet[:6], []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
		t.Fatalf("invalid packet header: %x", packet)
	}
	for offset := 6; offset < len(packet); offset += len(mac) {
		if !bytes.Equal(packet[offset:offset+len(mac)], mac) {
			t.Fatalf("packet MAC at %d = %x, want %x", offset, packet[offset:offset+len(mac)], mac)
		}
	}
}

func TestWakeMAC(t *testing.T) {
	if mac, err := wakeMAC("02:00:00:00:00:01"); err != nil || mac.String() != "02:00:00:00:00:01" {
		t.Fatalf("wakeMAC valid result = %v, %v", mac, err)
	}
	for _, input := range []string{"invalid", "02:00:00:00:00", "02:00:00:00:00:01:00:01"} {
		if _, err := wakeMAC(input); err == nil {
			t.Fatalf("wakeMAC accepted %q", input)
		}
	}
}

func TestDirectedBroadcast(t *testing.T) {
	broadcast, ok := directedBroadcast(net.IPv4(192, 168, 99, 161), net.CIDRMask(24, 32))
	if !ok || !broadcast.Equal(net.IPv4(192, 168, 99, 255)) {
		t.Fatalf("directed broadcast = %v, %v", broadcast, ok)
	}
	if _, ok := directedBroadcast(net.IPv6loopback, net.CIDRMask(64, 128)); ok {
		t.Fatal("accepted IPv6 network")
	}
}

func TestSendWakePacket(t *testing.T) {
	mac, _ := wakeMAC("02:00:00:00:00:01")
	if err := sendWakePacket(context.Background(), "127.0.0.1", mac); err != nil {
		t.Fatal(err)
	}
}

func TestSendWakePacketCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mac, _ := wakeMAC("02:00:00:00:00:01")
	if err := sendWakePacket(ctx, "provider.invalid", mac); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled wake = %v", err)
	}
}

func TestProviderEndpoint(t *testing.T) {
	for input, want := range map[string]string{
		"http://hal.localdomain:8110/": "hal.localdomain:8110",
		"https://hal.localdomain":      "hal.localdomain:443",
	} {
		endpoint, err := providerEndpoint(input)
		if err != nil || endpoint.address != want {
			t.Fatalf("providerEndpoint(%q) = %+v, %v", input, endpoint, err)
		}
	}
	for _, input := range []string{"", "hal.localdomain", "ftp://hal.localdomain", "ftp://hal.localdomain:8110"} {
		if _, err := providerEndpoint(input); err == nil {
			t.Fatalf("accepted invalid endpoint %q", input)
		}
	}
}
