package ensemble

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type streamObservations struct {
	mu       sync.Mutex
	values   []Observation
	fragment chan struct{}
	once     sync.Once
}

func (o *streamObservations) Observe(v Observation) {
	o.mu.Lock()
	o.values = append(o.values, v)
	o.mu.Unlock()
	if v.Kind == "part_delta" {
		o.once.Do(func() { close(o.fragment) })
	}
}
func streamWait(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("barrier not reached")
	}
}
func TestPublicStreamBarrierAcceptanceAndReplay(t *testing.T) {
	fragment := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["stream"] != true {
			t.Error("missing stream flag")
		}
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Hel\"}}]}\n\n")
		w.(http.Flusher).Flush()
		select {
		case <-fragment:
		case <-r.Context().Done():
			return
		}
		<-release
		fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"length\"}]}\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":2}}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()
	app := New(nil)
	defer app.Close()
	a, err := app.NewAgent(Config{Vendor: "openai", APIKey: "fixture", Model: "fixture", BaseURL: server.URL, LogPath: filepath.Join(t.TempDir(), "log")})
	if err != nil {
		t.Fatal(err)
	}
	observer := &streamObservations{fragment: fragment}
	app.Subscribe(a.ID(), observer)
	h, err := a.Submit("hello")
	if err != nil {
		t.Fatal(err)
	}
	streamWait(t, fragment)
	if a.Usage().Output != 0 {
		t.Fatal("provisional usage committed")
	}
	for _, e := range a.Events() {
		if e.Type == "response_ended" {
			t.Fatal("accepted before terminal")
		}
	}
	close(release)
	c, err := h.Wait(context.Background())
	if err != nil || c.Text != "Hello" || c.StopReason != "length" || c.Usage.Output != 2 {
		t.Fatalf("%#v %v", c, err)
	}
	// Wait for the nonblocking subscription to consume the final actor fact.
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		observer.mu.Lock()
		ended := false
		for _, o := range observer.values {
			ended = ended || o.Kind == "model_end"
		}
		observer.mu.Unlock()
		if ended {
			break
		}
		time.Sleep(time.Millisecond)
	}
	observer.mu.Lock()
	defer observer.mu.Unlock()
	begin, end, final := 0, 0, 0
	id := ""
	for _, o := range observer.values {
		switch o.Kind {
		case "model_begin":
			begin++
			id = o.OperationID
		case "part_delta":
			if o.OperationID != id || o.PartID < 1 {
				t.Fatal("bad correlation")
			}
		case "part_final":
			final++
			if o.PartID != 1 || o.ResponseSeq == 0 || *o.Part.Text != "Hello" {
				t.Fatal(o)
			}
		case "model_end":
			end++
			if !o.Accepted {
				t.Fatal(o)
			}
		}
	}
	if begin != 1 || end != 1 || final != 1 {
		t.Fatal(begin, end, final)
	}
	for _, e := range a.Events() {
		if e.Type == "request_sent" {
			if e.Request.Delivery != "stream" {
				t.Fatal(e)
			}
			body, err := a.ReconstructRequest(e.Seq)
			if err != nil || !strings.Contains(string(body), `"include_usage":true`) {
				t.Fatal(string(body), err)
			}
		}
	}
}
func TestIncompleteProposedCallHasNoEffect(t *testing.T) {
	for _, interrupt := range []bool{false, true} {
		t.Run(fmt.Sprint(interrupt), func(t *testing.T) {
			fragment := make(chan struct{})
			release := make(chan struct{})
			dir := t.TempDir()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(w, `data: {"type":"message_start","message":{"usage":{"input_tokens":1,"output_tokens":0}}}

data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"write-1","name":"write_file","input":{"path":"forbidden.txt","content":"never"}}}

data: {"type":"content_block_stop","index":0}

`)
				w.(http.Flusher).Flush()
				select {
				case <-release:
				case <-r.Context().Done():
				}
			}))
			defer server.Close()
			app := New(nil)
			defer app.Close()
			a, err := app.NewAgent(Config{APIKey: "fixture", Model: "fixture", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"write_file"}})
			if err != nil {
				t.Fatal(err)
			}
			app.Subscribe(a.ID(), &streamObservations{fragment: fragment})
			h, _ := a.Submit("write")
			streamWait(t, fragment)
			if interrupt {
				a.Interrupt()
			}
			close(release)
			c, err := h.Wait(context.Background())
			if err != nil || c.Outcome == "success" {
				t.Fatal(c, err)
			}
			if _, err = os.Stat(filepath.Join(dir, "forbidden.txt")); !os.IsNotExist(err) {
				t.Fatal("proposed tool acted")
			}
			if a.Usage().Output != 0 || a.Usage().Input != 0 {
				t.Fatal("incomplete usage accepted")
			}
			for _, e := range a.Events() {
				if e.Type == "response_ended" || e.Type == "tool_called" {
					t.Fatal("incomplete effect", e.Type)
				}
			}
		})
	}
}
