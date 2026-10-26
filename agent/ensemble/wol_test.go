package ensemble_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"ensemble/ensemble"
)

// TestWakeOnLANProviderLifecycle observes real loopback broadcasts and starts the
// HTTP peer only after its configured magic packet arrives. A private network
// namespace can provide UDP port 9 without host privileges or a real wake target.
func TestWakeOnLANProviderLifecycle(t *testing.T) {
	packets, err := net.ListenPacket("udp4", "127.255.255.255:9")
	if err != nil {
		t.Skipf("loopback broadcast receiver unavailable: %v", err)
	}
	defer packets.Close()
	// Missing ping exercises the supported ICMP-probe failure without a mock.
	t.Setenv("PATH", t.TempDir())
	mac := []byte{2, 0, 0, 0, 0, 1}
	wantPacket := append(bytes.Repeat([]byte{0xff}, 6), bytes.Repeat(mac, 16)...)
	for _, scenario := range []string{"ready", "cancel-before", "cancel-boot", "deadline-boot"} {
		t.Run(scenario, func(t *testing.T) {
			reservation, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			address := reservation.Addr().String()
			reservation.Close()
			agent, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{
				Model: "fixture", BaseURL: "http://" + address,
				WakeOnLANMAC: "02:00:00:00:00:01", DataDir: t.TempDir(), DisableStreaming: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer agent.Shutdown()
			wait := 3 * time.Second
			if scenario == "deadline-boot" {
				wait = 500 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(context.Background(), wait)
			defer cancel()
			if scenario == "cancel-before" {
				cancel()
			}
			result := make(chan error, 1)
			go func() {
				events, err := agent.Engine().Send(ctx, []byte(`{"model":"fixture","messages":[],"max_tokens":1024}`), ensemble.StreamCallbacks{})
				if scenario == "ready" && err == nil {
					found := false
					for _, event := range events {
						if event.Response != nil && len(event.Response.Parts) == 1 && event.Response.Parts[0].Text == "ready" {
							found = true
						}
					}
					if !found {
						err = errors.New("provider answer missing after wake")
					}
				}
				result <- err
			}()
			if scenario != "cancel-before" {
				packets.SetReadDeadline(time.Now().Add(2 * time.Second))
				buffer := make([]byte, 256)
				for {
					length, _, err := packets.ReadFrom(buffer)
					if err != nil {
						t.Fatalf("configured provider did not send its magic packet: %v", err)
					}
					if bytes.Equal(buffer[:length], wantPacket) {
						break
					}
				}
			}
			if scenario == "ready" {
				listener, err := net.Listen("tcp4", address)
				if err != nil {
					t.Fatal(err)
				}
				server := &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					io.Copy(io.Discard, request.Body)
					io.WriteString(writer, `{"content":[{"type":"text","text":"ready"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
				})}
				defer server.Close()
				go server.Serve(listener)
			} else if scenario == "cancel-boot" {
				cancel()
			}
			select {
			case err := <-result:
				want := context.Canceled
				if scenario == "ready" {
					want = nil
				} else if scenario == "deadline-boot" {
					want = context.DeadlineExceeded
				}
				if !errors.Is(err, want) {
					t.Errorf("result = %v, want %v", err, want)
				}
			case <-time.After(4 * time.Second):
				t.Fatal("provider wake ignored caller cancellation")
			}
			packets.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			buffer := make([]byte, 256)
			for {
				length, _, err := packets.ReadFrom(buffer)
				if err != nil {
					break
				}
				if bytes.Equal(buffer[:length], wantPacket) {
					t.Errorf("unexpected wake packet after completion/cancellation: %x", buffer[:length])
				}
			}
		})
	}
}
