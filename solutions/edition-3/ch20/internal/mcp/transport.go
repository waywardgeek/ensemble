package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os/exec"
	"sync"
	"syscall"

	"ensemble/internal/common"
)

type stream struct {
	rw     io.ReadWriteCloser
	reader *bufio.Reader
	write  sync.Mutex
	once   sync.Once
	err    error
}

func NewStream(rw io.ReadWriteCloser) common.MCPTransport {
	return &stream{rw: rw, reader: bufio.NewReader(rw)}
}
func (s *stream) Recv() (json.RawMessage, error) {
	b, err := s.reader.ReadBytes('\n')
	if len(b) > 0 && err == io.EOF {
		err = nil
	}
	return b, err
}
func (s *stream) Send(b json.RawMessage) error {
	s.write.Lock()
	defer s.write.Unlock()
	_, err := s.rw.Write(append(append([]byte(nil), b...), '\n'))
	return err
}

func (s *stream) Close() error { s.once.Do(func() { s.err = s.rw.Close() }); return s.err }

type processPipe struct {
	io.Reader
	io.Writer
	stdin  io.Closer
	stdout io.Closer
	cmd    *exec.Cmd
	once   sync.Once
	done   chan struct{}
}

func (p *processPipe) Close() error {
	p.once.Do(func() {
		_ = p.stdin.Close()
		_ = p.stdout.Close()
		_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
	})
	<-p.done
	return nil
}
func dial(ctx context.Context, c common.MCPServer, base string, diagnostics io.Writer, environment ...string) (common.MCPTransport, error) {
	return dialPrepared(ctx, c, base, diagnostics, nil, environment...)
}
func dialPrepared(ctx context.Context, c common.MCPServer, base string, diagnostics io.Writer, prepare func(*exec.Cmd) error, environment ...string) (common.MCPTransport, error) {
	switch c.Transport {
	case "url":
		return newHTTPTransport(c, diagnostics)
	case "stdio":
		if c.Command == "" {
			return nil, errors.New("MCP stdio requires command")
		}
		cmd := exec.Command(c.Command, c.Args...)
		cmd.Dir = base
		cmd.Env = environment
		if cmd.Env == nil {
			cmd.Env = common.ProcessEnvironment()
		}
		if prepare != nil {
			if err := prepare(cmd); err != nil {
				return nil, err
			}
		}
		cmd.Stderr = diagnostics
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		in, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		out, err := cmd.StdoutPipe()
		if err != nil {
			_ = in.Close()
			return nil, err
		}
		if err = cmd.Start(); err != nil {
			_ = in.Close()
			_ = out.Close()
			return nil, err
		}
		pipe := &processPipe{Reader: out, Writer: in, stdin: in, stdout: out, cmd: cmd, done: make(chan struct{})}
		go func() { _ = cmd.Wait(); close(pipe.done) }()
		return NewStream(pipe), nil
	case "tcp", "unix", "socket":
		network := c.Transport
		if network == "socket" {
			network = "tcp"
		}
		conn, err := (&net.Dialer{}).DialContext(ctx, network, c.Address)
		if err != nil {
			return nil, err
		}
		return NewStream(conn), nil
	default:
		return nil, errors.New("MCP transport unavailable: " + c.Transport)
	}
}
