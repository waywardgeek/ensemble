package jobs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/creack/pty"
)

func (j *job) StartProcess(command, cwd string) error {
	s := j.service()
	s.mu.Lock()
	defer s.mu.Unlock()
	if j.snapshot.Status != "running" || s.closed || s.fault != nil {
		return fmt.Errorf("job stopped before process startup")
	}
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = cwd
	if cwd == "" {
		cmd.Dir = s.parent.Workspace()
	}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "TERM=") {
			cmd.Env = append(cmd.Env, value)
		}
	}
	cmd.Env = append(cmd.Env, "TERM=dumb")
	terminal, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 50, Cols: 200})
	if err != nil {
		return fmt.Errorf("start PTY: %w", err)
	}
	j.terminal = terminal
	j.pid = cmd.Process.Pid
	j.snapshot.Cwd = cwd
	// There is no timer-driven Close: EOF/EIO after the slave closes drains all
	// queued bytes. A process-group kill closes inherited slave descriptors too.
	drained := make(chan struct{})
	go func() { defer close(drained); j.readProcess(terminal) }()
	go func() {
		waitErr := cmd.Wait()
		<-drained
		_ = terminal.Close()
		s.mu.Lock()
		defer s.mu.Unlock()
		defer close(j.reaped)
		j.terminal = nil
		if j.killing {
			s.finish(j, "killed", j.killReason)
			return
		}
		code := cmd.ProcessState.ExitCode()
		var exit *exec.ExitError
		if waitErr != nil && !errors.As(waitErr, &exit) {
			j.snapshot.IsError = true
			s.write(j, []byte(fmt.Sprintf("process wait failed: %v\n", waitErr)))
		}
		// Status text is produced only by a normal reaping winner, after reader join.
		prefix := ""
		if j.snapshot.Bytes > 0 {
			last := []byte{0}
			_, _ = j.file.ReadAt(last, j.snapshot.Bytes-1)
			if last[0] != '\n' {
				prefix = "\n"
			}
		}
		s.write(j, []byte(fmt.Sprintf("%sexit_code: %d\n", prefix, code)))
		j.snapshot.ExitCode = &code
		s.finish(j, "done", "")
	}()
	return nil
}
func (j *job) readProcess(reader io.Reader) {
	s := j.service()
	buffer := make([]byte, 32768)
	pendingCR := false
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			data := buffer[:n]
			// Hold one trailing CR to recognize CRLF even across read boundaries.
			if pendingCR {
				data = append([]byte{'\r'}, data...)
				pendingCR = false
			}
			if data[len(data)-1] == '\r' {
				pendingCR = true
				data = data[:len(data)-1]
			}
			data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
			s.mu.Lock()
			s.write(j, data)
			s.mu.Unlock()
		}
		if err != nil {
			s.mu.Lock()
			if pendingCR {
				s.write(j, []byte{'\r'})
			}
			// Linux PTYs signal slave closure as EIO; Darwin normally returns EOF.
			if !errors.Is(err, io.EOF) && !errors.Is(err, syscall.EIO) {
				s.fail(fmt.Errorf("PTY read failed: %w", err))
			}
			s.mu.Unlock()
			return
		}
	}
}
