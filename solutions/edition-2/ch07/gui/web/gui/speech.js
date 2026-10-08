import {partKey} from './artifacts.js';
// The page owns one queue. Every callback is fenced by its queue generation.
export class SpeechQueue {
  constructor(owner) {
    this.owner = owner;
    this.queue = []; this.generation = 0; this.current = null; this.cursors = new Map(); this.enabled = false;
  }
  busy() { return this.queue.length > 0 || this.current !== null; }
  service() { return this.owner.application().speech(); }
  enable(value) { this.enabled = value; if (value && !this.service().available()) { this.owner.diagnostic('Speech synthesis unavailable'); this.enabled = false; } }
  enqueue(key, text, operation = null) {
    if (this.closed || !text.trim()) return;
    if (!this.service().available()) { this.owner.diagnostic('Speech synthesis unavailable'); return; }
    this.queue.push({key, text, operation}); this.owner.reconcile().catch(e => this.owner.diagnostic(e.message)); this.pump();
  }
  async pump() {
    if (this.current || !this.queue.length) return;
    const item = this.queue.shift(), generation = this.generation; this.current = item;
    try { await this.owner.reconcile(); } catch (e) { if (generation === this.generation) this.cancel(); this.owner.diagnostic(e.message); return; }
    if (generation !== this.generation || this.current !== item) return;
    let settled = false;
    const settle = error => {
      if (settled || generation !== this.generation || this.current !== item) return;
      settled = true; this.current = null;
      if (error) this.owner.diagnostic('Speech error: ' + error);
      this.owner.speechEvent({type: error ? 'error' : 'end', key: item.key, error});
      if (this.queue.length) this.pump(); else this.owner.reconcile().catch(e => this.owner.diagnostic(e.message));
    };
    this.service().submit(this.owner, item.text, {
      start: () => { if (generation === this.generation) this.owner.speechEvent({type: 'start', key: item.key, text: item.text}); },
      end: settle,
    });
  }
  cancel(operation = null) {
    const retained = operation ? this.queue.filter(item => item.operation !== operation) : [];
    if (operation && this.current && this.current.operation !== operation) {
      this.queue = retained; this.owner.reconcile().catch(e => this.owner.diagnostic(e.message)); return;
    }
    this.generation++; this.current = null; this.queue = retained; this.service().cancel(this.owner);
    this.owner.reconcile().catch(e => this.owner.diagnostic(e.message)); this.pump();
  }
  reset() { this.cancel(); this.cursors.clear(); }
  close() { if (this.closed) return; this.closed = true; this.enabled = false; this.cancel(); this.cursors.clear(); }
  seed(partials) {
    for (const p of partials) {
      const text = [p.channels.thinking, p.channels.text].filter(v => v !== undefined).join('');
      this.cursors.set(partKey(p), {text, sent: text.length, tool: false});
    }
  }
  observe(o) {
    const operation = [o.agent_id, o.request_id, o.operation_id].join('/');
    if (o.kind === 'model_end') { if (!o.accepted) this.cancel(operation); return; }
    if (!this.enabled || !['part_delta', 'part_final'].includes(o.kind)) return;
    const key = partKey(o);
    let cursor = this.cursors.get(key);
    if (!cursor) { cursor = {text: '', sent: 0, tool: false}; this.cursors.set(key, cursor); }
    if (o.kind === 'part_delta' && (o.channel === 'text' || o.channel === 'thinking')) cursor.text += o.text;
    if (o.kind === 'part_final') {
      if (o.part.type === 'text' || o.part.type === 'thinking') cursor.text = o.part.text ?? '';
      if (o.part.type === 'tool_call' && !cursor.tool) {
        cursor.tool = true; const args = o.part.args || {}; const path = args.path || args.file_path || args.directory || '';
        this.enqueue(key, `Tool ${o.part.name}${path ? ': ' + path : ''}`, operation); return;
      }
    }
    if (o.kind !== 'part_delta' && o.kind !== 'part_final') return;
    const pending = cursor.text.slice(cursor.sent);
    const boundary = o.kind === 'part_final' ? pending.length : Math.max(...Array.from(pending.matchAll(/[.!?](?:\s|$)/g), m => m.index + m[0].length), 0);
    if (boundary > 0) { this.enqueue(key, pending.slice(0, boundary), operation); cursor.sent += boundary; }
  }
}
