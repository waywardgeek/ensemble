import {Connector} from './connector.js';
import {ArtifactScroll} from './artifacts.js';
import {SpeechQueue} from './speech.js';
import {SettingsPanel} from './preferences.js';

// Public page controller: an embedding may provide a different layout using the
// same labeled elements, or own Connector and ArtifactScroll directly.
export class Page {
  constructor(parent, root, url) {
    this.parent = parent; this.listeners = []; this.closed = false;
    this.root = root; this.input = root.querySelector('[data-input]'); this.status = root.querySelector('[data-status]');
    this.notice = root.querySelector('[data-notice]'); this.pauseStatus = root.querySelector('[data-pause]');
    const actions = root.querySelector('[data-actions]');
    this.artifacts = new ArtifactScroll(this, root.querySelector('[data-artifacts]'), actions ? 'chat' : 'all');
    this.actions = actions ? new ArtifactScroll(this, actions, 'actions') : null;
    this.speech = new SpeechQueue(this); this.connector = new Connector(this, url); this.connected = false; this.settings = new SettingsPanel(this);
    this.listen(this.input, 'input', () => this.reconcile().catch(e => this.diagnostic(e.message)));
    this.listen(this.input, 'keydown', e => { if (e.key === 'Escape' && this.input.value === '') { e.preventDefault(); this.speech.cancel(); } });
    this.listen(root.querySelector('[data-prompt]'), 'click', () => this.submit('prompt'));
    this.listen(root.querySelector('[data-hint]'), 'click', () => this.submit('hint'));
    this.listen(root.querySelector('[data-interrupt]'), 'click', () => this.connector.send('interrupt').catch(e => this.diagnostic(e.message)));
    this.listen(root.querySelector('[data-latest]'), 'click', () => { this.artifacts.latest(); this.actions?.latest(); });
    this.listen(root.querySelector('[data-cancel-speech]'), 'click', () => this.speech.cancel());
    const auto = root.querySelector('[data-auto-speech]');
    if (auto) this.listen(auto, 'click', () => { this.speech.activate(); this.settings.change({autoplay: !this.speech.enabled}, auto); });
    const playback = root.querySelector('[data-enable-playback]'); if (playback) this.listen(playback, 'click', () => this.speech.activate());
    this.connector.connect();
  }
  application() { return this.parent; }
  listen(element, type, callback) { element.addEventListener(type, callback); this.listeners.push(() => element.removeEventListener(type, callback)); }
  diagnostic(message) { if (!this.closed) this.notice.textContent = message; }
  speechEvent(detail) { if (!this.closed) this.root.dispatchEvent(new CustomEvent('ensemble-speech', {detail})); }
  connection(status) {
    if (this.closed) return;
    this.status.textContent = status;
    this.connected = status.startsWith('Connected'); this.lastCauses = null;
    if (!this.connected) { this.speech.cancel(); this.artifacts.incomplete(); this.actions?.incomplete(); this.pauseStatus.textContent = 'No pause held by this page'; }
    else this.reconcile().catch(e => this.diagnostic(e.message));
  }
  preferences(snapshot) { if (this.closed) return; this.settings.apply(snapshot); const auto=this.root.querySelector('[data-auto-speech]'); if(auto){auto.setAttribute('aria-pressed',String(snapshot.preferences.autoplay));auto.textContent='Autoplay new answers: '+(snapshot.preferences.autoplay?'on':'off');} }
  snapshot(snapshot) { if (this.closed) return; this.speech.reset(); this.speech.seed(snapshot.partials); this.artifacts.reset(snapshot); this.actions?.reset(snapshot); this.pause(snapshot.state); this.skills(snapshot.state.skills); this.settings.policy(snapshot.state.execution_policy, snapshot.state.active_max_model_requests); this.lifecycle(snapshot.state.lifecycle, snapshot.agent_id); }
  lifecycle(state, agent) { const el=this.root.querySelector('[data-agent]'); if(el){if(agent)this.agentID=agent;el.textContent=`${this.agentID || ''} — ${state}`;} }
  observation(o) { if (this.closed) return; this.artifacts.observation(o); this.actions?.observation(o); this.speech.observe(o); if(o.skills)this.skills(o.skills); if (o.kind === 'pause_changed') this.pause(o); if(o.kind==='policy_changed')this.settings.policy(o.execution_policy); if(o.kind==='state')this.lifecycle(o.state,o.agent_id); if(o.event?.type==='turn_started')this.settings.policy(this.settings.executionPolicy,o.event.turn.policy?.effective_max_model_requests ?? 16); if(o.event?.type==='turn_ended')this.settings.policy(this.settings.executionPolicy,null); }
  skills(state) {
    const element = this.root.querySelector('[data-skills]'); if (!element) return;
    element.replaceChildren();
    if (!state) { element.textContent = 'Skills disabled.'; return; }
    const line = (label, text) => { const p = document.createElement('p'); p.textContent = label + ': ' + text; element.append(p); };
    line('Primary', state.primary); line('Revision', String(state.revision)); line('Roots', state.roots.join(', ') || '(none)');
    line('Active', state.active.map(item => `${item.name} (${item.type})`).join(', '));
    line('Available', state.available.map(item => `${item.name}: ${item.description}`).join('; ') || '(none)');
    line('Tools', state.tools.join(', '));
  }
  pause(state) { this.pauseStatus.textContent = `${state.paused ? 'Tool admissions paused' : 'Tool admissions available'} — typing: ${state.typing_clients}, speaking: ${state.speaking_clients}. Already admitted work continues.`; }
  reply(m) {
    if (this.closed) return;
    if (m.type === 'completion') this.diagnostic(`Request ${m.request_id}: ${m.outcome}${m.error ? ' — ' + m.error.message : ''}`);
    if (m.type === 'ack' && Object.hasOwn(m, 'paused')) this.pause(m);
    if (m.type === 'ack' && Object.hasOwn(m, 'sent')) this.diagnostic(`Hint received for ${m.request_id}; sent=${m.sent}`);
    if (m.type === 'ack' && Object.hasOwn(m, 'accepted')) this.diagnostic(m.accepted ? `Interrupt accepted for ${m.request_id}` : 'No active request to interrupt');
  }
  reconcile() {
    if (this.closed || !this.connected) return Promise.resolve();
    const causes = {typing: this.input.value.length > 0, speaking: this.speech.busy()}, key = JSON.stringify(causes);
    if (key === this.lastCauses) return this.pausePromise || Promise.resolve();
    this.lastCauses = key; this.pausePromise = this.connector.send('pause', causes);
    this.pausePromise.catch(() => { if (this.lastCauses === key) this.lastCauses = null; }); return this.pausePromise;
  }
  async submit(type) {
    if (this.closed) return;
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
    if (!this.closed) this.input.focus();
  }
  speak(key, text) { if (!this.connected) { this.diagnostic('Connect before speaking'); return; } this.speech.activate(); this.speech.enqueue('manual/' + key, text); }
  close() {
    if (this.closed) return;
    this.closed = true; this.connected = false;
    for (const remove of this.listeners) remove();
    this.listeners = [];
    this.connector.close(); this.speech.close(); this.artifacts.close(); this.actions?.close(); this.settings.close(); this.parent.release(this);
  }
}
