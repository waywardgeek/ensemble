import { markdown, ansi, partText } from "./renderers.js";

// A scroll owns only its cards and supplied element, never application controls.
export class ArtifactScroll {
  cards = new Map();
  calls = new Map();
  constructor(element, speech = null) {
    // Speech is optional: a component consumer need not construct the chat shell.
    this.speech = speech;
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
    speaker.hidden = !this.speech;
    speaker.addEventListener("click", () => this.speech?.enqueue(card.text));
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
    if (!message || typeof message !== "object" || Array.isArray(message)) return;
    const id = message.part_id;
    if (message.type === "download_offered") {
      if (typeof message.token !== "string" || !/^[0-9a-f]{64}$/.test(message.token) ||
          typeof message.filename !== "string" || !message.filename.trim() ||
          typeof message.call_id !== "string" || !message.call_id.trim() ||
          (message.replay !== undefined && typeof message.replay !== "boolean")) return;
      const nearEnd =
        this.element.clientHeight + this.element.scrollTop >= this.element.scrollHeight - 180;
      const card = this.card(`download-${message.token}`, "download");
      const link = document.createElement("a");
      link.className = "artifact__download";
      link.setAttribute("role", "button");
      link.href = `/download/${message.token}`;
      link.download = message.filename;
      link.textContent = `Download ${message.filename}`;
      link.addEventListener("keydown", (event) => {
        if (event.key === " ") {
          event.preventDefault();
          link.click();
        }
      });
      card.text = link.textContent;
      card.content.replaceChildren(link);
      if (nearEnd) this.element.scrollTo(0, this.element.scrollHeight);
    } else if (message.type === "part_delta") {
      const card = this.card(id, message.kind);
      card.text += message.chunk;
      card.streamed = true;
      this.render(card);
      if (!message.replay && ["text", "thinking", "reasoning_summary"].includes(message.kind))
        this.speech?.chunk(id, message.chunk, false, message.kind);
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
        this.speech?.chunk(id, card.streamed ? "" : card.text, true, kind);
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
        this.speech?.flush();
        this.speech?.enqueue(message.name, "tool");
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
