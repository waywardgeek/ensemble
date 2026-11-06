'use strict';

const fs = require('node:fs');
const path = require('node:path');

// These locate public presentation, not application-owned JavaScript state.
// No field can supply code, change fixture payloads, or disable an assertion.
const selectors = Object.freeze([
  'connected', 'idle', 'composer', 'send', 'cards', 'snippet', 'code',
  'copy', 'feedback', 'manual', 'expand', 'expanded'
]);
const labels = Object.freeze(['copied', 'copying']);
const fields = Object.freeze(['version', ...selectors, ...labels]);

function validateBindings(value) {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error('Bindings must be a JSON object');
  }
  const keys = Object.keys(value).sort();
  if (JSON.stringify(keys) !== JSON.stringify([...fields].sort())) {
    throw new Error('Bindings must contain exactly the documented fields');
  }
  if (value.version !== 1) throw new Error('Unsupported bindings version');
  for (const name of [...selectors, ...labels]) {
    const text = value[name];
    if (typeof text !== 'string' || !text.trim() || text.length > 512 || text.includes('\0')) {
      throw new Error('Invalid binding: ' + name);
    }
  }
  for (const name of labels) {
    if (/[\r\n]/.test(value[name])) throw new Error('Feedback labels must be single-line');
  }
  if (value.copied === value.copying) {
    throw new Error('Pending and successful feedback must be distinct');
  }
  return Object.freeze({...value});
}

function readBindings(filename = path.join(__dirname, 'bindings.json')) {
  // Nonblocking open plus descriptor validation refuses FIFOs/devices without
  // waiting for a writer. Symlinks to ordinary explicitly selected files work.
  const descriptor = fs.openSync(filename, fs.constants.O_RDONLY | fs.constants.O_NONBLOCK);
  try {
    const info = fs.fstatSync(descriptor);
    if (!info.isFile()) throw new Error('Bindings must be a regular file');
    if (info.size > 16384) throw new Error('Bindings file exceeds 16 KiB');
    const bytes = Buffer.alloc(16385);
    let used = 0;
    while (used < bytes.length) {
      const count = fs.readSync(descriptor, bytes, used, bytes.length - used, null);
      if (count === 0) break;
      used += count;
    }
    if (used > 16384) throw new Error('Bindings file exceeds 16 KiB');
    const text = new TextDecoder('utf-8', {fatal: true}).decode(bytes.subarray(0, used));
    return validateBindings(JSON.parse(text));
  } finally {
    fs.closeSync(descriptor);
  }
}

module.exports = {readBindings, validateBindings};
