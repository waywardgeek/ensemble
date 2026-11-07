# File-reading renderers

The `renderers` mapping in `SKILL.md` associates a tool name with a relative
JavaScript entry file. It grants neither tools nor browser execution. Missing
or unapproved custom code uses the host's safe presentation.

## Directory preview API

The current bounded interface is an ES module with a default function:

```javascript
export default function render(document, text, directoryListing) {
    const node = document.createElement("pre");
    node.textContent = text;
    return node;
}
```

Return a detached HTML element. For results, text is the latest retained human
output window (at most 64 KiB), or the bounded tail of the recorded observation
when retrieval is unavailable. This does not alter model input or its cursor.
Complete structured evidence needs no extra fetch; incomplete evidence retains
its rich preview with a separately labeled plain-text tail. Ordinary result cards
fetch once; continuing command sessions own live polling. The host retains
identity, call/result pairing, timing, original speech text and audit payload. Use textContent
for filenames and output. Do not infer entries by splitting lines: a filename
can contain a newline. This preview API is not yet the complete streaming Artifact
lifecycle interface.

The optional third argument is recorded directory evidence:
`{total_entries, entries: [{name, kind}]}`. Names are individual strings, even
when they contain newlines. Kind is `directory`, `file`, or `unknown`; the native
operation currently confirms directories and leaves other kinds unknown. Display
at most the recorded subset and state omissions (`total_entries - entries.length`).
Evidence retains at most 256 entries and 32 KiB of UTF-8 names independently of
model observation clipping. It describes that completed operation, not the current
filesystem. Missing metadata means legacy text, not an empty directory.

The GUI's `renderers/shipped?renderer=TOOL_NAME` endpoint downloads the selected
embedded build renderer's source as
`{"entry":"relative.js","files":{"relative.js":"source"}}`. It does not
include optional disk-asset overrides. A candidate may include relative helper
modules; all declared files participate in its version hash. Review every file.
Ordinary JavaScript imports are not a containment mechanism; trusted code can
still access page authority and fetch undeclared external code.

Staging stores source without executing it. Try live approves this exact candidate
only in the current page. Keep accepts it for future pages. Reload discards an
unaccepted trial. Shipped-only recovery mode bypasses accepted custom code too,
without deleting the accepted selection. No automatic retry can resolve an
uncertain acceptance response; reload and check the durable selection.

Read `list-directory.js` only when adapting or debugging its implementation.

## File preview API

Read, write and edit renderers receive an optional fourth argument, `fileEvidence`.
The third argument remains directory evidence for compatibility. File evidence is
`{operation, path, start_line, start_column, content, removed}`. An excerpt is
`{text, total_bytes}`; fewer UTF-8 bytes in `text` means a truncated display.
`removed` is present only for edits. Positions are one-based; columns count UTF-16
code units. These facts describe the operation, not the file's present state.

The shipped `read-file.js` and code-tools write/edit entries share
`../../shared/file-excerpt.js`. That module imports `highlight` from
`./source-code.js`; both helpers are included in each file-renderer source package. Known source suffixes
receive lexical highlighting, not parsing or semantic analysis; unknown suffixes
remain literal text. Numbered lines are presentation, not a replacement for raw
recorded output. Empty content has no invented written line. Removed and inserted
sections are an applied replacement, not a whole-file diff.

The trial panel selects directory, read, write, edit, search, default tool output, prose, continuing command output, or registered tools discovered across the host's Agents. Accepted
and previous versions are independent per tool and shared across Agent views.
Only one trial is active through the panel at a time: Keep it or reload before
changing tools. Package review and promotion must match the selected tool.
Changing a helper changes the package version; no mutable page-global formatter
is an implicit dependency. Omitting the endpoint query retains directory behavior
for existing clients. Unknown or duplicate renderer queries are refused.

## Additional tools, including MCP

The guarded `renderers/tools` endpoint reads actual registered Tool objects across
this host's root-owned Agents. It returns distinct `{name, renderer}` entries,
not schemas, grants, credentials or source paths. Discovery never connects an MCP
server or loads a skill. Reload the page after later tool registrations.

For tools other than the five shipped specialized previews, `renderer` is
`tool-` followed by the lowercase SHA-256 of the exact UTF-8 tool name. Use the
returned key in selection, staging, package-review and promotion queries. This
separate namespace prevents a tool named `command` or `prose` from replacing a
host presentation slot. The same name shares its approved presentation across
Agents; it is not a server-provenance or output-schema compatibility guarantee.

Choose **Tool: name** in the panel and stage a reviewed package. There is no
invented shipped-source download for that tool. The default export uses the tool
preview signature above. Text is bounded; it is not necessarily complete JSON.
Native evidence is optional, not synthesized for MCP results. Return no element
when the available facts cannot support the view, allowing the default fallback.
Raw output and immutable call/result identity stay with the GUI.

Only explicitly accepted versions load on future pages, and recovery mode skips
them. If no Agent currently registers the name, it is not offered or loaded;
the saved selection is retained for when it returns. Selection/promotion requests
recheck registration and refuse a vanished name. Already loaded presentation code
lasts until reload; that does not grant execution of the tool. Immutable module
URLs use the verified snapshot hash, so relative helpers need no query string or
live-tool lookup. A module URL is not an automatic import or approval.

## Search preview API

`search-files.js` receives optional search evidence as its fifth argument; earlier
arguments are unchanged. It is `{lines, total_lines}`, with each line containing
`{path, number, text, matches}`. Each match is `{start, length}` in zero-based
UTF-16 code units, captured by the .NET regular expression engine. Empty matches
identify context lines; a span with zero length is a real zero-width match.

The display retains a prefix of at most 256 complete match/context lines, 32 KiB
of path-plus-text UTF-8 bytes, and 256 spans per line. It stops before a line that
cannot fit, rather than silently omit that line's highlights. `total_lines`
counts returned match/context lines, not files searched or matching occurrences.
Recursive searches remain best effort for inaccessible files. Full display does
not certify exhaustive access to the directory tree. Grouping never parses paths
from grep text, and highlights never rerun the regex in the browser. Raw output
and model observation limits remain independent.

## Default tool preview

`tool_default` uses the same `(document, boundedText)` entry interface. It handles
call previews and results without a usable specialized renderer, including unknown
tool names. A throwing or invalid specialized renderer falls back to this default;
a throwing, invalid or already-connected default result falls back to literal GUI
text. Neither customization changes raw observations, pairing or tool status.
The shipped source is `skills/shared/tool-default.js`, with no helper dependencies.


## Prose rendering

The `prose` selection uses an in-place function rather than the detached preview
interface: `render(parent, text, allowCopy)`. Chat and readable thinking share it.
Read `skills/shared/README.md` for lifecycle ownership and the complete three-file
parser/lexer package. The GUI retains original text, speech and a literal fallback.

## Continuing command output

The `command` selection exports `render(parent, bytes, rolledOut)`, a synchronous
in-place function. `bytes` is a fresh Uint8Array of at most 64 KiB; `rolledOut`
indicates that the beginning of the retained output is no longer in this window.
The shipped package includes `shared/command-output.js` and its exact
`shared/terminal-output.js` decoder. It is a terminal log, not an emulator.

The GUI owns Agent/Job/origin validation, fetching, cancellation, foreground
placement, notices, selection deferral and scroll position. It retains one bounded
snapshot to repaint when approval changes, including after completion. Custom
code owns only output presentation, not those controls or recorded observations.
Missing or throwing code falls back to literal decoded text, which may show ANSI
control fragments; this does not undo custom side effects. The shipped decoder
resets styling at rolling boundaries and makes complete terminal commands inert.
