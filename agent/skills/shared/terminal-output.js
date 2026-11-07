// Bounded terminal-log presentation, not a terminal emulator. No HTML or links.
// Each rolling snapshot starts with unknown prior terminal state and resets SGR.
export function terminalRuns(bytes, rolledOut = false) {
    if (!(bytes instanceof Uint8Array) || bytes.length > 65536) throw new Error("Invalid terminal window");
    let start = 0;
    if (rolledOut) while (start < bytes.length && (bytes[start] & 0xc0) === 0x80) start++;
    // A running writer may end mid-codepoint. Defer that suffix until a later
    // snapshot rather than briefly displaying a replacement character.
    const decoder = new TextDecoder("utf-8");
    const text = decoder.decode(bytes.subarray(start), { stream: true });
    const runs = [];
    let bold = false, foreground = null, background = null, buffer = "";
    function flush() {
        if (buffer) runs.push({ text: buffer, bold, foreground, background });
        buffer = "";
    }
    function sgr(sequence) {
        const values = sequence === "" ? [0] : sequence.split(";").map(Number);
        for (let n = 0; n < values.length; n++) {
            const value = values[n];
            if (value === 0) { bold = false; foreground = background = null; }
            else if (value === 1) bold = true;
            else if (value === 22) bold = false;
            else if (value === 39) foreground = null;
            else if (value === 49) background = null;
            else if (value >= 30 && value <= 37) foreground = value - 30;
            else if (value >= 90 && value <= 97) foreground = value - 90 + 8;
            else if (value >= 40 && value <= 47) background = value - 40;
            else if (value >= 100 && value <= 107) background = value - 100 + 8;
            // Skip unsupported extended-color arguments as a unit; their values
            // must not accidentally become independent SGR commands.
            else if (value === 38 || value === 48) {
                if (values[n + 1] === 5) n += 2;
                else if (values[n + 1] === 2) n += 4;
                else break;
            }
        }
    }
    for (let i = 0; i < text.length;) {
        const c = text[i++];
        if (c === "\x1b") {
            flush();
            const next = text[i++];
            if (next === "[") {
                const begin = i;
                while (i < text.length && !(text.charCodeAt(i) >= 0x40 && text.charCodeAt(i) <= 0x7e)) i++;
                if (i < text.length) {
                    const command = text[i++];
                    const argumentsText = text.slice(begin, i - 1);
                    if (command === "m" && /^[0-9;]*$/.test(argumentsText)) sgr(argumentsText);
                }
            } else if (next === "]" || next === "P" || next === "^" || next === "_") {
                // OSC/DCS strings (including hyperlinks) remain inert and hidden.
                while (i < text.length) {
                    if (text[i++] === "\x07") break;
                    if (text[i - 1] === "\x1b" && text[i] === "\\") { i++; break; }
                }
            }
        } else if (c === "\r") {
            // Preserve progress output as log lines; do not emulate cursor edits.
            if (text[i] !== "\n") buffer += "\n";
        } else if (c === "\n" || c === "\t" || c.charCodeAt(0) >= 32 && c.charCodeAt(0) !== 127) buffer += c;
    }
    flush();
    return runs;
}
