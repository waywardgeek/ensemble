package ensemble_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

func managedCompletion(t *testing.T, done <-chan ensemble.ChildCompletion) ensemble.ChildCompletion {
	t.Helper()
	select {
	case out := <-done:
		return out
	case <-time.After(5 * time.Second):
		t.Fatal("managed completion did not drain")
		return ensemble.ChildCompletion{}
	}
}

// TestCompactionManagedLifetime exercises real launch, provider cancellation,
// reliable completion, failed persistence and exact removal without delegation.
func TestCompactionManagedLifetime(t *testing.T) {
	for _, mode := range []string{"success", "save-failure", "log-failure", "kill", "owner-shutdown"} {
		t.Run(mode, func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			var first sync.Once
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				first.Do(func() { close(entered) })
				select {
				case <-release:
					compactionProviderReply(w, "anthropic", false, false, `{"memory":"I kept the decision and why."}`)
				case <-r.Context().Done():
				}
			}))
			t.Cleanup(peer.Close)
			t.Cleanup(func() {
				select {
				case <-release:
				default:
					close(release)
				}
			})
			app := ensemble.New(io.Discard)
			parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: true})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { parent.Shutdown() })
			child, done, err := parent.Fleet().StartCompaction("worker", "SOURCE DATA: decisions and reasons")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { child.Shutdown() })
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("worker did not request")
			}
			if app.FleetEnabled() || !app.ContainsAgent(child) || len(parent.Fleet().Children()) != 1 {
				t.Fatal("managed launch lost visibility or enabled delegation")
			}
			if child.ParentAgent() != parent || len(child.Declarations()) != 1 {
				t.Fatal("managed construction bypassed ownership/restriction")
			}
			sameJSON(t, child.Declarations()[0].Schema, []byte(memorySubmitSchema))
			select {
			case <-done:
				t.Fatal("completion preceded provider result")
			default:
			}
			if mode == "save-failure" {
				if err := os.Mkdir(child.Config().SavePath+".bak", 0700); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "log-failure" {
				if err := os.Mkdir(child.Config().LogPath, 0700); err != nil {
					t.Fatal(err)
				}
			}
			var stopErr error
			switch mode {
			case "kill":
				stopErr = parent.Fleet().ShutdownChild("worker")
			case "owner-shutdown":
				stopErr = parent.Shutdown()
			default:
				close(release)
			}
			if stopErr != nil {
				t.Fatal(stopErr)
			}
			out := managedCompletion(t, done)
			if app.ContainsAgent(child) || len(parent.Fleet().Children()) != 0 {
				t.Fatal("completed child retained in root/fleet")
			}
			if !errors.Is(child.Post(common.Hint{Text: "late input"}), common.ErrStopped) {
				t.Fatal("dismissed child still admits actor input")
			}
			if _, err := parent.Fleet().Find("worker"); err == nil {
				t.Fatal("dismissed child still addressable")
			}
			app.ForgetCompaction(child) // Idempotent exact-identity removal.
			if !app.ContainsAgent(parent) {
				t.Fatal("cleanup removed owner")
			}
			if mode == "kill" || mode == "owner-shutdown" {
				if !out.Cancelled || out.Result.Err == nil {
					t.Fatalf("cancellation concealed: %+v", out)
				}
				return
			}
			if out.Cancelled {
				t.Fatal("automatic cleanup marked success canceled")
			}
			if len(out.Usage) != 1 || out.Usage[0].Usage != (ensemble.Usage{Input: 80, CacheRead: 20, Output: 9000}) {
				t.Fatalf("drained usage lost: %+v", out.Usage)
			}
			sameJSON(t, out.Submission.Payload, []byte(`{"memory":"I kept the decision and why."}`))
			if mode == "save-failure" {
				if out.Result.Err != nil || out.CleanupErr == nil {
					t.Fatalf("snapshot failure lost: %+v", out)
				}
				return
			}
			if mode == "log-failure" {
				if out.Result.Err == nil || out.CleanupErr == nil {
					t.Fatalf("required log save failure lost: %+v", out)
				}
				return
			}
			if out.Result.Err != nil || out.CleanupErr != nil {
				t.Fatal(errors.Join(out.Result.Err, out.CleanupErr))
			}
			data, err := os.ReadFile(child.Config().LogPath)
			if err != nil || !strings.Contains(string(data), `"type":"tool_returned"`) {
				t.Fatalf("submission tail missing from saved log: %s %v", data, err)
			}
			if _, _, err := parent.Fleet().StartCompaction("worker", "must not reload old task"); err == nil {
				t.Fatal("managed child reloaded an old task directory")
			}
		})
	}
}

// TestCompactionRepeatedRemoval confirms multiple fresh workers release owned
// references while preserving their real snapshots, submissions and measured work.
func TestCompactionRepeatedRemoval(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		compactionProviderReply(w, "anthropic", false, false, `{"memory":"I kept the reason."}`)
	}))
	t.Cleanup(peer.Close)
	app := ensemble.New(io.Discard)
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("worker-%d", i)
		child, done, err := parent.Fleet().StartCompaction(id, "source")
		if err != nil {
			t.Fatal(err)
		}
		out := managedCompletion(t, done)
		if out.Result.Err != nil || out.CleanupErr != nil {
			t.Fatal(errors.Join(out.Result.Err, out.CleanupErr))
		}
		if app.ContainsAgent(child) || len(parent.Fleet().Children()) != 0 {
			t.Fatal("repeated worker retained")
		}
		data, err := os.ReadFile(filepath.Join(child.Config().DataDir, "submit.md"))
		if err != nil || !strings.Contains(string(data), "I kept the reason.") {
			t.Fatalf("trace removed: %s %v", data, err)
		}
		data, err = os.ReadFile(child.Config().SavePath)
		var saved common.SaveFile
		if err != nil || json.Unmarshal(data, &saved) != nil || saved.Context == nil || saved.Context.Turn != "idle" {
			t.Fatalf("completed snapshot missing: %v", err)
		}
		finished := false
		for _, event := range saved.Log {
			finished = finished || event.Type == "job_finished"
		}
		if !finished {
			t.Fatal("snapshot preceded actual submit job completion")
		}
		jobs := 0
		for _, event := range child.History().EventsAfter(0) {
			if event.Type == "tool_called" && event.Tool != nil && event.Tool.Job != nil {
				jobs++
				job, err := child.Jobs().Find(event.Tool.Job.Handle)
				if err != nil || job.Data().Status == "running" {
					t.Fatalf("worker job not drained: %v", err)
				}
			}
		}
		if jobs != 1 {
			t.Fatalf("expected actual submit job, found %d", jobs)
		}
	}
}

// TestCompactionDirectoryAllocation proves root allocation follows a monotonic
// sequence rather than the live registry that shrinks after worker dismissal.
func TestCompactionDirectoryAllocation(t *testing.T) {
	t.Chdir(t.TempDir())
	app := ensemble.New(io.Discard)
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-ch16-course"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	child, err := app.NewChildAgent(parent, ensemble.Config{Model: "claude-ch16-course"}, ensemble.AgentSpec{ID: "worker", Compaction: true})
	if err != nil {
		t.Fatal(err)
	}
	old := child.Config().DataDir
	if err := child.Shutdown(); err != nil {
		t.Fatal(err)
	}
	app.ForgetCompaction(child)
	app.ForgetCompaction(parent)
	if !app.ContainsAgent(parent) {
		t.Fatal("managed cleanup removed top-level owner")
	}
	next, err := app.NewAgent(ensemble.Config{Model: "claude-ch16-course"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { next.Shutdown() })
	if next.Config().DataDir == old {
		t.Fatal("dismissal reused a prior Agent directory")
	}
}

// TestCompactionLaunchPersistenceFailure rejects a real initial snapshot error
// before model admission and releases the constructed child from the root.
func TestCompactionLaunchPersistenceFailure(t *testing.T) {
	app := ensemble.New(io.Discard)
	var constructed ensemble.Agent
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		a, err := app.NewChildAgent(parent, parent.Config(), spec)
		if err != nil {
			return nil, err
		}
		constructed = a
		if err := os.MkdirAll(a.Config().SavePath, 0700); err != nil {
			return nil, err
		}
		return a, nil
	})
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: "http://127.0.0.1:1", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	if _, _, err := parent.Fleet().StartCompaction("failure", "source"); err == nil {
		t.Fatal("initial snapshot failure admitted worker")
	}
	if constructed == nil || app.ContainsAgent(constructed) || len(parent.Fleet().Children()) != 0 {
		t.Fatal("failed construction retained child")
	}
	if constructed.Engine().Usage() != (ensemble.Usage{}) {
		t.Fatal("failed launch retained measured usage")
	}
	if err := constructed.Shutdown(); err != nil {
		t.Fatal(err)
	}
}

// TestCompactionReservedAddress keeps cleanup bound to the manager's reserved
// key even when a trusted custom factory gives the child another display label.
func TestCompactionReservedAddress(t *testing.T) {
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		compactionProviderReply(w, "anthropic", false, false, `{"memory":"I kept the reason."}`)
	}))
	t.Cleanup(peer.Close)
	app := ensemble.New(io.Discard)
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		spec.ID = "factory-label"
		return app.NewChildAgent(parent, parent.Config(), spec)
	})
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	child, done, err := parent.Fleet().StartCompaction("reserved-address", "source")
	if err != nil {
		t.Fatal(err)
	}
	out := managedCompletion(t, done)
	if out.Result.Err != nil || out.CleanupErr != nil {
		t.Fatal(errors.Join(out.Result.Err, out.CleanupErr))
	}
	if app.ContainsAgent(child) || len(parent.Fleet().Children()) != 0 {
		t.Fatal("custom label prevented exact entry removal")
	}
}

// TestCompactionFactoryPaths preserves existing files when a factory changes
// any explicit managed persistence path, and drains the rejected real child.
func TestCompactionFactoryPaths(t *testing.T) {
	for _, changed := range []string{"data", "save", "log"} {
		t.Run(changed, func(t *testing.T) {
			app := ensemble.New(io.Discard)
			old := t.TempDir()
			save, log := filepath.Join(old, "save.json"), filepath.Join(old, "events.jsonl")
			before := []byte("prior owner bytes\n")
			for _, path := range []string{save, log} {
				if err := os.WriteFile(path, before, 0600); err != nil {
					t.Fatal(err)
				}
			}
			var constructed ensemble.Agent
			app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
				switch changed {
				case "data":
					spec.DataDir = old
				case "save":
					spec.SavePath = save
				case "log":
					spec.LogPath = log
				}
				a, err := app.NewChildAgent(parent, parent.Config(), spec)
				constructed = a
				return a, err
			})
			parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: "http://127.0.0.1:1", DataDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { parent.Shutdown() })
			child, done, launchErr := parent.Fleet().StartCompaction("fresh", "source")
			if launchErr == nil {
				managedCompletion(t, done)
				child.Shutdown()
			}
			for _, path := range []string{save, log} {
				after, err := os.ReadFile(path)
				if err != nil || string(after) != string(before) {
					t.Fatalf("factory refusal changed %s: %q %v", changed, after, err)
				}
			}
			if launchErr == nil {
				t.Fatal("factory changed managed persistence paths")
			}
			if constructed == nil || app.ContainsAgent(constructed) || len(parent.Fleet().Children()) != 0 {
				t.Fatal("factory refusal retained worker")
			}
			if !errors.Is(constructed.Post(common.Hint{Text: "late"}), common.ErrStopped) {
				t.Fatal("rejected worker still admits input")
			}
		})
	}
}

// TestCompactionFactoryPurposeRemoval covers an ordinary request whose factory
// returns a managed child; refusal must release that owned child's root entry.
func TestCompactionFactoryPurposeRemoval(t *testing.T) {
	app := ensemble.New(io.Discard)
	var constructed ensemble.Agent
	app.SetAgentFactory(func(parent ensemble.Agent, spec ensemble.AgentSpec) (ensemble.Agent, error) {
		spec.Compaction = true
		a, err := app.NewChildAgent(parent, parent.Config(), spec)
		constructed = a
		return a, err
	})
	parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { parent.Shutdown() })
	if _, err := app.CreateChild(parent, common.AgentSpec{ID: "rejected", DataDir: filepath.Join(parent.Config().DataDir, "rejected")}); err == nil {
		t.Fatal("changed purpose accepted")
	}
	if constructed == nil || app.ContainsAgent(constructed) {
		t.Fatal("rejected managed child retained")
	}
	if !errors.Is(constructed.Post(common.Hint{Text: "late"}), common.ErrStopped) {
		t.Fatal("rejected managed child still admits input")
	}
}

type lifetimeObserver func(ensemble.Agent, ensemble.Observation)

// ObserveAgent forwards nonblocking progress to the test's external barriers.
func (f lifetimeObserver) ObserveAgent(a ensemble.Agent, o ensemble.Observation) { f(a, o) }

// TestCompactionResultKillRace holds the final real snapshot at a FIFO after
// successful Result/shutdown, then races explicit kill against that shared save.
func TestCompactionResultKillRace(t *testing.T) {
	for _, failSave := range []bool{false, true} {
		t.Run(fmt.Sprint(failSave), func(t *testing.T) {
			entered, release := make(chan struct{}), make(chan struct{})
			peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				close(entered)
				<-release
				compactionProviderReply(w, "anthropic", false, false, `{"memory":"I retained the actual completed decision."}`)
			}))
			t.Cleanup(peer.Close)
			app := ensemble.New(io.Discard)
			parent, err := app.NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: peer.URL, DataDir: t.TempDir(), DisableStreaming: true})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { parent.Shutdown() })
			idle := make(chan struct{}, 1)
			app.AddObserver(lifetimeObserver(func(a ensemble.Agent, o ensemble.Observation) {
				if a.ParentAgent() == parent {
					if state, ok := o.(ensemble.ActivityChanged); ok && state.Activity.Phase == common.ActivityIdle && a.Fleet().Status().State == "submitted" {
						select {
						case idle <- struct{}{}:
						default:
						}
					}
				}
			}))
			child, done, err := parent.Fleet().StartCompaction("race", "source")
			if err != nil {
				t.Fatal(err)
			}
			<-entered
			backup := child.Config().SavePath + ".bak"
			if err := syscall.Mkfifo(backup, 0600); err != nil {
				t.Fatal(err)
			}
			close(release)
			select {
			case <-idle:
			case <-time.After(5 * time.Second):
				t.Fatal("no successful idle boundary")
			}
			// finish sets active=nil before publishing idle, and sends Result before the
			// actor can process stop. Joining here therefore puts successful Result first.
			if err := child.Shutdown(); err != nil {
				t.Fatal(err)
			}
			select {
			case <-done:
				t.Fatal("completion preceded held final snapshot")
			default:
			}
			if failSave {
				journal := child.Config().SavePath + ".journal"
				if err := os.Remove(journal); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(journal, 0700); err != nil {
					t.Fatal(err)
				}
			}
			cancelRequested := make(chan struct{}, 1)
			app.AddObserver(lifetimeObserver(func(a ensemble.Agent, o ensemble.Observation) {
				if a == parent {
					if _, ok := o.(ensemble.FleetChanged); ok {
						select {
						case cancelRequested <- struct{}{}:
						default:
						}
					}
				}
			}))
			killed := make(chan error, 1)
			go func() { killed <- parent.Fleet().ShutdownChild("race") }()
			select {
			case <-cancelRequested:
			case <-time.After(5 * time.Second):
				t.Fatal("kill did not publish cancellation before drain")
			}
			select {
			case <-killed:
				t.Fatal("kill returned before shared snapshot")
			default:
			}
			// A real reader releases the final snapshot's blocked backup open/write.
			copied := make(chan error, 1)
			go func() { _, err := os.ReadFile(backup); copied <- err }()
			out := managedCompletion(t, done)
			if err := <-copied; err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-killed:
				if fmt.Sprint(err) != fmt.Sprint(out.CleanupErr) {
					t.Fatalf("kill lost shared save error: %v / %v", err, out.CleanupErr)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("kill did not share final save")
			}
			if out.Result.Err != nil || (out.CleanupErr != nil) != failSave || !out.Cancelled || len(out.Submission.Payload) == 0 || len(out.Usage) != 1 {
				t.Fatalf("successful Result/kill ordering lost: %+v", out)
			}
			if app.ContainsAgent(child) || len(parent.Fleet().Children()) != 0 {
				t.Fatal("racing kill retained worker")
			}
		})
	}
}
