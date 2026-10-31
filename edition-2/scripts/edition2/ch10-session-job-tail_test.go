package ensemble

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// A real shell waits for an explicit file release. The first paired report and
// completed turn establish a settled candidate while the independent job lives.
// Its later terminal event must append while real checkpoint replacement waits.
func TestCh10ReviewIndependentJobTail(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		if requests.Add(1) == 1 {
			args, _ := json.Marshal(map[string]any{
				"command":             "printf CH10_READY; while [ ! -f release-review-job ]; do sleep 0.01; done; printf CH10_TERMINAL",
				"ai_callback_pattern": "CH10_READY", "ai_callback_delay": 1,
			})
			_ = json.NewEncoder(w).Encode(map[string]any{"model": "gpt-4.1-mini-2025-04-14",
				"choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": "", "tool_calls": []any{
					map[string]any{"id": "review-job", "type": "function", "function": map[string]any{"name": "run_command", "arguments": string(args)}},
				}}, "finish_reason": "tool_calls"}}, "usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1}})
			return
		}
		_, _ = io.WriteString(w, `{"model":"gpt-4.1-mini-2025-04-14","choices":[{"index":0,"message":{"role":"assistant","content":"Independent job remains running."},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
	}))
	defer server.Close()
	r, a, options := ch10ReviewOpen(t)
	config := a.Config()
	config.Vendor, config.Model, config.ResolvedModel = "openai", "gpt-4.1-mini-2025-04-14", "gpt-4.1-mini-2025-04-14"
	config.BaseURL, config.DisableStreaming = server.URL, true
	// Installed handlers are creation identity, so construct this session with its
	// intended ceiling rather than mutate that identity after opening.
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
	options.Config = config
	system := config.System
	options.System = &system
	options.Config.System = ""
	options.Config.DataDir = filepath.Join(config.Workspace, "job-session")
	options.Config.LogPath = ""
	options.Config.Builtins = []string{"run_command"}
	var err error
	a, err = r.OpenSession(options)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := a.Submit("Run the local controlled independent job.")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	completion, err := handle.Wait(ctx)
	if err != nil || completion.Outcome != "success" || requests.Load() != 2 {
		t.Fatalf("positive paired running-job turn: %+v %v requests=%d", completion, err, requests.Load())
	}
	state := a.Snapshot()
	if len(state.Jobs) != 1 {
		t.Fatalf("expected one real job, got %d", len(state.Jobs))
	}
	var job uint64
	for id, fact := range state.Jobs {
		job = id
		if fact.Status != "running" {
			t.Fatalf("job finished before fixture release: %+v", fact)
		}
	}
	boundary := state.LastSeq
	_, watch, err := a.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer watch.Close()
	disk := ch10ReviewWrap(a, "", "replace")
	var once sync.Once
	unblock := func() { once.Do(func() { close(disk.release) }) }
	defer unblock()
	type result struct {
		ack CheckpointAck
		err error
	}
	saved := make(chan result, 1)
	go func() { ack, e := a.Checkpoint(); saved <- result{ack, e} }()
	ch10ReviewWait(t, disk.entered, "real job checkpoint replacement")
	if err = os.WriteFile(filepath.Join(config.Workspace, "release-review-job"), []byte("release"), 0600); err != nil {
		t.Fatal(err)
	}
	var terminal uint64
	for {
		record, e := watch.Next(ctx)
		if e != nil {
			t.Fatal("independent job fact blocked by checkpoint writer", e)
		}
		event := record.Observation.Event
		if event.Type == "job_ended" && event.Job != nil && event.Job.Handle == job {
			terminal = event.Seq
			if event.Job.Status != "done" || event.Job.ExitCode == nil || *event.Job.ExitCode != 0 {
				t.Fatalf("not a genuine successful terminal fact: %+v", event.Job)
			}
			break
		}
	}
	if terminal <= boundary {
		t.Fatal("terminal event did not belong to a newer tail")
	}
	if a.Snapshot().Jobs[job].Status != "done" {
		t.Fatal("observed terminal fact was not accepted")
	}
	unblock()
	got := ch10ReviewWait(t, saved, "captured checkpoint completion")
	if got.err != nil || got.ack.AsOf != boundary {
		t.Fatalf("checkpoint anchor includes post-capture job event: %+v %v", got.ack, got.err)
	}
	checkpointPath := filepath.Join(options.Config.DataDir, "checkpoint.json")
	oldCheckpoint := ch10ReviewBytes(t, checkpointPath)
	inspection, err := r.InspectCheckpoint(oldCheckpoint)
	if err != nil || inspection.Snapshot().Jobs[job].Status != "running" || inspection.Boundary().LogSeq != boundary {
		t.Fatalf("captured job projection raced ahead of accepted prefix: %v", err)
	}
	if disk.count("replace") != 1 {
		t.Fatal("real replacement barrier was not exercised once")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	// Deliberately restore the earlier genuine checkpoint, keeping its real tail.
	if err = os.WriteFile(checkpointPath, oldCheckpoint, 0600); err != nil {
		t.Fatal(err)
	}
	a, err = r.OpenSession(options)
	if err != nil {
		t.Fatal("old checkpoint/new terminal tail failed resume", err)
	}
	if a.Snapshot().Jobs[job].Status != "done" || requests.Load() != 2 {
		t.Fatal("job tail was lost or startup sent HTTP")
	}
	count := 0
	for _, event := range a.Events() {
		if event.Type == "job_ended" && event.Job != nil && event.Job.Handle == job {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("terminal fact was replayed %d times", count)
	}
}

// Candidate validation failure must not be confused with a failed durable write.
func TestCh10ReviewInvalidAppendStaysUsable(t *testing.T) {
	_, a, _ := ch10ReviewOpen(t)
	if err := a.Append(Event{Type: "unknown_review_event"}); err == nil {
		t.Fatal("invalid candidate accepted")
	}
	if err := ch10ReviewAppend(a, "accepted after ordinary validation refusal"); err != nil {
		t.Fatal("ordinary refusal faulted Agent", err)
	}
	if _, err := a.Checkpoint(); err != nil {
		t.Fatal("healthy checkpoint after refused candidate", err)
	}
	if err := a.Close(); err != nil {
		t.Fatal("healthy close after refused candidate", err)
	}
}
