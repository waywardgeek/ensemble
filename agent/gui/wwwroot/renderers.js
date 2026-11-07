"use strict";
// Executable selection is shared; an unaccepted trial exists only in this page.
// This is trusted in-page code, not containment of a malicious renderer.
(() => {
    const byId = id => document.getElementById(id);
    const toolInput = byId("renderer-tool");
    let tool = "list_directory";
    const status = byId("renderer-status");
    const versionInput = byId("renderer-version");
    const source = byId("renderer-source");
    const fileInput = byId("renderer-package");
    const review = byId("renderer-review");
    const tryButton = byId("renderer-try");
    const keep = byId("renderer-keep");
    const rollback = byId("renderer-rollback");
    // Capture the fingerprinted script URL synchronously; currentScript is null
    // after an await. Missing shipped code leaves the host's plain-text fallback.
    const scriptUrl = document.currentScript.src;
    const shippedPaths = {
        prose: "shared/markdown.js",
        tool_default: "shared/tool-default.js",
        list_directory: "file-reading/renderers/list-directory.js",
        read_file: "file-reading/renderers/read-file.js",
        search_files: "file-reading/renderers/search-files.js",
        write_file: "code-tools/renderers/write-file.js",
        edit_file: "code-tools/renderers/edit-file.js"
    };
    window.EnsembleArtifactRenderers ||= Object.create(null);
    const shipped = Object.create(null);
    const safe = new URL(location.href);
    safe.searchParams.set("renderers", "shipped");
    byId("renderer-safe").href = safe.href;
    const safeMode = new URL(location.href).searchParams.get("renderers") === "shipped";
    let available = false, busy = false, reviewed = null, trial = null;
    let accepted = null, previous = null;
    const validVersion = value => typeof value === "string" && /^[a-f0-9]{64}$/.test(value);
    const endpoint = path => new URL("./renderers/" + path, location.href).href;
    function controls() {
        toolInput.disabled = busy || !!trial;
        fileInput.disabled = review.disabled = versionInput.disabled = busy || !available;
        tryButton.disabled = busy || !reviewed || !!trial || safeMode;
        keep.disabled = busy || !trial;
        rollback.disabled = busy || !available || (accepted === null && previous === null);
    }
    async function api(path, body, renderer = tool) {
        const url = new URL(endpoint(path));
        url.searchParams.set("renderer", renderer);
        const response = await fetch(url, {
            method: body === undefined ? "GET" : "POST", credentials: "same-origin", redirect: "error", cache: "no-store",
            headers: { "Content-Type": "application/json", "X-Ensemble-Renderer": "1" },
            body: body === undefined ? undefined : JSON.stringify(body), signal: AbortSignal.timeout(10000)
        });
        if (!response.ok) throw new Error("Renderer request refused");
        return response.json();
    }
    function selection(value) {
        if (!value || (value.accepted !== null && !validVersion(value.accepted)) ||
            (value.previous !== null && !validVersion(value.previous))) throw new Error("Invalid selection");
        accepted = value.accepted; previous = value.previous;
    }
    function install(render, renderer = tool) {
        window.EnsembleArtifactRenderers[renderer] = render;
        document.dispatchEvent(new Event("ensemble-renderers-changed"));
    }
    async function load(version, renderer = tool) {
        const pkg = await api("packages/" + version, undefined, renderer);
        if (pkg.renderer !== renderer || typeof pkg.entry !== "string" || !Object.hasOwn(pkg.files, pkg.entry))
            throw new Error("Invalid package");
        return pkg;
    }
    async function execute(version, pkg, renderer = tool) {
        const url = endpoint("modules/" + version + "/" + pkg.entry.split("/").map(encodeURIComponent).join("/"));
        const module = await import(url);
        if (typeof module.default !== "function") throw new Error("Renderer needs a default function export");
        install(module.default, renderer);
    }
    async function action(work, failure) {
        if (busy) return;
        busy = true; controls();
        try { await work(); }
        catch { status.textContent = failure; }
        finally { busy = false; controls(); }
    }
    async function inspect(version) {
        if (!validVersion(version)) throw new Error("Invalid version");
        const pkg = await load(version);
        reviewed = { version, pkg };
        source.textContent = "Entry: " + pkg.entry + "\n\n" + Object.entries(pkg.files)
            .map(([name, code]) => "--- " + name + " ---\n" + code).join("\n\n");
        status.textContent = "Candidate " + version + ". Review all files before Try live. Staging has not executed this code.";
    }
    versionInput.addEventListener("input", () => { reviewed = null; source.textContent = ""; controls(); });
    review.onclick = () => action(async () => {
        reviewed = null;
        await inspect(versionInput.value.trim());
    }, "Candidate could not be read. No trial started.");
    fileInput.onchange = () => action(async () => {
        reviewed = null;
        const file = fileInput.files[0];
        if (!file || file.size > 1024 * 1024) throw new Error("Package too large");
        const staged = await api("stage", JSON.parse(await file.text()));
        versionInput.value = staged.version;
        await inspect(staged.version);
    }, "Candidate staging or review failed. No trial started; staging may have completed.");
    tryButton.onclick = () => action(async () => {
        if (!reviewed || trial || safeMode) return;
        const candidate = reviewed;
        // Capture before importing: even a throwing module may have side effects.
        trial = { version: candidate.version, expected: accepted };
        status.textContent = "Trying candidate. Reload to recover if it fails.";
        try { await execute(candidate.version, candidate.pkg); }
        catch { trial = null; install(shipped[tool]); throw new Error("Trial failed"); }
        status.textContent = "Trial active: " + candidate.version + ". Keep explicitly, or reload to discard it.";
    }, "Candidate could not load. Shipped preview restored; reload to clear possible module side effects.");
    keep.onclick = () => action(async () => {
        if (!trial) return;
        const result = await api("accept", { version: trial.version, expected: trial.expected });
        selection(result.selection); trial = null;
        status.textContent = "Accepted " + accepted + ". Future pages use this version.";
    }, "Acceptance was not confirmed. Reload to check durable selection before trying again.");
    byId("renderer-revert").onclick = () => location.reload();
    rollback.onclick = () => action(async () => {
        await api("rollback", { expected: accepted });
        location.reload();
    }, "Rollback was not confirmed. Reload to check selection; no automatic retry was made.");
    function describe() {
        status.textContent = safeMode
            ? "Recovery mode: shipped renderers only. Accepted selection is unchanged; rollback remains available."
            : accepted === null ? "Using the shipped " + tool + " renderer." : "Using accepted version " + accepted + ".";
    }
    function downloadLink() {
        const url = new URL(endpoint("shipped")); url.searchParams.set("renderer", tool);
        byId("renderer-download").href = url.href;
        byId("renderer-download").download = tool + "-renderer.json";
    }
    toolInput.onchange = () => action(async () => {
        tool = toolInput.value;
        reviewed = null; accepted = previous = null; available = false;
        versionInput.value = source.textContent = fileInput.value = "";
        downloadLink();
        const state = await api("selection");
        available = state.available === true;
        if (!available) { status.textContent = "This host uses shipped renderers only."; return; }
        selection(state.selection); describe();
    }, "Renderer selection could not be read. Reload before trying again.");
    void action(async () => {
        const failures = [];
        for (const [renderer, path] of Object.entries(shippedPaths)) {
            try {
                const module = await import(new URL("./skills/" + path, scriptUrl).href);
                if (typeof module.default === "function") {
                    shipped[renderer] = module.default; install(module.default, renderer);
                }
            } catch { /* Each missing renderer independently keeps plain output. */ }
            try {
                const state = await api("selection", undefined, renderer);
                if (renderer === tool) {
                    available = state.available === true;
                    if (available) selection(state.selection);
                }
                if (!safeMode && state.available && state.selection.accepted !== null)
                    await execute(state.selection.accepted, await load(state.selection.accepted, renderer), renderer);
            } catch { failures.push(renderer); }
        }
        downloadLink();
        if (failures.length) {
            status.textContent = "Accepted renderer could not load: " + failures.join(", ") +
                ". Shipped rendering remains available. Reload with shipped renderers only to recover.";
        } else if (!available) status.textContent = "This host uses shipped renderers only.";
        else describe();
    }, "Renderer initialization failed. Reload with shipped renderers only to recover.");
})();
