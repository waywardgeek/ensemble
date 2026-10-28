"use strict";
// The browser owns presentation only. Replay replaces derived cards; stable IDs
// make finals, repeated frames, and metadata updates idempotent.
const transcript = document.querySelector("#transcript");
const cards = new Map();
let socket;
let reconnectTimer;
let replaying = false;
let activeSpeech = null;
const speechQueue = [];
let localPauseCause = false;
let agentClosed = false;
const speech = window.speechSynthesis;
const notice = text => { document.querySelector("#notice").textContent = text; };

function readable(part) {
    if (!part) return "";
    if (part.type === "text" || part.type === "thinking") return part.text || "";
    if (part.type === "tool_call") return `${part.name || "Tool"}\n${JSON.stringify(part.args || {}, null, 2)}`;
    if (part.type === "tool_result") return (part.parts || []).map(readable).join("\n");
    if (part.type === "blob") return `[${part.mime || "Media"} attachment]`;
    if (part.type === "redacted") return "[Redacted content]";
    return ""; // Never speak provider signatures or hidden opaque metadata.
}

// A deliberately small safe Markdown renderer: paragraphs, headings, lists,
// inline code, links, and fenced code. It never interprets provider HTML.
function inline(parent, text) {
    const token = /(`[^`]+`|\[([^\]]+)\]\((https?:\/\/[^\s)]+)\))/g;
    let start = 0;
    for (const match of text.matchAll(token)) {
        parent.append(document.createTextNode(text.slice(start, match.index)));
        if (match[0].startsWith("`")) {
            const code = document.createElement("code");
            code.textContent = match[0].slice(1, -1);
            parent.append(code);
        } else {
            const link = document.createElement("a");
            link.textContent = match[2];
            link.href = match[3];
            link.rel = "noopener noreferrer";
            parent.append(link);
        }
        start = match.index + match[0].length;
    }
    parent.append(document.createTextNode(text.slice(start)));
}
// Lexical display only, not a language parser. Every original byte remains
// in text nodes; provider code never becomes HTML or executable script.
function highlight(code, source) {
    const tokens = /("(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|\/\/[^\n]*|#[^\n]*|\b(?:class|public|private|static|partial|void|return|if|else|for|foreach|while|async|await|def|import|from|const|let|function|true|false|null|None|True|False)\b|\b\d+(?:\.\d+)?\b)/g;
    let offset = 0;
    for (const match of source.matchAll(tokens)) {
        code.append(document.createTextNode(source.slice(offset, match.index)));
        const span = document.createElement("span");
        const token = match[0];
        span.className = /^("|')/.test(token) ? "code-string"
            : /^(\/\/|#)/.test(token) ? "code-comment"
            : /^\d/.test(token) ? "code-number" : "code-keyword";
        span.textContent = token;
        code.append(span);
        offset = match.index + token.length;
    }
    code.append(document.createTextNode(source.slice(offset)));
}

function markdown(parent, text) {
    const segments = text.split(/```[^\n]*\n([\s\S]*?)(?:```|$)/g);
    segments.forEach((segment, index) => {
        if (index % 2) {
            const pre = document.createElement("pre");
            const code = document.createElement("code");
            highlight(code, segment);
            pre.append(code);
            parent.append(pre);
            return;
        }
        for (const paragraph of segment.split(/\n\s*\n/)) {
            if (!paragraph) continue;
            const lines = paragraph.split("\n");
            if (lines.every(line => /^\s*(?:[-*]|\d+\.)\s+/.test(line))) {
                const list = document.createElement(/^\s*\d/.test(lines[0]) ? "ol" : "ul");
                for (const line of lines) {
                    const item = document.createElement("li");
                    inline(item, line.replace(/^\s*(?:[-*]|\d+\.)\s+/, ""));
                    list.append(item);
                }
                parent.append(list);
                continue;
            }
            const heading = paragraph.match(/^(#{1,6})\s+([^\n]+)$/);
            const node = document.createElement(heading ? `h${Math.min(heading[1].length + 1, 6)}` : "p");
            inline(node, heading ? heading[2] : paragraph);
            parent.append(node);
        }
    });
}
// Queued speech counts as busy before playback starts. Only transitions emit
// controls: another tab's explicit unpause is never continuously vetoed here.
function updatePauseCause() {
    if (agentClosed) return;
    const typing = document.querySelector("#prompt").value.length > 0;
    const cause = typing || activeSpeech !== null || speechQueue.length > 0;
    if (cause !== localPauseCause && send(cause ? "pause" : "unpause")) {
        localPauseCause = cause;
    }
}
function startSpeech() {
    if (activeSpeech || speechQueue.length === 0) {
        updatePauseCause();
        return;
    }
    const item = speechQueue.shift();
    const utterance = new SpeechSynthesisUtterance(item.text);
    const selected = document.querySelector("#voice").value;
    utterance.voice = speech.getVoices().find(voice => voice.voiceURI === selected) || null;
    activeSpeech = { id: item.id, utterance };
    const settled = () => {
        if (activeSpeech?.utterance !== utterance) return;
        activeSpeech = null;
        startSpeech(); // Both completion and interruption advance the queue.
    };
    utterance.onend = settled;
    utterance.onerror = settled;
    updatePauseCause();
    try { speech.speak(utterance); }
    catch { settled(); notice("Speech could not be played."); }
}
function enqueueSpeech(id, text) {
    if (!speech || !window.SpeechSynthesisUtterance || !text.trim()) return;
    speechQueue.push({ id, text });
    updatePauseCause();
    startSpeech();
}
function cancelSpeech(update = true) {
    activeSpeech = null;
    speechQueue.length = 0;
    speech?.cancel();
    if (update) updatePauseCause();
}
function speak(card) {
    if (!speech || !window.SpeechSynthesisUtterance) {
        notice("Speech is unavailable in this browser.");
        return;
    }
    const same = activeSpeech?.id === card.id;
    cancelSpeech(false);
    if (!same) enqueueSpeech(card.id, card.text);
    updatePauseCause();
}
function narrate(card) {
    if (replaying || !document.querySelector("#narrate").checked) {
        card.narrated = card.text.length;
        return;
    }
    if (card.role !== "assistant") return;
    if (card.kind === "tool_call") {
        if (card.final && !card.narratedTool) {
            card.narratedTool = true;
            enqueueSpeech(card.id, `${card.name || "Tool"}: ${card.text.slice(0, 160)}`);
        }
        return;
    }
    if (card.kind !== "text" && card.kind !== "thinking") return;
    const pending = card.text.slice(card.narrated || 0);
    // Speak complete clauses while streaming, then flush the final remainder.
    // Final content remains the authority if a provider revises its draft.
    let length = card.final ? pending.length : 0;
    if (!card.final) {
        for (const match of pending.matchAll(/[.!?\n](?:\s|$)/g)) {
            length = match.index + match[0].length;
        }
    }
    if (length > 0) {
        enqueueSpeech(card.id, pending.slice(0, length));
        card.narrated = (card.narrated || 0) + length;
    }
}
function cardFor(id) {
    if (cards.has(id)) return cards.get(id);
    const node = document.createElement("article");
    node.dataset.partId = id;
    const card = { id, node, text: "", kind: "text", role: "assistant", status: "", final: false };
    cards.set(id, card);
    transcript.append(node);
    return card;
}
function render(card) {
    const oldDetails = card.node.querySelector("details");
    const wasOpen = oldDetails ? oldDetails.open : false;
    card.node.replaceChildren();
    card.node.dataset.role = card.role;
    card.node.dataset.kind = card.kind;
    card.node.dataset.status = card.status;
    const heading = document.createElement("div");
    heading.className = "card-heading";
    const title = document.createElement("strong");
    title.textContent = `${card.role} · ${card.kind}${card.status ? " · " + card.status : ""}`;
    const speaker = document.createElement("button");
    speaker.type = "button";
    speaker.textContent = "Speak";
    speaker.setAttribute("aria-label", `Speak ${card.role} ${card.kind}`);
    speaker.dataset.action = "speak";
    speaker.disabled = !card.text.trim() || !speech;
    speaker.onclick = () => speak(card);
    heading.append(title, speaker);
    card.node.append(heading);
    if (card.kind === "text") {
        const body = document.createElement("div");
        body.className = "message";
        markdown(body, card.text);
        card.node.append(body);
    } else {
        const details = document.createElement("details");
        details.open = wasOpen;
        const summary = document.createElement("summary");
        summary.textContent = card.kind === "thinking" ? "Thinking" : "Full tool details";
        const pre = document.createElement("pre");
        pre.textContent = card.text;
        details.append(summary);
        if (card.change) {
            const label = document.createElement("p");
            label.textContent = "Applied exact replacement (not a whole-file diff)";
            const change = document.createElement("pre");
            change.className = "edit-change";
            change.textContent = "- " + card.change.old_text.split("\n").join("\n- ")
                + "\n+ " + card.change.new_text.split("\n").join("\n+ ");
            details.append(label, change);
        }
        pre.className = card.toolName === "search_files" ? "search-results" : "tool-output";
        details.append(pre);
        card.node.append(details);
    }
}
function handle(frame) {
    if (frame.type === "replay") {
        replaying = true;
        cards.clear();
        transcript.replaceChildren();
        for (const event of frame.data.events || []) handle(event);
        for (const part of frame.data.partials || []) handle(part);
        for (const agent of frame.data.agents || []) handle(agent);
        replaying = false;
        return;
    }
    if (frame.type === "agent_state") {
        document.querySelector("#state").textContent = frame.state;
        agentClosed = frame.state === "shutdown";
        document.querySelector("#pause").disabled = agentClosed || frame.state === "paused";
        document.querySelector("#resume").disabled = agentClosed || frame.state !== "paused";
        for (const id of ["send", "hint", "interrupt"]) {
            document.querySelector(`#${id}`).disabled = agentClosed;
        }
        return;
    }
    if (frame.type === "state_changed") {
        document.querySelector("#state").textContent = frame.to;
        return;
    }
    if (frame.type === "error" || frame.type === "lifecycle") {
        notice(frame.message || frame.kind);
        return;
    }
    if (!frame.part_id) return;
    const card = cardFor(frame.part_id);
    if (frame.type === "part_delta") {
        if (card.final) return;
        card.kind = frame.kind;
        card.text += frame.chunk; // An empty chunk is a valid part opening.
        if (frame.tool_name) card.name = (card.name || "") + frame.tool_name;
    } else if (frame.type === "part_final") {
        const part = frame.part;
        card.kind = part.type;
        card.role = part.role || "assistant";
        card.text = readable(part);
        card.name = part.name || card.name;
        card.input = part.args || card.input;
        card.final = true;
        if (card.kind === "tool_call" && !card.status) card.status = "planned";
    } else if (frame.type === "tool_dispatched") {
        card.kind = "tool_call";
        card.status = "started";
        card.text = `${frame.name}\n${JSON.stringify(frame.input || {}, null, 2)}`;
        card.final = true;
    } else if (frame.type === "tool_finished") {
        card.kind = "tool_result";
        card.role = "tool";
        card.status = frame.is_error ? "error" : "finished";
        card.text = (frame.result || (frame.parts || []).map(readable).join("\n"))
            .replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, "");
        card.final = true;
        if (frame.tool_part_id && cards.has(frame.tool_part_id)) {
            const call = cards.get(frame.tool_part_id);
            call.status = frame.is_error ? "error" : "finished";
            card.toolName = call.name;
            if (!frame.is_error && call.name === "edit_file" &&
                typeof call.input?.old_text === "string" && typeof call.input?.new_text === "string") {
                card.change = call.input;
            }
            render(call);
        }
    }
    render(card);
    narrate(card);
}
function connect() {
    socket = new WebSocket(`${location.protocol === "https:" ? "wss:" : "ws:"}//${location.host}/ws`);
    socket.onopen = () => {
        document.querySelector("#connection").textContent = "Connected";
        send("subscribe");
        updatePauseCause();
    };
    socket.onmessage = event => {
        try { handle(JSON.parse(event.data)); }
        catch { notice("An invalid display frame was refused."); }
    };
    socket.onclose = () => {
        document.querySelector("#connection").textContent = "Disconnected — reconnecting";
        clearTimeout(reconnectTimer);
        reconnectTimer = setTimeout(connect, 800);
    };
    socket.onerror = () => { notice("Connection unavailable; work may still be running."); };
}
function send(type, text = "") {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        notice("Not connected; input was not sent.");
        return false;
    }
    socket.send(JSON.stringify({ type, text }));
    notice("");
    return true;
}
function submit(type) {
    const input = document.querySelector("#prompt");
    if (input.value.trim() && send(type, input.value)) {
        input.value = "";
        updatePauseCause();
    }
}
document.querySelector("#compose").onsubmit = event => { event.preventDefault(); submit("prompt"); };
document.querySelector("#hint").onclick = () => submit("hint");
for (const type of ["pause", "interrupt"]) document.querySelector(`#${type}`).onclick = () => send(type);
document.querySelector("#resume").onclick = () => send("unpause");
document.querySelector("#prompt").oninput = updatePauseCause;
document.querySelector("#prompt").onkeydown = event => {
    if (event.key === "Escape" && event.target.value.length === 0) {
        event.preventDefault();
        cancelSpeech();
    } else if ((event.altKey && event.key.toLowerCase() === "h") || (event.ctrlKey && event.key === "Enter")) {
        event.preventDefault(); submit("hint");
    } else if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault(); submit("prompt");
    }
};
document.querySelector("#stop-speech").onclick = () => cancelSpeech();
try { document.querySelector("#narrate").checked = localStorage.getItem("ensemble-narrate") === "true"; } catch { }
document.querySelector("#narrate").onchange = event => {
    try { localStorage.setItem("ensemble-narrate", String(event.target.checked)); } catch { }
};
function voices() {
    const select = document.querySelector("#voice");
    const chosen = select.value;
    select.replaceChildren(new Option("Browser default", ""));
    for (const voice of speech?.getVoices() || []) select.add(new Option(`${voice.name} (${voice.lang})`, voice.voiceURI));
    select.value = chosen;
}
if (speech) {
    speech.addEventListener("voiceschanged", voices);
    voices();
    document.addEventListener("visibilitychange", () => {
        if (!document.hidden) {
            // Recorded Chrome compatibility workaround, not a universal claim.
            const wakeup = new SpeechSynthesisUtterance("");
            wakeup.volume = 0;
            speech.speak(wakeup);
        }
    });
}
connect();
