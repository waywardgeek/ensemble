package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble"
	agentowner "ensemble/internal/agent"
	"ensemble/internal/common"
	engineowner "ensemble/internal/engine"
	historyowner "ensemble/internal/history"
	jobsowner "ensemble/internal/jobs"
	skillsowner "ensemble/internal/skills"
	toolowner "ensemble/internal/tool"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type callFailureHistory struct {
	common.History
	parent common.Agent
	mode   string
}

func (h *callFailureHistory) Parent() common.Agent { return h.parent }
func (h *callFailureHistory) Append(e common.Event) error {
	if h.mode == "append" && e.Type == common.ToolCalled {
		return errors.New("controlled call append failure")
	}
	return h.History.Append(e)
}
func (h *callFailureHistory) Dump(w io.Writer) error {
	if h.mode == "save" {
		for _, e := range h.History.Events() {
			if e.Type == common.ToolCalled {
				return errors.New("controlled call save failure")
			}
		}
	}
	return h.History.Dump(w)
}

type callReplyEngine struct {
	common.Engine
	call     common.Part
	requests int
}

func (e *callReplyEngine) ExchangeRequest(ctx context.Context, _ []byte, callbacks ...common.ParseCallbacks) ([]common.Event, error) {
	e.requests++
	msg := map[string]any{"content": "done"}
	stop := "stop"
	if e.requests == 1 {
		msg = map[string]any{"content": "I will perform the requested operation.", "tool_calls": []any{map[string]any{"id": e.call.CallID, "type": "function", "function": map[string]any{"name": e.call.Name, "arguments": string(e.call.Args)}}}}
		stop = "tool_calls"
	}
	body, _ := json.Marshal(map[string]any{"model": "actual", "choices": []any{map[string]any{"message": msg, "finish_reason": stop}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
	cb := common.ParseCallbacks{}
	if len(callbacks) > 0 {
		cb = callbacks[0]
	}
	return e.Engine.Parse(&http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(body)))}, cb)
}
func TestCh6R1CallRecordBarrier(t *testing.T) {
	for _, path := range []string{"invoke", "model", "prepare", "dispatch"} {
		for _, failure := range []string{"append", "save", "none"} {
			for _, kind := range []string{"registered", "write_file", "run_command"} {
				t.Run(path+"/"+failure+"/"+kind, func(t *testing.T) {
					dir := t.TempDir()
					logPath := filepath.Join(dir, "agent.jsonl")
					marker := filepath.Join(dir, "effect")
					app := ensemble.New(nil)
					defer app.Shutdown()
					a, err := agentowner.New(app, ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: "http://127.0.0.1:1", BaseDir: dir, DataDir: filepath.Join(dir, "jobs"), LogPath: logPath})
					if err != nil {
						t.Fatal(err)
					}
					transport, err := engineowner.New(a)
					if err != nil {
						t.Fatal(err)
					}
					fake := &callReplyEngine{Engine: transport}
					a.SetEngine(fake)
					a.SetHistory(&callFailureHistory{History: historyowner.New(a), mode: failure, parent: a})
					a.SetJobs(jobsowner.New(a))
					registry, err := toolowner.New(a)
					if err != nil {
						t.Fatal(err)
					}
					a.SetTools(registry)
					skillRegistry, err := skillsowner.New(a, registry.Catalog())
					if err != nil {
						t.Fatal(err)
					}
					a.SetSkills(skillRegistry)
					if err := skillRegistry.LoadInitial(); err != nil {
						t.Fatal(err)
					}

					a.Start()
					defer a.Shutdown()
					var effects atomic.Int32
					if err := a.Tools().Register(ensemble.ToolDefinition{Name: "registered", Description: "Controlled effect.", Parameters: json.RawMessage(`{"type":"object"}`)}, func(_ context.Context, _ ensemble.ToolCall, _ json.RawMessage) error {
						effects.Add(1)
						// Deliberately do not cooperate with cancellation: a mistakenly started
						// handler cannot be hidden by cancelling its observation or by shutdown.
						if err := os.WriteFile(marker, []byte("effect"), 0600); err != nil {
							return err
						}
						for _, event := range a.Log().Events() {
							if event.Type == ensemble.ToolCalled && event.Tool.Name == "registered" {
								saved, err := os.ReadFile(logPath)
								if err == nil && strings.Contains(string(saved), `"tool_called"`) {
									return nil
								}
							}
						}
						return errors.New("handler started without a saved call")
					}); err != nil {
						t.Fatal(err)
					}
					args := json.RawMessage(`{}`)
					switch kind {
					case "write_file":
						args, _ = json.Marshal(map[string]any{"path": marker, "content": "effect"})
					case "run_command":
						args, _ = json.Marshal(map[string]any{"command": fmt.Sprintf("printf effect > '%s'", marker)})
					}
					call := common.Part{Type: "tool_call", CallID: "barrier-call", Name: kind, Args: args}
					ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
					defer cancel()
					switch path {
					case "invoke":
						err = a.Tools().Invoke(ctx, call, "")
					case "model":
						fake.call = call
						_, err = a.Ask(ctx, "perform operation")
					case "prepare":
						d := a.Tools().Prepare(ctx, call, "", 0).Await(ctx)
						if d.IsError {
							err = errors.New(d.Parts[0].Text)
						}
					case "dispatch":
						parts, failed := a.Tools().Dispatch(ctx, call.Name, call.Args)
						if failed {
							err = errors.New(parts[0].Text)
						}
					}
					if failure == "none" {
						if err != nil {
							t.Fatal(err)
						}
						if _, err := os.Stat(marker); err != nil {
							t.Fatalf("successful effect missing: %v", err)
						}
						calls, returns := 0, 0
						for _, event := range a.Log().Events() {
							if event.Type == ensemble.ToolCalled {
								calls++
								if event.Tool.Job == nil || event.Tool.Job.Handle <= 0 {
									t.Fatal("call missing reserved handle")
								}
							}
							if event.Type == ensemble.ToolReturned {
								returns++
								if event.Tool.IsError {
									t.Fatalf("successful ordering failed: %+v", event.Tool)
								}
							}
						}
						if calls != 1 || returns != 1 {
							t.Fatalf("pairing calls=%d returns=%d", calls, returns)
						}
					} else {
						if err == nil {
							t.Error("record failure reported success")
						}
						for _, job := range app.JobTable() {
							if job.Data().Status == ensemble.JobRunning {
								t.Errorf("failed admission left an unstarted live reservation before shutdown: %+v", job)
							}
						}
						done := make(chan struct{})
						go func() { _ = a.Shutdown(); close(done) }()
						select {
						case <-done:
						case <-time.After(2 * time.Second):
							t.Fatal("shutdown blocked on unstarted job")
						}
						if effects.Load() != 0 {
							t.Errorf("handler effects before failed record: %d", effects.Load())
						}
						if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
							t.Errorf("file/process effect before failed record: %v", err)
						}
						calls, returns := 0, 0
						for _, event := range a.Log().Events() {
							if event.Type == ensemble.ToolCalled {
								calls++
							}
							if event.Type == ensemble.ToolReturned {
								returns++
								if !event.Tool.IsError {
									t.Error("failed admission recorded success")
								}
							}
						}
						if failure == "append" && (calls != 0 || returns != 0) {
							t.Errorf("invented pairing after failed append: %d/%d", calls, returns)
						}
						if failure == "save" && (calls != 1 || returns != 1) {
							t.Errorf("unpaired appended call after failed save: %d/%d", calls, returns)
						}
						for _, job := range app.JobTable() {
							if job.Data().Status != ensemble.JobDone {
								t.Errorf("failed admission left live/false-killed job: %+v", job.Data())
							}
						}
					}
				})
			}
		}
	}
}
func TestCh6R1EphemeraReplacement(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		requests <- string(b)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: server.URL, SystemPrompt: "stable primary"})
	if err != nil {
		t.Fatal(err)
	}
	post := func(s string) {
		t.Helper()
		if err := a.Post(ensemble.Ephemeral{Text: s}); err != nil {
			t.Fatal(err)
		}
	}
	post("old directive")
	post("replacement directive")
	if _, err := a.Ask(context.Background(), "first"); err == nil {
		t.Fatal("expected local HTTP failure")
	}
	if len(a.Log().Context().Ephemera) != 0 {
		t.Error("failed request did not consume ephemera")
	}
	post("clear me")
	post("")
	if _, err := a.Ask(context.Background(), "second"); err == nil {
		t.Fatal("expected local HTTP failure")
	}
	observed := []string{<-requests, <-requests}
	if strings.Contains(observed[0], "old directive") || !strings.Contains(observed[0], "replacement directive") || strings.Contains(observed[1], "directive") || strings.Contains(observed[1], "clear me") {
		t.Fatalf("actual requests: %q", observed)
	}
	for _, b := range observed {
		if !strings.Contains(b, "stable primary") {
			t.Fatalf("primary changed: %s", b)
		}
	}
	updates := 0
	for _, e := range a.Log().Events() {
		if e.Type == ensemble.MessageReceived && e.Message.Actor == ensemble.System {
			updates++
		}
	}
	if updates != 4 {
		t.Fatalf("lost append-only updates: %d", updates)
	}
}

func TestCh6R1EphemeraProjection(t *testing.T) {
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", SystemPrompt: "stable primary"})
	if err != nil {
		t.Fatal(err)
	}
	h := historyowner.New(a)
	for _, text := range []string{"old directive", "replacement directive", ""} {
		if err := h.Append(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.System, Parts: common.PartList{common.Text(text)}}}); err != nil {
			t.Fatal(err)
		}
		c := h.Context()
		if text == "" {
			if len(c.Ephemera) != 0 {
				t.Fatal(c.Ephemera)
			}
		} else if len(c.Ephemera) != 1 || c.Ephemera[0].Text != text {
			t.Fatal(c.Ephemera)
		}
		body, err := a.Engine().Render(c)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "stable primary") {
			t.Fatal(string(body))
		}
		if text == "replacement directive" && (strings.Contains(string(body), "old directive") || !strings.Contains(string(body), text)) {
			t.Fatal(string(body))
		}
		if text == "" && strings.Contains(string(body), "directive") {
			t.Fatal(string(body))
		}
	}
	if len(h.Events()) != 3 {
		t.Fatal("lost append-only updates")
	}
}
