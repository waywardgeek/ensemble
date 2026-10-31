"use strict";
// Account management is host UI, never conversation or browser-MCP output. Only
// redacted status reaches the DOM. In particular, a reauthentication URL can
// contain an ID-token hint: keep it request-local and navigate a separate window.
(() => {
    const panel = document.querySelector("#accounts-panel");
    const status = document.querySelector("#account-status");
    const notice = document.querySelector("#account-notice");
    const list = document.querySelector("#account-list");
    const label = document.querySelector("#account-label");
    const modelStatus = document.querySelector("#account-model-status");
    const models = document.querySelector("#account-model-list");
    const controls = [...panel.querySelectorAll("button")];

    async function request(operation, body) {
        const options = {
            method: body === undefined ? "GET" : "POST",
            headers: { "X-Ensemble-Auth": "1" },
            cache: "no-store",
            credentials: "same-origin",
            redirect: "error"
        };
        if (body !== undefined) {
            options.headers["Content-Type"] = "application/json";
            options.body = JSON.stringify(body);
        }
        const response = await fetch(`./auth/${operation}`, options);
        if (!response.ok) {
            // Neither an exception nor arbitrary HTTP text should leak an auth
            // URL to the console, speech queue, or snapshot-visible notice.
            throw new Error("Account operation failed. Refresh status; cancel a pending sign-in before retrying.");
        }
        return await response.json();
    }

    async function refresh() {
        const value = await request("status");
        status.textContent = `Running billing: ${value.billing}. Selected account: ${value.selected_account || "none"}. `
            + (value.enabled ? (value.pending ? "Sign-in pending." : "No sign-in pending.")
                : "Account management is disabled; launch with --credentials-dir.");
        list.replaceChildren();
        for (const account of value.accounts) {
            const item = document.createElement("li");
            item.textContent = `${account.label}${account.email ? ` (${account.email})` : ""}: `
                + (account.signed_in ? (account.plan_enabled ? "signed in, plan permission granted" : "signed in, no plan permission")
                    : "signed out");
            list.append(item);
        }
        // Controls reflect acknowledged availability, not local account state.
        for (const button of controls) {
            button.disabled = button.id !== "account-refresh" && !value.enabled;
        }
    }

    function clearModels() {
        models.replaceChildren();
        modelStatus.textContent = "No catalog fetched.";
    }

    async function showModels() {
        clearModels();
        notice.textContent = "";
        const selected = label.value;
        if (!selected.trim()) {
            modelStatus.textContent = "Enter an account label first.";
            label.focus();
            return;
        }
        for (const button of controls) button.disabled = true;
        modelStatus.textContent = "Fetching plan catalog.";
        try {
            const result = await request("models", { label: selected });
            // The text field remains editable; a response belongs to the label
            // that requested it, not whichever account is now being entered.
            if (label.value !== selected) return;
            for (const model of result.models) {
                const item = document.createElement("li");
                item.textContent = `${model.display_name} — ${model.slug}`;
                models.append(item);
            }
            modelStatus.textContent = result.models.length
                ? `Fetched ${result.models.length} plan models for ${selected}.`
                : `No visible plan models for ${selected}.`;
        } catch (_) {
            if (label.value === selected) {
                modelStatus.textContent = "Plan catalog unavailable. Check this account's sign-in and plan permission, then retry.";
            }
        } finally {
            await refreshSafely();
        }
    }

    async function perform(operation) {
        clearModels();
        notice.textContent = "";
        for (const button of controls) button.disabled = true;
        try {
            const body = operation === "cancel" ? {} : { label: label.value };
            const result = await request(operation, body);
            notice.textContent = result.message;
        } catch (_) {
            notice.textContent = "Account operation failed. Refresh status; cancel a pending sign-in before retrying.";
        } finally {
            await refreshSafely();
        }
    }

    async function refreshSafely() {
        try {
            await refresh();
        } catch (_) {
            status.textContent = "Account status unavailable.";
            // Fail closed: only refresh is available until the host responds.
            for (const button of controls) button.disabled = button.id !== "account-refresh";
        }
    }

    label.addEventListener("input", clearModels);
    document.querySelector("#account-models").onclick = showModels;
    document.querySelector("#account-refresh").onclick = async () => {
        clearModels();
        await refreshSafely();
    };
    document.querySelector("#account-cancel").onclick = () => perform("cancel");
    document.querySelector("#account-logout").onclick = () => perform("logout");
    document.querySelector("#account-signout").onclick = () => perform("signout");
    document.querySelector("#account-login").onclick = async () => {
        clearModels();
        notice.textContent = "";
        if (!label.value.trim()) {
            notice.textContent = "Enter an account label first.";
            label.focus();
            return;
        }
        // Open during the human gesture, before awaiting HTTP. No URL fallback
        // link: rendering one would expose private hints to DOM snapshots.
        const popup = window.open("about:blank", "_blank");
        if (!popup) {
            notice.textContent = "Allow a sign-in window, then try again. No sign-in was started.";
            return;
        }
        popup.opener = null;
        for (const button of controls) button.disabled = true;
        try {
            const result = await request("login", {
                label: label.value,
                consent: document.querySelector("#account-consent").checked
            });
            if (popup.closed) {
                throw new Error("Sign-in window closed.");
            }
            popup.location.replace(result.url);
            notice.textContent = "Complete sign-in in the separate window, then refresh status here. Cancel explicitly if you abandon it.";
        } catch (_) {
            popup.close();
            // A network failure may follow a successful Begin on the host. Do
            // not auto-cancel: another browser's attempt might now own the slot.
            notice.textContent = "Could not open sign-in. Refresh status and cancel any pending attempt before retrying.";
        } finally {
            await refreshSafely();
        }
    };
    refreshSafely();
})();
