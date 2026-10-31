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
Call gui_snapshot when you need to inspect the connected browser. Call tts_queue
when you need to inspect pending speech. These tools run only when requested;
no browser snapshot or speech queue is attached automatically.
Use gui_click with a visible CSS selector and gui_input with selector and text.
Call gui_snapshot again to confirm the effect. Hidden, omitted or previously
observed state is not evidence of current success: report what you can observe.
