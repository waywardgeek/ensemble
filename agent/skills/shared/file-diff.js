import { diffLines } from "./jsdiff/diff/line.js";
import { fileExcerpt, fileTitle } from "./file-excerpt.js";
import { highlight } from "./source-code.js";

// Compare captured replacement evidence, never the current file on disk.
export function fileDiff(document, evidence) {
    if (!evidence || evidence.operation !== "edit") return null;
    const bytes = text => new TextEncoder().encode(text).length;
    function fallback(reason) {
        const root = fileExcerpt(document, evidence, ["edit"]);
        const note = document.createElement("p");
        note.className = "file-truncation";
        note.textContent = reason;
        root.insertBefore(note, root.children[1]);
        return root;
    }
    const before = evidence.removed, after = evidence.content;
    if (bytes(before.text) !== before.total_bytes || bytes(after.text) !== after.total_bytes)
        return fallback("Incomplete evidence: showing captured before/after prefixes, not a unified diff.");
    if (bytes(before.text) > 32768 || bytes(after.text) > 32768)
        return fallback("Comparison display limit reached; showing before/after excerpts.");
    const changes = diffLines(before.text, after.text, { maxEditLength: 256, timeout: 25 });
    if (!changes) return fallback("Comparison budget reached; showing before/after excerpts.");

    const root = document.createElement("div");
    root.className = "file-preview file-diff";
    const title = fileTitle(document, evidence);
    title.textContent += " · replacement diff (old/new)";
    root.append(title);
    title.title += "; applied replacement only, not a whole-file diff; columns are old/new line numbers";
    title.setAttribute("aria-label", title.title);
    const pre = document.createElement("pre");
    pre.className = "file-source";
    pre.setAttribute("aria-label", "Applied replacement diff; old and new line numbers, minus removed, plus inserted");
    const rows = [];
    let oldLine = evidence.start_line, newLine = evidence.start_line;
    for (const change of changes) {
        const lines = change.value.match(/[^\n]*\n|[^\n]+$/g) || [];
        for (const line of lines) {
            rows.push({ marker: change.removed ? "−" : change.added ? "+" : " ",
                old: change.added ? "" : oldLine++, new: change.removed ? "" : newLine++,
                text: line.endsWith("\n") ? line.slice(0, -1) : line,
                newline: line.endsWith("\n"), changed: !!(change.added || change.removed) });
        }
    }
    const visible = new Set();
    rows.forEach((row, index) => {
        if (row.changed) for (let i = Math.max(0, index - 3); i <= Math.min(rows.length - 1, index + 3); i++) visible.add(i);
    });
    if (!visible.size) rows.forEach((_, i) => { if (i < 3 || i >= rows.length - 3) visible.add(i); });
    const source = /\.(cs|js|mjs|cjs|ts|tsx|jsx|py|go|java|c|h|cpp|hpp|rs|json|sh)$/i.test(evidence.path);
    let omitted = 0;
    function gap() {
        if (!omitted) return;
        const node = document.createElement("span");
        node.className = "file-source-line file-diff-gap";
        node.textContent = `… ${omitted} unchanged lines omitted …`;
        pre.append(node); omitted = 0;
    }
    rows.forEach((row, i) => {
        if (!visible.has(i)) { omitted++; return; }
        gap();
        const node = document.createElement("span");
        node.className = "file-source-line";
        for (const value of [row.old, row.new, row.marker]) {
            const gutter = document.createElement("span");
            gutter.className = "file-line-number";
            gutter.textContent = String(value);
            node.append(gutter);
        }
        const code = document.createElement("code");
        if (source && evidence.start_column === 1) highlight(code, row.text);
        else code.textContent = row.text;
        node.append(code); pre.append(node);
        if (!row.newline) {
            const notice = document.createElement("span");
            notice.className = "file-source-line file-diff-boundary";
            notice.textContent = `\\ No final newline in ${row.marker === "−" ? "removed" : row.marker === "+" ? "inserted" : "unchanged"} segment`;
            pre.append(notice);
        }
    });
    gap();
    root.append(pre);
    return root;
}
