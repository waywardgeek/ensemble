// The owner provides snapshot, observation, reply, connection and diagnostic methods.
// Connector owns the socket and generation; it never owns conversation state.
export class Connector {
  constructor(owner, url = new URL('/ws', location.href).href.replace(/^http/, 'ws')) {
    this.owner = owner; this.url = url; this.socket = null; this.staging = null;
    this.generation = null; this.revision = 0; this.pending = new Map(); this.closed = false;
  }
  connect() {
    if (this.closed) return;
    const socket = new WebSocket(this.url); this.socket = socket; this.count = 0;
    this.owner.connection('Connecting');
    socket.onopen = () => { if (socket === this.socket) this.send('subscribe').catch(e => this.owner.diagnostic(e.message)); };
    socket.onmessage = event => {
      if (socket !== this.socket) return;
      try { this.receive(JSON.parse(event.data)); } catch (e) { this.owner.diagnostic('Invalid server frame: ' + e.message); socket.close(); }
    };
    socket.onclose = () => {
      if (socket !== this.socket) return;
      this.staging = null; this.generation = null;
      this.rejectPending(); this.owner.connection('Disconnected — this page no longer holds a pause');
      if (!this.closed) this.timer = setTimeout(() => this.connect(), 500);
    };
    socket.onerror = () => { if (socket === this.socket) this.owner.diagnostic('Connection failed'); };
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
      try { this.socket.send(JSON.stringify({type, id, ...fields})); } catch (error) { this.pending.delete(id); reject(error); }
    });
  }
  receive(m) {
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
        this.owner.snapshot(s); this.owner.connection('Connected — ' + this.generation);
        const pending = this.pending.get(s.id); this.pending.delete(s.id); pending?.resolve(s);
      }
      return;
    }
    if (m.type === 'observation') {
      if (m.generation !== this.generation) return;
      if (m.revision !== this.revision + 1) throw new Error('observation gap');
      this.revision = m.revision; this.owner.observation(m.observation); return;
    }
    const pending = this.pending.get(m.id);
    if (pending) { this.pending.delete(m.id); m.type === 'error' ? pending.reject(new Error(m.message)) : pending.resolve(m); }
    this.owner.reply(m);
  }
}
