package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// TestCompactionConstruction exercises the default and trusted custom factory
// through real child construction, including pre-startup capability isolation.
func TestCompactionConstruction(t *testing.T) {
	for _, custom := range []bool{false, true} {
		for _, builtins := range []bool{false, true} {
			name := "default"
			if custom {
				name = "custom"
			}
			if builtins {
				name += "-builtins"
			}
			t.Run(name, func(t *testing.T) {
				app := ensemble.New(io.Discard)
				calls := 0
				if custom {
					app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
						calls++
						if spec.SkillsDir != "" || spec.PrimarySkill != "" || len(spec.StartupSkills) != 0 || len(spec.Variables) != 0 {
							t.Error("factory received unnormalized startup policy")
						}
						return app.NewChildAgent(parent, parent.Config(), spec)
					})
					app.EnableFleet(nil) // Must retain that custom constructor.
				}
				parent, err := app.NewAgent(ensemble.Config{Model: "claude-ch16-course", DataDir: t.TempDir(), BuiltinTools: builtins, SystemPrompt: "MAIN CODING ROLE: DO NOT SUBMIT"})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { parent.Shutdown() })
				before := wireText(map[string]any{"tools": parent.Declarations()})
				dir := filepath.Join(parent.Config().DataDir, "children", "compact")
				child, err := app.CreateChild(parent, common.AgentSpec{ID: "compact", Compaction: true, DataDir: dir, SkillsDir: filepath.Join(dir, "missing-skills"), PrimarySkill: "must-not-load", StartupSkills: []string{"must-not-start-mcp"}, Variables: map[string]func() string{"side-effect": func() string { t.Error("startup variable executed"); return "" }}})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { child.Shutdown() })
				if custom && calls != 1 {
					t.Fatalf("custom calls=%d", calls)
				}
				if !custom && app.FleetEnabled() {
					t.Fatal("internal factory enabled public delegation")
				}
				if !custom {
					if _, err := parent.Fleet().Spawn("ordinary", "must still refuse public delegation", ""); err == nil {
						t.Fatal("ordinary spawn admitted without public delegation")
					}
				}
				if wireText(map[string]any{"tools": parent.Declarations()}) != before {
					t.Fatal("child changed parent's tool surface")
				}
				decls := child.Declarations()
				if len(decls) != 1 || decls[0].Name != "submit" {
					t.Fatalf("worker tools=%+v", decls)
				}
				sameJSON(t, decls[0].Schema, []byte(memorySubmitSchema))
				if child.ParentAgent() != parent || child.Ensemble() != app || !child.Spec().Compaction {
					t.Fatal("worker lost actual owners/purpose")
				}
				if child.Skills() != nil || child.Recaller() != nil {
					t.Fatal("worker attached startup skills or recall")
				}
				for b := common.Band(0); b <= common.BandSession; b++ {
					if !child.Memory().Settings(b).Disabled {
						t.Fatalf("worker band %d enabled", b)
					}
				}
				prompt := child.Engine().SystemPrompt()
				if strings.Contains(prompt, "MAIN CODING ROLE") || !strings.Contains(prompt, "source owner's first-person memory") || !strings.Contains(prompt, "why") {
					t.Fatalf("wrong worker role: %s", prompt)
				}
				if err := child.SetMemorySettings(ensemble.MemorySettings{}); err == nil {
					t.Fatal("worker memory settings reopened")
				}
			})
		}
	}
}

// TestCompactionSafetyCeilings proves internal construction narrows through the
// same sandbox boundary and cannot place a worker outside its owner's root.
func TestCompactionSafetyCeilings(t *testing.T) {
	app := ensemble.New(io.Discard)
	dir := t.TempDir()
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-ch16-course", DataDir: dir}, ensemble.AgentSpec{SafeMode: true, SandboxRoot: dir})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	spec := common.AgentSpec{ID: "compact", Compaction: true, DataDir: filepath.Join(dir, "children", "compact"), SafeMode: true, SandboxRoot: parent.Spec().SandboxRoot}
	child, err := app.CreateChild(parent, spec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { child.Shutdown() })
	if !child.Spec().SafeMode || child.Spec().EnableWebSearch || child.Spec().SandboxRoot != parent.Spec().SandboxRoot {
		t.Fatal("compaction widened inherited safety ceilings")
	}
	spec.ID = "outside"
	spec.DataDir = t.TempDir()
	if _, err := app.CreateChild(parent, spec); err == nil {
		t.Fatal("compaction bypassed parent's filesystem ceiling")
	}
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		spec.Compaction = false
		return app.NewChildAgent(parent, parent.Config(), spec)
	})
	spec.ID = "wrong-purpose"
	spec.DataDir = filepath.Join(dir, "children", spec.ID)
	if _, err := app.CreateChild(parent, spec); err == nil || !strings.Contains(err.Error(), "purpose") {
		t.Fatalf("wrong-purpose factory was accepted: %v", err)
	}
}

// TestCompactionFiniteExecution keeps the actor, parser, jobs and submission
// real; a concurrent provider supplies only the responses at its HTTP boundary.
func TestCompactionFiniteExecution(t *testing.T) {
	for _, mode := range []string{"success", "invalid", "extra-call", "forged-tool", "prose-with-hint"} {
		t.Run(mode, func(t *testing.T) {
			var count atomic.Int32
			schemas := make(chan []byte, 8)
			entered, release := make(chan struct{}), make(chan struct{})
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := count.Add(1)
				var request map[string]any
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					http.Error(w, "bad JSON", 400)
					return
				}
				decls, _ := request["tools"].([]any)
				if len(decls) != 1 || decls[0].(map[string]any)["name"] != "submit" {
					t.Error("worker request lost its sole memory submission contract")
					http.Error(w, "wrong worker declarations", 400)
					return
				}
				schema, _ := json.Marshal(decls[0].(map[string]any)["input_schema"])
				schemas <- schema
				if mode == "prose-with-hint" && n == 1 {
					close(entered)
					select {
					case <-release:
					case <-r.Context().Done():
						return
					}
				}
				content := []any{map[string]any{"type": "text", "text": "I will preserve the decision and its reason."}}
				if mode != "prose-with-hint" {
					name, args := "submit", any(map[string]string{"memory": "I retained the decision because it avoids a repeated failure."})
					if mode == "invalid" {
						args = map[string]int{"memory": 7}
					}
					if mode == "forged-tool" {
						name = "forbidden"
						args = map[string]any{}
					}
					content = append(content, map[string]any{"type": "tool_use", "id": fmt.Sprint("call-", n), "name": name, "input": args})
					if mode == "extra-call" {
						content = append(content, map[string]any{"type": "tool_use", "id": fmt.Sprint("extra-", n), "name": "forbidden", "input": map[string]any{}})
					}
				}
				stop := "tool_use"
				if mode == "prose-with-hint" {
					stop = "end_turn"
				}
				json.NewEncoder(w).Encode(map[string]any{"content": content, "stop_reason": stop, "usage": map[string]int{"input_tokens": 5000, "output_tokens": 20}})
			}))
			t.Cleanup(peer.Close)
			app := ensemble.New(io.Discard)
			app.EnableFleet(nil)
			parent, err := app.NewAgent(ensemble.Config{Model: "claude-ch16-course", BaseURL: peer.URL, APIKey: "fake", DataDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { parent.Shutdown() })
			child, err := app.CreateChild(parent, common.AgentSpec{ID: "compact", Compaction: true, DataDir: filepath.Join(parent.Config().DataDir, "children", "compact")})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { child.Shutdown() })
			var invoked atomic.Bool
			if err := child.RegisterTool(ensemble.ToolDeclaration{Name: "forbidden", Schema: json.RawMessage(`{"type":"object"}`)}, func(ensemble.ToolContext, json.RawMessage) (string, error) {
				invoked.Store(true)
				return "forbidden", nil
			}); err == nil {
				t.Fatal("worker admitted forbidden tool registration")
			}
			if len(child.Declarations()) != 1 {
				t.Fatal("registration reopened worker capabilities")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result := child.Submit(ctx, "SOURCE DATA: We decided to retain the reason.")
			if mode == "prose-with-hint" {
				select {
				case <-entered:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				child.Post(ensemble.Hint{Text: "pending steering cannot create another task"})
				close(release)
			}
			var got ensemble.Result
			select {
			case got = <-result:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			if len(schemas) != int(count.Load()) {
				t.Fatal("provider did not capture every worker schema")
			}
			for range count.Load() {
				sameJSON(t, <-schemas, []byte(memorySubmitSchema))
			}
			if invoked.Load() {
				t.Fatal("forged worker capability ran")
			}
			if mode == "success" {
				if got.Err != nil || !strings.Contains(got.Text, `"memory"`) || count.Load() != 1 {
					t.Fatalf("successful worker: %+v requests=%d", got, count.Load())
				}
			} else {
				want := int32(3)
				if mode == "prose-with-hint" {
					want = 1
				}
				if got.Err == nil || count.Load() != want || child.Fleet().Status().State == "submitted" {
					t.Fatalf("failed worker: %+v requests=%d", got, count.Load())
				}
			}
			before := count.Load()
			if _, err := child.Ask(ctx, "a second task must not reset the budget"); err == nil || count.Load() != before {
				t.Fatalf("second task admitted: %v requests=%d", err, count.Load())
			}
		})
	}
}

// TestCompactionPurposeNotInherited checks trusted Go spawning separately from
// the worker's forbidden delegation tool: ordinary descendants retain defaults.
func TestCompactionPurposeNotInherited(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"content":[{"type":"tool_use","id":"result","name":"submit","input":{"result":"ordinary child result"}}],"stop_reason":"tool_use","usage":{"input_tokens":1,"output_tokens":1}}`)
	}))
	t.Cleanup(peer.Close)
	app := ensemble.New(io.Discard)
	app.EnableFleet(nil)
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-ch16-course", BaseURL: peer.URL, APIKey: "fake", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	worker, err := app.CreateChild(parent, common.AgentSpec{ID: "compact", Compaction: true, DataDir: filepath.Join(parent.Config().DataDir, "children", "compact")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { worker.Shutdown() })
	derived := ensemble.AgentSpec(worker.Spec()).Child(filepath.Join(worker.Config().DataDir, "derived"))
	if derived.Compaction || len(derived.SubmitSchema) != 0 {
		t.Fatal("public Child helper inherited worker purpose/schema")
	}
	derivedChild, err := app.NewChildAgent(worker, worker.Config(), derived)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { derivedChild.Shutdown() })
	if derivedChild.Spec().Compaction || len(derivedChild.Spec().SubmitSchema) != 0 {
		t.Fatal("public Child helper constructed another managed worker")
	}
	if _, err := worker.Fleet().Spawn("ordinary", "finish ordinary task", ""); err != nil {
		t.Fatal(err)
	}
	child, err := worker.Fleet().Find("ordinary")
	if err != nil {
		t.Fatal(err)
	}
	if child.Spec().Compaction || len(child.Spec().SubmitSchema) != 0 {
		t.Fatal("ordinary descendant inherited worker purpose/schema")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := worker.Fleet().Join(ctx, []string{"ordinary"}); err != nil {
		t.Fatal(err)
	}
	if result, err := worker.Fleet().Submission("ordinary"); err != nil || result != "ordinary child result" {
		t.Fatalf("ordinary result=%q err=%v", result, err)
	}
}
