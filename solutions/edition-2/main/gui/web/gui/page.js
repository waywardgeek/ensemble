import {Connector} from './connector.js';
import {ArtifactScroll} from './artifacts.js';
import {SpeechQueue} from './speech.js';

// Public page controller: an embedding may provide a different layout using the
// same labeled elements, or own Connector and ArtifactScroll directly.
export class Page {
  constructor(root, url) {
    this.root = root; this.input = root.querySelector('[data-input]'); this.status = root.querySelector('[data-status]');
    this.notice = root.querySelector('[data-notice]'); this.pauseStatus = root.querySelector('[data-pause]');
    this.artifacts = new ArtifactScroll(this, root.querySelector('[data-artifacts]'));
    this.speech = new SpeechQueue(this); this.connector = new Connector(this, url); this.connected = false;
    this.input.addEventListener('input', () => this.reconcile().catch(e => this.diagnostic(e.message)));
    this.input.addEventListener('keydown', e => { if (e.key === 'Escape' && this.input.value === '') { e.preventDefault(); this.speech.cancel(); } });
    root.querySelector('[data-prompt]').addEventListener('click', () => this.submit('prompt'));
    root.querySelector('[data-hint]').addEventListener('click', () => this.submit('hint'));
    root.querySelector('[data-interrupt]').addEventListener('click', () => this.connector.send('interrupt').catch(e => this.diagnostic(e.message)));
    root.querySelector('[data-latest]').addEventListener('click', () => this.artifacts.latest());
    root.querySelector('[data-cancel-speech]').addEventListener('click', () => this.speech.cancel());
    const auto = root.querySelector('[data-auto-speech]'); auto.setAttribute('aria-pressed', 'false');
    auto.addEventListener('click', () => { this.speech.enable(!this.speech.enabled); auto.setAttribute('aria-pressed', String(this.speech.enabled)); auto.textContent = this.speech.enabled ? 'Auto speech: on' : 'Auto speech: off'; });
    this.connector.connect();
  }
  diagnostic(message) { this.notice.textContent = message; }
  speechEvent(detail) { this.root.dispatchEvent(new CustomEvent('ensemble-speech', {detail})); }
  connection(status) {
    this.status.textContent = status;
    this.connected = status.startsWith('Connected'); this.lastCauses = null;
    if (!this.connected) { this.speech.cancel(); this.artifacts.incomplete(); this.pauseStatus.textContent = 'No pause held by this page'; }
    else this.reconcile().catch(e => this.diagnostic(e.message));
  }
  snapshot(snapshot) { this.speech.reset(); this.speech.seed(snapshot.partials); this.artifacts.reset(snapshot); this.pause(snapshot.state); }
  observation(o) { this.artifacts.observation(o); this.speech.observe(o); if (o.kind === 'pause_changed') this.pause(o); }
  pause(state) { this.pauseStatus.textContent = `${state.paused ? 'Tool admissions paused' : 'Tool admissions available'} — typing: ${state.typing_clients}, speaking: ${state.speaking_clients}. Already admitted work continues.`; }
  reply(m) {
    if (m.type === 'completion') this.diagnostic(`Request ${m.request_id}: ${m.outcome}${m.error ? ' — ' + m.error.message : ''}`);
    if (m.type === 'ack' && Object.hasOwn(m, 'paused')) this.pause(m);
    if (m.type === 'ack' && Object.hasOwn(m, 'sent')) this.diagnostic(`Hint received for ${m.request_id}; sent=${m.sent}`);
    if (m.type === 'ack' && Object.hasOwn(m, 'accepted')) this.diagnostic(m.accepted ? `Interrupt accepted for ${m.request_id}` : 'No active request to interrupt');
  }
  reconcile() {
    if (!this.connected) return Promise.resolve();
    const causes = {typing: this.input.value.length > 0, speaking: this.speech.busy()}, key = JSON.stringify(causes);
    if (key === this.lastCauses) return this.pausePromise || Promise.resolve();
    this.lastCauses = key; this.pausePromise = this.connector.send('pause', causes);
    this.pausePromise.catch(() => { if (this.lastCauses === key) this.lastCauses = null; }); return this.pausePromise;
  }
  async submit(type) {
    const text = this.input.value; if (!text.trim()) { this.diagnostic('Enter text first'); return; }
    try {
      const pending = this.connector.send(type, {text}); this.input.value = '';
      // Own both failures immediately. Acceptance uncertainty takes precedence
      // over a simultaneous failure to clear this connection's pause causes.
      const [submission, pause] = await Promise.allSettled([pending, this.reconcile()]);
      if (submission.status === 'rejected') throw submission.reason;
      const reply = submission.value;
      let message = type === 'prompt' ? `Accepted ${reply.request_id}` : `Hint received for ${reply.request_id}; sent=${reply.sent}`;
      if (pause.status === 'rejected') message += '; pause update unavailable after disconnect';
      this.diagnostic(message);
    } catch (e) { this.diagnostic(e.message); }
    this.input.focus();
  }
  speak(key, text) { if (!this.connected) { this.diagnostic('Connect before speaking'); return; } this.speech.enqueue('manual/' + key, text); }
  close() { this.connector.close(); this.speech.cancel(); }
}
