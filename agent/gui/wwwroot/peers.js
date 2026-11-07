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
    let choicesPending = false;
    const offerings = new Map();
    const preferences = new Map();

    async function change(operation, address) {
        if (pending) { notice.textContent = "An Agent operation is already in progress."; return; }
        const body = { name: operation === "create" ? name.value : address };
        if (operation === "create") {
            if (choicesPending || create.disabled) return;
            body.skill = skill.value;
            const offering = offerings.get(model.selectedOptions[0]?.dataset.offeringKey);
            const preference = preferences.get(model.selectedOptions[0]?.dataset.preferenceKey);
            if (offering) body.offering = offering;
            else if (preference) body.preference = preference;
            else body.model = model.value;
        }
        pending = new AbortController();
        for (const control of [name, model, skill, create, open]) control.disabled = true;
        cancel.hidden = operation !== "create";
        dialog.close();
        notice.textContent = operation === "create"
            ? "Creating Agent. If a browser is needed, select it in the tree and open its separate relay tab."
            : "Stopping and deleting Agent. Saved files will remain.";
        try {
            const response = await fetch(`./peers/${operation}`, {
                method: "POST", credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "Content-Type": "application/json", "X-Ensemble-Peers": "1" },
                body: JSON.stringify(body), signal: pending.signal
            });
            if (!response.ok) {
                notice.textContent = "Agent operation refused. Check the name, primary skill, saved funding, storage, and unfinished children. No funding fallback was used. The mounted Agent cannot be deleted here.";
                return;
            }
            const result = await response.json();
            notice.textContent = operation === "create"
                ? `Created Agent ${result.name}. Select it in the tree.${result.warning ? " " + result.warning : ""}`
                : `Deleted Agent ${result.name}; saved files retained.`;
        } catch (_) {
            // Response cancellation is not proof that commit did not win the race.
            notice.textContent = "Request interrupted. Check the Agent tree before retrying; the operation may already have completed.";
        } finally {
            pending = null;
            cancel.hidden = true;
            for (const control of [name, model, skill, create, open]) control.disabled = false;
        }
    }
    open.onclick = async () => {
        if (pending || choicesPending) return;
        dialog.showModal(); name.focus();
        choicesPending = true;
        choiceNotice.textContent = "Loading approved models and primary skills…";
        create.disabled = model.disabled = skill.disabled = true;
        offerings.clear();
        preferences.clear();
        model.replaceChildren(new Option("Inherit mounted Agent’s model", ""));
        skill.replaceChildren(new Option("Inherit mounted Agent’s primary skill", ""));
        try {
            const response = await fetch("./peers/choices", {
                credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "X-Ensemble-Peers": "1" }, signal: AbortSignal.timeout(10000)
            });
            if (!response.ok) throw new Error("choices unavailable");
            const choices = await response.json();
            for (const value of choices.model_choices || []) model.add(new Option(value, value));
            const groups = new Map();
            for (const choice of choices.preferred_models || []) {
                let group = groups.get(choice.vendor);
                if (!group) {
                    group = document.createElement("optgroup"); group.label = choice.vendor;
                    groups.set(choice.vendor, group); model.append(group);
                }
                const key = `preference:${preferences.size}`;
                preferences.set(key, choice.preference);
                const option = new Option(`${choice.label} · remembered funding: ${choice.funding}`, key);
                option.dataset.preferenceKey = key;
                option.disabled = !choice.enabled;
                group.append(option);
            }
            for (const choice of choices.configured_models || []) {
                let group = groups.get(choice.vendor);
                if (!group) {
                    group = document.createElement("optgroup"); group.label = choice.vendor;
                    groups.set(choice.vendor, group); model.append(group);
                }
                const key = `offering:${offerings.size}`;
                offerings.set(key, choice.offering);
                const option = new Option(`${choice.label} · ${choice.funding}`, key);
                option.dataset.offeringKey = key;
                option.disabled = !choice.enabled;
                group.append(option);
            }
            for (const choice of choices.primary_skills || []) {
                const option = new Option(choice.name, choice.name);
                option.title = choice.description;
                skill.add(option);
            }
            create.disabled = model.disabled = skill.disabled = false;
            choiceNotice.textContent = "";
        } catch (_) {
            choiceNotice.textContent = "Agent choices unavailable. Close and reopen the dialog to retry.";
        } finally { choicesPending = false; }
    };
    document.querySelector("#peer-close").onclick = () => dialog.close();
    document.querySelector("#peer-form").onsubmit = event => { event.preventDefault(); void change("create"); };
    document.addEventListener("click", event => {
        const remove = event.target.closest("button[data-delete-agent]");
        if (!remove) return;
        if (pending) { notice.textContent = "An Agent operation is already in progress."; return; }
        const address = remove.dataset.deleteAgent;
        if (window.confirm(`Delete ${address}? Running work and owned descendants will be stopped. Saved files will remain.`)) {
            void change("delete", address);
        }
    });
    cancel.onclick = () => pending?.abort();
})();
