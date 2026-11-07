// Bounded recorded text only. The host retains raw details and trusted controls.
export default function toolDefault(document, text) {
    const output = document.createElement("pre");
    output.className = "tool-output";
    output.textContent = text;
    return output;
}
