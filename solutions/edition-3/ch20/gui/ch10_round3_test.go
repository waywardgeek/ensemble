package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestCh10R3WireTraceRecords(t *testing.T) {
	dir := t.TempDir()
	root := ensemble.New(nil)
	defer root.Shutdown()
	a, err := root.NewAgent(ensemble.Config{Model: "fixture", SavePath: filepath.Join(dir, "save.json"), Tools: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewServerWithSettings(root, "web/gui", filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	path := filepath.Join(dir, "gui.log")
	trace, err := ensemble.OpenDiagnosticLog(path, func() int { return 2 })
	if err != nil {
		t.Fatal(err)
	}
	s.Trace = trace
	web := httptest.NewServer(s.Handler(s.Assets))
	defer web.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(web.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	_ = ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	var reply json.RawMessage
	if err = ws.ReadJSON(&reply); err != nil {
		t.Fatal(err)
	}
	// Consume the second startup frame before sending an acknowledged command.
	if err = ws.ReadJSON(&reply); err != nil {
		t.Fatal(err)
	}
	// Actual pretty-printed inbound traffic must still occupy one trace record.
	if err = ws.WriteMessage(websocket.TextMessage, []byte("{\n  \"type\": \"update_settings\",\n  \"settings\": {\"theme\": \"light\"}\n}")); err != nil {
		t.Fatal(err)
	}
	if err = ws.ReadJSON(&reply); err != nil {
		t.Fatal(err)
	}
	check := func(raw []byte) []string {
		t.Helper()
		if len(raw) == 0 || raw[len(raw)-1] != '\n' {
			t.Fatalf("incomplete trace: %q", raw)
		}
		lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
		for _, line := range lines {
			fields := strings.SplitN(line, " ", 3)
			if len(fields) != 3 {
				t.Fatalf("not timestamp/direction/JSON: %.180s", line)
			}
			if _, err := time.Parse(time.RFC3339Nano, fields[0]); err != nil {
				t.Fatal(err)
			}
			if fields[1] != "<" && fields[1] != ">" {
				t.Fatalf("direction %q", fields[1])
			}
			var compact bytes.Buffer
			if err := json.Compact(&compact, []byte(fields[2])); err != nil {
				t.Fatal(err)
			}
			if compact.String() != fields[2] {
				t.Fatal("payload is not compact JSON")
			}
		}
		return lines
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before := check(raw)
	if len(before) < 4 || !bytes.Contains(raw, []byte(" > {")) || !bytes.Contains(raw, []byte(" < {")) {
		t.Fatalf("missing actual WS traffic: %v", before)
	}
	// Retention is still an explicit successful-save boundary, never Write-time.
	if err = a.Save(); err != nil {
		t.Fatal(err)
	}
	if err = trace.Retain(); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	after := check(raw)
	if len(after) != 2 || strings.Join(after, "\n") != strings.Join(before[len(before)-2:], "\n") {
		t.Fatal("retention split or reordered records")
	}
	if err = ws.WriteMessage(websocket.TextMessage, []byte(`{"type":"update_settings","settings":{"theme":"dark"}}`)); err != nil {
		t.Fatal(err)
	}
	if err = ws.ReadJSON(&reply); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	appended := check(raw)
	if len(appended) != 4 || appended[0] != after[0] || appended[1] != after[1] {
		t.Fatal("post-retention append offset/boundary changed")
	}
}

func TestCh10R3StreamingFinalAfterCoalescedWake(t *testing.T) {
	var requests atomic.Int32
	releases := []chan struct{}{make(chan struct{}), make(chan struct{})}
	defer func() {
		for _, c := range releases {
			select {
			case <-c:
			default:
				close(c)
			}
		}
	}()
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(requests.Add(1)) - 1
		w.Header().Set("Content-Type", "text/event-stream")
		send := func(v string) { fmt.Fprintf(w, "data: %s\n\n", v); w.(http.Flusher).Flush() }
		send(`{"type":"message_start","message":{"model":"claude-sonnet-4-6","usage":{"input_tokens":2}}}`)
		send(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
		send(fmt.Sprintf(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":%q}}`, fmt.Sprintf("head-%d-", n)))
		if n > 0 && n <= len(releases) {
			select {
			case <-releases[n-1]:
			case <-r.Context().Done():
				return
			}
		}
		send(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"tail"}}`)
		send(`{"type":"content_block_stop","index":0}`)
		send(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":3}}`)
		send(`{"type":"message_stop"}`)
	}))
	defer peer.Close()
	dir := t.TempDir()
	cfg := ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "local-fake", BaseURL: peer.URL, StreamCapability: ensemble.StreamSSE, Tools: []string{}, SavePath: filepath.Join(dir, "save.json")}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	seed := ensemble.New(nil)
	old, err := seed.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = old.Ask(ctx, "historical"); err != nil {
		t.Fatal(err)
	}
	if err = seed.Shutdown(); err != nil {
		t.Fatal(err)
	}
	root := ensemble.New(nil)
	defer root.Shutdown()
	a, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	historical, err := json.Marshal(a.Log().Events())
	if err != nil {
		t.Fatal(err)
	}
	lastOld := a.Log().LastSeq()
	s, err := NewServerWithSettings(root, "web/gui", filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	web := httptest.NewServer(s.Handler(s.Assets))
	defer web.Close()
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(web.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	_ = ws.SetReadDeadline(time.Now().Add(8 * time.Second))
	var replay struct {
		Type string
		Data replayEnvelope
	}
	if err = ws.ReadJSON(&replay); err != nil {
		t.Fatal(err)
	}
	used := map[uint64]bool{}
	for _, f := range replay.Data.Events {
		if f.Part != nil && f.Part.Role == "agent" && f.Part.Kind == "text" {
			used[f.PartID] = true
		}
	}
	if len(used) == 0 {
		t.Fatal("historical seed absent")
	}
	for turn := 1; turn <= 2; turn++ {
		done := a.SubmitLive(ctx, fmt.Sprintf("fresh %d", turn))
		var id uint64
		var deltas string
		for id == 0 || deltas == "" {
			var f Wire
			if err = ws.ReadJSON(&f); err != nil {
				t.Fatal(err)
			}
			if f.Type == "part_delta" && f.Kind == "text" {
				if id != 0 && id != f.PartID {
					t.Fatal("delta identity changed")
				}
				id = f.PartID
				deltas += f.Chunk
				if f.AgentID != a.ID() {
					t.Fatal("wrong agent")
				}
			}
		}
		if used[id] {
			t.Fatalf("new response reused historical/prior display ID %d", id)
		}
		// A subscriber can miss an intermediate wake while the actor completes. Hold
		// only its existing lock; HTTP, actor, and reliable completion remain real.
		s.Hub.mu.Lock()
		close(releases[turn-1])
		result := <-done
		s.Hub.mu.Unlock()
		if result.Err != nil {
			t.Fatal(result.Err)
		}
		for {
			var f Wire
			if err = ws.ReadJSON(&f); err != nil {
				t.Fatal(err)
			}
			if f.Type == "part_delta" && f.Kind == "text" {
				if f.AgentID != a.ID() || f.PartID != id {
					t.Fatalf("delta key changed: %+v", f)
				}
				deltas += f.Chunk
			}
			if f.Type == "part_final" && f.Part != nil && f.Part.Role == "agent" && f.Part.Kind == "text" {
				if f.AgentID != a.ID() || f.PartID != id || f.Part.Text != deltas {
					t.Errorf("turn %d: key=%s/%d final=%s/%d deltas=%q final=%q", turn, a.ID(), id, f.AgentID, f.PartID, deltas, f.Part.Text)
				}
				if f.Part.Text != fmt.Sprintf("head-%d-tail", turn) {
					t.Error("incomplete final")
				}
				used[id] = true
				break
			}
		}
	}
	prior, err := json.Marshal(a.Log().EventsRange(0, lastOld))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(prior, historical) {
		t.Fatal("projection changed canonical historical events")
	}
}
