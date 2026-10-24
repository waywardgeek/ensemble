package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ensemble"
)

// A real MCP peer with the bundled schemas; only the selected automatic mode
// differs. It also runs as a local stdio child for effective-endpoint coverage.
func chapter12Peer(transport ensemble.MCPTransport, automaticTool, mode string) error {
	defer transport.Close()
	var tools []map[string]any
	if err := json.Unmarshal([]byte(`[
 {"name":"gui_snapshot","inputSchema":{"type":"object","properties":{},"additionalProperties":false}},
 {"name":"tts_queue","inputSchema":{"type":"object","properties":{},"additionalProperties":false}},
 {"name":"gui_click","inputSchema":{"type":"object","properties":{"selector":{"type":"string"}},"required":["selector"],"additionalProperties":false}},
 {"name":"gui_input","inputSchema":{"type":"object","properties":{"selector":{"type":"string"},"text":{"type":"string"}},"required":["selector","text"],"additionalProperties":false}}
 ]`), &tools); err != nil {
		return err
	}
	for _, tool := range tools {
		tool["description"] = "local fixture browser tool"
		if tool["name"] == automaticTool {
			tool["ephemeral"] = mode
		}
	}
	for {
		raw, err := transport.Recv()
		if err != nil {
			return err
		}
		var request struct {
			ID     json.RawMessage
			Method string
		}
		if err := json.Unmarshal(raw, &request); err != nil {
			return err
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
			result = map[string]any{"content": []map[string]string{{"type": "text", "text": "real override peer"}}}
		default:
			return fmt.Errorf("unexpected method %s", request.Method)
		}
		response, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result})
		if err != nil {
			return err
		}
		if err := transport.Send(response); err != nil {
			return err
		}
	}
}

func TestChapter12BrowserPeerProcess(t *testing.T) {
	if os.Args[len(os.Args)-1] != "chapter12-stdio-peer" {
		return
	}
	transport := ensemble.NewMCPStream(ensemble.SplitReadWriteCloser{Reader: os.Stdin, Writer: os.Stdout})
	if err := chapter12Peer(transport, "", ""); err != nil && err != io.EOF {
		t.Fatal(err)
	}
}

func chapter12BrowserAgent(t *testing.T, overrides []ensemble.MCPServer) ensemble.Agent {
	t.Helper()
	dir := t.TempDir()
	for name, text := range map[string]string{
		"base":    "---\nname: base\ntype: primary\ndescription: base\nloadable-skills: browser\n---\nbase",
		"browser": "---\nname: browser\ntype: loadable\ndescription: browser\ntools: gui_snapshot tts_queue gui_click gui_input\nmcp_servers:\n  - name: browser-debug\n    transport: websocket\n---\nOn demand only.",
	} {
		path := filepath.Join(dir, name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	en := ensemble.New(log.New(io.Discard, "", 0))
	t.Cleanup(func() { _ = en.Shutdown() })
	a, err := en.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "fixture", APIKey: "fixture", BaseDir: dir, DataDir: t.TempDir(), Skills: &ensemble.SkillSpec{Directory: dir, Primary: "base"}, StartupSkills: []string{"browser"}, MCPServers: overrides})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestChapter12BrowserAutomaticIdentity(t *testing.T) {
	for _, mode := range []string{"", "round", "turn"} {
		for _, name := range []string{"gui_snapshot", "tts_queue", "gui_click", "gui_input"} {
			t.Run(name+"/"+mode, func(t *testing.T) {
				a := chapter12BrowserAgent(t, nil)
				local, remote := ensemble.MCPPipe()
				defer local.Close()
				go func() { _ = chapter12Peer(remote, name, mode) }()
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				err := a.MCP().Connect(ctx, "browser-debug", local)
				if mode != "" {
					if err == nil {
						t.Fatal("same-schema automatic browser tool was accepted as ordinary")
					}
					if !strings.Contains(err.Error(), "automatic") {
						t.Fatalf("wrong rejection: %v", err)
					}
					if _, failed := a.Tools().Dispatch(ctx, name, json.RawMessage(`{}`)); !failed {
						t.Fatal("rejected peer remained callable")
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					if _, failed := a.Tools().Dispatch(ctx, "gui_snapshot", json.RawMessage(`{}`)); failed {
						t.Fatal("ordinary attached browser was not callable")
					}
				}
			})
		}
	}
}

func TestChapter12SkillFrameBrowserEligibility(t *testing.T) {
	for _, transport := range []string{"websocket", "stdio"} {
		t.Run(transport, func(t *testing.T) {
			var overrides []ensemble.MCPServer
			if transport == "stdio" {
				executable, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				overrides = []ensemble.MCPServer{{Name: "browser-debug", Transport: "stdio", Command: executable, Args: []string{"-test.run=^TestChapter12BrowserPeerProcess$", "--", "chapter12-stdio-peer"}}}
			}
			a := chapter12BrowserAgent(t, overrides)
			frame := skillFrame(a)
			raw, err := json.Marshal(frame)
			if err != nil {
				t.Fatal(err)
			}
			// Feed the actual emitted frame to the full production app, using the
			// existing controlled-DOM fixture rather than reproducing its predicate.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "node", "-e", `const fixture=require('./client_round1_test.cjs');const f=fixture();const frame=JSON.parse(require('fs').readFileSync(0,'utf8'));f.apply(frame);f.apply(frame);process.stdout.write(JSON.stringify(f.sent));`)
			cmd.Stdin = strings.NewReader(string(raw))
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("production app fixture: %v\n%s", err, output)
			}
			var commands []struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(output, &commands); err != nil {
				t.Fatalf("%v: %s", err, output)
			}
			opens := 0
			for _, command := range commands {
				if command.Type == "mcp_open" {
					opens++
				}
			}
			want := 1
			if transport == "stdio" {
				want = 0
			}
			if opens != want {
				t.Errorf("actual frame emitted %d tunnel opens, want %d: %s", opens, want, raw)
			}
			for _, skill := range frame.Skills {
				for _, server := range skill.MCPServers {
					if server.Transport != "websocket" || server.Command != "" || len(server.Args) != 0 {
						t.Fatalf("canonical declaration changed or endpoint leaked: %+v", server)
					}
				}
			}
			if transport == "stdio" {
				result, failed := a.Tools().Dispatch(ctx, "gui_snapshot", json.RawMessage(`{}`))
				if failed || !strings.Contains(fmt.Sprint(result), "real override peer") {
					t.Fatalf("real stdio override not used: %v, failed=%v", result, failed)
				}
			}
		})
	}
}
