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
	"testing"
)

type callbackObserver func(Observation)

func (f callbackObserver) Observe(o Observation) { f(o) }
func TestTurnBatchBoundAndFailureRetention(t *testing.T) {
	for _, mode := range []string{"ordinary", "round_limit", "transport"} {
		t.Run(mode, func(t *testing.T) {
			count := 0
			dir := t.TempDir()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				var request map[string]json.RawMessage
				json.NewDecoder(r.Body).Decode(&request)
				if !strings.Contains(string(request["tools"]), "write_file") {
					t.Error("continuation missing declarations")
				}
				if count > 1 && (!strings.Contains(string(request["messages"]), "tool_result") || !strings.Contains(string(request["messages"]), "unavailable")) {
					t.Error("missing complete batch")
				}
				if count > 1 && mode == "transport" {
					w.WriteHeader(503)
					return
				}
				if count > 1 && mode == "ordinary" {
					fmt.Fprint(w, `{"content":[{"type":"text","text":"final"}],"usage":{"input_tokens":2,"output_tokens":3}}`)
					return
				}
				fmt.Fprintf(w, `{"content":[{"type":"text","text":"intermediate"},{"type":"tool_use","id":"bad-%d","name":"disabled","input":{}},{"type":"tool_use","id":"write-%d","name":"write_file","input":{"path":"effect","content":"x","append":true}}],"usage":{"input_tokens":2,"output_tokens":3}}`, count, count)
			}))
			defer server.Close()
			a, err := New(nil).NewAgent(Config{APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"write_file"}})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			result, err := a.Prompt(context.Background(), "work")
			accepted := count
			if mode == "transport" {
				accepted--
			}
			if mode == "ordinary" {
				if err != nil || result.Text != "final" || count != 2 {
					t.Fatal(result, err, count)
				}
			} else if err == nil {
				t.Fatal("missing terminal error")
			}
			if mode == "round_limit" && (count != 16 || a.Events()[len(a.Events())-1].Error.Code != "round_limit") {
				t.Fatal("bad bound")
			}
			effects := 1
			if mode == "round_limit" {
				effects = 16
			}
			data, _ := os.ReadFile(filepath.Join(dir, "effect"))
			if string(data) != strings.Repeat("x", effects) {
				t.Fatal("batch skipped or repeated", string(data))
			}
			if a.Usage() != (Usage{Input: int64(accepted * 2), Output: int64(accepted * 3)}) {
				t.Fatal("accepted usage lost")
			}
			for _, event := range a.Events() {
				if event.Type == "tool_returned" && strings.HasPrefix(event.Tool.CallID, "bad") && !event.Tool.IsError {
					t.Fatal("ordinary failure flag lost")
				}
			}
			dump, _ := a.Dump()
			path := filepath.Join(dir, "replay")
			os.WriteFile(path, dump, 0600)
			loaded, err := New(nil).Load(path, Config{Model: "test"})
			if err != nil {
				t.Fatal(err)
			}
			loaded.Dump()
			loaded.Render(Config{Model: "test"})
			after, _ := os.ReadFile(filepath.Join(dir, "effect"))
			if string(after) != string(data) {
				t.Fatal("replay executed tools")
			}
		})
	}
}
func TestPersistenceExecutionBoundaries(t *testing.T) {
	for _, stage := range []string{"before_call", "after_effect"} {
		t.Run(stage, func(t *testing.T) {
			dir := t.TempDir()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"write","name":"write_file","input":{"path":"effect","content":"changed"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
			}))
			defer server.Close()
			app := New(nil)
			a, err := app.NewAgent(Config{APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"write_file"}})
			if err != nil {
				t.Fatal(err)
			}
			app.Subscribe(a.ID(), callbackObserver(func(o Observation) {
				if (stage == "before_call" && o.Kind == "response_ended") || (stage == "after_effect" && o.Kind == "tool_called") {
					a.log.Close()
				}
			}))
			_, err = a.Prompt(context.Background(), "write")
			if err == nil {
				t.Fatal("persistence failure accepted")
			}
			_, statErr := os.Stat(filepath.Join(dir, "effect"))
			if stage == "before_call" && !os.IsNotExist(statErr) {
				t.Fatal("tool ran before durable dispatch")
			}
			if stage == "after_effect" && (statErr != nil || !strings.Contains(err.Error(), "completion could not be recorded")) {
				t.Fatal("effect or diagnostic missing", err)
			}
			if _, err = a.Prompt(context.Background(), "retry"); err == nil {
				t.Fatal("faulted Agent continued")
			}
		})
	}
}
func TestWorkspaceCapabilityAndLiveDeclarations(t *testing.T) {
	app := New(nil)
	dirs := []string{t.TempDir(), t.TempDir()}
	for i, dir := range dirs {
		os.WriteFile(filepath.Join(dir, "same"), []byte(fmt.Sprint(i)), 0600)
	}
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body json.RawMessage
		json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, string(body))
		if len(bodies)%2 == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"read","name":"read_file","input":{"path":"same"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	for i, dir := range dirs {
		a, err := app.NewAgent(Config{APIKey: "test", Model: "test", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "log"), Builtins: []string{"read_file"}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = a.Prompt(context.Background(), "read"); err != nil {
			t.Fatal(err)
		}
		for _, event := range a.Events() {
			if event.Type == "tool_returned" {
				if got := *event.Tool.Parts[0].Text; got != fmt.Sprint(i) {
					t.Fatal("workspace leaked", got)
				}
			}
		}
		a.Close()
	}
	_, err := app.NewAgent(Config{APIKey: "test", Model: "test", LogPath: filepath.Join(t.TempDir(), "bad"), Tools: []ToolDefinition{{Name: "invented", Schema: json.RawMessage(`{"type":"object"}`)}}})
	if err == nil {
		t.Fatal("live arbitrary declarations accepted")
	}
}

func TestGeminiSignedContinuationAndErrorSurface(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if count == 1 {
			fmt.Fprint(w, `{"modelVersion":"resolved-model","candidates":[{"content":{"parts":[{"functionCall":{"name":"inspect","args":{"path":"x"}},"thoughtSignature":"signed"}]}}],"usageMetadata":{"promptTokenCount":150,"cachedContentTokenCount":30,"candidatesTokenCount":25,"thoughtsTokenCount":15}}`)
			return
		}
		var request json.RawMessage
		json.NewDecoder(r.Body).Decode(&request)
		for _, want := range []string{`"thoughtSignature":"signed"`, `"id":"call-3-0"`, `"name":"inspect"`, `"error":`} {
			if !strings.Contains(string(request), want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(string(request), `"tools"`) {
			t.Error("empty Registry emitted declarations")
		}
		fmt.Fprint(w, `{"modelVersion":"resolved-model","candidates":[{"content":{"parts":[{"text":""}]}}],"usageMetadata":{"promptTokenCount":0,"candidatesTokenCount":0}}`)
	}))
	defer server.Close()
	a, err := New(nil).NewAgent(Config{Vendor: "gemini", APIKey: "test", Model: "alias", ResolvedModel: "resolved-model", BaseURL: server.URL, LogPath: filepath.Join(t.TempDir(), "log")})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	answer, err := a.Prompt(context.Background(), "inspect")
	if err != nil || answer.Text != "" || count != 2 {
		t.Fatal(answer, err, count)
	}
	if got := a.Events()[2].Response.Parts[0].CallID; got != "call-3-0" {
		t.Fatal("synthesized ID", got)
	}
	if a.Usage() != (Usage{Input: 120, CacheRead: 30, Output: 40}) {
		t.Fatal(a.Usage())
	}
}

func TestErrorRenderingEverySurface(t *testing.T) {
	from := Provenance{Vendor: "anthropic", Model: "test", Surface: "messages"}
	a, err := New(nil).NewAgent(Config{APIKey: "test", Model: "test", LogPath: filepath.Join(t.TempDir(), "log")})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	events := []Event{
		{Type: "message_received", Message: &Entry{Actor: "human", Purpose: "dialogue", Parts: []Part{Text("read")}}},
		{Type: "response_ended", Response: &Response{From: from, Parts: []Part{{Type: "tool_call", CallID: "failed", From: &from, Name: "read_file", Args: json.RawMessage(`{"path":"missing"}`)}}, Usage: &Usage{}}},
		{Type: "tool_returned", Tool: &ToolEvent{CallID: "failed", Parts: []Part{Text("actionable path diagnostic")}, IsError: true}},
	}
	for _, event := range events {
		if err = a.Append(event); err != nil {
			t.Fatal(err)
		}
	}
	for vendor, marker := range map[string]string{"anthropic": `"is_error":true`, "openai": "Tool failed: actionable path diagnostic", "gemini": `"error":"actionable path diagnostic"`} {
		body, err := a.Render(Config{Vendor: vendor, Model: "test"})
		if err != nil || !strings.Contains(string(body), marker) {
			t.Fatal(vendor, string(body), err)
		}
	}
}
