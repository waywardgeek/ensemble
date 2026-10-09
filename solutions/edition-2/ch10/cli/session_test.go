package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type terminalCapture struct {
	mu      sync.Mutex
	data    bytes.Buffer
	changed chan struct{}
}

func (c *terminalCapture) Write(data []byte) (int, error) {
	c.mu.Lock()
	n, err := c.data.Write(data)
	c.mu.Unlock()
	select {
	case c.changed <- struct{}{}:
	default:
	}
	return n, err
}
func (c *terminalCapture) text() string { c.mu.Lock(); defer c.mu.Unlock(); return c.data.String() }
func (c *terminalCapture) await(t *testing.T, text string) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		if strings.Contains(c.text(), text) {
			return
		}
		select {
		case <-c.changed:
		case <-deadline:
			t.Fatalf("missing %q: %s", text, c.text())
		}
	}
}
func TestHumanControlsRemainReadableDuringBlockedRequest(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var mu sync.Mutex
	count := 0
	var second string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		mu.Lock()
		count++
		n := count
		if n == 2 {
			second = string(data)
		}
		mu.Unlock()
		if n == 1 {
			close(started)
			<-release
		}
		fmt.Fprint(w, `{"content":[{"type":"text","text":"second-answer"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	cliConfig(t, server.URL)
	reader, writer := io.Pipe()
	out := &terminalCapture{changed: make(chan struct{}, 1)}
	var diagnostics bytes.Buffer
	done := make(chan error, 1)
	go func() { done <- runArgs([]string{"chat"}, reader, out, &diagnostics) }()
	fmt.Fprintln(writer, "first")
	<-started
	fmt.Fprint(writer, "second\n/hint ONLY-NEXT\n/ephemeral blocked\n/interrupt\n")
	out.await(t, "interrupted=true")
	close(release)
	out.await(t, "second-answer")
	writer.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Accepted r1", "Accepted r2", "Hint received for r1", "Command refused: Agent busy", "Request r1 (interrupted", "Request r2 (success", "Final usage:"} {
		if !strings.Contains(out.text(), want) {
			t.Fatal(want, out.text())
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if !strings.Contains(second, "ONLY-NEXT") || strings.Contains(second, "blocked") {
		t.Fatal(second)
	}
}
func TestProtocolInvalidControlsRecoverAndEOFDrains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	defer server.Close()
	cliConfig(t, server.URL)
	var out, diagnostics bytes.Buffer
	input := `{"kind":"unknown"}
{"kind":"hint","text":"idle"}
{"kind":"interrupt","extra":true}
{"kind":"prompt","text":"one"}
{"kind":"prompt","text":"two"}
`
	if err := runArgs([]string{"protocol"}, strings.NewReader(input), &out, &diagnostics); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if strings.Count(text, `"invalid_control"`) != 3 || strings.Count(text, `"accepted":"prompt"`) != 2 || strings.Count(text, `"completion"`) != 2 || !strings.Contains(text, `"usage"`) {
		t.Fatal(text)
	}
	for _, id := range []string{"r1", "r2"} {
		accept := strings.Index(text, `"accepted":"prompt","request_id":"`+id+`"`)
		complete := strings.Index(text, `"outcome":"success","pending_hints":0,"request_id":"`+id+`"`)
		if accept < 0 || complete < accept {
			t.Fatal("acceptance/completion order", text)
		}
	}
}

func TestCLISessionRestartLocal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"content":[{"type":"text","text":"saved"}],"usage":{ "input_tokens": 1, "output_tokens": 1 }}`)
	}))
	defer server.Close()
	cliConfig(t, server.URL)
	t.Setenv("CH02_LOG", "")
	directory := t.TempDir() + "/session"
	for i := 0; i < 2; i++ {
		var out, diagnostics bytes.Buffer
		err := runArgs([]string{"--session-dir", directory, "protocol"}, strings.NewReader("{\"user\":\"hello\"}\n"), &out, &diagnostics)
		if err != nil {
			t.Fatalf("mount %d: %v\n%s\n%s", i, err, out.String(), diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if err := runArgs([]string{"session", "inspect", directory}, strings.NewReader(""), &out, &diagnostics); err != nil {
		t.Fatalf("inspect: %v; %s", err, diagnostics.String())
	}
}
