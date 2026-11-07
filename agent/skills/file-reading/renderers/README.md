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

Return a detached HTML element. The host supplies recorded, bounded preview text;
it retains identity, call/result pairing, timing and raw details. Use textContent
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

The GUI's `renderers/shipped` endpoint downloads the embedded build's source as
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
`../../shared/file-excerpt.js`. That module uses the host-loaded lexical helper
`window.EnsembleSourceCode` from `../../shared/source-code.js`. Known source suffixes
receive lexical highlighting, not parsing or semantic analysis; unknown suffixes
remain literal text. Numbered lines are presentation, not a replacement for raw
recorded output. Empty content has no invented written line. Removed and inserted
sections are an applied replacement, not a whole-file diff.

The current trial panel and source-package endpoint still target `list_directory`
only. File renderers are explicitly shipped assets; their metadata does not yet
make them selectable candidates in that panel.
