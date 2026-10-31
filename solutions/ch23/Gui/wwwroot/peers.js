"use strict";
// HTTP management runs beside the WebSocket: pending browser startup still
// needs tree selection and its independent relay tab to remain usable.
(() => {
    const name = document.querySelector("#peer-name");
    const model = document.querySelector("#peer-model");
    const skill = document.querySelector("#peer-skill");
    const create = document.querySelector("#peer-create");
    const remove = document.querySelector("#peer-delete");
    const cancel = document.querySelector("#peer-cancel");
    const confirm = document.querySelector("#peer-delete-confirm");
    const notice = document.querySelector("#peer-notice");
    let pending = null;

    async function change(operation) {
        if (pending) return;
        if (!name.value.trim()) {
            notice.textContent = "Enter a peer name.";
            name.focus();
            return;
        }
        if (operation === "delete" && !confirm.checked) {
            notice.textContent = "Confirm deletion of the named peer's registry entry first.";
            return;
        }
        const body = { name: name.value };
        if (operation === "create") Object.assign(body, { model: model.value, skill: skill.value });
        pending = new AbortController();
        for (const control of [name, model, skill, create, remove, confirm]) control.disabled = true;
        cancel.hidden = operation !== "create";
        notice.textContent = operation === "create"
            ? "Creating peer. If its skill needs a browser, select it in the conversation picker and open its independent relay."
            : "Stopping and deleting peer. Saved files will remain.";
        try {
            const response = await fetch(`./peers/${operation}`, {
                method: "POST", credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "Content-Type": "application/json", "X-Ensemble-Peers": "1" },
                body: JSON.stringify(body), signal: pending.signal
            });
            if (!response.ok) {
                notice.textContent = "Peer operation refused. Check the name, primary skill, storage, and unfinished children. The mounted Agent cannot be deleted here.";
                return;
            }
            const result = await response.json();
            notice.textContent = operation === "create"
                ? `Created peer ${result.name}. Select it to view its conversation.`
                : `Deleted peer ${result.name}; saved files retained.`;
        } catch (_) {
            // Aborting a response is not proof that commit did not win the race.
            notice.textContent = "Request interrupted. Check the Agent tree before retrying; an operation may already have completed.";
        } finally {
            pending = null;
            cancel.hidden = true;
            confirm.checked = false;
            for (const control of [name, model, skill, create, remove, confirm]) control.disabled = false;
        }
    }
    create.onclick = () => change("create");
    remove.onclick = () => change("delete");
    cancel.onclick = () => pending?.abort();
})();
