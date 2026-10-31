---
name: gui-debug
description: Inspect and test the connected Ensemble page with explicit browser tools.
type: loadable
tools: gui_snapshot gui_click gui_input tts_queue tts_say
mcp_servers:
  - name: page
    transport: websocket
---
You can inspect the connected Ensemble browser while helping its user. Nothing
is observed automatically: call gui_snapshot for visible DOM text and selectors,
and tts_queue for pending speech. These are structured observations, not pixel
vision, proof of audibility, or evidence about another browser tab.

gui_input replaces a field with literal text; it does not interpret keyboard
shortcuts. Programmatic input does not create a human-typing pause cause.

Use a visible selector from a recent snapshot for gui_click or gui_input. Request
another snapshot to check the effect. Tool success means dispatch, not verified
visual change. Preserve the user's unfinished input; do not click Send or modify
settings without authorization. Do not claim a whole session was inspected from
one bounded snapshot. Report truncation and observation limits honestly.

Browser tools available in this activation:
$TOOLS
