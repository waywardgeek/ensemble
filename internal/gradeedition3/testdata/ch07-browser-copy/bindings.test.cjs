'use strict';
const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {readBindings, validateBindings} = require('./bindings.cjs');

const defaults = readBindings();
test('documented defaults and alternative public selectors are immutable values', () => {
  assert(Object.isFrozen(defaults));
  const input = {...defaults, composer: 'textarea[aria-label="Compose"]'};
  const result = validateBindings(input);
  input.composer = '#later';
  assert.equal(result.composer, 'textarea[aria-label="Compose"]');
  assert(Object.isFrozen(result));
});

test('binding data cannot omit fields or add code and skip switches', () => {
  const missing = {...defaults};
  delete missing.copy;
  assert.throws(() => validateBindings(missing), /exactly/);
  for (const extra of ['evaluate', 'skipClipboard', '__proto__']) {
    const input = JSON.parse(JSON.stringify(defaults));
    Object.defineProperty(input, extra, {value: 'anything', enumerable: true});
    assert.throws(() => validateBindings(input), /exactly/);
  }
});

test('invalid shape, version, labels and selector values refuse', () => {
  for (const value of [null, [], 'bindings', 1]) assert.throws(() => validateBindings(value));
  assert.throws(() => validateBindings({...defaults, version: 2}), /version/);
  for (const value of ['', ' ', null, 2, 'x\0y', 'x'.repeat(513)]) {
    assert.throws(() => validateBindings({...defaults, composer: value}), /composer/);
  }
  assert.throws(() => validateBindings({...defaults, copied: 'done\nnow'}), /single-line/);
  assert.throws(() => validateBindings({...defaults, copied: defaults.copying}), /distinct/);
});

test('file reads reject malformed and oversized data without evaluation', () => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ch07-bindings-test-'));
  try {
    const filename = path.join(directory, 'bindings.json');
    fs.writeFileSync(filename, '{broken');
    assert.throws(() => readBindings(filename), SyntaxError);
    fs.writeFileSync(filename, ' '.repeat(16385));
    assert.throws(() => readBindings(filename), /16 KiB/);
    fs.writeFileSync(filename, Buffer.from([0xff]));
    assert.throws(() => readBindings(filename), TypeError);
    const folder = path.join(directory, 'folder');
    fs.mkdirSync(folder);
    assert.throws(() => readBindings(folder), /regular file/);
    fs.writeFileSync(filename, JSON.stringify(defaults));
    assert.deepEqual(readBindings(filename), defaults);
  } finally {
    fs.rmSync(directory, {recursive: true}); // Only this test's newly owned directory.
  }
});
