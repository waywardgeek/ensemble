package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"ensemble"
)

func c4Agent(t *testing.T) (*ensemble.Ensemble, ensemble.Agent, string) {
	t.Helper()
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
	return app, a, dir
}
func c4Dispatch(t *testing.T, a ensemble.Agent, name, args string, wantError bool) string {
	t.Helper()
	parts, failed := a.Tools().Dispatch(context.Background(), name, json.RawMessage(args))
	var text strings.Builder
	for _, p := range parts {
		text.WriteString(p.Text)
	}
	if failed != wantError {
		t.Fatalf("%s failed=%v want %v: %s", name, failed, wantError, text.String())
	}
	return text.String()
}
func c4Register(t *testing.T, a ensemble.Agent, name string, handler ensemble.ToolHandler) {
	t.Helper()
	if err := a.Tools().Register(ensemble.ToolDefinition{Name: name, Description: "Focused local regression.", Parameters: json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)}, handler); err != nil {
		t.Fatal(err)
	}
}
func c4Signal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("channel handshake timed out")
	}
}
func c4Killed(t *testing.T, a ensemble.Agent, job ensemble.Job, reason string) {
	t.Helper()
	data := job.Data()
	if data.Status != ensemble.JobKilledStatus || data.ExitCode != nil || data.Reason != reason {
		t.Fatalf("killed data: %+v", data)
	}
	report, err := job.Wait(context.Background(), ensemble.ToolLimits{MaxOutputBytes: 1024})
	if err != nil || report.IsError || report.Job.ExitCode != nil || report.Job.Status != ensemble.JobKilledStatus {
		t.Fatalf("killed report: %+v, %v", report, err)
	}
	count := 0
	for _, event := range a.Log().Events() {
		if event.Type != ensemble.JobKilled || event.Job.Handle != data.Handle {
			continue
		}
		count++
		raw, err := json.Marshal(event)
		if err != nil || bytes.Contains(raw, []byte(`"exit_code"`)) {
			t.Fatalf("killed serialization: %s %v", raw, err)
		}
	}
	if count != 1 {
		t.Fatalf("job_killed count=%d", count)
	}
}

func TestChapter4Round1KilledMetadata(t *testing.T) {
	for _, reason := range []string{"kill_job", "shutdown"} {
		t.Run(reason, func(t *testing.T) {
			app, a, dir := c4Agent(t)
			ownerReady := make(chan struct{})
			late := make(chan struct{})
			lateDone := make(chan struct{})
			cancelled := make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(late) }) }
			t.Cleanup(release)
			c4Register(t, a, "late", func(ctx context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
				call.SetExitCode(7)
				if _, err := io.WriteString(call, "before kill"); err != nil {
					return err
				}
				close(ownerReady)
				if reason == "shutdown" {
					<-ctx.Done()
					close(cancelled)
				}
				<-late // Deliberately uncooperative handler: late effects must remain inert.
				call.SetExitCode(99)
				_, err := io.WriteString(call, "FORBIDDEN LATE OUTPUT")
				close(lateDone)
				return err
			})
			c4Dispatch(t, a, "tool_limits", `{"ai_callback_delay":0}`, false)
			c4Dispatch(t, a, "late", `{}`, false)
			c4Signal(t, ownerReady)
			job, err := a.Jobs().Get(1)
			if err != nil {
				t.Fatal(err)
			}
			if job.Data().ExitCode == nil {
				t.Fatal("test did not set the pre-kill exit metadata")
			}
			if _, err := job.Wait(context.Background(), ensemble.ToolLimits{MaxOutputBytes: 1024}); err != nil {
				t.Fatal(err)
			}
			blocked, armed := c4Observe(t, job, "")
			c4Signal(t, armed)
			if reason == "shutdown" {
				stopped := make(chan error, 1)
				go func() { stopped <- app.Shutdown() }()
				c4Signal(t, cancelled)
				select {
				case err := <-stopped:
					t.Fatalf("shutdown abandoned held Go work: %v", err)
				default:
				}
				if d := job.Data(); d.Status != ensemble.JobRunning || d.ExitCode == nil || *d.ExitCode != 7 || d.Reason != "" {
					t.Fatalf("invented held shutdown outcome: %+v", d)
				}
				select {
				case got := <-blocked:
					t.Fatalf("shutdown fabricated observation completion: %+v", got)
				default:
				}
				release()
				c4Signal(t, lateDone)
				select {
				case err := <-stopped:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("shutdown did not join released handler")
				}
				wake := c4Result(t, blocked)
				if !strings.Contains(wake.Text, "FORBIDDEN LATE OUTPUT") {
					t.Fatalf("shutdown lost real output wake: %+v", wake)
				}
				// Wait may wake on output while the handler is still running.
				// Shutdown has now joined it; inspect fresh terminal data rather
				// than requiring the earlier report to update itself.
				data := job.Data()
				if data.Status != ensemble.JobDone || data.ExitCode == nil || *data.ExitCode != 99 || data.Reason != "" {
					t.Fatalf("shutdown lost real outcome: data=%+v earlier_wake=%+v", data, wake)
				}
				disk, err := os.ReadFile(filepath.Join(dir, "cr", "io", "1"))
				if err != nil || string(disk) != "before killFORBIDDEN LATE OUTPUT" {
					t.Fatalf("shutdown discarded real output: %q %v", disk, err)
				}
				for _, e := range a.Log().Events() {
					if e.Type == ensemble.JobKilled {
						t.Fatal("shutdown fabricated a killed Go job")
					}
				}
				return
			}
			err = job.Kill(reason)
			if err != nil {
				t.Fatal(err)
			}
			wake := c4Result(t, blocked)
			if wake.Job.Status != ensemble.JobKilledStatus || wake.Job.ExitCode != nil {
				t.Fatalf("kill wake: %+v", wake)
			}
			c4Killed(t, a, job, reason)
			before := job.Data()
			release()
			c4Signal(t, lateDone)
			c4Killed(t, a, job, reason)
			after := job.Data()
			if before.Bytes != after.Bytes {
				t.Fatalf("late writer changed bytes: %d -> %d", before.Bytes, after.Bytes)
			}
			disk, err := os.ReadFile(filepath.Join(dir, "cr", "io", "1"))
			if err != nil || string(disk) != "before kill" {
				t.Fatalf("late disk write: %q, %v", disk, err)
			}
		})
	}
	for _, reason := range []string{"kill_job", "shutdown"} {
		t.Run("process-"+reason, func(t *testing.T) {
			app, a, dir := c4Agent(t)
			c4Dispatch(t, a, "run_command", `{"command":"printf '%s' $$ > pid; printf 'ready> '; read value","ai_callback_pattern":"ready> ","ai_callback_delay":2}`, false)
			job, err := a.Jobs().Get(1)
			if err != nil {
				t.Fatal(err)
			}
			pidBytes, err := os.ReadFile(filepath.Join(dir, "pid"))
			if err != nil {
				t.Fatal(err)
			}
			pid, err := strconv.Atoi(string(pidBytes))
			if err != nil {
				t.Fatal(err)
			}
			if reason == "shutdown" {
				err = app.Shutdown()
			} else {
				err = job.Kill(reason)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = syscall.Kill(pid, 0); err != syscall.ESRCH {
				t.Fatalf("process %d was not reaped: %v", pid, err)
			}
			c4Killed(t, a, job, reason)
			if err = job.Kill("kill_job"); err != nil {
				t.Fatal(err)
			}
			c4Killed(t, a, job, reason)
		})
	}
}

// Done is evaluated only after Wait snapshots its changed channel and unlocks.
// This handshake coordinates real blocked observers without sleeps or test hooks
// in production. Repeated evaluations also show a regex wait re-armed.
type c4WaitContext struct {
	context.Context
	armed chan struct{}
}

func (c *c4WaitContext) Done() <-chan struct{} { c.armed <- struct{}{}; return c.Context.Done() }

type c4Observed struct {
	report ensemble.JobReport
	err    error
}

func c4Observe(t *testing.T, job ensemble.Job, pattern string) (<-chan c4Observed, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	observed := make(chan c4Observed, 1)
	watch := &c4WaitContext{Context: ctx, armed: make(chan struct{}, 16)}
	go func() {
		r, err := job.Wait(watch, ensemble.ToolLimits{Delay: time.Hour, Pattern: pattern, MaxOutputBytes: 1024})
		observed <- c4Observed{r, err}
	}()
	return observed, watch.armed
}
func c4Result(t *testing.T, result <-chan c4Observed) ensemble.JobReport {
	t.Helper()
	select {
	case r := <-result:
		if r.err != nil {
			t.Fatal(r.err)
		}
		return r.report
	case <-time.After(3 * time.Second):
		t.Fatal("observer failed to return")
		return ensemble.JobReport{}
	}
}
func TestChapter4Round1OutputWakeups(t *testing.T) {
	_, a, _ := c4Agent(t)
	writes := make(chan string)
	ownerReady := make(chan struct{})
	var once sync.Once
	finish := func() { once.Do(func() { close(writes) }) }
	t.Cleanup(finish)
	c4Register(t, a, "writer", func(_ context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
		close(ownerReady)
		for text := range writes {
			if _, err := io.WriteString(call, text); err != nil {
				return err
			}
		}
		return nil
	})
	c4Dispatch(t, a, "tool_limits", `{"ai_callback_delay":0}`, false)
	c4Dispatch(t, a, "writer", `{}`, false)
	c4Signal(t, ownerReady)
	job, err := a.Jobs().Get(1)
	if err != nil {
		t.Fatal(err)
	}
	first, armed := c4Observe(t, job, "")
	c4Signal(t, armed)
	writes <- "MATCH already observed\n"
	report := c4Result(t, first)
	if report.Text != "MATCH already observed\n" || report.Job.Status != ensemble.JobRunning || report.Job.Handle != 1 {
		t.Fatalf("output wake: %+v", report)
	}
	matching, armed := c4Observe(t, job, "MATCH")
	c4Signal(t, armed) // Earlier MATCH must not satisfy this wait.
	writes <- "unmatched chunk\n"
	select {
	case <-armed: // Nonmatching new bytes woke the loop, but did not return.
	case r := <-matching:
		t.Fatalf("regex returned on old/nonmatching bytes: %+v", r)
	case <-time.After(3 * time.Second):
		t.Fatal("regex wait did not re-arm")
	}
	writes <- "fresh MATCH\n"
	report = c4Result(t, matching)
	if report.Text != "unmatched chunk\nfresh MATCH\n" {
		t.Fatalf("regex cursor: %+v", report)
	}
	done, armed := c4Observe(t, job, "")
	c4Signal(t, armed)
	finish()
	report = c4Result(t, done)
	if report.Job.Status != ensemble.JobDone || report.Text != "" {
		t.Fatalf("completion wake: %+v", report)
	}
	repeat, err := job.Wait(context.Background(), ensemble.ToolLimits{Delay: time.Hour, MaxOutputBytes: 1024})
	if err != nil || repeat.Text != "" || repeat.Job.Status != ensemble.JobDone {
		t.Fatalf("terminal repeat: %+v %v", repeat, err)
	}
}

func TestChapter4Round1ConsumedLimits(t *testing.T) {
	_, a, dir := c4Agent(t)
	if err := os.WriteFile(filepath.Join(dir, "small"), []byte("small\n"), 0600); err != nil {
		t.Fatal(err)
	}
	setter := func() {
		c4Dispatch(t, a, "tool_limits", `{"ai_callback_delay":0.5,"ai_callback_pattern":"unused","max_output_bytes":37}`, false)
	}
	check := func(text, name, delay, pattern, cap string) {
		t.Helper()
		want := fmt.Sprintf("[tool_limits consumed by this %s call: ai_callback_delay %s, ai_callback_pattern %s, max_output_bytes %s]\n", name, delay, pattern, cap)
		if !strings.HasPrefix(text, want) {
			t.Fatalf("missing prefix %q in %q", want, text)
		}
	}
	setter()
	check(c4Dispatch(t, a, "read_file", `{"path":"small"}`, false), "read_file", "500ms", `"unused"`, "37")
	if text := c4Dispatch(t, a, "read_file", `{"path":"small"}`, false); text != "small\n" {
		t.Fatalf("fast/default parity: %q", text)
	}
	for _, tc := range []struct{ name, args string }{
		{"read_file", `{"path":17}`}, {"missing", `{}`}, {"wait_for_job", `{"handle":999}`}, {"tool_limits", `{"max_output_bytes":0}`},
	} {
		setter()
		check(c4Dispatch(t, a, tc.name, tc.args, true), tc.name, "500ms", `"unused"`, "37")
		if text := c4Dispatch(t, a, "read_file", `{"path":"small"}`, false); text != "small\n" {
			t.Fatalf("failed consumer left pending limits: %q", text)
		}
	}
	setter()
	check(c4Dispatch(t, a, "wait_for_job", `{"handle":999,"ai_callback_delay":0,"ai_callback_pattern":"","max_output_bytes":19}`, true), "wait_for_job", "0s", "none", "19")
	setter()
	check(c4Dispatch(t, a, "run_command", `{"command":"printf okay","ai_callback_delay":1,"ai_callback_pattern":"","max_output_bytes":128}`, false), "run_command", "1s", "none", "128")
	setter()
	reply := c4Dispatch(t, a, "tool_limits", `{"max_output_bytes":2048}`, false)
	check(reply, "tool_limits", "500ms", `"unused"`, "37")
	if !strings.Contains(reply, "whichever tool") || !strings.Contains(reply, "then spent") {
		t.Fatalf("setter does not explain consumption: %q", reply)
	}
	check(c4Dispatch(t, a, "read_file", `{"path":"small"}`, false), "read_file", "3s", "none", "2048")
	setter()
	if err := a.Tools().Invoke(context.Background(), ensemble.Part{Type: "tool_call", CallID: "refused-exact", Name: "read_file", Args: json.RawMessage(`{"path":"small"}`)}, "policy refusal"); err != nil {
		t.Fatal(err)
	}
	events := a.Log().Events()
	returned := events[len(events)-1].Tool
	if returned == nil || !returned.IsError || returned.CallID != "refused-exact" {
		t.Fatalf("refused result: %+v", returned)
	}
	check(returned.Parts[0].Text, "read_file", "500ms", `"unused"`, "37")
}

func TestChapter4Round1Cwd(t *testing.T) {
	_, a, dir := c4Agent(t)
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0700); err != nil {
		t.Fatal(err)
	}
	start, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ id, args, want string }{
		{"cwd-override", `{"command":"printf '%s' \"$PWD\"","cwd":"sub"}`, sub},
		{"cwd-default", `{"command":"printf '%s' \"$PWD\""}`, ""},
	} {
		if err := a.Tools().Invoke(context.Background(), ensemble.Part{Type: "tool_call", CallID: tc.id, Name: "run_command", Args: json.RawMessage(tc.args)}, ""); err != nil {
			t.Fatal(err)
		}
		for _, event := range a.Log().Events() {
			if event.Tool == nil || event.Tool.CallID != tc.id {
				continue
			}
			data := event.Tool.Job
			if data == nil || data.Cwd != tc.want {
				t.Fatalf("cwd metadata: %+v", data)
			}
			raw, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(raw, []byte(`"cwd"`)) != (tc.want != "") {
				t.Fatalf("optional cwd JSON: %s", raw)
			}
			if event.Type == "tool_returned" {
				text := event.Tool.Parts[0].Text
				want := tc.want
				if want == "" {
					want = dir
				}
				if !strings.Contains(text, want) || (tc.want != "" && !strings.Contains(text, "cwd: "+sub)) {
					t.Fatalf("cwd report: %q", text)
				}
			}
		}
	}
	if after, err := os.Getwd(); err != nil || after != start || a.Config().BaseDir != dir {
		t.Fatalf("global/base directory changed: %s %v", after, err)
	}
}

func TestChapter4Round1FileStreaming(t *testing.T) {
	_, a, dir := c4Agent(t)
	if err := os.Mkdir(filepath.Join(dir, "files"), 0700); err != nil {
		t.Fatal(err)
	}
	var listing, search strings.Builder
	for i := 0; i < 96; i++ {
		name := fmt.Sprintf("file-%03d", i)
		content := "context\nHIT one\nHIT two\nend\n"
		if err := os.WriteFile(filepath.Join(dir, "files", name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintln(&listing, name)
		if i > 0 {
			search.WriteString("--\n")
		}
		fmt.Fprintf(&search, "files/%s-1-context\nfiles/%s:2:HIT one\nfiles/%s:3:HIT two\nfiles/%s-4-end\n", name, name, name, name)
	}
	for i, tc := range []struct{ name, args, want string }{
		{"list_directory", `{"path":"files"}`, listing.String()},
		{"search_files", `{"path":"files","pattern":"HIT","context_lines":1}`, search.String()},
	} {
		c4Dispatch(t, a, "tool_limits", `{"max_output_bytes":80}`, false)
		text := c4Dispatch(t, a, tc.name, tc.args, false)
		if !strings.Contains(text, "truncated") || len(text) >= len(tc.want) {
			t.Fatalf("unbounded %s inline result", tc.name)
		}
		job, err := a.Jobs().Get(i + 1)
		if err != nil {
			t.Fatal(err)
		}
		data := job.Data()
		outputPath := filepath.FromSlash(data.Output.Locator)
		if !filepath.IsAbs(outputPath) {
			outputPath = filepath.Join(dir, outputPath)
		}
		raw, err := os.ReadFile(outputPath)
		if err != nil || string(raw) != tc.want || data.Bytes != int64(len(tc.want)) {
			t.Fatalf("%s full disk/order/context: bytes=%d want=%d err=%v", tc.name, len(raw), len(tc.want), err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "small"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "small", "a"), []byte("HIT\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if text := c4Dispatch(t, a, "list_directory", `{"path":"small"}`, false); text != "a\n" {
		t.Fatalf("small list parity: %q", text)
	}
	if text := c4Dispatch(t, a, "search_files", `{"path":"small","pattern":"HIT"}`, false); text != "small/a:1:HIT\n" {
		t.Fatalf("small search parity: %q", text)
	}
}

func TestChapter4Round1SetupFailureCorrelation(t *testing.T) {
	for _, mode := range []string{"mkdir", "open", "admission"} {
		t.Run(mode, func(t *testing.T) {
			// This test deliberately breaks job storage, not recovery storage.
			dir := t.TempDir()
			app := ensemble.New(nil)
			t.Cleanup(func() {
				if err := app.Shutdown(); err != nil {
					t.Error(err)
				}
			})
			if err := os.WriteFile(filepath.Join(dir, "input"), []byte("readable\n"), 0600); err != nil {
				t.Fatal(err)
			}
			var effects atomic.Int32
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					http.Error(w, "read", 500)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if requests.Add(1) == 1 {
					io.WriteString(w, `{"content":[{"type":"tool_use","id":"limits-exact","name":"tool_limits","input":{"max_output_bytes":80}},{"type":"tool_use","id":"write-exact","name":"write_file","input":{"path":"FORBIDDEN","content":"side effect"}},{"type":"tool_use","id":"custom-exact","name":"custom","input":{}},{"type":"tool_use","id":"read-exact","name":"read_file","input":{"path":"input"}}],"usage":{"input_tokens":1,"output_tokens":2},"stop_reason":"tool_use"}`)
					return
				}
				var body map[string]any
				if err := json.Unmarshal(raw, &body); err != nil {
					t.Error(err)
				}
				results := map[string]map[string]any{}
				var collect func(any)
				collect = func(value any) {
					switch value := value.(type) {
					case map[string]any:
						if value["type"] == "tool_result" {
							id, _ := value["tool_use_id"].(string)
							results[id] = value
						}
						for _, nested := range value {
							collect(nested)
						}
					case []any:
						for _, nested := range value {
							collect(nested)
						}
					}
				}
				collect(body)
				for _, id := range []string{"limits-exact", "write-exact", "custom-exact", "read-exact"} {
					value, ok := results[id]
					wantError := id != "limits-exact" && !(mode == "open" && id == "read-exact")
					failed, _ := value["is_error"].(bool)
					if !ok || failed != wantError {
						t.Errorf("followup %s: %+v wantError=%v", id, value, wantError)
					}
				}
				io.WriteString(w, `{"content":[{"type":"text","text":"all results correlated"}],"usage":{"input_tokens":3,"output_tokens":4},"stop_reason":"end_turn"}`)
			}))
			defer server.Close()
			// Construct once, after the loopback endpoint exists; do not clone
			// an already managed identity or its terminal/recovery state.
			a, err := app.NewAgent(ensemble.Config{
				Model: "local-only", BaseDir: dir, DataDir: filepath.Join(dir, "cr"),
				SavePath: filepath.Join(t.TempDir(), "save.json"),
				BaseURL:  server.URL, APIKey: "local-regression-only",
			})
			if err != nil {
				t.Fatal(err)
			}
			c4Register(t, a, "custom", func(context.Context, ensemble.ToolCall, json.RawMessage) error { effects.Add(1); return nil })
			// Obstruct the actual tool-output boundary after construction,
			// without damaging diagnostics or the independent recovery file.
			ioDir := filepath.Join(a.Config().DataDir, "io")
			switch mode {
			case "mkdir":
				if err := os.WriteFile(ioDir, []byte("not a directory"), 0600); err != nil {
					t.Fatal(err)
				}
			case "open":
				for _, n := range []string{"1", "2"} {
					if err := os.MkdirAll(filepath.Join(ioDir, n), 0700); err != nil {
						t.Fatal(err)
					}
				}
			}
			if mode == "admission" {
				if err := a.Jobs().Shutdown(); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			answer, err := a.Ask(ctx, "Run the local file checks.")
			if err != nil || answer != "all results correlated" || requests.Load() != 2 {
				t.Fatalf("Agent aborted the batch: %q %v requests=%d", answer, err, requests.Load())
			}
			if effects.Load() != 0 {
				t.Fatal("registered handler ran without an output owner")
			}
			if _, err := os.Stat(filepath.Join(dir, "FORBIDDEN")); !os.IsNotExist(err) {
				t.Fatalf("file effect ran: %v", err)
			}
			pairs := map[string][2]int{}
			for _, event := range a.Log().Events() {
				if event.Tool == nil {
					continue
				}
				pair := pairs[event.Tool.CallID]
				switch event.Type {
				case "tool_called":
					pair[0]++
				case "tool_returned":
					pair[1]++
				}
				pairs[event.Tool.CallID] = pair
				if event.Type == "tool_returned" {
					text := event.Tool.Parts[0].Text
					consumed := strings.HasPrefix(text, "[tool_limits consumed by this write_file call:")
					if consumed != (event.Tool.CallID == "write-exact") {
						t.Fatalf("failed setup consumption: %s %q", event.Tool.CallID, text)
					}
				}
			}
			for _, id := range []string{"limits-exact", "write-exact", "custom-exact", "read-exact"} {
				if pairs[id] != [2]int{1, 1} {
					t.Fatalf("lost pair %s: %v", id, pairs[id])
				}
			}
			expectedJobs := 3
			if mode == "admission" {
				expectedJobs = 0
			}
			if len(app.JobTable()) != expectedJobs {
				t.Fatalf("job count=%d want %d", len(app.JobTable()), expectedJobs)
			}
			for _, job := range app.JobTable() {
				data := job.Data()
				if data.Status != ensemble.JobDone {
					t.Fatalf("failed setup left job running: %+v", data)
				}
				if mode == "open" && data.Handle == 3 {
					continue
				}
				if data.Bytes != 0 || data.ExitCode != nil || data.Output.Locator != "" {
					t.Fatalf("invented failed-setup output: %+v", data)
				}
				report, err := job.Wait(ctx, ensemble.ToolLimits{MaxOutputBytes: 1024})
				if err != nil || !report.IsError || !strings.Contains(report.Text, "preparing job output") {
					t.Fatalf("failed setup report: %+v %v", report, err)
				}
			}
		})
	}
}
