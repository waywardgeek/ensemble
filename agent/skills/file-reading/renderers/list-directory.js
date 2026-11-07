"use strict";
// Trusted shipped renderer. The host supplies recorded, bounded text, not a
// filesystem handle. Do not split lines into entries: filenames may contain LF.
(() => {
    const renderers = window.EnsembleArtifactRenderers ||= Object.create(null);
    renderers.list_directory = function directoryPreview(document, text) {
        const body = document.createElement("div");
        body.className = "directory-preview";
        const label = document.createElement("div");
        label.textContent = "Directory output";
        const output = document.createElement("pre");
        output.className = "tool-output";
        output.textContent = text;
        body.append(label, output);
        return body;
    };
})();
