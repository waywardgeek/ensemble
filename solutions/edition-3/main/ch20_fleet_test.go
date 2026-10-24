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

	"ensemble"
)

func TestCh20FleetActualAgentSubmit(t *testing.T) {
	var requests atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(string(body), "root followup") {
			fmt.Fprint(w, `{"id":"root","usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2},"choices":[{"message":{"role":"assistant","content":"root remains usable"},"finish_reason":"stop"}]}`)
			return
		}
		fmt.Fprint(w, `{"id":"child","usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2},"choices":[{"message":{"role":"assistant","content":null,"tool_calls":[{"id":"finish","type":"function","function":{"name":"submit","arguments":"{\"result\":\"done\"}"}}]},"finish_reason":"tool_calls"}]}`)
	}))
	defer peer.Close()
	root := ensemble.New(nil)
	defer root.Shutdown()
	var builds atomic.Int32
	root.SetAgentFactory(func(ctx context.Context, request ensemble.Config) (ensemble.Agent, error) {
		builds.Add(1)
		config := request
		if request.ParentID != "" {
			parent, ok := root.Agent(request.ParentID)
			if !ok {
				return nil, fmt.Errorf("missing parent")
			}
			config = parent.Config().Child(filepath.Join(parent.Config().DataDir, "children", request.ID))
			config.ID, config.ParentID, config.Prompt = request.ID, request.ParentID, request.Prompt
			config.SubmitSchema = append(json.RawMessage(nil), request.SubmitSchema...)
		}
		return root.NewAgent(config)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	workspace := t.TempDir()
	parent, err := root.CreateAgent(ctx, ensemble.Config{ID: "root", Vendor: ensemble.OpenAI, Model: "fixture", APIKey: "dummy", BaseURL: peer.URL, BaseDir: workspace, DataDir: filepath.Join(workspace, "state"), SavePath: filepath.Join(workspace, "root-save.json")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "bad", Prompt: "task", SubmitSchema: json.RawMessage(`{"type":"object","unknown":true}`)})
	if err == nil || builds.Load() != 1 {
		t.Fatalf("schema admission: err=%v builds=%d", err, builds.Load())
	}
	child, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "worker", Prompt: "finish the task"})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := root.Wait(ctx, child.ID())
	if err != nil {
		t.Fatal(err)
	}
	if outcome.State != "submitted" || string(outcome.Submission) != `{"result":"done"}` {
		t.Fatalf("outcome: %+v", outcome)
	}
	if requests.Load() != 1 {
		t.Fatalf("submit made %d provider requests", requests.Load())
	}
	if child.Config().BaseDir != parent.Config().BaseDir || child.Config().DataDir == parent.Config().DataDir {
		t.Fatal("workspace/state ownership mismatch")
	}
	outcome.Submission[0] = '!'
	repeated, err := root.Wait(ctx, child.ID())
	if err != nil || string(repeated.Submission) != `{"result":"done"}` {
		t.Fatalf("retained outcome: %+v %v", repeated, err)
	}
	if _, err := root.Spawn(ctx, parent.ID(), ensemble.Config{ID: "worker", Prompt: "duplicate"}); err == nil {
		t.Fatal("duplicate terminal ID accepted")
	}
	reply := <-parent.Submit(ctx, "root followup")
	if reply.Err != nil || reply.Text != "root remains usable" {
		t.Fatalf("root reply: %+v", reply)
	}
}
