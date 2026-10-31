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
const send = (message) => {
  if (socket?.readyState === WebSocket.OPEN)
    socket.send(JSON.stringify(message));
};
// Only local cause transitions send automatic commands. An explicit unpause
// elsewhere must remain effective while this tab has an unchanged local cause.
function pace() {
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
    if (!this.audio || !voice.checked) return;
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
    utterance.rate = Number(document.querySelector("#tts-speed").value);
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


function markdown(node, text) {
  // A deliberately small Markdown subset: headings, bullets, fences, bold and
  // inline code. All source text remains reachable even when syntax is unknown.
  let code = false;
  for (const line of text.split("\n")) {
    if (line.startsWith("```")) {
      code = !code;
      continue;
    }
    const heading = line.match(/^(#{1,6})\s+(.*)/);
    const row = document.createElement(
      code ? "pre" : heading ? `h${heading[1].length}` : "p",
    );
    const value = heading ? heading[2] : line.replace(/^[-*] /, "• ");
    if (code) row.textContent = value;
    else {
      for (const token of value.split(/(\*\*[^*]+\*\*|`[^`]+`)/g)) {
        const tag =
          token.startsWith("**") && token.endsWith("**")
            ? "strong"
            : token.startsWith("`") && token.endsWith("`")
              ? "code"
              : "span";
        const child = document.createElement(tag);
        child.textContent =
          tag === "strong"
            ? token.slice(2, -2)
            : tag === "code"
              ? token.slice(1, -1)
              : token;
        row.append(child);
      }
    }
    node.append(row);
  }
}
// Recognize display colour only, without interpreting arbitrary terminal
// controls or treating shell output as HTML. Unknown colour codes reset style.
function ansi(node, text) {
  let colour = "";
  for (const token of text.split(/(\x1b\[[0-9;]*m)/g)) {
    if (token.startsWith("\x1b[")) {
      const colours = { 31: "red", 32: "green", 33: "yellow", 34: "blue" };
      colour =
        token
          .slice(2, -1)
          .split(";")
          .map((n) => colours[n])
          .find(Boolean) || "";
    } else {
      const span = document.createElement("span");
      span.className = colour ? `ansi-${colour}` : "";
      span.textContent = token;
      node.append(span);
    }
  }
}
// Final parts replace provisional text. Private provider material is retained
// in history; only its readable thinking text is useful in the scroll.
function partText(part) {
  if (part.type === "tool_call")
    return `${part.name}\n${JSON.stringify(part.args, null, 2)}`;
  if (part.type === "opaque") {
    const value = part.data || {};
    // Signatures and redacted thinking are replay bytes, never spoken prose.
    if (value.type === "reasoning") return (value.summary ?? []).map(part => part.text ?? "").join("");
    if (value.type === "thinking") return value.thinking ?? "";
    if (value.thought === true) return value.text ?? "";
    return "";
  }
  // An empty text part is still text; its provenance is not a fallback body.
  if (part.type === "text") return part.text ?? "";
  return part.stub ?? "";
}

class ArtifactScroll {
  cards = new Map();
  calls = new Map();
  constructor(element) {
    this.element = element;
  }
  reset() {
    this.cards.clear();
    this.calls.clear();
    this.element.replaceChildren();
  }
  // One card shape serves every kind. The speaker closure reads current text,
  // so a card created during streaming speaks its full finalized value later.
  card(id, kind) {
    id = String(id);
    if (this.cards.has(id)) return this.cards.get(id);
    const element = document.createElement("article");
    element.className = `artifact artifact--${kind}`;
    element.dataset.id = id;
    const header = document.createElement("header");
    const label = document.createElement("span");
    label.textContent = kind;
    const speaker = document.createElement("button");
    speaker.textContent = "🔊";
    speaker.setAttribute("aria-label", "Read artifact aloud");
    const content = document.createElement("div");
    content.className = "content";
    header.append(label, speaker);
    element.append(header, content);
    this.element.append(element);
    const card = { element, label, content, text: "", kind, streamed: false };
    speaker.addEventListener("click", () => speech.enqueue(card.text));
    this.cards.set(id, card);
    return card;
  }
  // Follow the stream only when the reader was already at the bottom. Reading
  // an earlier artifact should not fight an unconditional scroll-to-bottom.
  render(card) {
    const nearEnd =
      this.element.clientHeight + this.element.scrollTop >= this.element.scrollHeight - 180;
    card.content.replaceChildren();
    // A cap bounds the normal card while the expander retains every byte.
    const shown = card.text.slice(0, 4000);
    const body = document.createElement("div");
    card.content.append(body);
    if (card.kind === "text" || ["thinking", "reasoning_summary"].includes(card.kind)) markdown(body, shown);
    else {
      const pre = document.createElement("pre");
      if (card.label.textContent === "edit_file") {
        for (const line of shown.split("\n")) {
          const row = document.createElement("div");
          row.className = line.startsWith("+")
            ? "diff-add"
            : line.startsWith("-")
              ? "diff-remove"
              : "";
          row.textContent = line;
          pre.append(row);
        }
      } else ansi(pre, shown);
      body.append(pre);
    }
    if (["thinking", "reasoning_summary"].includes(card.kind)) {
      const details = document.createElement("details");
      const summary = document.createElement("summary");
      summary.textContent = "Thinking";
      body.replaceWith(details);
      details.append(summary, body);
    }
    if (card.text.length > shown.length) {
      const details = document.createElement("details");
      const summary = document.createElement("summary");
      summary.textContent = "Full artifact";
      const full = document.createElement("pre");
      full.textContent = card.text;
      details.append(summary, full);
      card.content.append(details);
    }
    if (nearEnd) this.element.scrollTo(0, this.element.scrollHeight);
  }
  // History replay is silent. Finals reconcile display text, while already
  // streamed speech is never re-enqueued just because a final arrives.
  receive(message) {
    const id = message.part_id;
    if (message.type === "part_delta") {
      const card = this.card(id, message.kind);
      card.text += message.chunk;
      card.streamed = true;
      this.render(card);
      if (!message.replay && ["text", "thinking", "reasoning_summary"].includes(message.kind))
        speech.chunk(id, message.chunk, false, message.kind);
    } else if (
      message.type === "part_final" ||
      message.type === "user_message"
    ) {
      const part = message.part;
      const kind =
        message.type === "user_message"
          ? "user"
          : part.type === "opaque"
            ? (part.from?.surface === "responses" ? "reasoning_summary" : "thinking")
            : part.type;
      const card = this.card(id, kind);
      card.kind = kind;
      card.text = partText(part);
      this.render(card);
      if (part.type === "tool_call") this.calls.set(part.call_id, card);
      if (!message.replay && ["text", "thinking", "reasoning_summary"].includes(kind))
        speech.chunk(id, card.streamed ? "" : card.text, true, kind);
    } else if (message.type === "tool_dispatched") {
      const card =
        this.calls.get(message.call_id) ||
        this.card(`call-${message.call_id}`, "tool");
      this.calls.set(message.call_id, card);
      card.kind = "tool";
      card.label.textContent = message.name;
      card.text = `${message.name}\n${JSON.stringify(message.input, null, 2)}`;
      this.render(card);
      if (!message.replay) {
        speech.flush();
        speech.enqueue(message.name, "tool");
      }
    } else if (message.type === "tool_finished") {
      // Tool output stays silent automatically, but the same speaker button
      // still provides full on-demand access to arguments and result text.
      const card =
        this.calls.get(message.call_id) ||
        this.card(`call-${message.call_id}`, "tool");
      card.text += `\n\n${message.is_error ? "Error: " : ""}${message.result}`;
      if (message.is_error) card.element.classList.add("artifact--error");
      this.render(card);
    }
  }
}
const chatScroll = new ArtifactScroll(document.querySelector("#artifacts"));
const actionsScroll = new ArtifactScroll(document.querySelector("#actions"));
// Provider tool parts and dispatch/results share the actions renderer, retaining
// call identity. Other content stays in chat without a second rendering path.
function route(message) {
  const tool = ["tool_dispatched", "tool_finished"].includes(message.type) ||
    message.kind === "tool_call" || message.part?.type === "tool_call";
  (tool ? actionsScroll : chatScroll).receive(message);
}
const agents = [{ id: "root", name: "Ensemble", state: "idle", children: [] }];
function renderAgents(nodes, parent) {
  parent.replaceChildren();
  for (const agent of nodes) {
    const node = document.createElement("li");
    node.dataset.state = agent.state;
    const indicator = document.createElement("span");
    indicator.className = "agent-indicator";
    node.append(indicator, `${agent.name} — ${agent.state}`);
    if (agent.children.length) {
      const children = document.createElement("ul");
      renderAgents(agent.children, children);
      node.append(children);
    }
    parent.append(node);
  }
}
renderAgents(agents, document.querySelector("#agent-tree"));
const theme = document.querySelector("#theme");
const model = document.querySelector("#model");
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
function applySettings(settings) {
  if (!editingContext.has(model)) model.value = settings.model;
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
    chatScroll.reset();
    actionsScroll.reset();
    send({ type: "subscribe" });
  };
  socket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    if (message.type === "jsonrpc") { receiveMCP(message.payload); return; }
    route(message);
    if (["current_settings", "settings_changed"].includes(message.type)) applySettings(message.settings);
    if (message.type === "settings_error") {
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
      agents[0].state = message.to;
      renderAgents(agents, document.querySelector("#agent-tree"));
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
