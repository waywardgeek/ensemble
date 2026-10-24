package gui_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"ensemble"
	"ensemble/gui"
)

func TestCh20PublicEmbeddingAndRelay(t *testing.T) {
	root := ensemble.New(nil)
	defer root.Shutdown()
	dir := t.TempDir()
	agent, err := root.NewAgent(ensemble.Config{Model: "fixture", APIKey: "dummy", BaseURL: "http://127.0.0.1:1", BaseDir: dir, DataDir: filepath.Join(dir, "state"), SavePath: filepath.Join(dir, "save.json")})
	if err != nil {
		t.Fatal(err)
	}
	server, err := gui.New(root, agent, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if _, err := gui.New(root, agent, nil); err == nil {
		t.Fatal("nil logger accepted")
	}
	request := httptest.NewRequest("GET", "http://gui/", nil)
	response := httptest.NewRecorder()
	server.Handler("").ServeHTTP(response, request)
	if response.Code != 200 || !strings.Contains(response.Body.String(), "<html") {
		t.Fatalf("embedded interface: %d", response.Code)
	}
	// Check the actual browser entry, not JS transformed into a classic test script.
	entries := regexp.MustCompile(`<script\b[^>]*\bsrc\s*=\s*["'](?:\./|/)?app\.js["'][^>]*>`).FindAllString(response.Body.String(), -1)
	if len(entries) != 1 || !regexp.MustCompile(`\btype\s*=\s*["']module["']`).MatchString(entries[0]) {
		t.Fatalf("served app.js must have exactly one module entry: %q", entries)
	}
	for _, name := range []string{"app.js", "artifact-scroll.js", "renderers.js"} {
		asset := httptest.NewRecorder()
		server.Handler("").ServeHTTP(asset, httptest.NewRequest("GET", "http://gui/"+name, nil))
		want, err := fs.ReadFile(gui.Assets(), name)
		if err != nil {
			t.Fatal(err)
		}
		if asset.Code != 200 || !bytes.Equal(asset.Body.Bytes(), want) {
			t.Fatalf("served %s differs from the actual module: status %d", name, asset.Code)
		}
		mediaType, _, err := mime.ParseMediaType(asset.Header().Get("Content-Type"))
		if err != nil || (mediaType != "text/javascript" && mediaType != "application/javascript") {
			t.Fatalf("%s has non-JavaScript module MIME %q: %v", name, asset.Header().Get("Content-Type"), err)
		}
	}
	override := t.TempDir()
	if err := os.WriteFile(filepath.Join(override, "index.html"), []byte("disk override"), 0600); err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	server.Handler(override).ServeHTTP(response, httptest.NewRequest("GET", "http://gui/", nil))
	if response.Body.String() != "disk override" {
		t.Fatal("disk override was not served")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	stopped := make(chan error, 1)
	go func() { stopped <- server.ServeMCP(listener) }()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := io.WriteString(conn, `{"jsonrpc":"2.0","id":"same-id","method":"tools/list"}`+"\n"); err != nil {
		t.Fatal(err)
	}
	var reply struct {
		ID    string          `json:"id"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.NewDecoder(conn).Decode(&reply); err != nil {
		t.Fatal(err)
	}
	if reply.ID != "same-id" || len(reply.Error) == 0 {
		t.Fatalf("uncorrelated no-browser reply: %+v", reply)
	}
	server.Close()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("relay did not close")
	}
	if err := agent.Post(ensemble.Hint{Text: "server no longer owns this view"}); err != nil {
		t.Fatalf("server stopped Agent: %v", err)
	}
}

// This intentionally records the remaining reusable-component HOLD gap.
func TestCh20ReusableRendererModules(t *testing.T) {
	for _, name := range []string{"artifact-scroll.js", "renderers.js"} {
		data, err := gui.Assets().Open(name)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(data)
		data.Close()
		if len(body) == 0 {
			t.Fatalf("empty %s", name)
		}
	}
}
