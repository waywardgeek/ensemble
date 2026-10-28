"use strict";
// Browser-side MCP service for this page only. No server-side eval, navigation,
// hidden browser profile, or automatic snapshot is offered by this catalogue.
const browserTools = [
    { name: "gui_snapshot", description: "Inspect visible page text and CSS selectors; not pixels.",
        inputSchema: { type: "object", properties: {}, additionalProperties: false } },
    { name: "gui_click", description: "Click one visible element identified by a snapshot selector.",
        inputSchema: { type: "object", properties: { selector: { type: "string" } }, required: ["selector"], additionalProperties: false } },
    { name: "gui_input", description: "Replace a visible text field's value and dispatch its input event.",
        inputSchema: { type: "object", properties: { selector: { type: "string" }, text: { type: "string" } }, required: ["selector", "text"], additionalProperties: false } },
    { name: "tts_queue", description: "Inspect this page's current and queued speech; does not establish audibility.",
        inputSchema: { type: "object", properties: {}, additionalProperties: false } },
    { name: "tts_say", description: "Queue speech, optionally replacing current and queued speech.",
        inputSchema: { type: "object", properties: { text: { type: "string" }, interrupt: { type: "boolean" } }, required: ["text"], additionalProperties: false } }
];

function visibleElement(element) {
    return element instanceof HTMLElement && element.getClientRects().length > 0
        && getComputedStyle(element).visibility !== "hidden";
}
function elementSelector(element) {
    if (element.id) return `#${CSS.escape(element.id)}`;
    const path = [];
    for (let node = element; node && node !== document.body; node = node.parentElement) {
        const siblings = [...node.parentElement.children].filter(other => other.tagName === node.tagName);
        path.unshift(`${node.localName}:nth-of-type(${siblings.indexOf(node) + 1})`);
    }
    return "body" + (path.length ? " > " + path.join(" > ") : "");
}
function browserSnapshot() {
    const visible = [...document.querySelectorAll("button,input,textarea,select,a,summary,h1,h2,h3,[role]")]
        .filter(visibleElement);
    const text = document.body.innerText;
    return {
        title: document.title, url: location.href,
        text: text.slice(0, 20000), text_truncated: text.length > 20000,
        elements_truncated: visible.length > 200,
        elements: visible.slice(0, 200).map(element => ({
            selector: elementSelector(element), tag: element.localName,
            text: element.innerText, label: element.getAttribute("aria-label"),
            value: element.type === "password" ? undefined : element.value,
            disabled: Boolean(element.disabled)
        }))
    };
}
function targetElement(selector) {
    const matches = document.querySelectorAll(selector);
    if (matches.length !== 1 || !visibleElement(matches[0])) {
        throw new Error("Selector must identify exactly one visible element; request a fresh snapshot.");
    }
    if (matches[0].disabled) throw new Error("Selected element is disabled.");
    return matches[0];
}
function browserCall(name, args) {
    const tool = browserTools.find(candidate => candidate.name === name);
    if (!tool) throw new Error(`Unknown browser tool: ${name}`);
    if (!args || typeof args !== "object" || Array.isArray(args)) throw new Error("Arguments must be an object.");
    for (const key of Object.keys(args)) {
        const expected = tool.inputSchema.properties[key];
        if (!expected || typeof args[key] !== expected.type) throw new Error(`Invalid argument: ${key}`);
    }
    for (const key of tool.inputSchema.required || []) {
        if (!(key in args)) throw new Error(`Missing argument: ${key}`);
    }
    switch (name) {
        case "gui_snapshot": return browserSnapshot();
        case "gui_click":
            targetElement(args.selector).click();
            return "Click dispatched. Request a fresh snapshot to observe its effect.";
        case "gui_input": {
            const element = targetElement(args.selector);
            if (!(element instanceof HTMLTextAreaElement ||
                (element instanceof HTMLInputElement && ["text", "search", "email", "url", "tel", "password"].includes(element.type)))) {
                throw new Error("Selected element is not a text input.");
            }
            if (element.readOnly) throw new Error("Selected field is read-only.");
            element.value = args.text;
            element.dispatchEvent(new Event("input", { bubbles: true }));
            return "Input dispatched. Request a fresh snapshot to observe its effect.";
        }
        case "tts_queue":
            return { current: activeSpeech ? { id: activeSpeech.id, text: activeSpeech.utterance.text } : null,
                queued: speechQueue.map(item => ({ id: item.id, text: item.text })) };
        case "tts_say":
            if (!speech || !window.SpeechSynthesisUtterance) throw new Error("Speech synthesis is unavailable.");
            if (!args.text.trim()) throw new Error("Speech text must not be empty.");
            if (args.interrupt) cancelSpeech();
            enqueueSpeech(`mcp:${crypto.randomUUID()}`, args.text);
            return "Speech queued; playback and audibility are not verified.";
    }
}
function handleMcp(connection, frame) {
    // Replies belong to the receiving connection, not the mutable global socket.
    // A reconnect cannot accidentally receive an old session's RPC response.
    if (frame.agent_id !== agentID) throw new Error("MCP targets a different Agent.");
    const request = frame.data;
    if (!request || request.jsonrpc !== "2.0" || typeof request.method !== "string") {
        throw new Error("Invalid MCP request.");
    }
    if (!("id" in request)) {
        if (request.method === "notifications/initialized") connection.mcpInitialized = true;
        return; // Cancellation notifications cannot undo already-completed DOM effects.
    }
    const reply = { jsonrpc: "2.0", id: request.id };
    if (request.method === "initialize") {
        if (request.params?.protocolVersion !== "2024-11-05") {
            reply.error = { code: -32602, message: "Unsupported MCP protocol version." };
        } else {
            reply.result = { protocolVersion: "2024-11-05", capabilities: { tools: {} },
                serverInfo: { name: "ensemble-browser", version: "1" } };
        }
    } else if (!connection.mcpInitialized) {
        reply.error = { code: -32000, message: "Initialize the browser service first." };
    } else if (request.method === "tools/list") {
        reply.result = { tools: browserTools };
    } else if (request.method === "tools/call") {
        try {
            const value = browserCall(request.params?.name, request.params?.arguments ?? {});
            reply.result = { content: [{ type: "text", text: typeof value === "string" ? value : JSON.stringify(value) }] };
        } catch (error) {
            reply.result = { isError: true, content: [{ type: "text", text: error.message }] };
        }
    } else {
        reply.error = { code: -32601, message: "Unknown browser MCP method." };
    }
    if (connection.readyState === WebSocket.OPEN) {
        connection.send(JSON.stringify({ type: "mcp", agent_id: frame.agent_id, data: reply }));
    }
}
