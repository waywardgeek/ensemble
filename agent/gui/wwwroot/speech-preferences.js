"use strict";
// Listening preferences belong to this browser profile, never to an Agent.
// Do not seed enabled from legacy server settings: unattended devices stay quiet.
class BrowserSpeech {
    constructor() {
        this.key = "ensemble.speech.v1";
        this.value = this.read(null);
        try { this.value = this.read(localStorage.getItem(this.key)); } catch { /* Page-session fallback. */ }
        window.addEventListener("storage", event => {
            try { if (event.storageArea !== localStorage) return; } catch { return; }
            if (event.key !== this.key && event.key !== null) return;
            const previous = this.value;
            this.value = this.read(event.newValue);
            this.onchange?.(previous);
        });
    }
    read(text) {
        let value;
        try { value = JSON.parse(text); } catch { /* Invalid values use silent defaults. */ }
        return {
            enabled: value?.enabled === true,
            speed: typeof value?.speed === "number" && Number.isFinite(value.speed)
                ? Math.min(4, Math.max(0.25, value.speed)) : 1,
            pause: value?.pause !== false,
            voice: typeof value?.voice === "string" ? value.voice : ""
        };
    }
    select(patch) {
        const previous = this.value;
        this.value = this.read(JSON.stringify({ ...this.value, ...patch }));
        try { localStorage.setItem(this.key, JSON.stringify(this.value)); } catch { /* Keep local controls usable. */ }
        this.onchange?.(previous);
    }
}
const browserSpeech = new BrowserSpeech();
