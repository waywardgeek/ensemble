import { terminalRuns } from "./terminal-output.js";

// Presentation only: the GUI owns window validation, notices and Job lifetime.
export default function render(parent, bytes, rolledOut) {
    const nodes = terminalRuns(bytes, rolledOut).map(run => {
        const span = parent.ownerDocument.createElement("span");
        span.textContent = run.text;
        if (run.bold) span.classList.add("terminal-bold");
        if (run.foreground !== null) span.classList.add(`terminal-fg-${run.foreground}`);
        if (run.background !== null) span.classList.add(`terminal-bg-${run.background}`);
        return span;
    });
    parent.replaceChildren(...nodes);
}
