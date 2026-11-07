// Recorded .NET spans, not browser regex evaluation or parsing grep text.
export default function searchFiles(document, preview, directory, file, evidence) {
    if (!evidence) return null;
    const root = document.createElement("div");
    root.className = "search-preview file-preview";
    const summary = document.createElement("p");
    summary.textContent = evidence.lines.length < evidence.total_lines
        ? `Truncated display: ${evidence.lines.length} of ${evidence.total_lines} returned match/context lines; ${evidence.total_lines - evidence.lines.length} omitted.`
        : `${evidence.total_lines} returned match/context lines.`;
    root.append(summary);
    const groups = new Map();
    for (const line of evidence.lines) {
        let group = groups.get(line.path);
        if (!group) {
            const section = document.createElement("section");
            section.className = "search-file";
            const heading = document.createElement("p");
            heading.className = "search-path";
            heading.textContent = line.path;
            const pre = document.createElement("pre");
            pre.className = "file-source";
            section.append(heading, pre); root.append(section);
            group = { pre, previous: null }; groups.set(line.path, group);
        }
        if (group.previous !== null && line.number !== group.previous + 1) {
            const gap = document.createElement("span");
            gap.className = "file-source-line search-gap";
            gap.textContent = "… context gap …";
            group.pre.append(gap);
        }
        group.previous = line.number;
        const row = document.createElement("span"); row.className = "file-source-line";
        const number = document.createElement("span"); number.className = "file-line-number";
        number.textContent = String(line.number);
        const code = document.createElement("code");
        const label = document.createElement("span"); label.className = "search-line-kind";
        label.textContent = line.matches.length ? "match: " : "context: "; code.append(label);
        let cursor = 0;
        for (const span of line.matches) {
            code.append(document.createTextNode(line.text.slice(cursor, span.start)));
            const mark = document.createElement("mark");
            if (span.length === 0) {
                mark.className = "search-zero-match";
                mark.textContent = "[zero-width match]";
                mark.title = `Match at UTF-16 column ${span.start + 1}`;
            } else mark.textContent = line.text.slice(span.start, span.start + span.length);
            code.append(mark); cursor = span.start + span.length;
        }
        code.append(document.createTextNode(line.text.slice(cursor)));
        row.append(number, code); group.pre.append(row);
    }
    return root;
}
