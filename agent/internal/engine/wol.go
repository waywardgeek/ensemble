package engine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"ensemble/internal/common"
)

const wakeProbeTimeout = time.Second
const wakeBootTimeout = 90 * time.Second

// Wake checks belong to the Engine's transport path, using this request's Agent
// configuration and cancellation rather than a separate background wake service.
func (e *engine) wakeUnavailableProvider(ctx context.Context, cfg common.Config) error {
	if cfg.WakeOnLANMAC == "" {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	endpoint, err := providerEndpoint(cfg.BaseURL)
	if err != nil {
		return err
	}
	// An open TCP port or an ICMP reply is enough to skip waking the host.
	// ICMP can succeed while inference is unavailable; neither probe proves model
	// readiness. Only failure of both probes triggers a wake attempt.
	if tcpReachable(ctx, endpoint.address) || icmpReachable(ctx, endpoint.host) {
		return ctx.Err()
	}
	mac, err := wakeMAC(cfg.WakeOnLANMAC)
	if err != nil {
		return errors.New("invalid Wake-on-LAN MAC")
	}
	if err := sendWakePacket(ctx, endpoint.host, mac); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("sending Wake-on-LAN packet: %w", err)
	}
	e.parent.Ensemble().Logf("provider endpoint unavailable; sent Wake-on-LAN packet")
	// After one packet, allow up to 90 seconds for the provider's TCP listener,
	// still honoring earlier caller cancellation. HTTP inference remains the
	// readiness check once the port opens; this wait does not load or test a model.
	deadline, cancel := context.WithTimeout(ctx, wakeBootTimeout)
	defer cancel()
	for {
		if tcpReachable(deadline, endpoint.address) {
			return ctx.Err()
		}
		select {
		case <-deadline.Done():
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errors.New("provider remained unreachable after Wake-on-LAN")
		case <-time.After(wakeProbeTimeout):
		}
	}
}

func wakeMAC(value string) (net.HardwareAddr, error) {
	mac, err := net.ParseMAC(value)
	if err != nil || len(mac) != 6 {
		return nil, errors.New("invalid Wake-on-LAN MAC")
	}
	return mac, nil
}

type providerAddress struct {
	host    string
	address string
}

func providerEndpoint(baseURL string) (providerAddress, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Hostname() == "" {
		return providerAddress{}, errors.New("Wake-on-LAN requires a provider base URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return providerAddress{}, errors.New("Wake-on-LAN requires an http(s) provider URL")
	}
	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}
	return providerAddress{host: u.Hostname(), address: net.JoinHostPort(u.Hostname(), port)}, nil
}

func tcpReachable(ctx context.Context, address string) bool {
	probe, cancel := context.WithTimeout(ctx, wakeProbeTimeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(probe, "tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Missing ping, blocked ICMP and probe timeout all count as failed probes, not
// proof that the host is asleep. TCP is checked first to avoid unnecessary wakes.
func icmpReachable(ctx context.Context, host string) bool {
	probe, cancel := context.WithTimeout(ctx, wakeProbeTimeout)
	defer cancel()
	args := []string{"-c", "1", "-W", "1", host}
	if runtime.GOOS == "darwin" {
		// Darwin's ping expresses this timeout in milliseconds rather than seconds.
		args[3] = "1000"
	}
	return exec.CommandContext(probe, "ping", args...).Run() == nil
}

func sendWakePacket(ctx context.Context, host string, mac net.HardwareAddr) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	addresses, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return err
	}
	for _, address := range addresses {
		if address = address.To4(); address == nil {
			continue
		}
		broadcast, err := broadcastAddress(address)
		if err != nil {
			continue
		}
		conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: broadcast, Port: 9})
		if err != nil {
			continue
		}
		if err := ctx.Err(); err != nil {
			conn.Close()
			return err
		}
		_, err = conn.Write(wakePacket(mac))
		conn.Close()
		if err == nil {
			return nil
		}
	}
	return errors.New("no local broadcast route to provider")
}

// Let the OS route select the local source address without sending a probe.
// Its interface supplies the subnet mask for the directed broadcast, keeping
// network topology in the transport rather than in per-Agent configuration.
func broadcastAddress(destination net.IP) (net.IP, error) {
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: destination, Port: 9})
	if err != nil {
		return nil, err
	}
	local := conn.LocalAddr().(*net.UDPAddr).IP
	conn.Close()
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range addrs {
			network, ok := address.(*net.IPNet)
			if !ok || !network.Contains(local) {
				continue
			}
			broadcast, ok := directedBroadcast(network.IP, network.Mask)
			if !ok {
				continue
			}
			return broadcast, nil
		}
	}
	return nil, errors.New("no local route to provider")
}

func directedBroadcast(ip net.IP, mask net.IPMask) (net.IP, bool) {
	ip = ip.To4()
	if ip == nil || len(mask) != net.IPv4len {
		return nil, false
	}
	broadcast := make(net.IP, net.IPv4len)
	for index := range broadcast {
		broadcast[index] = ip[index] | ^mask[index]
	}
	return broadcast, true
}

func wakePacket(mac net.HardwareAddr) []byte {
	packet := make([]byte, 6+16*len(mac))
	for index := 0; index < 6; index++ {
		packet[index] = 0xff
	}
	for index := 6; index < len(packet); index += len(mac) {
		copy(packet[index:], mac)
	}
	return packet
}
