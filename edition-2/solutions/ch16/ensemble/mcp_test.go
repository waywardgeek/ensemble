package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"ensemble/ensemble"
)

// This is an external protocol peer, not a replacement for jobs or dispatch.
// Its wire is independently the chapter's initialize/list/call/content contract.
type mcpPeer struct {
	transport ensemble.MCPTransport
	mu        sync.Mutex
	counts    map[string]int
	requests  chan map[string]json.RawMessage
	responses chan map[string]json.RawMessage
}

func peer(t *testing.T, a ensemble.Agent, name string, held ...string) *mcpPeer {
	t.Helper()
	local, remote := ensemble.MCPPipe()
	p := &mcpPeer{transport: remote, counts: map[string]int{}, requests: make(chan map[string]json.RawMessage, 20), responses: make(chan map[string]json.RawMessage, 20)}
	t.Cleanup(func() { remote.Close() })
	go func() {
		for {
			data, err := remote.Recv()
			if err != nil {
				return
			}
			var m map[string]json.RawMessage
			if json.Unmarshal(data, &m) != nil {
				return
			}
			var method string
			json.Unmarshal(m["method"], &method)
			switch method {
			case "initialize":
				p.reply(m["id"], map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}})
			case "notifications/initialized":
				p.mu.Lock()
				p.counts[method]++
				p.mu.Unlock()
			case "tools/list":
				var tools []map[string]any
				for _, name := range []string{"echo", "read_marker", "snapshot", "startup", "slow", "paired", "forbidden", "broken"} {
					mode := ""
					if name == "snapshot" || name == "broken" {
						mode = "round"
					}
					if name == "startup" {
						mode = "turn"
					}
					tools = append(tools, map[string]any{"name": name, "description": name, "inputSchema": map[string]any{"type": "object"}, "ephemeral": mode})
				}
				p.reply(m["id"], map[string]any{"tools": tools})
			case "tools/call":
				var params struct {
					// Name selects the external peer operation.
					Name string `json:"name"`
					// Arguments supplies the actual tool inputs.
					Arguments map[string]any `json:"arguments"`
				}
				json.Unmarshal(m["params"], &params)
				p.mu.Lock()
				p.counts[params.Name]++
				n := p.counts[params.Name]
				p.mu.Unlock()
				if params.Name == "slow" || params.Name == "paired" || slices.Contains(held, params.Name) {
					p.requests <- m
					continue
				}
				text := fmt.Sprintf("%s-%d", params.Name, n)
				if params.Name == "echo" {
					text = fmt.Sprint(params.Arguments["text"])
				}
				p.reply(m["id"], map[string]any{"content": []map[string]string{{"type": "text", "text": text}}, "isError": params.Name == "broken"})
			case "$/cancelRequest":
				p.requests <- m
			case "":
				p.responses <- m
			}
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := a.MCP().Connect(ctx, name, local); err != nil {
		t.Fatal(err)
	}
	return p
}

func (p *mcpPeer) reply(id json.RawMessage, result any) {
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
	p.transport.Send(data)
}
func mcpAgent(t *testing.T, url string) ensemble.Agent {
	t.Helper()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", BaseURL: url, BuiltinTools: true, Workspace: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	return a
}

func receiveRPC(t *testing.T, channel <-chan map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	select {
	case message := <-channel:
		return message
	case <-time.After(3 * time.Second):
		t.Fatal("MCP peer did not receive message")
		return nil
	}
}

// TestMCPLoopEphemeraAndReverse crosses the real HTTP, actor, job and MCP seams.
func TestMCPLoopEphemeraAndReverse(t *testing.T) {
	var requests []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		requests = append(requests, body)
		if len(requests) == 1 {
			io.WriteString(w, `{"content":[{"type":"tool_use","id":"call","name":"echo","input":{"text":"remote-result"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			io.WriteString(w, `{"content":[{"type":"text","text":"done"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	a := mcpAgent(t, server.URL)
	p := peer(t, a, "test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for range 2 {
		if text, err := a.Ask(ctx, "Use remote tool"); err != nil || text != "done" {
			t.Fatalf("turn %q %v", text, err)
		}
	}
	if len(requests) != 3 {
		t.Fatalf("rounds = %d", len(requests))
	}
	for i, request := range requests {
		wire, _ := json.Marshal(request)
		if strings.Contains(string(wire), "broken-") {
			t.Error("failed ephemeral result was injected")
		}
		if !strings.Contains(string(wire), fmt.Sprintf("snapshot-%d", i+1)) {
			t.Errorf("round %d missing fresh snapshot", i+1)
		}
		if i == 1 && strings.Contains(string(wire), "startup-") {
			t.Error("turn ephemera repeated in second round")
		}
		if i != 1 && !strings.Contains(string(wire), fmt.Sprintf("startup-%d", 1+i/2)) {
			t.Error("turn ephemera absent")
		}
		tools, _ := json.Marshal(request["tools"])
		if strings.Contains(string(tools), "snapshot") || strings.Contains(string(tools), "startup") {
			t.Error("ephemeral tools exposed to model")
		}
	}
	// Inspect the result block, not the earlier call arguments carrying the same
	// marker. Otherwise dropping every outgoing argument can falsely pass.
	resultSeen := false
	for _, message := range requests[1]["messages"].([]any) {
		for _, block := range message.(map[string]any)["content"].([]any) {
			value := block.(map[string]any)
			if value["type"] == "tool_result" {
				data, _ := json.Marshal(value["content"])
				if strings.Contains(string(data), "remote-result") {
					resultSeen = true
				}
			}
		}
	}
	if !resultSeen {
		t.Error("MCP result never reached real renderer")
	}
	if _, err := a.Invoke(ctx, "snapshot", nil); err == nil {
		t.Error("reverse caller invoked hidden automatic tool")
	}
	path := filepath.Join(a.Config().Workspace, "reverse.txt")
	os.WriteFile(path, []byte("reverse-file-content"), 0600)
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 901, "method": "tools/call", "params": map[string]any{"name": "read_file", "arguments": map[string]string{"path": path}}})
	p.transport.Send(data)
	reply := receiveRPC(t, p.responses)
	if string(reply["id"]) != "901" || !strings.Contains(string(reply["result"]), "reverse-file-content") {
		t.Fatalf("reverse reply: %s", reply)
	}
	p.mu.Lock()
	initialized := p.counts["notifications/initialized"]
	broken := p.counts["broken"]
	p.mu.Unlock()
	if initialized != 1 || broken != 3 {
		t.Fatalf("handshake/ephemeral error path: %d %d", initialized, broken)
	}
}

// TestMCPCorrelationKillAndDisconnect separates request identity and job lifetime.
func TestMCPCorrelationKillAndDisconnect(t *testing.T) {
	a := mcpAgent(t, "http://unused")
	p := peer(t, a, "test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	results := make(chan string, 2)
	for _, label := range []string{"first", "second"} {
		go func(label string) {
			out, err := a.Invoke(ctx, "paired", json.RawMessage(fmt.Sprintf(`{"text":%q}`, label)))
			results <- label + ":" + out + fmt.Sprint(err)
		}(label)
	}
	first, second := receiveRPC(t, p.requests), receiveRPC(t, p.requests)
	for _, request := range []map[string]json.RawMessage{second, first} {
		var params struct {
			// Arguments preserves each concurrent request's independent label.
			Arguments map[string]string `json:"arguments"`
		}
		json.Unmarshal(request["params"], &params)
		p.reply(request["id"], map[string]any{"content": []map[string]string{{"type": "text", "text": params.Arguments["text"]}}})
	}
	for range 2 {
		result := <-results
		if result != "first:first<nil>" && result != "second:second<nil>" {
			t.Fatalf("mis-correlated response %q", result)
		}
	}
	out, err := a.Invoke(ctx, "slow", json.RawMessage(`{"ai_callback_delay":0}`))
	if err != nil || !strings.Contains(out, "running") {
		t.Fatalf("callback %q %v", out, err)
	}
	request := receiveRPC(t, p.requests)
	var handle int
	for _, event := range a.History().EventsAfter(0) {
		if event.Tool != nil && event.Tool.Name == "slow" && event.Tool.Job != nil {
			handle = event.Tool.Job.Handle
		}
	}
	job, err := a.Jobs().Find(handle)
	if err != nil {
		t.Fatal(err)
	}
	if job.Context().Err() != nil {
		t.Fatal("callback delay canceled execution")
	}
	if err := job.Kill("test"); err != nil {
		t.Fatal(err)
	}
	cancellation := receiveRPC(t, p.requests)
	var params map[string]json.RawMessage
	json.Unmarshal(cancellation["params"], &params)
	if string(cancellation["method"]) != `"$/cancelRequest"` || string(params["id"]) != string(request["id"]) {
		t.Fatalf("wrong cancellation: %s", cancellation)
	}
	p.transport.Close()
	select {
	case <-a.MCP().Done("test"):
	case <-ctx.Done():
		t.Fatal("disconnect did not wake connection")
	}
	if slices.Contains(toolNames(a), "echo") {
		t.Error("disconnected tool still advertised")
	}
}

// This fixture is the first response from the successful October 10 Anthropic
// live MCP run. Empty partial_json is real wire data, not a guessed edge case.
func TestMCPAnthropicEmptyToolArguments(t *testing.T) {
	stream, err := os.ReadFile("testdata/anthropic-empty-tool-args.sse")
	if err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "text/event-stream")
		if requests == 1 {
			w.Write(stream)
			return
		}
		io.WriteString(w, `event: message_start
data: {"type":"message_start","message":{"id":"reply","model":"claude-sonnet-5","usage":{"input_tokens":1,"output_tokens":1}}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":"done"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_stop
data: {"type":"message_stop"}

`)
	}))
	defer server.Close()
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: server.URL, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	peer(t, a, "test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if text, err := a.Ask(ctx, "read marker"); err != nil || text != "done" {
		t.Fatalf("captured empty arguments: %q %v", text, err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d", requests)
	}
	for _, event := range a.History().EventsAfter(0) {
		if event.Type == "tool_called" && event.Tool.Name == "read_marker" && string(event.Tool.Args) == "{}" {
			return
		}
	}
	t.Fatal("captured empty arguments did not reach real MCP dispatch")
}
