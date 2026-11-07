"use strict";
// Chat can arrive before the optional lexer. Keep literal text until it loads.
let highlight = (code, source) => { code.textContent = source; };
void import(new URL("./source-code.js", document.currentScript.src).href)
    .then(module => {
        highlight = module.highlight;
        document.dispatchEvent(new Event("ensemble-renderers-changed"));
    }).catch(() => { /* Literal source remains readable. */ });

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
function closedCodeFence(token) {
    if (token?.type !== "fence" || !token.map) return false;
    // The pinned parser removes container prefixes without removing body lines.
    // Its exclusive source span includes the closer only when one was consumed.
    // Count a final unterminated body line too: EOF is not a closing fence.
    const bodyLines = token.content.length === 0 ? 0
        : token.content.split("\n").length - Number(token.content.endsWith("\n"));
    return token.map[1] - token.map[0] === bodyLines + 2;
}
function snippetCopyControls(snippet, source) {
    const controls = document.createElement("div");
    controls.className = "snippet-controls";
    const copy = document.createElement("button");
    copy.type = "button";
    copy.className = "btn btn-quiet";
    copy.dataset.action = "copy-code";
    copy.textContent = "Copy code";
    const status = document.createElement("span");
    status.setAttribute("role", "status");
    let manual = null;
    copy.onclick = async () => {
        // This closure owns the click-time logical payload, not a mutable card
        // lookup or a truncated preview. No clipboard reads or command execution.
        copy.disabled = true;
        status.textContent = "Copying…";
        try {
            if (!navigator.clipboard?.writeText) throw new Error("Clipboard unavailable");
            await navigator.clipboard.writeText(source);
            if (!snippet.isConnected) return;
            status.textContent = "Copied";
            if (manual) manual.hidden = true;
        } catch {
            if (!snippet.isConnected) return;
            status.textContent = "Clipboard unavailable. Select and copy the code below.";
            if (!manual) {
                manual = document.createElement("textarea");
                manual.className = "snippet-manual";
                manual.readOnly = true;
                manual.setAttribute("aria-label", "Code for manual copying");
                manual.value = source;
                controls.after(manual);
            }
            manual.hidden = false;
            manual.focus({ preventScroll: true });
            manual.select();
        } finally {
            copy.disabled = false;
        }
    };
    controls.append(copy, status);
    snippet.append(controls);
}
function markdown(parent, text, allowCopy = false) {
    const oldSnippets = new Map([...parent.querySelectorAll(".code-snippet")]
        .map(node => [node.dataset.codeIndex, node]));
    const focused = parent.contains(document.activeElement) ? document.activeElement : null;
    const openCode = new Set([...parent.querySelectorAll("details[data-code-index][open]")]
        .map(node => node.dataset.codeIndex));
    const tableScroll = [...parent.querySelectorAll(".markdown-table")].map(node => node.scrollLeft);
    const tokens = messageMarkdown.parse(text, {});
    const codeTokens = tokens.filter(token => token.type === "fence" || token.type === "code_block");
    // Generated only by the pinned parser with raw HTML disabled, restrictive
    // links, and image tokens replaced by escaped text. Parse once for both uses.
    parent.innerHTML = messageMarkdown.renderer.render(tokens, messageMarkdown.options, {});
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
        const token = codeTokens[index];
        const source = token.content;
        const copyable = allowCopy && closedCodeFence(token);
        const previous = oldSnippets.get(String(index));
        // Keep feedback, expansion, pending copy and manual selection on an
        // unchanged block while later prose streams. Changed blocks get new UI.
        if (previous && previous.dataset.copyable === String(copyable)
            && previous.querySelector("code").textContent === source
            && previous.dataset.language === token.info) {
            pre.replaceWith(previous);
            continue;
        }
        const snippet = document.createElement("div");
        snippet.className = "code-snippet";
        snippet.dataset.codeIndex = String(index);
        snippet.dataset.copyable = String(copyable);
        snippet.dataset.language = token.info;
        pre.before(snippet);
        if (copyable) snippetCopyControls(snippet, source);
        snippet.append(pre);
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
    // Reparenting a retained control can blur it. Restore only that same node;
    // never move focus to a replacement snippet with different click-time bytes.
    if (focused && parent.contains(focused)) focused.focus({ preventScroll: true });
}
