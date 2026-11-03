"use strict";
// Presentation and keyboard choices belong to this origin/browser profile.
class BrowserDevice {
    constructor() {
        this.key = "ensemble.device.v1";
        this.value = this.read(null);
        try { this.value = this.read(localStorage.getItem(this.key)); } catch { /* Page fallback. */ }
        window.addEventListener("storage", event => {
            try { if (event.storageArea !== localStorage) return; } catch { return; }
            if (event.key !== this.key && event.key !== null) return;
            // Clearing storage resets locally; it must not revive a server preference.
            this.value = this.read(event.newValue);
            if (this.value.font === null) this.value.font = 18;
            this.onchange?.();
        });
    }
    read(text) {
        let value;
        try { value = JSON.parse(text); } catch { /* Invalid input uses defaults. */ }
        return {
            font: Number.isInteger(value?.font) ? Math.max(12, Math.min(32, value.font)) : null,
            shortcut: ["AltRight", "F8", "off"].includes(value?.shortcut) ? value.shortcut : "off"
        };
    }
    seed(font) {
        if (this.value.font !== null) return;
        this.select({ font: Number.isInteger(font) ? font : 18 });
    }
    select(patch) {
        this.value = this.read(JSON.stringify({ ...this.value, ...patch }));
        try { localStorage.setItem(this.key, JSON.stringify(this.value)); } catch { /* Page fallback. */ }
        this.onchange?.();
    }
}
const browserDevice = new BrowserDevice();
