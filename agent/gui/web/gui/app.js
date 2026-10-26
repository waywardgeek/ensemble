import { ArtifactScroll } from "./artifact-scroll.js";
import { browserMCP } from "./mcp.js";
import { SpeechBuffer, speechText } from "./speech.js";

// Each pane owns its artifacts through the same renderer. Renderers construct DOM nodes from text;
// filenames, tool results and model prose can never become executable markup.
const input = document.querySelector("#input");
const voice = document.querySelector("#voice");
const state = document.querySelector("#state");
const pauseState = document.querySelector("#pause-state");
const speechStatus = document.querySelector("#speech-status");
let socket,
  localPaused = false,
  speaking = false,
  humanTyping = false;
let pendingSelection = null;
const send = (message) => {
  if (socket?.readyState === WebSocket.OPEN)
    socket.send(JSON.stringify(message));
};
// Only local cause transitions send automatic commands. An explicit unpause
// elsewhere must remain effective while this tab has an unchanged local cause.
function pace() {
  if (pendingSelection !== null) return;
  // Field contents can come from MCP. Keep the human cause separately so a
  // later speech callback cannot reinterpret programmatic text as typing.
  const next = speaking || humanTyping;
  if (next !== localPaused) {
    localPaused = next;
    send({ type: next ? "pause" : "unpause" });
  }
}

// Queue complete phrases, not individual network tokens. Holding pause from
// enqueue through the last end/error gives a listener time to hear each phrase.
class Speech {
  queue = [];
  active = null;
  buffer = new SpeechBuffer(this);
  transcript = [];
  // Quiet observation still exercises this exact pipeline, not a second renderer.
  audio = new URLSearchParams(location.search).get("speech") !== "transcript";
  enqueue(text, source = "artifact") {
    text = speechText(text);
    if (!text) return;
    this.transcript.push({ text, source });
    window.dispatchEvent(new CustomEvent("ensemble-speech", { detail: { text, source } }));
    if (pendingSelection !== null || !this.audio || !voice.checked) return;
    this.queue.push({ text, source, queued_at: Date.now() });
    speaking = true;
    pace();
    this.advance();
  }
  // Keep the utterance alive until its callback. Some browser engines need the
  // JavaScript reference as well as the speech service's internal queued value.
  advance() {
    if (this.active) return;
    const pending = this.queue.shift();
    if (pending === undefined) {
      speaking = false;
      pace();
      return;
    }
    const utterance = new SpeechSynthesisUtterance(pending.text);
    utterance.source = pending.source;
    utterance.queued_at = pending.queued_at;
    utterance.started_at = null;
    utterance.onstart = () => { utterance.started_at = Date.now(); };
    // Eloquence accelerates nonlinearly above rate 2. Bill's observed
    // anchors are 2x at 2.0, about 3.5x at 2.6 and 5x at 2.9. Interpolate
    // those rather than passing 5 to the engine; other voices may differ.
    const speed = Number(document.querySelector("#tts-speed").value);
    utterance.rate = speed <= 2 ? speed
      : speed <= 3.5 ? 2 + (speed - 2) * 0.4
      : 2.6 + (speed - 3.5) * 0.2;
    this.active = utterance;
    const done = (event) => {
      if (this.active !== utterance) return; // A cancelled utterance cannot finish a newer one.
      speechStatus.textContent =
        event.type === "error" ? `Speech: ${event.error}` : "";
      this.active = null;
      this.advance();
    };
    utterance.onend = done;
    utterance.onerror = done;
    speechSynthesis.speak(utterance);
  }
  // MCP observes actual native speech progress, plus utterances waiting locally.
  pending() {
    const active = this.active ? [{ text: this.active.text, source: this.active.source, state: speechSynthesis.paused ? "paused" : this.active.started_at ? "speaking" : "starting", queued_at: this.active.queued_at, started_at: this.active.started_at, elapsed_ms: this.active.started_at ? Date.now() - this.active.started_at : 0 }] : [];
    return active.concat(this.queue.map(item => ({ ...item, state: "queued", started_at: null, elapsed_ms: 0 })));
  }
  // Part identity prevents a thinking tail from joining response prose. The
  // renderer already knows whether a final had deltas; reuse that information.
  chunk(id, text, final = false, source = "text") {
    this.buffer.chunk(id, text, source);
    if (final) this.buffer.flush(id);
  }
  flush() { this.buffer.flush(); }
  // Clear our generation before invoking native cancellation: its asynchronous
  // error callback may arrive after a new artifact has already queued speech.
  cancel() {
    this.active = null;
    this.queue = [];
    this.buffer.fragments.clear();
    speechSynthesis.cancel();
    speaking = false;
    pace();
  }
}
const speech = new Speech();
const receiveMCP = browserMCP(send, speech);
// The virtual listener receives only this transcript through its hear tool.
// Exposing the application-owned object also lets browser checks inspect native
// state without replacing the synthesizer or duplicating the speech algorithm.
window.ensembleSpeech = speech;
window.ensembleTurn = 0;
// Chrome can suspend its speech service after a tab switch. This silent nudge
// is deliberately outside the audible queue and therefore does not hold pause.
document.addEventListener("visibilitychange", () => {
  if (!speech.audio || !voice.checked) return;
  const wake = new SpeechSynthesisUtterance("");
  wake.volume = 0;
  speechSynthesis.speak(wake);
});


const chatScroll = new ArtifactScroll(document.querySelector("#artifacts"), speech);
const actionsScroll = new ArtifactScroll(document.querySelector("#actions"), speech);
// Provider tool parts and dispatch/results share the actions renderer, retaining
// call identity. Other content stays in chat without a second rendering path.
function route(message) {
  const tool = ["tool_dispatched", "tool_finished"].includes(message.type) ||
    message.kind === "tool_call" || message.part?.type === "tool_call";
  (tool ? actionsScroll : chatScroll).receive(message);
}
let agents = [], selectedAgent = "";
// The same socket orders release before selection. Suppress outgoing pacing
// until the requested replay boundary. Only one selection may be outstanding.
function selectAgent(key = "main") {
  if (pendingSelection !== null) return;
  pendingSelection = key;
  renderAgents(agents, document.querySelector("#agent-tree"));
  humanTyping = localPaused = false;
  speech.cancel();
  send({ type: "unpause" });
  send({ type: "select_agent", key });
}
function renderAgents(nodes, parent) {
  parent.replaceChildren();
  for (const agent of nodes) {
    const node = document.createElement("li");
    node.dataset.state = agent.state;
    node.dataset.activity = agent.activity?.phase || "idle";
    const select = document.createElement("button");
    select.type = "button";
    select.dataset.agentKey = agent.key;
    select.setAttribute("aria-pressed", String(agent.key === selectedAgent));
    const phase = agent.activity?.phase || "idle";
    const name = agent.display_name || agent.id;
    select.textContent = `${name}${agent.persistent ? " · saved" : ""} — ${agent.state} · ${phase}`;
    select.title = agent.activity?.detail || "Select conversation";
    select.disabled = pendingSelection !== null;
    select.onclick = () => selectAgent(agent.key);
    node.append(select);
    if (agent.key) {
      if (agent.key === selectedAgent) document.querySelector("#conversation-title").textContent = name;
      const rename = document.createElement("button");
      rename.type = "button";
      rename.textContent = "Rename";
      rename.setAttribute("aria-label", `Rename ${name}`);
      rename.onclick = () => {
        const displayName = window.prompt("Agent display name", name);
        if (displayName !== null) send({ type: "update_settings", key: agent.key, settings: { display_name: displayName } });
      };
      node.append(rename);
    }
    if (agent.deletable) {
      const remove = document.createElement("button");
      remove.type = "button";
      remove.textContent = "Delete";
      remove.setAttribute("aria-label", `Delete ${name}`);
      remove.onclick = () => send({ type: "delete_agent", name: agent.id });
      node.append(remove);
    }
    if (agent.children.length) {
      const children = document.createElement("ul");
      renderAgents(agent.children, children);
      node.append(children);
    }
    parent.append(node);
  }
}
document.querySelector("#create-agent").onsubmit = (event) => {
  event.preventDefault();
  send({ type: "create_agent", name: document.querySelector("#agent-name").value,
    model: document.querySelector("#agent-model").value, skill: document.querySelector("#agent-skill").value });
};
const theme = document.querySelector("#theme");
const model = document.querySelector("#model");
// Application fields are a separate explicit-save draft. Preference broadcasts
// may update active/restart facts, but must not replace unsaved provider edits.
const applicationPanel = document.querySelector("#application-settings");
const applicationStatus = document.querySelector("#application-status");
const saveApplication = document.querySelector("#save-application-settings");
const providerControls = Object.fromEntries([
  "vendor", "model", "billing_mode", "base_url", "credential_file", "reasoning_effort", "reasoning_summary",
].map(field => [field, document.getElementById(`provider-${field.replaceAll("_", "-")}`)]));
let applicationDirty = false, applicationRequest = null;
const providerDraft = () => Object.fromEntries(
  Object.entries(providerControls).map(([field, control]) => [field, control.value]));
const sameProvider = (a, b) => Object.keys(providerControls).every(field => (a?.[field] ?? "") === (b?.[field] ?? ""));
function resetApplicationDraft() {
  applicationDirty = false;
  applicationRequest = null;
  saveApplication.disabled = true;
  applicationStatus.textContent = "";
}
for (const control of Object.values(providerControls)) control.addEventListener("input", () => {
  applicationDirty = true;
  saveApplication.disabled = false;
  applicationStatus.textContent = "Unsaved application changes — save, then restart to apply.";
});
saveApplication.onclick = () => {
  if (applicationPanel.hidden || !applicationDirty) return;
  if (socket?.readyState !== WebSocket.OPEN) {
    applicationStatus.textContent = "Not connected. Application changes have not been saved.";
    return;
  }
  applicationRequest = { id: crypto.randomUUID(), provider: providerDraft() };
  send({ type: "update_settings", request_id: applicationRequest.id,
    settings: { provider: applicationRequest.provider } });
  applicationStatus.textContent = "Saving application settings; changes apply after restart.";
};
function applyApplicationSettings(application, requestID) {
  // The server includes this capability only for the file-configured main Agent.
  applicationPanel.hidden = !application;
  document.querySelector("#legacy-model-setting").hidden = !!application;
  model.disabled = !!application;
  if (!application) return;
  document.querySelector("#application-path").textContent = application.path;
  const active = application.active_provider;
  document.querySelector("#application-active").textContent =
    `Active now: ${active.vendor} · ${active.model} · ${active.billing_mode}`;
  document.querySelector("#application-restart").hidden = !application.restart_required;
  // A successful save may normalize values (such as relative credential paths).
  // Match the request identity, not saved values; retain edits made after Save.
  if (applicationRequest && requestID === applicationRequest.id) {
    if (sameProvider(providerDraft(), applicationRequest.provider)) {
      applicationDirty = false;
      applicationStatus.textContent = "Application settings saved. Changes apply after restart.";
    }
    applicationRequest = null;
  }
  if (!applicationDirty) {
    for (const [field, control] of Object.entries(providerControls)) control.value = application.provider[field] ?? "";
  }
  saveApplication.disabled = !applicationDirty;
}
const speed = document.querySelector("#tts-speed");
const font = document.querySelector("#font-size");
const scheme = matchMedia("(prefers-color-scheme: dark)");
function applyTheme() {
  document.body.dataset.theme = theme.value === "system" ?
    (scheme.matches ? "dark" : "light") : theme.value;
}
scheme.addEventListener("change", applyTheme);
const contextTarget = document.querySelector("#context-target");
const logRetention = document.querySelector("#log-retention");
// A snapshot acknowledging one field must not erase another field being edited.
// Number inputs commit on change/blur, so retain only their unsent local text.
const editingContext = new Set();
for (const control of [contextTarget, logRetention, model]) {
  control.addEventListener("input", () => editingContext.add(control));
}
// Band switches and byte budgets travel through the same persisted settings path.
const memoryControls = new Map();
for (const band of ["soul", "memory", "64x", "8x", "session", "conversation"]) {
  const row = document.createElement("div");
  const enabled = document.createElement("input"); enabled.type = "checkbox";
  const label = document.createElement("label"); label.append(enabled, ` ${band} enabled `);
  row.append(label);
  const controls = { disabled: enabled };
  for (const field of ["budget", "high_watermark", "low_watermark"]) {
    const input = document.createElement("input"); input.type = "number"; input.min = "0"; input.step = "1";
    input.setAttribute("aria-label", `${band} ${field} bytes (0 = default)`);
    input.style.width = "7em";
    input.addEventListener("input", () => editingContext.add(input));
    controls[field] = input;
    const caption = document.createElement("label");
    caption.append(` ${field.replaceAll("_", " ")} `, input); row.append(caption);
  }
  for (const [field, input] of Object.entries(controls)) input.addEventListener("change", () => {
    editingContext.delete(input);
    send({ type: "update_settings", settings: { memory: { [band]: { [field]: field === "disabled" ? !input.checked : Number(input.value) } } } });
  });
  memoryControls.set(band, controls);
  document.querySelector("#memory-settings").append(row);
}
function applySettings(settings, requestID) {
  applyApplicationSettings(settings.application, requestID);
  if (!editingContext.has(model)) model.value = settings.model;
  for (const [band, controls] of memoryControls) for (const [field, input] of Object.entries(controls)) {
    if (field === "disabled") input.checked = !settings.memory?.[band]?.disabled;
    else if (!editingContext.has(input)) input.value = settings.memory?.[band]?.[field] ?? 0;
  }
  if (!editingContext.has(contextTarget)) contextTarget.value = settings.context_target;
  if (!editingContext.has(logRetention)) logRetention.value = settings.log_retention;
  theme.value = settings.theme;
  applyTheme();
  voice.checked = settings.tts_enabled;
  if (!voice.checked) speech.cancel();
  speed.value = settings.tts_speed;
  document.querySelector("#speed-value").value = settings.tts_speed;
  // Loaded/API values can be any integer in range, not only the menu presets.
  if (![...font.options].some(option => Number(option.value) === settings.font_size))
    font.add(new Option(String(settings.font_size), String(settings.font_size)));
  font.value = settings.font_size;
  document.body.style.fontSize = `${settings.font_size}px`;
  document.querySelector("#settings-status").textContent = "";
}
for (const [control, field, value] of [
  [theme, "theme", () => theme.value],
  [model, "model", () => model.value.trim()],
  [contextTarget, "context_target", () => Number(contextTarget.value)],
  [logRetention, "log_retention", () => Number(logRetention.value)],
  [voice, "tts_enabled", () => voice.checked],
  [speed, "tts_speed", () => Number(speed.value)],
  [font, "font_size", () => Number(font.value)],
]) control.addEventListener("change", () => {
  if (control.disabled) return;
  editingContext.delete(control);
  send({ type: "update_settings", settings: { [field]: value() } });
});
// Native buttons and selected attributes expose the active panel to readers.
for (const tab of document.querySelectorAll('[role="tab"]')) {
  tab.onclick = () => {
    for (const other of document.querySelectorAll('[role="tab"]')) {
      const selected = other === tab;
      other.setAttribute("aria-selected", selected);
      document.getElementById(other.getAttribute("aria-controls")).hidden = !selected;
    }
  };
}
// Pointer capture survives leaving the narrow divider. Keyboard arrows provide
// the same resizing operation without requiring a precise mouse gesture.
for (const [id, selector, minimum, sign] of [
  ["sidebar-drag", ".sidebar", 180, 1], ["actions-drag", ".actions", 200, -1],
]) {
  const bar = document.getElementById(id), pane = document.querySelector(selector);
  const resize = (width) => {
    const maximum = Math.max(minimum, pane.offsetWidth + document.querySelector(".center").offsetWidth - 300);
    const next = Math.max(minimum, Math.min(maximum, width));
    pane.style.width = `${next}px`;
    bar.setAttribute("aria-valuenow", Math.round(next));
  };
  bar.setAttribute("aria-valuemin", minimum);
  bar.setAttribute("aria-valuenow", pane.offsetWidth);
  bar.onpointerdown = (event) => {
    const start = event.clientX, width = pane.offsetWidth;
    bar.setPointerCapture(event.pointerId);
    bar.onpointermove = (move) => resize(width + sign * (move.clientX - start));
    bar.onpointerup = bar.onpointercancel = () => { bar.onpointermove = null; };
  };
  bar.onkeydown = (event) => {
    if (["ArrowLeft", "ArrowRight"].includes(event.key)) {
      event.preventDefault();
      resize(pane.offsetWidth + sign * (event.key === "ArrowRight" ? 20 : -20));
    }
  };
}
function connect() {
  socket = new WebSocket(
    `${location.protocol === "https:" ? "wss" : "ws"}://${location.host}/ws`,
  );
  // Reconnection requests history only; stale local causes must not veto another tab's unpause.
  socket.onopen = () => {
    pendingSelection = null;
    chatScroll.reset();
    actionsScroll.reset();
    send({ type: "subscribe", key: selectedAgent });
  };
  socket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === "jsonrpc") { receiveMCP(message.payload); return; }
    if (message.type === "agent_selected") {
      // Reset only at the server's ordered replay boundary. Older queued frames
      // have arrived already; new live frames follow the selected Agent's replay.
      // Cancellation belongs to the outgoing view, never the new selection.
      humanTyping = localPaused = false;
      speech.cancel(); chatScroll.reset(); actionsScroll.reset();
      if (pendingSelection === message.key) pendingSelection = null;
      if (selectedAgent !== message.key) {
        resetApplicationDraft();
        applicationPanel.hidden = true;
        editingContext.clear();
      }
      selectedAgent = message.key;
      document.querySelector("#conversation-title").textContent = message.key;
    }
    if (message.type === "agent_tree") {
      agents = message.root.children;
      renderAgents(agents, document.querySelector("#agent-tree"));
    }
    if (message.type === "agent_error") {
      pendingSelection = null;
      renderAgents(agents, document.querySelector("#agent-tree"));
      document.querySelector("#agent-status").textContent = message.error;
    }
    if (message.type === "agent_created") document.querySelector("#agent-status").textContent = `Created ${message.key}`;
    if (message.type === "agent_deleted") {
      document.querySelector("#agent-status").textContent = `Deleted ${message.key}`;
      if (selectedAgent === message.key || selectedAgent.startsWith(message.key + "/")) selectAgent();
    }
    route(message);
    if (["current_settings", "settings_changed", "settings_update"].includes(message.type)) applySettings(message.settings, message.request_id);
    if (message.type === "settings_error") {
      if (applicationRequest && message.request_id === applicationRequest.id) {
        applicationRequest = null;
        applicationStatus.textContent = `Application settings were not saved: ${message.error}`;
      }
      document.querySelector("#settings-status").textContent = message.error;
      speech.flush();
      speech.enqueue(message.error, "error");
    }
    if (message.type === "billing_selected") {
      state.textContent = `Billing: ${message.source}`;
      speech.enqueue(`Billing: ${message.source}`, "status");
    }
    if (message.type === "state_changed") {
      state.textContent = message.to;
    }
    if (message.type === "pause_changed")
      pauseState.textContent = message.paused ? "paused" : "";
    if (message.type === "usage") {
      const meter = document.querySelector("#usage");
      const input = message.input + message.cache_write + message.cache_read;
      const cost = message.priced ? `$${message.cost_usd.toFixed(4)}` : "—";
      meter.textContent = `Cost ${cost} · Input ${input} · Cache ${message.cache_read} · Hit ${(100 * message.hit_rate).toFixed(1)}%`;
      const total = message.total;
      meter.title = `Session tokens: input ${total.input}, cache write ${total.cache_write}, cache read ${total.cache_read}, output ${total.output}`;
    }
    if (message.type === "turn_ended") {
      // End of stream must release even a trailing colon or interrupted phrase.
      speech.flush();
      if (message.error) {
        state.textContent = message.error;
        speech.enqueue(message.error, "error");
      }
      window.ensembleTurn++;
    }
  };
  socket.onclose = () => {
    state.textContent = "disconnected — reconnecting";
    setTimeout(connect, 1000);
  };
}
// Typing holds new tool starts even when speech is disabled. Submitting clears
// that cause, but queued speech may independently keep the local cause true.
input.addEventListener("input", event => {
  humanTyping = !event.detail?.mcp && input.value.length > 0;
  pace();
});
document.querySelector("#input-form").addEventListener("submit", (event) => {
  event.preventDefault();
  if (!input.value.trim() || socket?.readyState !== WebSocket.OPEN) return;
  send({ type: "prompt", text: input.value });
  input.value = "";
  humanTyping = false;
  pace();
});
input.addEventListener("keydown", (event) => {
  if (event.key === "Enter" && !event.shiftKey) {
    event.preventDefault();
    document.querySelector("#input-form").requestSubmit();
  }
});
// Escape works after a speaker click too, when focus is on the card button.
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") speech.cancel();
});
document.querySelector("#interrupt").onclick = () => {
  speech.cancel();
  send({ type: "interrupt" });
};
// This is an explicit global action. Other tabs keep their local causes, but
// cannot veto it or immediately reassert pause without a new local transition.
document.querySelector("#resume").onclick = () => send({ type: "unpause" });
connect();
