export const partKey = p => [p.agent_id, p.request_id, p.operation_id, p.part_id].join('/');
const clean = text => String(text ?? '').replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '').replace(/\x1b\][^\x07]*(?:\x07|\x1b\\)/g, '').replace(/[\x00-\x08\x0b-\x1f\x7f]/g, '');
function inline(node, text) {
  // Raw HTML and links remain text. Only paired emphasis is interpreted.
  const pattern = /(\*\*([^*]+)\*\*|\*([^*]+)\*)/g; let start = 0;
  for (const match of text.matchAll(pattern)) {
    node.append(document.createTextNode(text.slice(start, match.index)));
    const mark = document.createElement(match[2] ? 'strong' : 'em'); mark.textContent = match[2] || match[3]; node.append(mark); start = match.index + match[0].length;
  }
  node.append(document.createTextNode(text.slice(start)));
}
export function safeMarkdown(owner, text) {
  const fragment = document.createDocumentFragment(); let code = null, list = null, paragraph = [];
  const flush = () => { if (paragraph.length) { const p = document.createElement('p'); inline(p, paragraph.join('\n')); fragment.append(p); paragraph = []; } };
  for (const line of clean(text).split('\n')) {
    if (line.startsWith('```')) { flush(); list = null; if (code) code = null; else { const pre = document.createElement('pre'); code = document.createElement('code'); pre.append(code); fragment.append(pre); } continue; }
    if (code) { code.append(document.createTextNode(line + '\n')); continue; }
    const item = line.match(/^\s*(?:[-*]|\d+\.)\s+(.*)$/);
    if (item) { flush(); if (!list) { list = document.createElement('ul'); fragment.append(list); } const li = document.createElement('li'); inline(li, item[1]); list.append(li); }
    else { list = null; if (!line.trim()) flush(); else paragraph.push(line); }
  }
  flush(); return fragment;
}
export class Artifact {
  constructor(owner, key, title, text, markdown = false) {
    this.owner = owner; this.key = key; this.title = title; this.text = text; this.markdown = markdown; this.expanded = false;
    this.element = document.createElement('article'); this.element.className = 'artifact'; this.element.dataset.key = key;
    this.heading = document.createElement('h2'); this.content = document.createElement('div'); this.meta = document.createElement('p'); this.meta.className = 'meta';
    this.expand = document.createElement('button'); this.expand.addEventListener('click', () => { this.expanded = !this.expanded; this.render(); });
    this.speaker = document.createElement('button'); this.speaker.textContent = 'Speak'; this.speaker.setAttribute('aria-label', 'Speak full card'); this.speaker.addEventListener('click', () => this.owner.speak(this.key, this.text));
    this.element.append(this.heading, this.content, this.meta, this.expand, this.speaker); this.render();
  }
  render() {
    this.heading.textContent = clean(this.title); const limit = 1200; const short = !this.expanded && this.text.length > limit;
    const shown = clean(short ? this.text.slice(0, limit) : this.text); this.content.replaceChildren();
    if (this.markdown) this.content.append(safeMarkdown(this.owner, shown)); else { const pre = document.createElement('pre'); pre.textContent = shown; this.content.append(pre); }
    this.meta.textContent = [this.status || '', short ? `${this.text.length - limit} characters omitted from preview` : '', this.reference || ''].filter(Boolean).join(' · ');
    this.expand.hidden = this.text.length <= limit; this.expand.textContent = this.expanded ? 'Show preview' : 'Expand full retained text'; this.expand.setAttribute('aria-expanded', String(this.expanded));
  }
}
// ArtifactScroll accepts projected records; its owner supplies speech/control.
export class ArtifactScroll {
  constructor(owner, element) {
    this.owner = owner; this.element = element; this.cards = new Map(); this.calls = new Map(); this.jobs = new Map(); this.follow = true;
    element.addEventListener('scroll', () => { this.follow = element.scrollHeight - element.scrollTop - element.clientHeight < 80; });
  }
  speak(key, text) { this.owner.speak(key, text); }
  diagnostic(message) { this.owner.diagnostic(message); }
  latest() { this.follow = true; this.element.scrollTop = this.element.scrollHeight; }
  changed() { if (this.follow) this.element.scrollTop = this.element.scrollHeight; }
  card(key, title, text, markdown = false) {
    let card = this.cards.get(key);
    if (!card) { card = new Artifact(this, key, title, text, markdown); this.cards.set(key, card); this.element.append(card.element); }
    else { card.title = title; card.text = text; card.markdown = markdown; card.render(); }
    this.changed(); return card;
  }
  reset(snapshot) {
    this.cards.clear(); this.calls.clear(); this.jobs.clear(); this.element.replaceChildren(); this.agent = snapshot.agent_id;
    if (snapshot.omitted) this.card('omitted', 'Earlier history omitted', `${snapshot.omitted} earlier renderable events omitted; this view contains the most recent 100 events.`);
    for (const e of snapshot.events) this.event(e, true, snapshot.agent_id);
    for (const p of snapshot.partials) this.partial(p, p.channels);
    this.changed();
  }
  partial(p, channels) {
    const key = 'p/' + partKey(p); const card = this.cards.get(key);
    const values = {...card?.channels, ...channels}; const tool = Object.hasOwn(values, 'tool_name') || Object.hasOwn(values, 'tool_args');
    const text = tool ? `${values.tool_name || ''}\n${values.tool_args || ''}` : [values.thinking, values.text].filter(v => v !== undefined).join('\n');
    const updated = this.card(key, tool ? 'Proposed tool' : Object.hasOwn(values, 'thinking') ? 'Thinking / answer' : 'Answer', text, !tool);
    updated.channels = values; updated.status = 'Provisional'; updated.identity = p; updated.render(); return updated;
  }
  final(o) {
    const provisional = 'p/' + partKey(o), key = `d/${o.agent_id}/${o.response_seq}/${o.part_index}`; const card = this.cards.get(provisional);
    if (card) { this.cards.delete(provisional); card.key = key; card.element.dataset.key = key; this.cards.set(key, card); }
    this.part(key, o.part, o.agent_id);
  }
  part(key, part, agent) {
    let title = 'Unknown content', text = JSON.stringify(part), markdown = false;
    if (part.type === 'text' || part.type === 'thinking') { title = part.type === 'thinking' ? 'Thinking' : 'Answer'; text = part.text ?? ''; markdown = true; }
    if (part.type === 'opaque') { title = 'Opaque provider content'; text = '[Opaque provider content — unavailable for display]'; }
    if (part.type === 'tool_call') { title = 'Proposed tool: ' + part.name; text = `${part.name}\n${JSON.stringify(part.args, null, 2)}`; this.calls.set(`${agent}/${part.call_id}`, key); }
    const card = this.card(key, title, text, markdown); card.status = part.type === 'tool_call' ? 'Accepted proposal — not yet started' : 'Accepted'; card.render(); return card;
  }
  event(e, replay = false, agent = this.agent) {
    const key = `e/${agent}/${e.seq}`;
    if (e.type === 'message_received') this.card(key, e.message.actor === 'human' ? 'You' : 'Message', e.message.parts.map(p => p.text ?? `[${p.type}]`).join('\n'));
    if (e.type === 'hint_received') this.card(key, 'Hint received', e.hint.text);
    if (e.type === 'response_ended' && replay) e.response.parts.forEach((p, i) => this.part(`d/${agent}/${e.seq}/${i}`, p, agent));
    if (e.type === 'tool_called' || e.type === 'tool_returned') {
      const tool = e.tool, callKey = `${agent}/${tool.call_id}`; let existing = this.calls.get(callKey); let card;
      if (e.type === 'tool_called') {
        existing ||= key; card = this.card(existing, 'Tool: ' + tool.name, `${tool.name}\n${JSON.stringify(tool.args, null, 2)}`); this.calls.set(callKey, existing); card.status = 'Started';
      } else {
        const prior = existing ? this.cards.get(existing) : null;
        const report = tool.parts.map(p => p.text ?? (p.ref ? `[artifact reference: ${p.ref.locator}]` : `[${p.type}]`)).join('\n');
        card = this.card(key, prior ? 'Tool result: ' + tool.call_id : 'Earlier-call result: ' + tool.call_id, report);
        card.status = [prior ? '' : 'Missing call context outside retained window', tool.is_error ? 'Error report' : 'Report delivered', tool.job?.status || ''].filter(Boolean).join(' · ');
        if (prior) { prior.status = 'Report delivered' + (tool.job ? ' — ' + tool.job.status : ''); prior.render(); }
        if (tool.parts.some(p => p.ref)) card.reference = 'This is the full retained report. The larger artifact is only a reference; ask the Agent to retrieve it.';
      }
      if (tool.job) { const jobKey = `${agent}/${tool.job.handle}`; const cards = this.jobs.get(jobKey) || []; cards.push(card); this.jobs.set(jobKey, cards); }
      card.render();
    }
    if (e.type === 'job_ended' || e.type === 'job_killed') for (const card of this.jobs.get(`${agent}/${e.job.handle}`) || []) { card.status = 'Job ' + e.job.status; card.render(); }
    if (e.type === 'turn_ended') this.card(key, 'Request ' + e.turn.request_id, 'Outcome: ' + e.turn.outcome);
    if (e.type === 'error_occurred') this.card(key, 'Error', e.error.message);
  }
  observation(o) {
    if (o.event?.type) this.event(o.event, false, o.agent_id);
    if (o.kind === 'part_delta') { const key = 'p/' + partKey(o), previous = this.cards.get(key)?.channels?.[o.channel] || ''; this.partial(o, {[o.channel]: previous + o.text}); }
    if (o.kind === 'part_final') this.final(o);
    if (o.kind === 'model_end' && !o.accepted) this.incomplete(o);
    this.changed();
  }
  incomplete(operation = null) {
    for (const card of this.cards.values()) if (card.status === 'Provisional' && (!operation || card.identity?.operation_id === operation.operation_id && card.identity?.agent_id === operation.agent_id && card.identity?.request_id === operation.request_id)) { card.status = 'Incomplete — not accepted'; card.render(); }
  }
}
