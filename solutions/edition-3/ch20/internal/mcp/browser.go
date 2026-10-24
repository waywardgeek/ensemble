package mcp

import "ensemble/internal/common"

// The bundled browser attaches after the agent and GUI server are constructed.
// These expected definitions permit startup skill admission, not observations or
// a pretend connection. An attached peer supplies its own catalog; Connect
// validates these expectations only after they have been admitted. Calls still
// require a connected real peer.
// Keep these four definitions in sync with gui/web/gui/app.js browserTools.
func browserDefinitions(spec common.MCPServer) []remoteTool {
	if spec.Name != "browser-debug" || spec.Transport != "websocket" {
		return nil
	}
	return []remoteTool{
		{Name: "gui_snapshot", Description: "Inspect current visible browser DOM, stable node/part identity, visibility and omission limits.", InputSchema: []byte(`{"type":"object","properties":{},"additionalProperties":false}`)},
		{Name: "tts_queue", Description: "Inspect actual browser speech queue and recent native speech callbacks on demand.", InputSchema: []byte(`{"type":"object","properties":{},"additionalProperties":false}`)},
		{Name: "gui_click", Description: "Click a currently visible element by CSS selector; inspect again to verify effects.", InputSchema: []byte(`{"type":"object","properties":{"selector":{"type":"string"}},"required":["selector"],"additionalProperties":false}`)},
		{Name: "gui_input", Description: "Set a visible text input by CSS selector and dispatch input/change events.", InputSchema: []byte(`{"type":"object","properties":{"selector":{"type":"string"},"text":{"type":"string"}},"required":["selector","text"],"additionalProperties":false}`)},
	}
}
