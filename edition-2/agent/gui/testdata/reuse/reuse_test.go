package reuse_test

import (
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Public construction and actual actor history protect the boundary that a
// callback fixture would bypass. Only the remote HTTP provider is replaced.
func TestPublicEvents(t *testing.T) {
	requests := make(chan string, 1)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- string(body)
		io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"EXTERNAL_REPLY"},"finish_reason":"stop"}],"usage":{"prompt_tokens":2,"completion_tokens":1}}`)
	}))
	defer provider.Close()
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o-mini", BaseURL: provider.URL, APIKey: "fixture", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	err = a.History().Append(ensemble.Event{Type: "message_received", Message: &ensemble.MessageData{Actor: ensemble.Human, Parts: []ensemble.Part{{Type: "text", Text: "PLANTED_PUBLIC_EVENT"}}}})
	if err != nil {
		t.Fatal(err)
	}
	hub, err := gui.New(root, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	server := httptest.NewServer(hub.Handler(""))
	defer server.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	ws.WriteJSON(map[string]string{"type": "subscribe"})
	planted := false
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		planted = planted || strings.Contains(string(data), "PLANTED_PUBLIC_EVENT")
		if strings.Contains(string(data), "current_settings") {
			break
		}
	}
	if !planted {
		t.Fatal("public planted event absent from replay")
	}
	ws.WriteJSON(map[string]string{"type": "prompt", "text": "PUBLIC_PROMPT"})
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "EXTERNAL_REPLY") {
			break
		}
	}
	select {
	case request := <-requests:
		if !strings.Contains(request, "PUBLIC_PROMPT") || !strings.Contains(request, "PLANTED_PUBLIC_EVENT") {
			t.Fatalf("real actor request lost public input: %s", request)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("prompt did not reach provider")
	}
}

// Both asset sources are exercised positively, then independent embedded modules
// are mounted. An independent browser imports the component without app.js.
func TestReusableAssets(t *testing.T) {
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o-mini", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	hub, err := gui.New(root, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	for _, name := range []string{"index.html", "app.js", "artifact-scroll.js", "renderers.js", "speech.js", "mcp.js", "style.css"} {
		if data, err := fs.ReadFile(gui.Assets(), name); err != nil || len(data) == 0 {
			t.Fatalf("embedded %s: %v", name, err)
		}
	}
	disk := t.TempDir()
	os.WriteFile(filepath.Join(disk, "index.html"), []byte("DISK_OVERRIDE"), 0600)
	for _, directory := range []string{"", disk} {
		r := httptest.NewRecorder()
		hub.Handler(directory).ServeHTTP(r, httptest.NewRequest("GET", "/", nil))
		if r.Code != 200 {
			t.Fatalf("asset status %s: %d", directory, r.Code)
		}
		if directory == "" && !strings.Contains(r.Body.String(), `<title>Ensemble</title>`) {
			t.Fatal("embedded default returned no application shell")
		}
		if (directory == disk) != strings.Contains(r.Body.String(), "DISK_OVERRIDE") {
			t.Fatal("disk override did not select its explicit source")
		}
	}
	components := http.NewServeMux()
	for _, name := range []string{"artifact-scroll.js", "renderers.js", "mcp.js"} {
		components.HandleFunc("/"+name, func(w http.ResponseWriter, r *http.Request) {
			data, err := fs.ReadFile(gui.Assets(), name)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.Header().Set("Content-Type", "text/javascript")
			w.Write(data)
		})
	}
	server := httptest.NewServer(components)
	defer server.Close()
	response, err := http.Get(server.URL + "/index.html")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatal("component-only server exposed shell")
	}
	cmd := exec.Command("node", "components.test.mjs")
	cmd.Dir = os.Getenv("GUI_SOURCE")
	cmd.Env = append(os.Environ(), "ASSET_URL="+server.URL)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("standalone browser: %v\n%s", err, out)
	}
}

// Raw requests retain their original ID even when clients reuse IDs and browser
// replies arrive in reverse order. The real core MCP handshake shares the wire.
func TestPublicRelay(t *testing.T) {
	root := ensemble.New(io.Discard)
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-4o-mini", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	hub, err := gui.New(root, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ended := make(chan error, 1)
	go func() { ended <- hub.ServeMCP(listener) }()
	connect := func() net.Conn {
		c, e := net.Dial("tcp", listener.Addr().String())
		if e != nil {
			t.Fatal(e)
		}
		c.SetDeadline(time.Now().Add(5 * time.Second))
		t.Cleanup(func() { c.Close() })
		return c
	}
	read := func(c net.Conn) map[string]any {
		var value map[string]any
		if err := json.NewDecoder(c).Decode(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	first := connect()
	defer first.Close()
	fmt.Fprintln(first, `{"jsonrpc":"2.0","id":"same","method":"tools/list"}`)
	value := read(first)
	if value["id"] != "same" || value["error"] == nil {
		t.Fatalf("no-browser correlated error: %v", value)
	}
	server := httptest.NewServer(hub.Handler(""))
	defer server.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	ws.WriteJSON(map[string]string{"type": "subscribe"})
	// Answer the actual Agent's initialize/tools/list before exercising relay.
	for initialized := false; !initialized; {
		var frame map[string]json.RawMessage
		if err := ws.ReadJSON(&frame); err != nil {
			t.Fatal(err)
		}
		var rpc map[string]any
		if json.Unmarshal(frame["payload"], &rpc) != nil || rpc["id"] == nil {
			continue
		}
		result := map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]string{"name": "boundary fake", "version": "1"}}
		if rpc["method"] == "tools/list" {
			result = map[string]any{"tools": []any{}}
			initialized = true
		}
		ws.WriteJSON(map[string]any{"type": "jsonrpc", "payload": map[string]any{"jsonrpc": "2.0", "id": rpc["id"], "result": result}})
	}
	second := connect()
	defer second.Close()
	fmt.Fprintln(first, `{"jsonrpc":"2.0","id":"same","method":"tools/list","params":{"marker":"first"}}`)
	fmt.Fprintln(second, `{"jsonrpc":"2.0","id":"same","method":"tools/list","params":{"marker":"second"}}`)
	var pending []map[string]any
	for len(pending) < 2 {
		var frame map[string]json.RawMessage
		if err := ws.ReadJSON(&frame); err != nil {
			t.Fatal(err)
		}
		var rpc map[string]any
		if json.Unmarshal(frame["payload"], &rpc) == nil && rpc["id"] != nil {
			pending = append(pending, rpc)
		}
	}
	if pending[0]["id"] == pending[1]["id"] {
		t.Fatal("relay IDs collide")
	}
	for i := 1; i >= 0; i-- {
		rpc := pending[i]
		ws.WriteJSON(map[string]any{"type": "jsonrpc", "payload": map[string]any{"jsonrpc": "2.0", "id": rpc["id"], "result": rpc["params"]}})
	}
	for i, c := range []net.Conn{first, second} {
		value := read(c)
		want := []string{"first", "second"}[i]
		if value["id"] != "same" || value["result"].(map[string]any)["marker"] != want {
			t.Fatalf("relay correlation %s: %v", want, value)
		}
	}
	// Browser disconnect must return the pending error with its original ID.
	// Then Server shutdown must wake a new silent-browser request and join it.
	fmt.Fprintln(first, `{"jsonrpc":"2.0","id":7,"method":"tools/list"}`)
	for {
		var frame map[string]json.RawMessage
		if err := ws.ReadJSON(&frame); err != nil {
			t.Fatal(err)
		}
		if frame["payload"] != nil {
			break
		}
	}
	ws.Close()
	value = read(first)
	if value["id"] != float64(7) || value["error"] == nil {
		t.Fatalf("browser disconnect lost request correlation: %v", value)
	}
	ws, _, err = websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	fmt.Fprintln(first, `{"jsonrpc":"2.0","id":8,"method":"tools/list"}`)
	var pendingClose map[string]json.RawMessage
	if err := ws.ReadJSON(&pendingClose); err != nil {
		t.Fatal(err)
	}
	if pendingClose["payload"] == nil {
		t.Fatal("shutdown probe did not reach browser")
	}
	closed := make(chan struct{})
	go func() { hub.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("relay shutdown stuck")
	}
	select {
	case <-ended:
	case <-time.After(time.Second):
		t.Fatal("relay listener survived close")
	}
}
