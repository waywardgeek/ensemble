// One scroll owns every artifact. Renderers construct DOM nodes from text;
// filenames, tool results and model prose can never become executable markup.
const input = document.querySelector("#input");
const voice = document.querySelector("#voice");
const state = document.querySelector("#state");
const pauseState = document.querySelector("#pause-state");
const speechStatus = document.querySelector("#speech-status");
let socket,
  localPaused = false,
  speaking = false;
const send = (message) => {
  if (socket?.readyState === WebSocket.OPEN)
    socket.send(JSON.stringify(message));
};
// Only local cause transitions send automatic commands. An explicit unpause
// elsewhere must remain effective while this tab has an unchanged local cause.
function pace() {
  const next = speaking || input.value.length > 0;
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
  fragments = new Map();
  enqueue(text) {
    if (!text.trim()) return;
    this.queue.push(text);
    speaking = true;
    pace();
    this.advance();
  }
  // Keep the utterance alive until its callback. Some browser engines need the
  // JavaScript reference as well as the speech service's internal queued value.
  advance() {
    if (this.active) return;
    const text = this.queue.shift();
    if (text === undefined) {
      speaking = false;
      pace();
      return;
    }
    const utterance = new SpeechSynthesisUtterance(text);
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
  // Phrase buffers belong to artifact IDs, so a thinking fragment cannot finish
  // an unrelated chat sentence. Finals flush only the unspoken remainder.
  chunk(id, text, final = false) {
    if (!voice.checked) return;
    let pending = (this.fragments.get(id) || "") + text;
    let phrase;
    while ((phrase = pending.match(/^([\s\S]*?(?:[.!?](?:\s+|$)|\n))/))) {
      this.enqueue(phrase[1]);
      pending = pending.slice(phrase[1].length);
    }
    if (final) {
      this.enqueue(pending);
      this.fragments.delete(id);
    } else this.fragments.set(id, pending);
  }
  // Clear our generation before invoking native cancellation: its asynchronous
  // error callback may arrive after a new artifact has already queued speech.
  cancel() {
    this.active = null;
    this.queue = [];
    this.fragments.clear();
    speechSynthesis.cancel();
    speaking = false;
    pace();
  }
}
const speech = new Speech();
// Chrome can suspend its speech service after a tab switch. This silent nudge
// is deliberately outside the audible queue and therefore does not hold pause.
document.addEventListener("visibilitychange", () => {
  const wake = new SpeechSynthesisUtterance("");
  wake.volume = 0;
  speechSynthesis.speak(wake);
});
voice.addEventListener("change", () => {
  if (!voice.checked) speech.cancel();
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
      window.innerHeight + window.scrollY >= document.body.scrollHeight - 180;
    card.content.replaceChildren();
    // A cap bounds the normal card while the expander retains every byte.
    const shown = card.text.slice(0, 4000);
    const body = document.createElement("div");
    card.content.append(body);
    if (card.kind === "text" || card.kind === "thinking") markdown(body, shown);
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
    if (card.kind === "thinking") {
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
    if (nearEnd) card.element.scrollIntoView({ block: "end" });
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
      if (!message.replay && ["text", "thinking"].includes(message.kind))
        speech.chunk(id, message.chunk);
    } else if (
      message.type === "part_final" ||
      message.type === "user_message"
    ) {
      const part = message.part;
      const kind =
        message.type === "user_message"
          ? "user"
          : part.type === "opaque"
            ? "thinking"
            : part.type;
      const card = this.card(id, kind);
      card.kind = kind;
      card.text = partText(part);
      this.render(card);
      if (part.type === "tool_call") this.calls.set(part.call_id, card);
      if (!message.replay && ["text", "thinking"].includes(kind))
        speech.chunk(id, card.streamed ? "" : card.text, true);
    } else if (message.type === "tool_dispatched") {
      const card =
        this.calls.get(message.call_id) ||
        this.card(`call-${message.call_id}`, "tool");
      this.calls.set(message.call_id, card);
      card.kind = "tool";
      card.label.textContent = message.name;
      card.text = `${message.name}\n${JSON.stringify(message.input, null, 2)}`;
      this.render(card);
      if (!message.replay && voice.checked)
        speech.enqueue(
          `${message.name.replaceAll("_", " ")}: ${JSON.stringify(message.input).slice(0, 140)}`,
        );
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
const scroll = new ArtifactScroll(document.querySelector("#artifacts"));
function connect() {
  socket = new WebSocket(
    `${location.protocol === "https:" ? "wss" : "ws"}://${location.host}/ws`,
  );
  // Reconnection requests history only; stale local causes must not veto another tab's unpause.
  socket.onopen = () => {
    scroll.reset();
    send({ type: "subscribe" });
  };
  socket.onmessage = (event) => {
    const message = JSON.parse(event.data);
    scroll.receive(message);
    if (message.type === "state_changed") state.textContent = message.to;
    if (message.type === "pause_changed")
      pauseState.textContent = message.paused ? "paused" : "";
    if (message.type === "turn_ended") {
      // An interrupted partial has no final callback to clear its phrase buffer.
      speech.fragments.clear();
      if (message.error) state.textContent = message.error;
    }
  };
  socket.onclose = () => {
    state.textContent = "disconnected — reconnecting";
    setTimeout(connect, 1000);
  };
}
// Typing holds new tool starts even when speech is disabled. Submitting clears
// that cause, but queued speech may independently keep the local cause true.
input.addEventListener("input", pace);
document.querySelector("#input-form").addEventListener("submit", (event) => {
  event.preventDefault();
  if (!input.value.trim() || socket?.readyState !== WebSocket.OPEN) return;
  send({ type: "prompt", text: input.value });
  input.value = "";
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
  if (event.key === "Escape" && input.value === "") speech.cancel();
});
document.querySelector("#interrupt").onclick = () => {
  speech.cancel();
  send({ type: "interrupt" });
};
// This is an explicit global action. Other tabs keep their local causes, but
// cannot veto it or immediately reassert pause without a new local transition.
document.querySelector("#resume").onclick = () => send({ type: "unpause" });
connect();
