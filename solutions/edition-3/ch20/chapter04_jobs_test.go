package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble"
)

// A deliberately small implementation smoke check; lifetime/race gates are separate.
func TestChapter4InitialJobSmoke(t *testing.T) {
	dir := t.TempDir()
	app := ensemble.New(nil)
	t.Cleanup(func() {
		if err := app.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	a, err := app.NewAgent(ensemble.Config{Model: "local-only", BaseDir: dir, DataDir: "cr"})
	if err != nil {
		t.Fatal(err)
	}
	call := func(name, args string) string {
		t.Helper()
		parts, failed := a.Tools().Dispatch(context.Background(), name, json.RawMessage(args))
		if failed {
			t.Fatalf("%s: %+v", name, parts)
		}
		var text strings.Builder
		for _, p := range parts {
			text.WriteString(p.Text)
		}
		return text.String()
	}
	payload := "HEAD\n" + strings.Repeat("x", 4096) + "\nTAIL\n"
	if err := os.WriteFile(filepath.Join(dir, "input"), []byte(payload), 0600); err != nil {
		t.Fatal(err)
	}
	call("tool_limits", `{"max_output_bytes":64}`)
	text := call("read_file", `{"path":"input"}`)
	if !strings.Contains(text, "HEAD") || !strings.Contains(text, "TAIL") || !strings.Contains(text, "truncated") {
		t.Fatalf("unbounded/missing output: %s", text)
	}
	data, err := os.ReadFile(filepath.Join(dir, "cr", "io", "1"))
	if err != nil || string(data) != payload {
		t.Fatalf("disk output: %v %d", err, len(data))
	}
	if len(app.JobTable()) != 1 {
		t.Fatal("limits became a job")
	}
	err = a.Tools().Register(ensemble.ToolDefinition{Name: "upper", Description: "Uppercase text.", Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`)}, func(_ context.Context, call ensemble.ToolCall, raw json.RawMessage) error {
		if call.Parent() != a || call.Parent().Parent() != app {
			t.Error("lost parents")
		}
		var args struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return err
		}
		_, err := io.WriteString(call, strings.ToUpper(args.Text))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if text = call("upper", `{"text":"hello"}`); text != "HELLO" {
		t.Fatalf("registered output: %s", text)
	}
	text = call("run_command", `{"command":"printf 'ready> '; read answer; printf 'received:%s\\n' \"$answer\"","ai_callback_pattern":"ready> ","ai_callback_delay":2}`)
	if !strings.Contains(text, "running") || !strings.Contains(text, "ready>") {
		t.Fatalf("PTY prompt: %s", text)
	}
	text = call("send_input", `{"handle":3,"input":"hello\n","ai_callback_delay":2,"ai_callback_pattern":"received:hello"}`)
	if !strings.Contains(text, "received:hello") {
		t.Fatalf("PTY reply: %s", text)
	}
	text = call("wait_for_job", `{"handle":3,"ai_callback_delay":2,"ai_callback_pattern":"never-written-marker"}`)
	if !strings.Contains(text, "done") {
		t.Fatalf("PTY completion: %s", text)
	}
	text = call("wait_for_job", `{"handle":3,"ai_callback_delay":0}`)
	if strings.Contains(text, "received:hello") {
		t.Fatalf("cursor replayed: %s", text)
	}
	if len(app.JobTable()) != 3 {
		t.Fatal("supervision became a job")
	}
}
