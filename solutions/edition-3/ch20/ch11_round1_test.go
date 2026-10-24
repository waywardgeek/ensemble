package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble"
)

// Deliberately only the chapter's public methods, not an internal adapter API.
type chapterTransport struct{ ensemble.MCPTransport }

func (t chapterTransport) Send(raw json.RawMessage) error { return t.MCPTransport.Send(raw) }
func (t chapterTransport) Recv() (json.RawMessage, error) { return t.MCPTransport.Recv() }

var _ ensemble.MCPTransport = chapterTransport{}

func TestMCPPublicCompatibility(t *testing.T) {
	stream := ensemble.NewMCPTransport(strings.NewReader("{\"jsonrpc\":\"2.0\"}\n"), io.Discard, io.NopCloser(strings.NewReader("")))
	if _, err := stream.Recv(); err != nil {
		t.Fatal(err)
	}
	stream.Close()
	dir := t.TempDir()
	c9Skill(t, dir, "base", "primary", "loadable-skills: external\n", "base")
	c9Skill(t, dir, "external", "loadable", "tools: echo\nmcp_servers:\n  - name: fixture\n    transport: websocket\n", "external")
	a := c9Agent(t, dir, nil)
	local, remote := ensemble.MCPPipe()
	defer remote.Close()
	done := make(chan error, 1)
	go func() {
		for {
			raw, err := remote.Recv()
			if err != nil {
				done <- err
				return
			}
			var req struct {
				ID     json.RawMessage
				Method string
			}
			if err = json.Unmarshal(raw, &req); err != nil {
				done <- err
				return
			}
			if len(req.ID) == 0 {
				continue
			}
			var result any
			switch req.Method {
			case "initialize":
				result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}}
			case "tools/list":
				result = map[string]any{"tools": []map[string]any{{"name": "echo", "description": "Echo", "inputSchema": map[string]any{"type": "object"}}}}
			default:
				done <- io.ErrUnexpectedEOF
				return
			}
			reply, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
			if err = remote.Send(reply); err != nil {
				done <- err
				return
			}
		}
	}()
	if err := a.MCP().Connect(context.Background(), "fixture", chapterTransport{local}); err != nil {
		t.Fatal(err)
	}
	if err := a.Skills().LoadSkill("external"); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range a.Skills().Snapshot().Definitions {
		if d.Name == "echo" {
			found = true
		}
	}
	if !found {
		t.Fatal("Connect discovery was not available to skill admission")
	}
	a.MCP().Close()
	<-done
}

// A peer supplies protocol replies, not fake Agent execution or job outcomes.
func c11Transport(t *testing.T, tools []map[string]any) (ensemble.MCPTransport, <-chan string) {
	t.Helper()
	local, remote := ensemble.MCPPipe()
	return local, c11ServePeer(t, remote, tools)
}

func c11ServePeer(t *testing.T, remote ensemble.MCPTransport, tools []map[string]any) <-chan string {
	t.Helper()
	calls := make(chan string, 32)
	done := make(chan struct{})
	t.Cleanup(func() { remote.Close(); <-done })
	go func() {
		defer close(done)
		for {
			raw, err := remote.Recv()
			if err != nil {
				return
			}
			var request struct {
				ID     json.RawMessage
				Method string
				Params struct{ Name string }
			}
			if json.Unmarshal(raw, &request) != nil {
				return
			}
			if len(request.ID) == 0 {
				continue
			}
			var result any
			switch request.Method {
			case "initialize":
				result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}}
			case "tools/list":
				result = map[string]any{"tools": tools}
			case "tools/call":
				calls <- request.Params.Name
				result = map[string]any{"content": []map[string]any{{"type": "text", "text": "actual peer: " + request.Params.Name}}}
			}
			reply, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
			if remote.Send(reply) != nil {
				return
			}
		}
	}()
	return calls
}
func c11Peer(t *testing.T, a ensemble.Agent, name string, tools []map[string]any) <-chan string {
	t.Helper()
	local, calls := c11Transport(t, tools)
	if err := a.MCP().Connect(context.Background(), name, local); err != nil {
		t.Fatal(err)
	}
	return calls
}
func c11Tool(name, mode string) map[string]any {
	return map[string]any{"name": name, "description": "fixture " + name, "inputSchema": map[string]any{"type": "object"}, "ephemeral": mode}
}
func TestMCPExplicitOwningGrants(t *testing.T) {
	dir := t.TempDir()
	c9Skill(t, dir, "base", "primary", "loadable-skills: owner other collision\n", "base")
	c9Skill(t, dir, "owner", "loadable", "tools: automatic\nmcp_servers:\n  - name: first\n    transport: websocket\n", "owner\n$TOOLS")
	c9Skill(t, dir, "other", "loadable", "tools: same_name\nmcp_servers:\n  - name: second\n    transport: websocket\n", "other")
	c9Skill(t, dir, "collision", "loadable", "tools: read_file\nmcp_servers:\n  - name: collision\n    transport: websocket\n", "collision")
	a := c9Agent(t, dir, nil)
	second := c11Peer(t, a, "second", []map[string]any{c11Tool("same_name", "")})
	if err := a.Skills().LoadSkill("other"); err != nil {
		t.Fatal(err)
	}
	first := c11Peer(t, a, "first", []map[string]any{c11Tool("automatic", "round"), c11Tool("ungranted", "turn"), c11Tool("same_name", "round")})
	if err := a.Skills().LoadSkill("owner"); err != nil {
		t.Fatal(err)
	}
	snap := a.Skills().Snapshot()
	if strings.Contains(snap.Instructions, "automatic") {
		t.Fatalf("hidden automatic name leaked in tool menu: %s", snap.Instructions)
	}
	for _, d := range snap.Definitions {
		if d.Name == "automatic" || d.Name == "ungranted" {
			t.Fatalf("automatic name is model-visible: %s", d.Name)
		}
	}
	for _, s := range snap.Tools {
		if s.Name == "ungranted" || s.Name == "same_name" && s.Source != "mcp:second" {
			t.Fatalf("cross-owner or missing grant admitted: %+v", s)
		}
	}
	round, _ := a.MCP().Gather(context.Background(), true)
	if len(round) == 0 {
		t.Fatalf("gather: %v", round)
	}
	if got := <-first; got != "automatic" {
		t.Fatalf("automatic invocation %q", got)
	}
	for _, name := range []string{"automatic", "ungranted"} {
		parts, failed := a.Tools().Dispatch(context.Background(), name, json.RawMessage(`{}`))
		if len(parts) != 1 || !failed {
			t.Fatalf("guessed %s admitted: %+v", name, parts)
		}
	}
	select {
	case name := <-first:
		t.Fatalf("ungranted/guessed invocation %s", name)
	default:
	}
	select {
	case name := <-second:
		t.Fatalf("ordinary other-server tool became automatic: %s", name)
	default:
	}
	c11Peer(t, a, "collision", []map[string]any{c11Tool("read_file", "")})
	if err := a.Skills().LoadSkill("collision"); err == nil {
		t.Fatal("external tool shadowed compiled owner")
	}
}
func TestMCPIncomingAliasesUseJobs(t *testing.T) {
	dir := t.TempDir()
	c9Skill(t, dir, "base", "primary", "tools: read_file\n", "base")
	a := c9Agent(t, dir, nil)
	file := filepath.Join(t.TempDir(), "local.txt")
	if err := os.WriteFile(file, []byte("local job read result"), 0600); err != nil {
		t.Fatal(err)
	}
	left, right := ensemble.MCPPipe()
	defer right.Close()
	served := make(chan error, 1)
	go func() { served <- a.ServeMCPConnection(context.Background(), left) }()
	invoke := func(id int, method string, params any) map[string]any {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
		if err := right.Send(raw); err != nil {
			t.Fatal(err)
		}
		raw, err := right.Recv()
		if err != nil {
			t.Fatal(err)
		}
		var response map[string]any
		if err = json.Unmarshal(raw, &response); err != nil {
			t.Fatal(err)
		}
		if response["id"] != float64(id) {
			t.Fatalf("correlation: %s", raw)
		}
		return response
	}
	for i, method := range []string{"tools/builtin/list", "tools/builtin/call"} {
		if invoke(i+1, method, map[string]any{})["error"] == nil {
			t.Fatal("preinitialize alias admitted")
		}
	}
	if invoke(3, "initialize", map[string]any{})["error"] != nil {
		t.Fatal("initialize failed")
	}
	if err := right.Send(json.RawMessage(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)); err != nil {
		t.Fatal(err)
	}
	result := invoke(4, "tools/builtin/call", map[string]any{"name": "read_file", "arguments": map[string]any{"path": file}})
	raw, _ := json.Marshal(result)
	if !strings.Contains(string(raw), "local job read result") {
		t.Fatalf("actual local job result missing: %s", raw)
	}
	refused := invoke(5, "tools/builtin/call", map[string]any{"name": "not_admitted", "arguments": map[string]any{}})
	if r, ok := refused["result"].(map[string]any); !ok || r["isError"] != true {
		t.Fatalf("unknown tool accepted: %v", refused)
	}
	right.Close()
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

func TestMCPApplicationConnectAndCollision(t *testing.T) {
	app := ensemble.New(nil)
	t.Cleanup(func() { _ = app.Shutdown() })
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "fixture", BaseDir: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	bad, _ := c11Transport(t, []map[string]any{c11Tool("would_be_partial", ""), c11Tool("read_file", "")})
	if err := a.MCP().Connect(context.Background(), "collision", bad); err == nil {
		t.Fatal("application catalogue shadowed local tool")
	}
	for _, tool := range a.Skills().Snapshot().Tools {
		if tool.Name == "would_be_partial" {
			t.Fatal("failed catalogue partially published")
		}
	}
	calls := c11Peer(t, a, "application", []map[string]any{c11Tool("application_echo", "")})
	result, failed := a.Tools().Dispatch(context.Background(), "application_echo", json.RawMessage(`{}`))
	raw, _ := json.Marshal(result)
	if failed || !strings.Contains(string(raw), "actual peer: application_echo") {
		t.Fatalf("public connection not executable through jobs: %s", raw)
	}
	if got := <-calls; got != "application_echo" {
		t.Fatalf("call %s", got)
	}
}
