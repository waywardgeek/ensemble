package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func TestChapter09Round1RegistrationValidation(t *testing.T) {
	for _, api := range []string{"Tools.Register", "Skills.RegisterTool"} {
		for _, bad := range []struct{ name, description, schema string }{
			{"blank description", "", `{"type":"object"}`},
			{"whitespace description", " \t", `{"type":"object"}`},
			{"missing type", "custom", `{}`},
			{"nonobject type", "custom", `{"type":"string"}`},
			{"nonstring type", "custom", `{"type":1}`},
		} {
			t.Run(api+"/"+bad.name, func(t *testing.T) {
				a := c9Agent(t, c9Fixture(t), nil)
				var calls atomic.Int32
				handler := func(_ context.Context, c ensemble.ToolCall, _ json.RawMessage) error {
					calls.Add(1)
					_, err := io.WriteString(c, "registered output")
					return err
				}
				register := func(d ensemble.ToolDefinition) error {
					if api == "Tools.Register" {
						return a.Tools().Register(d, handler)
					}
					return a.Skills().RegisterTool(ensemble.ToolSpec{ToolDefinition: d, Handler: handler, Enabled: true})
				}
				before := a.Tools().Definitions()
				wireBefore, err := a.Engine().Render(ensemble.Context{})
				if err != nil {
					t.Fatal(err)
				}
				def := ensemble.ToolDefinition{Name: "round1_custom", Description: bad.description, Parameters: json.RawMessage(bad.schema)}
				if err := register(def); err == nil {
					t.Errorf("accepted invalid registration: %+v", def)
				}
				if !reflect.DeepEqual(before, a.Tools().Definitions()) {
					t.Error("failed registration changed declarations")
				}
				wireAfter, err := a.Engine().Render(ensemble.Context{})
				if err != nil {
					t.Fatal(err)
				}
				if string(wireBefore) != string(wireAfter) {
					t.Error("failed registration changed provider catalog")
				}
				if _, failed := a.Tools().Dispatch(context.Background(), def.Name, json.RawMessage(`{}`)); !failed {
					t.Error("invalid registration became dispatchable")
				}
				if calls.Load() != 0 {
					t.Fatal("invalid handler ran")
				}
				def.Description = "custom"
				def.Parameters = json.RawMessage(`{"type":"object"}`)
				if err := register(def); err != nil {
					t.Fatal(err)
				}
				if _, failed := a.Tools().Dispatch(context.Background(), def.Name, json.RawMessage(`{}`)); failed {
					t.Fatal("valid registration not dispatchable")
				}
				if calls.Load() != 1 || len(a.Tools().Definitions()) != len(before)+1 {
					t.Fatal("valid registration not published exactly once")
				}
			})
		}
	}
}

func c9HTTPAgent(t *testing.T, dir, url string) ensemble.Agent {
	t.Helper()
	app := ensemble.New(nil)
	t.Cleanup(func() { _ = app.Shutdown() })
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", APIKey: "local-fixture-only", BaseURL: url, BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}})
	if err != nil {
		t.Fatal(err)
	}
	return a
}
func c9Tool(t *testing.T, a ensemble.Agent, name string) ensemble.ToolSpec {
	t.Helper()
	for _, tool := range a.Skills().Snapshot().Tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("missing tool %s", name)
	return ensemble.ToolSpec{}
}
func c9LoadEvents(a ensemble.Agent) []ensemble.Event {
	var events []ensemble.Event
	for _, e := range a.Log().EventsRange(0, a.Log().LastSeq()) {
		if e.Type == ensemble.ToolCalled && e.Tool != nil && e.Tool.Name == "load_skill" {
			events = append(events, e)
		}
	}
	return events
}
func c9CheckWire(t *testing.T, b []byte, deferred bool, referenceCall string) {
	t.Helper()
	var wire struct {
		Tools []struct {
			Name     string `json:"name"`
			Deferred bool   `json:"defer_loading"`
		}
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tool := range wire.Tools {
		if tool.Name == "write_file" {
			found = true
			if tool.Deferred != deferred {
				t.Errorf("writer deferred=%v want %v", tool.Deferred, deferred)
			}
		}
		if tool.Name == "read_file" && tool.Deferred {
			t.Error("startup tool became deferred")
		}
	}
	if !found {
		t.Fatal("missing writer catalog entry")
	}
	refs := 0
	uses := map[string]bool{}
	for _, message := range wire.Messages {
		var blocks []struct {
			Type      string          `json:"type"`
			ID        string          `json:"id"`
			ToolUseID string          `json:"tool_use_id"`
			Content   json.RawMessage `json:"content"`
			IsError   bool            `json:"is_error"`
		}
		if err := json.Unmarshal(message.Content, &blocks); err != nil {
			var text string
			if err := json.Unmarshal(message.Content, &text); err != nil {
				t.Fatal(err)
			}
			continue
		}
		for _, block := range blocks {
			if block.Type == "tool_use" && message.Role == "assistant" {
				uses[block.ID] = true
			}
			if block.Type != "tool_result" {
				continue
			}
			var content []struct {
				Type     string `json:"type"`
				ToolName string `json:"tool_name"`
			}
			if json.Unmarshal(block.Content, &content) != nil {
				continue
			}
			if block.ToolUseID == referenceCall && referenceCall != "" {
				if block.IsError {
					t.Error("successful load result became an error")
				}
				for _, part := range content {
					if part.Type != "tool_reference" {
						t.Errorf("reference-bearing result mixed with %q content", part.Type)
					}
				}
			}
			for _, part := range content {
				if part.Type == "tool_reference" && part.ToolName == "write_file" {
					refs++
					if message.Role != "user" || block.ToolUseID != referenceCall || !uses[block.ToolUseID] {
						t.Errorf("reference without its real pair: %s", b)
					}
				}
			}
		}
	}
	want := 0
	if referenceCall != "" {
		want = 1
	}
	if refs != want {
		t.Errorf("writer references=%d want %d", refs, want)
	}
	if referenceCall == "" && len(uses) != 0 {
		t.Error("unexpected fabricated assistant call")
	}
}

func TestChapter09Round1DirectLoadExposure(t *testing.T) {
	for _, path := range []string{"Dispatch", "Invoke", "Manual"} {
		t.Run(path, func(t *testing.T) {
			captured := make(chan []byte, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				captured <- b
				ch5Reply(w, map[string]string{"type": "text", "text": "done"})
			}))
			defer server.Close()
			a := c9HTTPAgent(t, c9Fixture(t), server.URL)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if path == "Manual" {
				if err := a.Skills().LoadSkill("writer"); err != nil {
					t.Fatal(err)
				}
			} else if path == "Invoke" {
				if err := a.Tools().Invoke(ctx, ensemble.Part{Type: "tool_call", Name: "load_skill", CallID: "direct-load", Args: json.RawMessage(`{"name":"writer"}`)}, ""); err != nil {
					t.Fatal(err)
				}
			} else if _, failed := a.Tools().Dispatch(ctx, "load_skill", json.RawMessage(`{"name":"writer"}`)); failed {
				t.Fatal("load failed")
			}
			events := c9LoadEvents(a)
			tool := c9Tool(t, a, "write_file")
			if path == "Manual" {
				if len(events) != 0 || !tool.Enabled || tool.EventSeq != 0 || tool.Origin != "manual" {
					t.Fatalf("manual load invented tool provenance: %+v / %+v", tool, events)
				}
			} else {
				if len(events) != 1 {
					t.Fatalf("real admissions=%d want 1", len(events))
				}
				if !tool.Enabled || tool.EventSeq != events[0].Seq || tool.Origin != "tool" {
					t.Fatalf("provenance not actual admission: %+v / %+v", tool, events)
				}
			}
			// Direct admission is not an assistant-authored model call. The next caller-
			// supplied context has no pair; rendering must neither fabricate one nor hide
			// an enabled capability. Generic orphan history rendering is not repaired here.
			seq := a.Log().LastSeq()
			if _, err := a.Engine().Exchange(ctx, ensemble.Context{}); err != nil {
				t.Fatal(err)
			}
			c9CheckWire(t, <-captured, false, "")
			if a.Log().LastSeq() != seq {
				t.Fatal("render/exchange rewrote History")
			}
			if _, failed := a.Tools().Dispatch(ctx, "load_skill", json.RawMessage(`{"name":"writer"}`)); failed {
				t.Fatal("reload failed")
			}
			again := c9Tool(t, a, "write_file")
			if again.EventSeq != tool.EventSeq || again.Origin != tool.Origin || again.Skill != tool.Skill || again.Source != tool.Source {
				t.Fatalf("reload rebased provenance: %+v", again)
			}
			b, err := a.Engine().Render(ensemble.Context{})
			if err != nil {
				t.Fatal(err)
			}
			c9CheckWire(t, b, false, "")
		})
	}
}

func TestChapter09EnablingEventAndRendering(t *testing.T) {
	for _, skill := range []string{"writer", "broken"} {
		t.Run(skill, func(t *testing.T) {
			captured := make(chan []byte, 3)
			var count atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				captured <- b
				if count.Add(1) == 1 {
					ch5Reply(w, ch5Call("model-load", "load_skill", map[string]string{"name": skill}))
				} else {
					ch5Reply(w, map[string]string{"type": "text", "text": "done"})
				}
			}))
			defer server.Close()
			a := c9HTTPAgent(t, c9Fixture(t), server.URL)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := a.Ask(ctx, "load the skill"); err != nil {
				t.Fatal(err)
			}
			if count.Load() != 2 {
				t.Fatalf("requests=%d", count.Load())
			}
			<-captured
			after := <-captured
			events := c9LoadEvents(a)
			if len(events) != 1 || events[0].Tool.CallID != "model-load" {
				t.Fatalf("actual model admissions: %+v", events)
			}
			tool := c9Tool(t, a, "write_file")
			if skill == "broken" {
				if tool.Enabled || tool.EventSeq != 0 {
					t.Fatalf("failed load published tool: %+v", tool)
				}
				// A real failed call/result may be present, but it grants no reference.
				if strings.Contains(string(after), `"tool_reference"`) {
					t.Fatal("failed load published reference")
				}
				if _, failed := a.Tools().Dispatch(ctx, "write_file", json.RawMessage(`{}`)); !failed {
					t.Fatal("failed load enabled dispatch")
				}
				return
			}
			if !tool.Enabled || tool.EventSeq != events[0].Seq || tool.Origin != "tool" {
				t.Fatalf("provenance: %+v", tool)
			}
			c9CheckWire(t, after, true, "model-load")
			if shared := c9Tool(t, a, "read_file"); shared.EventSeq != 0 || shared.Origin != "initial" {
				t.Fatalf("shared startup provenance changed: %+v", shared)
			}
			history := a.Log().Context()
			status := ""
			for _, entry := range history.Dialogue {
				for _, part := range entry.Parts {
					if part.Type == "tool_result" && part.CallID == "model-load" {
						if part.IsError {
							t.Fatal("canonical successful load marked as error")
						}
						for _, value := range part.Parts {
							if value.Type == "text" {
								status += value.Text
							}
						}
					}
				}
			}
			if !strings.Contains(status, "writer") {
				t.Fatalf("canonical handler status missing: %q", status)
			}
			seq := a.Log().LastSeq()
			canonicalBefore, err := json.Marshal(a.Log().EventsRange(0, seq))
			if err != nil {
				t.Fatal(err)
			}
			contextBefore, err := json.Marshal(history)
			if err != nil {
				t.Fatal(err)
			}
			replay, err := a.Engine().Render(history)
			if err != nil {
				t.Fatal(err)
			}
			c9CheckWire(t, replay, true, "model-load")
			absent, err := a.Engine().Render(ensemble.Context{})
			if err != nil {
				t.Fatal(err)
			}
			c9CheckWire(t, absent, false, "")
			canonicalAfter, err := json.Marshal(a.Log().EventsRange(0, a.Log().LastSeq()))
			if err != nil {
				t.Fatal(err)
			}
			contextAfter, err := json.Marshal(history)
			if err != nil {
				t.Fatal(err)
			}
			if a.Log().LastSeq() != seq || string(canonicalAfter) != string(canonicalBefore) || string(contextAfter) != string(contextBefore) {
				t.Fatal("projection changed canonical History or context")
			}
			if err := a.Skills().UnloadSkill("writer"); err != nil {
				t.Fatal(err)
			}
			pending, err := a.Engine().Render(history)
			// The still-active dependency dep independently grants write_file.
			if err != nil || string(pending) != string(replay) {
				t.Fatal("pending unload changed independently granted replay")
			}
			if _, failed := a.Tools().Dispatch(ctx, "load_skill", json.RawMessage(`{"name":"writer"}`)); failed {
				t.Fatal("reload failed")
			}
			if again := c9Tool(t, a, "write_file"); again.EventSeq != tool.EventSeq || again.Skill != tool.Skill || again.Source != tool.Source || again.Origin != tool.Origin {
				t.Fatalf("reload changed metadata: %+v", again)
			}
			replay, err = a.Engine().Render(history)
			if err != nil {
				t.Fatal(err)
			}
			c9CheckWire(t, replay, true, "model-load")
		})
	}
}

func TestChapter09Round1DuplicateDiscoveredIdentity(t *testing.T) {
	dir := t.TempDir()
	for _, folder := range []string{"one", "two"} {
		path := filepath.Join(dir, folder, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("---\nname: base\ntype: primary\n---\nbody"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	_, err := app.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "claude-sonnet-4-6", BaseDir: t.TempDir(), DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}})
	if err == nil || !strings.Contains(err.Error(), "duplicate skill") {
		t.Fatalf("duplicate discovered identity: %v", err)
	}
}
