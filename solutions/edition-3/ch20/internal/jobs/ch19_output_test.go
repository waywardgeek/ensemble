package jobs

import (
	"context"
	"ensemble/internal/common"
	"ensemble/internal/sandbox"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type ch19Agent struct {
	common.Agent
	cfg   common.Config
	files common.Sandbox
}

func (a *ch19Agent) Config() common.Config   { return a.cfg }
func (a *ch19Agent) Sandbox() common.Sandbox { return a.files }
func TestCh19RetainedOutputRooted(t *testing.T) {
	for _, method := range []string{"observe", "pattern", "complete"} {
		t.Run(method, func(t *testing.T) {
			dir := t.TempDir()
			a := &ch19Agent{cfg: common.Config{BaseDir: dir, SandboxRoot: dir}}
			var err error
			a.files, err = sandbox.New(a)
			if err != nil {
				t.Fatal(err)
			}
			defer a.files.Close()
			path := filepath.Join(dir, "output")
			f, err := a.files.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			f.WriteString("inside")
			f.Close()
			outside := filepath.Join(t.TempDir(), "private")
			os.WriteFile(outside, []byte("OUTSIDE"), 0600)
			os.Remove(path)
			if err = os.Symlink(outside, path); err != nil {
				t.Fatal(err)
			}
			j := &job{parent: &collection{parent: a}, data: common.JobData{Status: common.JobRunning, Bytes: 7, Output: common.Ref{Locator: path}}, settled: make(chan struct{}), changed: make(chan struct{})}
			var report common.JobReport
			switch method {
			case "observe":
				report, err = j.observeLocked(common.ToolLimits{MaxOutputBytes: 100})
			case "pattern":
				report, err = j.wait(context.Background(), common.ToolLimits{Delay: time.Second, MaxOutputBytes: 100, Pattern: "OUTSIDE"}, true)
			case "complete":
				close(j.settled)
				report, err = j.Complete(context.Background())
			}
			if err == nil || report.Text != "" {
				t.Fatalf("outside output disclosed: %+v, %v", report, err)
			}
		})
	}
}
