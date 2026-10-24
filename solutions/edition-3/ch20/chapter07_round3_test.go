package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
	agentowner "ensemble/internal/agent"
	"ensemble/internal/common"
	engineowner "ensemble/internal/engine"
	historyowner "ensemble/internal/history"
	jobsowner "ensemble/internal/jobs"
	skillsowner "ensemble/internal/skills"
	toolowner "ensemble/internal/tool"
)

type ch7SavedBoundary struct {
	raw              []byte
	readErr, saveErr error
}

// This wrapper delegates the actual Agent save callback. Failures occur inside
// the serializer passed to that callback, or at a redundant ordinary Dump;
// neither is the earlier pre-callback failure used by the R2 gate fixture.
type ch7ActualSaveHistory struct {
	common.History
	path, mode                  string
	saved                       chan ch7SavedBoundary
	release                     chan struct{}
	serializers, duplicateSaves atomic.Int32
	committed                   atomic.Bool
}

func (h *ch7ActualSaveHistory) AppendSaved(e common.Event, save func(func(io.Writer) error) error) error {
	err := h.History.AppendSaved(e, func(dump func(io.Writer) error) error {
		err := save(func(w io.Writer) error {
			h.serializers.Add(1)
			if err := dump(w); err != nil {
				return err
			}
			if h.mode == "serializer_failure" {
				return errors.New("controlled staged serializer failure")
			}
			return nil
		})
		raw, readErr := os.ReadFile(h.path)
		h.saved <- ch7SavedBoundary{raw: raw, readErr: readErr, saveErr: err}
		<-h.release
		return err
	})
	if err == nil {
		h.committed.Store(true)
	}
	return err
}
func (h *ch7ActualSaveHistory) Dump(w io.Writer) error {
	if h.committed.Load() {
		last := h.EventsRange(h.LastSeq()-1, h.LastSeq())
		if len(last) == 1 && last[0].Type == ensemble.ToolReleased {
			h.duplicateSaves.Add(1)
			if h.mode == "duplicate_save_failure" {
				return errors.New("redundant post-commit save must not run")
			}
		}
	}
	return h.History.Dump(w)
}
func ch7ReadSavedEvents(parent common.Agent, raw []byte) ([]ensemble.Event, error) {
	h := historyowner.New(parent)
	if err := h.Load(bytes.NewReader(raw)); err != nil {
		return nil, err
	}
	return h.Events(), nil
}
func ch7HasDispatch(events []ensemble.Event) bool {
	for _, e := range events {
		if e.Type == ensemble.ToolReleased || (e.Tool != nil && e.Tool.Dispatched) {
			return true
		}
	}
	return false
}
func TestCh7R3ActualAdmissionSave(t *testing.T) {
	for _, mode := range []string{"inspect_saved_boundary", "serializer_failure", "duplicate_save_failure"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "agent.jsonl")
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := agentowner.New(app, ensemble.Config{Vendor: ensemble.OpenAI, Model: "offline", APIKey: "fake", BaseURL: "http://127.0.0.1:1", BaseDir: dir, DataDir: filepath.Join(dir, "jobs"), LogPath: path})
			if err != nil {
				t.Fatal(err)
			}
			h := &ch7ActualSaveHistory{History: historyowner.New(a), path: path, mode: mode, saved: make(chan ch7SavedBoundary, 1), release: make(chan struct{})}
			a.SetHistory(h)
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
				case <-h.release:
				default:
					close(h.release)
				}
				_ = a.Shutdown()
			}()
			q := a.Observe()
			defer q.Close()
			var effects atomic.Int32
			handlerChecks := make(chan error, 1)
			err = a.Tools().Register(ensemble.ToolDefinition{Name: "saved_effect", Description: "Actual admission-save probe", Parameters: json.RawMessage(`{"type":"object"}`)}, func(context.Context, ensemble.ToolCall, json.RawMessage) error {
				raw, err := os.ReadFile(path)
				if err == nil {
					var saved []ensemble.Event
					saved, err = ch7ReadSavedEvents(a, raw)
					if err == nil && !ch7HasDispatch(saved) {
						err = errors.New("handler preceded dispatched fact in actual saved file")
					}
				}
				if !ch7HasDispatch(a.Log().EventsRange(0, a.View().Through)) {
					err = errors.Join(err, errors.New("handler preceded committed replay frontier"))
				}
				queued := false
				for _, o := range q.Drain() {
					if _, ok := o.(ensemble.ToolDispatched); ok {
						queued = true
					}
				}
				if !queued {
					err = errors.Join(err, errors.New("handler preceded queued dispatch observation"))
				}
				handlerChecks <- err
				effects.Add(1)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			result := make(chan error, 1)
			go func() {
				result <- a.Tools().Invoke(context.Background(), ensemble.Part{Type: "tool_call", CallID: "saved-effect", Name: "saved_effect", Args: json.RawMessage(`{}`)}, "")
			}()
			var boundary ch7SavedBoundary
			select {
			case boundary = <-h.saved:
			case <-time.After(3 * time.Second):
				t.Fatal("actual admission save callback did not complete")
			}
			if boundary.readErr != nil {
				t.Fatal(boundary.readErr)
			}
			saved, err := ch7ReadSavedEvents(a, boundary.raw)
			if err != nil {
				t.Fatal(err)
			}
			if effects.Load() != 0 {
				t.Error("effect ran before admission callback returned")
			}
			if ch7HasDispatch(a.Log().Events()) || ch7HasDispatch(a.Log().EventsRange(0, a.View().Through)) {
				t.Error("staged boundary was installed/published before its save callback returned")
			}
			if h.serializers.Load() != 1 {
				t.Errorf("actual save invoked staged serializer %d times, want 1", h.serializers.Load())
			}
			wantFailure := mode == "serializer_failure"
			if wantFailure {
				if boundary.saveErr == nil {
					t.Error("production save ignored staged serializer failure")
				}
				if ch7HasDispatch(saved) {
					t.Error("failed serializer replaced actual saved file with dispatch")
				}
			} else {
				if boundary.saveErr != nil {
					t.Error(boundary.saveErr)
				}
				if !ch7HasDispatch(saved) {
					t.Error("successful admission callback saved old History, not staged ToolReleased")
				}
			}
			close(h.release)
			select {
			case err = <-result:
			case <-time.After(3 * time.Second):
				t.Fatal("admission did not settle")
			}
			if wantFailure {
				if err == nil {
					t.Error("failed production save admitted the handler")
				}
				if effects.Load() != 0 {
					t.Errorf("failed production save executed handler %d times", effects.Load())
				}
				if effects.Load() == 0 {
					for _, o := range q.Drain() {
						if _, ok := o.(ensemble.ToolDispatched); ok {
							t.Error("failed save queued dispatch")
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("admitted handler aborted after successful boundary save: %v", err)
				}
				if effects.Load() != 1 {
					t.Errorf("successful boundary executed handler %d times, want 1", effects.Load())
				} else {
					if err := <-handlerChecks; err != nil {
						t.Error(err)
					}
				}
			}
			if h.duplicateSaves.Load() != 0 {
				t.Errorf("redundant post-commit save count=%d, want 0", h.duplicateSaves.Load())
			}
			// Force an ordinary subsequent actor save. It must not reveal a dispatch
			// from the failed transaction through memory, disk, or GUI's replay range.
			before := a.View().Through
			if err := a.Post(ensemble.Ephemeral{Text: "later successful save"}); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(3 * time.Second)
			for a.View().Through <= before && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			if a.View().Through <= before {
				t.Fatal("later ordinary save did not publish")
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			saved, err = ch7ReadSavedEvents(a, raw)
			if err != nil {
				t.Fatal(err)
			}
			sets := []struct {
				name   string
				events []ensemble.Event
			}{{"memory", a.Log().Events()}, {"saved file", saved}, {"replay frontier", a.Log().EventsRange(0, a.View().Through)}}
			for _, set := range sets {
				if ch7HasDispatch(set.events) == wantFailure {
					t.Errorf("%s dispatch truth wrong after later successful save (failure=%t)", set.name, wantFailure)
				}
			}
			if effects.Load() == 0 && ch7HasDispatch(saved) {
				t.Error(fmt.Sprintf("saved dispatch exists although no handler effect was admitted (%s)", mode))
			}
		})
	}
}
