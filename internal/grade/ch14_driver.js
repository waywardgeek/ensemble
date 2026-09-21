'use strict';
//
// Chapter 14 grader driver.
//
// The speech pipeline is a pure function from a sequence of text fragments to a
// sequence of utterances, with exactly one browser dependency. So the grader loads
// the student's own files under node, stubs that dependency, feeds a fixed fragment
// sequence, and asserts on what reached the channel. No browser, no API key, no
// model, no flake.
//
// The observation point is the speechSynthesis stub itself. That is deliberate: it
// is literally the channel under test, and it means the grader never has to know
// the name of any field inside the student's module.
//
// Usage: node ch14_driver.js '<spec json>'
//   spec = {ttsFile, ttsGlobal, artifactFile, artifactGlobal, guiFile}
//
// Output: one JSON object on stdout, {checks: {id: {ok, err}}, error?: string}.

const fs = require('fs');

const spec = JSON.parse(process.argv[2]);

// ── Environment ──────────────────────────────────────────────────────────────

function makeElement(tag) {
  const el = {
    tagName: String(tag || 'div').toUpperCase(),
    id: '',
    className: '',
    value: '',
    textContent: '',
    innerHTML: '',
    title: '',
    hidden: false,
    disabled: false,
    checked: false,
    scrollTop: 0,
    scrollHeight: 0,
    clientHeight: 0,
    offsetHeight: 0,
    style: {},
    dataset: {},
    children: [],
    childNodes: [],
    parentNode: null,
    _listeners: Object.create(null),
    _attrs: Object.create(null),
  };
  el.classList = {
    _s: new Set(),
    add(...c) { c.forEach((x) => this._s.add(x)); },
    remove(...c) { c.forEach((x) => this._s.delete(x)); },
    contains(c) { return this._s.has(c); },
    toggle(c, f) {
      if (f === undefined) { this._s.has(c) ? this._s.delete(c) : this._s.add(c); return this._s.has(c); }
      if (f) this._s.add(c); else this._s.delete(c);
      return !!f;
    },
  };
  el.appendChild = function (c) { if (c) { c.parentNode = el; } el.children.push(c); el.childNodes.push(c); return c; };
  el.insertBefore = function (c) { return el.appendChild(c); };
  el.removeChild = function (c) {
    const i = el.children.indexOf(c);
    if (i >= 0) { el.children.splice(i, 1); el.childNodes.splice(i, 1); }
    return c;
  };
  el.remove = function () { if (el.parentNode) el.parentNode.removeChild(el); };
  el.addEventListener = function (t, h) { (el._listeners[t] = el._listeners[t] || []).push(h); };
  el.removeEventListener = function (t, h) {
    const l = el._listeners[t]; if (!l) return;
    const i = l.indexOf(h); if (i >= 0) l.splice(i, 1);
  };
  el.dispatchEvent = function (ev) { fire(el, ev && ev.type, ev); return true; };
  el.setAttribute = function (k, v) { el._attrs[k] = String(v); };
  el.getAttribute = function (k) { return k in el._attrs ? el._attrs[k] : null; };
  el.removeAttribute = function (k) { delete el._attrs[k]; };
  el.hasAttribute = function (k) { return k in el._attrs; };
  el.querySelector = function () { return null; };
  el.querySelectorAll = function () { return []; };
  el.closest = function () { return null; };
  el.focus = function () {};
  el.blur = function () {};
  el.select = function () {};
  el.scrollIntoView = function () {};
  el.getBoundingClientRect = function () { return { top: 0, left: 0, right: 0, bottom: 0, width: 0, height: 0 }; };
  Object.defineProperty(el, 'firstChild', { get() { return el.childNodes[0] || null; } });
  Object.defineProperty(el, 'lastChild', { get() { return el.childNodes[el.childNodes.length - 1] || null; } });
  return el;
}

// Invoke every listener registered for `type` on `el`.
function fire(el, type, ev) {
  if (!el || !type) return;
  const ls = (el._listeners && el._listeners[type]) || [];
  for (const h of ls.slice()) {
    try { h.call(el, ev || { type, target: el, preventDefault() {}, stopPropagation() {} }); } catch (e) { /* student handler threw; scenario assertions will show it */ }
  }
}

// Build a fresh global environment. Returns the observation handles.
function makeEnv() {
  const env = {
    spoken: [],      // text handed to speechSynthesis.speak, in order
    pending: [],     // utterances awaiting onend
    cancels: 0,
    wsSent: [],      // raw strings passed to WebSocket.send
    byId: new Map(),
    docListeners: Object.create(null),
  };

  const synth = {
    speaking: false,
    paused: false,
    pending: false,
    speak(u) {
      // The Chrome wake-up trick speaks a zero-volume empty utterance on tab focus.
      // That is not content, so it is not part of the channel.
      const t = u && typeof u.text === 'string' ? u.text : '';
      if (t.trim() !== '' && u.volume !== 0) env.spoken.push(t);
      env.pending.push(u);
    },
    cancel() { env.cancels++; env.pending.length = 0; },
    pause() {}, resume() {},
    getVoices() { return []; },
    addEventListener() {}, removeEventListener() {},
  };

  class SpeechSynthesisUtterance {
    constructor(text) {
      this.text = text; this.rate = 1; this.pitch = 1; this.volume = 1;
      this.voice = null; this.lang = 'en-US';
      this.onend = null; this.onerror = null; this.onstart = null;
    }
    addEventListener(t, h) { this['on' + t] = h; }
  }

  class FakeWebSocket {
    constructor(url) {
      this.url = url; this.readyState = 1;
      this.onopen = null; this.onmessage = null; this.onclose = null; this.onerror = null;
      env.ws = this;
    }
    send(d) { env.wsSent.push(typeof d === 'string' ? d : String(d)); }
    close() { this.readyState = 3; }
    addEventListener(t, h) { this['on' + t] = h; }
    removeEventListener() {}
  }
  FakeWebSocket.CONNECTING = 0; FakeWebSocket.OPEN = 1;
  FakeWebSocket.CLOSING = 2; FakeWebSocket.CLOSED = 3;

  const doc = {
    readyState: 'complete',
    title: '',
    body: makeElement('body'),
    documentElement: makeElement('html'),
    getElementById(id) {
      if (!env.byId.has(id)) { const e = makeElement('div'); e.id = id; env.byId.set(id, e); }
      return env.byId.get(id);
    },
    createElement(t) { return makeElement(t); },
    createTextNode(t) { return { textContent: String(t), nodeType: 3 }; },
    createDocumentFragment() { return makeElement('fragment'); },
    querySelector() { return null; },
    querySelectorAll() { return []; },
    addEventListener(t, h) { (env.docListeners[t] = env.docListeners[t] || []).push(h); },
    removeEventListener() {},
    dispatchEvent() { return true; },
  };

  // Recent node versions define some of these (navigator, fetch, localStorage) as
  // getter-only accessors, so a plain assignment throws. defineProperty always wins.
  const def = (name, value) => Object.defineProperty(global, name, { value, writable: true, configurable: true, enumerable: true });

  def('document', doc);
  def('window', global);
  def('self', global);
  def('speechSynthesis', synth);
  def('SpeechSynthesisUtterance', SpeechSynthesisUtterance);
  def('WebSocket', FakeWebSocket);
  def('location', { protocol: 'http:', host: 'localhost:8084', hostname: 'localhost', port: '8084', href: 'http://localhost:8084/', origin: 'http://localhost:8084' });
  def('navigator', { userAgent: 'node', language: 'en-US', clipboard: { writeText() { return Promise.resolve(); } } });
  def('matchMedia', () => ({ matches: false, media: '', addEventListener() {}, removeEventListener() {}, addListener() {}, removeListener() {} }));
  def('localStorage', {
    _m: new Map(),
    getItem(k) { return this._m.has(k) ? this._m.get(k) : null; },
    setItem(k, v) { this._m.set(k, String(v)); },
    removeItem(k) { this._m.delete(k); },
    clear() { this._m.clear(); },
  });
  def('requestAnimationFrame', (fn) => { fn(0); return 0; });
  def('cancelAnimationFrame', () => {});
  def('getComputedStyle', () => ({ getPropertyValue: () => '' }));
  def('alert', () => {}); def('confirm', () => true); def('prompt', () => null);
  def('fetch', () => Promise.resolve({ ok: true, json: () => Promise.resolve({}), text: () => Promise.resolve('') }));
  // Renderers live in a sibling file and only affect pixels, never speech.
  def('Renderers', new Proxy({}, { get: () => (x) => (x == null ? '' : String(x)), has: () => true }));

  env.synth = synth;
  env.document = doc;
  return env;
}

// Drain queued audio: fire onend until nothing is pending. A student's onend
// handler normally speaks the next entry, so this loop is what advances the queue.
function finishAudio(env) {
  let guard = 0;
  while (env.pending.length && guard++ < 5000) {
    const u = env.pending.shift();
    if (u && typeof u.onend === 'function') u.onend();
  }
}

// Load a source file and hand back the binding it declares.
//
// A top-level `const` inside eval does NOT leak into the surrounding scope, so the
// export is appended to the source. Using direct eval inside a function also keeps
// each load independent, which is what makes a fresh module per scenario possible.
function loadModule(file, globalName) {
  const src = fs.readFileSync(file, 'utf8');
  return eval(src + '\n;(' + globalName + ')');
}

function loadTTS(env) {
  const tts = loadModule(spec.ttsFile, spec.ttsGlobal);
  // Sibling modules reach the speech pipeline as a global, so publish it under the
  // name the student's own source declares.
  Object.defineProperty(global, spec.ttsGlobal, { value: tts, writable: true, configurable: true, enumerable: true });
  // Some implementations defer wiring to DOMContentLoaded.
  fire({ _listeners: env.docListeners }, 'DOMContentLoaded');
  return tts;
}

// ── Assertion helpers ────────────────────────────────────────────────────────

const checks = {};
function record(id, fn) {
  try {
    const err = fn();
    checks[id] = err ? { ok: false, err: String(err) } : { ok: true, err: '' };
  } catch (e) {
    checks[id] = { ok: false, err: 'scenario threw: ' + (e && e.stack ? e.stack.split('\n').slice(0, 3).join(' | ') : e) };
  }
}

const show = (a) => JSON.stringify(a);

// ── Scenarios ────────────────────────────────────────────────────────────────

// Feed fragments through the student's chunk entry point, draining audio after each
// so the queue advances exactly as it would in a browser.
function speakFragments(frags, { flush = true } = {}) {
  const env = makeEnv();
  const tts = loadTTS(env);
  for (const f of frags) { tts.queueChunk(f); finishAudio(env); }
  if (flush) { tts.flush(); finishAudio(env); }
  return { env, tts, spoken: env.spoken };
}

record('tts-buffers-fragments', () => {
  const env = makeEnv();
  const tts = loadTTS(env);

  tts.queueChunk('Hel'); finishAudio(env);
  if (env.spoken.length !== 0) return 'spoke ' + show(env.spoken) + ' after the fragment "Hel"; nothing may be spoken before a phrase boundary arrives';
  tts.queueChunk('lo wor'); finishAudio(env);
  if (env.spoken.length !== 0) return 'spoke ' + show(env.spoken) + ' mid-word; "Hello wor" is not a phrase';

  tts.queueChunk('ld. Bye.'); finishAudio(env);
  if (env.spoken.length === 0) return 'spoke nothing after a sentence-ending period arrived; the completed phrase must be released';
  if (env.spoken[0] !== 'Hello world.') return 'first utterance was ' + show(env.spoken[0]) + ', want "Hello world."; a word split across three fragments is still one word';

  tts.flush(); finishAudio(env);
  const joined = env.spoken.join(' ').replace(/\s+/g, ' ').trim();
  if (joined !== 'Hello world. Bye.') return 'utterances ' + show(env.spoken) + ' do not reassemble to "Hello world. Bye."';
  return '';
});

record('tts-filters-markup', () => {
  const fence = 'Here it is:\n```js\nconst x = 1;\n```\nThat was the code.';

  // Both orderings matter. The single-chunk case is the one that disproves the
  // theory that streaming is what shatters a fence.
  const whole = speakFragments([fence]);
  const split = speakFragments(['Here it is:\n``', '`js\nconst x', ' = 1;\n', '``', '`\nThat was the code.']);

  for (const [label, r] of [['as one chunk', whole], ['split across deltas', split]]) {
    const all = r.spoken.join(' ');
    if (all.includes('`')) return 'a backtick reached the channel with the fence ' + label + ': ' + show(r.spoken);
    if (!/code block/i.test(all)) return 'the fenced block was not named when fed ' + label + '; heard ' + show(r.spoken) + ', want something like "code block"';
    if (all.includes('const x = 1')) return 'the body of the fenced block was read aloud ' + label + '; a code block is named, not read';
  }

  const marks = [
    ['**ready** now.', '*', 'emphasis markers'],
    ['Call `flush` now.', '`', 'inline code markers'],
    ['# A Heading\n\nBody text here.', '#', 'heading markers'],
    ['- first item\n\n- second item\n\n', '-', 'bullet markers'],
  ];
  for (const [input, bad, what] of marks) {
    const r = speakFragments([input]);
    const all = r.spoken.join(' ');
    if (all.includes(bad)) return what + ' reached the channel: input ' + show(input) + ' produced ' + show(r.spoken);
    if (all.trim() === '') return 'nothing at all was spoken for ' + show(input) + '; filtering must remove the marks, not the words';
  }

  const link = speakFragments(['See [the design doc](http://example.com/design.md) for more.']);
  const linkAll = link.spoken.join(' ');
  if (/http|\]|\(/.test(linkAll)) return 'link syntax reached the channel: ' + show(link.spoken) + '; keep the text, drop the URL';
  if (!/the design doc/.test(linkAll)) return 'the link text was lost: ' + show(link.spoken);
  return '';
});

record('tts-boundaries', () => {
  // A single newline is a wrap, not a phrase boundary.
  const wrap = speakFragments(['The quick brown\nfox jumps over it.']);
  if (wrap.spoken.length !== 1) return 'a sentence wrapped across one newline produced ' + wrap.spoken.length + ' utterances ' + show(wrap.spoken) + ', want 1; a lone newline is a space';
  if (!/brown fox/.test(wrap.spoken[0])) return 'the wrap was not joined as a space: ' + show(wrap.spoken[0]);

  // A blank line is a real paragraph break.
  const para = speakFragments(['First paragraph here\n\nSecond paragraph here']);
  if (para.spoken.length !== 2) return 'a blank line produced ' + para.spoken.length + ' utterances ' + show(para.spoken) + ', want 2; a blank line is a boundary';

  // End of stream is itself a boundary.
  const env = makeEnv();
  const tts = loadTTS(env);
  tts.queueChunk('Next up is this:'); finishAudio(env);
  if (env.spoken.length !== 0) return 'text with no terminal punctuation was spoken before flush: ' + show(env.spoken);
  tts.flush(); finishAudio(env);
  if (env.spoken.length === 0) return 'flush() spoke nothing; a response ending in a colon is otherwise never heard';
  if (!/Next up is this/.test(env.spoken.join(' '))) return 'flush() spoke ' + show(env.spoken) + ', want the stranded text "Next up is this:"';
  return '';
});

record('tts-expands-identifiers', () => {
  const cases = [
    ['The queueChunk call.', /queue Chunk/, 'camelCase must split into words'],
    ['The snake_case_name value.', /snake case name/, 'snake_case must split into words'],
    ['An HTTPServer restart.', /HTTP Server/, 'an acronym run must split before the capitalised word'],
    ['Reading RPGLit lately.', /RPG Lit/, 'an acronym run must split before the capitalised word'],
  ];
  for (const [input, want, why] of cases) {
    const r = speakFragments([input]);
    const all = r.spoken.join(' ');
    if (!want.test(all)) return 'input ' + show(input) + ' produced ' + show(r.spoken) + '; ' + why;
  }

  // The mutation guard. An implementation that inserts a space before every capital
  // passes every case above and fails here, spelling ordinary words letter by letter.
  const plain = speakFragments(['The quick brown fox jumps over the lazy dog.']);
  if (plain.spoken.join(' ').replace(/\s+/g, ' ').trim() !== 'The quick brown fox jumps over the lazy dog.') {
    return 'ordinary words were altered: ' + show(plain.spoken) + '; only identifiers are expanded';
  }
  const caps = speakFragments(['ALL CAPS WORDS ARE FINE.']);
  const capsOut = caps.spoken.join(' ').replace(/\s+/g, ' ').trim();
  if (capsOut !== 'ALL CAPS WORDS ARE FINE.') {
    return 'an all-caps word was broken up: ' + show(caps.spoken) + '; ALL CAPS is read as words, not spelled out letter by letter';
  }
  return '';
});

record('tts-speaks-unstreamed', () => {
  if (!spec.artifactFile) return 'could not find the module that feeds the speech channel';

  function run(msgs) {
    const env = makeEnv();
    const tts = loadTTS(env);
    const Scroll = loadModule(spec.artifactFile, spec.artifactGlobal);
    const scroll = new Scroll(makeElement('div'), {});
    for (const m of msgs) { scroll.handleMessage(m); finishAudio(env); }
    return env.spoken.join(' ');
  }

  // Direction 1: a part that arrives whole, with no deltas, is still spoken.
  const whole = run([{ type: 'part_final', part_id: 'p1', kind: 'text', text: 'A whole part arrived.' }]);
  if (!/A whole part arrived/.test(whole)) {
    return 'a part delivered with no deltas was never spoken (heard ' + show(whole) + '); with streaming off the agent is silent while the screen looks healthy';
  }

  // Direction 2: a part that streamed is not spoken a second time when it finalises.
  const streamed = run([
    { type: 'part_delta', part_id: 'p2', kind: 'text', chunk: 'Streamed text ' },
    { type: 'part_delta', part_id: 'p2', kind: 'text', chunk: 'arrived here.' },
    { type: 'part_final', part_id: 'p2', kind: 'text', text: 'Streamed text arrived here.' },
  ]);
  const hits = (streamed.match(/Streamed text arrived here/g) || []).length;
  if (hits === 0) return 'a streamed part was never spoken: ' + show(streamed);
  if (hits > 1) return 'a streamed part was spoken ' + hits + ' times: ' + show(streamed) + '; text already delivered by deltas must not be repeated at part_final';
  return '';
});

record('tts-gate-both-causes', () => {
  if (!spec.guiFile) return 'could not find the module that computes the pause gate';

  const env = makeEnv();
  const tts = loadTTS(env);
  if (spec.artifactFile) { try { global[spec.artifactGlobal] = loadModule(spec.artifactFile, spec.artifactGlobal); } catch (e) { /* optional */ } }
  global[spec.ttsGlobal] = tts;

  try { loadModule(spec.guiFile, '0'); } catch (e) {
    return 'could not load the module holding the pause gate: ' + (e && e.message ? e.message : e);
  }
  fire({ _listeners: env.docListeners }, 'DOMContentLoaded');
  fire({ _listeners: env.docListeners }, 'load');
  if (env.ws && typeof env.ws.onopen === 'function') { try { env.ws.onopen({ type: 'open' }); } catch (e) {} }

  // The input box is whichever element registered an 'input' listener. Deriving it
  // from the student's own tree means the check never depends on an element id.
  let input = null;
  for (const el of env.byId.values()) {
    if (el._listeners['input'] && el._listeners['input'].length) { input = el; break; }
  }
  if (!input) return 'no element registered an "input" listener, so the typing half of the gate is not wired';

  const gate = () => env.wsSent
    .map((s) => { try { return JSON.parse(s); } catch (e) { return null; } })
    .filter((m) => m && (m.type === 'pause' || m.type === 'unpause'))
    .map((m) => m.type);

  const before = gate().length;
  const since = () => gate().slice(before);

  // Cause one: speech starts.
  tts.queueChunk('Speaking now. ');
  if (show(since()) !== show(['pause'])) return 'starting speech sent ' + show(since()) + ', want exactly one "pause"; speaking is half the gate';

  // Cause two arrives while already paused: no edge, so nothing is sent.
  input.value = ' ';
  fire(input, 'input');
  if (show(since()) !== show(['pause'])) return 'a second cause while already paused sent ' + show(since()) + '; send only on an edge, not at every call site';

  // Speech ends, but a space is still in the box. A space is input.
  finishAudio(env);
  if (show(since()) !== show(['pause'])) {
    return 'unpaused while a space remained in the input box: ' + show(since()) + '; a space counts as input, and unpausing requires BOTH causes clear';
  }

  // Now the last cause clears.
  input.value = '';
  fire(input, 'input');
  if (show(since()) !== show(['pause', 'unpause'])) return 'clearing the last cause sent ' + show(since()) + ', want exactly one "unpause" after the "pause"';

  // Still clear: no edge, no message.
  fire(input, 'input');
  if (show(since()) !== show(['pause', 'unpause'])) return 'an unchanged gate value sent another message: ' + show(since());
  return '';
});

process.stdout.write(JSON.stringify({ checks }, null, 2) + '\n');
