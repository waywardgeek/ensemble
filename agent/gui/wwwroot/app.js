"use strict";
// The browser owns presentation only. Replay replaces derived cards; stable IDs
// make finals, repeated frames, and metadata updates idempotent.
const transcript = document.querySelector("#transcript");
const cards = new Map();
let localErrorSequence = 0;
const toolTranscript = document.querySelector("#tool-transcript");
let jobSessions = null;
import(new URL("./job-sessions.js", document.currentScript.src).href).then(module => {
    jobSessions = module.createJobSessions(toolTranscript, fetch, () => {
        for (const card of cards.values()) if (card.kind === "tool_result") queueCardRender(card);
    });
    if (socket?.readyState === WebSocket.OPEN && !selectionPending)
        for (const card of cards.values()) if (card.kind === "tool_result") jobSessions.observe(card);
}).catch(() => { /* Recorded cards remain usable without continuing views. */ });
let acknowledgedState = null;
let workingContext = null;
let workingContextForm = null;
let selectionPending = true;
const collapsedAgents = new Set();
let socket;
let agentID = null; // Display identity changes only on admitted replay.
let agentRows = []; // Bounded server tree projection, not a client Agent registry.
let reconnectTimer;
let replaying = false;
let activeSpeech = null;
let speechWatchTimer;
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
// Origin is a fact, not something speech callbacks can infer from field contents.
let humanTyping = false;
let agentClosed = false;
// Last acknowledged server snapshot. Controls request changes; they never
// become a second preference store. Theme and speech belong to this browser.
let settings = null;
let goalSelectionPending = false;
let goalDirty = false; // Unsaved form edits only, never a second live configuration.
const speech = window.speechSynthesis;
const notice = text => { document.querySelector("#notice").textContent = text; };
let pageCapture = null;
let captureConsent = null;
let voiceControl = null;
let captureDraft = null; // One gesture, original socket and selected Agent.
const imageComposer = new ImageComposer(document.querySelector("#prompt"),
    document.querySelector("#image-previews"), document.querySelector("#image-status"),
    document.querySelector("#discard-image-draft"), () => {
        resizeComposer();
        humanTyping = document.querySelector("#prompt").value.length > 0;
        renderControls();
        updatePauseCause();
    }, notice);

// Preload before enabling the camera. Awaiting an import in its click handler
// would lose the transient browser activation required by getDisplayMedia.
const pageCaptureReady = import("./page-capture.js").then(({ PageCapture }) => {
    pageCapture = new PageCapture();
    renderControls();
    return pageCapture;
}).catch(() => { notice("Page capture could not load. You can still attach an image."); return null; });
const captureConsentReady = Promise.all([pageCaptureReady, import("./capture-consent.js")])
    .then(([capture, { CaptureConsent }]) => {
        if (!capture) return null;
        captureConsent = new CaptureConsent(capture, document.querySelector("#capture-consent"));
        captureConsent.addEventListener("change", renderControls);
        return captureConsent;
    }).catch(() => null);
async function requestBrowserCapture(connection, target, id, args) {
    if (!args || typeof args !== "object" || Array.isArray(args) || Object.keys(args).length)
        throw new Error("Page capture takes no arguments.");
    const current = () => socket === connection && connection.readyState === WebSocket.OPEN
        && connection.mcpInitialized && connection.relayAgentID === target
        && agentID === target && !selectionPending && !agentClosed;
    if (!current()) throw new Error("Select the requesting Agent before asking for capture.");
    const consent = await captureConsentReady;
    if (!current()) throw new Error("Capture request is no longer current.");
    if (!consent) throw new Error("Capture confirmation is unavailable; attach a manual screenshot.");
    if (imageComposer.pending || captureDraft) throw new Error("Finish the current image send or capture first.");
    const label = document.querySelector("#agent-choice").selectedOptions[0]?.textContent || target;
    return consent.request(connection, target, id, label);
}
function cancelBrowserCapture(connection, target, id) {
    return captureConsent?.cancelMatching(connection, target, id) ?? Promise.resolve();
}
function closeBrowserCapture(connection) {
    return captureConsent?.cancelConnection(connection) ?? Promise.resolve();
}
function cancelPageCapture(message) {
    captureConsent?.cancel(message);
    if (!captureDraft) return;
    captureDraft = null; // Invalidate draft admission before stopping native work.
    pageCapture.cancel().then(renderControls);
    if (message) notice(message);
    renderControls();
}
async function capturePage(event) {
    if (captureDraft) {
        cancelPageCapture("Page capture cancelled. Any late image will be discarded.");
        return;
    }
    if (document.querySelector("#capture-page").disabled) return;
    if (!event?.isTrusted) {
        notice("Please press the camera yourself to authorize capture.");
        return;
    }
    const admission = { socket, agent: agentID };
    captureDraft = admission;
    try {
        // Call before awaiting anything: the native picker requires this click.
        const result = pageCapture.capture();
        notice("Choose this Ensemble tab. Press the camera again to cancel.");
        renderControls();
        const captured = await result;
        if (captureDraft !== admission || socket !== admission.socket
            || socket.readyState !== WebSocket.OPEN || selectionPending
            || agentID !== admission.agent || agentClosed || imageComposer.pending) return;
        imageComposer.addFiles([captured.file], admission.agent);
        if (imageComposer.entries.some(entry => entry.file === captured.file))
            notice("Screenshot added to the draft. Send when ready.");
    } catch (error) {
        if (captureDraft === admission)
            notice(error?.name === "AbortError" ? "Page capture cancelled." : String(error?.message || "Page capture failed."));
    } finally {
        if (captureDraft === admission) captureDraft = null;
        renderControls();
    }
}
document.querySelector("#capture-page").onclick = capturePage;

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
    if (part.type === "text" || part.type === "thinking" || part.type === "reasoning_summary" || part.type === "handoff" || part.type === "recall") return part.text || "";
    if (part.type === "tool_call") return `${part.name || "Tool"}\n${JSON.stringify(part.args || {}, null, 2)}`;
    if (part.type === "tool_result") return (part.parts || []).map(readable).join("\n");
    if (part.type === "blob") return `[${part.mime || "Media"} attachment]`;
    if (part.type === "redacted") return "[Redacted content]";
    return ""; // Never speak provider signatures or hidden opaque metadata.
}

// Send only local activity edges. A different device can resume the Agent;
// unchanged activity here must not reassert its previous pause.
function updatePauseCause() {
    if (agentClosed || selectionPending || replaying || !acknowledgedState ||
        socket?.readyState !== WebSocket.OPEN) return;
    const speaking = activeSpeech !== null || speechQueue.length > 0;
    const cause = humanTyping || (browserSpeech.value.pause && speaking);
    if (cause === localPauseCause) return;
    if (!send(cause ? "pause" : "unpause")) return;
    localPauseCause = cause;
    recordSpeech(cause ? "pause" : "resume", { cause: humanTyping ? "typing" : speaking ? "speech" : "clear" });
    renderControls();
}
// Native speech may stop without an end/error callback (notably around browser
// focus/visibility changes). Observe its state, never estimate reading duration.
// Allow startup time because speak() can return before native state is published.
function checkSpeechCompletion() {
    const active = activeSpeech;
    if (!active || performance.now() - active.admittedAt < 2000) return;
    if (speech.speaking || speech.pending) return;
    recordSpeech("completion_recovered", { part_id: active.id });
    active.settled();
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
        utterance.rate = browserSpeech.value.speed;
        utterance.voice = speech.getVoices().find(voice => voice.voiceURI === selected) || null;
        const settled = () => {
            if (activeSpeech?.utterance !== utterance) return;
            clearInterval(speechWatchTimer);
            activeSpeech = null;
            startSpeech(); // Both completion and interruption advance the queue.
        };
        activeSpeech = { id: item.id, utterance, admittedAt: performance.now(), settled };
        speechWatchTimer = setInterval(checkSpeechCompletion, 500);
        utterance.onend = settled;
        utterance.onerror = settled;
        updatePauseCause();
        try { speech.speak(utterance); }
        catch { settled(); notice("Speech could not be played."); }
        return;
    }
    updatePauseCause();
}
function enqueueSpeech(id, text, source = "explicit", announceThinking = false) {
    if (!text.trim()) return false;
    if (!transcriptSpeech && (!speech || !window.SpeechSynthesisUtterance)) return false;
    const thinking = source === "reasoning_summary" || source === "thinking";
    const label = thinking && announceThinking ? "Thinking: "
        : source === "error" ? "Error: " : "";
    speechQueue.push({ id, text: label + text, source });
    updatePauseCause();
    startSpeech();
    return true;
}
// Manual artifact reading and MCP speech use the same filter as narration.
function speakText(id, text, source = "explicit") {
    let announced = false;
    for (const phrase of speechPhrases(text, 0, true).phrases) {
        if (enqueueSpeech(id, phrase, source, !announced)) announced = true;
    }
}
function cancelSpeech(update = true) {
    clearInterval(speechWatchTimer);
    activeSpeech = null;
    speechQueue.length = 0;
    speech?.cancel();
    if (update) updatePauseCause();
}
function stopSpeechAndResume() {
    cancelSpeech(false);
    // Preserve the draft and its edge baseline, but explicitly resume from any device.
    localPauseCause = humanTyping;
    if (!selectionPending && !agentClosed) send("unpause");
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
    if (selectionPending || replaying || !browserSpeech.value.enabled) {
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
    // A provider part is one passage even when its sentences arrive in many
    // deltas. Keep its announcement beside its cursor, not on each queue item.
    for (const phrase of next.phrases) {
        if (enqueueSpeech(card.id, phrase, card.kind, !card.speechAnnounced)) {
            card.speechAnnounced = true;
        }
    }
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
function receivedTail(text) {
    const bytes = new TextEncoder().encode(text);
    let start = Math.max(0, bytes.length - 65536);
    while (start < bytes.length && (bytes[start] & 0xc0) === 0x80) start++;
    return new TextDecoder().decode(bytes.subarray(start));
}
function completePresentation(card) {
    if (card.directoryListing) return card.directoryListing.entries.length === card.directoryListing.total_entries;
    if (card.searchEvidence) return card.searchEvidence.lines.length === card.searchEvidence.total_lines;
    if (card.fileEvidence) {
        const segments = [card.fileEvidence.content, card.fileEvidence.removed].filter(Boolean);
        return segments.every(segment => new TextEncoder().encode(segment.text).length === segment.total_bytes);
    }
    return false;
}
// Human output is independent of the immutable model observation in card.text.
function loadResultWindow(card) {
    // Continuing commands have one fetch owner, including late module startup.
    if (jobSessions?.owns(card)) { card.windowController?.abort(); return; }
    if (card.windowRequested || card.kind !== "tool_result" || card.status === "error" || !Number.isSafeInteger(card.job?.handle)) return;
    if (completePresentation(card)) return;
    card.windowRequested = true;
    const controller = new AbortController();
    card.windowController = controller;
    const deadline = setTimeout(() => controller.abort(), 10000);
    const url = new URL("./jobs/window", document.baseURI);
    url.searchParams.set("agent", card.agent);
    url.searchParams.set("handle", card.job.handle);
    (async () => {
        try {
            const response = await fetch(url, { headers: { "X-Ensemble-Output": "1" },
                signal: controller.signal, credentials: "same-origin", redirect: "error", cache: "no-store" });
            if (!response.ok) throw new Error("Unavailable");
            const value = await response.json();
            if (!currentCard(card)) return;
            if (value.handle !== card.job.handle || value.encoding !== "base64" ||
                !Number.isSafeInteger(value.start) || !Number.isSafeInteger(value.end) ||
                value.start < 0 || value.end < value.start || value.end - value.start > 65536 ||
                typeof value.data !== "string" || value.data.length > 87384 ||
                value.rolled_out !== (value.start > 0) ||
                (card.job.origin_call_id && value.origin_call_id !== card.job.origin_call_id)) throw new Error("Invalid window");
            const bytes = Uint8Array.from(atob(value.data), c => c.charCodeAt(0));
            if (bytes.length !== value.end - value.start) throw new Error("Invalid length");
            let start = 0;
            if (value.rolled_out) while (start < bytes.length && (bytes[start] & 0xc0) === 0x80) start++;
            card.windowText = new TextDecoder().decode(bytes.subarray(start));
            card.windowNotice = value.rolled_out ? "Latest 64 KiB; earlier output remains on disk." : "";
        } catch {
            if (!currentCard(card) || controller.signal.aborted) return;
            card.windowNotice = "Retained output unavailable; showing the recorded observation.";
        } finally {
            clearTimeout(deadline);
            if (card.windowController === controller) card.windowController = null;
            if (currentCard(card) && !controller.signal.aborted) queueCardRender(card);
        }
    })();
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
    download.setAttribute("aria-label", "Download complete output");
    download.title = "Download complete output";
    download.dataset.action = "download-output";
    setButtonIcon(download, "icon-download");
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
    card.node.querySelector(".card-heading").insertBefore(download, card.node.querySelector('[data-action="speak"]'));
    links.append(notice);
    card.node.append(links);
}
function workProductLinks(card) {
    const offer = card.workProduct;
    if (card.kind !== "tool_result" || card.status === "error" || !offer
        || typeof offer.id !== "string" || !/^[a-f0-9]{32}$/i.test(offer.id)
        || typeof offer.name !== "string" || !offer.name.trim() || /[\x00-\x1f\x7f/\\]/.test(offer.name)
        || !Number.isSafeInteger(offer.offered_bytes) || offer.offered_bytes < 0
        || offer.offered_bytes > 64 * 1024 * 1024) return;
    const box = document.createElement("div");
    box.className = "work-product-links";
    const button = document.createElement("button");
    button.type = "button";
    button.className = "btn btn-sm btn-outline-primary";
    const label = "Download " + offer.name;
    button.textContent = label;
    const policy = document.createElement("small");
    policy.textContent = `${offer.offered_bytes.toLocaleString()} bytes when offered · current contents on click · 64 MiB limit · expires on restart`;
    const notice = document.createElement("span");
    notice.setAttribute("role", "status");
    const url = new URL("./work-products/download", document.baseURI);
    url.searchParams.set("agent", card.agent);
    url.searchParams.set("offer", offer.id);
    button.addEventListener("click", async () => {
        if (card.workProductController) { card.workProductController.abort(); return; }
        if (selectionPending || !currentCard(card)) return;
        const controller = new AbortController();
        card.workProductController = controller;
        button.textContent = "Cancel download";
        notice.textContent = "Downloading current file contents…";
        try {
            const response = await fetch(url, {
                headers: { "X-Ensemble-Work-Product": "1" }, signal: controller.signal
            });
            if (!response.ok) throw new Error(response.status === 403
                ? "Download access refused."
                : "Offer expired, unavailable, or exceeds the 64 MiB limit.");
            const limit = 64 * 1024 * 1024;
            if (Number(response.headers.get("Content-Length")) > limit)
                throw new Error("Work product exceeds the 64 MiB download limit.");
            const reader = response.body.getReader();
            const chunks = [];
            let bytes = 0;
            while (true) {
                const chunk = await reader.read();
                if (chunk.done) break;
                bytes += chunk.value.byteLength;
                if (bytes > limit) throw new Error("Work product exceeds the 64 MiB download limit.");
                chunks.push(chunk.value);
            }
            if (controller.signal.aborted || !currentCard(card))
                throw new DOMException("Offer view changed", "AbortError");
            const local = URL.createObjectURL(new Blob(chunks, { type: "application/octet-stream" }));
            const anchor = document.createElement("a");
            anchor.href = local;
            anchor.download = offer.name;
            anchor.click();
            setTimeout(() => URL.revokeObjectURL(local), 1000);
            notice.textContent = `Downloaded ${bytes.toLocaleString()} bytes.`;
        } catch (error) {
            if (currentCard(card)) notice.textContent = error.name === "AbortError"
                ? "Download cancelled." : "Download failed: " + error.message;
        } finally {
            controller.abort();
            if (card.workProductController === controller) card.workProductController = null;
            button.textContent = label;
        }
    });
    box.append(button, policy, notice);
    card.node.append(box);
}
// Coalesce streamed chunks without delaying authoritative text or speech.
const pendingCardRenders = new Set();
let cardRenderTimer = null;
function cardHasSelection(card) {
    const selection = window.getSelection();
    if (!selection || selection.isCollapsed) return false;
    for (let index = 0; index < selection.rangeCount; index++) {
        if (selection.getRangeAt(index).intersectsNode(card.node)) return true;
    }
    return false;
}
function flushCardRenders() {
    cardRenderTimer = null;
    for (const card of pendingCardRenders) {
        if (cardHasSelection(card)) continue;
        pendingCardRenders.delete(card);
        render(card);
    }
}
function queueCardRender(card) {
    if (cards.get(card.id) !== card || card.agent !== agentID) return;
    pendingCardRenders.add(card);
    if (cardRenderTimer === null) cardRenderTimer = setTimeout(flushCardRenders, 50);
}
document.addEventListener("selectionchange", () => {
    if (pendingCardRenders.size && cardRenderTimer === null)
        cardRenderTimer = setTimeout(flushCardRenders, 50);
});
function toolObservationLabel(card) {
    // These are call/observation facts, never claims about current execution.
    switch (card.status) {
        case "planned": return "Requested";
        case "started": return "Dispatched · awaiting result";
        case "finished": return "Result returned";
        case "error": return "Error reported";
        default: return "Observation state unavailable";
    }
}
function recordedJobLabel(job) {
    if (!Number.isSafeInteger(job?.handle) || job.handle <= 0) return "";
    // A retained snapshot is not a live Job monitor. Replay must not imply that
    // an old running observation still describes the process now.
    let state = "unavailable";
    switch (job.status) {
        case "running": state = "running when observed"; break;
        case "done": state = "completed"; break;
        case "killed": state = "killed"; break;
    }
    const exit = Number.isSafeInteger(job.exit_code) ? ` · exit code ${job.exit_code}` : "";
    return `Job ${job.handle} ${state}${exit}`;
}
function currentCard(card) {
    return card && cards.get(card.id) === card && card.agent === agentID && card.node.isConnected;
}
function renderPairLink(card) {
    const target = card.kind === "tool_call" ? card.latestResult : card.callCard;
    let button = card.node.querySelector('[data-action="view-pair"]');
    // The result may be assigned before its first DOM insertion. Keep the
    // existing control through that update; clicks still require connected cards.
    if (!target || cards.get(target.id) !== target || target.agent !== card.agent) {
        button?.remove();
        return;
    }
    if (!button) {
        button = actionButton("", () => {
            // Resolve the actual relationship on click; replay invalidates old
            // objects even when the new transcript reuses the same wire IDs.
            const counterpart = card.kind === "tool_call" ? card.latestResult : card.callCard;
            if (!currentCard(card) || !currentCard(counterpart)) return;
            const observations = counterpart.node.closest(".job-observations");
            if (observations) observations.open = true;
            counterpart.node.tabIndex = -1;
            counterpart.node.focus({ preventScroll: true });
            counterpart.node.scrollIntoView({ block: "nearest" });
        });
        button.dataset.action = "view-pair";
        setButtonIcon(button, "icon-pair");
        const heading = card.node.querySelector(".card-heading");
        heading.insertBefore(button, heading.querySelector('[data-action="speak"]'));
    }
    button.title = card.kind === "tool_call" ? "View latest result" : "View call";
    button.setAttribute("aria-label", button.title);
}
function renderRecordedTiming(card) {
    let stamp = card.node.querySelector(":scope > .recorded-time");
    const parsed = typeof card.recordedAt === "string" ? new Date(card.recordedAt) : null;
    if (parsed && Number.isFinite(parsed.getTime())) {
        if (!stamp) {
            stamp = document.createElement("time");
            stamp.className = "recorded-time";
            card.node.append(stamp);
        }
        stamp.dateTime = card.recordedAt;
        stamp.textContent = parsed.toLocaleTimeString();
        stamp.title = "Finalized fact recorded at " + card.recordedAt;
        stamp.setAttribute("aria-label", stamp.title);
    } else stamp?.remove();
    let elapsed = card.node.querySelector(":scope > .observation-duration");
    if (card.kind === "tool_result" && Number.isFinite(card.observationMs) && card.observationMs >= 0) {
        if (!elapsed) {
            elapsed = document.createElement("div");
            elapsed.className = "observation-duration";
            card.node.append(elapsed);
        }
        elapsed.textContent = (parsed && Number.isFinite(parsed.getTime()) ? " · " : "") + (card.observationMs / 1000).toLocaleString(undefined,
            { maximumFractionDigits: 3 }) + " s";
        elapsed.title = "Dispatch to returned observation: " + card.observationMs + " ms; not background Job runtime";
    } else elapsed?.remove();
}
function render(card) {
    // Timers and Job callbacks may outlive a replay or display-window eviction.
    // They must never reinsert a former Agent's card into the current pane.
    if (cards.get(card.id) !== card || card.agent !== agentID) return;
    if (cardHasSelection(card)) { queueCardRender(card); return; }
    pendingCardRenders.delete(card);
    const artifact = ["tool_call", "tool_result", "blob"].includes(card.kind);
    const pane = artifact ? toolTranscript : transcript;
    const savedScrollTop = pane.scrollTop;
    const reading = pane.contains(document.activeElement) || [...cards.values()].some(
        item => pane.contains(item.node) && cardHasSelection(item));
    const nearBottom = !reading && pane.scrollHeight - pane.scrollTop - pane.clientHeight < 90;
    const container = card.kind === "tool_result" ? jobSessions?.container(card) || pane : pane;
    if (card.node.parentElement !== container) container.append(card.node);
    const prose = ["text", "thinking", "reasoning_summary"].includes(card.kind);
    const keepBody = prose && card.node.querySelector(":scope > .message");
    if (!keepBody && !artifact) card.node.replaceChildren();
    card.node.className = "part";
    card.node.dataset.role = card.role;
    card.node.dataset.kind = card.kind;
    card.node.dataset.status = card.status;
    // Retain stream identity/order and original text, but do not reserve space
    // for thinking openings or opaque summaries with no readable content.
    card.node.hidden = ["thinking", "reasoning_summary"].includes(card.kind) && !card.text.trim();
    const heading = card.node.querySelector(":scope > .card-heading") || document.createElement("div");
    heading.className = "card-heading";
    const title = heading.querySelector("strong") || document.createElement("strong");
    const argument = card.kind === "tool_call" ? Object.values(card.input || {}).find(value => typeof value === "string") : null;
    const summary = argument ? ` · ${argument.slice(0, 100)}` : "";
    const kind = card.kind === "reasoning_summary" || card.kind === "thinking" ? "Thinking"
        : card.kind === "error" ? "Error" : card.kind;
    // Labels are projections of this card, not invented people or timestamps.
    // Keep wire-level details in a tooltip while prose reads as conversation.
    const author = card.role === "human" ? (settings?.human_name || "Human")
        : card.role === "assistant" ? (agentRows.find(row => row.key === card.agent)?.id || card.agent || "Agent") : card.role;
    title.textContent = artifact ? card.name || card.toolName || kind
        : card.kind === "text" ? author : kind;
    const statusLabel = artifact && card.status ? toolObservationLabel(card) : card.status;
    title.title = `${author} · ${card.name || card.toolName || kind}${summary}${statusLabel ? " · " + statusLabel : ""}`;
    card.node.querySelectorAll(":scope > .tool-state").forEach(node => node.remove());
    if (card.kind === "tool_result") {
        const jobLabel = recordedJobLabel(card.job);
        const resultLabel = card.toolName ? `${card.toolName} result` : toolObservationLabel(card);
        title.textContent = card.toolName
            ? resultLabel + (jobLabel ? ` · ${jobLabel}` : "") + (card.status === "error" ? " · error" : "")
            : (jobLabel || resultLabel) + (card.status === "error" && card.job ? " · error" : "");
    }
    const speaker = heading.querySelector('[data-action="speak"]') || actionButton("Speak", () => speak(card));
    speaker.setAttribute("aria-label", `Speak ${author} ${card.kind}`);
    speaker.title = speaker.getAttribute("aria-label");
    if (!speaker.dataset.action) setButtonIcon(speaker, "icon-speaker");
    speaker.dataset.action = "speak";
    speaker.disabled = !card.text.trim() || (!speech && !transcriptSpeech);
    // Playback belongs with the speaker label, not in the prose flow.
    if (!title.parentElement) heading.append(title);
    if (!speaker.parentElement) heading.append(speaker);
    if (!heading.parentElement) card.node.append(heading);
    // Supplied thinking is readable prose, not a hidden tool payload. Keep its
    // original kind/provenance; the friendly label does not invent hidden reasoning.
    if (card.kind === "error") {
        // Diagnostics are visible literal text, never provider HTML or a hidden payload.
        const body = document.createElement("div");
        body.className = "message";
        body.style.whiteSpace = "pre-wrap";
        body.textContent = card.text;
        card.node.append(body);
    } else if (prose) {
        const body = keepBody || document.createElement("div");
        body.className = "message";
        const proseRenderer = window.EnsembleArtifactRenderers?.prose;
        try {
            if (typeof proseRenderer !== "function") throw new Error("Prose renderer unavailable");
            proseRenderer(body, card.text, card.kind === "text" && card.role === "assistant");
        } catch {
            // Presentation failure never changes recorded text or speech. This
            // literal fallback is outside the customizable renderer chain.
            body.replaceChildren(document.createTextNode(card.text));
        }
        if (!body.parentElement) card.node.append(body);
    } else {
        const result = card.kind === "tool_result";
        const sessionOutput = result ? jobSessions?.output(card) : null;
        let details = card.node.querySelector(":scope > details");
        if (!details) {
            details = document.createElement("details");
            const summary = document.createElement("summary");
            summary.textContent = "Expand complete received payload";
            details.append(summary, document.createElement("pre"));
            card.node.append(details);
        }
        details.hidden = result && !sessionOutput;
        details.querySelector("summary").textContent = result ? "Recorded observation" : "Expand complete received payload";
        const pre = details.querySelector("pre:not(.edit-change)");
        if (pre.textContent !== card.text) pre.textContent = card.text;
        pre.className = card.toolName === "search_files" ? "search-results" : "tool-output";
        if (card.change && !details.querySelector(".edit-change")) {
            const label = document.createElement("p");
            label.textContent = "Applied exact replacement (not a whole-file diff)";
            const change = document.createElement("pre");
            change.className = "edit-change";
            change.textContent = "- " + card.change.old_text.split("\n").join("\n- ")
                + "\n+ " + card.change.new_text.split("\n").join("\n+ ");
            details.insertBefore(label, pre);
            details.insertBefore(change, pre);
        }
        if (artifact && card.kind !== "blob") {
            let preview = card.node.querySelector(":scope > .preview");
            if (!preview) {
                preview = document.createElement("div");
                preview.className = "preview";
                card.node.insertBefore(preview, details);
                details.ontoggle = () => { preview.hidden = details.open; };
            }
            const clipped = card.text.slice(0, 1400).split("\n").slice(0, 18).join("\n");
            const previewText = result ? (jobSessions?.owns(card) ? receivedTail(card.text) : (card.windowText ?? receivedTail(card.text))) : clipped + (clipped.length < card.text.length ? "\n… (bounded preview)" : "");
            if (sessionOutput) {
                if (card.previewSessionOutput !== sessionOutput) {
                    const link = document.createElement("button");
                    link.type = "button";
                    link.className = "btn btn-sm btn-outline-secondary session-output-link";
                    link.textContent = `View Job ${card.job.handle} output`;
                    link.onclick = () => {
                        if (!currentCard(card) || jobSessions?.output(card) !== sessionOutput) return;
                        sessionOutput.focus({ preventScroll: true });
                        sessionOutput.scrollIntoView({ block: "nearest" });
                    };
                    preview.replaceChildren(link);
                    card.previewSessionOutput = sessionOutput;
                }
            } else {
                const renderer = card.kind === "tool_result" && card.status !== "error"
                    ? (window.EnsembleToolRenderers?.[card.toolName] ||
                        (["list_directory", "read_file", "search_files", "write_file", "edit_file"].includes(card.toolName)
                            ? window.EnsembleArtifactRenderers?.[card.toolName] : null)) : null;
                const defaultRenderer = window.EnsembleArtifactRenderers?.tool_default;
                if (card.previewSessionOutput || card.previewText !== previewText || card.previewDefaultRenderer !== defaultRenderer || card.previewRenderer !== renderer || card.previewDirectoryListing !== card.directoryListing || card.previewFileEvidence !== card.fileEvidence || card.previewSearchEvidence !== card.searchEvidence) {
                    let content = null;
                    if (typeof renderer === "function") {
                        try { content = renderer(document, previewText, card.directoryListing, card.fileEvidence, card.searchEvidence); }
                        catch { /* A presentation failure must not hide the recorded result. */ }
                    }
                    if ((!(content instanceof HTMLElement) || content.isConnected) && typeof defaultRenderer === "function") {
                        try { content = defaultRenderer(document, previewText); }
                        catch { /* The final literal fallback remains independent of custom code. */ }
                    }
                    if (!(content instanceof HTMLElement) || content.isConnected) {
                        content = document.createElement("pre");
                        content.textContent = previewText;
                    }
                    preview.replaceChildren(content);
                    if (card.windowText !== undefined && !completePresentation(card) &&
                        (card.directoryListing || card.fileEvidence || card.searchEvidence)) {
                        const label = document.createElement("p");
                        label.textContent = "Retained output tail (plain text)";
                        const tail = document.createElement("pre");
                        tail.className = "tool-output";
                        tail.textContent = card.windowText;
                        preview.append(label, tail);
                    }
                    card.previewSessionOutput = null;
                    card.previewText = previewText;
                    card.previewRenderer = renderer;
                    card.previewDefaultRenderer = defaultRenderer;
                    card.previewDirectoryListing = card.directoryListing;
                    card.previewFileEvidence = card.fileEvidence;
                    card.previewSearchEvidence = card.searchEvidence;
                }
            }
            preview.hidden = result ? false : details.open;
            preview.tabIndex = 0;
            preview.setAttribute("aria-label", sessionOutput ? "Continuing output reference" : "Scrollable tool output");
            if (result) {
                let windowNotice = card.node.querySelector(":scope > .output-window-notice");
                if (!windowNotice) {
                    windowNotice = document.createElement("small");
                    windowNotice.className = "output-window-notice";
                    card.node.insertBefore(windowNotice, details);
                }
                windowNotice.textContent = card.windowNotice || (new TextEncoder().encode(card.text).length > 65536 && card.windowText === undefined ? "Latest 64 KiB of the received observation; retained output not yet available." : "");
                windowNotice.hidden = !!jobSessions?.owns(card) || !windowNotice.textContent;
                loadResultWindow(card);
            }
        }
    }
    // Passive attachments and the owned-output control are separate from raw
    // detail. A call status update must not replace a focused summary or an
    // in-flight download button. Parts are immutable projection values.
    if (card.renderedParts !== card.parts) {
        card.node.querySelector(":scope > .artifact-media")?.remove();
        const attachments = document.createElement("div");
        attachments.className = "artifact-media";
        for (const part of card.parts || []) media(attachments, part, card);
        if (attachments.childNodes.length) card.node.append(attachments);
        card.renderedParts = card.parts;
    }
    if (card.renderedJobHandle !== card.job?.handle) {
        card.node.querySelector(":scope > .artifact-links")?.remove();
        card.node.querySelector('[data-action="download-output"]')?.remove();
        fullOutputLinks(card);
        card.renderedJobHandle = card.job?.handle;
    }
    if (card.renderedWorkProductID !== card.workProduct?.id) {
        card.workProductController?.abort();
        card.node.querySelector(":scope > .work-product-links")?.remove();
        workProductLinks(card);
        card.renderedWorkProductID = card.workProduct?.id;
    }
    renderRecordedTiming(card);
    if (artifact) renderPairLink(card);

    pane.scrollTop = nearBottom ? pane.scrollHeight : savedScrollTop;
}
// State is acknowledged, not inferred from transcript content or tool activity.
function renderControls() {
    const connected = socket?.readyState === WebSocket.OPEN;
    const ready = connected && !selectionPending && acknowledgedState && !agentClosed;
    voiceControl?.bind(socket, agentID, !!ready, !!document.querySelector("#prompt").value.trim()
        || imageComposer.hasImages || imageComposer.blocked || !!pageCapture?.busy || !!captureDraft || !!captureConsent?.busy);
    const busy = ["inflight", "toolspending"].includes(acknowledgedState);
    const paused = acknowledgedState === "paused";
    const status = document.querySelector("#state");
    const label = !connected ? "disconnected" : selectionPending ? "selecting" : acknowledgedState || "unknown";
    const context = connected && !selectionPending ? workingContext : null;
    const percent = Number.isFinite(context?.percent) ? `${Math.floor(context.percent)}% context` : "context unknown";
    status.textContent = `${label} · ${percent}`;
    status.title = context?.input_tokens != null
        ? `Last reported main request: ${context.input_tokens.toLocaleString()} input tokens / ${context.working_tokens.toLocaleString()} working limit. Not a prediction of the next request.`
        : "Working context: no current provider-reported input measurement. Unknown is not zero.";
    document.querySelector("#interrupt").disabled = !ready || !(busy || paused);
    // The actor classifies ordinary input when dequeued, even during a turn.
    document.querySelector("#send").disabled = !ready || !!voiceControl?.blocked || imageComposer.blocked || !!pageCapture?.busy || !!captureDraft || !!captureConsent?.busy;
    document.querySelector("#attach-images").disabled = !ready || !!voiceControl?.blocked || !!imageComposer.pending || !!pageCapture?.busy || !!captureDraft || !!captureConsent?.busy;
    const camera = document.querySelector("#capture-page");
    camera.disabled = !ready || !!voiceControl?.blocked || !pageCapture || !!captureConsent?.busy || !!imageComposer.pending || (pageCapture.busy && !captureDraft);
    camera.setAttribute("aria-pressed", String(!!captureDraft));
    camera.setAttribute("aria-label", captureDraft ? "Cancel page capture" : "Capture this page");
    camera.title = captureDraft ? "Cancel page capture" : "Capture this page";
    document.querySelector("#clear-tools").disabled = !ready;
    // Pause may be an idle speech/manual hold, not active work. The actor
    // still refuses model changes when a paused turn or queued work exists.
    document.querySelector("#model-choice").disabled = !ready || busy;
    // Theme is browser-local and remains available while disconnected.
}
// Format acknowledged Engine values only. No browser counters or price table.
function showUsage(frame) {
    const picker = document.querySelector("#model-choice");
    picker.replaceChildren();
    // Configured offerings change the complete arrangement; legacy curated
    // names preserve their route. The server validates every actual selection.
    const groups = new Map();
    const add = (vendor, id, label, enabled, offering = null, preference = null) => {
        if (!groups.has(vendor)) {
            const group = document.createElement("optgroup");
            group.label = vendor === "Gemini" ? "Google" : vendor;
            groups.set(vendor, group);
            picker.append(group);
        }
        const option = document.createElement("option");
        option.value = id;
        if (offering) option.dataset.offering = JSON.stringify(offering);
        if (preference) option.dataset.preference = JSON.stringify(preference);
        option.textContent = label;
        option.disabled = !enabled;
        if (!enabled) option.title = preference
            ? "Default funding is unavailable or ambiguous. Choose funding in Settings → AI."
            : offering ? "Configured offering unavailable under this Agent's current policy."
            : "Unavailable legacy choice for this Agent's current route.";
        groups.get(vendor).append(option);
    };
    const configured = frame.configured_models || [];
    const preferred = frame.preferred_models || [];
    const shown = new Set();
    let selected = frame.selected_model;
    const key = (vendor, model) => `${vendor.toLowerCase()}:${model}`;
    for (const model of preferred) {
        const id = "preference:" + JSON.stringify(model.preference);
        add(model.vendor, id, model.label + (model.selected && frame.selection_available === false ? " (unavailable)" : ""), model.enabled, null, model.preference);
        shown.add(key(model.vendor, model.model || model.preference.model));
        if (model.selected) selected = id;
    }
    // Embedded callers may have no global funding store. A sole explicit route
    // still works; ambiguous routes require the host's funding configuration.
    for (const model of configured) {
        const identity = key(model.vendor, model.model);
        if (shown.has(identity)) continue;
        shown.add(identity);
        const matches = configured.filter(value => key(value.vendor, value.model) === identity);
        const id = JSON.stringify(model.offering);
        add(model.vendor, id, model.label, matches.length === 1 && model.enabled, model.offering);
        if (frame.selected_model === model.model && frame.selected_offering?.vendor === model.offering.vendor) selected = id;
    }
    for (const model of frame.recommended_models || []) {
        if (shown.has(key(model.vendor, model.id))) continue;
        add(model.vendor, model.id, model.label, model.enabled);
        shown.add(key(model.vendor, model.id));
    }
    if (![...picker.options].some(option => option.value === selected))
        add(frame.model_vendor || "Current", selected, frame.selected_model + " (current)", frame.selection_available !== false);
    picker.title = "Select a model. Change its default funding in Settings → AI.";
    picker.value = selected;
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
        const label = row.display_name || (row.key === "" ? "Original unnamed conversation" : row.key ?? row.id);
        option.title = row.key || row.id;
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
    const name = selected?.display_name || selected?.id || agentID || "Original unnamed conversation";
    const input = document.querySelector("#prompt");
    input.placeholder = `Ask ${name} anything…`;
    input.setAttribute("aria-label", `Message to ${name}`);
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
    if (frame.type === "display_window") {
        const retained = new Set(frame.part_ids);
        for (const [id, card] of cards) {
            if (retained.has(id)) continue;
            card.workProductController?.abort();
            card.windowController?.abort();
            pendingCardRenders.delete(card);
            card.node.remove();
            cards.delete(id);
            if (card.callCard) card.callCard.latestResult = null;
            if (card.latestResult) card.latestResult.callCard = null;
        }
        jobSessions?.retain(cards.values());
        for (const card of cards.values()) renderPairLink(card);
        return;
    }
    if (frame.type === "goal_status" || frame.type === "goal_updated") {
        if (goalSelectionPending) return;
        goalDirty = false;
        showGoal(frame.goal);
        if (frame.type === "goal_updated") notice("Goal applied for this session.");
        return;
    }
    if (frame.type === "usage") {
        showWorkingContext(frame.working_context);
        showUsage(frame);
        return;
    }
    if (frame.type === "replay") {
        if (agentID !== frame.agent_id) {
            workingContext = null;
            settingsDrafts.clear();
            voiceControl?.cancel();
            cancelPageCapture("Page capture cancelled on Agent change.");
            cancelSpeech(false);
            localPauseCause = humanTyping;
        }
        imageComposer.selectionChanged(frame.agent_id);
        agentID = frame.agent_id;
        selectionPending = false;
        acknowledgedState = null;
        renderControls();
        goalSelectionPending = false;
        goalDirty = false;
        showGoal(agentRows.find(row => row.key === agentID)?.goal);
        send("get_goal");
        const relayLink = document.querySelector("#relay-link");
        relayLink.hidden = !agentID || agentID === socket?.relayAgentID;
        if (agentID) {
            const relayURL = new URL(location.href);
            relayURL.searchParams.set("agent_id", agentID);
            relayLink.href = relayURL.href;
        }
        showAgents();
        replaying = true;
        clearTimeout(cardRenderTimer);
        cardRenderTimer = null;
        pendingCardRenders.clear();
        for (const card of cards.values()) { card.workProductController?.abort(); card.windowController?.abort(); }
        cards.clear();
        // reset() notifies remaining cards; clear ownership before that callback.
        jobSessions?.reset();
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
        const nameChanged = settings?.human_name !== frame.settings.human_name;
        settings = frame.settings;
        applySettings();
        if (nameChanged) for (const card of cards.values()) render(card);
        notice(settingsDrafts.size ? "Settings received; unsaved edits remain in the form." : "");
        return;
    }
    if (frame.type === "agent_state" || frame.type === "state_changed") {
        acknowledgedState = frame.state || frame.to;
        agentClosed = acknowledgedState === "shutdown";
        if (agentClosed) cancelPageCapture("Page capture cancelled because the Agent closed.");
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
            // Durable errors share the server's retained display identity. Control
            // refusals are local and expire with the next authoritative window.
            const id = frame.part_id || (frame.kind === "error_occurred"
                ? `${agentID}:e${frame.seq}:error` : `local-error:${++localErrorSequence}`);
            const alreadyShown = cards.has(id);
            const card = cardFor(id);
            Object.assign(card, { kind: "error", role: "system", final: true,
                text: frame.message || "Request failed.", recordedAt: frame.recorded_at });
            render(card);
            applySettings();
            if (!alreadyShown && !replaying && browserSpeech.value.enabled) {
                flushNarration();
                speakText(id, card.text, "error");
            }
        }
        notice(frame.message || frame.kind);
        return;
    }
    if (!frame.part_id) return;
    const card = cardFor(frame.part_id);
    if (["part_final", "tool_finished"].includes(frame.type)
        || (frame.type === "tool_dispatched" && !card.recordedAt)) {
        card.recordedAt = frame.recorded_at;
    }
    if (frame.type === "tool_finished") card.observationMs = frame.observation_ms;
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
        card.toolName = frame.name;
        card.workProduct = frame.is_error ? null : frame.work_product;
        card.directoryListing = frame.is_error ? null : frame.directory_listing;
        card.fileEvidence = frame.is_error ? null : frame.file_evidence;
        card.searchEvidence = frame.is_error ? null : frame.search_evidence;
        card.role = "tool";
        card.status = frame.is_error ? "error" : "finished";
        card.text = (frame.result || (frame.parts || []).map(readable).join("\n"))
            .replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, "");
        card.final = true;
        if (frame.tool_part_id && cards.get(frame.tool_part_id)?.kind === "tool_call" &&
            cards.get(frame.tool_part_id)?.agent === card.agent) {
            const call = cards.get(frame.tool_part_id);
            call.latestResult = card;
            card.callCard = call;
            call.status = frame.is_error ? "error" : "finished";
            card.toolName = call.name;
            if (!frame.is_error && call.name === "edit_file" &&
                typeof call.input?.old_text === "string" && typeof call.input?.new_text === "string") {
                card.change = call.input;
            }
            render(call);
        }
    }
    if (frame.type === "part_delta" && !replaying) {
        // Insert once in arrival order; updates only replace this card's body.
        if (!card.node.parentElement) render(card);
        else queueCardRender(card);
    } else render(card);
    if (card.kind === "tool_result") jobSessions?.observe(card,
        ["wait_for_job", "send_input"].includes(card.toolName));
    // Once a continuing view exists, keep each immutable observation inside its
    // collapsed history instead of adding another output card to the pane.
    if (card.kind === "tool_result" && jobSessions?.container(card)) render(card);
    if (card.callCard) renderPairLink(card.callCard);
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
            if (voiceControl?.receive(connection, frame)) return;
            if (imageComposer.acknowledge(frame, connection)) return;
            if (frame.type === "connection") {
                connection.relayAgentID = frame.relay_agent_id;
                connection.voiceFunding = frame.transcription?.enabled === true
                    && frame.transcription.payer === "openai-api-key";
                renderControls();
            } else if (frame.type === "mcp") handleMcp(connection, frame);
            else handle(frame);
        }
        catch { notice("An invalid display frame was refused."); }
    };
    socket.onclose = () => {
        if (connection !== socket) return;
        jobSessions?.reset();
        for (const card of cards.values()) card.windowController?.abort();
        document.querySelector("#connection").textContent = "Disconnected — reconnecting";
        selectionPending = true;
        voiceControl?.cancel();
        cancelPageCapture("Page capture cancelled on disconnect; no screenshot was added.");
        imageComposer.disconnect();
        cancelSpeech(false);
        // Remember remaining local activity; reconnect must not reassert a stale pause.
        localPauseCause = humanTyping;
        renderControls();
        showGoal(undefined);
        clearTimeout(reconnectTimer);
        reconnectTimer = setTimeout(connect, 800);
    };
    socket.onerror = () => { notice("Connection unavailable; work may still be running."); };
}
function send(type, text = "", offering = null, preference = null) {
    if (!socket || socket.readyState !== WebSocket.OPEN) {
        notice("Not connected; input was not sent.");
        return false;
    }
    socket.send(JSON.stringify({ type, text, ...(offering ? { offering } : {}), ...(preference ? { preference } : {}), ...(agentID ? { agent_id: agentID } : {}) }));
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
    voiceControl?.cancel();
    cancelPageCapture("Page capture cancelled on Agent change.");
    jobSessions?.reset();
    for (const card of cards.values()) card.windowController?.abort();
    imageComposer.selectionChanged(key);
    // Release this tab's old pause cause before asking the server to switch.
    // The incoming view starts with an activity baseline, not a stale pause.
    cancelSpeech(false);
    if (localPauseCause) send("unpause");
    localPauseCause = false;
    selectionPending = true;
    renderControls();
    goalSelectionPending = true;
    showGoal(undefined);
    socket.send(JSON.stringify({ type: "select_agent", agent_id: key }));
}
// Fit rendered lines, including wrapping, within the CSS one-to-five-line bounds.
// Reset first so deleting text can shrink the input as well as grow it.
function resizeComposer() {
    const input = document.querySelector("#prompt");
    if (!input.clientWidth) return; // Hidden compact panes are measured when shown.
    const style = getComputedStyle(input);
    const border = parseFloat(style.borderTopWidth) + parseFloat(style.borderBottomWidth);
    input.style.height = style.minHeight;
    input.style.height = `${Math.min(input.scrollHeight + border, parseFloat(style.maxHeight))}px`;
}
let composerWidth = -1;
new ResizeObserver(entries => {
    const width = entries[0].contentRect.width;
    if (width === composerWidth) return; // Our height change is not a new sizing cause.
    composerWidth = width;
    resizeComposer();
}).observe(document.querySelector("#prompt"));
resizeComposer();

// One input path. Browser state must not classify a message as a hint.
function submit() {
    if (voiceControl?.blocked) { notice("Finish or cancel voice before sending a draft."); return; }
    if (document.querySelector("#send").disabled) {
        notice(pageCapture?.busy || captureDraft || captureConsent?.busy ? "Finish or cancel page capture before sending."
            : "Wait for the selected Agent to connect; input was not sent.");
        return;
    }
    const input = document.querySelector("#prompt");
    if (imageComposer.hasImages) {
        imageComposer.send(socket, agentID);
        return;
    }
    if (input.value.trim() && send("prompt", input.value)) {
        input.value = "";
        resizeComposer();
        humanTyping = false;
        updatePauseCause();
    }
}
document.querySelector("#compose").onsubmit = event => { event.preventDefault(); submit(); };
document.querySelector("#prompt").addEventListener("paste", event => {
    if (!socket || socket.readyState !== WebSocket.OPEN || selectionPending || !agentID || agentClosed) {
        if (Array.from(event.clipboardData?.items || []).some(item => item.kind === "file")) {
            event.preventDefault();
            notice("Wait for the selected Agent before pasting images.");
        }
        return;
    }
    imageComposer.paste(event, agentID);
});
const imageFiles = document.querySelector("#image-files");
document.querySelector("#attach-images").onclick = () => {
    if (document.querySelector("#attach-images").disabled) return;
    imageComposer.openPicker(imageFiles, socket, agentID);
};
imageFiles.onchange = () => {
    if (!socket || socket.readyState !== WebSocket.OPEN || selectionPending || agentClosed) {
        imageComposer.picker = null;
        imageFiles.value = "";
        return;
    }
    imageComposer.acceptPicker(imageFiles, socket, agentID);
};
imageFiles.addEventListener("cancel", () => { imageComposer.picker = null; });
window.addEventListener("pagehide", () => {
    cancelPageCapture();
    imageComposer.clear();
});
document.querySelector("#interrupt").onclick = () => send("interrupt");
document.querySelector("#clear-tools").onclick = () => send("clear_tools");
document.addEventListener("ensemble-funding-changed", () => {
    if (socket?.readyState === WebSocket.OPEN) socket.send(JSON.stringify({ type: "get_model_choices" }));
});
document.querySelector("#model-choice").onchange = event => {
    // A choice is the action. The server's usage frame confirms it or restores
    // the unchanged model after refusal; never infer acceptance from the click.
    const option = event.target.selectedOptions[0];
    const offering = option?.dataset.offering ? JSON.parse(option.dataset.offering) : null;
    const preference = option?.dataset.preference ? JSON.parse(option.dataset.preference) : null;
    if (send("set_model", offering || preference ? "" : event.target.value, offering, preference)) event.target.disabled = true;
};
document.querySelector("#prompt").oninput = event => {
    resizeComposer();
    humanTyping = !event.detail?.mcp && event.target.value.length > 0;
    updatePauseCause();
};
document.querySelector("#prompt").onkeydown = event => {
    if (event.key === "Escape") {
        event.preventDefault();
        stopSpeechAndResume();
    } else if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault(); submit();
    }
};
document.querySelector("#stop-speech").onclick = stopSpeechAndResume;
// Composer Escape is handled above; honour dialog/widget handlers first.
document.addEventListener("keydown", event => {
    if (event.key === "Escape" && !event.defaultPrevented) {
        // Do not prevent the native dialog-cancel action or its focus return.
        stopSpeechAndResume();
    }
});
function sameWorkingContextModel(first, second) {
    return first?.vendor === second?.vendor && first?.model === second?.model
        && first?.offering?.vendor === second?.offering?.vendor
        && first?.offering?.endpoint === second?.offering?.endpoint
        && first?.offering?.model === second?.offering?.model;
}
// Drafts are presentation only: acknowledged settings remain the policy authority.
const settingsDrafts = new Set();
const contextSettingIDs = ["working-context-tokens", "context-capacity-tokens"];
function forgetSettingsDrafts(ids) {
    for (const id of ids) settingsDrafts.delete(id);
}
function settingControlValue(control) {
    return control.type === "checkbox" ? control.checked : control.value;
}
function autosaveSettings(controls, save) {
    for (const control of controls) {
        control.addEventListener("input", () => settingsDrafts.add(control.id));
        control.addEventListener("change", () => {
            settingsDrafts.add(control.id);
            save();
        });
    }
}
function showWorkingContext(value) {
    const modelChanged = !sameWorkingContextModel(workingContext, value);
    workingContext = value || null;
    if (value && modelChanged) {
        forgetSettingsDrafts(contextSettingIDs);
        fillWorkingContextSettings(value);
    }
    renderControls();
}
function fillWorkingContextSettings(value) {
    workingContextForm = { vendor: value.vendor, model: value.model,
        ...(value.offering ? { offering: { ...value.offering } } : {}) };
    document.querySelector("#working-context-model").textContent = `${value.vendor} / ${value.model}`;
    document.querySelector("#working-context-tokens").value = value.working_tokens ?? "";
    document.querySelector("#context-capacity-tokens").value = value.capacity_tokens ?? "";
}
function saveWorkingContextSettings() {
    if (!workingContextForm) return;
    const working = Number(document.querySelector("#working-context-tokens").value);
    const capacityText = document.querySelector("#context-capacity-tokens").value;
    const capacity = capacityText === "" ? null : Number(capacityText);
    if (!Number.isSafeInteger(working) || working <= 0 || (capacity !== null
        && (!Number.isSafeInteger(capacity) || capacity <= 0 || working > capacity))) {
        notice("Enter positive whole input tokens; the working limit must not exceed the capacity ceiling.");
        return;
    }
    patchSettings({ working_context: { ...workingContextForm,
        working_tokens: working, capacity_tokens: capacity } });
}
document.querySelector("#save-working-context").onclick = saveWorkingContextSettings;
autosaveSettings(contextSettingIDs.map(id => document.getElementById(id)), saveWorkingContextSettings);

function applySettings() {
    if (!settings) return;
    const drafts = [...settingsDrafts].map(id => {
        const control = document.getElementById(id);
        return {control, value: settingControlValue(control)};
    });
    document.querySelector("#agent-display-name").value = settings.display_name || "";
    document.querySelector("#human-name").value = settings.human_name || "";
    document.querySelector("#human-description").value = settings.human_description || "";
    browserTheme.seed(settings.theme);
    browserDevice.seed(settings.font_size);
    resizeComposer();
    renderControls();
    document.querySelector("#max-tool-rounds").value = settings.max_tool_rounds;
    document.querySelector("#context-target-bytes").value = settings.context_target_bytes;
    // Settings acknowledgment/rollback restores form values, not the live usage
    // measurement. The usage stream may already be newer than this snapshot.
    const savedContext = settings.working_context;
    const currentModel = workingContext && !sameWorkingContextModel(workingContext, savedContext);
    if (currentModel) fillWorkingContextSettings(workingContext);
    else if (savedContext) fillWorkingContextSettings(savedContext);
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
        document.querySelector(`#memory-${band}-budget`).value = policy.budget;
    }
    for (const {control, value} of drafts) {
        if (settingControlValue(control) === value) settingsDrafts.delete(control.id);
        else if (control.type === "checkbox") control.checked = value;
        else control.value = value;
    }
    updatePauseCause(); // Disabling speech pausing does not cancel the queue.
}
function saveDisplayName() {
    const control = document.querySelector("#agent-display-name");
    if (!control.reportValidity()) return;
    control.value = control.value.trim();
    patchSettings({display_name: control.value || null});
}
autosaveSettings([document.querySelector("#agent-display-name")], saveDisplayName);
document.querySelector("#agent-name-settings").addEventListener("submit", event => {
    event.preventDefault();
    saveDisplayName();
});
document.querySelector("#user-settings").addEventListener("submit", event => {
    event.preventDefault();
    patchSettings({ human_name: document.querySelector("#human-name").value,
        human_description: document.querySelector("#human-description").value });
});
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
        busy: "Working", queued: "Continuation queued", idle: "Idle",
        retry_wait: "Waiting to retry", paused: "Paused",
        off: "Reminders off — goal retained", file_error: "Goal file error — reminders off"
    };
    const header = document.querySelector("#header-goal");
    header.classList.toggle("motto", !!goal && !goal.file);
    if (!goal) {
        header.textContent = "Goal …";
        header.title = "Waiting for acknowledged goal configuration.";
    } else if (!goal.file) {
        header.textContent = "Don’t Be Evil";
        header.title = "No goal configured.";
    } else {
        const configured = goal.last_error ? "Goal file error" : goal.available ? "Goal on" : "Goal unavailable";
        const reminders = goal.enabled ? "continue on idle" : "reminders off";
        header.textContent = `${configured} · ${reminders}`;
        header.title = `${header.textContent} — goal configuration, not Agent activity.`;
    }
    const status = document.querySelector("#goal-status");
    if (goal) {
        const details = [labels[goal.phase] || goal.phase];
        if (!goal.enabled && (goal.phase === "busy" || goal.phase === "paused")) {
            details.push("reminders off");
        }
        if (goal.file) details.push(goal.file);
        if (goal.due_utc) details.push(`next retry ${goal.due_utc}`);
        if (goal.retry_error) details.push(`failures ${goal.failure_count}: ${goal.retry_error}`);
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
    for (const id of ["goal-file", "goal-enabled", "goal-apply", "goal-off", "goal-clear"]) {
        document.querySelector(`#${id}`).disabled = !goal?.available;
    }
    document.querySelector("#goal-retry").disabled = !goal?.available || !goal.enabled || goal.phase !== "retry_wait";
    if (!goalDirty) {
        document.querySelector("#goal-file").value = goal?.file || "";
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
    updateGoal({file: document.querySelector("#goal-file").value,
        enabled: document.querySelector("#goal-enabled").checked});
});
document.querySelector("#goal-retry").onclick = () => updateGoal({retry_now: true});
document.querySelector("#goal-off").onclick = () => updateGoal({enabled: false});
document.querySelector("#goal-clear").onclick = () => updateGoal({file: ""});
document.querySelector("#goal-reload").onclick = () => send("get_goal");
function saveMemorySettings() {
    if (!document.querySelector("#memory-settings-form").checkValidity()) {
        notice("Complete the memory fields with values in their allowed ranges before saving.");
        return;
    }
    const memory = {};
    for (const band of ["conversation", "soul", "memory", "archive64", "archive8", "session"]) {
        const policy = {disabled: !document.querySelector(`#memory-${band}-enabled`).checked};
        const value = Number(document.querySelector(`#memory-${band}-budget`).value);
        if (!Number.isSafeInteger(value) || value < 0) {
            notice("Memory sizes must be nonnegative whole bytes within the browser’s exact integer range.");
            return;
        }
        policy.budget = value;
        memory[band] = policy;
    }
    patchSettings({memory});
}
function saveRecallSettings() {
    if (!document.querySelector("#recall-settings-form").checkValidity()) {
        notice("Complete the recall fields with values in their allowed ranges before saving.");
        return;
    }
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
}
for (const [group, save] of [["memory", saveMemorySettings], ["recall", saveRecallSettings]]) {
    const form = document.querySelector(`#${group}-settings-form`);
    const controls = [...form.querySelectorAll("input, select")];
    autosaveSettings(controls, save);
    form.addEventListener("submit", event => {
        event.preventDefault();
        save();
    });
    document.querySelector(`#${group}-settings-reload`).onclick = () => {
        forgetSettingsDrafts(controls.map(control => control.id));
        send("get_settings");
    };
}
document.querySelector("#font-size").onchange = event => browserDevice.select({ font: Number(event.target.value) });
document.querySelector("#max-tool-rounds").onchange = event => patchSettings({ max_tool_rounds: Number(event.target.value) });
document.querySelector("#context-target-bytes").onchange = event => patchSettings({ context_target_bytes: Number(event.target.value) });
document.querySelector("#tts-speed").onchange = event => browserSpeech.select({ speed: Number(event.target.value) });
document.querySelector("#narrate").onchange = event => browserSpeech.select({ enabled: event.target.checked });
document.querySelector("#voice-shortcut").onchange = event => browserDevice.select({ shortcut: event.target.value });
document.querySelector("#pause-speaking").onchange = event => browserSpeech.select({ pause: event.target.checked });
function renderDevicePreferences() {
    document.documentElement.style.fontSize = `${browserDevice.value.font ?? 18}px`;
    document.querySelector("#font-size").value = browserDevice.value.font ?? 18;
    document.querySelector("#voice-shortcut").value = browserDevice.value.shortcut;
    voiceControl?.setShortcut(browserDevice.value.shortcut);
    resizeComposer();
}
browserDevice.onchange = renderDevicePreferences;
renderDevicePreferences();
function renderSpeechPreferences() {
    const value = browserSpeech.value;
    document.querySelector("#tts-speed").value = value.speed;
    document.querySelector("#speed-label").value = `${value.speed}×`;
    document.querySelector("#narrate").checked = value.enabled;
    document.querySelector(".speech-toggle").title = value.enabled
        ? "Speech on in this browser" : "Speech off in this browser";
    document.querySelector("#pause-speaking").checked = value.pause;
    document.querySelector("#voice").value = value.voice;
}
browserSpeech.onchange = previous => {
    renderSpeechPreferences();
    if (previous.enabled && !browserSpeech.value.enabled) cancelSpeech(false);
    updatePauseCause();
};
document.querySelector("#voice").onchange = event => browserSpeech.select({ voice: event.target.value });
renderSpeechPreferences();
function voices() {
    const select = document.querySelector("#voice");
    const chosen = browserSpeech.value.voice;
    select.replaceChildren(new Option("Browser default", ""));
    for (const voice of speech?.getVoices() || []) select.add(new Option(`${voice.name} (${voice.lang})`, voice.voiceURI));
    select.value = chosen;
}
if (speech) {
    speech.addEventListener("voiceschanged", voices);
    voices();
    window.addEventListener("focus", checkSpeechCompletion);
    window.addEventListener("pageshow", checkSpeechCompletion);
    document.addEventListener("visibilitychange", () => {
        checkSpeechCompletion();
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
    function add(row, group, level = 1) {
        const children = row.key == null ? [] : agentRows.filter(child => child.parent === row.key);
        const item = document.createElement("li");
        item.setAttribute("role", "treeitem");
        item.setAttribute("aria-level", String(level));
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
            toggle.dataset.toggleAgent = row.key;
            toggle.tabIndex = -1;
            toggle.setAttribute("aria-label", `${expanded ? "Collapse" : "Expand"} ${row.display_name || row.id}`);
            line.append(toggle);
        } else {
            const spacer = document.createElement("span");
            spacer.className = "tree-expander-space";
            spacer.setAttribute("aria-hidden", "true");
            line.append(spacer);
        }
        const choose = actionButton("", () => selectAgent(row.key));
        choose.classList.add("tree-select");
        choose.disabled = row.key == null;
        choose.dataset.agent = row.key ?? row.id;
        choose.tabIndex = row.key === (focused ?? agentID) ? 0 : -1;
        choose.title = `${row.key || row.id} · ${row.paused ? "paused" : row.state} · ${[row.phase, row.tool, row.detail].filter(Boolean).join(" — ")}`;
        const name = document.createElement("span");
        name.className = "tree-name";
        name.textContent = row.display_name || row.id;
        const state = document.createElement("span");
        state.className = "tree-state";
        state.textContent = [row.submitted ? "submitted" : "", row.paused ? "paused" : row.state].filter(Boolean).join(" · ");
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
        if (row.can_delete) {
            const remove = document.createElement("button");
            remove.type = "button";
            remove.className = "btn tree-delete icon-button";
            remove.dataset.deleteAgent = row.key;
            remove.setAttribute("aria-label", `Delete ${row.display_name || row.id}`);
            remove.title = `Delete ${row.key}; saved files will remain`;
            remove.innerHTML = '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7M14 10v7"/></svg>';
            line.append(remove);
        }
        item.append(line);
        if (children.length) {
            const nested = document.createElement("ul");
            nested.setAttribute("role", "group");
            nested.hidden = collapsedAgents.has(row.key);
            for (const child of children) add(child, nested, level + 1);
            item.append(nested);
        }
        group.append(item);
    }
    for (const row of agentRows.filter(row => row.parent == null)) add(row, tree);
    if (!tree.querySelector('[tabindex="0"]')) tree.querySelector(".tree-select:not(:disabled)")?.setAttribute("tabindex", "0");
    if (focused != null) tree.querySelector(`[data-agent="${CSS.escape(focused)}"]`)?.focus();
}

connect();

// This module does not own or release speech/manual pauses. Voice is explicitly
// funded by the host and starts only from a trusted held control gesture.
import("./voice-control.js").then(({ VoiceControl }) => {
    voiceControl = new VoiceControl(document.querySelector("#voice-hold"),
        document.querySelector("#voice-cancel"), document.querySelector("#voice-status"));
    voiceControl.setShortcut(browserDevice.value.shortcut);
    voiceControl.addEventListener("change", renderControls);
    renderControls();
}).catch(() => {
    document.querySelector("#voice-status").textContent = "Voice controls could not load.";
});
document.querySelector("#prompt").addEventListener("input", () => {
    voiceControl?.cancel();
    renderControls();
});

// Renderer selection changes presentation, never recorded tool facts.
document.addEventListener("ensemble-renderers-changed", () => {
    for (const card of cards.values()) queueCardRender(card);
});
