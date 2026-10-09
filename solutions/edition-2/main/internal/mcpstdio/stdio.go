// Package mcpstdio owns LF framing and the directly launched process endpoint.
package mcpstdio

import (
	"bufio"
	"context"
	"example.com/ensemble/internal/common"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type Constructor struct{ options common.MCPStdioOptions }

func New(options common.MCPStdioOptions) (*Constructor, error) {
	if !filepath.IsAbs(options.Command) || !filepath.IsAbs(options.CWD) {
		return nil, fmt.Errorf("stdio command and cwd must be absolute paths")
	}
	options.Args = append([]string{}, options.Args...)
	options.EnvAllowlist = append([]string{}, options.EnvAllowlist...)
	seen := map[string]bool{}
	for _, n := range options.EnvAllowlist {
		valid, _ := regexp.MatchString(`^[A-Za-z_][A-Za-z0-9_]*$`, n)
		if !valid || seen[n] {
			return nil, fmt.Errorf("invalid environment allowlist")
		}
		seen[n] = true
	}
	return &Constructor{options: options}, nil
}
func (c *Constructor) Open(ctx context.Context, parent common.MCPConnection) (common.MCPTransport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !supported() {
		return nil, fmt.Errorf("stdio processes supported only on macOS/Linux")
	}
	cmd := exec.Command(c.options.Command, c.options.Args...)
	cmd.Dir = c.options.CWD
	cmd.Env = []string{}
	for _, n := range c.options.EnvAllowlist {
		v, ok := os.LookupEnv(n)
		if !ok {
			return nil, fmt.Errorf("allowlisted environment value missing")
		}
		cmd.Env = append(cmd.Env, n+"="+v)
	}
	processGroup(cmd)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, outWriter, err := os.Pipe()
	if err != nil {
		in.Close()
		return nil, err
	}
	cmd.Stdout=outWriter
 diagnostic, diagnosticWriter, err := os.Pipe()
	if err != nil {
		in.Close()
		out.Close();outWriter.Close()
		return nil, err
	}
	cmd.Stderr=diagnosticWriter
 if err = cmd.Start(); err != nil {
		in.Close()
		out.Close()
		diagnostic.Close();outWriter.Close();diagnosticWriter.Close()
		return nil, fmt.Errorf("stdio process start failed")
	}
	outWriter.Close();diagnosticWriter.Close()
 t := &transport{parent: parent, command: cmd, input: in, output: out, stderr: diagnostic, reader: bufio.NewReaderSize(out, 32768), exited: make(chan struct{})}
	t.workers.Add(2)
	go func() { defer t.workers.Done(); _ = cmd.Wait(); close(t.exited) }()
	go func() {
		defer t.workers.Done()
		buffer := make([]byte, 4096)
		for {
			n, e := diagnostic.Read(buffer)
			if n > 0 {
				t.mu.Lock()
				t.tail = append(t.tail, buffer[:n]...)
				if len(t.tail) > 16<<10 {
					t.truncated = true
					t.tail = append([]byte(nil), t.tail[len(t.tail)-(16<<10):]...)
				}
				t.mu.Unlock()
			}
			if e != nil {
				return
			}
		}
	}()
	if err = ctx.Err(); err != nil {
		t.Close()
		return nil, err
	}
	return t, nil
}

type transport struct {
	parent          common.MCPConnection
	command         *exec.Cmd
	input           io.WriteCloser
	output, stderr  io.ReadCloser
	reader          *bufio.Reader
	mu              sync.Mutex
	closed          bool
	tail            []byte
	truncated       bool
	active, workers sync.WaitGroup
	exited          chan struct{}
	once            sync.Once
}

func (t *transport) Connection() common.MCPConnection { return t.parent }
func (t *transport) begin() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return false
	}
	t.active.Add(1)
	return true
}
func (t *transport) Send(ctx context.Context, m common.MCPMessage) error {
	if !t.begin() {
		return fmt.Errorf("stdio closed")
	}
	defer t.active.Done()
	if len(m.JSON) == 0 || len(m.JSON) > common.MCPMessageLimit || strings.ContainsRune(string(m.JSON), '\n') || !utf8.Valid(m.JSON) {
		return fmt.Errorf("invalid complete message")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	callback:=make(chan struct{});stop := context.AfterFunc(ctx, func() { _ = t.input.Close();close(callback) })
 defer func(){if !stop(){<-callback}}()
	data := append(append([]byte(nil), m.JSON...), '\n')
	for len(data) > 0 {
		n, e := t.input.Write(data)
		if e != nil {
			return fmt.Errorf("stdio write failed")
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return ctx.Err()
}
func (t *transport) Receive(ctx context.Context) ([]byte, error) {
	if !t.begin() {
		return nil, io.EOF
	}
	defer t.active.Done()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	callback:=make(chan struct{});stop := context.AfterFunc(ctx, func() { _ = t.output.Close();close(callback) })
 defer func(){if !stop(){<-callback}}()
	var message []byte
	for {
		chunk, err := t.reader.ReadSlice('\n')
		if len(message)+len(chunk) > common.MCPMessageLimit+1 {
			return nil, fmt.Errorf("stdio line limit")
		}
		message = append(message, chunk...)
		if err == bufio.ErrBufferFull {
			continue
		}
		if err != nil {
			if len(message) != 0 {
				return nil, fmt.Errorf("nonempty EOF fragment")
			}
			return nil, err
		}
		break
	}
	message = message[:len(message)-1]
	if len(message) == 0 || len(message) > common.MCPMessageLimit || !utf8.Valid(message) {
		return nil, fmt.Errorf("invalid stdio line")
	}
	return message, nil
}
func (t *transport) Abandon(ctx context.Context, id string, b []byte) error {
	return t.Send(ctx, common.MCPMessage{RequestID: id, JSON: b})
}
func (t *transport) Diagnostics() common.MCPDiagnostics {
	t.mu.Lock()
	defer t.mu.Unlock()
	return common.MCPDiagnostics{StderrTail: append([]byte(nil), t.tail...), Truncated: t.truncated}
}
func (t *transport) Close() error {
	t.once.Do(func() {
		t.mu.Lock()
		t.closed = true
		t.mu.Unlock()
		_ = t.input.Close()
		_ = t.output.Close()
		_ = t.stderr.Close()
		select {
		case <-t.exited:
		case <-time.After(time.Second):
		}
		terminate(t.command, false)
		select {
		case <-t.exited:
		case <-time.After(time.Second):
		}
		terminate(t.command, true)
		<-t.exited
		t.active.Wait()
		t.workers.Wait()
	})
	return nil
}
