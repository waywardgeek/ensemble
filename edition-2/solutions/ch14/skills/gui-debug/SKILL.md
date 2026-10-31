---
name: gui-debug
type: loadable
description: Inspect and operate the connected browser
depends: ensemble
tools: gui_snapshot tts_queue gui_click gui_input
mcp_servers:
  - name: browser-debug
    transport: websocket
---
The current browser snapshot and speech queue arrive automatically before each
round. Use gui_click with a visible CSS selector and gui_input with selector and
text. Observe the next snapshot to confirm the effect. Hidden or omitted state
is not evidence of success: report what you can actually observe.
