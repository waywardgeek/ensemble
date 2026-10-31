// Render into caller-owned DOM nodes; importing this module touches no shell.
export function markdown(node, text) {
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
