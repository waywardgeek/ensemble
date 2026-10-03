package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/llm"
)

// Exercise real request rendering, registry dispatch, jobs, and the actor's
// mailbox. Only the provider and the tool's work are fakes. Channel handshakes
// make the stale completion arrive before the real one without sleeps.
func TestActorSendsRealToolResultAfterStaleCompletion(t *testing.T) {
	t.Chdir(t.TempDir())
	release := make(chan struct{})
	var once sync.Once
	finishTool := func() { once.Do(func() { close(release) }) }
	defer finishTool()

	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		request := requests.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\n", `{"type":"message_start","message":{"id":"test","type":"message","role":"assistant","model":"claude-sonnet-5","content":[],"usage":{"input_tokens":1,"output_tokens":1}}}`)
		if request == 1 {
			fmt.Fprintf(w, "data: %s\n\n", `{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"current","name":"blocked_tool","input":{}}}`)
			fmt.Fprintf(w, "data: %s\n\n", `{"type":"content_block_stop","index":0}`)
			fmt.Fprintf(w, "data: %s\n\n", `{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":1}}`)
		} else {
			// With the old wait, this request arrives before the hint has
			// released the tool. Unblock it even on that failure path.
			finishTool()
			if !strings.Contains(string(body), "actual-tool-output") || strings.Contains(string(body), "No result was recorded") {
				t.Error("next provider request did not carry the real tool result")
			}
			fmt.Fprintf(w, "data: %s\n\n", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":"done"}}`)
			fmt.Fprintf(w, "data: %s\n\n", `{"type":"content_block_stop","index":0}`)
			fmt.Fprintf(w, "data: %s\n\n", `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}`)
		}
		fmt.Fprintf(w, "data: %s\n\n", `{"type":"message_stop"}`)
	}))
	defer srv.Close()

	agent := NewBareAgent(Config{Model: "claude-sonnet-5", Vendor: common.VendorAnthropic, BaseURL: srv.URL, APIKey: "test-key"}, "events.jsonl")
	defer agent.Shutdown()
	agent.reg.RegisterInitial(common.Tool{
		Name: "blocked_tool", Description: "Wait for release.",
		Schema: json.RawMessage(`{"type":"object","properties":{}}`),
		Run: func(*common.Call, json.RawMessage) (string, error) {
			<-release
			return "actual-tool-output", nil
		},
	})
	agent.eng.Cfg.Tools = agent.reg.Declarations()
	a := agent.NewActor()
	a.Attach(llm.ObserverFunc(func(obs common.Observation) {
		switch m := obs.(type) {
		case common.ToolDispatched:
			a.Send(common.ToolCompleted{CallID: "old", Result: "late result"})
		case common.ToolFinished:
			if m.CallID == "old" {
				a.Send(common.Hint{Text: "release the current tool"})
			}
		case common.PartDelta:
			if m.Chunk == "hint:release the current tool" {
				finishTool()
			}
		}
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	done := make(chan struct{})
	go func() {
		defer close(done)
		a.Run(ctx)
	}()
	defer func() { finishTool(); cancel(); <-done }()
	a.Send(common.UserMessage{Text: "Run the tool."})
	obs, err := a.Wait(ctx, func(obs common.Observation) bool {
		_, ok := obs.(common.TurnEnded)
		return ok
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	<-done
	if ended := obs.(common.TurnEnded); ended.Err != "" || ended.Text != "done" {
		t.Fatalf("turn ended with %+v", ended)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("provider received %d requests, want 2", got)
	}
	for _, ev := range agent.eng.Log.Events {
		if ev.Type == common.ToolResultLost {
			t.Fatalf("running tool falsely declared lost: %+v", ev.Tool)
		}
	}
}
