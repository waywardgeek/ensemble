"use strict";
// This dialog moves no state into a second UI model. The existing controls and
// their acknowledgement/save handlers remain the sole editors of settings.
(() => {
    const dialog = document.querySelector("#settings-dialog");
    const open = document.querySelector("#settings-open");
    const close = document.querySelector("#settings-close");

    open.addEventListener("click", () => {
        if (dialog.open) return;
        dialog.showModal();
        void loadFunding();
        dialog.querySelector('[role="tab"][aria-selected="true"]').focus();
    });
    close.addEventListener("click", () => dialog.close());
    dialog.addEventListener("close", () => open.focus({ preventScroll: true }));
    dialog.addEventListener("shown.bs.tab", () => {
        dialog.querySelector(".settings-body").scrollTop = 0;
    });
    async function loadFunding() {
        const list = document.querySelector("#funding-choices");
        const status = document.querySelector("#funding-status");
        status.textContent = "Loading configured funding…";
        list.replaceChildren();
        try {
            const response = await fetch("./peers/choices", {
                credentials: "same-origin", redirect: "error", cache: "no-store",
                headers: { "X-Ensemble-Peers": "1" }, signal: AbortSignal.timeout(10000)
            });
            if (!response.ok) throw new Error("funding unavailable");
            const choices = (await response.json()).funding_choices || [];
            const groups = new Map();
            for (const choice of choices) {
                if (!groups.has(choice.vendor_id)) groups.set(choice.vendor_id, []);
                groups.get(choice.vendor_id).push(choice);
            }
            for (const rows of groups.values()) {
                const label = document.createElement("label");
                label.textContent = rows[0].vendor + " ";
                const select = document.createElement("select");
                select.className = "form-select";
                select.setAttribute("aria-label", rows[0].vendor + " default funding");
                select.add(new Option("Choose default funding", ""));
                for (const [index, row] of rows.entries()) {
                    const option = new Option(row.label, String(index));
                    option.selected = row.selected;
                    select.add(option);
                }
                // A sole available binding already resolves without a saved default.
                if (rows.length === 1 && !rows.some(row => row.selected)) select.value = "0";
                const savedStatus = document.createElement("span");
                savedStatus.setAttribute("role", "status");
                let confirmed = select.value;
                select.onchange = async () => {
                    if (select.value === "") return;
                    select.disabled = true;
                    savedStatus.textContent = " Saving…";
                    try {
                        const response = await fetch("./models/funding", {
                            method: "POST", credentials: "same-origin", redirect: "error", cache: "no-store",
                            headers: { "Content-Type": "application/json", "X-Ensemble-Peers": "1" },
                            body: JSON.stringify({ offering: rows[Number(select.value)].offering }),
                            signal: AbortSignal.timeout(10000)
                        });
                        if (!response.ok) throw new Error("save refused");
                        confirmed = select.value;
                        savedStatus.textContent = " Saved";
                        status.textContent = "Select a model to apply its default funding; running Agents retain their current funding.";
                        document.dispatchEvent(new Event("ensemble-funding-changed"));
                    } catch (_) {
                        select.value = confirmed;
                        savedStatus.textContent = " Save was not confirmed. Refresh funding to check, then choose again to retry.";
                    } finally { select.disabled = false; }
                };
                label.append(select, savedStatus); list.append(label);
            }
            status.textContent = choices.length ? "" : "No funding sources are configured for this host.";
        } catch (_) { status.textContent = "Funding settings unavailable. Refresh to retry."; }
    }
    document.querySelector("#funding-refresh").onclick = loadFunding;
})();
