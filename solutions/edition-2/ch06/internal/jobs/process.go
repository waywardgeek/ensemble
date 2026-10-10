package jobs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"ensemble/internal/common"
	"github.com/creack/pty"
)

// A process and its terminal share the job lifetime. The reader owns natural
// completion; Kill changes the status first so late EOF cannot relabel it done.
type process struct {
	command  *exec.Cmd
	terminal *os.File
	done     chan struct{}
}

// StartProcess transfers completion to the PTY reader and rejects stopped jobs.
func (j *job) StartProcess(command *exec.Cmd) error {
	// Hold the state lock through launch: a kill that wins forbids launch,
	// and a kill that loses sees the actual process group it must stop.
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.data.Status != "running" || j.stopping {
		return errors.New("job is no longer running")
	}
	terminal, err := pty.StartWithSize(command, &pty.Winsize{Rows: 50, Cols: 200})
	if err != nil {
		return err
	} // Pipes cannot substitute for an interactive tty.
	j.process = &process{command: command, terminal: terminal, done: make(chan struct{})}
	// Only overrides belong in the per-call cwd record. An absent field
	// means the default, so a later bare call cannot inherit hidden state.
	if command.Dir != j.parent.Agent().Config().Workspace {
		j.data.Cwd = command.Dir
	}
	go j.capture(j.process)
	return nil
}
func (j *job) capture(p *process) {
	defer close(p.done)
	// Keep a trailing CR across reads: the terminal can split a CRLF pair
	// anywhere, and normalization must not depend on read-buffer boundaries.
	buffer := make([]byte, 32*1024)
	pending := ""
	for {
		n, err := p.terminal.Read(buffer)
		text := pending + string(buffer[:n])
		pending = ""
		if strings.HasSuffix(text, "\r") && err == nil {
			pending = "\r"
			text = strings.TrimSuffix(text, "\r")
		}
		j.mu.Lock()
		j.write(strings.ReplaceAll(text, "\r\n", "\n"))
		j.mu.Unlock()
		if err != nil {
			// Unix PTYs signal EOF as either EOF or EIO. Other I/O failures are
			// diagnostics, not invented command exit codes.
			if err != io.EOF && !errors.Is(err, syscall.EIO) {
				j.parent.Agent().Ensemble().Logf("job terminal read: %v", err)
			}
			break
		}
	}
	// Drain before reaping/closing, or the last output chunk can disappear.
	// ExitError is ordinary command data, not failure to invoke the tool.
	err := p.command.Wait()
	p.terminal.Close()
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			code = exit.ExitCode()
		}
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.data.Status != "running" {
		return
	}
	j.data.ExitCode = &code
	// The exit marker comes last because output arrives incrementally. Keeping
	// completion here prevents dispatch from closing the file under the reader.
	j.complete(pending+fmt.Sprintf("\nexit_code: %d\n", code), nil)
}

// Send writes a line to the running terminal without holding the job state lock
// during I/O.
func (j *job) Send(input string) error {
	j.mu.Lock()
	if j.data.Status != "running" || j.process == nil {
		j.mu.Unlock()
		return errors.New("job has no running input terminal")
	}
	terminal := j.process.terminal
	j.mu.Unlock()
	// Release the state lock before I/O: the reader must publish replies,
	// and a kill must remain able to reach the process while input is sent.
	// Input is a line by default. An explicit newline is already submitted;
	// retaining terminal echo makes the model's keystroke visible in the log.
	if !strings.HasSuffix(input, "\n") {
		input += "\n"
	}
	_, err := io.WriteString(terminal, input)
	return err
}

// Kill explicitly stops a process group and suppresses subsequent output for that
// job.
func (j *job) Kill(reason string) error {
	j.mu.Lock()
	if j.data.Status != "running" {
		j.mu.Unlock()
		return nil
	}
	if j.process != nil {
		// pty.Start creates a session whose process group is the shell's PID.
		// Killing only that PID leaves its children running with an open terminal.
		err := syscall.Kill(-j.process.command.Process.Pid, syscall.SIGKILL)
		if err != nil && !errors.Is(err, syscall.ESRCH) {
			j.mu.Unlock()
			return err
		}
	}
	// Killed is its own terminal state, not done with an exit code of -1.
	// Clear the optional exit value before notifying waiters or recording it.
	j.data.Status = "killed"
	j.data.ExitCode = nil
	j.file.Close()
	data := j.data
	j.notify()
	j.mu.Unlock()
	// Reap the killed shell before returning, without holding the state lock
	// needed by its reader. Waiters were already notified above.
	if j.process != nil {
		<-j.process.done
	}
	return j.parent.Agent().Record(common.Event{Type: "job_killed", Job: &common.JobKilledData{JobData: data, Reason: reason}})
}

// Closing process admission also covers a handler that has not launched yet.
// Without it, shutdown could scan an empty job and miss its later child process.
func (j *job) Shutdown() error {
	j.mu.Lock()
	j.stopping = true
	hasProcess := j.process != nil
	j.mu.Unlock()
	if hasProcess {
		return j.Kill("shutdown")
	}
	return nil
}
