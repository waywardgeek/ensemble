package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// StdioTransport spawns a subprocess and talks newline-delimited JSON-RPC
// over its stdin/stdout. This is the standard MCP transport for local
// subprocess servers (Python, Go, etc.).
//
// The subprocess is a managed process — bounded shutdown like Job.awaitReaped.
type StdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Scanner
	mu     sync.Mutex
	closed bool
}

// NewStdioTransport starts a subprocess and returns a transport connected
// to its stdin/stdout. The process starts immediately.
func NewStdioTransport(command string, args []string, env []string) (*StdioTransport, error) {
	cmd := exec.Command(command, args...)
	if len(env) > 0 {
		cmd.Env = env
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp stdio: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("mcp stdio: stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("mcp stdio: start %s: %w", command, err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	return &StdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		reader: scanner,
	}, nil
}

func (s *StdioTransport) Send(msg json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return fmt.Errorf("transport closed")
	}
	_, err := fmt.Fprintf(s.stdin, "%s\n", msg)
	return err
}

func (s *StdioTransport) Recv() (json.RawMessage, error) {
	if s.reader.Scan() {
		line := s.reader.Bytes()
		cp := make([]byte, len(line))
		copy(cp, line)
		return json.RawMessage(cp), nil
	}
	if err := s.reader.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

// Close sends EOF on stdin, then waits for the process to exit with a
// bounded timeout — same discipline as Job.awaitReaped.
func (s *StdioTransport) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()

	// Close stdin to signal EOF to the subprocess.
	s.stdin.Close()

	// Wait with a bounded timeout.
	done := make(chan error, 1)
	go func() { done <- s.cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		// Process did not exit in time — kill it.
		s.cmd.Process.Kill()
		<-done
	}
	return nil
}
