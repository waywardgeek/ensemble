"use strict";
// Executable selection is shared; an unaccepted trial exists only in this page.
// This is trusted in-page code, not containment of a malicious renderer.
(() => {
    const byId = id => document.getElementById(id);
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
    const shippedUrl = new URL("./skills/file-reading/renderers/list-directory.js", scriptUrl).href;
    window.EnsembleArtifactRenderers ||= Object.create(null);
    let shipped;
    const safe = new URL(location.href);
    safe.searchParams.set("renderers", "shipped");
    byId("renderer-safe").href = safe.href;
    const safeMode = new URL(location.href).searchParams.get("renderers") === "shipped";
    let available = false, busy = false, reviewed = null, trial = null;
    let accepted = null, previous = null;
    const validVersion = value => typeof value === "string" && /^[a-f0-9]{64}$/.test(value);
    const endpoint = path => new URL("./renderers/" + path, location.href).href;
    function controls() {
        fileInput.disabled = review.disabled = versionInput.disabled = busy || !available;
        tryButton.disabled = busy || !reviewed || !!trial || safeMode;
        keep.disabled = busy || !trial;
        rollback.disabled = busy || !available || (accepted === null && previous === null);
    }
    async function api(path, body) {
        const response = await fetch(endpoint(path), {
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
    function install(render) {
        window.EnsembleArtifactRenderers.list_directory = render;
        document.dispatchEvent(new Event("ensemble-renderers-changed"));
    }
    async function load(version) {
        const pkg = await api("packages/" + version);
        if (pkg.renderer !== "list_directory" || typeof pkg.entry !== "string" || !Object.hasOwn(pkg.files, pkg.entry))
            throw new Error("Invalid package");
        return pkg;
    }
    async function execute(version, pkg) {
        const url = endpoint("modules/" + version + "/" + pkg.entry.split("/").map(encodeURIComponent).join("/"));
        const module = await import(url);
        if (typeof module.default !== "function") throw new Error("Renderer needs a default function export");
        install(module.default);
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
        catch { trial = null; install(shipped); throw new Error("Trial failed"); }
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
    void action(async () => {
        try {
            const module = await import(shippedUrl);
            if (typeof module.default === "function") { shipped = module.default; install(shipped); }
        } catch { /* Plain-text fallback remains available. */ }
        for (const [tool, path] of [
            ["read_file", "file-reading/renderers/read-file.js"],
            ["search_files", "file-reading/renderers/search-files.js"],
            ["write_file", "code-tools/renderers/write-file.js"],
            ["edit_file", "code-tools/renderers/edit-file.js"]]) {
            try {
                const module = await import(new URL("./skills/" + path, scriptUrl).href);
                if (typeof module.default === "function") window.EnsembleArtifactRenderers[tool] = module.default;
            } catch { /* Each missing renderer independently keeps plain output. */ }
        }
        document.dispatchEvent(new Event("ensemble-renderers-changed"));
        const state = await api("selection");
        available = state.available === true;
        if (!available) { status.textContent = "This host uses shipped renderers only."; return; }
        selection(state.selection);
        if (safeMode) {
            status.textContent = "Recovery mode: shipped renderers only. Accepted selection is unchanged; rollback remains available.";
            return;
        }
        if (accepted !== null) await execute(accepted, await load(accepted));
        status.textContent = accepted === null ? "Using the shipped directory renderer." : "Using accepted version " + accepted + ".";
    }, "Accepted renderer could not load. Shipped rendering remains available. Reload with shipped renderers only to recover.");
})();
