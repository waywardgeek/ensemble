// Render into caller-owned DOM nodes; importing this module touches no shell.
export function markdown(node, text, { copyable = false, boxes = [] } = {}) {
  // A deliberately small Markdown subset: headings, bullets, fences, bold and
  // inline code. All source text remains reachable even when syntax is unknown.
  let fence = null;
  let lines = [];
  for (const line of text.split("\n")) {
    if (fence) {
      const closing = line.match(/^(`{3,})\s*$/);
      if (closing && closing[1].length >= fence.length) {
        const value = lines.join("\n");
        if (copyable) node.append(clipboardBox(value, boxes));
        else {
          const pre = document.createElement("pre");
          pre.textContent = value;
          node.append(pre);
        }
        fence = null;
        lines = [];
      } else lines.push(line);
      continue;
    }
    const opening = line.match(/^(`{3,})([^`]*)$/);
    if (opening) {
      fence = { length: opening[1].length };
      continue;
    }
    const heading = line.match(/^(#{1,6})\s+(.*)/);
    const row = document.createElement(
      heading ? `h${heading[1].length}` : "p",
    );
    const value = heading ? heading[2] : line.replace(/^[-*] /, "• ");
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
    node.append(row);
  }
  if (fence) {
    const pre = document.createElement("pre");
    pre.textContent = lines.join("\n");
    node.append(pre);
  }
}

function clipboardBox(text, boxes) {
  const index = boxes.findIndex(box => box.querySelector("textarea").value === text);
  if (index !== -1) return boxes.splice(index, 1)[0];
  const box = document.createElement("div");
  box.className = "clipboard-block";
  const field = document.createElement("textarea");
  field.setAttribute("aria-label", "Copyable text");
  field.readOnly = true;
  field.value = text;
  field.rows = Math.min(12, Math.max(2, text.split("\n").length));
  const copy = document.createElement("button");
  copy.type = "button";
  copy.textContent = "Copy";
  copy.setAttribute("aria-label", "Copy text");
  const status = document.createElement("span");
  status.setAttribute("role", "status");
  let copying = false;
  copy.addEventListener("click", async () => {
    if (copying) return;
    const value = field.value;
    copying = true;
    copy.setAttribute("aria-busy", "true");
    status.textContent = "";
    try {
      await copyText(value);
      status.textContent = "Copied";
    } catch {
      if (field.isConnected) {
        field.focus();
        field.select();
      }
      status.textContent = "Copy blocked. Press Ctrl+C or Command+C to copy the selected text.";
    } finally {
      copying = false;
      copy.removeAttribute("aria-busy");
    }
  });
  box.append(field, copy, status);
  return box;
}

async function copyText(text) {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return;
    } catch {}
  }
  const focused = document.activeElement;
  const selection = document.getSelection();
  const ranges = [];
  for (let index = 0; index < (selection?.rangeCount || 0); index++)
    ranges.push(selection.getRangeAt(index).cloneRange());
  const field = document.createElement("textarea");
  field.value = text;
  field.readOnly = true;
  field.tabIndex = -1;
  field.setAttribute("aria-hidden", "true");
  field.style.position = "fixed";
  field.style.opacity = "0";
  document.body.append(field);
  let copied = false;
  try {
    field.focus({ preventScroll: true });
    field.select();
    copied = document.execCommand("copy");
  } finally {
    field.remove();
    focused?.focus({ preventScroll: true });
    selection?.removeAllRanges();
    for (const range of ranges) selection.addRange(range);
  }
  if (!copied) throw new Error("Clipboard unavailable");
}
// Recognize display colour only, without interpreting arbitrary terminal
// controls or treating shell output as HTML. Unknown colour codes reset style.
export function ansi(node, text) {
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
export function partText(part) {
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

