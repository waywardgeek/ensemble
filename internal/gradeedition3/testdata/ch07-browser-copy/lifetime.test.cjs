'use strict';
const assert = require('node:assert/strict');
const test = require('node:test');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const {spawn} = require('node:child_process');

const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));

async function waitFor(predicate, message) {
  const deadline = Date.now() + 5000;
  while (!predicate() && Date.now() < deadline) await delay(10);
  assert(predicate(), message);
}

test('interruption joins browser cleanup before reporting; no-join variant is detected',
  {timeout: 15000}, async () => {
    const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'ch07-browser-lifetime-'));
    try {
      const fixture = path.join(directory, 'fixture.json');
      fs.writeFileSync(fixture, JSON.stringify({
        prompt: 'prompt', marker: 'marker', human: 'human', markdown: 'fixture',
        closed: 'closed', later: 'later', finished: 'finished', payload: 'code',
        payloads: ['a', 'b', 'c', '']
      }));
      const modulePath = path.join(directory, 'controlled-playwright.cjs');
      fs.writeFileSync(modulePath, `
let rejectNavigation;
let releaseClose;
let closing = false;
const close = new Promise(resolve => { releaseClose = resolve; });
process.on('message', message => {
  if (message.kind === 'release-close') {
    releaseClose();
    process.disconnect();
  }
});
const page = {
  setDefaultTimeout() {}, on() {},
  goto() {
    const pending = new Promise((resolve, reject) => { rejectNavigation = reject; });
    process.send({kind: 'navigation-pending'});
    return pending;
  }
};
const context = {async addInitScript() {}, async route() {}, async newPage() { return page; }};
const browser = {
  async newContext() { return context; },
  close() {
    if (!closing) {
      closing = true;
      process.send({kind: 'close-pending'});
      rejectNavigation(new Error('controlled browser closure'));
    }
    return close;
  }
};
module.exports = {chromium: {async launch() { return browser; }}};
`);
      for (const filename of ['bindings.cjs', 'bindings.json']) {
        fs.copyFileSync(path.join(__dirname, filename), path.join(directory, filename));
      }
      async function run(script, streaming) {
        const childArguments = [script, 'http://127.0.0.1:1/', fixture];
        if (streaming) childArguments.push('http://127.0.0.1:2/');
        childArguments.push(modulePath);
        const child = spawn(process.execPath, childArguments, {
          stdio: ['ignore', 'pipe', 'pipe', 'ipc']
        });
        const messages = [];
        let output = '', errors = '';
        child.on('message', value => messages.push(value));
        child.stdout.on('data', value => { output += value; });
        child.stderr.on('data', value => { errors += value; });
        const exited = new Promise((resolve, reject) => {
          child.once('error', reject);
          child.once('close', (code, signal) => resolve({code, signal}));
        });
        try {
          await waitFor(() => messages.some(value => value.kind === 'navigation-pending'),
            'component reaches controlled navigation');
          child.kill('SIGTERM');
          await waitFor(() => messages.some(value => value.kind === 'close-pending'),
            'interruption reaches browser cleanup');
          await delay(100);
          const publishedBeforeJoin = output.trim().length > 0;
          assert.equal(child.exitCode, null, 'controlled cleanup remains held');
          child.send({kind: 'release-close'});
          const exit = await exited;
          assert.equal(exit.code, 2, errors);
          assert.equal(exit.signal, null);
          const report = JSON.parse(output);
          assert.equal(report.passed, false);
          return publishedBeforeJoin;
        } finally {
          if (child.exitCode === null && child.signalCode === null) child.kill('SIGKILL');
          await exited;
        }
      }
      for (const component of ['buffered.cjs', 'streaming.cjs']) {
        const original = fs.readFileSync(path.join(__dirname, component), 'utf8');
        const anchor = '    await browser?.close();\n    process.removeListener';
        assert.equal(original.split(anchor).length, 2, component + ': one final browser join');
        const mutant = path.join(directory, 'no-join-' + component);
        fs.writeFileSync(mutant, original.replace(anchor,
          '    browser?.close();\n    process.removeListener'));
        const streaming = component === 'streaming.cjs';
        assert.equal(await run(path.join(__dirname, component), streaming), false,
          component + ': result is withheld until cleanup joins');
        assert.equal(await run(mutant, streaming), true,
          component + ': removing the join publishes a premature result');
      }
    } finally {
      fs.rmSync(directory, {recursive: true});
    }
  });
