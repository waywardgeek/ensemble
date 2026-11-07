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
            if (operation === "create") body.skill = skill.value;
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
            const shown = new Set((choices.preferred_models || []).map(choice =>
                `${choice.vendor}:${choice.model || choice.preference.model}`));
            for (const choice of choices.configured_models || []) {
                const identity = `${choice.vendor}:${choice.model}`;
                if (shown.has(identity)) continue;
                shown.add(identity);
                const matches = choices.configured_models.filter(value => `${value.vendor}:${value.model}` === identity);
                const key = `offering:${offerings.size}`;
                offerings.set(key, choice.offering);
                const option = new Option(choice.label, key);
                option.dataset.offeringKey = key;
                option.disabled = matches.length !== 1 || !choice.enabled;
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
