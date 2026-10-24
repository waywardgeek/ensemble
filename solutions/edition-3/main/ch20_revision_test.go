package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func ch20Rig(t *testing.T, respond func(http.ResponseWriter, []byte)) (*ensemble.Ensemble, ensemble.Agent) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { raw, _ := io.ReadAll(r.Body); respond(w, raw) }))
	t.Cleanup(server.Close)
	base := t.TempDir()
	root := ensemble.New(log.New(io.Discard, "", 0))
	t.Cleanup(func() { _ = root.Shutdown() })
	cfg := ensemble.Config{ID: "root", Vendor: ensemble.OpenAI, Model: "fixture", APIKey: "dummy", BaseURL: server.URL, BaseDir: base, DataDir: filepath.Join(base, "root"), SavePath: filepath.Join(base, "root", "save.json")}
	a, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	root.SetAgentFactory(func(ctx context.Context, request ensemble.Config) (ensemble.Agent, error) {
		parent, ok := root.Agent(request.ParentID)
		if !ok {
			return nil, errors.New("missing parent")
		}
		cfg := parent.Config().Child(filepath.Join(base, request.ID))
		cfg.ID, cfg.ParentID, cfg.Prompt = request.ID, request.ParentID, request.Prompt
		cfg.SubmitSchema = append(json.RawMessage(nil), request.SubmitSchema...)
		return root.NewAgent(cfg)
	})
	return root, a
}
func ch20Reply(w http.ResponseWriter, content string, submit bool) {
	message := map[string]any{"role": "assistant", "content": content}
	finish := "stop"
	if submit {
		message["tool_calls"] = []any{map[string]any{"id": "call-submit", "type": "function", "function": map[string]any{"name": "submit", "arguments": content}}}
		message["content"] = ""
		finish = "tool_calls"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"id": "fixture", "object": "chat.completion", "model": "fixture", "choices": []any{map[string]any{"index": 0, "message": message, "finish_reason": finish}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}})
}
func ch20Context(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestCh20WakesAndFailureResolution(t *testing.T) {
	var calls atomic.Int32
	root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) { calls.Add(1); ch20Reply(w, "ordinary finish", false) })
	ctx := ch20Context(t)
	_, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "worker", Prompt: "do work"})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := root.Wait(ctx, "worker")
	if err != nil || outcome.State != "failed" || !strings.Contains(outcome.Error, "three corrective wakes") {
		t.Fatalf("outcome %+v %v", outcome, err)
	}
	if calls.Load() != 4 {
		t.Fatalf("got %d provider calls, want initial + three wakes", calls.Load())
	}
	if err := root.AcceptSubmission(parent.ID(), json.RawMessage(`{"result":"done"}`)); err == nil {
		t.Fatal("unresolved failure allowed parent submit")
	}
	if err := root.ShutdownAgent("worker"); err != nil {
		t.Fatal(err)
	}
	if err := root.AcceptSubmission(parent.ID(), json.RawMessage(`{"result":"resolved"}`)); err != nil {
		t.Fatal(err)
	}
}

func TestCh20RootSubmitQueueAndRecovery(t *testing.T) {
	entered, release := make(chan struct{}, 1), make(chan struct{})
	var calls atomic.Int32
	root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) {
		calls.Add(1)
		entered <- struct{}{}
		<-release
		ch20Reply(w, `{"result":"done","ignored":true}`, true)
	})
	ctx := ch20Context(t)
	first := parent.Submit(ctx, "terminal root")
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	queued := parent.Submit(ctx, "must not run")
	close(release)
	if result := <-first; result.Err != nil || string(result.Submission) != `{"result":"done"}` {
		t.Fatalf("first %+v", result)
	}
	if result := <-queued; result.Err == nil {
		t.Fatal("queued turn ran after submit")
	}
	outcome, err := root.Wait(ctx, parent.ID())
	if err != nil || outcome.State != "submitted" {
		t.Fatalf("outcome %+v %v", outcome, err)
	}
	if calls.Load() != 1 {
		t.Fatal("provider request after terminal submit")
	}
	cfg := parent.Config()
	raw, err := os.ReadFile(filepath.Join(cfg.DataDir, "submit.md"))
	if err != nil || strings.Count(string(raw), "\n") != 1 || !strings.Contains(string(raw), `"version":1`) {
		t.Fatalf("durable row %s %v", raw, err)
	}
	if err := root.Shutdown(); err != nil {
		t.Fatal(err)
	}
	restored := ensemble.New(log.New(io.Discard, "", 0))
	defer restored.Shutdown()
	again, err := restored.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := restored.Wait(ctx, again.ID())
	if err != nil || string(recovered.Submission) != `{"result":"done"}` {
		t.Fatalf("recover %+v %v", recovered, err)
	}
	if result := <-again.Submit(ctx, "must stay stopped"); result.Err == nil {
		t.Fatal("restored terminal activated")
	}
	after, _ := os.ReadFile(filepath.Join(cfg.DataDir, "submit.md"))
	if string(raw) != string(after) {
		t.Fatal("recovery duplicated row")
	}
	_ = restored.Shutdown()
	if err := os.WriteFile(filepath.Join(cfg.DataDir, "submit.md"), []byte("not json\n"), 0600); err != nil {
		t.Fatal(err)
	}
	bad := ensemble.New(log.New(io.Discard, "", 0))
	defer bad.Shutdown()
	if _, err := bad.NewAgent(cfg); err == nil {
		t.Fatal("malformed submission was ignored")
	}
}

func TestCh20JoinQuestionsAndHumanSink(t *testing.T) {
	release := make(chan struct{})
	root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) { <-release; ch20Reply(w, `{"result":"done"}`, true) })
	defer close(release)
	ctx := ch20Context(t)
	child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "questioner", Prompt: "wait"})
	if err != nil {
		t.Fatal(err)
	}
	answer := make(chan string, 1)
	go func() {
		reply, err := root.AskParent(ctx, child.ID(), "which target?")
		if err != nil {
			answer <- err.Error()
		} else {
			answer <- reply
		}
	}()
	joined, err := root.Join(ctx, parent.ID(), nil)
	if err != nil || len(joined.Questions) != 1 {
		t.Fatalf("join %+v %v", joined, err)
	}
	q := joined.Questions[0]
	if err := root.RespondToQuestion(parent.ID(), child.ID(), q.RequestID, "chosen"); err != nil {
		t.Fatal(err)
	}
	if got := <-answer; got != "chosen" {
		t.Fatal(got)
	}
	presented := make(chan ensemble.Question, 1)
	root.SetHumanQuestionSink(func(q ensemble.Question) { presented <- q })
	go func() {
		reply, err := root.AskParent(ctx, parent.ID(), "human help")
		if err != nil {
			answer <- err.Error()
		} else {
			answer <- reply
		}
	}()
	human := <-presented
	if err := root.RespondToHumanQuestion(parent.ID(), human.RequestID, "human answer"); err != nil {
		t.Fatal(err)
	}
	if got := <-answer; got != "human answer" {
		t.Fatal(got)
	}
}

func TestCh20AdmissionCleanupAndPaths(t *testing.T) {
	root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) { t.Error("unexpected provider call"); ch20Reply(w, "", false) })
	ctx := ch20Context(t)
	root.SetAgentFactory(func(ctx context.Context, request ensemble.Config) (ensemble.Agent, error) {
		cfg := parent.Config().Child(filepath.Join(parent.Config().BaseDir, request.ID))
		cfg.ID, cfg.ParentID = request.ID, request.ParentID
		a, err := root.NewAgent(cfg)
		if err != nil {
			return nil, err
		}
		return a, errors.New("factory failed after construction")
	})
	if _, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "rejected", Prompt: "task"}); err == nil {
		t.Fatal("factory error lost")
	}
	if _, ok := root.Agent("rejected"); ok {
		t.Fatal("rejected Agent retained")
	}
	root.SetAgentFactory(func(context.Context, ensemble.Config) (ensemble.Agent, error) { return parent, nil })
	if _, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "foreign", Prompt: "task"}); err == nil {
		t.Fatal("foreign return admitted")
	}
	if err := parent.Post(ensemble.Hint{Text: "still usable"}); err != nil {
		t.Fatal(err)
	}
	bad := parent.Config().Child(filepath.Join(parent.Config().BaseDir, "bad"))
	bad.ID = "../bad"
	if _, err := root.NewAgent(bad); err == nil {
		t.Fatal("bad root ID admitted")
	}
	for i, path := range []string{parent.Config().SavePath, parent.Config().SavePath + ".bak", filepath.Join(parent.Config().DataDir, "submit.md")} {
		cfg := parent.Config().Child(filepath.Join(parent.Config().BaseDir, "collision"))
		cfg.ID = string(rune('a' + i))
		cfg.LogPath = path
		if _, err := root.NewAgent(cfg); err == nil {
			t.Fatal("cross-kind state collision", path)
		}
	}
}

func TestCh20CustomSubmitCorrectionAndPersistenceFailures(t *testing.T) {
	t.Run("custom correction", func(t *testing.T) {
		var count atomic.Int32
		root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) {
			n := count.Add(1)
			args := `{"answer":"wrong"}`
			if n > 1 {
				if !strings.Contains(string(b), "$.answer: violates type") {
					t.Error("missing validation feedback")
				}
				args = `{"answer":2}`
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "fixture", "choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "tool_calls": []any{map[string]any{"id": fmt.Sprintf("submit-%d", n), "type": "function", "function": map[string]any{"name": "submit", "arguments": args}}}}, "finish_reason": "tool_calls"}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}})
		})
		ctx := ch20Context(t)
		schema := json.RawMessage(`{"type":"object","properties":{"answer":{"type":"integer","minimum":1}},"required":["answer"],"additionalProperties":false}`)
		child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "custom", Prompt: "answer", SubmitSchema: schema})
		if err != nil {
			t.Fatal(err)
		}
		outcome, err := root.Wait(ctx, child.ID())
		if err != nil || outcome.State != "submitted" || string(outcome.Submission) != `{"answer":2}` || count.Load() != 2 {
			t.Fatalf("outcome %+v %v requests=%d", outcome, err, count.Load())
		}
		row, err := os.ReadFile(filepath.Join(child.Config().DataDir, "submit.md"))
		if err != nil || !strings.Contains(string(row), `"payload":{"answer":2}`) {
			t.Fatalf("row %s %v", row, err)
		}
	})
	for _, failure := range []string{"submit row", "final save"} {
		t.Run(failure, func(t *testing.T) {
			arrived, release := make(chan struct{}, 1), make(chan struct{})
			root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) {
				arrived <- struct{}{}
				<-release
				ch20Reply(w, `{"result":"done"}`, true)
			})
			ctx := ch20Context(t)
			child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "worker", Prompt: "finish"})
			if err != nil {
				t.Fatal(err)
			}
			select {
			case <-arrived:
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			path := filepath.Join(child.Config().DataDir, "submit.md")
			if failure == "final save" {
				path = child.Config().SavePath
				_ = os.Remove(path)
			}
			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
			close(release)
			outcome, err := root.Wait(ctx, child.ID())
			if err != nil || outcome.State != "failed" || outcome.Error == "" || len(outcome.Submission) != 0 {
				t.Fatalf("false success %+v %v", outcome, err)
			}
			if err := parent.Post(ensemble.Hint{Text: "root remains usable"}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCh20AdmissionCancellationMutationAndAliases(t *testing.T) {
	root, parent := ch20Rig(t, func(w http.ResponseWriter, b []byte) { t.Error("unexpected provider request"); ch20Reply(w, "", false) })
	ctx, cancel := context.WithCancel(ch20Context(t))
	root.SetAgentFactory(func(ctx context.Context, request ensemble.Config) (ensemble.Agent, error) {
		cfg := parent.Config().Child(filepath.Join(parent.Config().BaseDir, request.ID))
		cfg.ID, cfg.ParentID = request.ID, request.ParentID
		cfg.SubmitSchema = request.SubmitSchema
		a, err := root.NewAgent(cfg)
		cancel()
		return a, err
	})
	if _, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "cancelled", Prompt: "task"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation %v", err)
	}
	if _, ok := root.Agent("cancelled"); ok {
		t.Fatal("cancelled admission remains registered")
	}
	schema := json.RawMessage(`{"type":"object","properties":{"x":{"type":"number"}}}`)
	before := string(schema)
	root.SetAgentFactory(func(ctx context.Context, request ensemble.Config) (ensemble.Agent, error) {
		for i := range request.SubmitSchema {
			if request.SubmitSchema[i] == 'x' {
				request.SubmitSchema[i] = 'y'
				break
			}
		}
		cfg := parent.Config().Child(filepath.Join(parent.Config().BaseDir, request.ID))
		cfg.ID, cfg.ParentID = request.ID, request.ParentID
		cfg.SubmitSchema = request.SubmitSchema
		return root.NewAgent(cfg)
	})
	if _, err := root.Spawn(ch20Context(t), parent.ID(), ensemble.Config{ID: "mutated", Prompt: "task", SubmitSchema: schema}); err == nil {
		t.Fatal("factory schema mutation admitted")
	}
	if string(schema) != before {
		t.Fatal("caller schema mutated")
	}
	if _, ok := root.Agent("mutated"); ok {
		t.Fatal("mutated admission retained")
	}
	alias := filepath.Join(parent.Config().BaseDir, "alias")
	if err := os.Symlink(parent.Config().DataDir, alias); err != nil {
		t.Fatal(err)
	}
	cfg := parent.Config().Child(alias)
	cfg.ID = "alias-agent"
	cfg.SavePath = filepath.Join(parent.Config().BaseDir, "separate-save.json")
	if _, err := root.NewAgent(cfg); err == nil {
		t.Fatal("canonical submit.md alias admitted")
	}
	cfg = parent.Config().Child(filepath.Join(parent.Config().BaseDir, "explicit"))
	cfg.ID = "agent-2"
	if _, err := root.NewAgent(cfg); err != nil {
		t.Fatal(err)
	}
	cfg = parent.Config().Child(filepath.Join(parent.Config().BaseDir, "generated"))
	cfg.ID = ""
	generated, err := root.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if generated.ID() == "agent-2" {
		t.Fatal("generated ID reused explicit identity")
	}
}
