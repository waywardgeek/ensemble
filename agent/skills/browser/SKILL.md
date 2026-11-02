---
name: browser
description: Inspect, operate or request a human-approved capture of the connected Ensemble page, and inspect or queue its speech.
type: loadable
tools: gui_snapshot gui_capture gui_click gui_input tts_queue tts_say
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

Request gui_capture for actual PNG pixels of the visible connected Ensemble tab.
It takes no arguments. The human must approve the dedicated dialog and choose
this tab in the browser picker; gui_click cannot approve capture. This is not a
full-scroll screenshot, a DOM reconstruction, or authority to capture another
tab, window or the desktop. Unsupported browsers (including Android Chrome
without display capture) and unsupported model/tool-image placements refuse;
the human can attach an OS screenshot instead.

The requesting Agent must remain selected on the original connection. Decline,
selection changes, disconnect, session closure and explicit cancellation make
late pixels ineligible. A foreground observation may return a running Job while
consent is pending; that is not a successful screenshot or permission to retry.
Use the same Job to observe completion, or cancel it explicitly. Only a completed
result carrying image content establishes image delivery. Capture itself does
not send a human chat message or alter the human draft.

Request tts_queue to inspect the current speech and FIFO. tts_say queues text;
interrupt:true replaces current and queued speech. These results establish queue
operations, not audible playback or a human's ability to hear them. There are no
automatic browser snapshots or speech-queue attachments.
