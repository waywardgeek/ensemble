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
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCh7R1SlowIdentifiedSubscriber(t *testing.T) {
	app := ensemble.New(nil)
	entered, release := make(chan struct{}, 1), make(chan struct{})
	off := app.SubscribeIdentified(func(ensemble.AgentObservation) {
		select {
		case entered <- struct{}{}:
		default:
		}
		<-release
	})
	defer func() { close(release); off(); app.Shutdown() }()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", PauseOnStart: true})
	if err != nil {
		t.Fatal(err)
	}
	result := a.Submit(context.Background(), "wait at gate")
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("subscriber not reached")
	}
	a.Post(ensemble.Interrupt{})
	select {
	case r := <-result:
		if r.Err == nil {
			t.Fatal("interrupt did not cancel turn")
		}
	case <-time.After(time.Second):
		t.Fatal("slow identified subscriber blocked Agent interrupt")
	}
}

type ch7SaveGate struct {
	common.History
	entered, release chan struct{}
	once             sync.Once
	fail             bool
}

func (h *ch7SaveGate) Dump(w io.Writer) error {
	blocked := false
	for _, e := range h.History.Events() {
		if e.Type == ensemble.ToolCalled {
			h.once.Do(func() { blocked = true; close(h.entered); <-h.release })
			break
		}
	}
	if blocked && h.fail {
		return errors.New("controlled one-shot save failure")
	}
	return h.History.Dump(w)
}
func TestCh7R1SavedAdmissionFrontier(t *testing.T) {
	for _, tc := range []struct{ fail, kill bool }{{}, {fail: true}, {kill: true}} {
		fail := tc.fail
		t.Run(fmt.Sprintf("failure=%t/kill-before-release=%t", fail, tc.kill), func(t *testing.T) {
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := agentowner.New(app, ensemble.Config{Model: "offline", BaseDir: t.TempDir(), LogPath: filepath.Join(t.TempDir(), "history.jsonl")})
			if err != nil {
				t.Fatal(err)
			}
			gate := &ch7SaveGate{History: historyowner.New(a), entered: make(chan struct{}), release: make(chan struct{}), fail: fail}
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
			var effects atomic.Int32
			if err := a.Tools().Register(ensemble.ToolDefinition{Name: "effect", Description: "Controlled effect", Parameters: json.RawMessage(`{"type":"object"}`)}, func(context.Context, ensemble.ToolCall, json.RawMessage) error { effects.Add(1); return nil }); err != nil {
				t.Fatal(err)
			}
			result := make(chan error, 1)
			go func() {
				result <- a.Tools().Invoke(context.Background(), ensemble.Part{Type: "tool_call", CallID: "request", Name: "effect", Args: json.RawMessage(`{}`)}, "")
			}()
			select {
			case <-gate.entered:
			case <-time.After(time.Second):
				t.Fatal("save not entered")
			}
			snapshot := a.View()
			if snapshot.Through != 0 || len(a.Log().EventsRange(0, snapshot.Through)) != 0 {
				t.Fatal("pending save became published")
			}
			if effects.Load() != 0 {
				t.Fatal("effect crossed pending intent save")
			}
			var killed chan error
			if tc.kill {
				for _, e := range gate.Events() {
					if e.Type == ensemble.ToolCalled && e.Tool.Job != nil {
						job, _ := app.LookupJob(e.Tool.Job.Handle)
						killed = make(chan error, 1)
						go func() { killed <- job.Kill("kill_job") }()
						deadline := time.Now().Add(time.Second)
						for job.Data().Status != ensemble.JobKilledStatus {
							if time.Now().After(deadline) {
								t.Fatal("reserved job not killed")
							}
							time.Sleep(time.Millisecond)
						}
					}
				}
			}
			close(gate.release)
			select {
			case err := <-result:
				if (err != nil) != fail {
					t.Fatalf("result=%v failure=%t", err, fail)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("invocation did not finish")
			}
			if killed != nil {
				select {
				case err := <-killed:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(time.Second):
					t.Fatal("kill fact not delivered")
				}
			}
			visible := a.Log().EventsRange(0, a.View().Through)
			dispatches := 0
			called := 0
			for _, e := range visible {
				if e.Type == ensemble.ToolCalled {
					called++
					if e.Tool.Dispatched {
						t.Fatal("intent claimed dispatch")
					}
				}
				if e.Type == ensemble.ToolReleased {
					dispatches++
				}
			}
			if called != 1 {
				t.Fatalf("called=%d", called)
			}
			if fail || tc.kill {
				if effects.Load() != 0 || dispatches != 0 {
					t.Fatal("failed saved intent became dispatch after later successful save")
				}
			} else if effects.Load() != 1 || dispatches != 1 {
				t.Fatalf("effect=%d dispatches=%d", effects.Load(), dispatches)
			}
		})
	}
}

type ch7BlockedTrace struct {
	entered, release chan struct{}
	once             sync.Once
}

func (w *ch7BlockedTrace) Write(p []byte) (int, error) {
	w.once.Do(func() { close(w.entered) })
	<-w.release
	return len(p), nil
}
func TestCh7R1BlockedTraceDoesNotBlockAgent(t *testing.T) {
	app := ensemble.New(nil)
	w := &ch7BlockedTrace{entered: make(chan struct{}), release: make(chan struct{})}
	app.SetTrace(w)
	defer func() { close(w.release); app.Shutdown() }()
	app.Trace(ensemble.TraceRecord{Stage: "block"})
	select {
	case <-w.entered:
	case <-time.After(time.Second):
		t.Fatal("trace not entered")
	}
	a, err := app.NewAgent(ensemble.Config{Model: "offline", PauseOnStart: true})
	if err != nil {
		t.Fatal(err)
	}
	result := a.Submit(context.Background(), "wait")
	a.Post(ensemble.Interrupt{})
	select {
	case <-result:
	case <-time.After(time.Second):
		t.Fatal("trace writer blocked Agent")
	}
}
