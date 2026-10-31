// The browser is an MCP server. Keeping protocol dispatch here leaves the chat
// renderer independent of tools, and every reply uses its original request ID.
export function browserMCP(send, speech) {
  const tools = [
    { name: "gui_snapshot", description: "Visible panes, controls and artifact previews", ephemeral: "round", inputSchema: { type: "object", properties: {} } },
    { name: "gui_click", description: "Click a visible control by CSS selector", inputSchema: { type: "object", properties: { selector: { type: "string" } }, required: ["selector"] } },
    { name: "gui_input", description: "Set an input value and dispatch input/change", inputSchema: { type: "object", properties: { selector: { type: "string" }, text: { type: "string" } }, required: ["selector", "text"] } },
    { name: "tts_queue", description: "Pending utterances with state and timing", ephemeral: "round", inputSchema: { type: "object", properties: {} } },
  ];
  return (request) => {
    if (request.id === undefined) return; // Initialized/cancel are notifications.
    const reply = { jsonrpc: "2.0", id: request.id };
    if (request.method === "initialize") {
      reply.result = { protocolVersion: "2024-11-05", capabilities: { tools: {} }, serverInfo: { name: "Ensemble browser", version: "1" } };
    } else if (request.method === "tools/list") reply.result = { tools };
    else if (request.method === "tools/call") {
      let text, isError = false;
      try {
        const { name, arguments: args = {} } = request.params;
        if (name === "gui_snapshot") text = snapshot();
        else if (name === "tts_queue") text = JSON.stringify(speech.pending());
        else if (name === "gui_click" || name === "gui_input") {
          const element = document.querySelector(args.selector);
          if (!element || !visible(element)) throw new Error("No visible element matches selector");
          if (name === "gui_click") element.click();
          else {
            if (!("value" in element) || typeof args.text !== "string") throw new Error("gui_input requires a value control and text");
            element.value = args.text;
            element.dispatchEvent(new Event("input", { bubbles: true }));
            element.dispatchEvent(new Event("change", { bubbles: true }));
          }
          text = "ok";
        } else throw new Error(`Unknown tool ${name}`);
      } catch (error) { text = error.message; isError = true; }
      reply.result = { content: [{ type: "text", text }], isError };
    } else reply.error = { code: -32601, message: "Method not found" };
    send({ type: "jsonrpc", payload: reply });
  };
}

function visible(element) {
  return element.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true });
}

// IDs are preferred; otherwise the structural selector remains usable even for
// repeated speaker buttons. Labels alone are neither unique nor valid selectors.
function selector(element) {
  if (element === document.documentElement) return "html";
  if (element.id) return `#${CSS.escape(element.id)}`;
  const siblings = [...element.parentElement.children].filter(child => child.tagName === element.tagName);
  return `${selector(element.parentElement)} > ${element.tagName.toLowerCase()}:nth-of-type(${siblings.indexOf(element) + 1})`;
}

function snapshot() {
  const units = [];
  for (const pane of document.querySelectorAll("main > aside, main > section")) {
    if (visible(pane)) units.push({ text: `## ${pane.getAttribute("aria-label") || pane.className} (${Math.round(pane.getBoundingClientRect().width)}px wide)\n` });
  }
  for (const element of document.querySelectorAll("button,input,textarea,select,summary,[role=tab],[role=separator]")) {
    if (!visible(element)) continue;
    const state = {};
    for (const attribute of element.attributes) {
      if (attribute.name.startsWith("aria-") || ["type", "role"].includes(attribute.name)) state[attribute.name] = attribute.value;
    }
    for (const key of ["value", "checked", "disabled", "selected", "readOnly", "required"]) {
      if (key in element) state[key] = element[key];
    }
    if (element.tagName === "SUMMARY") state.expanded = element.parentElement.open;
    units.push({ text: `- ${selector(element)} ${JSON.stringify(element.getAttribute("aria-label") || element.textContent.trim())} ${JSON.stringify(state)}\n` });
  }
  for (const artifact of document.querySelectorAll(".artifact")) {
    if (visible(artifact)) units.push({ artifact: true, text: `Artifact ${selector(artifact)}: ${artifact.innerText}\n` });
  }
  // Keep controls whole: a truncated state object invites guesses. Artifact
  // previews can be shortened, but the footer accounts for every omitted char.
  const encoder = new TextEncoder();
  let text = "# Visible GUI\n", elided = 0, omitted = 0;
  for (const unit of units) {
    let shown = unit.text;
    const available = 4096 - 128 - encoder.encode(text).length;
    if (encoder.encode(shown).length > available) {
      if (!unit.artifact) shown = "";
      else {
        shown = [...shown].slice(0, Math.max(0, available)).join("");
        while (encoder.encode(shown).length > available) shown = [...shown].slice(0, -1).join("");
      }
    }
    if (unit.artifact && !shown) omitted++;
    elided += [...unit.text].length - [...shown].length;
    text += shown;
  }
  return `${text}\n[Snapshot: ${elided} characters elided; ${omitted} artifacts omitted.]`;
}
