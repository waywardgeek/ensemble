package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/mcp"
)

func TestPipeHandshake(t *testing.T) {
	// Create a PipeTransport pair.
	clientT, serverT := mcp.NewPipeTransport()

	// Run fake MCP server.
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Read initialize request.
		raw, err := serverT.Recv()
		if err != nil {
			t.Errorf("server recv initialize: %v", err)
			return
		}
		var req struct {
			JSONRPC string `json:"jsonrpc"`
			ID      int    `json:"id"`
			Method  string `json:"method"`
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("server unmarshal initialize: %v", err)
			return
		}
		if req.Method != "initialize" {
			t.Errorf("first message should be initialize, got %q", req.Method)
			return
		}
		t.Logf("server: got initialize (id=%d)", req.ID)

		// Respond to initialize.
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result": map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo":      map[string]string{"name": "test", "version": "1.0"},
				"capabilities":    map[string]any{"tools": map[string]any{}},
			},
		}
		data, _ := json.Marshal(resp)
		if err := serverT.Send(data); err != nil {
			t.Errorf("server send init response: %v", err)
			return
		}
		t.Log("server: sent init response")

		// Read notifications/initialized.
		raw, err = serverT.Recv()
		if err != nil {
			t.Errorf("server recv notification: %v", err)
			return
		}
		t.Logf("server: got notification: %s", string(raw))

		// Read tools/list.
		raw, err = serverT.Recv()
		if err != nil {
			t.Errorf("server recv tools/list: %v", err)
			return
		}
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("server unmarshal tools/list: %v", err)
			return
		}
		t.Logf("server: got tools/list (id=%d)", req.ID)

		// Respond with tools.
		toolsResp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result": map[string]any{
				"tools": []map[string]any{
					{
						"name":        "echo",
						"description": "Echoes input",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{"text": map[string]string{"type": "string"}},
						},
					},
				},
			},
		}
		data, _ = json.Marshal(toolsResp)
		if err := serverT.Send(data); err != nil {
			t.Errorf("server send tools response: %v", err)
			return
		}
		t.Log("server: sent tools response")
	}()

	// Client side.
	client := mcp.NewClient(clientT)
	defer client.Close()

	ctx := t.Context()
	if err := client.Initialize(ctx); err != nil {
		t.Fatalf("client Initialize: %v", err)
	}
	t.Log("client: Initialize OK")

	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("client ListTools: %v", err)
	}
	t.Logf("client: ListTools returned %d tools", len(tools))

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "echo" {
		t.Errorf("tool name = %q, want echo", tools[0].Name)
	}

	<-done
}
