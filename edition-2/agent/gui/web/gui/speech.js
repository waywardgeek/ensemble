// Speech text is independent of the browser synthesizer. Its parent owns the
// audible queue and transcript; this child owns only unfinished part buffers.
export class SpeechBuffer {
  fragments = new Map();
  constructor(speech) {
    if (!speech) throw new Error("SpeechBuffer requires its speech owner");
    this.speech = speech;
  }
  chunk(id, text, source = "text") {
    const previous = this.fragments.get(id);
    this.fragments.set(id, { text: (previous?.text || "") + text, source });
    this.drain(id, false);
  }
  // Finals, tool dispatch and errors explicitly release a colon-ended tail.
  flush(id) {
    for (const key of id === undefined ? [...this.fragments.keys()] : [id])
      this.drain(key, true);
  }
  drain(id, final) {
    const part = this.fragments.get(id);
    if (!part) return;
    // Resolve fences before segmentation: their embedded newlines and periods
    // must never become phrase boundaries. Hold an unfinished fence across deltas.
    let text = part.text.replace(/```[^\n]*\n[\s\S]*?```/g, " code block ");
    if (text.includes("```")) {
      if (!final) return;
      text = text.replace(/```[\s\S]*$/, " code block ");
    }
    let end;
    while ((end = boundary(text)) > 0) {
      this.speech.enqueue(text.slice(0, end), part.source);
      text = text.slice(end);
    }
    if (final) {
      this.speech.enqueue(text, part.source);
      this.fragments.delete(id);
    } else part.text = text;
  }
}

// A single wrap is whitespace; punctuation and blank lines end phrases. Do not
// split inside inline code or links, whose syntax can span network fragments.
function boundary(text) {
  let code = false,
    brackets = 0,
    parentheses = 0;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (c === "`") code = !code;
    if (code) continue;
    if (c === "[") brackets++;
    if (c === "]") brackets = Math.max(0, brackets - 1);
    if (c === "(" && text[i - 1] === "]") parentheses++;
    if (c === ")") parentheses = Math.max(0, parentheses - 1);
    if (brackets || parentheses) continue;
    if (/[.!?]/.test(c) && (i + 1 === text.length || /\s/.test(text[i + 1])))
      return i + 1;
    if (c === "\n" && /^\n[ \t]*\n/.test(text.slice(i)))
      return i + text.slice(i).match(/^\n[ \t]*\n/)[0].length;
  }
  return 0;
}

// Filter only at utterance boundaries, after the fragments have reassembled.
// Identifier expansion precedes underscore removal so snake_case keeps its words.
export function speechText(text) {
  return text
    .replace(/```[^\n]*\n[\s\S]*?```/g, " code block ")
    .replace(/!?\[([^\]]*)\]\([^)]*\)/g, "$1")
    .replace(/^\s{0,3}#{1,6}\s+/gm, "")
    .replace(/^\s*([-*+] |\d+\. )/gm, "")
    .replace(/([A-Z]+)([A-Z][a-z])/g, "$1 $2")
    .replace(/([a-z\d])([A-Z])/g, "$1 $2")
    .replace(/([\p{L}\d])_+(?=[\p{L}\d])/gu, "$1 ")
    .replace(/[*_`]/g, "")
    .replace(/\s+/g, " ")
    .trim();
}
