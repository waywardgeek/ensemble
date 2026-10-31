"use strict";

// One composer owns its local draft. An admission acknowledgment, not send(),
// authorizes clearing it. File readers and preview URLs die with their entries.
class ImageComposer {
    constructor(input, panel, status, discard, changed, notify) {
        this.input = input;
        this.panel = panel;
        this.status = status;
        this.discard = discard;
        this.changed = changed;
        this.notify = notify;
        this.entries = [];
        this.agent = null;
        this.pending = null;
        discard.onclick = () => this.clear(true);
    }
    get blocked() {
        return !!this.pending || this.entries.some(entry => entry.state !== "ready");
    }
    get hasImages() { return this.entries.length > 0; }
    render() {
        this.panel.replaceChildren();
        for (const entry of this.entries) {
            const figure = document.createElement("figure");
            const image = document.createElement("img");
            image.src = entry.url;
            image.alt = `Attached image ${this.entries.indexOf(entry) + 1}`;
            const caption = document.createElement("figcaption");
            caption.textContent = entry.state === "ready" ? entry.file.name || "Pasted image"
                : entry.state === "reading" ? "Reading image…" : "Read failed — remove this image";
            const remove = document.createElement("button");
            remove.type = "button";
            remove.className = "btn btn-quiet";
            remove.textContent = "Remove";
            remove.setAttribute("aria-label", `Remove image ${this.entries.indexOf(entry) + 1}`);
            remove.disabled = !!this.pending;
            remove.onclick = () => { this.remove(entry); this.render(); };
            figure.append(image, caption, remove);
            this.panel.append(figure);
        }
        this.panel.hidden = !this.hasImages;
        this.input.readOnly = !!this.pending;
        this.discard.hidden = !this.pending?.uncertain;
        this.status.textContent = this.pending?.uncertain
            ? "Connection lost: image send outcome unknown. Check History. This draft will not be resent."
            : this.pending ? "Sending images — waiting for admission…"
            : this.entries.some(entry => entry.state === "reading") ? "Reading pasted images…" : "";
        this.status.hidden = !this.status.textContent;
        this.changed();
    }
    remove(entry) {
        const index = this.entries.indexOf(entry);
        if (index < 0) return;
        this.entries.splice(index, 1); // Invalidate callbacks before aborting.
        if (entry.reader.readyState === FileReader.LOADING) entry.reader.abort();
        URL.revokeObjectURL(entry.url);
        if (!this.hasImages) this.agent = null;
    }
    clear(clearText = false) {
        for (const entry of [...this.entries]) this.remove(entry);
        this.pending = null;
        if (clearText) this.input.value = "";
        this.render();
    }
    selectionChanged(agent) {
        if (!this.hasImages || this.agent === agent) return;
        const wasPending = !!this.pending;
        this.clear(wasPending);
        this.notify(wasPending ? "Image draft cleared on selection change; the previous send may already have been admitted."
            : "Unsent image attachments discarded on Agent change.");
    }
    disconnect() {
        if (this.pending) {
            this.pending.uncertain = true;
            this.render();
        } else if (this.hasImages) {
            this.clear();
            this.notify("Disconnected: unsent images discarded. Text retained.");
        }
    }
    paste(event, agent) {
        const files = Array.from(event.clipboardData?.items || [])
            .filter(item => item.kind === "file").map(item => item.getAsFile()).filter(Boolean);
        if (!files.length) return;
        event.preventDefault();
        if (this.pending) return;
        // Image clipboard events can contain a caption too. Suppressing native
        // image paste must not suppress that accompanying plain text.
        const text = event.clipboardData.getData("text/plain");
        if (text) {
            this.input.setRangeText(text, this.input.selectionStart, this.input.selectionEnd, "end");
            this.input.dispatchEvent(new Event("input", { bubbles: true }));
        }
        const sourceBytes = [...this.entries.map(entry => entry.file), ...files]
            .reduce((total, file) => total + file.size, 0);
        if (this.entries.length + files.length > 4 || sourceBytes > 32 * 1024 * 1024) {
            this.notify("Paste refused: at most four images and 32 MiB total source data.");
            return;
        }
        if (files.some(file => !["image/png", "image/jpeg", "image/webp", "image/gif"].includes(file.type) || !file.size)) {
            this.notify("Paste refused: use nonempty PNG, JPEG, WebP or GIF files. The server validates pixels and refuses animation.");
            return;
        }
        this.agent = agent;
        for (const file of files) {
            const reader = new FileReader();
            const entry = { file, reader, url: URL.createObjectURL(file), state: "reading", data: null };
            this.entries.push(entry);
            reader.onload = () => {
                if (!this.entries.includes(entry) || entry.state !== "reading") return;
                const result = String(reader.result);
                entry.data = result.slice(result.indexOf(",") + 1);
                entry.state = "ready";
                this.render();
            };
            reader.onerror = () => {
                if (!this.entries.includes(entry)) return;
                entry.state = "failed";
                this.render();
            };
            reader.readAsDataURL(file);
        }
        this.render();
    }
    send(connection, agent) {
        if (!this.hasImages || this.blocked || this.agent !== agent) return false;
        if (new TextEncoder().encode(this.input.value).length > 32 * 1024) {
            this.notify("Image message text exceeds 32 KiB; shorten it before sending.");
            return false;
        }
        const sequence = (connection.imageSequence || 0) + 1;
        if (!Number.isSafeInteger(sequence)) return false;
        connection.imageSequence = sequence;
        this.pending = { connection, agent, sequence, uncertain: false };
        this.render(); // Disable repeated clicks before the socket write.
        try {
            connection.send(JSON.stringify({ type: "image_prompt", agent_id: agent, sequence,
                text: this.input.value, images: this.entries.map(entry => ({ mime: entry.file.type, data: entry.data })) }));
        } catch {
            this.pending.uncertain = true;
            this.render();
        }
        return true;
    }
    acknowledge(frame, connection) {
        const pending = this.pending;
        if (!pending || pending.uncertain || pending.connection !== connection) return false;
        const sequence = frame.type === "image_accepted" ? frame.sequence : frame.image_sequence;
        if (sequence !== pending.sequence) return false;
        if (frame.type === "image_accepted" && frame.agent_id === pending.agent) {
            const reduced = (frame.images || []).filter(image => image.resized);
            this.clear(true);
            this.notify(reduced.length ? `Images admitted; ${reduced.length} resized for delivery.` : "Images admitted.");
            return true;
        }
        if (frame.type === "error") {
            this.pending = null;
            this.render();
            this.notify("Image message refused; draft retained. " + (frame.message || ""));
            return true;
        }
        return false;
    }
}
