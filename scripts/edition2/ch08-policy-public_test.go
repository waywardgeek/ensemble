package ensemble_test

// This external consumer uses only the declared public API and JSON contracts.
// It intentionally has no optional GUI import or internal package access.
import (
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
	"sync"
	"testing"
	"time"

	"example.com/ensemble"
)

func ch08JSON(t *testing.T, value any) map[string]any {
	t.Helper()
	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err = json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func ch08Policy(t *testing.T, value any, revision, stored int, persistent bool) {
	t.Helper()
	effective := stored
	if effective == 0 {
		effective = 16
	}
	want := map[string]any{"revision": float64(revision), "persistent": persistent,
		"max_model_requests": float64(stored), "effective_max_model_requests": float64(effective)}
	if got := ch08JSON(t, value); !reflect.DeepEqual(got, want) {
		t.Fatalf("unsafe or incorrect policy: got %v, want %v", got, want)
	}
}

func ch08Agent(t *testing.T, app *ensemble.Ensemble, directory, name, policy string) *ensemble.Agent {
	t.Helper()
	a, err := app.NewAgent(ensemble.Config{Vendor: "openai", Model: "fixture-ch08-public", APIKey: "LOCAL-ONLY",
		Workspace: directory, LogPath: filepath.Join(directory, name+".log"), PolicyPath: policy})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a
}

func TestCh08PublicMemoryOwnershipAndWatch(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	dir := t.TempDir()
	a := ch08Agent(t, app, dir, "a", "")
	b := ch08Agent(t, app, dir, "b", "")
	ch08Policy(t, a.ExecutionPolicy(), 0, 0, false)
	_, watch, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer watch.Close()
	ack, err := a.UpdatePolicy(0, json.RawMessage(`{"max_model_requests":2}`))
	if err != nil {
		t.Fatal(err)
	}
	ch08Policy(t, a.ExecutionPolicy(), 1, 2, false)
	ch08Policy(t, b.ExecutionPolicy(), 0, 0, false)
	// Mutate the returned public value without depending on Go field spelling.
	owned := a.ExecutionPolicy()
	if err = json.Unmarshal([]byte(`{"revision":999,"persistent":true,"max_model_requests":200,"effective_max_model_requests":200}`), &owned); err != nil {
		t.Fatal(err)
	}
	ch08Policy(t, a.ExecutionPolicy(), 1, 2, false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	record, err := watch.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got := ch08JSON(t, record)
	if got["revision"] != ch08JSON(t, ack)["watch_revision"] {
		t.Fatal("public acknowledgement disagrees with published watch revision")
	}
	observation := got["observation"].(map[string]any)
	if observation["kind"] != "policy_changed" {
		t.Fatal("policy change bypassed public watch")
	}
	ch08Policy(t, observation["execution_policy"], 1, 2, false)
	snapshot, tail, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()
	safe := ch08JSON(t, snapshot)
	if safe["watermark"] != got["revision"] {
		t.Fatal("watch snapshot regressed after policy acknowledgement")
	}
	ch08Policy(t, safe["state"].(map[string]any)["execution_policy"], 1, 2, false)
	if _, err = a.UpdatePolicy(0, json.RawMessage(`{"max_model_requests":3}`)); err == nil {
		t.Fatal("stale public patch accepted")
	}
	if _, err = a.UpdatePolicy(1, json.RawMessage(`{"max_model_requests":0,"max_model_requests":4}`)); err == nil {
		t.Fatal("duplicate public patch accepted")
	}
	ch08Policy(t, a.ExecutionPolicy(), 1, 2, false)
	if _, err = a.UpdatePolicy(1, json.RawMessage(`{"max_model_requests":0}`)); err != nil {
		t.Fatal(err)
	}
	ch08Policy(t, a.ExecutionPolicy(), 2, 0, false)
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = a.UpdatePolicy(2, json.RawMessage(`{"max_model_requests":3}`)); err == nil {
		t.Fatal("closed public owner accepted update")
	}
}

func TestCh08PublicPersistenceIdentity(t *testing.T) {
	app := ensemble.New(io.Discard)
	defer app.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "absent-parent", "POLICY-PATH-MARKER.json")
	a := ch08Agent(t, app, dir, "first", path)
	ch08Policy(t, a.ExecutionPolicy(), 0, 0, true)
	if _, err := a.UpdatePolicy(0, json.RawMessage(`{"max_model_requests":17}`)); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.UpdatePolicy(1, json.RawMessage(`{"max_model_requests":17}`)); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	later, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) || !os.SameFile(stat, later) || stat.ModTime() != later.ModTime() {
		t.Fatal("no-change public patch performed persistence work")
	}
	changed := a.Config()
	changed.PolicyPath = filepath.Join(dir, "other.json")
	if a.SetConfig(changed) == nil {
		t.Fatal("creation-only policy path changed")
	}
	duplicate := a.Config()
	duplicate.LogPath = filepath.Join(dir, "duplicate.log")
	if other, err := app.NewAgent(duplicate); err == nil {
		_ = other.Close()
		t.Fatal("two live owners claimed same policy file")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := ch08Agent(t, app, dir, "reopened", path)
	ch08Policy(t, reopened.ExecutionPolicy(), 1, 17, true)
}

func TestCh08PublicTwoAgentEffectsAndReplay(t *testing.T) {
	var mu sync.Mutex
	bodies := map[string][]map[string]any{}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			http.Error(w, "bad fixture request", 500)
			return
		}
		label := ""
		for _, item := range body["messages"].([]any) {
			message := item.(map[string]any)
			if message["role"] == "user" {
				switch text := message["content"].(type) {
				case string:
					label = text
				case []any:
					label = ""
					for _, part := range text {
						if p, ok := part.(map[string]any); ok {
							if s, ok := p["text"].(string); ok {
								label += s
							}
						}
					}
				}
			}
		}
		if label != "A" && label != "B" {
			t.Errorf("unexpected fixture label %q", label)
			http.Error(w, "bad fixture label", 500)
			return
		}
		mu.Lock()
		bodies[label] = append(bodies[label], body)
		number := len(bodies[label])
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"model": "fixture-ch08-public", "choices": []any{map[string]any{"finish_reason": "tool_calls", "message": map[string]any{"role": "assistant", "content": nil, "tool_calls": []any{map[string]any{"id": fmt.Sprintf("%s-%d", label, number), "type": "function", "function": map[string]any{"name": "read_file", "arguments": `{"path":"seed.txt"}`}}}}}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
	}))
	defer backend.Close()
	app := ensemble.New(io.Discard)
	defer app.Close()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("PUBLIC-EFFECT"), 0600); err != nil {
		t.Fatal(err)
	}
	agents := map[string]*ensemble.Agent{}
	for i, label := range []string{"A", "B"} {
		config := ensemble.Config{Vendor: "openai", Model: "fixture-ch08-public", APIKey: "LOCAL-ONLY", BaseURL: backend.URL, DisableStreaming: true, Workspace: dir, LogPath: filepath.Join(dir, label+".log"), PolicyPath: filepath.Join(dir, label+".json"), Builtins: []string{"read_file"}}
		a, err := app.NewAgent(config)
		if err != nil {
			t.Fatal(err)
		}
		agents[label] = a
		if _, err = a.UpdatePolicy(0, json.RawMessage(fmt.Sprintf(`{"max_model_requests":%d}`, i+1))); err != nil {
			t.Fatal(err)
		}
	}
	handles := map[string]ensemble.RequestHandle{}
	for label, a := range agents {
		h, err := a.Submit(label)
		if err != nil {
			t.Fatal(err)
		}
		handles[label] = h
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, label := range []string{"A", "B"} {
		completed, err := handles[label].Wait(ctx)
		if err != nil || completed.Outcome != "round_limit" {
			t.Fatalf("%s public completion: %v %v", label, completed, err)
		}
	}
	for i, label := range []string{"A", "B"} {
		a := agents[label]
		mu.Lock()
		sent := append([]map[string]any(nil), bodies[label]...)
		mu.Unlock()
		if len(sent) != i+1 {
			t.Fatalf("%s policy did not control real HTTP: %d", label, len(sent))
		}
		requests, returns := 0, 0
		for _, event := range a.Events() {
			if event.Type == "tool_returned" {
				returns++
				if !strings.Contains(stringMustJSON(t, event), "PUBLIC-EFFECT") {
					t.Fatal("tool result lacks actual file contents")
				}
			}
			if event.Type == "request_sent" {
				body, err := a.ReconstructRequest(event.Seq)
				if err != nil {
					t.Fatal(err)
				}
				var replay map[string]any
				if err = json.Unmarshal(body, &replay); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(replay, sent[requests]) {
					t.Fatal("policy capture changed exact request reconstruction")
				}
				requests++
			}
		}
		if requests != i+1 || returns != i+1 {
			t.Fatal("final accepted tool not paired before round limit")
		}
		if _, err := a.UpdatePolicy(1, json.RawMessage(`{"max_model_requests":5}`)); err != nil {
			t.Fatal(err)
		}
		config := a.Config()
		if err := a.Close(); err != nil {
			t.Fatal(err)
		}
		before, err := os.ReadFile(config.PolicyPath)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := app.Load(config.LogPath, config)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = loaded.ReconstructRequest(firstRequestSeq(t, loaded)); err != nil {
			t.Fatal(err)
		}
		_ = loaded.Close()
		after, err := os.ReadFile(config.PolicyPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(before) != string(after) {
			t.Fatal("historical capture rewrote current policy file")
		}
	}
}

func stringMustJSON(t *testing.T, value any) string {
	t.Helper()
	b, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func firstRequestSeq(t *testing.T, a *ensemble.Agent) uint64 {
	t.Helper()
	for _, e := range a.Events() {
		if e.Type == "request_sent" {
			return e.Seq
		}
	}
	t.Fatal("missing request")
	return 0
}

func TestCh08PublicRecordedPolicyValidation(t *testing.T) {
	cases := []struct {
		name, policy string
		valid        bool
	}{
		{"historical-absent", "", true},
		{"default", `{"revision":0,"max_model_requests":0,"effective_max_model_requests":16}`, true},
		{"positive", `{"revision":9,"max_model_requests":17,"effective_max_model_requests":17}`, true},
		{"present-null", `null`, false},
		{"present-empty", `{}`, false},
		{"missing-revision", `{"max_model_requests":0,"effective_max_model_requests":16}`, false},
		{"missing-stored", `{"revision":0,"effective_max_model_requests":16}`, false},
		{"missing-effective", `{"revision":0,"max_model_requests":0}`, false},
		{"negative-revision", `{"revision":-1,"max_model_requests":0,"effective_max_model_requests":16}`, false},
		{"fractional-revision", `{"revision":0.5,"max_model_requests":0,"effective_max_model_requests":16}`, false},
		{"wrong-kind", `{"revision":0,"max_model_requests":false,"effective_max_model_requests":16}`, false},
		{"null-stored", `{"revision":0,"max_model_requests":null,"effective_max_model_requests":16}`, false},
		{"over-range", `{"revision":0,"max_model_requests":257,"effective_max_model_requests":257}`, false},
		{"wrong-default", `{"revision":0,"max_model_requests":0,"effective_max_model_requests":1}`, false},
		{"wrong-positive", `{"revision":0,"max_model_requests":17,"effective_max_model_requests":16}`, false},
	}
	positives := 0
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.valid && positives != 3 {
				t.Fatal("negative capture control requires all three passing append/load fixtures")
			}
			member := ""
			if tc.policy != "" {
				member = `,"policy":` + tc.policy
			}
			raw := `{"seq":1,"type":"turn_started","time":"2026-01-01T00:00:00Z","turn":{"request_id":"historical"` + member + `}}`
			app := ensemble.New(io.Discard)
			defer app.Close()
			dir := t.TempDir()
			live := ch08Agent(t, app, dir, "live", "")
			var event ensemble.Event
			err := json.Unmarshal([]byte(raw), &event)
			if err == nil {
				err = live.Append(event)
			}
			if (err == nil) != tc.valid {
				t.Fatalf("append validity=%v, want %v: %v", err == nil, tc.valid, err)
			}
			if !tc.valid && len(live.Events()) != 0 {
				t.Fatal("malformed present capture mutated history")
			}
			path := filepath.Join(dir, "history.log")
			log := "{\"log_version\":1}\n" + raw + "\n"
			if err = os.WriteFile(path, []byte(log), 0600); err != nil {
				t.Fatal(err)
			}
			config := live.Config()
			config.LogPath = path
			loaded, err := app.Load(path, config)
			if (err == nil) != tc.valid {
				t.Fatalf("load validity=%v, want %v: %v", err == nil, tc.valid, err)
			}
			if loaded != nil {
				_ = loaded.Close()
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != log {
				t.Fatal("load rewrote historical policy capture")
			}
			if tc.valid {
				positives++
			}
		})
	}
}
