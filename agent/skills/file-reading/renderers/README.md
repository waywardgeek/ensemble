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
