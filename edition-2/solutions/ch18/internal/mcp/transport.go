// Package mcp owns external tool connections without importing any GUI code.
package mcp

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"ensemble/internal/common"
)

// RawTransport adapts caller-owned streams to the exercise's newline JSON wire.
// The closer owns interruption of both streams; it must unblock blocked I/O.
type RawTransport struct {
	decoder *json.Decoder
	encoder *json.Encoder
	closer  io.Closer
	mu      sync.Mutex
}

// NewRaw wraps a boundary resource, rather than creating another domain owner.
func NewRaw(reader io.Reader, writer io.Writer, closer io.Closer) *RawTransport {
	if reader == nil || writer == nil || closer == nil {
		panic("MCP raw transport requires streams and closer")
	}
	return &RawTransport{decoder: json.NewDecoder(reader), encoder: json.NewEncoder(writer), closer: closer}
}

// Send serializes whole messages so reverse replies cannot split an outgoing request.
func (t *RawTransport) Send(message json.RawMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.encoder.Encode(message)
}

// Recv decodes one message; the protocol's sole reader owns this decoder.
func (t *RawTransport) Recv() (json.RawMessage, error) {
	var message json.RawMessage
	err := t.decoder.Decode(&message)
	return message, err
}

// Close interrupts reads and writes using the boundary's own resource owner.
func (t *RawTransport) Close() error { return t.closer.Close() }

// Pipe creates connected in-process endpoints exercising the same wire codec.
func Pipe() (common.MCPTransport, common.MCPTransport) {
	a, b := net.Pipe()
	return NewRaw(a, a, a), NewRaw(b, b, b)
}

// A stdio peer has a real process lifetime. Closing kills its group and reaps it,
// preventing descendants from retaining pipes after skill unload or shutdown.
type stdio struct {
	*RawTransport
	command *exec.Cmd
	input   io.WriteCloser
	output  io.ReadCloser
	once    sync.Once
}

func startStdio(owner common.Agent, server common.MCPServer) (common.MCPTransport, error) {
	command := exec.Command(server.Command, server.Args...)
	command.Dir = owner.Config().Workspace
	// Do not pass model credentials into unrelated tool subprocesses. Ordinary
	// executable lookup and OS identity still use the process environment.
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "LANG"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	input, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		input.Close()
		return nil, err
	}
	command.Stderr = diagnosticWriter{owner}
	if err := command.Start(); err != nil {
		input.Close()
		output.Close()
		return nil, err
	}
	t := &stdio{command: command, input: input, output: output}
	t.RawTransport = NewRaw(output, input, t)
	return t, nil
}

// Close is idempotent because EOF and explicit unload may race to release a peer.
func (t *stdio) Close() error {
	t.once.Do(func() {
		t.input.Close()
		t.output.Close()
		syscall.Kill(-t.command.Process.Pid, syscall.SIGKILL)
		t.command.Wait()
	})
	return nil
}

type diagnosticWriter struct{ owner common.Agent }

// Write routes server stderr through the owning Ensemble's serialized logger.
func (w diagnosticWriter) Write(data []byte) (int, error) {
	w.owner.Ensemble().Logf("MCP server: %s", data)
	return len(data), nil
}

var errClosed = errors.New("MCP connection closed")
