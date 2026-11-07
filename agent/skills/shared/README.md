# Shared presentation helpers

This directory is shared source, not an activatable skill and not a tool grant.
The GUI explicitly packages reviewed files; directory presence does not execute
code. Other implementation languages may use shared source here as needs arise.

`source-code.js` contains the existing safe lexical highlighter, used by Markdown
and file excerpts. It appends text nodes and spans, never HTML from source text.
`file-excerpt.js` builds bounded numbered file views from recorded operation facts.
Read a renderer's API note before loading implementation source into model context.
These are trusted in-page helpers, not a sandbox or an independent approval system.

`markdown.js` renders chat and supplied readable thinking in place. The GUI owns
card identity, streaming scheduling, selection deferral and speech; this renderer
preserves unchanged snippet controls, expansion and table scroll. It uses the
page-loaded pinned Markdown parser with raw HTML disabled, restricted links and
escaped image placeholders, and imports `source-code.js` for lexical formatting.
It is shipped source, not yet an independently selectable approval package.

`tool-default.js` is the configurable bounded-text preview for tool calls and
results without a usable specialized renderer. Its approval key is `tool_default`;
it uses the same reviewed-version lifecycle as specialized tool previews. The GUI
retains a final literal-text fallback if custom code throws or returns an invalid
or already-connected node. Raw details and controls remain GUI-owned.
