package common

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
)

// MCPServer is an application endpoint. A connection alone grants no tools.
type MCPServer struct {
	Name      string   `json:"name"`
	Transport string   `json:"transport"`
	URL       string   `json:"url,omitempty"`
	AuthEnv   string   `json:"auth_env,omitempty"`
	Command   string   `json:"command,omitempty"`
	Args      []string `json:"args,omitempty"`
	Address   string   `json:"address,omitempty"`
}

// MCPTransport carries complete JSON messages. Stream transports delimit with LF;
// the GUI adapter frames the same messages inside its existing WebSocket.
type MCPTransport interface {
	Recv() (json.RawMessage, error)
	Send(json.RawMessage) error
	Close() error
}
type MCPManager interface {
	Discover(context.Context, []Skill) ([]ToolSpec, error)
	Reconcile(CapabilitySnapshot)
	Gather(context.Context, bool) (PartList, PartList)
	Connect(context.Context, string, MCPTransport) error
	Attach(string, MCPTransport) error
	Serve(context.Context, MCPTransport) error
	Close() error
}

// MCPContent preserves wire content, including unsupported structured blocks.
// Remote content is rendered as text only at the existing job output boundary.
type MCPContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	Data     string          `json:"data,omitempty"`
	MIMEType string          `json:"mimeType,omitempty"`
	Resource json.RawMessage `json:"resource,omitempty"`
}
type MCPCallResult struct {
	Content []json.RawMessage `json:"content"`
	IsError bool              `json:"isError,omitempty"`
}

// SplitReadWriteCloser keeps protocol stdin and stdout separate.
type SplitReadWriteCloser struct {
	Reader    io.Reader
	Writer    io.Writer
	CloseFunc func() error
}

func (s SplitReadWriteCloser) Read(p []byte) (int, error)  { return s.Reader.Read(p) }
func (s SplitReadWriteCloser) Write(p []byte) (int, error) { return s.Writer.Write(p) }
func (s SplitReadWriteCloser) Close() error {
	if s.CloseFunc != nil {
		return s.CloseFunc()
	}
	return nil
}

// ValidateMCPURL rejects ambiguous endpoints without echoing URL secrets.
func ValidateMCPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return errors.New("MCP requires an absolute HTTP(S) URL without userinfo or fragment")
	}
	return nil
}
