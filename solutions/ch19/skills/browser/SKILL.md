---
name: browser
description: Inspect and operate the connected Ensemble page, and inspect or queue its speech.
type: loadable
tools: gui_snapshot gui_click gui_input tts_queue tts_say
mcp_servers:
  - name: page
    transport: websocket
---
These tools act on the one connected Ensemble browser tab, not arbitrary browser
sessions. Loading requires a connected GUI. Other tabs may observe the Agent but
are not additional tool targets. Disconnecting removes this connection's tools;
unloading closes its MCP session. Reconnect before establishing a fresh session.

Request gui_snapshot when you need visible page text and CSS selectors. It is a
structured DOM observation, not pixel vision; heed its truncation flags. Use a
selector from the snapshot for gui_click or gui_input, then request another
snapshot to observe the effect. A successful dispatch is not proof of its effect.

Request tts_queue to inspect the current speech and FIFO. tts_say queues text;
interrupt:true replaces current and queued speech. These results establish queue
operations, not audible playback or a human's ability to hear them. There are no
automatic browser snapshots or speech-queue attachments.
