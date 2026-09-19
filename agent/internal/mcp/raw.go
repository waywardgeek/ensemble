package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// RawTransport is a transport over arbitrary io.Reader/io.Writer pairs.
// Used for --mcp-pipe mode where the process's own stdin/stdout become the
// MCP wire, and for testing.
type RawTransport struct {
	reader *bufio.Scanner
	writer io.Writer
	mu     sync.Mutex
	closed bool
}

// NewRawTransport creates a transport from an io.Reader and io.Writer.
// Messages are newline-delimited JSON.
func NewRawTransport(r io.Reader, w io.Writer) *RawTransport {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	return &RawTransport{
		reader: scanner,
		writer: w,
	}
}

func (t *RawTransport) Send(msg json.RawMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return fmt.Errorf("transport closed")
	}
	_, err := fmt.Fprintf(t.writer, "%s\n", msg)
	return err
}

func (t *RawTransport) Recv() (json.RawMessage, error) {
	if t.reader.Scan() {
		line := t.reader.Bytes()
		cp := make([]byte, len(line))
		copy(cp, line)
		return json.RawMessage(cp), nil
	}
	if err := t.reader.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

func (t *RawTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	// Close the writer if it supports it.
	if c, ok := t.writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
