package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// PipeTransport is an in-process transport using io.Pipe.
// NewPipeTransport returns both ends: client and server.
// Messages are newline-delimited JSON (one JSON object per line).
type PipeTransport struct {
	reader *bufio.Scanner
	writer io.WriteCloser
	mu     sync.Mutex // protects writer
	closed bool
}

// NewPipeTransport creates a pair of connected transports.
// Messages written to one side are readable from the other.
func NewPipeTransport() (client Transport, server Transport) {
	// client writes → server reads
	cr, sw := io.Pipe()
	// server writes → client reads
	sr, cw := io.Pipe()

	clientT := &PipeTransport{
		reader: bufio.NewScanner(cr),
		writer: cw,
	}
	clientT.reader.Buffer(make([]byte, 1024*1024), 1024*1024)

	serverT := &PipeTransport{
		reader: bufio.NewScanner(sr),
		writer: sw,
	}
	serverT.reader.Buffer(make([]byte, 1024*1024), 1024*1024)

	return clientT, serverT
}

func (p *PipeTransport) Send(msg json.RawMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return fmt.Errorf("transport closed")
	}
	_, err := fmt.Fprintf(p.writer, "%s\n", msg)
	return err
}

func (p *PipeTransport) Recv() (json.RawMessage, error) {
	if p.reader.Scan() {
		return json.RawMessage(p.reader.Bytes()), nil
	}
	if err := p.reader.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

func (p *PipeTransport) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return p.writer.Close()
}
