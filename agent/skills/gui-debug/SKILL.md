---
name: gui-debug
description: Look at your own GUI to debug it
depends:
  - ensemble
---

## GUI Debug Mode

Call `view_gui` to see the GUI as the user currently sees it in the browser.
It returns a markdown snapshot: the panes, every interactive element with a
CSS selector, and previews of the artifacts on screen. The tool exists
whenever the agent runs with `--port`, loaded skill or not; if no browser has
the GUI open, it says so.

Use it to spot layout problems, find broken or missing elements, and check
that a change to the GUI took effect. Look again after each change rather
than trusting an earlier snapshot.

What you cannot do from here yet: click, type, or read the speech queue. The
browser's MCP server offers those tools (`gui_click`, `gui_input`,
`tts_queue`), but this agent does not connect to it. An agent in another
process can, through `--mcp-port` and `mcp-connect`.
