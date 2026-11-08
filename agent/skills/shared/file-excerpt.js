import { highlight } from "./source-code.js";
// Shared presentation helper. No path reads: only captured operation evidence.
export function fileExcerpt(document, evidence, operations) {
    if (!evidence || !operations.includes(evidence.operation)) return null;
    const root = document.createElement("div");
    root.className = "file-preview";
    root.append(fileTitle(document, evidence));
    const partial = evidence.start_column > 1;
    const source = /\.(cs|js|mjs|cjs|ts|tsx|jsx|py|go|java|c|h|cpp|hpp|rs|json|sh)$/i.test(evidence.path);
    function section(label, excerpt) {
        if (label) {
            const heading = document.createElement("p");
            heading.textContent = label;
            root.append(heading);
        }
        if (new TextEncoder().encode(excerpt.text).length < excerpt.total_bytes) {
            const notice = document.createElement("p");
            notice.className = "file-truncation";
            notice.textContent = `Truncated display: captured prefix of ${excerpt.total_bytes} UTF-8 bytes.`;
            root.append(notice);
        }
        if (!excerpt.text.length) {
            const empty = document.createElement("p");
            empty.textContent = "No content in this segment.";
            root.append(empty);
            return;
        }
        const pre = document.createElement("pre");
        pre.className = "file-source";
        const lines = excerpt.text.split("\n");
        if (lines.at(-1) === "") lines.pop();
        for (const [index, text] of lines.entries()) {
            const row = document.createElement("span");
            row.className = "file-source-line";
            const number = document.createElement("span");
            number.className = "file-line-number";
            number.textContent = String(evidence.start_line + index);
            const code = document.createElement("code");
            // A mid-line fragment can start inside a string or comment. Without
            // the preceding source, literal display is more faithful than guessed tokens.
            if (source && !partial) highlight(code, text);
            else code.textContent = text;
            row.append(number, code);
            pre.append(row);
        }
        root.append(pre);
    }
    if (evidence.operation === "edit") {
        const note = document.createElement("p");
        note.textContent = "Applied exact replacement, not a whole-file diff. Inserted positions describe the replacement at the recorded anchor.";
        root.append(note);
        section("Removed", evidence.removed);
        section("Inserted", evidence.content);
    } else section("", evidence.content);
    return root;
}

// Shared compact heading; does not construct or discard source DOM.
export function fileTitle(document, evidence) {
    const title = document.createElement("p");
    title.className = "file-name";
    title.textContent = evidence.path.split(/[\\/]/).at(-1) || evidence.path;
    const partial = evidence.start_column > 1;
    const append = evidence.operation === "append";
    let description = evidence.path;
    if (append || partial) {
        const action = append ? (evidence.content.text.length ? "appended" : "empty append") : "partial line";
        title.textContent += ` · ${action} at ${evidence.start_line}:${evidence.start_column}`;
        description += `; ${action} at line ${evidence.start_line}, column ${evidence.start_column} (UTF-16)`;
    }
    title.title = description;
    title.setAttribute("aria-label", description);
    return title;
}
