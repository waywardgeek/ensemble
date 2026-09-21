// tts-browser.js - a headless stand-in for the browser that runs the GUI.
//
// This is half of the chapter 14 test harness. The other half is
// tts-harness.sh, which starts the agent and then runs this.
//
// The job here is to be a browser: load the real GUI scripts, in the real
// order index.html loads them, connect to the real agent over a real
// WebSocket, type a real prompt, and wait for the turn to end. Nothing in
// this file reimplements any part of the speech pipeline. If it did, the
// harness would be testing itself instead of the GUI.
//
// One thing is deliberately not real: the speech engine. See makeSynth below.
//
// Usage:
//   node tts-browser.js --url http://127.0.0.1:8099 --prompt "hello"
//   node tts-browser.js --url ... --prompt "..." --type-during-turn "typing"

'use strict';

const fs = require('fs');
const path = require('path');
const vm = require('vm');

// ---------------------------------------------------------------- arguments

function parseArgs(argv) {
  const out = { url: '', prompt: '', typeDuringTurn: '', timeoutMs: 30000 };
  for (let i = 0; i < argv.length; i++) {
    const next = () => argv[++i];
    switch (argv[i]) {
      case '--url': out.url = next(); break;
      case '--prompt': out.prompt = next(); break;
      case '--type-during-turn': out.typeDuringTurn = next(); break;
      case '--timeout-ms': out.timeoutMs = parseInt(next(), 10); break;
    }
  }
  if (!out.url) throw new Error('--url is required');
  return out;
}

const args = parseArgs(process.argv.slice(2));
const guiDir = path.join(__dirname, '..', 'web', 'gui');

// ------------------------------------------------------------------ the DOM
//
// A minimum viable DOM. Elements are created on demand and cached by id, so
// two lookups of the same id return the same object and a listener registered
// on the first is still there on the second.

function makeElement(tag, id) {
  const el = {
    tagName: (tag || 'div').toUpperCase(),
    id: id || '',
    _listeners: {},
    children: [],
    style: {},
    dataset: {},
    _attrs: {},
    textContent: '',
    innerHTML: '',
    value: '',
    scrollTop: 0,
    scrollHeight: 0,
    clientHeight: 0,
    classList: {
      _set: new Set(),
      add(c) { this._set.add(c); },
      remove(c) { this._set.delete(c); },
      toggle(c, on) { if (on === undefined) { this._set.has(c) ? this._set.delete(c) : this._set.add(c); } else if (on) { this._set.add(c); } else { this._set.delete(c); } },
      contains(c) { return this._set.has(c); },
    },
    addEventListener(type, fn) { (this._listeners[type] = this._listeners[type] || []).push(fn); },
    removeEventListener(type, fn) {
      const l = this._listeners[type];
      if (l) this._listeners[type] = l.filter((f) => f !== fn);
    },
    appendChild(c) { this.children.push(c); return c; },
    removeChild(c) { this.children = this.children.filter((x) => x !== c); return c; },
    insertBefore(c) { this.children.unshift(c); return c; },
    replaceChildren(...kids) { this.children = kids; },
    append(...kids) { for (const k of kids) this.children.push(k); },
    prepend(...kids) { this.children.unshift(...kids); },
    remove() {},
    contains() { return false; },
    closest() { return null; },
    matches() { return false; },
    scrollIntoView() {},
    focus() {},
    blur() {},
    click() { this.dispatchEvent({ type: 'click', target: this }); },
    setAttribute(k, v) { this._attrs[k] = String(v); },
    getAttribute(k) { return Object.prototype.hasOwnProperty.call(this._attrs, k) ? this._attrs[k] : null; },
    removeAttribute(k) { delete this._attrs[k]; },
    hasAttribute(k) { return Object.prototype.hasOwnProperty.call(this._attrs, k); },
    querySelector() { return makeElement('div'); },
    querySelectorAll() { return []; },
    getBoundingClientRect() { return { top: 0, left: 0, width: 800, height: 600, bottom: 600, right: 800 }; },
    focus() {}, blur() {}, click() { fire(this, 'click', {}); },
    scrollIntoView() {}, remove() {},
    closest() { return null; },
  };
  return el;
}

// Dispatch an event to whatever listeners an element registered. This is how
// the harness "types": it is the same path a real keystroke takes.
function fire(el, type, props) {
  const ev = Object.assign({
    type,
    preventDefault() {}, stopPropagation() {},
    target: el, currentTarget: el,
  }, props || {});
  for (const fn of (el._listeners[type] || [])) fn(ev);
  return ev;
}

// ------------------------------------------------------------ speech engine
//
// The one deliberately unreal part of this harness, and the reason it exists.
//
// A real speechSynthesis cannot be observed: it has no transcript API, so
// there is no way to ask it what it said. Worse, in a headless browser it does
// not work at all - zero voices installed, and speak() fails outright with
// "not-allowed". A queue that advances when an utterance reports it finished
// would stall forever on the first one.
//
// So record mode REPLACES the engine rather than driving it. speak() records
// the utterance and immediately reports that it finished, which keeps the
// queue moving at full speed. The recording that matters is not made here
// though: it is made by the GUI itself, sent to the agent, and written to the
// speech log. This stub exists only so the pipeline has somewhere to speak to.

function makeSynth() {
  const synth = {
    speaking: false,
    paused: false,
    _spoken: [],
    getVoices() { return [{ name: 'Harness Voice', lang: 'en-US', default: true }]; },
    speak(u) {
      synth.speaking = true;
      synth._spoken.push(u.text);
      // Complete on a later tick, the way a real engine would, so callers
      // that set onend after calling speak() still see it fire.
      setTimeout(() => {
        synth.speaking = false;
        if (typeof u.onend === 'function') u.onend({ type: 'end' });
      }, 0);
    },
    cancel() {
      synth.speaking = false;
    },
    pause() { synth.paused = true; },
    resume() { synth.paused = false; },
  };
  return synth;
}

function SpeechSynthesisUtterance(text) {
  this.text = text;
  this.lang = 'en-US';
  this.rate = 1;
  this.pitch = 1;
  this.volume = 1;
  this.voice = null;
  this.onend = null;
  this.onerror = null;
  this.onstart = null;
}

// ------------------------------------------------------------------ globals
//
// node 21 and later define some of these as getter-only properties, so every
// global goes in through defineProperty rather than plain assignment.

function define(name, value) {
  Object.defineProperty(globalThis, name, {
    value, writable: true, configurable: true, enumerable: true,
  });
}

const elementsById = new Map();

const documentStub = {
  _listeners: {},
  body: makeElement('body'),
  documentElement: makeElement('html'),
  readyState: 'complete',
  title: '',
  getElementById(id) {
    if (!elementsById.has(id)) elementsById.set(id, makeElement('div', id));
    return elementsById.get(id);
  },
  createElement(tag) { return makeElement(tag); },
  createTextNode(t) { const e = makeElement('span'); e.textContent = t; return e; },
  createDocumentFragment() { return makeElement('fragment'); },
  querySelector(sel) {
    const id = sel.startsWith('#') ? sel.slice(1) : sel.replace(/[^a-zA-Z0-9_-]/g, '');
    return documentStub.getElementById(id);
  },
  querySelectorAll() { return []; },
  addEventListener(type, fn) { (documentStub._listeners[type] = documentStub._listeners[type] || []).push(fn); },
  removeEventListener() {},
  dispatchEvent() { return true; },
};

const url = new URL(args.url);

define('document', documentStub);
define('location', {
  protocol: url.protocol,
  host: url.host,
  hostname: url.hostname,
  port: url.port,
  href: args.url,
  origin: url.origin,
  pathname: '/',
  search: '',
  reload() {},
});
define('navigator', { userAgent: 'tts-harness', language: 'en-US', platform: 'harness' });
define('localStorage', {
  _d: new Map(),
  getItem(k) { return this._d.has(k) ? this._d.get(k) : null; },
  setItem(k, v) { this._d.set(k, String(v)); },
  removeItem(k) { this._d.delete(k); },
  clear() { this._d.clear(); },
});
define('speechSynthesis', makeSynth());
define('SpeechSynthesisUtterance', SpeechSynthesisUtterance);
define('requestAnimationFrame', (fn) => setTimeout(() => fn(Date.now()), 0));
define('cancelAnimationFrame', (h) => clearTimeout(h));
define('matchMedia', () => ({ matches: false, addEventListener() {}, removeEventListener() {} }));
define('alert', () => {});
define('scrollTo', () => {});
define('getComputedStyle', () => ({ getPropertyValue: () => '' }));

// The two libraries index.html pulls from a CDN. The speech channel does not
// depend on either, so a pass-through is enough.
define('marked', { parse: (s) => String(s), setOptions() {} });
define('AnsiUp', function AnsiUp() { this.ansi_to_html = (s) => String(s); });

// window is this same global object, the way it is in a browser.
define('window', globalThis);
globalThis.addEventListener = (type, fn) => documentStub.addEventListener(type, fn);
globalThis.removeEventListener = () => {};

// Count what arrives on the socket. Only used by HARNESS_DEBUG, but it has to
// be installed before the GUI opens its connection.
globalThis.__harnessMsgCount = 0;
globalThis.__harnessKinds = {};
const NativeWebSocket = globalThis.WebSocket;
function CountingWebSocket(u, p) {
  const ws = new NativeWebSocket(u, p);
  ws.addEventListener('message', (ev) => {
    globalThis.__harnessMsgCount++;
    let kind = 'unparsed';
    try { kind = (JSON.parse(ev.data) || {}).type || 'no-type'; } catch (_) { /* leave it */ }
    globalThis.__harnessKinds[kind] = (globalThis.__harnessKinds[kind] || 0) + 1;
  });
  return ws;
}
// The GUI checks ws.readyState against WebSocket.OPEN, so the stand-in has to
// carry the same constants the real constructor does. Without them that check
// compares a number against undefined and quietly never sends anything.
for (const k of ['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED']) {
  CountingWebSocket[k] = NativeWebSocket[k];
}
define('WebSocket', CountingWebSocket);

// ------------------------------------------------------------ script loader
//
// vm.runInThisContext shares one global lexical scope across calls, so a
// top-level `const TTS = ...` in tts.js is visible to artifact-scroll.js and
// gui.js afterwards. That is exactly how <script> tags behave, which makes
// this a faithful loader rather than a clever one.

function loadScript(file) {
  const full = path.join(guiDir, file);
  const src = fs.readFileSync(full, 'utf8');
  try {
    vm.runInThisContext(src, { filename: full });
  } catch (err) {
    console.error(`[harness] failed loading ${file}: ${err.message}`);
    throw err;
  }
}

// The same order index.html uses, minus mcp.js. The speech channel does not
// involve MCP, and mcp.js replaces the WebSocket constructor to attach itself,
// which would put a second consumer on the socket for no benefit here.
for (const f of ['renderers.js', 'tts.js', 'artifact-scroll.js', 'gui.js']) {
  loadScript(f);
}

// A top-level `const TTS = ...` in a script goes into the global lexical
// scope, which later scripts can see but plain module code here cannot. So
// reach into that scope by name when the harness needs to look at something
// the GUI declared.
function readGlobal(name) {
  return vm.runInThisContext(`typeof ${name} !== 'undefined' ? ${name} : undefined`);
}

// ----------------------------------------------------------------- the turn

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

async function waitFor(predicate, timeoutMs, what) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    if (predicate()) return true;
    await sleep(20);
  }
  throw new Error(`timed out after ${timeoutMs}ms waiting for ${what}`);
}

function typeInto(text, submit) {
  // The harness belongs to the same project as the GUI, so it is allowed to
  // know what the GUI calls its input box. Typing goes through the element's
  // own listeners, which is the same path a real keystroke takes.
  const el = documentStub.getElementById('prompt-input');
  el.value = text;
  fire(el, 'input', { target: el });
  if (submit) {
    fire(el, 'keydown', { key: 'Enter', shiftKey: false, target: el });
  }
  return el;
}

// The GUI keeps speech switched off until the settings say otherwise, so the
// harness turns it on the way a person would: by changing the setting. The
// agent's own settings have to arrive first, because applying them afterwards
// would put the switch straight back.
async function enableSpeech() {
  await waitFor(
    () => (globalThis.__harnessKinds || {}).current_settings,
    10000,
    "the agent's settings to arrive",
  );
  const box = documentStub.getElementById('set-tts-enabled');
  box.checked = true;
  fire(box, 'change', { target: box });
  await waitFor(
    () => { const t = readGlobal('TTS'); return t && t.enabled; },
    5000,
    'speech to be switched on',
  );
}

async function main() {
  // gui.js opened its own WebSocket on load. Wait for the agent to answer.
  await waitFor(() => globalThis.agentState !== undefined, 10000, 'the GUI to initialise');
  await sleep(300); // let the socket settle and the first state arrive

  await enableSpeech();

  if (!args.prompt) {
    console.error('[harness] no --prompt given; nothing to do');
    return;
  }

  typeInto(args.prompt, true);

  // Wait for the agent to say it has finished. The turn-ended message is the
  // reliable signal: polling the displayed state can miss a fast turn
  // entirely, and then the harness sits here until it times out.
  const deadline = Date.now() + args.timeoutMs;
  let typedDuringTurn = false;

  while (Date.now() < deadline) {
    if ((globalThis.__harnessKinds || {}).turn_ended) {
      // Do not leave the instant the turn ends. A failure that arrives just
      // behind the turn would otherwise never be recorded, and silence in
      // the log would be blamed on the system rather than on this harness.
      await sleep(600);
      break;
    }

    // The pause-gate scenario: type while the agent is still talking.
    if (args.typeDuringTurn && !typedDuringTurn && (globalThis.__harnessKinds || {}).part_delta) {
      typedDuringTurn = true;
      typeInto(args.typeDuringTurn, false);
      // Hold the keystroke long enough for the gate to shut, then clear it so
      // the gate reopens and the log shows both edges.
      await sleep(1200);
      const el = documentStub.getElementById('prompt-input');
      el.value = '';
      fire(el, 'input', { target: el });
    }

    await sleep(20);
  }

  // Speak anything still sitting in the buffer. Harmless if the GUI already
  // flushed on turn end.
  const tts = readGlobal('TTS');
  if (tts && typeof tts.flush === 'function') tts.flush();

  if (process.env.HARNESS_DEBUG) {
    console.error('[debug] agentState      =', globalThis.agentState);
    console.error('[debug] TTS present     =', typeof tts);
    console.error('[debug] TTS.enabled     =', tts && tts.enabled);
    console.error('[debug] TTS.onRecord    =', tts && typeof tts.onRecord);
    console.error('[debug] transcript len  =', tts && tts.transcript && tts.transcript.length);
    console.error('[debug] engine spoke    =', globalThis.speechSynthesis._spoken.length);
    console.error('[debug] ws msgs seen    =', globalThis.__harnessMsgCount || 0);
    console.error('[debug] ws kinds        =', JSON.stringify(globalThis.__harnessKinds || {}));
  }

  // The GUI reports each utterance to the agent over the socket, and the
  // agent writes the log. Give those last messages time to arrive before the
  // process exits and the socket closes.
  await sleep(600);
}

main().then(
  () => process.exit(0),
  (err) => { console.error(`[harness] ${err.message}`); process.exit(1); },
);
