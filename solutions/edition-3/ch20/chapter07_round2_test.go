package ensemble_test

import (
	"context"
	"encoding/json"
	"ensemble"
	agentowner "ensemble/internal/agent"
	"ensemble/internal/common"
	engineowner "ensemble/internal/engine"
	historyowner "ensemble/internal/history"
	jobsowner "ensemble/internal/jobs"
	skillsowner "ensemble/internal/skills"
	toolowner "ensemble/internal/tool"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type ch7BoundarySave struct {
	common.History
	entered, release chan struct{}
	once             sync.Once
	fail             bool
}

func (h *ch7BoundarySave) barrier() error {
	h.once.Do(func() { close(h.entered); <-h.release })
	if h.fail {
		return errors.New("controlled dispatch boundary save failure")
	}
	return nil
}
func (h *ch7BoundarySave) Dump(w io.Writer) error {
	for _, e := range h.Events() {
		if e.Type == ensemble.ToolReleased {
			if err := h.barrier(); err != nil {
				return err
			}
			break
		}
	}
	return h.History.Dump(w)
}
func (h *ch7BoundarySave) AppendSaved(e common.Event, save func(func(io.Writer) error) error) error {
	return h.History.AppendSaved(e, func(dump func(io.Writer) error) error {
		if e.Type == ensemble.ToolReleased {
			if err := h.barrier(); err != nil {
				return err
			}
		}
		return save(dump)
	})
}
func TestCh7R2BeforeExecutionBoundary(t *testing.T) {
	for _, tc := range []struct {
		path string
		fail bool
	}{{"invoke", false}, {"prepare-await", false}, {"model", false}, {"invoke", true}, {"prepare-await", true}, {"model", true}} {
		path := tc.path
		t.Run(fmt.Sprintf("%s/fail=%t", path, tc.fail), func(t *testing.T) {
			var requests atomic.Int32
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if requests.Add(1) == 1 {
					fmt.Fprint(w, `{"usage":{"prompt_tokens":1,"completion_tokens":1},"choices":[{"message":{"tool_calls":[{"id":"call","type":"function","function":{"name":"effect","arguments":"{}"}}]}}]}`)
				} else {
					fmt.Fprint(w, `{"usage":{"prompt_tokens":1,"completion_tokens":1},"choices":[{"message":{"content":"done"}}]}`)
				}
			}))
			defer peer.Close()
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := agentowner.New(app, ensemble.Config{Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, StreamCapability: ensemble.StreamPlain, Model: "offline", APIKey: "fake", BaseURL: peer.URL, BaseDir: t.TempDir(), LogPath: filepath.Join(t.TempDir(), "history.jsonl")})
			if err != nil {
				t.Fatal(err)
			}
			gate := &ch7BoundarySave{fail: tc.fail, History: historyowner.New(a), entered: make(chan struct{}), release: make(chan struct{})}
			a.SetHistory(gate)
			engine, err := engineowner.New(a)
			if err != nil {
				t.Fatal(err)
			}
			a.SetEngine(engine)
			a.SetJobs(jobsowner.New(a))
			tools, err := toolowner.New(a)
			if err != nil {
				t.Fatal(err)
			}
			a.SetTools(tools)
			skillRegistry, err := skillsowner.New(a, tools.Catalog())
			if err != nil {
				t.Fatal(err)
			}
			a.SetSkills(skillRegistry)
			if err := skillRegistry.LoadInitial(); err != nil {
				t.Fatal(err)
			}

			a.Start()
			defer func() {
				select {
				case <-gate.release:
				default:
					close(gate.release)
				}
				_ = a.Shutdown()
			}()
			q := a.Observe()
			defer q.Close()
			effects := make(chan bool, 1)
			err = a.Tools().Register(ensemble.ToolDefinition{Name: "effect", Description: "Boundary probe", Parameters: json.RawMessage(`{"type":"object"}`)}, func(context.Context, ensemble.ToolCall, json.RawMessage) error {
				saved, queued := false, false
				for _, e := range a.Log().EventsRange(0, a.View().Through) {
					saved = saved || e.Type == ensemble.ToolReleased
				}
				for _, o := range q.Drain() {
					if _, ok := o.(ensemble.ToolDispatched); ok {
						queued = true
					}
				}
				effects <- saved && queued
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			result := make(chan error, 1)
			go func() {
				call := ensemble.Part{Type: "tool_call", CallID: "call", Name: "effect", Args: json.RawMessage(`{}`)}
				switch path {
				case "invoke":
					result <- a.Tools().Invoke(context.Background(), call, "")
				case "prepare-await":
					d := a.Tools().Prepare(context.Background(), call, "", 0).Await(context.Background())
					if d.IsError {
						result <- fmt.Errorf("%v", d.Parts)
					} else {
						result <- nil
					}
				case "model":
					r := <-a.Submit(context.Background(), "run effect")
					result <- r.Err
				}
			}()
			select {
			case <-gate.entered:
			case <-time.After(3 * time.Second):
				select {
				case err := <-result:
					t.Fatalf("boundary save not entered; result=%v", err)
				default:
					t.Fatal("boundary save not entered")
				}
			}
			select {
			case <-effects:
				t.Fatal("handler ran before dispatch-boundary save/publication")
			case <-time.After(35 * time.Millisecond):
			}
			close(gate.release)
			if tc.fail {
				select {
				case err := <-result:
					if err == nil {
						t.Fatal("missing save failure")
					}
				case <-time.After(3 * time.Second):
					t.Fatal("failed boundary did not settle")
				}
				select {
				case <-effects:
					t.Fatal("failed dispatch save executed handler")
				default:
				}
				for _, e := range a.Log().EventsRange(0, a.View().Through) {
					if e.Type == ensemble.ToolReleased || (e.Tool != nil && e.Tool.Dispatched) {
						t.Fatalf("failed release later appeared dispatched: %#v", e)
					}
				}
				for _, o := range q.Drain() {
					if _, ok := o.(ensemble.ToolDispatched); ok {
						t.Fatal("failed release published dispatch")
					}
				}
				return
			}
			select {
			case correct := <-effects:
				if !correct {
					t.Fatal("handler began before saved and queued dispatch notification")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("handler did not run")
			}
			select {
			case err := <-result:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("call did not settle")
			}
		})
	}
}

func TestCh7R2ControlPrepareHasNoEarlyEffect(t *testing.T) {
	_, a, _ := c4Agent(t)
	body := strings.Repeat("payload", 20)
	c4Register(t, a, "control_probe", func(ctx context.Context, call ensemble.ToolCall, _ json.RawMessage) error {
		_, err := io.WriteString(call, body)
		return err
	})
	prepared := a.Tools().Prepare(context.Background(), ensemble.Part{Type: "tool_call", CallID: "set-later", Name: "tool_limits", Args: json.RawMessage(`{"max_output_bytes":37}`)}, "", 0)
	if got := c4Dispatch(t, a, "control_probe", `{}`, false); got != body {
		t.Fatalf("Prepare applied control before admission: %q", got)
	}
	if d := prepared.Await(context.Background()); d.IsError {
		t.Fatalf("control admission failed: %+v", d)
	}
	if got := c4Dispatch(t, a, "control_probe", `{}`, false); !strings.Contains(got, "max_output_bytes 37") {
		t.Fatalf("admitted literal-next control was lost: %q", got)
	}
}
