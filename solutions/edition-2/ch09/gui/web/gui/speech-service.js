// Pages own logical queues. This application child also holds a browser lease
// before using native synthesis: Chrome shares its speech queue across tabs.
// Cooperating documents in one origin/storage bucket use the same lock name.
export class SpeechService {
  constructor(parent) { this.parent = parent; this.pending = []; this.active = null; this.closed = false; }
  available() { return !this.closed && !!this.parent.synthesis && !!this.parent.Utterance && !!this.parent.locks?.request; }
  submit(page, text, callbacks, rate = 1) {
    if (!this.available()) { callbacks.end('Speech synthesis or native coordination unavailable'); return; }
    this.pending.push({page, text, callbacks, rate}); this.pump();
  }
  pump() {
    if (this.closed || this.active || !this.pending.length) return;
    const request = this.pending.shift(); this.active = request;
    request.wait = new AbortController();
    try {
      this.parent.locks.request('ensemble-native-speech', {signal: request.wait.signal}, () => {
        if (this.active !== request || this.closed) return;
        return new Promise(release => { request.release = release; this.speak(request); });
      }).catch(error => this.finish(request, error.message));
    } catch (error) { this.finish(request, error.message); }
  }
  speak(request) {
    try {
      const utterance = new this.parent.Utterance(request.text); utterance.rate = request.rate;
      utterance.onstart = () => { if (this.active === request) { clearTimeout(request.timer); request.callbacks.start(); } };
      utterance.onend = () => this.finish(request);
      utterance.onerror = event => this.finish(request, event.error || 'unknown');
      // Waiting for another document's lease is not blocked playback. Start the
      // bounded no-start check only after this service owns the native request.
      request.timer = setTimeout(() => {
        if (this.active !== request) return;
        this.cancel(request.page);
        request.callbacks.end('Playback did not start; enable playback on this page');
      }, 5000);
      this.parent.synthesis.speak(utterance);
    } catch (error) { this.finish(request, error.message); }
  }
  finish(request, error) {
    if (this.active !== request) return;
    clearTimeout(request.timer); this.active = null; request.release?.();
    request.callbacks.end(error); this.pump();
  }
  cancel(page) {
    this.pending = this.pending.filter(request => request.page !== page);
    if (this.active?.page !== page) return;
    const request = this.active; clearTimeout(request.timer); this.active = null;
    request.wait.abort();
    // A waiting or stale request never owns native cancellation. Fence callbacks
    // first, then cancel while the lease is still held, and release after cleanup.
    if (request.release) {
      try { this.parent.synthesis.cancel(); } catch (error) { this.parent.diagnostic(error.message); }
      request.release();
    }
    this.pump();
  }
  close() {
    if (this.closed) return;
    this.closed = true; this.pending = [];
    if (this.active) this.cancel(this.active.page);
  }
}
