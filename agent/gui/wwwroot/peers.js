"use strict";
// HTTP management remains independent of the conversation socket. Closing the
// creation dialog leaves the tree and separate relay tab usable during startup.
(() => {
    const name = document.querySelector("#peer-name");
    const model = document.querySelector("#peer-model");
    const skill = document.querySelector("#peer-skill");
    const dialog = document.querySelector("#peer-dialog");
    const open = document.querySelector("#peer-open");
    const create = document.querySelector("#peer-create");
    const cancel = document.querySelector("#peer-cancel");
    const notice = document.querySelector("#peer-notice");
    const choiceNotice = document.querySelector("#peer-choice-status");
    const access = document.querySelector("#peer-access");
    const accessMode = document.querySelector("#peer-access-mode");
    const workspace = document.querySelector("#peer-workspace");
    const browser = document.querySelector("#peer-directory-browser");
    const folders = document.querySelector("#peer-directories");
    const directoryStatus = document.querySelector("#peer-directory-status");
    const up = document.querySelector("#peer-directory-up");
    const use = document.querySelector("#peer-directory-use");
    const imports = document.querySelector("#peer-imports");
    const importPath = document.querySelector("#peer-import-path");
    const importStatus = document.querySelector("#peer-import-status");
    let browseTarget = "workspace";
    function addImport(path, label = "Read-only import", available = true, selected = true) {
        const existing = [...imports.querySelectorAll("input")].find(input => input.value === path);
        if (existing) { if (!existing.disabled) existing.checked = selected; return; }
        const row = document.createElement("label"); row.className = "d-block";
        const input = document.createElement("input"); input.type = "checkbox"; input.value = path;
        input.checked = selected && available; input.disabled = !available;
        row.append(input, document.createTextNode(` ${label}: ${path}${available ? "" : " (unavailable)"}`));
        imports.append(row);
    }
    document.querySelector("#peer-import-add").onclick = () => {
        const path = importPath.value.trim();
        if (!path.startsWith("/") || path.length > 4096 || /[\x00-\x1f\x7f]/.test(path)) {
            importStatus.textContent = "Enter an absolute server directory path."; return;
        }
        if (imports.querySelectorAll("input").length >= 64) {
            importStatus.textContent = "At most 64 imports are supported."; return;
        }
        addImport(path); importPath.value = "";
        importStatus.textContent = "Selected. Access is validated when the Agent is created.";
    };
    let directoryPath = "", parentPath = null, directoryGeneration = 0;
    accessMode.onchange = () => {
        const confined = accessMode.value === "workspace";
        document.querySelector("#peer-workspace-controls").hidden = !confined;
        workspace.required = confined && !recovering;
    };
    async function browse(path = "") {
        const generation = ++directoryGeneration;
        browser.hidden = false;
        directoryStatus.textContent = "Loading server folders…";
        folders.replaceChildren(); up.disabled = use.disabled = true;
        try {
            const response = await fetch(`./peers/directories?path=${encodeURIComponent(path)}`, {
                credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "X-Ensemble-Peers": "1" }, signal: AbortSignal.timeout(10000)
            });
            if (!response.ok) throw new Error("directory unavailable");
            const result = await response.json();
            if (generation !== directoryGeneration) return;
            directoryPath = result.path; parentPath = result.parent;
            for (const folder of result.directories) folders.add(new Option(folder, folder));
            directoryStatus.textContent = result.path + (result.truncated ? " — first 200 folders; enter a path for others." : "");
            up.disabled = !parentPath; use.disabled = false;
        } catch (_) {
            if (generation === directoryGeneration) directoryStatus.textContent = "Folder unavailable. Choose another path or retry.";
        }
    }
    document.querySelector("#peer-browse").onclick = () => { browseTarget = "workspace"; void browse(); };
    document.querySelector("#peer-import-browse").onclick = () => { browseTarget = "import"; void browse(); };
    up.onclick = () => { if (parentPath) void browse(parentPath); };
    document.querySelector("#peer-directory-open").onclick = () => {
        if (folders.value) void browse(directoryPath.replace(/\/$/, "") + "/" + folders.value);
    };
    use.onclick = () => {
        if (browseTarget === "import") {
            addImport(directoryPath); importStatus.textContent = "Selected. Access is validated when the Agent is created.";
            importPath.focus();
        } else { workspace.value = directoryPath; workspace.focus(); }
        browser.hidden = true;
    };
    let pending = null;
    let recovering = null;
    let choicesPending = false;
    const offerings = new Map();
    const preferences = new Map();

    async function change(operation, address) {
        if (pending) { notice.textContent = "An Agent operation is already in progress."; return; }
        const body = { name: operation === "create" || operation === "restore" ? name.value : address };
        if (operation === "create" || operation === "restore") {
            if (choicesPending || create.disabled) return;
            if (operation === "create") {
                body.skill = skill.value;
                if (accessMode.value === "workspace") {
                    if (!workspace.reportValidity()) return;
                    body.sandbox_root = workspace.value;
                    body.process_read_paths = [...imports.querySelectorAll("input:checked")].map(input => input.value);
                }
            }
            const offering = offerings.get(model.selectedOptions[0]?.dataset.offeringKey);
            const preference = preferences.get(model.selectedOptions[0]?.dataset.preferenceKey);
            if (offering) body.offering = offering;
            else if (preference) body.preference = preference;
            else if (operation === "create") body.model = model.value;
        }
        pending = new AbortController();
        for (const control of [name, model, skill, create, open]) control.disabled = true;
        cancel.hidden = operation === "delete";
        dialog.close();
        notice.textContent = operation !== "delete"
            ? "Starting Agent. If a browser is needed, select it in the tree and open its separate relay tab."
            : "Stopping and deleting Agent. Saved files will remain.";
        try {
            const response = await fetch(`./peers/${operation}`, {
                method: "POST", credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "Content-Type": "application/json", "X-Ensemble-Peers": "1" },
                body: JSON.stringify(body), signal: pending.signal
            });
            if (!response.ok) {
                const result = await response.json().catch(() => ({}));
                notice.textContent = result.error || "Agent operation refused. Check model funding in Settings and retry. Saved conversations are retained.";
                return;
            }
            const result = await response.json();
            notice.textContent = operation !== "delete"
                ? `${operation === "restore" ? "Restored" : "Created"} Agent ${result.name}. Select it in the tree.`
                : `Deleted Agent ${result.name}; saved files retained.`;
        } catch (_) {
            // Response cancellation is not proof that commit did not win the race.
            notice.textContent = "Request interrupted. Check the Agent tree before retrying; the operation may already have completed.";
        } finally {
            pending = null;
            cancel.hidden = true;
            void refreshSaved();
            for (const control of [name, model, skill, create, open]) control.disabled = false;
        }
    }
    async function openDialog(saved = null) {
        if (pending || choicesPending) return;
        recovering = saved;
        access.hidden = !!saved;
        workspace.required = !saved && accessMode.value === "workspace";
        browser.hidden = true;
        ++directoryGeneration;
        name.readOnly = !!saved;
        if (saved) name.value = saved.name;
        document.querySelector("#peer-dialog-title").textContent = saved ? "Restore Agent" : "Create Agent";
        create.textContent = saved ? "Save and retry" : "Create";
        skill.hidden = !!saved;
        document.querySelector('label[for="peer-skill"]').hidden = !!saved;
        dialog.showModal(); name.focus();
        choicesPending = true;
        choiceNotice.textContent = "Loading approved models and primary skills…";
        create.disabled = model.disabled = skill.disabled = true;
        offerings.clear();
        preferences.clear();
        model.replaceChildren(new Option(saved ? "Retry saved model" : "Inherit mounted Agent’s model", ""));
        skill.replaceChildren(new Option("Inherit mounted Agent’s primary skill", ""));
        try {
            const response = await fetch("./peers/choices", {
                credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "X-Ensemble-Peers": "1" }, signal: AbortSignal.timeout(10000)
            });
            if (!response.ok) throw new Error("choices unavailable");
            const choices = await response.json();
            document.querySelector("#peer-inherited-access").textContent = choices.sandbox_root
                ? `Inherited workspace: ${choices.sandbox_root}`
                : "Inherited access is unrestricted. Choose a workspace to confine this Agent.";

            imports.replaceChildren(); importStatus.textContent = ""; importPath.value = "";
            for (const suggestion of choices.process_read_suggestions || [])
                addImport(suggestion.path, suggestion.label, suggestion.available, suggestion.selected);

            const groups = new Map();
            for (const choice of choices.preferred_models || []) {
                let group = groups.get(choice.vendor);
                if (!group) {
                    group = document.createElement("optgroup"); group.label = choice.vendor;
                    groups.set(choice.vendor, group); model.append(group);
                }
                const key = `preference:${preferences.size}`;
                preferences.set(key, choice.preference);
                const option = new Option(choice.label, key);
                option.dataset.preferenceKey = key;
                option.disabled = !choice.enabled;
                group.append(option);
            }
            const shown = new Set((choices.preferred_models || []).filter(choice => choice.enabled).map(choice =>
                `${choice.vendor}:${choice.model || choice.preference.model}`));
            for (const choice of choices.configured_models || []) {
                const identity = `${choice.vendor}:${choice.model}`;
                if (shown.has(identity)) continue;
                // An unresolved group must leave exact routes available for recovery.
                const key = `offering:${offerings.size}`;
                offerings.set(key, choice.offering);
                const option = new Option(choice.label, key);
                option.dataset.offeringKey = key;
                option.disabled = !choice.enabled;
                model.add(option);
            }
            if (!saved) for (const value of choices.model_choices || []) {
                if (!(choices.configured_models || []).some(choice => choice.model === value))
                    model.add(new Option(value, value));
            }
            for (const choice of choices.primary_skills || []) {
                const option = new Option(choice.name, choice.name);
                option.title = choice.description;
                skill.add(option);
            }
            create.disabled = model.disabled = skill.disabled = false;
            choiceNotice.textContent = "Funding is selected in Settings → AI.";
        } catch (_) {
            choiceNotice.textContent = "Agent choices unavailable. Close and reopen the dialog to retry.";
        } finally { choicesPending = false; }
    }
    open.onclick = () => openDialog();
    document.querySelector("#peer-close").onclick = () => dialog.close();
    document.querySelector("#peer-form").onsubmit = event => { event.preventDefault(); void change(recovering ? "restore" : "create"); };
    document.addEventListener("click", event => {
        const remove = event.target.closest("button[data-delete-agent]");
        if (!remove) return;
        if (pending) { notice.textContent = "An Agent operation is already in progress."; return; }
        const address = remove.dataset.deleteAgent;
        if (window.confirm(`Delete ${address}? Running work and owned descendants will be stopped. Saved files will remain.`)) {
            void change("delete", address);
        }
    });
    async function refreshSaved() {
        const section = document.querySelector("#saved-peers");
        const list = document.querySelector("#saved-peer-list");
        try {
            const response = await fetch("./peers/saved", {
                credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "X-Ensemble-Peers": "1" }, signal: AbortSignal.timeout(10000)
            });
            if (!response.ok) throw new Error("saved agents unavailable");
            const rows = await response.json();
            list.replaceChildren();
            section.hidden = rows.length === 0;
            for (const row of rows) {
                const item = document.createElement("p");
                const label = document.createElement("span");
                label.textContent = `${row.name}: ${row.message} `;
                const button = document.createElement("button");
                button.type = "button"; button.className = "btn btn-secondary";
                button.textContent = `Restore ${row.name}`;
                button.onclick = () => openDialog(row);
                item.append(label, button); list.append(item);
            }
        } catch (_) {
            section.hidden = false;
            list.textContent = "Saved agents could not be listed. Check storage and reload the page; files have been preserved.";
        }
    }
    cancel.onclick = () => pending?.abort();
    void refreshSaved();
})();
