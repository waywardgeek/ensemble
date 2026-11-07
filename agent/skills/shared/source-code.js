"use strict";
// Lexical display only, not a language parser. Every original byte remains
// in text nodes; provider code never becomes HTML or executable script.
export function highlight(code, source) {
    const tokens = /("(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|\/\/[^\n]*|#[^\n]*|\b(?:class|public|private|static|partial|void|return|if|else|for|foreach|while|async|await|def|import|from|const|let|function|true|false|null|None|True|False)\b|\b\d+(?:\.\d+)?\b)/g;
    let offset = 0;
    for (const match of source.matchAll(tokens)) {
        code.append(code.ownerDocument.createTextNode(source.slice(offset, match.index)));
        const span = code.ownerDocument.createElement("span");
        const token = match[0];
        span.className = /^("|')/.test(token) ? "code-string"
            : /^(\/\/|#)/.test(token) ? "code-comment"
            : /^\d/.test(token) ? "code-number" : "code-keyword";
        span.textContent = token;
        code.append(span);
        offset = match.index + token.length;
    }
    code.append(code.ownerDocument.createTextNode(source.slice(offset)));
}
