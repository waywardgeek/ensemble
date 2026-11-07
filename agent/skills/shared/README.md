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
preserves unchanged snippet controls, expansion and table scroll. Its `prose` approval
key covers chat and readable thinking together. The default
export is synchronous `render(parent, text, allowCopy)`: update the supplied body
in place; do not take ownership of the card, original text, or speech controls.
The GUI defers updates during selection and supplies the assistant-text copy policy.
A missing or throwing renderer falls back to literal text; reload remains recovery
from custom code side effects, not a sandbox guarantee.

The package includes `markdown.js`, `markdown-parser.js` and `source-code.js`.
The parser adapter encloses the unchanged pinned markdown-it 14.1.0 UMD payload
in module-local CommonJS bindings and exports its factory, without a page global.
Its comment includes the upstream MIT license. The enclosed payload SHA-256 is
`38c70a1e7ca91ab40e2d9e6e60129851a717ed1c7d4acbbdd41bf9503791cf68`;
original provenance is recorded in the GUI vendor README. Raw HTML remains
disabled, links restricted and image placeholders escaped by `markdown.js`.

`tool-default.js` is the configurable bounded-text preview for tool calls and
results without a usable specialized renderer. Its approval key is `tool_default`;
it uses the same reviewed-version lifecycle as specialized tool previews. The GUI
retains a final literal-text fallback if custom code throws or returns an invalid
or already-connected node. Raw details and controls remain GUI-owned.

`command-output.js` presents a bounded byte window in place, using the exact
`terminal-output.js` helper bundled with its `command` approval package. Read
these sources for terminal-log customization; polling and Job lifetime remain in
the GUI. See the renderer API note for selection and rolling-window limits.
