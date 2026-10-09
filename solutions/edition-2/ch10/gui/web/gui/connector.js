// The owner provides snapshot, observation, reply, connection and diagnostic methods.
// Connector owns the socket and generation; it never owns conversation state.
export class Connector {
  constructor(owner, url = new URL('/ws', location.href).href.replace(/^http/, 'ws')) {
    this.owner = owner; this.url = url; this.socket = null; this.staging = null;
    this.generation = null; this.revision = 0; this.pending = new Map(); this.closed = false;
  }
  connect() {
    if (this.closed) return;
    const socket = new WebSocket(this.url); this.socket = socket; this.count = 0; this.preferences = null; this.settingsReady = false;
    this.owner.connection('Connecting');
    socket.onopen = () => { if (socket === this.socket) this.send('subscribe').catch(e => this.owner.diagnostic(e.message)); };
    socket.onmessage = event => {
      if (socket !== this.socket) return;
      try { this.receive(this.decode(event.data)); } catch (e) { this.settingsReady = false; this.owner.diagnostic('Invalid server frame: ' + e.message); socket.close(); }
    };
    socket.onclose = () => {
      if (socket !== this.socket) return;
      this.staging = null; this.generation = null; this.settingsReady = false;
      this.rejectPending(); this.owner.connection('Disconnected — this page no longer holds a pause');
      if (!this.closed) this.timer = setTimeout(() => this.connect(), 500);
    };
    socket.onerror = () => { if (socket === this.socket) this.owner.diagnostic('Connection failed'); };
  }
  // Disk and wire revisions remain uint64 JSON integers. Number is used only
  // while exact; unsafe revisions stay BigInt through decode, comparison and send.
  decode(raw) {
    // Retain lexemes by their containing object, then normalize only protocol
    // counters. A tool argument named revision must not become an unexpected
    // BigInt and break artifact rendering; strings are never rewritten.
    const sources = new WeakMap();
    const message = JSON.parse(raw, function(key, value, context) {
      if (typeof value === 'number' && !Number.isSafeInteger(value)) {
        let fields = sources.get(this); if (!fields) sources.set(this, fields = new Map());
        fields.set(key, context?.source);
      }
      return value;
    });
    const counter = (object, key = 'revision') => {
      if (!object || object[key] === undefined || object[key] === null) return;
      const value = object[key];
      if (typeof value !== 'number' || !Number.isInteger(value) || value < 0) throw new Error('Invalid server counter');
      if (Number.isSafeInteger(value)) return;
      const source = sources.get(object)?.get(key);
      if (!source) throw new Error('This browser cannot retain exact server revisions; JSON source support is required');
      if (!/^(0|[1-9][0-9]{0,19})$/.test(source) || BigInt(source) > 18446744073709551615n) throw new Error('Invalid server counter');
      object[key] = BigInt(source);
    };
    counter(message); counter(message, 'watermark'); counter(message, 'watch_revision');
    for (const key of ['first_seq','last_seq','log_seq','as_of']) counter(message,key);
    counter(message.state?.session,'checkpoint_seq'); counter(message.observation?.session,'checkpoint_seq');
    for (const job of message.state?.job_access || []) counter(job,'handle');
    const eventCounters = e => {
      if (!e) return;
      counter(e,'seq'); counter(e.job,'handle'); counter(e.tool?.job,'handle');
      counter(e.turn,'request_index');
      for (const key of ['from','to']) counter(e.redact,key);
      for (const key of ['hints','ephemera']) for (let i=0;i<(e.request?.[key]?.length||0);i++) counter(e.request[key],String(i));
    };
    eventCounters(message.event); eventCounters(message.observation?.event);
    counter(message.observation,'response_seq'); counter(message.observation,'seq');
    counter(message.state?.execution_policy); counter(message.observation?.execution_policy);
    const skillState = state => {
      if (!state) return;
      counter(state);
      for (const item of [...state.active, ...state.retired]) counter(item, 'activation');
    };
    const skillEvent = event => {
      if (!event?.skills) return;
      skillState(event.skills.state);
      for (const item of event.skills.activated) {
        counter(item, 'activation');
        for (let i = 0; i < item.dependencies.length; i++) counter(item.dependencies, String(i));
      }
    };
    skillState(message.state?.skills); skillState(message.observation?.skills);
    skillEvent(message.event); skillEvent(message.observation?.event);
    if (message.type === 'error' && ['preferences', 'policy'].includes(message.domain)) counter(message.current);
    return message;
  }
  encode(message) {
    if (!['preferences_update', 'policy_update'].includes(message.type)) return JSON.stringify(message);
    const base = message.base_revision;
    if (typeof base === 'number' && !Number.isSafeInteger(base)) throw new Error('Settings base revision must be exact');
    if (typeof base !== 'bigint') return JSON.stringify(message);
    if (base < 0n || base > 18446744073709551615n) throw new Error('Invalid settings base revision');
    // Serialize the other fields normally, then append only this known numeric
    // field. No replacement can accidentally alter user text or patch contents.
    return JSON.stringify({...message, base_revision: undefined}).slice(0, -1) + ',"base_revision":' + base.toString() + '}';
  }
  rejectPending() {
    for (const p of this.pending.values()) {
      const message = p.type === 'prompt' ? 'Prompt acceptance unknown; inspect history after reconnect. Do not resend automatically.' : 'Connection lost';
      p.reject(new Error(message));
    }
    this.pending.clear();
  }
  close() {
    if (this.closed) return;
    this.closed = true; clearTimeout(this.timer);
    const socket = this.socket; this.socket = null; this.staging = null; this.generation = null;
    this.rejectPending();
    if (socket) { socket.onopen = socket.onmessage = socket.onclose = socket.onerror = null; socket.close(); }
  }
  send(type, fields = {}) {
    if (this.closed || !this.socket || this.socket.readyState !== WebSocket.OPEN) return Promise.reject(new Error('Not connected'));
    const id = 'c' + (++this.count);
    return new Promise((resolve, reject) => {
      this.pending.set(id, {type, resolve, reject});
      try { this.socket.send(this.encode({type, id, ...fields})); } catch (error) { this.pending.delete(id); reject(error); }
    });
  }
  receive(m) {
    if (m.type === "preferences_snapshot" || m.type === "preferences_changed") {
      if (m.type === "preferences_changed" && (!this.preferences || BigInt(m.revision) !== BigInt(this.preferences.revision) + 1n)) throw new Error("preferences revision gap");
      this.preferences = {revision: m.revision, preferences: m.preferences}; this.owner.preferences?.(this.preferences); return;
    }
    if (m.type === 'snapshot_begin') {
      this.staging = {...m, events: [], partials: []}; return;
    }
    if (m.type.startsWith('snapshot_')) {
      if (!this.staging || m.generation !== this.staging.generation) return;
      if (m.type === 'snapshot_event') this.staging.events.push(m.event);
      if (m.type === 'snapshot_partial') this.staging.partials.push(m);
      if (m.type === 'snapshot_end') {
        if (m.watermark !== this.staging.watermark) throw new Error('snapshot watermark mismatch');
        const s = this.staging; this.staging = null; this.generation = s.generation; this.revision = s.watermark;
        if (!this.preferences || !s.state.execution_policy) throw new Error('settings snapshot missing');
        this.settingsReady = true; this.owner.snapshot(s); this.owner.connection('Connected — ' + this.generation);
        const pending = this.pending.get(s.id); this.pending.delete(s.id); pending?.resolve(s);
      }
      return;
    }
    if (m.type === 'observation') {
      if (m.generation !== this.generation) return;
      if (BigInt(m.revision) !== BigInt(this.revision) + 1n) throw new Error('observation gap');
      this.revision = m.revision; this.owner.observation(m.observation); return;
    }
    const pending = this.pending.get(m.id);
    if (pending) { this.pending.delete(m.id); ['error','command_error'].includes(m.type) ? pending.reject(Object.assign(new Error(m.message), {code:m.code, domain:m.domain, current:m.current})) : pending.resolve(m); }
    this.owner.reply(m);
  }
}
