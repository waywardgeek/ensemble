// Pages own logical queues; this application child serializes the shared native
// API. Only the active request's owner may invoke native cancellation.
export class SpeechService {
  constructor(parent) { this.parent = parent; this.pending = []; this.active = null; this.closed = false; }
  available() { return !this.closed && !!this.parent.synthesis && !!this.parent.Utterance; }
  submit(page, text, callbacks) {
    if (!this.available()) { callbacks.end('Speech synthesis unavailable'); return; }
    this.pending.push({page, text, callbacks}); this.pump();
  }
  pump() {
    if (this.closed || this.active || !this.pending.length) return;
    const request = this.pending.shift(); this.active = request;
    const finish = error => {
      if (this.active !== request) return;
      this.active = null; request.callbacks.end(error); this.pump();
    };
    try {
      const utterance = new this.parent.Utterance(request.text);
      utterance.onstart = () => { if (this.active === request) request.callbacks.start(); };
      utterance.onend = () => finish();
      utterance.onerror = event => finish(event.error || 'unknown');
      this.parent.synthesis.speak(utterance);
    } catch (error) { finish(error.message); }
  }
  cancel(page) {
    this.pending = this.pending.filter(request => request.page !== page);
    if (this.active?.page !== page) return;
    // Fence callbacks before cancel: some engines call them synchronously.
    this.active = null;
    try { this.parent.synthesis.cancel(); } catch (error) { this.parent.diagnostic(error.message); }
    this.pump();
  }
  close() {
    if (this.closed) return;
    this.closed = true; this.pending = [];
    if (this.active) this.cancel(this.active.page);
  }
}
