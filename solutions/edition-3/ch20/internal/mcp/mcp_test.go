package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"testing"

	"ensemble/internal/common"
)

func TestCorrelationAndDisconnect(t *testing.T) {
	left, right := net.Pipe()
	p := newPeer(NewStream(left), nil, false)
	defer p.close(io.EOF)
	remote := NewStream(right)
	defer remote.Close()
	result := make(chan string, 2)
	for _, method := range []string{"alpha", "beta"} {
		go func(method string) {
			var value string
			err := p.call(context.Background(), method, map[string]any{}, &value)
			if err != nil {
				result <- "error: " + err.Error()
			} else {
				result <- value
			}
		}(method)
	}
	messages := []message{}
	for i := 0; i < 2; i++ {
		raw, err := remote.Recv()
		if err != nil {
			t.Fatal(err)
		}
		var m message
		if err = json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		messages = append(messages, m)
	}
	for i := 1; i >= 0; i-- {
		value, _ := json.Marshal(messages[i].Method)
		raw, _ := json.Marshal(message{JSONRPC: "2.0", ID: messages[i].ID, Result: value})
		if err := remote.Send(raw); err != nil {
			t.Fatal(err)
		}
	}
	got := map[string]bool{<-result: true, <-result: true}
	if !got["alpha"] || !got["beta"] {
		t.Fatalf("wrong correlated responses: %v", got)
	}
	pending := make(chan error, 1)
	go func() { var value any; pending <- p.call(context.Background(), "pending", nil, &value) }()
	if _, err := remote.Recv(); err != nil {
		t.Fatal(err)
	}
	_ = remote.Close()
	if err := <-pending; err == nil {
		t.Fatal("disconnect must release pending call")
	}
}
func TestSchemaAndContent(t *testing.T) {
	raw, err := normalizeSchema(json.RawMessage(`{"$schema":"ignored","type":"object","properties":{"n":{"type":"integer","minimum":0}},"required":["n"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "$schema") || !strings.Contains(string(raw), `"minimum":0`) {
		t.Fatalf("normalization: %s", raw)
	}
	if _, err = normalizeSchema(json.RawMessage(`{"type":"array"}`)); err == nil {
		t.Fatal("non-object arguments accepted")
	}
	var output bytes.Buffer
	content := []json.RawMessage{json.RawMessage(`{"type":"text","text":"hello"}`), json.RawMessage(`{"type":"image","data":"aGVsbG8=","mimeType":"image/png"}`)}
	if err = writeContent(&output, content); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "hello\n{") || !strings.Contains(output.String(), "image/png") {
		t.Fatalf("lost content: %s", output.String())
	}
}

// This helper is launched by the real stdio transport, never replaced with an
// in-memory fake. Its stdout is protocol-only; test diagnostics stay on stderr.
func TestMCPChildProcess(t *testing.T) {
	if len(os.Args) == 0 || os.Args[len(os.Args)-1] != "mcp-child" {
		return
	}
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		var m message
		if json.Unmarshal(scan.Bytes(), &m) != nil {
			os.Exit(2)
		}
		if len(m.ID) == 0 {
			continue
		}
		var value any
		switch m.Method {
		case "initialize":
			value = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}}
		case "tools/list":
			value = map[string]any{"tools": []remoteTool{{Name: "echo", Description: "Echo input", InputSchema: json.RawMessage(`{"type":"object"}`)}}}
		case "tools/call":
			value = map[string]any{"content": []common.MCPContent{{Type: "text", Text: string(m.Params)}}}
		default:
			os.Exit(3)
		}
		data, _ := json.Marshal(value)
		reply, _ := json.Marshal(message{JSONRPC: "2.0", ID: m.ID, Result: data})
		_, _ = os.Stdout.Write(append(reply, '\n'))
	}
	os.Exit(0)
}
func TestRealStdioChild(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	transport, err := dial(context.Background(), common.MCPServer{Transport: "stdio", Command: executable, Args: []string{"-test.run=^TestMCPChildProcess$", "mcp-child"}}, t.TempDir(), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	p := newPeer(transport, nil, false)
	defer p.close(io.EOF)
	c := &connection{peer: p}
	m := &manager{}
	if err = m.initialize(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	if len(c.tools) != 1 || c.tools[0].Name != "echo" {
		t.Fatalf("discovery: %v", c.tools)
	}
	var result common.MCPCallResult
	if err = p.call(context.Background(), "tools/call", map[string]any{"name": "echo", "arguments": map[string]string{"text": "split streams"}}, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 1 || !bytes.Contains(result.Content[0], []byte("split streams")) {
		t.Fatalf("result: %s", result.Content)
	}
}
func TestCancelledCallDoesNotWaitForProvider(t *testing.T) {
	left, right := net.Pipe()
	p := newPeer(NewStream(left), nil, false)
	defer p.close(io.EOF)
	remote := NewStream(right)
	defer remote.Close()
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan error, 1)
	go func() { var result any; finished <- p.call(ctx, "tools/call", nil, &result) }()
	if _, err := remote.Recv(); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := <-finished; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context: %v", err)
	}
}
