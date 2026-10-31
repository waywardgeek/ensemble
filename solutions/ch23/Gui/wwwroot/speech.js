"use strict";
// Speech formatting is a value transformation, not another conversation store.
// Callers own the raw text and offset. Offsets always count source characters,
// never the shorter filtered text, so a final cannot repeat streamed speech.
function cleanSpeech(text) {
    return text
        // Resolve a complete (or final unterminated) fence BEFORE flattening lines.
        .replace(/```[^\n]*\n[\s\S]*?(?:```|$)/g, " code block ")
        .replace(/```[^\n]*$/g, " code block ")
        .replace(/!?\[([^\]]*)\]\([^)]*\)/g, "$1")
        .replace(/^[ \t]*(?:#{1,6}[ \t]+|>[ \t]*|[-*+][ \t]+|\d+\.[ \t]+)/gm, "")
        .replace(/(\*\*|__)([\s\S]*?)\1/g, "$2")
        .replace(/\*([^*]+)\*/g, "$1")
        .replace(/\b_([^_]+)_\b/g, "$1")
        .replace(/[`*]/g, "")
        .replace(/([A-Z]+)([A-Z][a-z])/g, "$1 $2")
        .replace(/([a-z0-9])([A-Z])/g, "$1 $2")
        .replace(/([\p{L}\p{N}])_+(?=[\p{L}\p{N}])/gu, "$1 ")
        .replace(/\s+/g, " ")
        .trim();
}

// Scan the unconsumed source, not individually filtered deltas. Incomplete
// Markdown stays buffered: splitting a fence/link first would expose its syntax
// or contents to the voice. A lone source newline is whitespace, not a boundary.
// This intentionally supports the same small Markdown vocabulary as the GUI.
function speechPhrases(text, offset, final) {
    const phrases = [];
    let start = offset;
    let cursor = offset;
    const emit = end => {
        const filtered = cleanSpeech(text.slice(start, end));
        if (filtered) phrases.push(filtered);
        start = end;
    };
    while (cursor < text.length) {
        if (text[cursor] === "`") {
            const fence = text.startsWith("```", cursor);
            const marker = fence ? "```" : "`";
            // One or two trailing ticks might become a fence in the next delta.
            if (!final && /^`{1,2}$/.test(text.slice(cursor))) break;
            const closing = text.indexOf(marker, cursor + marker.length);
            if (closing < 0) break;
            cursor = closing + marker.length;
            continue;
        }
        if (text[cursor] === "[") {
            const labelEnd = text.indexOf("]", cursor + 1);
            if (labelEnd < 0) break;
            if (labelEnd + 1 === text.length && !final) break;
            if (text[labelEnd + 1] === "(") {
                const linkEnd = text.indexOf(")", labelEnd + 2);
                if (linkEnd < 0) break;
                cursor = linkEnd + 1;
                continue;
            }
        }
        const paragraph = text.slice(cursor).match(/^\n[ \t]*\n/);
        if (paragraph) {
            cursor += paragraph[0].length;
            emit(cursor);
            continue;
        }
        // Require following whitespace while streaming: an end-of-delta period
        // might belong to a decimal or path whose next characters have not arrived.
        const sentence = /[.!?]/.test(text[cursor]) &&
            (cursor + 1 < text.length ? /\s/.test(text[cursor + 1]) : final);
        const numberedBullet = text[cursor] === "." && /(?:^|\n)[ \t]*\d+$/.test(text.slice(start, cursor));
        if (sentence && !numberedBullet) {
            cursor++;
            emit(cursor);
            continue;
        }
        cursor++;
    }
    if (final) emit(text.length); // Also consumes markup-only and empty tails.
    return { offset: start, phrases };
}
