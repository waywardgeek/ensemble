package ensemble

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
)

// Rebase every represented reference in a genuine one-job snapshot. This is
// synthetic high-ID history, not a claim that the original real job ran at that
// ID. A global root may burn IDs belonging to other Agents; the represented
// job, matching text/locators and both exact maxima remain complete and public
// InspectCheckpoint/ImportSession must validate this candidate before use.
func ch10EdgeRebaseJob(t *testing.T, cp *common.Checkpoint, from, to uint64) {
	t.Helper()
	oldLocator, newLocator := fmt.Sprintf("cr/io/%d", from), fmt.Sprintf("cr/io/%d", to)
	job := func(j *common.JobSnapshot) {
		if j != nil && j.Handle == from {
			j.Handle, j.Output.Locator = to, newLocator
		}
	}
	var parts func([]common.Part)
	parts = func(values []common.Part) {
		for i := range values {
			p := &values[i]
			if p.Ref != nil && p.Ref.Locator == oldLocator {
				p.Ref.Locator = newLocator
			}
			if p.Text != nil {
				s := strings.ReplaceAll(*p.Text, oldLocator, newLocator)
				s = strings.ReplaceAll(s, fmt.Sprintf("job %d status:", from), fmt.Sprintf("job %d status:", to))
				p.Text = &s
			}
			parts(p.Parts)
		}
	}
	c := &cp.State.Context
	if len(c.Jobs) != 1 || c.Jobs[from].Handle != from {
		t.Fatal("rebase parent must have exactly one genuine job")
	}
	j := c.Jobs[from]
	job(&j)
	delete(c.Jobs, from)
	c.Jobs[to] = j
	refs := 0
	for id, state := range c.Calls {
		if state.JobHandle == from {
			state.JobHandle = to
			refs++
		}
		owned := []common.Part{state.Part}
		parts(owned)
		state.Part = owned[0]
		c.Calls[id] = state
	}
	if refs != 1 {
		t.Fatalf("one genuine called job reference required: %d", refs)
	}
	for _, entries := range [][]common.Entry{c.Entries, c.Instructions, c.Ephemera, c.Hints, c.PendingSkills, c.DeferredSkills} {
		for i := range entries {
			parts(entries[i].Parts)
		}
	}
	if c.Pending != nil {
		parts(c.Pending.Parts)
	}
	for i := range cp.State.Window.Events {
		e := &cp.State.Window.Events[i].Event
		job(e.Job)
		if e.Tool != nil {
			job(e.Tool.Job)
			parts(e.Tool.Parts)
		}
		if e.Message != nil {
			parts(e.Message.Parts)
		}
		if e.Response != nil {
			parts(e.Response.Parts)
		}
	}
	cp.HighWatermarks.Job, cp.State.Session.HighWatermarks.Job = to, to
}

func TestCh10EdgesJobUint64AndCollision(t *testing.T) {
	parentServer, parentCount := ch10EdgeServer(t, ch10EdgeCall{"run_command", `{"command":"printf parent-job"}`})
	parent := New(io.Discard)
	t.Cleanup(func() { _ = parent.Close() })
	for i := 0; i < 100; i++ {
		parent.AllocateHandle()
	}
	a, err := parent.OpenSession(ch10EdgeOptions(t, parentServer.URL, "run_command"))
	if err != nil {
		t.Fatal(err)
	}
	ch10EdgeTurn(t, a, true)
	x, err := a.ExportCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = parent.InspectCheckpoint(x.Bytes); err != nil {
		t.Fatalf("real101 parent refused: %v", err)
	}
	cp, err := a.Codec().Decode(x.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	ch10EdgeRebaseJob(t, &cp, 101, math.MaxUint64-1)
	high, err := a.Codec().Encode(cp)
	if err != nil {
		t.Fatal(err)
	}
	inspection, err := parent.InspectCheckpoint(high)
	if err != nil || inspection.Snapshot().Jobs[math.MaxUint64-1].Handle != math.MaxUint64-1 {
		t.Fatalf("complete synthetic high-ID parent refused: %v", err)
	}
	if parentCount.Load() != 2 {
		t.Fatal("inspection executed historical work")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	for _, occupied := range []bool{false, true} {
		t.Run(fmt.Sprint(occupied), func(t *testing.T) {
			server, count := ch10EdgeNamedServer(t, "maximum", ch10EdgeCall{"run_command", `{"command":"printf LAST_VALID > accepted-marker"}`}, ch10EdgeCall{"run_command", `{"command":"printf BAD > forbidden-marker"}`})
			r := New(io.Discard)
			t.Cleanup(func() { _ = r.Close() })
			o := ch10EdgeOptions(t, server.URL, "run_command")
			path := filepath.Join(o.Config.Workspace, "cr", "io", "18446744073709551615")
			canary := []byte("occupied final artifact must remain unchanged")
			if occupied {
				if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(path, canary, 0600); err != nil {
					t.Fatal(err)
				}
			}
			b, err := r.ImportSession(high, o)
			if err != nil {
				t.Fatalf("complete represented high job import refused: %v", err)
			}
			if count.Load() != 0 {
				t.Fatal("high-ID import performed work")
			}
			if !occupied {
				ch10EdgeTurn(t, b, true)
				if b.Snapshot().Jobs[math.MaxUint64].Status != "done" {
					t.Fatal("last uint64 job identity was not allocated exactly")
				}
				if marker, err := os.ReadFile(filepath.Join(o.Config.Workspace, "accepted-marker")); err != nil || string(marker) != "LAST_VALID" {
					t.Fatal("last valid job did not execute")
				}
				if _, err = b.Checkpoint(); err != nil {
					t.Fatalf("maximum durable job cannot checkpoint: %v", err)
				}
			}
			before := len(b.Snapshot().Jobs)
			c := ch10EdgeTurn(t, b, false)
			if c.Error == nil || c.Error.Message == "" {
				t.Fatal("exhaustion has no truthful failure")
			}
			if len(b.Snapshot().Jobs) != before {
				t.Fatal("exhaustion invented a job")
			}
			if r.AllocateHandle() != 0 || r.AllocateHandle() != 0 {
				t.Fatal("exhausted root wrapped or reused a handle")
			}
			for _, name := range []string{"forbidden-marker", "cr/io/0", "cr/io/1"} {
				if _, err := os.Lstat(filepath.Join(o.Config.Workspace, name)); !os.IsNotExist(err) {
					t.Fatalf("exhaustion created effect %s", name)
				}
			}
			if occupied {
				if _, err := os.Lstat(filepath.Join(o.Config.Workspace, "accepted-marker")); !os.IsNotExist(err) {
					t.Fatal("collision/exhaustion started handler")
				}
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, canary) {
					t.Fatal("occupied maximum artifact was changed")
				}
			}
			// Chapter4 §4.2 requires infrastructure creation failure to terminate
			// before dispatch, without a handler or subsequent model request.
			wantRequests := int32(3)
			if occupied {
				wantRequests = 1
			}
			if count.Load() != wantRequests {
				t.Fatal("job creation failure continued to another model request")
			}
			ch10RemainingCode(t, b.Close(), "session_limit")
		})
	}
}
