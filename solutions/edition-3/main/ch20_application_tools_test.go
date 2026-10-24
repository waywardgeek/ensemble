package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"ensemble"
)

func TestCh20ApplicationSubmitOwnershipAndRecovery(t *testing.T) {
	for _, source := range []string{"application", "application-dialogue", "compiled:fixture"} {
		t.Run(source, func(t *testing.T) {
			var requests, effects, tailEffects atomic.Int32
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requests.Add(1) == 1 {
					w.Header().Set("Content-Type", "application/json")
					io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[{"id":"application-submit","type":"function","function":{"name":"submit","arguments":"{\"composer_text\":\"hello application\"}"}},{"id":"after-submit","type":"function","function":{"name":"after_submit","arguments":"{}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)
				} else {
					ch20Reply(w, "application turn finished", false)
				}
			}))
			defer peer.Close()
			root := ensemble.New(log.New(io.Discard, "", 0))
			defer root.Shutdown()
			dir := t.TempDir()
			cfg := ensemble.Config{ID: "application", Model: "fixture", Vendor: ensemble.OpenAI, APIKey: "dummy", BaseURL: peer.URL,
				BaseDir: dir, DataDir: filepath.Join(dir, "state"), Tools: []string{}, LogMode: ensemble.LogFull,
				SubmitSchema: json.RawMessage(`{"type":"object","properties":{"fleet_only":{"type":"integer"}},"required":["fleet_only"],"additionalProperties":false}`)}
			a, err := root.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if len(a.Tools().Definitions()) != 0 {
				t.Fatal("application-only Agent gained startup tools")
			}
			definition := ensemble.ToolDefinition{Name: "submit", Description: "Send an application composer", Parameters: json.RawMessage(`{"type":"object","properties":{"composer_text":{"type":"string"}},"required":["composer_text"],"additionalProperties":false}`)}
			handler := func(ctx context.Context, out ensemble.ToolCall, raw json.RawMessage) error {
				var args map[string]string
				if err := json.Unmarshal(raw, &args); err != nil {
					return err
				}
				if len(args) != 1 || args["composer_text"] != "hello application" {
					return errors.New("wrong application schema/arguments")
				}
				effects.Add(1)
				// Text emitted by an application cannot assert the canonical fact.
				_, err := io.WriteString(out, `{"sent":true,"fleet_completion":true}`)
				return err
			}
			if strings.HasPrefix(source, "application") {
				err = a.RegisterExternalTool(definition, handler)
			} else {
				err = a.Skills().RegisterTool(ensemble.ToolSpec{ToolDefinition: definition, Handler: handler, Source: source, Enabled: true})
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := a.RegisterExternalTool(ensemble.ToolDefinition{Name: "after_submit", Description: "Ordinary later call in the same batch", Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}, func(context.Context, ensemble.ToolCall, json.RawMessage) error {
				tailEffects.Add(1)
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			ctx := ch20Context(t)
			for _, prompt := range []string{"use application submit", "continue ordinary conversation"} {
				answer, err := a.Ask(ctx, prompt)
				if err != nil || answer != "application turn finished" {
					t.Fatalf("application turn: %q %v", answer, err)
				}
			}
			if effects.Load() != 1 || tailEffects.Load() != 1 || requests.Load() != 3 {
				t.Fatalf("handler=%d tail=%d provider=%d", effects.Load(), tailEffects.Load(), requests.Load())
			}
			if outcome, err := root.Outcome(a.ID()); err == nil {
				t.Fatalf("application submit created a terminal outcome: %+v", outcome)
			}
			found := false
			for _, e := range a.Log().Events() {
				if e.Type == ensemble.ToolReturned && e.Tool != nil && e.Tool.Name == "submit" {
					found = true
					if e.Tool.IsError || e.Tool.FleetCompletion || e.Tool.Job == nil {
						t.Fatalf("application result lost ordinary ownership: %+v", e.Tool)
					}
				}
			}
			if !found {
				t.Fatal("application submit result not recorded")
			}
			if a.Log().Context().FleetCompletion {
				t.Fatal("application output forged accepted completion context")
			}
			if source == "application-dialogue" {
				if err := a.SetPersistencePolicy(ensemble.LogDialogueOnly, *a.Config().DiagnosticMaxRecords); err != nil {
					t.Fatal(err)
				}
			}
			cfg = a.Config()
			if err := a.Save(); err != nil {
				t.Fatal(err)
			}
			if err := root.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(cfg.DataDir, "submit.md")); !os.IsNotExist(err) {
				t.Fatalf("application wrote a fleet submission row: %v", err)
			}

			// Historical application actions must not be reinterpreted by a later
			// catalogue, even when it now installs the genuine fleet submit.
			for _, tools := range [][]string{{}, nil} {
				cfg.Tools = tools
				reopened := ensemble.New(log.New(io.Discard, "", 0))
				t.Cleanup(func() { _ = reopened.Shutdown() })
				again, err := reopened.NewAgent(cfg)
				if err != nil {
					t.Fatalf("reopening application history: %v", err)
				}
				if answer, err := again.Ask(ctx, "ordinary reopened turn"); err != nil || answer != "application turn finished" {
					t.Fatalf("reopened turn %q %v", answer, err)
				}
				if _, err := reopened.Outcome(again.ID()); err == nil {
					t.Fatal("reopened application history became fleet completion")
				}
				if err := reopened.Shutdown(); err != nil {
					t.Fatal(err)
				}
			}
			if effects.Load() != 1 {
				t.Fatal("recovery replayed the application handler")
			}
			if _, err := os.Stat(filepath.Join(cfg.DataDir, "submit.md")); !os.IsNotExist(err) {
				t.Fatalf("recovery invented a submission row: %v", err)
			}
		})
	}
}

func TestCh20ApplicationCatalogueBoundaries(t *testing.T) {
	makeAgent := func(tools []string, skill *ensemble.SkillSpec) ensemble.Agent {
		t.Helper()
		root := ensemble.New(log.New(io.Discard, "", 0))
		t.Cleanup(func() { _ = root.Shutdown() })
		dir := t.TempDir()
		a, err := root.NewAgent(ensemble.Config{Model: "fixture", BaseDir: dir, DataDir: filepath.Join(dir, "state"), Tools: tools, Skills: skill})
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	noop := func(context.Context, ensemble.ToolCall, json.RawMessage) error { return nil }
	def := func(name string) ensemble.ToolDefinition {
		return ensemble.ToolDefinition{Name: name, Description: name, Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}
	}
	a := makeAgent([]string{}, nil)
	if _, failed := a.Tools().Dispatch(ch20Context(t), "submit", json.RawMessage(`{"result":"guessed completion"}`)); !failed {
		t.Fatal("unregistered application-only submit became a fleet operation")
	}
	for _, name := range []string{"hear", "type", "submit", "wait", "sleep"} {
		if err := a.RegisterExternalTool(def(name), noop); err != nil {
			t.Fatal(err)
		}
	}
	names := []string{}
	for _, d := range a.Tools().Definitions() {
		names = append(names, d.Name)
	}
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"hear", "sleep", "submit", "type", "wait"}) {
		t.Fatalf("listener declarations widened: %v", names)
	}
	if err := a.RegisterExternalTool(def("submit"), noop); err == nil {
		t.Fatal("duplicate application handler replaced")
	}
	if err := a.RegisterExternalTool(def("run_command"), noop); err == nil {
		t.Fatal("disabled installed core tool replaced")
	}
	if err := a.Skills().RegisterTool(ensemble.ToolSpec{ToolDefinition: def("compiled_owned"), Source: "compiled:fixture", Handler: noop}); err != nil {
		t.Fatal(err)
	}
	if err := a.Skills().RegisterExternal([]ensemble.ToolSpec{{ToolDefinition: def("remote_owned"), Source: "mcp:fixture", Handler: noop}}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"compiled_owned", "remote_owned"} {
		if err := a.RegisterExternalTool(def(name), noop); err == nil {
			t.Fatalf("replaced installed owner %s", name)
		}
	}
	ctx := ch20Context(t)
	for _, name := range []string{"run_command", "read_file", "spawn_sub_agent", "join_agents", "agents_status", "shutdown_agent", "send_message_to_parent", "respond_to_parent_message", "compiled_owned", "remote_owned"} {
		if _, failed := a.Tools().Dispatch(ctx, name, json.RawMessage(`{}`)); !failed {
			t.Fatalf("guessed capability %s executed", name)
		}
	}
	for _, tools := range [][]string{nil, {"read_file"}} {
		ordinary := makeAgent(tools, nil)
		if err := ordinary.RegisterExternalTool(def("submit"), noop); err == nil {
			t.Fatal("default/subset fleet catalogue lost collision protection")
		}
	}
	directory := t.TempDir()
	if err := os.MkdirAll(filepath.Join(directory, "coordinator"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "coordinator", "SKILL.md"), []byte("---\nname: coordinator\ntype: primary\ndescription: Fleet coordinator\ntools: [submit]\n---\nUse the declared terminal tool.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	skilled := makeAgent(nil, &ensemble.SkillSpec{Directory: directory, Primary: "coordinator"})
	if err := skilled.RegisterExternalTool(def("submit"), noop); err == nil {
		t.Fatal("skill-selected fleet ownership was shadowed")
	}
	found := false
	for _, definition := range skilled.Tools().Definitions() {
		if definition.Name == "submit" {
			found = true
		}
	}
	if !found {
		t.Fatal("skill-selected fleet submit unavailable")
	}
}

func TestCh20AcceptedCompletionFactRecovery(t *testing.T) {
	for _, mode := range []string{"fleet-tool", "Complete", "dialogue-only/fleet-tool", "dialogue-only/Complete"} {
		t.Run(mode, func(t *testing.T) {
			root, a := ch20Rig(t, func(w http.ResponseWriter, b []byte) { ch20Reply(w, `{"result":"done"}`, true) })
			ctx := ch20Context(t)
			dialogueOnly := strings.HasPrefix(mode, "dialogue-only/")
			if dialogueOnly {
				if err := a.SetPersistencePolicy(ensemble.LogDialogueOnly, *a.Config().DiagnosticMaxRecords); err != nil {
					t.Fatal(err)
				}
			}
			var result ensemble.Result
			if strings.HasSuffix(mode, "Complete") {
				result = <-a.Complete(ctx, "done")
			} else {
				result = <-a.Submit(ctx, "finish")
			}
			if result.Err != nil || string(result.Submission) != `{"result":"done"}` {
				t.Fatalf("terminal result %+v", result)
			}
			if _, err := root.Wait(ctx, a.ID()); err != nil {
				t.Fatal(err)
			}
			marked := 0
			for _, e := range a.Log().Events() {
				if e.Type == ensemble.ToolReturned && e.Tool != nil && e.Tool.FleetCompletion {
					marked++
				}
			}
			if !dialogueOnly && marked != 1 {
				t.Fatalf("accepted completion facts=%d", marked)
			}
			if !a.Log().Context().FleetCompletion {
				t.Fatal("accepted completion missing from projected context")
			}
			cfg := a.Config()
			if err := root.Shutdown(); err != nil {
				t.Fatal(err)
			}
			restored := ensemble.New(log.New(io.Discard, "", 0))
			again, err := restored.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if out, err := restored.Wait(ctx, again.ID()); err != nil || out.State != "submitted" {
				t.Fatalf("durable recovery %+v %v", out, err)
			}
			if r := <-again.Submit(ctx, "must remain terminal"); r.Err == nil {
				t.Fatal("recovered completion accepted new work")
			}
			if err := restored.Shutdown(); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filepath.Join(cfg.DataDir, "submit.md")); err != nil {
				t.Fatal(err)
			}
			missing := ensemble.New(log.New(io.Discard, "", 0))
			defer missing.Shutdown()
			if _, err := missing.NewAgent(cfg); err == nil || !strings.Contains(err.Error(), "accepted submit history has no durable submission row") {
				t.Fatalf("missing accepted row not refused: %v", err)
			}
		})
	}
}
