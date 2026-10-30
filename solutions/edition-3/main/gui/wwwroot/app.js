"use strict";
// The browser owns presentation only. Replay replaces derived cards; stable IDs
// make finals, repeated frames, and metadata updates idempotent.
const transcript = document.querySelector("#transcript");
const cards = new Map();
const toolTranscript = document.querySelector("#tool-transcript");
let acknowledgedState = null;
let selectionPending = true;
const collapsedAgents = new Set();
let socket;
let agentID = null; // Display identity changes only on admitted replay.
let agentRows = []; // Bounded server tree projection, not a client Agent registry.
let reconnectTimer;
let replaying = false;
let activeSpeech = null;
const speechQueue = [];
// The transcript records channel admission, not acoustic hearing. Quiet mode
// shares this exact path and merely skips the native synthesis boundary.
const transcriptSpeech = new URLSearchParams(location.search).get("speech") === "transcript";
const speechLog = [];
window._ttsLog = speechLog;
function recordSpeech(kind, fields) {
    speechLog.push(Object.freeze({ seq: speechLog.length + 1, ms: performance.now(), kind, ...fields }));
}
let localPauseCause = false;
// This tab's control intent, not another copy of the actor's pause state.
// Only a pause requested automatically by this tab may be released on clearing
// typing/speech. Explicit Pause takes over that intent until explicit Resume.
let pauseIntent = null; // "automatic", "manual", or no locally owned pause.
// Origin is a fact, not something speech callbacks can infer from field contents.
let humanTyping = false;
let agentClosed = false;
// Last acknowledged server snapshot. Controls request changes; they never
// become a second preference store. The server is also authoritative on reload.
let settings = null;
let goalSelectionPending = false;
let goalDirty = false; // Unsaved form edits only, never a second live configuration.
const speech = window.speechSynthesis;
const notice = text => { document.querySelector("#notice").textContent = text; };

// Only fixed local symbol names are passed by this module. Model text never
// becomes markup; accessible names belong to the surrounding native buttons.
function setButtonIcon(button, symbol) {
    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    svg.setAttribute("viewBox", "0 0 24 24");
    svg.setAttribute("aria-hidden", "true");
    svg.setAttribute("focusable", "false");
    const use = document.createElementNS(svg.namespaceURI, "use");
    use.setAttribute("href", `#${symbol}`);
    svg.append(use);
    button.classList.add("icon-button");
    button.replaceChildren(svg);
}

function readable(part) {
    if (!part) return "";
    if (part.type === "text" || part.type === "thinking" || part.type === "reasoning_summary" || part.type === "handoff") return part.text || "";
    if (part.type === "tool_call") return `${part.name || "Tool"}\n${JSON.stringify(part.args || {}, null, 2)}`;
    if (part.type === "tool_result") return (part.parts || []).map(readable).join("\n");
    if (part.type === "blob") return `[${part.mime || "Media"} attachment]`;
    if (part.type === "redacted") return "[Redacted content]";
    return ""; // Never speak provider signatures or hidden opaque metadata.
}

// A deliberately small safe Markdown renderer: paragraphs, headings, lists,
// inline code, links, and fenced code. It never interprets provider HTML.
function inline(parent, text) {
    const token = /(`[^`]+`|\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)|\*\*[^*]+\*\*|\*[^*]+\*)/g;
    let start = 0;
    for (const match of text.matchAll(token)) {
        parent.append(document.createTextNode(text.slice(start, match.index)));
        if (match[0].startsWith("`")) {
            const code = document.createElement("code");
            code.textContent = match[0].slice(1, -1);
            parent.append(code);
        } else if (match[0].startsWith("*")) {
            const strong = match[0].startsWith("**");
            const emphasis = document.createElement(strong ? "strong" : "em");
            const delimiter = strong ? 2 : 1;
            emphasis.textContent = match[0].slice(delimiter, -delimiter);
            parent.append(emphasis);
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
            const long = segment.length > 1800 || segment.split("\n").length > 24;
            const details = document.createElement("details");
            const summary = document.createElement("summary");
            summary.textContent = "Expand complete code";
            const preview = document.createElement("pre");
            preview.textContent = segment.slice(0, 1800).split("\n").slice(0, 24).join("\n") + "\n…";
            highlight(code, segment);
            pre.append(code);
            if (long) {
                details.append(summary, pre);
                details.ontoggle = () => { preview.hidden = details.open; };
                parent.append(preview, details);
            } else {
                parent.append(pre);
            }
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
// Queued speech counts as busy before playback starts. Act only on cause
// transitions: explicit Resume (including from another tab) must not be vetoed
// continuously by an unchanged draft or speech queue.
function updatePauseCause() {
    if (agentClosed || selectionPending || replaying || !acknowledgedState ||
        socket?.readyState !== WebSocket.OPEN) return;
    const speaking = activeSpeech !== null || speechQueue.length > 0;
    const cause = humanTyping || (settings?.pause_while_speaking !== false && speaking);
    if (cause === localPauseCause) return;
    localPauseCause = cause;
    if (cause) {
        // An already paused Agent may have been paused elsewhere. Do not claim
        // ownership of that pause merely because a local draft now exists.
        if (pauseIntent === "manual" || acknowledgedState === "paused") return;
        if (!send("pause")) return;
        pauseIntent = "automatic";
        const sources = [];
        if (humanTyping) sources.push("typing");
        if (speaking && settings?.pause_while_speaking !== false) sources.push("speech");
        recordSpeech("pause", { cause: sources.join("+") });
    } else if (pauseIntent === "automatic" && send("unpause")) {
        pauseIntent = null;
        recordSpeech("resume", { cause: "clear" });
    }
    renderControls();
}
function startSpeech() {
    if (activeSpeech) return;
    while (speechQueue.length > 0) {
        const item = speechQueue.shift();
        // Future queued items are NOT in the transcript. Cancelling the queue
        // therefore cannot make unsaid later paragraphs appear as heard.
        recordSpeech("utterance", { part_id: item.id, text: item.text, source: item.source });
        if (transcriptSpeech) continue;
        const utterance = new SpeechSynthesisUtterance(item.text);
        const selected = document.querySelector("#voice").value;
        utterance.rate = settings?.tts_speed ?? 1;
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
        return;
    }
    updatePauseCause();
}
function enqueueSpeech(id, text, source = "explicit") {
    if (!text.trim()) return;
    if (!transcriptSpeech && (!speech || !window.SpeechSynthesisUtterance)) return;
    const label = source === "reasoning_summary" ? "Reasoning summary: "
        : source === "thinking" ? "Thinking: " : source === "error" ? "Error: " : "";
    speechQueue.push({ id, text: label + text, source });
    updatePauseCause();
    startSpeech();
}
// Manual artifact reading and MCP speech use the same filter as narration.
function speakText(id, text, source = "explicit") {
    for (const phrase of speechPhrases(text, 0, true).phrases) {
        enqueueSpeech(id, phrase, source);
    }
}
function cancelSpeech(update = true) {
    activeSpeech = null;
    speechQueue.length = 0;
    speech?.cancel();
    if (update) updatePauseCause();
}
function speak(card) {
    if (!transcriptSpeech && (!speech || !window.SpeechSynthesisUtterance)) {
        notice("Speech is unavailable in this browser.");
        return;
    }
    const same = activeSpeech?.id === card.id;
    cancelSpeech(false);
    if (!same) speakText(card.id, card.text);
    updatePauseCause();
}
// Explicit flush sites (dispatch, error, final) release tails without changing
// the displayed part's lifecycle. Each part retains its own raw source cursor.
function flushNarration() {
    for (const card of cards.values()) {
        if (card.kind === "text" || card.kind === "thinking" || card.kind === "reasoning_summary") narrate(card, true);
    }
}
function narrate(card, flush = false) {
    if (replaying || !settings?.tts_enabled) {
        // A partial replay or muted fragment may end inside Markdown. Do not
        // resume midway through that part and accidentally speak a code body.
        // The next new part can narrate normally; this restored part stays quiet.
        card.speechSuppressed = true;
        card.narrated = card.text.length;
        if (card.final) card.narratedTool = true;
        return;
    }
    if (card.speechSuppressed) return;
    if (card.kind === "tool_result") {
        if (card.status === "error" && !card.narratedTool) {
            card.narratedTool = true;
            flushNarration();
            speakText(card.id, card.text || "Tool failed.", "error");
        }
        return; // Successful tool bodies remain automatically silent.
    }
    if (card.role !== "assistant") return;
    if (card.kind === "tool_call") {
        if (card.status === "started" && !card.narratedTool) {
            card.narratedTool = true;
            flushNarration();
            speakText(card.id, `Running ${card.name || "tool"}.`, "tool-announcement");
        }
        return;
    }
    if (card.kind !== "text" && card.kind !== "thinking" && card.kind !== "reasoning_summary") return;
    const next = speechPhrases(card.text, card.narrated || 0, card.final || flush);
    card.narrated = next.offset;
    for (const phrase of next.phrases) enqueueSpeech(card.id, phrase, card.kind);
}
function actionButton(label, action) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "btn btn-quiet";
    button.textContent = label;
    button.onclick = action;
    return button;
}
function cardFor(id) {
    if (cards.has(id)) return cards.get(id);
    const node = document.createElement("article");
    node.dataset.partId = id;
    const card = { id, agent: agentID, node, text: "", kind: "text", role: "assistant", status: "", final: false };
    cards.set(id, card);
    return card;
}
function media(parent, part, card) {
    if (part.type === "tool_result") {
        for (const child of part.parts || []) media(parent, child, card);
        return;
    }
    if (part.type !== "blob") return;
    // Only passive media receives a data URL. Provider HTML/SVG is never markup.
    const mime = part.mime || "";
    if (!/^(image\/(png|jpeg|gif|webp)|audio\/(mpeg|wav|ogg|webm)|video\/(mp4|webm))$/.test(mime)) return;
    const url = part.ref?.locator;
    if (typeof url !== "string" || !url.startsWith(`data:${mime};base64,`)) return;
    const box = document.createElement("div");
    box.className = "media";
    const element = document.createElement(mime.startsWith("image/") ? "img" : mime.startsWith("audio/") ? "audio" : "video");
    element.src = url;
    if (element.tagName === "IMG") element.alt = "Agent image attachment";
    else element.controls = true;
    const download = document.createElement("a");
    download.href = url;
    download.download = `attachment-${card.id.replace(/[^a-z0-9]/gi, "-")}`;
    download.textContent = "Download attachment";
    box.append(element, download);
    parent.append(box);
}
function fullOutputLinks(card) {
    if (card.kind !== "tool_result") return;
    // Job metadata is published by the real projection. A handle is not a
    // browser path, and the server rechecks this card's Agent ownership.
    const handle = card.job?.handle;
    if (!Number.isSafeInteger(handle) || handle <= 0) return;
    const links = document.createElement("div");
    links.className = "artifact-links";
    const url = new URL("./jobs/output", document.baseURI);
    url.searchParams.set("agent", card.agent);
    url.searchParams.set("handle", handle);
    const download = document.createElement("button");
    download.type = "button";
    download.className = "btn btn-sm btn-outline-primary";
    download.textContent = "Download complete output";
    const notice = document.createElement("span");
    notice.setAttribute("role", "status");
    download.addEventListener("click", async () => {
        download.disabled = true;
        notice.textContent = "Reading completed Job output (16 MiB maximum)…";
        const controller = new AbortController();
        const limit = 16 * 1024 * 1024;
        try {
            const response = await fetch(url, {
                headers: { "X-Ensemble-Output": "1" },
                signal: controller.signal
            });
            if (!response.ok) {
                const refusal = await response.json().catch(() => null);
                throw new Error(refusal?.error || "This Agent no longer owns available output for this Job.");
            }
            if (Number(response.headers.get("Content-Length")) > limit) {
                throw new Error("Output exceeds the 16 MiB browser limit. Use the owned workspace file locally.");
            }
            // Do not use unbounded fetch().blob(). Enforce the cap even if a
            // response lacks Content-Length. The bounded Blob is released after
            // the download starts; this is not claimed to be streaming to disk.
            const reader = response.body.getReader();
            const chunks = [];
            let bytes = 0;
            while (true) {
                const chunk = await reader.read();
                if (chunk.done) break;
                bytes += chunk.value.byteLength;
                if (bytes > limit) {
                    throw new Error("Output exceeds the 16 MiB browser limit.");
                }
                chunks.push(chunk.value);
            }
            const local = URL.createObjectURL(new Blob(chunks, { type: "text/plain" }));
            const anchor = document.createElement("a");
            anchor.href = local;
            anchor.download = `job-${handle}-output.txt`;
            anchor.click();
            setTimeout(() => URL.revokeObjectURL(local), 1000);
            notice.textContent = `Downloaded ${bytes.toLocaleString()} complete bytes.`;
        } catch (error) {
            notice.textContent = error.message;
        } finally {
            controller.abort();
            download.disabled = false;
        }
    });
    links.append(download, notice);
    card.node.append(links);
}
function render(card) {
    const artifact = ["tool_call", "tool_result", "blob"].includes(card.kind);
    const pane = artifact ? toolTranscript : transcript;
    const nearBottom = pane.scrollHeight - pane.scrollTop - pane.clientHeight < 90;
    if (card.node.parentElement !== pane) pane.append(card.node);
    const oldDetails = card.node.querySelector("details");
    const wasOpen = oldDetails ? oldDetails.open : false;
    card.node.replaceChildren();
    card.node.className = "part";
    card.node.dataset.role = card.role;
    card.node.dataset.kind = card.kind;
    card.node.dataset.status = card.status;
    const heading = document.createElement("div");
    heading.className = "card-heading";
    const title = document.createElement("strong");
    const argument = card.kind === "tool_call" ? Object.values(card.input || {}).find(value => typeof value === "string") : null;
    const summary = argument ? ` · ${argument.slice(0, 100)}` : "";
    const kind = card.kind === "reasoning_summary" ? "Reasoning summary" : card.kind === "thinking" ? "Thinking" : card.kind;
    // Labels are projections of this card, not invented people or timestamps.
    // Keep wire-level details in a tooltip while prose reads as conversation.
    const author = card.role === "human" ? "You" : card.role === "assistant" ? "Assistant" : card.role;
    title.textContent = artifact ? card.name || card.toolName || kind
        : card.kind === "text" ? author : kind;
    title.title = `${card.role} · ${card.name || card.toolName || kind}${summary}${card.status ? " · " + card.status : ""}`;
    if (artifact && card.status) {
        const status = document.createElement("div");
        status.className = "tool-state";
        status.textContent = card.status;
        card.node.append(status);
    }
    const speaker = actionButton("Speak", () => speak(card));
    speaker.setAttribute("aria-label", `Speak ${card.role} ${card.kind}`);
    speaker.title = speaker.getAttribute("aria-label");
    setButtonIcon(speaker, "icon-speaker");
    speaker.dataset.action = "speak";
    speaker.disabled = !card.text.trim() || (!speech && !transcriptSpeech);
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
        summary.textContent = card.kind === "reasoning_summary" ? "Reasoning summary"
            : card.kind === "thinking" ? "Thinking" : "Expand complete received payload";
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
        if (artifact && card.kind !== "blob") {
            const preview = document.createElement("pre");
            preview.className = "preview";
            const clipped = card.text.slice(0, 1400).split("\n").slice(0, 18).join("\n");
            preview.textContent = clipped + (clipped.length < card.text.length ? "\n… (bounded preview)" : "");
            preview.hidden = wasOpen;
            details.ontoggle = () => { preview.hidden = details.open; };
            card.node.append(preview);
        }
        card.node.append(details);
    }
    for (const part of card.parts || []) media(card.node, part, card);
    fullOutputLinks(card);
    if (nearBottom) pane.scrollTop = pane.scrollHeight;
}
// State is acknowledged, not inferred from transcript content or tool activity.
function renderControls() {
    const connected = socket?.readyState === WebSocket.OPEN;
    const ready = connected && !selectionPending && acknowledgedState && !agentClosed;
    const busy = ["inflight", "toolspending"].includes(acknowledgedState);
    const paused = acknowledgedState === "paused";
    document.querySelector("#state").textContent = !connected ? "disconnected"
        : selectionPending ? "selecting" : acknowledgedState || "unknown";
    // Fixed slots avoid moving adjacent controls as activity changes. Pause
    // stays available during this tab's automatic pause so a person can pin it.
    document.querySelector("#pause").disabled = !ready || (paused && pauseIntent !== "automatic");
    document.querySelector("#resume").disabled = !ready || !paused;
    document.querySelector("#interrupt").disabled = !ready || !(busy || paused);
    // The actor classifies ordinary input when dequeued, even during a turn.
    document.querySelector("#send").disabled = !ready;
    document.querySelector("#clear-tools").disabled = !ready;
    document.querySelector("#model-choice").disabled = !ready || busy || paused;
    document.querySelector("#set-model").disabled = document.querySelector("#model-choice").disabled;
    document.querySelector("#theme-toggle").disabled = !connected || !settings;
}
// Format acknowledged Engine values only. No browser counters or price table.
function showUsage(frame) {
    document.querySelector("#funding-source").textContent = frame.funding || "Funding information unavailable.";
    const picker = document.querySelector("#model-choice");
    picker.replaceChildren();
    for (const model of frame.model_choices) {
        const option = document.createElement("option");
        option.value = model;
        option.textContent = model;
        picker.append(option);
    }
    picker.value = frame.selected_model;
    renderControls();
    const count = value => value == null ? "?" : value.toLocaleString("en-US");
    const latest = frame.usage;
    const totals = latest.totals;
    document.querySelector("#usage-cost").textContent = frame.total_cost_usd == null
        ? "?" : "$" + frame.total_cost_usd.toFixed(2);
    document.querySelector("#usage-all").textContent = count(latest.all_input);
    document.querySelector("#usage-cache").textContent = count(latest.cache_read_tokens);
    document.querySelector("#usage-hit").textContent = frame.hit_rate == null
        ? "?" : (frame.hit_rate * 100).toFixed(1) + "%";
    const text = `Session totals: input ${count(totals.input_tokens)}, cache write ${count(totals.cache_write_tokens)}, cache read ${count(totals.cache_read_tokens)}, output ${count(totals.output_tokens)}. Latest output ${count(latest.output_tokens)}.`;
    document.querySelector("#usage-totals").textContent = text;
    document.querySelector("#usage-meter").title = text;
}
function showAgents() {
    const picker = document.querySelector("#agent-choice");
    picker.replaceChildren();
    for (const row of agentRows) {
        const option = document.createElement("option");
        // Disabled unnamed rows must not share the initial view's empty value.
        if (row.key != null) option.value = row.key;
        option.disabled = row.key == null;
        const activity = [row.phase, row.tool, row.detail].filter(Boolean).join(" — ");
        const label = row.key === "" ? "Original unnamed conversation" : row.key ?? row.id;
        option.textContent = `${label} — ${row.paused ? "paused" : row.state}${activity ? " — " + activity : ""}`;
        picker.append(option);
    }
    // An unnamed borrowed host can still display its original conversation,
    // but does not acquire a fabricated global address in the tree.
    if (agentID === "" && !agentRows.some(row => row.key === "")) {
        const original = document.createElement("option");
        original.value = "";
        original.textContent = "Original unnamed conversation";
        picker.append(original);
    }
    picker.value = agentID ?? "";
    const selected = agentRows.find(row => row.key === agentID);
    const name = document.querySelector("#selected-name");
    name.textContent = selected?.id || agentID || "Original unnamed conversation";
    name.title = selected?.key || name.textContent;
    const conversation = document.querySelector("#conversation-agent");
    conversation.textContent = name.textContent;
    conversation.title = name.title;
    const input = document.querySelector("#prompt");
    input.placeholder = `Ask ${name.textContent} anything…`;
    input.setAttribute("aria-label", `Message to ${name.textContent}`);
    renderAgentTree();
}
function handle(frame) {
    if (frame.type === "agent_tree") {
        agentRows = frame.agents || [];
        showAgents();
        showGoal(agentRows.find(row => row.key === agentID)?.goal);
        return;
    }
    if (frame.type !== "replay" && frame.agent_id != null && frame.agent_id !== agentID) return;
    if (frame.type === "goal_status" || frame.type === "goal_updated") {
        if (goalSelectionPending) return;
        goalDirty = false;
        showGoal(frame.goal);
        if (frame.type === "goal_updated") notice("Goal applied for this session.");
        return;
    }
    if (frame.type === "usage") {
        showUsage(frame);
        return;
    }
    if (frame.type === "replay") {
        if (agentID !== frame.agent_id) {
            cancelSpeech(false);
            localPauseCause = false;
            pauseIntent = null;
        }
        agentID = frame.agent_id;
        selectionPending = false;
        acknowledgedState = null;
        renderControls();
        goalSelectionPending = false;
        goalDirty = false;
        showGoal(agentRows.find(row => row.key === agentID)?.goal);
        send("get_goal");
        const relayLink = document.querySelector("#relay-link");
        relayLink.hidden = !agentID;
        if (agentID) {
            const relayURL = new URL(location.href);
            relayURL.searchParams.set("agent_id", agentID);
            relayLink.href = relayURL.href;
        }
        showAgents();
        replaying = true;
        cards.clear();
        transcript.replaceChildren();
        toolTranscript.replaceChildren();
        for (const event of frame.data.events || []) handle(event);
        for (const part of frame.data.partials || []) handle(part);
        for (const agent of frame.data.agents || []) handle(agent);
        replaying = false;
        updatePauseCause();
        return;
    }
    if (frame.type === "current_settings" || frame.type === "settings_updated") {
        settings = frame.settings;
        applySettings();
        notice("");
        return;
    }
    if (frame.type === "agent_state" || frame.type === "state_changed") {
        const wasPaused = acknowledgedState === "paused";
        acknowledgedState = frame.state || frame.to;
        agentClosed = acknowledgedState === "shutdown";
        if (wasPaused && acknowledgedState !== "paused") pauseIntent = null;
        renderControls();
        updatePauseCause();
        return;
    }
    if (frame.type === "error" || frame.type === "lifecycle") {
        if (frame.type === "error" && selectionPending) {
            selectionPending = false;
            showAgents();
            renderControls();
        }
        if (frame.type === "error" && goalSelectionPending) {
            goalSelectionPending = false;
            showGoal(agentRows.find(row => row.key === agentID)?.goal);
        }
        // Provider failures arrive as durable lifecycle facts; control refusals
        // arrive as error frames. Turn completion is not a second error source.
        const isError = frame.type === "error" || frame.kind === "error_occurred";
        if (isError) {
            applySettings();
            if (!replaying && settings?.tts_enabled) {
                flushNarration();
                speakText("error", frame.message || "Request failed.", "error");
            }
        }
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
        card.parts = part.type === "tool_result" ? part.parts : [part];
        card.kind = part.type;
        card.role = part.role || "assistant";
        card.text = readable(part);
        card.name = part.name || card.name;
        card.input = part.args || card.input;
        card.final = true;
        if (card.kind === "tool_call" && !card.status) card.status = "planned";
    } else if (frame.type === "tool_dispatched") {
        card.kind = "tool_call";
        card.name = frame.name;
        card.input = frame.input;
        card.status = "started";
        card.text = `${frame.name}\n${JSON.stringify(frame.input || {}, null, 2)}`;
        card.final = true;
    } else if (frame.type === "tool_finished") {
        card.kind = "tool_result";
        card.parts = frame.parts || [];
        card.job = frame.job;
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
    const endpoint = new URL("./ws", document.baseURI);
    const relay = new URLSearchParams(location.search).get("agent_id");
    if (relay !== null) endpoint.searchParams.set("agent_id", relay);
    endpoint.protocol = location.protocol === "https:" ? "wss:" : "ws:";
    socket = new WebSocket(endpoint);
    const connection = socket;
    agentID = null;
    socket.onopen = () => {
        document.querySelector("#connection").textContent = "Connected";
        renderControls();
        send("subscribe");
        updatePauseCause();
    };
    socket.onmessage = event => {
        if (connection !== socket) return;
        try {
            const frame = JSON.parse(event.data);
            if (frame.type === "connection") {
                connection.relayAgentID = frame.relay_agent_id;
                document.querySelector("#relay-agent").textContent = frame.relay_agent_id || "Original unnamed Agent";
            } else if (frame.type === "mcp") handleMcp(connection, frame);
            else handle(frame);
        }
        catch { notice("An invalid display frame was refused."); }
    };
    socket.onclose = () => {
        if (connection !== socket) return;
        document.querySelector("#connection").textContent = "Disconnected — reconnecting";
        selectionPending = true;
        pauseIntent = null;
        localPauseCause = false;
        cancelSpeech(false);
        renderControls();
        showGoal(undefined);
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
    socket.send(JSON.stringify({ type, text, ...(agentID ? { agent_id: agentID } : {}) }));
    notice("");
    return true;
}
document.querySelector("#agent-choice").onchange = event => selectAgent(event.target.value);
function selectAgent(key) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        showAgents();
        notice("Not connected; selection was not sent.");
        return;
    }
    // Release this tab's old pause cause before asking the server to switch.
    // The next replay can apply a still-present typing cause to the new view.
    cancelSpeech(false);
    if (pauseIntent === "automatic") send("unpause");
    pauseIntent = null;
    localPauseCause = false;
    selectionPending = true;
    renderControls();
    goalSelectionPending = true;
    showGoal(undefined);
    socket.send(JSON.stringify({ type: "select_agent", agent_id: key }));
}
// One input path. Browser state must not classify a message as a hint.
function submit() {
    if (document.querySelector("#send").disabled) {
        notice("Wait for the selected Agent to connect; input was not sent.");
        return;
    }
    const input = document.querySelector("#prompt");
    if (input.value.trim() && send("prompt", input.value)) {
        input.value = "";
        humanTyping = false;
        updatePauseCause();
    }
}
document.querySelector("#compose").onsubmit = event => { event.preventDefault(); submit(); };
document.querySelector("#pause").onclick = () => {
    if (send("pause")) pauseIntent = "manual";
    renderControls();
};
document.querySelector("#interrupt").onclick = () => {
    if (send("interrupt")) pauseIntent = null;
};
document.querySelector("#clear-tools").onclick = () => send("clear_tools");
document.querySelector("#set-model").onclick = () => {
    send("set_model", document.querySelector("#model-choice").value);
};
document.querySelector("#resume").onclick = () => {
    if (send("unpause")) pauseIntent = null;
};
document.querySelector("#prompt").oninput = event => {
    humanTyping = !event.detail?.mcp && event.target.value.length > 0;
    updatePauseCause();
};
document.querySelector("#prompt").onkeydown = event => {
    if (event.key === "Escape") {
        event.preventDefault();
        cancelSpeech();
    } else if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault(); submit();
    }
};
document.querySelector("#stop-speech").onclick = () => cancelSpeech();
function applySettings() {
    if (!settings) return;
    document.documentElement.dataset.theme = settings.theme;
    const dark = settings.theme === "dark" || (settings.theme === "system" && matchMedia("(prefers-color-scheme: dark)").matches);
    document.documentElement.dataset.bsTheme = dark ? "dark" : "light";
    const themeButton = document.querySelector("#theme-toggle");
    setButtonIcon(themeButton, dark ? "icon-moon" : "icon-sun");
    themeButton.setAttribute("aria-label", dark ? "Switch to light theme" : "Switch to dark theme");
    themeButton.title = dark ? "Dark theme; switch to light theme" : "Light theme; switch to dark theme";
    document.documentElement.style.fontSize = `${settings.font_size}px`;
    renderControls();
    document.querySelector("#theme").value = settings.theme;
    document.querySelector("#font-size").value = settings.font_size;
    document.querySelector("#max-tool-rounds").value = settings.max_tool_rounds;
    document.querySelector("#context-target-bytes").value = settings.context_target_bytes;
    if (settings.recall) {
        const recall = settings.recall;
        document.querySelector("#recall-enabled").checked = !recall.disabled;
        document.querySelector("#recall-judge-enabled").checked = !recall.disable_judge;
        for (const field of ["judge_model", "min_score", "max_snippets", "min_query_length",
            "candidates", "context_messages", "max_recall_bytes"]) {
            document.querySelector(`#recall-${field}`).value = recall[field];
        }
    }
    for (const [band, policy] of Object.entries(settings.memory || {})) {
        document.querySelector(`#memory-${band}-enabled`).checked = !policy.disabled;
        for (const field of ["budget", "high_watermark", "low_watermark"]) {
            document.querySelector(`#memory-${band}-${field}`).value = policy[field];
        }
    }
    document.querySelector("#tts-speed").value = settings.tts_speed;
    document.querySelector("#speed-label").value = `${settings.tts_speed}×`;
    document.querySelector("#narrate").checked = settings.tts_enabled;
    document.querySelector(".speech-toggle").title = settings.tts_enabled
        ? "Speech on; stop automatically reading new messages" : "Speech off; enable reading new messages aloud";
    document.querySelector("#pause-speaking").checked = settings.pause_while_speaking;
    updatePauseCause(); // Disabling speech pausing does not cancel the queue.
}
function patchSettings(patch) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        notice("Not connected; settings were not saved.");
        applySettings();
        return;
    }
    socket.send(JSON.stringify({ type: "update_settings", settings: patch, agent_id: agentID }));
    // Keep a proposed edit while saving; only acknowledged settings affect policy.
    notice("Saving settings…");
}
// Status is an actor observation, not a browser timer or completion verdict.
function showGoal(goal) {
    const connected = socket?.readyState === WebSocket.OPEN;
    if (!connected || goalSelectionPending) goal = undefined;
    const labels = {
        unloaded: "Goals skill unloaded", unconfigured: "No goal configured",
        busy: "Working", scheduled: "Waiting for idle reminder", paused: "Paused",
        off: "Reminders off — goal retained", file_error: "Goal file error — reminders off"
    };
    const status = document.querySelector("#goal-status");
    if (goal) {
        const details = [labels[goal.phase] || goal.phase];
        if (!goal.enabled && (goal.phase === "busy" || goal.phase === "paused")) {
            details.push("reminders off");
        }
        if (goal.file) details.push(goal.file);
        if (goal.due_utc) details.push(`due ${goal.due_utc}`);
        if (goal.last_error) details.push(goal.last_error);
        status.textContent = `Goal: ${details.join(" — ")}`;
    } else if (!connected) {
        status.textContent = "Goal: disconnected — current state unknown";
    } else if (goalSelectionPending) {
        status.textContent = "Goal: switching Agent — waiting for status";
    } else {
        status.textContent = "Goal: waiting for Agent status";
    }
    document.querySelector("#goal-reload").disabled = !connected || goalSelectionPending;
    for (const id of ["goal-file", "goal-delay", "goal-enabled", "goal-apply", "goal-off", "goal-clear"]) {
        document.querySelector(`#${id}`).disabled = !goal?.available;
    }
    if (!goalDirty) {
        document.querySelector("#goal-file").value = goal?.file || "";
        document.querySelector("#goal-delay").value = goal?.idle_seconds ?? 480;
        document.querySelector("#goal-enabled").checked = goal?.enabled ?? false;
    }
}
function updateGoal(goal) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        notice("Not connected; goal was not changed.");
        return;
    }
    socket.send(JSON.stringify({type: "update_goal", agent_id: agentID, goal}));
    notice("Applying goal…");
}
document.querySelector("#goal-form").addEventListener("input", () => { goalDirty = true; });
document.querySelector("#goal-form").addEventListener("submit", event => {
    event.preventDefault();
    const seconds = Number(document.querySelector("#goal-delay").value);
    if (!Number.isFinite(seconds) || seconds < .001 || seconds > 4294967.294) {
        notice("Idle delay must be between .001 and 4294967.294 seconds.");
        return;
    }
    updateGoal({file: document.querySelector("#goal-file").value,
        enabled: document.querySelector("#goal-enabled").checked, idle_seconds: seconds});
});
document.querySelector("#goal-off").onclick = () => updateGoal({enabled: false});
document.querySelector("#goal-clear").onclick = () => updateGoal({file: ""});
document.querySelector("#goal-reload").onclick = () => send("get_goal");
document.querySelector("#memory-settings-form").addEventListener("submit", event => {
    event.preventDefault();
    const memory = {};
    for (const band of ["conversation", "soul", "memory", "archive64", "archive8", "session"]) {
        const policy = {disabled: !document.querySelector(`#memory-${band}-enabled`).checked};
        for (const field of ["budget", "high_watermark", "low_watermark"]) {
            const value = Number(document.querySelector(`#memory-${band}-${field}`).value);
            if (!Number.isSafeInteger(value) || value < 0) {
                notice("Memory sizes must be nonnegative whole bytes within the browser’s exact integer range.");
                return;
            }
            policy[field] = value;
        }
        memory[band] = policy;
    }
    patchSettings({memory});
});
document.querySelector("#memory-settings-reload").onclick = () => send("get_settings");
document.querySelector("#recall-settings-form").addEventListener("submit", event => {
    event.preventDefault();
    const recall = {
        disabled: !document.querySelector("#recall-enabled").checked,
        disable_judge: !document.querySelector("#recall-judge-enabled").checked,
        judge_model: document.querySelector("#recall-judge_model").value
    };
    for (const field of ["min_score", "max_snippets", "min_query_length", "candidates",
        "context_messages", "max_recall_bytes"]) {
        const value = Number(document.querySelector(`#recall-${field}`).value);
        if (!Number.isFinite(value) || value < 0 || (field !== "min_score" && !Number.isSafeInteger(value))) {
            notice("Recall sizes/counts must be nonnegative integers; score must be a finite nonnegative number.");
            return;
        }
        recall[field] = value;
    }
    patchSettings({recall});
});
document.querySelector("#recall-settings-reload").onclick = () => send("get_settings");
document.querySelector("#theme-toggle").onclick = () => patchSettings({theme: document.documentElement.dataset.bsTheme === "dark" ? "light" : "dark"});
document.querySelector("#theme").onchange = event => patchSettings({ theme: event.target.value });
document.querySelector("#font-size").onchange = event => patchSettings({ font_size: Number(event.target.value) });
document.querySelector("#max-tool-rounds").onchange = event => patchSettings({ max_tool_rounds: Number(event.target.value) });
document.querySelector("#context-target-bytes").onchange = event => patchSettings({ context_target_bytes: Number(event.target.value) });
document.querySelector("#tts-speed").onchange = event => patchSettings({ tts_speed: Number(event.target.value) });
document.querySelector("#narrate").onchange = event => patchSettings({ tts_enabled: event.target.checked });
document.querySelector("#pause-speaking").onchange = event => patchSettings({ pause_while_speaking: event.target.checked });
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
        if (!document.hidden && !transcriptSpeech) {
            // Recorded Chrome compatibility workaround, not a universal claim.
            const wakeup = new SpeechSynthesisUtterance("");
            wakeup.volume = 0;
            speech.speak(wakeup);
        }
    });
}
// The tree is only a view of the server projection. Selecting never changes the
// fixed relay target or creates an Agent. Arrow keys traverse visible rows.
function renderAgentTree() {
    const tree = document.querySelector("#agent-tree");
    const focused = tree.contains(document.activeElement) ? document.activeElement.dataset.agent : null;
    tree.replaceChildren();
    function add(row, group) {
        const children = row.key == null ? [] : agentRows.filter(child => child.parent === row.key);
        const item = document.createElement("li");
        item.setAttribute("role", "treeitem");
        item.setAttribute("aria-selected", String(row.key === agentID));
        const line = document.createElement("div");
        line.className = "tree-row";
        if (children.length) {
            const expanded = !collapsedAgents.has(row.key);
            item.setAttribute("aria-expanded", String(expanded));
            const toggle = actionButton(expanded ? "−" : "+", () => {
                if (expanded) collapsedAgents.add(row.key); else collapsedAgents.delete(row.key);
                renderAgentTree();
                tree.querySelector(`[data-agent="${CSS.escape(row.key)}"]`)?.focus();
            });
            toggle.classList.add("tree-expand");
            toggle.tabIndex = -1;
            toggle.setAttribute("aria-label", `${expanded ? "Collapse" : "Expand"} ${row.id}`);
            line.append(toggle);
        }
        const choose = actionButton("", () => selectAgent(row.key));
        choose.classList.add("tree-select");
        choose.disabled = row.key == null;
        choose.dataset.agent = row.key ?? row.id;
        choose.tabIndex = row.key === (focused ?? agentID) ? 0 : -1;
        choose.title = `${row.key || row.id} · ${row.paused ? "paused" : row.state} · ${[row.phase, row.tool, row.detail].filter(Boolean).join(" — ")}`;
        const name = document.createElement("span");
        name.className = "tree-name";
        name.textContent = row.id;
        const state = document.createElement("span");
        state.className = "tree-state";
        state.textContent = row.paused ? "paused" : row.state;
        choose.append(name, state);
        choose.onkeydown = event => {
            if (!["ArrowDown", "ArrowUp", "ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
            event.preventDefault();
            const visible = [...tree.querySelectorAll(".tree-select:not(:disabled)")].filter(button => button.offsetParent !== null);
            let next = visible.indexOf(choose);
            if (event.key === "ArrowDown") next = Math.min(next + 1, visible.length - 1);
            if (event.key === "ArrowUp") next = Math.max(next - 1, 0);
            if (event.key === "Home") next = 0;
            if (event.key === "End") next = visible.length - 1;
            if (event.key === "ArrowRight" && children.length) {
                if (collapsedAgents.delete(row.key)) {
                    renderAgentTree();
                    tree.querySelector(`[data-agent="${CSS.escape(row.key)}"]`)?.focus();
                    return;
                }
                next = Math.min(next + 1, visible.length - 1);
            }
            if (event.key === "ArrowLeft") {
                if (children.length && !collapsedAgents.has(row.key)) {
                    collapsedAgents.add(row.key);
                    renderAgentTree();
                    tree.querySelector(`[data-agent="${CSS.escape(row.key)}"]`)?.focus();
                    return;
                }
                const parent = visible.findIndex(button => button.dataset.agent === row.parent);
                if (parent >= 0) next = parent;
            }
            for (const button of visible) button.tabIndex = -1;
            visible[next].tabIndex = 0;
            visible[next].focus();
        };
        line.append(choose);
        item.append(line);
        if (children.length) {
            const nested = document.createElement("ul");
            nested.setAttribute("role", "group");
            nested.hidden = collapsedAgents.has(row.key);
            for (const child of children) add(child, nested);
            item.append(nested);
        }
        group.append(item);
    }
    for (const row of agentRows.filter(row => row.parent == null)) add(row, tree);
    if (!tree.querySelector('[tabindex="0"]')) tree.querySelector(".tree-select:not(:disabled)")?.setAttribute("tabindex", "0");
    if (focused != null) tree.querySelector(`[data-agent="${CSS.escape(focused)}"]`)?.focus();
}

connect();
