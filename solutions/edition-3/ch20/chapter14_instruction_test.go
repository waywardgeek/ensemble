package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
	"ensemble/internal/common"
	engineowner "ensemble/internal/engine"
	historyowner "ensemble/internal/history"
)

type c14InstructionWire struct {
	System     string          `json:"-"`
	SystemWire json.RawMessage `json:"system"`
	Tools      json.RawMessage `json:"tools"`
	Messages   []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"messages"`
}

func c14InstructionRequest(t *testing.T, data []byte, loaded bool) c14InstructionWire {
	t.Helper()
	var wire c14InstructionWire
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	// Chapter 15 may legally carry SYSTEM as marked text blocks.
	if json.Unmarshal(wire.SystemWire, &wire.System) != nil {
		var blocks []struct{ Type, Text string }
		if err := json.Unmarshal(wire.SystemWire, &blocks); err != nil {
			t.Fatal(err)
		}
		for _, block := range blocks {
			if block.Type != "text" {
				t.Fatal("non-text SYSTEM block")
			}
			wire.System += block.Text
		}
	}
	dialogue, _ := json.Marshal(wire.Messages)
	for _, body := range []string{"PREAMBLE-R4", "PRIMARY-R4-MANUAL", "RUNTIME-R4-MANUAL"} {
		want := 1
		if body == "RUNTIME-R4-MANUAL" && !loaded {
			want = 0
		}
		if got := strings.Count(wire.System, body); got != want {
			t.Errorf("SYSTEM %s count=%d want %d", body, got, want)
		}
		if strings.Contains(string(dialogue), body) {
			t.Errorf("instruction %s appears in dialogue", body)
		}
	}
	return wire
}
func c14DeclarationGroups(t *testing.T, wire c14InstructionWire, expected []string) {
	t.Helper()
	var groups []string
	for _, message := range wire.Messages {
		var blocks []struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(message.Content, &blocks) != nil {
			continue
		}
		for _, block := range blocks {
			if block.Type == "tool_addition" || block.Type == "tool_removal" {
				if message.Role != "system" {
					t.Error("declaration delta left its synthetic SYSTEM lane")
				}
				groups = append(groups, block.Type)
			}
		}
	}
	if strings.Join(groups, ",") != strings.Join(expected, ",") {
		t.Errorf("declaration groups=%v want %v", groups, expected)
	}
}
func c14CompleteLoadBatch(t *testing.T, wire c14InstructionWire) {
	t.Helper()
	callIndex, resultIndex, deltaIndex := -1, -1, -1
	calls, results := map[string]int{}, map[string]int{}
	for i, message := range wire.Messages {
		var blocks []struct {
			Type, ID  string
			ToolUseID string `json:"tool_use_id"`
		}
		if json.Unmarshal(message.Content, &blocks) != nil {
			continue
		}
		for _, block := range blocks {
			switch block.Type {
			case "tool_use":
				if message.Role != "assistant" {
					t.Error("call lost assistant provenance")
				}
				calls[block.ID]++
				callIndex = i
			case "tool_result":
				if message.Role != "user" {
					t.Error("result lost user role")
				}
				results[block.ToolUseID]++
				resultIndex = i
			case "tool_addition":
				deltaIndex = i
			}
		}
	}
	if len(calls) != 2 || len(results) != 2 || calls["r4-load"] != 1 || calls["r4-read"] != 1 || results["r4-load"] != 1 || results["r4-read"] != 1 || resultIndex != callIndex+1 {
		t.Fatalf("incomplete/split batch: calls=%v results=%v indexes=%d/%d", calls, results, callIndex, resultIndex)
	}
	if deltaIndex >= 0 && deltaIndex <= resultIndex {
		t.Fatal("declaration delta split the complete call/result batch")
	}
}

func TestChapter14InstructionSystemLifecycle(t *testing.T) {
	for _, model := range []string{"claude-sonnet-5-course", "claude-opus-5-course", "claude-sonnet-4-6"} {
		t.Run(model, func(t *testing.T) {
			dir, base := t.TempDir(), t.TempDir()
			c9Skill(t, dir, "base", "primary", "tools: read_file\nloadable-skills: extra\n", "PRIMARY-R4-MANUAL")
			c9Skill(t, dir, "extra", "loadable", "tools: write_file\n", "RUNTIME-R4-MANUAL")
			path := filepath.Join(base, "input.txt")
			if err := os.WriteFile(path, []byte("FILE-R4-CONTENT"), 0600); err != nil {
				t.Fatal(err)
			}
			captured := make(chan []byte, 8)
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					return
				}
				captured <- body
				parts := []any{map[string]string{"type": "text", "text": "visible-r4-answer"}}
				if requests.Add(1) == 1 {
					parts = []any{ch5Call("r4-load", "load_skill", map[string]string{"name": "extra"}), ch5Call("r4-read", "read_file", map[string]string{"path": path})}
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"model": model, "content": parts, "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
			}))
			defer server.Close()
			cfg := ensemble.Config{Vendor: ensemble.Anthropic, Surface: ensemble.Messages, Model: model, APIKey: "local-fixture-only", BaseURL: server.URL, BaseDir: base, DataDir: t.TempDir(), SavePath: filepath.Join(t.TempDir(), "session.save"), SystemPrompt: "PREAMBLE-R4", Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}}
			app := ensemble.New(nil)
			closed := false
			defer func() {
				if !closed {
					_ = app.Shutdown()
				}
			}()
			a, err := app.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			ask := func(agent ensemble.Agent, prompt string) {
				t.Helper()
				if _, err := agent.Ask(ctx, prompt); err != nil {
					t.Fatal(err)
				}
			}
			next := func() []byte {
				t.Helper()
				select {
				case b := <-captured:
					return b
				case <-ctx.Done():
					t.Fatal(ctx.Err())
					return nil
				}
			}
			ask(a, "load extra and read the local file")
			first := c14InstructionRequest(t, next(), false)
			loadedBody := next()
			loaded := c14InstructionRequest(t, loadedBody, true)
			c14CompleteLoadBatch(t, loaded)
			if !c14Has(a, "write_file") {
				t.Fatal("load did not establish its grant")
			}
			inline := model == "claude-opus-5-course"
			if inline {
				if !bytes.Equal(first.Tools, loaded.Tools) {
					t.Fatal("runtime load rewrote startup declarations")
				}
				c14DeclarationGroups(t, loaded, []string{"tool_addition"})
			} else {
				c14DeclarationGroups(t, loaded, nil)
			}
			if model == "claude-sonnet-4-6" {
				// The real paired native payload must contain only references,
				// never the canonical status prose or the skill manual.
				c9CheckWire(t, loadedBody, true, "r4-load")
			}
			if err := a.Log().Append(common.Event{Type: common.Redacted, Redact: &common.RedactData{From: 1, To: a.Log().LastSeq(), Level: common.RedactTool}}); err != nil {
				t.Fatal(err)
			}
			ask(a, "continue after clearing tool parts")
			clearedBody := next()
			c14InstructionRequest(t, clearedBody, true)
			if strings.Contains(string(clearedBody), `"tool_use"`) || strings.Contains(string(clearedBody), `"tool_result"`) {
				t.Fatal("clear retained a tool exchange")
			}
			if model == "claude-sonnet-4-6" {
				c9CheckWire(t, clearedBody, false, "")
			}
			if err := a.Skills().UnloadSkill("extra"); err != nil {
				t.Fatal(err)
			}
			if c14Has(a, "write_file") {
				t.Fatal("pending unload retained the grant")
			}
			ask(a, "continue with pending instructions")
			pending := c14InstructionRequest(t, next(), true)
			if inline {
				if !bytes.Equal(first.Tools, pending.Tools) {
					t.Fatal("unload rewrote startup declarations")
				}
				c14DeclarationGroups(t, pending, []string{"tool_addition", "tool_removal"})
			} else {
				c14DeclarationGroups(t, pending, nil)
			}
			if _, failed := a.Tools().Dispatch(ctx, "micro_handoff", json.RawMessage(`{"text":"R4-CHECKPOINT"}`)); failed {
				t.Fatal("handoff failed")
			}
			ask(a, "continue after handoff")
			handoff := c14InstructionRequest(t, next(), true)
			dialogue, _ := json.Marshal(handoff.Messages)
			if strings.Contains(string(dialogue), `"tool_use"`) || strings.Contains(string(dialogue), `"tool_result"`) || strings.Count(string(dialogue), "R4-CHECKPOINT") != 1 {
				t.Fatal("handoff changed tool/summary shape")
			}
			if inline {
				c14DeclarationGroups(t, handoff, []string{"tool_addition", "tool_removal"})
			}
			survivor := false
			for _, s := range a.Log().Context().ActiveSkills {
				if s.Name == "extra" {
					survivor = s.PendingUnload && s.Body == "RUNTIME-R4-MANUAL"
				}
			}
			if !survivor {
				t.Fatal("lost pending instruction survivor")
			}
			if err := a.Save(); err != nil {
				t.Fatal(err)
			}
			var canonical bytes.Buffer
			if err := a.Log().Dump(&canonical); err != nil {
				t.Fatal(err)
			}
			replay := historyowner.New(nil)
			if err := replay.Load(bytes.NewReader(canonical.Bytes())); err != nil {
				t.Fatal(err)
			}
			renderer, err := engineowner.New(a)
			if err != nil {
				t.Fatal(err)
			}
			current, err := a.Render()
			if err != nil {
				t.Fatal(err)
			}
			replayed, err := renderer.Render(replay.Context())
			if err != nil {
				t.Fatal(err)
			}
			c14InstructionRequest(t, replayed, true)
			if !bytes.Equal(current, replayed) {
				t.Fatal("replay changed request projection")
			}
			var afterRender bytes.Buffer
			if err := a.Log().Dump(&afterRender); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(canonical.Bytes(), afterRender.Bytes()) {
				t.Fatal("render changed canonical History")
			}
			if err := app.Shutdown(); err != nil {
				t.Fatal(err)
			}
			closed = true
			restoredRoot := ensemble.New(nil)
			defer restoredRoot.Shutdown()
			restored, err := restoredRoot.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			restoredBody, err := restored.Render()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(current, restoredBody) {
				t.Fatal("snapshot/journal restart changed projection")
			}
			if c14Has(restored, "write_file") {
				t.Fatal("restart restored the withdrawn grant")
			}
			ask(restored, "continue after restart")
			c14InstructionRequest(t, next(), true)
			if requests.Load() != 6 {
				t.Fatalf("requests=%d want 6", requests.Load())
			}
		})
	}
}
