---
name: gui-debug
description: Debug the GUI through browser MCP tools
depends:
  - ensemble
mcp_servers:
  - name: browser-debug
    transport: websocket
---

## GUI Debug Mode

Loading this skill connects to the browser's MCP server over the
existing WebSocket connection. Four tools become available:

**Ephemeral (auto-injected every round trip):**
- `gui_snapshot` — markdown summary of the current DOM: panes,
  interactive elements with CSS selectors, artifact previews
- `tts_queue` — pending TTS utterances as JSON

**Callable tools:**
- `gui_click(selector)` — click an element by CSS selector
- `gui_input(selector, text)` — set text on an input element

The DOM snapshot appears in your context automatically. Use it to
spot layout issues, find broken elements, verify that actions
produced the expected UI change.
