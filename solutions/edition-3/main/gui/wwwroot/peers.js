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
    let pending = null;

    async function change(operation, address) {
        if (pending) { notice.textContent = "An Agent operation is already in progress."; return; }
        const body = { name: operation === "create" ? name.value : address };
        if (operation === "create") Object.assign(body, { model: model.value, skill: skill.value });
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
                notice.textContent = "Agent operation refused. Check the name, primary skill, storage, and unfinished children. The mounted Agent cannot be deleted here.";
                return;
            }
            const result = await response.json();
            notice.textContent = operation === "create"
                ? `Created Agent ${result.name}. Select it in the tree.`
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
    open.onclick = () => { dialog.showModal(); name.focus(); };
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
