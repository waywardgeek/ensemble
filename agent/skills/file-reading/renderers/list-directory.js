// Recorded operation evidence, never a filesystem lookup. Older History has
// only bounded text; do not split it into entries because names may contain LF.
export default function directoryPreview(document, text, listing) {
    const body = document.createElement("div");
    body.className = "directory-preview";
    const label = document.createElement("div");
    body.append(label);
    if (!listing) {
        label.remove();
        const output = document.createElement("pre");
        output.className = "tool-output";
        output.textContent = text;
        body.append(output);
        return body;
    }
    const count = listing.entries.length;
    label.textContent = count === listing.total_entries
        ? `${count} entries` : `${count} of ${listing.total_entries} entries shown (${listing.total_entries - count} omitted)`;
    const entries = document.createElement("ul");
    for (const entry of listing.entries) {
        const item = document.createElement("li");
        const name = document.createElement("code");
        name.style.whiteSpace = "pre-wrap";
        name.textContent = entry.name;
        item.append(name);
        if (entry.kind === "directory") item.append(document.createTextNode(" / (directory)"));
        entries.append(item);
    }
    body.append(entries);
    return body;
}
