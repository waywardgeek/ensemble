package ensemble_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func c4r2Call(t *testing.T, a ensemble.Agent, name string, args map[string]any, wantError bool) string {
	t.Helper()
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatal(err)
	}
	return c4Dispatch(t, a, name, string(raw), wantError)
}

const c4EchoCommand = `printf 'ECHO_READY\n'; while IFS= read -r line; do if [ "$line" = quit ]; then printf 'ECHO_BYE\n'; exit 0; fi; printf 'echo: %s\n' "$line"; done`

func TestChapter4Round2TerminalSettingsAndParity(t *testing.T) {
	t.Setenv("TERM", "inherited-must-not-survive")
	_, a, dir := c4Agent(t)
	command := `printf 'TERM=%s\n' "$TERM"; stty size; printf 'utf8:π☃\n'; printf 'CR-only:\rX'`
	text := c4r2Call(t, a, "run_command", map[string]any{"command": command}, false)
	data, disk := c3JobOutput(t, a)
	want := "TERM=dumb\n50 200\nutf8:π☃\nCR-only:\rX\nexit_code: 0\n"
	if disk != want || text != disk || data.Bytes != int64(len(want)) || data.ExitCode == nil || *data.ExitCode != 0 {
		t.Fatalf("PTY settings, normalization or parity: job=%+v disk=%q report=%q", data, disk, text)
	}
	if strings.Count(disk, "exit_code:") != 1 {
		t.Fatalf("footer count: %q", disk)
	}
	c4r2Call(t, a, "tool_limits", map[string]any{"max_output_bytes": 8}, false)
	text = c4r2Call(t, a, "run_command", map[string]any{"command": "printf 0123456789abcdef"}, false)
	data, disk = c3JobOutput(t, a)
	if !strings.HasPrefix(text, "[tool_limits consumed by this run_command call:") || !strings.Contains(text, fmt.Sprintf("job %d done;", data.Handle)) || !strings.Contains(text, "output truncated;") || disk != "0123456789abcdef\nexit_code: 0\n" {
		t.Fatalf("truncation/header/diagnostic: %q disk=%q", text, disk)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0700); err != nil {
		t.Fatal(err)
	}
	text = c4r2Call(t, a, "run_command", map[string]any{"command": "printf cwd", "cwd": "sub"}, false)
	if !strings.Contains(text, "; cwd: "+filepath.Join(dir, "sub")) {
		t.Fatalf("cwd header lost: %q", text)
	}
}

func TestChapter4Round2LineInputAndTerminalCursor(t *testing.T) {
	_, a, _ := c4Agent(t)
	text := c4r2Call(t, a, "run_command", map[string]any{"command": c4EchoCommand, "ai_callback_pattern": "ECHO_READY\n", "ai_callback_delay": 2}, false)
	if !strings.Contains(text, "job 1 running;") || !strings.HasSuffix(text, "ECHO_READY\n") {
		t.Fatalf("initial prompt: %q", text)
	}
	for _, line := range []string{"hello world", "already\n", "carriage\r", ""} {
		// No supplied callback pattern: echo alone is not the whole first response.
		text = c4r2Call(t, a, "send_input", map[string]any{"handle": 1, "input": line, "ai_callback_delay": 0.1}, false)
		want := "echo: " + strings.TrimRight(line, "\r\n") + "\n"
		if !strings.Contains(text, want) || strings.Count(text, "echo: ") != 1 || strings.Contains(text, "ECHO_READY") || strings.Contains(text, "\r") {
			t.Fatalf("line %q: %q", line, text)
		}
	}
	text = c4r2Call(t, a, "send_input", map[string]any{"handle": 1, "input": "quit", "ai_callback_delay": 2}, false)
	if !strings.Contains(text, "job 1 done;") || !strings.HasSuffix(text, "quit\nECHO_BYE\n\nexit_code: 0\n") {
		t.Fatalf("quit without caller newline: %q", text)
	}
	data, disk := c3JobOutput(t, a)
	want := "ECHO_READY\nhello world\necho: hello world\nalready\necho: already\ncarriage\necho: carriage\n\necho: \nquit\nECHO_BYE\n\nexit_code: 0\n"
	if disk != want || data.Bytes != int64(len(want)) {
		t.Fatalf("echo/terminator/full retained bytes: %q job=%+v", disk, data)
	}
	text = c4r2Call(t, a, "wait_for_job", map[string]any{"handle": 1, "ai_callback_delay": 0}, false)
	if strings.Contains(text, "ECHO_BYE") || strings.Contains(text, "echo: ") {
		t.Fatalf("terminal cursor replay: %q", text)
	}
	for _, definition := range a.Tools().Definitions() {
		if definition.Name == "send_input" && (!strings.Contains(definition.Description, "line") || strings.Contains(definition.Description, "without appending")) {
			t.Fatalf("misleading input description: %q", definition.Description)
		}
	}
}

func TestChapter4Round2StagedCRLFAndUnseenPattern(t *testing.T) {
	_, a, _ := c4Agent(t)
	// Disable only this fixture's ONLCR/echo so the raw CR and LF writes are
	// deliberate. The production terminal keeps echo, checked in the test above.
	command := `stty -echo -onlcr; printf 'first\r'; IFS= read -r line; printf '\nsecond\rX π☃\r'; IFS= read -r line; printf '\nEND\r'`
	text := c4r2Call(t, a, "run_command", map[string]any{"command": command, "ai_callback_pattern": "first", "ai_callback_delay": 2}, false)
	if !strings.HasSuffix(text, "first") {
		t.Fatalf("first chunk: %q", text)
	}
	job, err := a.Jobs().Get(1)
	if err != nil {
		t.Fatal(err)
	}
	if job.Data().Bytes != 5 {
		t.Fatalf("pending CR was retained prematurely: %+v", job.Data())
	}
	text = c4r2Call(t, a, "send_input", map[string]any{"handle": 1, "input": "go", "ai_callback_pattern": "^" + regexp.QuoteMeta("\nsecond\rX π☃") + "$", "ai_callback_delay": 2}, false)
	if !strings.Contains(text, "job 1 running;") || !strings.HasSuffix(text, "\nsecond\rX π☃") || strings.Contains(text, "first") {
		t.Fatalf("normalized unseen pattern/cursor: %q", text)
	}
	text = c4r2Call(t, a, "send_input", map[string]any{"handle": 1, "input": "finish", "ai_callback_delay": 2}, false)
	if !strings.Contains(text, "job 1 done;") || !strings.HasSuffix(text, "\nEND\r\nexit_code: 0\n") || strings.Contains(text, "second") {
		t.Fatalf("terminal pending CR/cursor: %q", text)
	}
	data, disk := c3JobOutput(t, a)
	want := "first\nsecond\rX π☃\nEND\r\nexit_code: 0\n"
	if disk != want || data.Bytes != int64(len(want)) {
		t.Fatalf("normalized disk/bytes: %q %+v", disk, data)
	}
}

func TestChapter4Round2AgentWireAndKilledObservation(t *testing.T) {
	dir := t.TempDir()
	app := ensemble.New(nil)
	t.Cleanup(func() {
		if err := app.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	var requests atomic.Int32
	type wireResult struct {
		text   string
		failed bool
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			http.Error(w, "read", 500)
			return
		}
		results := map[string]wireResult{}
		var collect func(any)
		collect = func(v any) {
			switch v := v.(type) {
			case map[string]any:
				if v["type"] == "tool_result" {
					id, _ := v["tool_use_id"].(string)
					failed, _ := v["is_error"].(bool)
					var text strings.Builder
					switch content := v["content"].(type) {
					case string:
						text.WriteString(content)
					case []any:
						for _, part := range content {
							if part, ok := part.(map[string]any); ok {
								if value, ok := part["text"].(string); ok {
									text.WriteString(value)
								}
							}
						}
					}
					results[id] = wireResult{text.String(), failed}
				}
				for _, nested := range v {
					collect(nested)
				}
			case []any:
				for _, nested := range v {
					collect(nested)
				}
			}
		}
		collect(body)
		check := func(id, contains string, failed bool) {
			result, ok := results[id]
			if !ok || result.failed != failed || !strings.Contains(result.text, contains) {
				t.Errorf("wire result %s: %+v want %q failed=%v", id, result, contains, failed)
			}
		}
		call := func(id, name string, args map[string]any) map[string]any {
			return map[string]any{"type": "tool_use", "id": id, "name": name, "input": args}
		}
		var content []map[string]any
		switch requests.Add(1) {
		case 1:
			content = []map[string]any{call("fast-exact", "run_command", map[string]any{"command": "printf 'out\\n'; printf 'err\\n' >&2"}), call("prompt-exact", "run_command", map[string]any{"command": c4EchoCommand, "ai_callback_pattern": "ECHO_READY\n", "ai_callback_delay": 2})}
		case 2:
			check("fast-exact", "out\nerr\n\nexit_code: 0\n", false)
			if results["fast-exact"].text != "out\nerr\n\nexit_code: 0\n" {
				t.Errorf("fast wire result is not the bare retained output: %+v", results["fast-exact"])
			}
			check("prompt-exact", "ECHO_READY\n", false)
			content = []map[string]any{call("input-exact", "send_input", map[string]any{"handle": 2, "input": "hello world", "ai_callback_pattern": "echo: hello world\n", "ai_callback_delay": 2})}
		case 3:
			check("input-exact", "hello world\necho: hello world\n", false)
			content = []map[string]any{call("quit-exact", "send_input", map[string]any{"handle": 2, "input": "quit", "ai_callback_delay": 2})}
		case 4:
			check("quit-exact", "ECHO_BYE\n\nexit_code: 0\n", false)
			content = []map[string]any{call("held-exact", "run_command", map[string]any{"command": "printf 'HOLD_READY\\n'; read line", "ai_callback_pattern": "HOLD_READY\n", "ai_callback_delay": 2})}
		case 5:
			check("held-exact", "HOLD_READY\n", false)
			content = []map[string]any{call("kill-exact", "kill_job", map[string]any{"handle": 3}), call("observe-exact", "wait_for_job", map[string]any{"handle": 3, "ai_callback_delay": 0}), call("missing-exact", "wait_for_job", map[string]any{"handle": 9999, "ai_callback_delay": 0})}
		case 6:
			check("kill-exact", "job 3 killed;", false)
			check("observe-exact", "reason: kill_job", false)
			check("missing-exact", "unknown", true)
			if strings.Contains(results["observe-exact"].text, "exit_code") {
				t.Error("killed observation invented exit metadata")
			}
			content = []map[string]any{{"type": "text", "text": "wire checks complete"}}
		default:
			t.Error("unexpected extra request")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"content": content, "stop_reason": "end_turn", "usage": map[string]int{"input_tokens": 1, "output_tokens": 1}})
	}))
	defer server.Close()
	// One unnamed Agent, with caller-owned state and a derived SavePath.
	a, err := app.NewAgent(ensemble.Config{
		Model: "local-only", BaseDir: dir, DataDir: filepath.Join(dir, "cr"),
		BaseURL: server.URL, APIKey: "loopback-only",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	answer, err := a.Ask(ctx, "Exercise the local PTY and supervision tools.")
	if err != nil || answer != "wire checks complete" || requests.Load() != 6 {
		t.Fatalf("Agent wire: %q %v requests=%d", answer, err, requests.Load())
	}
	killed, err := a.Jobs().Get(3)
	if err != nil {
		t.Fatal(err)
	}
	c4Killed(t, a, killed, "kill_job")
	killedPath := killed.Data().Output.Locator
	if !filepath.IsAbs(killedPath) {
		killedPath = filepath.Join(a.Config().BaseDir, killedPath)
	}
	killedBytes, readErr := os.ReadFile(killedPath)
	if readErr != nil || string(killedBytes) != "HOLD_READY\n" || killed.Data().Bytes != int64(len(killedBytes)) {
		t.Fatalf("killed output/footer/bytes: %q %+v %v", killedBytes, killed.Data(), readErr)
	}
	c4r2Call(t, a, "send_input", map[string]any{"handle": 3, "input": "too late"}, true)
	report, err := killed.Wait(ctx, ensemble.ToolLimits{MaxOutputBytes: 1024})
	if err != nil || report.IsError {
		t.Fatalf("known killed job is not a successful observation: %+v %v", report, err)
	}
	pairs := map[string][2]int{}
	for _, event := range a.Log().Events() {
		if event.Tool != nil {
			pair := pairs[event.Tool.CallID]
			if event.Type == "tool_called" {
				pair[0]++
			}
			if event.Type == "tool_returned" {
				pair[1]++
			}
			pairs[event.Tool.CallID] = pair
		}
	}
	for _, id := range []string{"fast-exact", "prompt-exact", "input-exact", "quit-exact", "held-exact", "kill-exact", "observe-exact", "missing-exact"} {
		if pairs[id] != [2]int{1, 1} {
			t.Fatalf("call pairing %s: %v", id, pairs[id])
		}
	}
	c4Register(t, a, "actual_error", func(context.Context, ensemble.ToolCall, json.RawMessage) error {
		return errors.New("handler failure preserved")
	})
	c4r2Call(t, a, "actual_error", map[string]any{}, true)
	c4r2Call(t, a, "wait_for_job", map[string]any{"handle": 4, "ai_callback_delay": 0}, true)
	c4r2Call(t, a, "send_input", map[string]any{"handle": 4, "input": "unsupported"}, true)
	if err := app.Shutdown(); err != nil {
		t.Fatal(err)
	}
}
