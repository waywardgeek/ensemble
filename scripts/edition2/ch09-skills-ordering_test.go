package ch09_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"example.com/ensemble"
)

type exchange struct {
	body  []byte
	reply chan []byte
}
type endpoint struct {
	server   *httptest.Server
	requests chan exchange
}

func localEndpoint(t *testing.T) *endpoint {
	t.Helper()
	e := &endpoint{requests: make(chan exchange, 8)}
	e.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		x := exchange{data, make(chan []byte, 1)}
		select {
		case e.requests <- x:
		case <-r.Context().Done():
			return
		}
		select {
		case data = <-x.reply:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(e.server.Close)
	return e
}
func request(t *testing.T, e *endpoint) exchange {
	t.Helper()
	select {
	case x := <-e.requests:
		return x
	case <-time.After(4 * time.Second):
		t.Fatal("model request did not reach controlled endpoint")
		return exchange{}
	}
}
func answer(t *testing.T, vendor string, x exchange, tool string, args map[string]any) {
	t.Helper()
	var body any
	switch vendor {
	case "anthropic":
		content := []any{map[string]any{"type": "text", "text": "done"}}
		stop := "end_turn"
		if tool != "" {
			content = []any{map[string]any{"type": "tool_use", "id": "c1", "name": tool, "input": args}}
			stop = "tool_use"
		}
		body = map[string]any{"model": "claude-sonnet-4-6", "content": content, "stop_reason": stop, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}}
	case "openai":
		message := map[string]any{"role": "assistant", "content": "done"}
		stop := "stop"
		if tool != "" {
			raw, _ := json.Marshal(args)
			message["content"] = ""
			message["tool_calls"] = []any{map[string]any{"id": "c1", "type": "function", "function": map[string]any{"name": tool, "arguments": string(raw)}}}
			stop = "tool_calls"
		}
		body = map[string]any{"model": "gpt-4.1-mini-2025-04-14", "choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": stop}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}}
	case "gemini":
		parts := []any{map[string]any{"text": "done"}}
		if tool != "" {
			parts = []any{map[string]any{"functionCall": map[string]any{"id": "c1", "name": tool, "args": args}}}
		}
		body = map[string]any{"modelVersion": "gemini-3.8-flash", "candidates": []any{map[string]any{"content": map[string]any{"role": "model", "parts": parts}, "finishReason": "STOP"}}, "usageMetadata": map[string]int{"promptTokenCount": 1, "candidatesTokenCount": 1}}
	}
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	x.reply <- data
}
func wireConfig(t *testing.T, vendor string, e *endpoint, name string) ensemble.Config {
	t.Helper()
	c := configuration(t, name)
	c.Vendor = vendor
	c.Model = map[string]string{"anthropic": "claude-sonnet-4-6", "openai": "gpt-4.1-mini-2025-04-14", "gemini": "models/gemini-3.8-flash"}[vendor]
	c.ResolvedModel = strings.TrimPrefix(c.Model, "models/")
	c.BaseURL = e.server.URL
	c.DisableStreaming = true
	return c
}
func oneRequest(t *testing.T, a *ensemble.Agent) {
	t.Helper()
	if _, err := a.UpdatePolicy(a.ExecutionPolicy().Revision, json.RawMessage(`{"max_model_requests":1}`)); err != nil {
		t.Fatal(err)
	}
}
func submit(t *testing.T, a *ensemble.Agent, text string) ensemble.RequestHandle {
	t.Helper()
	h, err := a.Submit(text)
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func complete(t *testing.T, h ensemble.RequestHandle, outcome string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := h.Wait(ctx)
	if err != nil || c.Outcome != outcome {
		t.Fatalf("completion %+v, %v; want %s", c, err, outcome)
	}
}
func changeWhileHeld(t *testing.T, a *ensemble.Agent, load bool) {
	t.Helper()
	result := make(chan error, 1)
	go func() {
		var err error
		if load {
			_, err = a.LoadSkill("edit")
		} else {
			_, err = a.UnloadSkill("edit")
		}
		result <- err
	}()
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("public skill control parked behind held external work")
	}
}
func waitFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(4 * time.Second)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("missing real worker barrier %s", filepath.Base(path))
		}
		time.Sleep(time.Millisecond * 5)
	}
}
func history(t *testing.T, a *ensemble.Agent) []ensemble.Event {
	t.Helper()
	raw, err := a.Dump()
	if err != nil {
		t.Fatal(err)
	}
	var events []ensemble.Event
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var e ensemble.Event
		if err = json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatal(err)
		}
		if e.Type != "" {
			events = append(events, e)
		}
	}
	return events
}
func TestCh09PublicUnloadDuringHeldHTTP(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		t.Run(vendor, func(t *testing.T) {
			e := localEndpoint(t)
			app := ensemble.New(io.Discard)
			defer app.Close()
			c := wireConfig(t, vendor, e, "held")
			a := agent(t, app, c)
			load(t, a)
			oneRequest(t, a)
			h := submit(t, a, "Attempt a controlled write.")
			x := request(t, e)
			if !strings.Contains(string(x.body), "write_file") {
				t.Fatal("positive request did not advertise admitted capability")
			}
			changeWhileHeld(t, a, false)
			answer(t, vendor, x, "write_file", map[string]any{"path": "forbidden.txt", "content": "wrong"})
			complete(t, h, "round_limit")
			if _, err := os.Stat(filepath.Join(c.Workspace, "forbidden.txt")); !os.IsNotExist(err) {
				t.Fatalf("revoked effect occurred: %v", err)
			}
			paired := false
			for _, event := range history(t, a) {
				if event.Type == "tool_returned" && event.Tool.CallID == "c1" {
					paired = event.Tool.IsError
				}
			}
			if !paired {
				t.Fatal("revoked call lacks paired refusal")
			}
		})
	}
}
func TestCh09PublicAdmittedJobSurvivesUnload(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		t.Run(vendor, func(t *testing.T) {
			e := localEndpoint(t)
			app := ensemble.New(io.Discard)
			defer app.Close()
			c := wireConfig(t, vendor, e, "admitted")
			c.Builtins = append(c.Builtins, "run_command")
			c.Skills.Catalog["edit"] = []byte("---\nname: edit\ndescription: Edit\ntype: loadable\ntools: run_command\n---\nCommand manual.\n")
			a := agent(t, app, c)
			load(t, a)
			oneRequest(t, a)
			release := filepath.Join(c.Workspace, "release")
			t.Cleanup(func() { _ = os.WriteFile(release, []byte("go"), 0600) })
			h := submit(t, a, "Run held local operation.")
			x := request(t, e)
			answer(t, vendor, x, "run_command", map[string]any{"command": "printf started > admitted; while [ ! -f release ]; do sleep 0.01; done; printf survived > effect; printf done", "ai_callback_delay": 60})
			waitFile(t, filepath.Join(c.Workspace, "admitted"))
			changeWhileHeld(t, a, false)
			if err := os.WriteFile(release, []byte("go"), 0600); err != nil {
				t.Fatal(err)
			}
			complete(t, h, "round_limit")
			bytes, err := os.ReadFile(filepath.Join(c.Workspace, "effect"))
			if err != nil || string(bytes) != "survived" {
				t.Fatalf("admitted job revoked mid-effect: %q %v", bytes, err)
			}
			for _, event := range history(t, a) {
				if event.Type == "job_killed" {
					t.Fatal("unload killed an already admitted job")
				}
				if event.Type == "tool_returned" && event.Tool.CallID == "c1" && event.Tool.IsError {
					t.Fatal("admitted result reauthorized after unload")
				}
			}
		})
	}
}
func tokens(t *testing.T, vendor string, data []byte) []string {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	field := "messages"
	if vendor == "gemini" {
		field = "contents"
	}
	var out []string
	for _, raw := range body[field].([]any) {
		m := raw.(map[string]any)
		if vendor == "openai" {
			if m["role"] == "tool" {
				out = append(out, "result:"+m["tool_call_id"].(string))
			} else if m["role"] == "user" {
				if text, ok := m["content"].(string); ok {
					out = append(out, text)
				}
			}
			continue
		}
		parts := "content"
		if vendor == "gemini" {
			parts = "parts"
		}
		for _, v := range m[parts].([]any) {
			p := v.(map[string]any)
			if text, ok := p["text"].(string); ok && m["role"] == "user" {
				out = append(out, text)
			}
			if p["type"] == "tool_result" {
				out = append(out, "result:"+p["tool_use_id"].(string))
			}
			if f, ok := p["functionResponse"].(map[string]any); ok {
				out = append(out, "result:"+f["id"].(string))
			}
		}
	}
	return out
}
func suffix(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) < len(want) || !reflect.DeepEqual(got[len(got)-len(want):], want) {
		t.Fatalf("suffix\n got %#v\nwant %#v", got, want)
	}
}
func TestCh09PublicAnchoredHintSkillPromptAndReplay(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		for _, skills := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/skills-%v", vendor, skills), func(t *testing.T) {
				e := localEndpoint(t)
				app := ensemble.New(io.Discard)
				defer app.Close()
				c := wireConfig(t, vendor, e, "anchor")
				c.Builtins = append(c.Builtins, "run_command")
				c.Skills.Catalog["base"] = []byte("---\nname: base\ndescription: Base\ntype: primary\ntools: run_command\nloadable-skills: edit\n---\nPrimary once.\n")
				c.Skills.Catalog["edit"] = []byte("---\nname: edit\ndescription: Edit\ntype: loadable\ntools: write_file\n---\nUse write_file.")
				if !skills {
					c.Skills = nil
					c.Builtins = []string{"run_command", "read_file", "write_file"}
				}
				a := agent(t, app, c)
				oneRequest(t, a)
				release := filepath.Join(c.Workspace, "release")
				t.Cleanup(func() { _ = os.WriteFile(release, []byte("go"), 0600) })
				h := submit(t, a, "Initial P.")
				first := request(t, e)
				early, err := a.Hint("Early H.")
				if err != nil {
					t.Fatal(err)
				}
				answer(t, vendor, first, "run_command", map[string]any{"command": "printf started > admitted; while [ ! -f release ]; do sleep 0.01; done; printf done", "ai_callback_delay": 60})
				waitFile(t, filepath.Join(c.Workspace, "admitted"))
				held, err := a.Hint("Remember H.")
				if err != nil {
					t.Fatal(err)
				}
				if skills {
					changeWhileHeld(t, a, true)
				}
				renderConfig := a.Config()
				if skills {
					renderConfig.System = ""
				}
				if _, err = a.Render(renderConfig); err == nil {
					t.Fatal("unresolved prefix rendered before matched result")
				}
				if err = os.WriteFile(release, []byte("go"), 0600); err != nil {
					t.Fatal(err)
				}
				complete(t, h, "round_limit")
				late, err := a.Hint("Late H.")
				if err != nil {
					t.Fatal(err)
				}
				before, _ := a.Dump()
				r1, err := a.Render(renderConfig)
				if err != nil {
					t.Fatal(err)
				}
				r2, err := a.Render(renderConfig)
				if err != nil || string(r1) != string(r2) {
					t.Fatal("repeated render changed pending material")
				}
				after, _ := a.Dump()
				if string(before) != string(after) {
					t.Fatal("render mutated durable history")
				}
				h = submit(t, a, "Continue P.")
				second := request(t, e)
				manual := "[skill edit activation 2]\nUse write_file.\n[/skill]"
				want := []string{"result:c1", "Early H.", "Remember H.", manual, "Continue P.", "Late H."}
				if !skills {
					want = []string{"result:c1", "Continue P.", "Early H.", "Remember H.", "Late H."}
				}
				suffix(t, tokens(t, vendor, second.body), want)
				answer(t, vendor, second, "", nil)
				complete(t, h, "success")
				h = submit(t, a, "Final P.")
				third := request(t, e)
				want = []string{"result:c1", manual, "Continue P.", "Final P."}
				if !skills {
					want = []string{"result:c1", "Continue P.", "Final P."}
				}
				suffix(t, tokens(t, vendor, third.body), want)
				answer(t, vendor, third, "", nil)
				complete(t, h, "success")
				captured := [][]byte{first.body, second.body, third.body}
				index := 0
				for _, event := range history(t, a) {
					if event.Type != "request_sent" {
						continue
					}
					if event.Request.Configuration == nil {
						t.Fatal("missing captured request configuration")
					}
					if skills && event.Request.Configuration.System != "" {
						t.Fatal("primary duplicated in request capture")
					}
					if !skills && event.Request.Configuration.System == "" {
						t.Fatal("no-skills capture lost inherited System")
					}
					if index == 1 && !reflect.DeepEqual(event.Request.Hints, []uint64{early.Seq, held.Seq, late.Seq}) {
						t.Fatalf("hint capture receipt order: %v", event.Request.Hints)
					}
					raw, err := a.ReconstructRequest(event.Seq)
					if err != nil {
						t.Fatal(err)
					}

					if !bytes.Equal(raw, captured[index]) {
						t.Fatalf("prefix replay differs for request %d", index)
					}
					index++
				}
				if index != 3 {
					t.Fatalf("unexpected request count %d", index)
				}
				// Public offline construction has no skill catalog/variables and no endpoint.
				offline := ensemble.New(io.Discard)
				defer offline.Close()
				reader, err := offline.Load(c.LogPath, ensemble.Config{Vendor: c.Vendor, Model: c.Model, ResolvedModel: c.ResolvedModel, BaseURL: "http://127.0.0.1:1"})
				if err != nil {
					t.Fatal(err)
				}
				index = 0
				for _, event := range history(t, a) {
					if event.Type != "request_sent" {
						continue
					}
					raw, err := reader.ReconstructRequest(event.Seq)
					if err != nil {
						t.Fatal(err)
					}

					if !bytes.Equal(raw, captured[index]) {
						t.Fatal("offline replay did not preserve anchored placement")
					}
					index++
				}
			})
		}
	}
}
