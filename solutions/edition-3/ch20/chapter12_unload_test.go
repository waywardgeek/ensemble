package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

// Observe the real manager closing its transport; protocol replies still come
// from the existing MCP peer fixture, not a substitute Agent or job handler.
type c12ObservedTransport struct {
	ensemble.MCPTransport
	closed chan struct{}
	once   sync.Once
}

func (p *c12ObservedTransport) Close() error {
	err := p.MCPTransport.Close()
	p.once.Do(func() { close(p.closed) })
	return err
}

func c12PairedResult(t *testing.T, raw []byte, id, name string) json.RawMessage {
	t.Helper()
	var wire struct {
		Messages []struct {
			Role    string
			Content json.RawMessage
		}
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	calls, results := 0, 0
	var content json.RawMessage
	for _, message := range wire.Messages {
		var blocks []struct {
			Type, ID, Name string
			ToolUseID      string `json:"tool_use_id"`
			IsError        bool   `json:"is_error"`
			Content        json.RawMessage
		}
		if json.Unmarshal(message.Content, &blocks) != nil {
			continue // Ordinary text messages have no tool pair.
		}
		for _, block := range blocks {
			if block.Type == "tool_use" && block.ID == id {
				calls++
				if message.Role != "assistant" || block.Name != name {
					t.Errorf("wrong real call %s: %+v", id, block)
				}
			}
			if block.Type == "tool_result" && block.ToolUseID == id {
				results++
				if message.Role != "user" || block.IsError {
					t.Errorf("unsuccessful/mispaired result %s: %+v", id, block)
				}
				content = block.Content
			}
		}
	}
	if calls != 1 || results != 1 {
		t.Fatalf("real pair %s: calls=%d results=%d", id, calls, results)
	}
	return content
}

func TestChapter12UnloadAnthropicCatalog(t *testing.T) {
	for _, compiled := range []bool{false, true} {
		name := "external-only"
		if compiled {
			name = "with-compiled-grant"
		}
		t.Run(name, func(t *testing.T) {
			browser := []string{"gui_snapshot", "tts_queue", "gui_click", "gui_input"}
			tools := append(append([]string{}, browser...), "dynamic_auto")
			if compiled {
				tools = append(tools, "write_file")
			}
			dir := t.TempDir()
			c9Skill(t, dir, "base", "primary", "tools: read_file\nloadable-skills: gui-debug\n", "PRIMARY MANUAL")
			c9Skill(t, dir, "gui-debug", "loadable", "tools: "+strings.Join(tools, " ")+"\nmcp_servers:\n  - name: browser-debug\n    transport: websocket\n", "PENDING BROWSER MANUAL")
			captured := make(chan []byte, 3)
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					http.Error(w, "read request", 400)
					return
				}
				n := requests.Add(1)
				if n > 3 {
					t.Error("unexpected extra model request")
					http.Error(w, "request bound", 400)
					return
				}
				captured <- raw
				switch n {
				case 1:
					ch5Reply(w, ch5Call("browser-load", "load_skill", map[string]string{"name": "gui-debug"}))
				case 2:
					ch5Reply(w, ch5Call("browser-unload", "unload_skill", map[string]string{"name": "gui-debug"}))
				case 3:
					ch5Reply(w, map[string]string{"type": "text", "text": "finished"})
				}
			}))
			defer server.Close()
			a := c9HTTPAgent(t, dir, server.URL)
			peerTools := []map[string]any{c11Tool("dynamic_auto", "round")}
			for _, name := range browser {
				peerTools = append(peerTools, c11Tool(name, ""))
			}
			transport, peerCalls := c11Transport(t, peerTools)
			observed := &c12ObservedTransport{MCPTransport: transport, closed: make(chan struct{})}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := a.MCP().Connect(ctx, "browser-debug", observed); err != nil {
				t.Fatal(err)
			}
			if _, err := a.Ask(ctx, "Load gui-debug, unload it, then finish."); err != nil {
				t.Fatal(err)
			}
			if requests.Load() != 3 {
				t.Fatalf("requests=%d", requests.Load())
			}
			before, loaded, after := <-captured, <-captured, <-captured
			beforeCatalog := c11ProjectionCatalog(t, before)
			loadedCatalog := c11ProjectionCatalog(t, loaded)
			afterCatalog := c11ProjectionCatalog(t, after)
			for _, name := range browser {
				if _, ok := beforeCatalog[name]; ok {
					t.Errorf("offered external tool declared before load: %s", name)
				}
				if d, ok := loadedCatalog[name]; !ok || !d.Deferred {
					t.Errorf("real load lost external declaration/reference: %s %+v", name, d)
				}
				if d, ok := afterCatalog[name]; ok {
					t.Errorf("unloaded external declaration remains: %s deferred=%v", name, d.Deferred)
				}
				if spec := c9Tool(t, a, name); spec.Enabled || spec.Source != "mcp:browser-debug" {
					t.Errorf("external capability truth changed: %+v", spec)
				}
			}
			if d, ok := afterCatalog["edit_file"]; !ok || !d.Deferred {
				t.Error("ordinary dormant compiled declaration lost")
			}
			if d, ok := afterCatalog["read_file"]; !ok || d.Deferred {
				t.Error("ordinary initial declaration changed")
			}
			if compiled {
				// Chapter 14 withdraws the sole compiled grant, not its instruction survivor.
				if spec := c9Tool(t, a, "write_file"); spec.Enabled || spec.Origin != "tool" {
					t.Errorf("unloaded compiled grant/identity: %+v", spec)
				}
				if _, ok := afterCatalog["write_file"]; ok {
					t.Error("unloaded compiled declaration remains")
				}
				if _, failed := a.Tools().Dispatch(ctx, "write_file", json.RawMessage(`{}`)); !failed {
					t.Error("unloaded compiled tool callable")
				}
			}
			var last struct {
				System json.RawMessage
				Tools  []struct{ Name, Description string }
			}
			if err := json.Unmarshal(after, &last); err != nil {
				t.Fatal(err)
			}
			var system string
			if json.Unmarshal(last.System, &system) != nil {
				var blocks []struct{ Type, Text string }
				if err := json.Unmarshal(last.System, &blocks); err != nil {
					t.Fatal(err)
				}
				for _, block := range blocks {
					if block.Type != "text" {
						t.Fatalf("non-text SYSTEM block: %+v", block)
					}
					system += block.Text
				}
			}
			if !strings.Contains(system, "PENDING BROWSER MANUAL") {
				t.Error("pending manual removed from actual next request")
			}
			pending := false
			for _, skill := range a.Skills().ListSkills() {
				if skill.Name == "gui-debug" {
					pending = skill.Enabled && skill.PendingUnload
				}
			}
			if !pending {
				t.Error("pending skill truth lost")
			}
			select {
			case <-observed.closed:
			default:
				t.Error("unload did not close the actual peer transport")
			}
			if spec := c9Tool(t, a, "dynamic_auto"); spec.Enabled || spec.Ephemeral != "round" {
				t.Errorf("automatic capability truth changed: %+v", spec)
			}
			select {
			case name := <-peerCalls:
				if name != "dynamic_auto" {
					t.Errorf("unexpected peer call %s", name)
				}
			default:
				t.Error("genuinely loaded automatic tool did not run")
			}
			select {
			case name := <-peerCalls:
				t.Errorf("extra/post-unload automatic call: %s", name)
			default:
			}

			// Read genuine paired history. Rendering may project references, but must
			// not replace its status results, IDs, or instruction authority.
			history := a.Log().Context()
			statuses := map[string]string{}
			for _, entry := range history.Dialogue {
				for _, part := range entry.Parts {
					if part.Type == "tool_result" && (part.CallID == "browser-load" || part.CallID == "browser-unload") {
						if part.IsError {
							t.Errorf("canonical status failed: %s", part.CallID)
						}
						for _, text := range part.Parts {
							if text.Type == "text" {
								statuses[part.CallID] += text.Text
							}
						}
					}
				}
			}
			loadContent := c12PairedResult(t, after, "browser-load", "load_skill")
			{
				var status string
				if err := json.Unmarshal(loadContent, &status); err != nil || status == "" || status != statuses["browser-load"] {
					t.Errorf("external-only load lost canonical-status fallback: %s", loadContent)
				}
			}
			var status string
			if err := json.Unmarshal(c12PairedResult(t, after, "browser-unload", "unload_skill"), &status); err != nil || status != statuses["browser-unload"] {
				t.Errorf("wire unload status differs from canonical result: %q", status)
			}
			for _, term := range []string{"compiled", "external grants", "closed"} {
				if !strings.Contains(strings.ToLower(status), term) {
					t.Errorf("unload status omits %q semantics: %q", term, status)
				}
				for _, tool := range last.Tools {
					if tool.Name == "unload_skill" && !strings.Contains(strings.ToLower(tool.Description), term) {
						t.Errorf("unload description omits %q semantics: %q", term, tool.Description)
					}
				}
			}
			canonical, _ := json.Marshal(history)
			if _, err := a.Engine().Render(history); err != nil {
				t.Fatal(err)
			}
			unchanged, _ := json.Marshal(a.Log().Context())
			if string(canonical) != string(unchanged) {
				t.Error("projection mutated canonical history")
			}
			for _, name := range browser {
				if _, failed := a.Tools().Dispatch(ctx, name, json.RawMessage(`{}`)); !failed {
					t.Errorf("unloaded external tool callable: %s", name)
				}
			}
		})
	}
}
