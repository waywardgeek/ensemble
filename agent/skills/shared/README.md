# Shared presentation helpers

This directory is shared source, not an activatable skill and not a tool grant.
The GUI explicitly packages reviewed files; directory presence does not execute
code. Other implementation languages may use shared source here as needs arise.

`source-code.js` contains the existing safe lexical highlighter, used by Markdown
and file excerpts. It appends text nodes and spans, never HTML from source text.
`file-excerpt.js` builds bounded numbered file views from recorded operation facts.
Read a renderer's API note before loading implementation source into model context.
These are trusted in-page helpers, not a sandbox or an independent approval system.
