"use strict";
// Presentation only: callers retain original Markdown for History and speech.
// Never enable raw HTML or plugins without reviewing this security boundary.
const messageMarkdown = window.markdownit({ html: false, linkify: false, typographer: false });
messageMarkdown.validateLink = url => /^https?:\/\//i.test(url);
messageMarkdown.renderer.rules.image = (tokens, index) =>
    messageMarkdown.utils.escapeHtml(`[Image: ${tokens[index].content}]`);
messageMarkdown.renderer.rules.link_open = (tokens, index, options, env, self) => {
    tokens[index].attrSet("rel", "noopener noreferrer");
    return self.renderToken(tokens, index, options);
};
function markdown(parent, text) {
    const openCode = new Set([...parent.querySelectorAll("details[data-code-index][open]")]
        .map(node => node.dataset.codeIndex));
    const tableScroll = [...parent.querySelectorAll(".markdown-table")].map(node => node.scrollLeft);
    // This HTML is generated only by the pinned parser with raw HTML disabled,
    // restrictive links, and image tokens replaced by escaped text.
    parent.innerHTML = messageMarkdown.render(text);
    for (const [index, table] of [...parent.querySelectorAll("table")].entries()) {
        const wrapper = document.createElement("div");
        wrapper.className = "markdown-table";
        wrapper.tabIndex = 0;
        wrapper.setAttribute("role", "region");
        wrapper.setAttribute("aria-label", "Scrollable table");
        table.before(wrapper);
        wrapper.append(table);
        wrapper.scrollLeft = tableScroll[index] || 0;
        for (const cell of table.querySelectorAll("thead th")) cell.scope = "col";
    }
    for (const [index, pre] of [...parent.querySelectorAll("pre")].entries()) {
        const code = pre.querySelector("code");
        if (!code) continue;
        const source = code.textContent;
        code.replaceChildren();
        highlight(code, source);
        if (source.length <= 1800 && source.split("\n").length <= 24) continue;
        const details = document.createElement("details");
        details.dataset.codeIndex = String(index);
        details.open = openCode.has(String(index));
        const summary = document.createElement("summary");
        summary.textContent = "Expand complete code";
        const preview = document.createElement("pre");
        preview.textContent = source.slice(0, 1800).split("\n").slice(0, 24).join("\n") + "\n…";
        preview.hidden = details.open;
        pre.before(preview, details);
        details.append(summary, pre);
        details.ontoggle = () => { preview.hidden = details.open; };
    }
}
