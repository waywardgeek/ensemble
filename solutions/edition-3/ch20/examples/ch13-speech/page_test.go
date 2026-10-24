package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// Exercise the real page transport without a browser or provider. Diagnostics
// must identify the fixed CDP operation, retain its cause, and omit payloads.
func TestPageCommandFailureMetadata(t *testing.T) {
	const secret = "diagnostic-private-parameter-and-body"
	for _, stage := range []string{"read", "write", "rejected"} {
		t.Run(stage, func(t *testing.T) {
			upgrade := websocket.Upgrader{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgrade.Upgrade(w, r, nil)
				if err != nil {
					return
				}
				defer conn.Close()
				var request struct {
					ID int `json:"id"`
				}
				if conn.ReadJSON(&request) != nil {
					return
				}
				if stage == "rejected" {
					conn.WriteJSON(map[string]any{"id": request.ID, "error": map[string]any{"code": -32000, "message": secret, "data": secret}})
				}
			}))
			defer server.Close()
			conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if stage == "write" {
				conn.Close()
			}
			err = (&page{conn: conn}).call(context.Background(), "Page.navigate", map[string]any{"url": secret}, nil)
			if err == nil {
				t.Fatal("missing page failure")
			}
			if !strings.Contains(err.Error(), "Page.navigate") || !strings.Contains(err.Error(), stage) {
				t.Errorf("failure lacks method/stage: %v", err)
			}
			if strings.Contains(err.Error(), secret) {
				t.Errorf("diagnostic disclosed params/protocol body: %v", err)
			}
			switch stage {
			case "read":
				var cause *websocket.CloseError
				if !errors.As(err, &cause) || cause.Code != websocket.CloseAbnormalClosure {
					t.Errorf("read cause not preserved: %v", err)
				}
				if !strings.Contains(err.Error(), "unexpected EOF") {
					t.Errorf("read cause not actionable: %v", err)
				}
			case "write":
				if !errors.Is(err, net.ErrClosed) {
					t.Errorf("write cause not preserved: %v", err)
				}
			}
		})
	}
}

func TestPageCommandCancelledCause(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (&page{}).call(ctx, "Page.enable", nil, nil)
	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "Page.enable") {
		t.Fatalf("cancelled command metadata/cause: %v", err)
	}
}
