package jobs

import (
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func TestStopProcessAcrossAttachment(t *testing.T) {
	for _, before := range []bool{false, true} {
		name := "attached"
		if before {
			name = "late_attachment"
		}
		t.Run(name, func(t *testing.T) {
			j := &Job{status: common.StatusRunning, changed: make(chan struct{})}
			cmd := exec.Command("/bin/sh", "-c", "exec sleep 60")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer cmd.Process.Kill()
			if before {
				j.StopProcess("shutdown")
			}
			j.Attach(cmd.Process, nil)
			if !before {
				j.StopProcess("shutdown")
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case err := <-done:
				if err == nil {
					t.Fatal("process was not killed")
				}
			case <-time.After(2 * time.Second):
				t.Fatal("shutdown let a process escape")
			}
			if j.Status() != common.StatusKilled {
				t.Fatalf("status = %s", j.Status())
			}
		})
	}
}

func TestStopProcessPreservesGoResultButKillDoesNot(t *testing.T) {
	j := &Job{status: common.StatusRunning, changed: make(chan struct{})}
	j.StopProcess("shutdown")
	j.Finish("actual output", nil)
	if j.Status() != common.StatusDone || j.Report(common.WokeDone, common.DefaultLimits()) != "actual output" {
		t.Fatal("process-only shutdown discarded the Go tool's result")
	}

	j = &Job{status: common.StatusRunning, changed: make(chan struct{})}
	j.Kill("user requested")
	j.Finish("late contradictory output", nil)
	if j.Status() != common.StatusKilled || string(j.out.Bytes()) != "\n[job 0 killed: user requested]\n" {
		t.Fatal("explicit kill no longer suppresses late output")
	}
}
