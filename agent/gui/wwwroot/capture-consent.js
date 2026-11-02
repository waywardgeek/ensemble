// One Agent-requested capture. The page owns this component and the shared
// PageCapture; this component cancels only capture work it actually started.
export class CaptureConsent extends EventTarget {
    #capture;
    #dialog;
    #requester;
    #status;
    #operation = null;

    constructor(capture, dialog) {
        super();
        this.#capture = capture;
        this.#dialog = dialog;
        this.#requester = dialog.querySelector('[data-capture-requester]');
        this.#status = dialog.querySelector('[data-capture-status]');
        dialog.querySelector('[data-capture-share]').addEventListener('click', event => {
            // gui_click and synthetic keyboard events cannot authorize pixels.
            if (!event.isTrusted) {
                this.#status.textContent = 'Please press Share yourself. An Agent cannot approve this request.';
                return;
            }
            this.#share();
        });
        dialog.querySelector('[data-capture-cancel]').addEventListener('click', () => this.cancel());
        dialog.addEventListener('cancel', event => {
            event.preventDefault();
            this.cancel();
        });
    }

    get busy() { return this.#operation !== null; }
    #error(message) { return { isError: true, content: [{ type: 'text', text: message }] }; }
    #changed() { this.dispatchEvent(new Event('change')); }

    request(connection, agent, id, label = agent) {
        // The canonical address of an unnamed mounted root is the empty string.
        // Its socket and RPC identity still bind the request; absence is different.
        if (!Number.isSafeInteger(id) || id <= 0 || typeof agent !== 'string' || connection?.readyState !== 1)
            return Promise.resolve(this.#error('Capture requires a live addressed request.'));
        if (!this.#capture.supported)
            return Promise.resolve(this.#error('Page capture is unavailable in this browser. Attach a manual screenshot instead.'));
        if (this.busy || this.#capture.busy || document.querySelector('dialog[open]'))
            return Promise.resolve(this.#error('Finish or cancel the current capture or dialog first.'));
        let resolve;
        const reply = new Promise(complete => { resolve = complete; });
        const operation = { connection, agent, id, resolve, started: false, cancelled: false,
            work: null, reader: null, cleanup: null };
        this.#operation = operation;
        this.#requester.textContent = String(label || 'current Agent').slice(0, 160);
        this.#status.textContent = 'Only this Ensemble tab may be shared. The screenshot returns to the requesting Agent, not to your draft.';
        try { this.#dialog.showModal(); }
        catch {
            this.#finish(operation, this.#error('The capture confirmation could not open.'));
        }
        this.#changed();
        return reply;
    }

    cancelMatching(connection, agent, id, message = 'Capture request cancelled.') {
        const operation = this.#operation;
        if (!operation || operation.connection !== connection || operation.agent !== agent || operation.id !== id)
            return Promise.resolve();
        return this.cancel(message);
    }

    cancelConnection(connection, message = 'Browser tool session closed.') {
        if (this.#operation?.connection !== connection) return Promise.resolve();
        return this.cancel(message);
    }

    cancel(message = 'Capture declined or cancelled.') {
        const operation = this.#operation;
        if (!operation) return Promise.resolve();
        if (operation.cleanup) return operation.cleanup;
        operation.cancelled = true; // Invalidate before stopping native work.
        let complete;
        operation.cleanup = new Promise(resolve => { complete = resolve; });
        if (this.#dialog.open) this.#dialog.close();
        if (operation.reader?.readyState === FileReader.LOADING) operation.reader.abort();
        let stopping;
        try { stopping = operation.started ? this.#capture.cancel() : Promise.resolve(); }
        catch (error) { stopping = Promise.reject(error); }
        Promise.allSettled([stopping, operation.work]).then(results => {
            const cleanupFailed = results[0].status === 'rejected';
            this.#finish(operation, this.#error(cleanupFailed ? 'Capture cleanup failed; no image was returned.' : message));
            complete();
        });
        this.#changed();
        return operation.cleanup;
    }

    #share() {
        const operation = this.#operation;
        if (!operation || operation.started || operation.cancelled) return;
        if (operation.connection.readyState !== 1 || this.#capture.busy) {
            this.cancel('The connection or capture state changed. Ask again when ready.');
            return;
        }
        operation.started = true;
        this.#dialog.close();
        // #perform calls native capture before its first await, preserving this
        // trusted click's browser activation. Native permission is still required.
        operation.work = this.#perform(operation);
        operation.work.then(result => {
            if (!operation.cancelled) this.#finish(operation, result);
        }, () => {
            if (!operation.cancelled)
                this.#finish(operation, this.#error('Page capture failed or permission was refused. No image was returned.'));
        });
        this.#changed();
    }

    #check(operation) {
        if (this.#operation !== operation || operation.cancelled || operation.connection.readyState !== 1)
            throw new DOMException('Capture is no longer current.', 'AbortError');
    }

    async #perform(operation) {
        const captured = await this.#capture.capture();
        this.#check(operation);
        if (captured.file?.type !== 'image/png' || captured.file.size <= 0 || captured.file.size > 32 * 1024 * 1024)
            throw new Error('Invalid capture image.');
        const reader = new FileReader();
        operation.reader = reader;
        try {
            const encoded = await new Promise((resolve, reject) => {
                reader.onload = () => resolve(reader.result);
                reader.onerror = () => reject(new Error('Capture encoding failed.'));
                reader.onabort = () => reject(new DOMException('Capture encoding cancelled.', 'AbortError'));
                reader.readAsDataURL(captured.file);
            });
            this.#check(operation);
            const prefix = 'data:image/png;base64,';
            if (typeof encoded !== 'string' || !encoded.startsWith(prefix)) throw new Error('Invalid capture encoding.');
            return { content: [{ type: 'image', mimeType: 'image/png', data: encoded.slice(prefix.length) }] };
        } finally {
            reader.onload = reader.onerror = reader.onabort = null;
            operation.reader = null;
        }
    }

    #finish(operation, result) {
        if (this.#operation !== operation) return;
        this.#operation = null;
        if (this.#dialog.open) this.#dialog.close();
        operation.resolve(result);
        this.#changed();
    }
}
