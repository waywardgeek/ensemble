package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"ensemble/internal/common"
	"github.com/creack/pty"
)

func commandDefinition() common.ToolDefinition {
	return common.ToolDefinition{Name: "run_command", Description: "Start a fresh /bin/sh -c in a PTY with merged output and exit_code. Optional cwd applies only to this command, relative to the Agent base directory. Provider credentials are removed. ai_callback_delay defaults to 3 seconds, ai_callback_pattern watches unseen output, and max_output_bytes defaults to 16384; these bound observation, never execution. Running reports include a handle for wait_for_job/send_input/kill_job. Full output stays on disk.", Parameters: json.RawMessage(`{"type":"object","properties":{"command":{"type":"string"},"cwd":{"type":"string"},"ai_callback_delay":{"type":"number","minimum":0},"ai_callback_pattern":{"type":"string"},"max_output_bytes":{"type":"integer","minimum":1}},"required":["command"],"additionalProperties":false}`)}
}
func (r *registry) run(ctx context.Context, owner common.ToolCall, raw json.RawMessage) error {
	a, err := decode(raw)
	if err != nil {
		return err
	}
	if err = a.only("command", "cwd", "ai_callback_delay", "ai_callback_pattern", "max_output_bytes"); err != nil {
		return err
	}
	command, err := a.text("command", true)
	if err != nil {
		return err
	}
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("command must not be empty")
	}
	cwd, err := a.text("cwd", false)
	if err != nil {
		return err
	}
	if _, present := a["cwd"]; present && cwd == "" {
		return fmt.Errorf("cwd must not be empty")
	}
	cmd := exec.Command("/bin/sh", "-c", command)
	// Confined preparation needs the supplied name for refusals. Unconfined
	// commands retain their existing cwd expansion and absolute-path semantics.
	cmd.Dir = cwd
	if r.parent.Config().SandboxRoot == "" {
		cmd.Dir = r.path(cwd)
	}
	if err := r.parent.Sandbox().ValidateCommand(command); err != nil {
		return err
	}
	if err := r.parent.Sandbox().PrepareCommand(cmd); err != nil {
		return err
	}
	var terminal *os.File
	var waitErr, errorRead error
	reaped := make(chan struct{})
	drained := make(chan struct{})
	var stopOnce sync.Once
	stop := func() error {
		stopOnce.Do(func() {
			// pty.Start creates a session/process-group leader on Unix.
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-reaped
			owner.SetExitCode(cmd.ProcessState.ExitCode())
			_ = terminal.Close()
			<-drained
		})
		return nil
	}
	err = owner.StartProcess(func() (io.Writer, func() error, error) {
		var err error
		terminal, err = pty.StartWithSize(cmd, &pty.Winsize{Rows: 50, Cols: 200})
		if err != nil {
			return nil, nil, fmt.Errorf("starting command: %w", err)
		}
		go func() { errorRead = copyPTY(owner, terminal); close(drained) }()
		go func() { waitErr = cmd.Wait(); close(reaped) }()
		return terminal, stop, nil
	})
	if err != nil {
		return err
	}
	defer terminal.Close()
	select {
	case <-reaped:
		// A background descendant must not retain the PTY after the shell has exited.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-drained
	case <-ctx.Done():
		_ = stop()
	}
	code := cmd.ProcessState.ExitCode()
	owner.SetExitCode(code)
	if waitErr != nil {
		if _, ok := waitErr.(*exec.ExitError); !ok {
			return fmt.Errorf("waiting for command: %w", waitErr)
		}
	}
	// PTY EOF commonly arrives as EIO on Unix, rather than io.EOF.
	if errorRead != nil && !os.IsNotExist(errorRead) && !strings.Contains(errorRead.Error(), "input/output error") && !strings.Contains(errorRead.Error(), "file already closed") {
		return fmt.Errorf("reading PTY: %w", errorRead)
	}
	_, err = fmt.Fprintf(owner, "\nexit_code: %d\n", code)
	return err
}

// copyPTY normalizes CRLF before bytes reach the job's file or cursor. A CR at
// a read boundary is held until its successor (or EOF), preserving lone CRs and
// every non-CR byte, including split UTF-8. Memory is bounded by the read chunk.
func copyPTY(dst io.Writer, src io.Reader) error {
	raw := make([]byte, 32*1024)
	normalized := make([]byte, 0, len(raw)+1)
	pendingCR := false
	for {
		n, readErr := src.Read(raw)
		normalized = normalized[:0]
		for _, b := range raw[:n] {
			if pendingCR {
				if b != '\n' {
					normalized = append(normalized, '\r')
				}
				pendingCR = false
			}
			if b == '\r' {
				pendingCR = true
			} else {
				normalized = append(normalized, b)
			}
		}
		if readErr != nil && pendingCR {
			normalized = append(normalized, '\r')
			pendingCR = false
		}
		if len(normalized) > 0 {
			written, err := dst.Write(normalized)
			if err != nil {
				return err
			}
			if written != len(normalized) {
				return io.ErrShortWrite
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}
