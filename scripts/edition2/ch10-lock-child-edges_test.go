package ensemble

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func ch10EdgeEventually(t *testing.T, f func() bool, message string) {
	t.Helper()
	deadline := time.NewTimer(6 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if f() {
			return
		}
		select {
		case <-tick.C:
		case <-deadline.C:
			t.Fatal(message)
		}
	}
}

// The helper is an actual independent public consumer process. It runs only
// when the supervising test supplies an explicit scratch directory and server.
func TestCh10EdgesLockHolder(t *testing.T) {
	w := os.Getenv("ENSEMBLE_CH10_EDGE_HOLDER")
	if w == "" {
		t.Skip("subprocess helper, exercised by SurvivingChildLock")
	}
	r := New(io.Discard)
	defer r.Close()
	o := SessionOptions{Config: Config{Vendor: "openai", Model: "gpt-4.1-mini-2025-04-14", APIKey: "edge-noncredential", BaseURL: os.Getenv("ENSEMBLE_CH10_EDGE_URL"), DisableStreaming: true, Workspace: w, DataDir: filepath.Join(w, "session"), Builtins: []string{"run_command"}}}
	a, err := r.OpenSession(o)
	if err != nil {
		t.Fatal(err)
	}
	ch10EdgeTurn(t, a, true)
	if _, err = a.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(w, "holder-ready"), []byte("ready"), 0600); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestCh10EdgesSurvivingChildLock(t *testing.T) {
	w := t.TempDir()
	marker := fmt.Sprintf("ch10-edge-owned-%d", time.Now().UnixNano())
	args, err := json.Marshal(map[string]any{
		"command":             ": " + marker + `; trap '' HUP; echo $$ > edge-child.pid; printf EDGE_READY; i=0; while [ ! -f edge-child.release ]; do i=$((i+1)); printf '%s' "$i" > edge-child.heartbeat; sleep 0.02; done`,
		"ai_callback_pattern": "EDGE_READY", "ai_callback_delay": 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	server, count := ch10EdgeServer(t, ch10EdgeCall{"run_command", string(args)})
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "-test.run=^TestCh10EdgesLockHolder$", "-test.v")
	cmd.Env = append(os.Environ(), "ENSEMBLE_CH10_EDGE_HOLDER="+w, "ENSEMBLE_CH10_EDGE_URL="+server.URL)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pid := 0
	waited := false
	defer func() {
		_ = os.WriteFile(filepath.Join(w, "edge-child.release"), []byte("release"), 0600)
		// Also handle a fixture failure before holder-ready. Only signal the PID
		// written by this tool and still carrying this test's unique command token.
		if pid == 0 {
			b, _ := os.ReadFile(filepath.Join(w, "edge-child.pid"))
			pid, _ = strconv.Atoi(strings.TrimSpace(string(b)))
		}
		if pid > 1 && pid != os.Getpid() && pid != cmd.Process.Pid {
			command, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
			if err == nil && bytes.Contains(command, []byte(marker)) {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
		if !waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	if ledger := os.Getenv("ENSEMBLE_CH10_EDGE_LEDGER"); ledger != "" {
		record, err := json.Marshal(map[string]any{"workspace": w, "marker": marker, "holder_pid": cmd.Process.Pid, "binary": binary})
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(ledger, record, 0600); err != nil {
			t.Fatal(err)
		}
	}
	ch10EdgeEventually(t, func() bool { _, err := os.Stat(filepath.Join(w, "holder-ready")); return err == nil }, "independent holder did not finish the real paired running-job turn")
	rawPID, err := os.ReadFile(filepath.Join(w, "edge-child.pid"))
	if err != nil {
		t.Fatal(err)
	}
	pid, err = strconv.Atoi(strings.TrimSpace(string(rawPID)))
	if err != nil || pid <= 1 || pid == os.Getpid() {
		t.Fatal("controlled child PID missing")
	}
	if count.Load() != 2 {
		t.Fatal("positive child did not use two local requests")
	}
	command, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil || !bytes.Contains(command, []byte(marker)) {
		t.Fatal("live tool child command identity does not match this test")
	}
	lockPath := filepath.Join(w, "session", "owner.lock")
	lockBefore, err := os.Stat(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("holder_pid=%d child_pid=%d child_identity=%s lock_inode=%d", cmd.Process.Pid, pid, marker, lockBefore.Sys().(*syscall.Stat_t).Ino)
	r := New(io.Discard)
	t.Cleanup(func() { _ = r.Close() })
	o := SessionOptions{Config: Config{Vendor: "openai", Model: "gpt-4.1-mini-2025-04-14", APIKey: "edge-noncredential", BaseURL: server.URL, DisableStreaming: true, Workspace: w, DataDir: filepath.Join(w, "session"), Builtins: []string{"run_command"}}}
	_, err = r.OpenSession(o)
	ch10RemainingCode(t, err, "session_in_use")
	if err = cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	waited = true
	// A changed heartbeat after the owner has died proves the real tool child
	// still executes; kill(pid,0) alone could mistake a zombie for that proof.
	heartbeat := filepath.Join(w, "edge-child.heartbeat")
	before, err := os.ReadFile(heartbeat)
	if err != nil {
		t.Fatal(err)
	}
	ch10EdgeEventually(t, func() bool {
		after, err := os.ReadFile(heartbeat)
		return err == nil && len(after) != 0 && !bytes.Equal(before, after)
	}, "tool child did not survive owner death")
	afterHeartbeat, err := os.ReadFile(heartbeat)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("post-death live child heartbeat: before=%q after=%q", before, afterHeartbeat)
	a, err := r.OpenSession(o)
	if err != nil {
		t.Fatalf("surviving tool child retained store lock: %v; holder output: %s", err, output.String())
	}
	if len(a.Snapshot().Jobs) != 1 {
		t.Fatal("surviving OS child did not remain historical in resumed state")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	lockAfter, err := os.Stat(lockPath)
	if err != nil || !os.SameFile(lockBefore, lockAfter) {
		t.Fatal("owner death or reopen replaced lock inode")
	}
	t.Logf("reopened same lock_inode=%d while historical tool child remained alive", lockAfter.Sys().(*syscall.Stat_t).Ino)
	if count.Load() != 2 {
		t.Fatal("resume contacted model or restarted tool work")
	}
}
