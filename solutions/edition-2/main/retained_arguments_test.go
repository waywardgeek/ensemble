package ensemble

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
	"strings"
	"testing"
	"time"
)

// Each provider makes the same four attempted calls. These are actual localhost
// exchanges through Actor; in particular no limit fact or result is fabricated.
func TestRetainedArgumentsContinuationAndReplay(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		for _, session := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/session=%t", vendor, session), func(t *testing.T) {
				calls := []struct{ name, args string }{
					{"tool_limits", `{"max_output_bytes":1}`},
					{"load_skill", `{"name": "edit", "name": "edit"}`},
					{"read_file", `{"path": "note.txt"}`},
					{"load_skill", `{"name": "edit"}`},
				}
				var bodies [][]byte
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
					}
					bodies = append(bodies, body)
					i := len(bodies) - 1
					var response any
					switch vendor {
					case "anthropic":
						part := map[string]any{"type": "text", "text": "continued"}
						if i < len(calls) {
							part = map[string]any{"type": "tool_use", "id": fmt.Sprint("c", i), "name": calls[i].name, "input": json.RawMessage(calls[i].args)}
						}
						response = map[string]any{"model": "fixture", "content": []any{part}, "usage": map[string]int{"input_tokens": 2, "output_tokens": 1}}
					case "openai":
						message := map[string]any{"role": "assistant", "content": "continued"}
						if i < len(calls) {
							message = map[string]any{"role": "assistant", "content": nil, "tool_calls": []any{map[string]any{"id": fmt.Sprint("c", i), "type": "function", "function": map[string]any{"name": calls[i].name, "arguments": calls[i].args}}}}
						}
						response = map[string]any{"model": "fixture", "choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": "stop"}}, "usage": map[string]int{"prompt_tokens": 2, "completion_tokens": 1}}
					case "gemini":
						part := map[string]any{"text": "continued"}
						if i < len(calls) {
							part = map[string]any{"functionCall": map[string]any{"id": fmt.Sprint("c", i), "name": calls[i].name, "args": json.RawMessage(calls[i].args)}}
						}
						response = map[string]any{"modelVersion": "fixture", "candidates": []any{map[string]any{"index": 0, "content": map[string]any{"role": "model", "parts": []any{part}}, "finishReason": "STOP"}}, "usageMetadata": map[string]int{"promptTokenCount": 2, "candidatesTokenCount": 1}}
					}
					if err = json.NewEncoder(w).Encode(response); err != nil {
						t.Error(err)
					}
				}))
				defer server.Close()
				root := New(io.Discard)
				defer root.Close()
				config := persistenceSkillConfig(t)
				config.Vendor = vendor
				config.BaseURL = server.URL
				config.Builtins = []string{"tool_limits", "load_skill", "unload_skill", "read_file", "write_file"}
				config.Skills.Catalog["base"] = []byte("---\nname: base\ndescription: Base\ntype: primary\ntools: read_file tool_limits\nloadable-skills: edit\n---\nIdentity.")
				if err := os.WriteFile(filepath.Join(config.Workspace, "note.txt"), []byte("untruncated ordinary read"), 0600); err != nil {
					t.Fatal(err)
				}
				var a *Agent
				var err error
				if session {
					config.LogPath = ""
					config.DataDir = "session"
					a, err = root.OpenSession(SessionOptions{Config: config})
				} else {
					a, err = root.NewAgent(config)
				}
				if err != nil {
					t.Fatal(err)
				}
				var older []byte
				if session {
					cp, e := a.ExportCheckpoint()
					if e != nil {
						t.Fatal(e)
					}
					older = cp.Bytes
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				answer, err := a.Ask(ctx, "exercise bounded call sequence")
				if err != nil || answer != "continued" {
					t.Fatalf("continuation: %q %v", answer, err)
				}
				if len(bodies) != 5 {
					t.Fatalf("got %d requests", len(bodies))
				}
				called, returned, changes, consumed := 0, 0, 0, 0
				for _, e := range a.Events() {
					if e.Type == "tool_called" {
						called++
						if e.Tool.Name == "load_skill" && called == 2 && e.Tool.Job != nil {
							t.Fatal("invalid management call created Job")
						}
					}
					if e.Type == "tool_returned" {
						returned++
						raw, _ := json.Marshal(e.Tool.Parts)
						if returned == 1 && e.Tool.IsError {
							t.Fatalf("setter failed: %s", raw)
						}
						if returned == 2 && (!e.Tool.IsError || !bytes.Contains(raw, []byte("invalid_skill_arguments"))) {
							t.Fatalf("duplicate result: %s", raw)
						}
						if returned == 2 {
							text := *e.Tool.Parts[0].Text
							var acknowledgment struct {
								Error    string
								Name     string
								Revision uint64
							}
							at := strings.LastIndex(text, "{")
							if at < 0 || json.Unmarshal([]byte(text[at:]), &acknowledgment) != nil || acknowledgment.Error != "invalid_skill_arguments" || acknowledgment.Name != "" || acknowledgment.Revision != 0 {
								t.Fatalf("ambiguous name acknowledgment: %s", text)
							}
						}
						if returned == 3 && (e.Tool.IsError || !bytes.Contains(raw, []byte("untruncated ordinary read"))) {
							t.Fatalf("pending limit leaked into ordinary call: %s", raw)
						}
					}
					if e.Type == "skills_changed" {
						changes++
					}
					if e.Type == "tool_limits_consumed" {
						consumed++
					}
				}
				if called != 4 || returned != 4 || changes != 1 {
					t.Fatalf("paired effects %d %d %d", called, returned, changes)
				}
				if session && consumed != 1 {
					t.Fatalf("consumption count %d", consumed)
				}
				type replaySource interface {
					Events() []Event
					ReconstructRequest(uint64) ([]byte, error)
				}
				check := func(label string, source replaySource) {
					t.Helper()
					n := 0
					for _, e := range source.Events() {
						if e.Type != "request_sent" {
							continue
						}
						got, err := source.ReconstructRequest(e.Seq)
						if err != nil {
							t.Fatalf("%s reconstruct: %v", label, err)
						}
						var actual, expected any
						if json.Unmarshal(got, &actual) != nil || json.Unmarshal(bodies[n], &expected) != nil {
							t.Fatal("bad request JSON")
						}
						ga, _ := json.Marshal(actual)
						gb, _ := json.Marshal(expected)
						if !bytes.Equal(ga, gb) {
							t.Fatalf("%s request%d mismatch\n%s\n%s", label, n, ga, gb)
						}
						if vendor == "openai" && n > 1 && !bytes.Contains(got, []byte(`\"name\": \"edit\"`)) {
							t.Fatal("original arguments STRING whitespace lost")
						}
						n++
					}
					if n != 5 {
						t.Fatalf("%s requests %d", label, n)
					}
				}
				check("accepted", a)
				if err = a.Close(); err != nil {
					t.Fatal(err)
				}
				if !session {
					loaded, e := root.Load(config.LogPath, config)
					if e != nil {
						t.Fatal(e)
					}
					check("standalone full log", loaded)
					return
				}
				directory := filepath.Join(config.Workspace, config.DataDir)
				checkpoint := filepath.Join(directory, "checkpoint.json")
				latest, e := os.ReadFile(checkpoint)
				if e != nil {
					t.Fatal(e)
				}
				snapshot, e := root.InspectCheckpoint(latest)
				if e != nil {
					t.Fatal("semantic checkpoint", e)
				}
				wanted, e := a.Render(a.Config())
				if e != nil {
					t.Fatal(e)
				}
				actual, e := snapshot.Render(a.Config())
				if e != nil || !bytes.Equal(wanted, actual) {
					t.Fatalf("semantic checkpoint render differs: %v", e)
				}

				var null map[string]any
				if json.Unmarshal(latest, &null) != nil {
					t.Fatal("checkpoint JSON")
				}
				null["state"] = nil
				null["state_sha256"] = nil
				nullRaw, _ := json.Marshal(null)
				for _, branch := range []struct {
					name string
					raw  []byte
				}{{"latest", latest}, {"older-tail", older}, {"null-state", nullRaw}, {"full-log", nil}} {
					if branch.raw == nil {
						err = os.Remove(checkpoint)
					} else {
						err = os.WriteFile(checkpoint, branch.raw, 0600)
					}
					if err != nil {
						t.Fatal(err)
					}
					inspected, e := root.InspectSession(directory)
					if e != nil {
						t.Fatalf("%s: %v", branch.name, e)
					}
					check(branch.name, inspected)
				}
				if err = os.WriteFile(checkpoint, latest, 0600); err != nil {
					t.Fatal(err)
				}
				b, e := root.OpenSession(SessionOptions{Config: config})
				if e != nil {
					t.Fatal(e)
				}
				check("reopened", b)
				if b.Usage() != a.Usage() {
					t.Fatal("usage restored twice")
				}
				if _, e = b.Ask(ctx, "continued after reopen"); e != nil {
					t.Fatal(e)
				}
				if e = b.Close(); e != nil {
					t.Fatal(e)
				}
				importedConfig := config
				importedConfig.DataDir = "imported"
				imported, e := root.ImportSession(latest, SessionOptions{Config: importedConfig})
				if e != nil {
					t.Fatal("semantic import", e)
				}
				actual, e = imported.Render(imported.Config())
				if e != nil || !bytes.Equal(wanted, actual) {
					t.Fatalf("import render differs: %v", e)
				}
				if _, e = imported.Ask(ctx, "tail after import"); e != nil {
					t.Fatal(e)
				}
				if e = imported.Close(); e != nil {
					t.Fatal(e)
				}
				tail, e := root.InspectSession(filepath.Join(config.Workspace, "imported"))
				if e != nil {
					t.Fatal(e)
				}
				for _, event := range tail.Events() {
					if event.Type == "request_sent" {
						raw, e := tail.ReconstructRequest(event.Seq)
						if e != nil {
							t.Fatal(e)
						}
						var got, want any
						json.Unmarshal(raw, &got)
						json.Unmarshal(bodies[len(bodies)-1], &want)
						x, _ := json.Marshal(got)
						y, _ := json.Marshal(want)
						if !bytes.Equal(x, y) {
							t.Fatal("imported snapshot-tail reconstruction differs")
						}
					}
				}

			})
		}
	}
}

func TestRetainedChatReplayStringCorrespondenceRefusal(t *testing.T) {
	root := New(io.Discard)
	defer root.Close()
	a, err := root.NewAgent(Config{Vendor: "openai", Model: "fixture", APIKey: "fixture", Workspace: t.TempDir(), LogPath: filepath.Join(t.TempDir(), "log")})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Append(Event{Type: "message_received", Message: &Entry{Actor: "human", Purpose: "dialogue", Parts: []Part{Text("pending")}}}); err != nil {
		t.Fatal(err)
	}
	from := Provenance{Vendor: "openai", Model: "fixture", Surface: "chat_completions"}
	for _, original := range []string{`{"name":"other"}`, `[]`, `{"name":}`} {
		err = a.Append(Event{Type: "response_ended", Response: &Response{From: from, Usage: &Usage{}, Parts: []Part{{Type: "tool_call", CallID: "bad", Name: "load_skill", From: &from, Args: json.RawMessage(`{"name":"edit"}`), ArgumentsText: &original}}}})
		if err == nil || !strings.Contains(err.Error(), "argument replay") {
			t.Fatalf("correspondence refusal: %v", err)
		}
	}
	if err = a.Close(); err != nil {
		t.Fatal("validation refusal became terminal", err)
	}
}
