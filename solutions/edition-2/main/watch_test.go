package ensemble_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"example.com/ensemble"
)

func TestWatchRetainsKilledJobOutcome(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch requests {
		case 1:
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"start","name":"run_command","input":{"command":"sleep 30","ai_callback_delay":0}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		case 2:
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"stop","name":"kill_job","input":{"handle":1}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		default:
			fmt.Fprint(w, `{"content":[{"type":"text","text":"stopped"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Close()
	dir := t.TempDir()
	agent, err := app.NewAgent(ensemble.Config{DisableStreaming: true, APIKey: "fixture", Model: "fixture", BaseURL: server.URL, Workspace: dir, LogPath: filepath.Join(dir, "events"), Builtins: []string{"run_command", "kill_job"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := agent.Prompt(context.Background(), "start then kill the scratch job"); err != nil {
		t.Fatal(err)
	}
	snapshot, watch, err := agent.Watch()
	if err != nil {
		t.Fatal(err)
	}
	defer watch.Close()
	var running, killed uint64
	for _, event := range snapshot.Events {
		if event.Type == "tool_returned" && event.Tool.CallID == "start" && event.Tool.Job != nil && event.Tool.Job.Status == "running" {
			running = event.Seq
		}
		if event.Type == "job_killed" && event.Job.Handle == 1 && event.Job.Status == "killed" {
			killed = event.Seq
		}
	}
	if running == 0 || killed <= running {
		t.Fatalf("reconnect must retain the terminal fact after the running report: running=%d killed=%d", running, killed)
	}
}
