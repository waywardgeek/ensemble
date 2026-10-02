package ws

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// fakeBrowser stands in for a GUI tab running mcp.js: it subscribes so the hub
// marks it live, then answers every jsonrpc request by echoing the source tag,
// which is the one thing mcp.js must do for routing to work.
func fakeBrowser(t *testing.T, url, name string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial hub: %v", err)
	}
	conn.WriteJSON(map[string]string{"type": "subscribe"})
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var env struct {
				Type    string          `json:"type"`
				Source  string          `json:"source"`
				Payload json.RawMessage `json:"payload"`
			}
			if json.Unmarshal(data, &env) != nil || env.Type != "jsonrpc" {
				continue
			}
			var req rpcHead
			json.Unmarshal(env.Payload, &req)
			if req.Method == "" || req.ID == nil {
				continue
			}
			conn.WriteJSON(map[string]any{
				"type":   "jsonrpc",
				"source": env.Source,
				"payload": map[string]any{
					"jsonrpc": "2.0",
					"id":      req.ID,
					"result":  map[string]string{"from": name, "method": req.Method},
				},
			})
		}
	}()
	return conn
}

// startMCPPort runs a hub behind a real HTTP server and a real TCP listener.
func startMCPPort(t *testing.T) (h *Hub, wsURL string, tcpAddr string) {
	t.Helper()
	h = NewHub(nil, nil, "", &common.Log{}, nil, nil)
	srv := httptest.NewServer(http.HandlerFunc(h.ServeWS))
	t.Cleanup(srv.Close)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go h.ServeMCP(ln)
	return h, "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws", ln.Addr().String()
}

func dialMCP(t *testing.T, addr string) (net.Conn, *bufio.Reader) {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn, bufio.NewReader(conn)
}

func readLine(t *testing.T, conn net.Conn, r *bufio.Reader) map[string]any {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	line, err := r.ReadBytes('\n')
	if err != nil {
		t.Fatalf("read reply: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(line, &m); err != nil {
		t.Fatalf("reply is not JSON: %q", line)
	}
	return m
}

// waitLive blocks until the hub has n live clients, so a request is not sent
// before the fake browsers' subscribe messages have been processed.
func waitLive(t *testing.T, h *Hub, n int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		h.mu.Lock()
		live := 0
		for c := range h.clients {
			if c.live {
				live++
			}
		}
		h.mu.Unlock()
		if live == n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("hub never reached %d live clients", n)
}

// A request on the TCP port reaches the browser and the reply comes back on
// the same connection, with the id intact. Two tabs are open, and both answer;
// the client must still see exactly one reply per request. The second request
// proves it: if the duplicate of the first had been forwarded, it would be the
// next line read, and its id would be 1 rather than 2.
func TestMCPPortRoundTripOneReplyPerRequest(t *testing.T) {
	h, wsURL, addr := startMCPPort(t)
	fakeBrowser(t, wsURL, "tab-a")
	fakeBrowser(t, wsURL, "tab-b")
	waitLive(t, h, 2)

	conn, r := dialMCP(t, addr)
	for id := 1; id <= 2; id++ {
		req, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": "tools/list"})
		conn.Write(append(req, '\n'))
		got := readLine(t, conn, r)
		if got["id"] != float64(id) {
			t.Fatalf("request %d: reply has id %v (a duplicate from the second tab leaked through?)", id, got["id"])
		}
		res, _ := got["result"].(map[string]any)
		if res["method"] != "tools/list" {
			t.Fatalf("request %d: reply did not come from the browser: %v", id, got)
		}
	}
}

// With no browser open, nothing will ever answer, so the relay answers at once
// with an error carrying the request's id instead of letting the client hang.
func TestMCPPortNoGUIAnswersImmediately(t *testing.T) {
	_, _, addr := startMCPPort(t)
	conn, r := dialMCP(t, addr)
	conn.Write([]byte(`{"jsonrpc":"2.0","id":"abc","method":"tools/call"}` + "\n"))
	got := readLine(t, conn, r)
	if got["id"] != "abc" {
		t.Fatalf("error reply has id %v, want abc", got["id"])
	}
	e, _ := got["error"].(map[string]any)
	if e == nil || !strings.Contains(e["message"].(string), "no GUI") {
		t.Fatalf("want a no-GUI error, got %v", got)
	}
}

// Closing the TCP connection unregisters its source tag, so a long-running
// server does not accumulate receivers for clients that have gone.
func TestMCPPortDisconnectRemovesReceiver(t *testing.T) {
	h, _, addr := startMCPPort(t)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	count := func() int {
		h.mu.Lock()
		defer h.mu.Unlock()
		n := 0
		for tag := range h.mcpReceivers {
			if strings.HasPrefix(tag, "tcp-") {
				n++
			}
		}
		return n
	}
	for i := 0; i < 300 && count() == 0; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if count() != 1 {
		t.Fatalf("connection registered %d receivers, want 1", count())
	}
	conn.Close()
	for i := 0; i < 300 && count() != 0; i++ {
		time.Sleep(10 * time.Millisecond)
	}
	if count() != 0 {
		t.Fatalf("receiver still registered after disconnect")
	}
}
