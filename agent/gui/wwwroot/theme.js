"use strict";
// Appearance belongs to this origin/browser profile, not to the selected Agent.
class BrowserTheme {
    constructor() {
        this.key = "ensemble.theme.v1";
        this.preference = null;
        this.system = matchMedia("(prefers-color-scheme: dark)");
        try { this.preference = this.valid(localStorage.getItem(this.key)); } catch { /* Session-only when storage is blocked. */ }
        this.system.addEventListener("change", () => this.render());
        window.addEventListener("storage", event => {
            // sessionStorage also emits storage events; it is not our preference.
            try { if (event.storageArea !== localStorage) return; } catch { return; }
            if (event.key !== this.key && event.key !== null) return;
            // Removal means use the browser's system preference, not server state.
            this.preference = this.valid(event.newValue) || "system";
            this.render();
        });
        document.querySelector("#theme-toggle").onclick = () => this.select(
            document.documentElement.dataset.bsTheme === "dark" ? "light" : "dark");
        document.querySelector("#theme").onchange = event => this.select(event.target.value);
        this.render();
    }
    valid(value) { return ["dark", "light", "system"].includes(value) ? value : null; }
    seed(legacyTheme) {
        if (this.preference !== null) return;
        this.select(this.valid(legacyTheme) || "system");
    }
    select(value) {
        if (!this.valid(value)) return;
        this.preference = value;
        try { localStorage.setItem(this.key, value); } catch { /* Keep this tab usable without durable storage. */ }
        this.render();
    }
    render() {
        const preference = this.preference || "system";
        const dark = preference === "dark" || (preference === "system" && this.system.matches);
        document.documentElement.dataset.theme = preference;
        document.documentElement.dataset.bsTheme = dark ? "dark" : "light";
        const button = document.querySelector("#theme-toggle");
        button.disabled = false;
        button.querySelector("use").setAttribute("href", dark ? "#icon-moon" : "#icon-sun");
        button.setAttribute("aria-label", dark ? "Switch to light theme" : "Switch to dark theme");
        button.title = dark ? "Dark theme; switch to light theme" : "Light theme; switch to dark theme";
        document.querySelector("#theme").value = preference;
    }
}
const browserTheme = new BrowserTheme();
