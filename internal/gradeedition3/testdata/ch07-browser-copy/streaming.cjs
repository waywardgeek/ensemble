'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const {readBindings} = require('./bindings.cjs');

async function main(argv) {
  if (argv.length < 4 || argv.length > 5) {
    throw new Error('Expected owned GUI URL, fixture JSON, provider URL, Playwright module, optional bindings JSON');
  }
  const [url, fixturePath, provider, modulePath, bindingPath] = argv;
  const target = new URL(url);
  if (target.protocol !== 'http:' || !['127.0.0.1', 'localhost', '[::1]'].includes(target.hostname)) {
    throw new Error('The browser component requires an owned loopback HTTP GUI');
  }
  const providerTarget = new URL(provider);
  if (providerTarget.protocol !== 'http:' || providerTarget.hostname !== '127.0.0.1') {
    throw new Error('Streaming gates require the owned loopback provider');
  }
  const binding = readBindings(bindingPath);
  const fixture = JSON.parse(fs.readFileSync(fixturePath, 'utf8'));
  for (const key of ['prompt', 'marker', 'closed', 'later', 'finished', 'payload', 'markdown']) {
    if (typeof fixture[key] !== 'string' || !fixture[key]) throw new Error('Invalid fixture: ' + key);
  }
  const {chromium} = require(modulePath);
  const checks = [], sent = [], received = [], errors = [], blocked = [];
  let browser;
  let failure = null;
  let interrupted = false;
  const stop = () => {
    interrupted = true;
    void browser?.close().catch(error => {
      failure ??= 'Browser interruption cleanup failed: ' + error.message;
    });
  };
  process.once('SIGTERM', stop);
  try {
    browser = await chromium.launch({channel: 'chrome', headless: true, timeout: 15000});
    if (interrupted) throw new Error('Browser component interrupted');
    const context = await browser.newContext();
    await context.addInitScript(() => {
      // Fixture-owned state substitutes native I/O only. It is not application state.
      const clipboard = window.streamFixtureClipboard = {writes: [], denied: false, reads: 0};
      Object.defineProperty(navigator, 'clipboard', {
        configurable: true,
        value: {
          async writeText(text) {
            clipboard.writes.push(text);
            if (clipboard.denied) throw new Error('fixture denial');
          },
          async readText() { clipboard.reads++; throw new Error('clipboard read forbidden'); },
          async read() { clipboard.reads++; throw new Error('clipboard read forbidden'); }
        }
      });
      Object.defineProperty(window, 'speechSynthesis', {
        configurable: true,
        value: {
          getVoices: () => [], addEventListener() {},
          speak(utterance) { queueMicrotask(() => utterance.onend?.()); },
          cancel() {}, pause() {}, resume() {}
        }
      });
      window.SpeechSynthesisUtterance = class { constructor(text) { this.text = text; } };
    });
    await context.route('**/*', route => {
      if (new URL(route.request().url()).origin === target.origin) return route.continue();
      blocked.push(route.request().url());
      return route.abort();
    });
    const page = await context.newPage();
    page.setDefaultTimeout(8000);
    page.on('pageerror', error => errors.push(error.message));
    page.on('websocket', socket => {
      socket.on('framereceived', frame => {
        try { received.push(JSON.parse(String(frame.payload))); }
        catch (error) { errors.push(error.message); }
      });
      socket.on('framesent', frame => {
        try { sent.push(JSON.parse(String(frame.payload))); }
        catch (error) { errors.push(error.message); }
      });
    });
    async function step(number) {
      const response = await page.request.post(providerTarget.origin + '/fixture/step/' + number);
      assert.equal(response.status(), 200, 'owned-provider-gate');
    }
    await page.goto(url);
    await page.locator(binding.connected).waitFor();
    await page.locator(binding.composer).fill(fixture.prompt);
    await page.locator(binding.send).click();
    const answer = page.locator(binding.cards).filter({hasText: fixture.marker});
    await answer.getByText(fixture.marker, {exact: true}).waitFor();
    assert.equal(await answer.locator(binding.copy).count(), 0,
      'unfinished-stream-fence-has-no-copy');
    checks.push('unfinished-fence-visible-without-copy-before-release');

    await step(1);
    await answer.getByText(fixture.closed, {exact: true}).waitFor();
    const snippet = answer.locator(binding.snippet).filter({has: page.locator(binding.copy)});
    const copy = snippet.locator(binding.copy);
    await copy.waitFor();
    await snippet.locator(binding.expand).click();
    await copy.click();
    await snippet.locator(binding.feedback).filter({hasText: binding.copied}).waitFor();
    assert.equal(await page.evaluate(() => window.streamFixtureClipboard.writes.at(-1)),
      fixture.payload);
    checks.push('closed-stream-fence-copies-exact-payload');
    await copy.focus();
    const original = await copy.elementHandle();

    await step(2);
    await answer.getByText(fixture.later, {exact: true}).waitFor();
    assert(await original.evaluate(node => node.isConnected && document.activeElement === node),
      'stream-update-retains-original-focused-copy-control');
    checks.push('stream-update-retains-original-focused-copy-control');
    assert.equal(await snippet.locator(binding.expanded).count(), 1);
    assert.equal(await snippet.locator(binding.feedback).innerText(), binding.copied);
    checks.push('stream-update-retains-expansion-and-feedback');
    await page.evaluate(() => { window.streamFixtureClipboard.denied = true; });
    await copy.click();
    const manual = snippet.locator(binding.manual);
    await manual.waitFor();
    assert.equal(await manual.inputValue(), fixture.payload);
    const originalManual = await manual.elementHandle();
    assert(await originalManual.evaluate(node => node.readOnly && document.activeElement === node &&
      node.selectionStart === 0 && node.selectionEnd === node.value.length));
    checks.push('manual-fallback-established-before-final-delta');

    await step(3);
    await answer.getByText(fixture.finished, {exact: true}).waitFor();
    const deadline = Date.now() + 8000;
    while (!received.some(frame => frame.type === 'turn_ended' && frame.reason === 'completed') &&
           Date.now() < deadline) await page.waitForTimeout(10);
    assert(received.some(frame => frame.type === 'turn_ended' && frame.reason === 'completed'),
      'completed-stream-turn-on-public-wire');
    assert.equal(received.filter(frame => frame.type === 'part_final' &&
      frame.part?.type === 'text' && frame.part.text === fixture.markdown).length, 1,
      'exact-stream-final-on-public-wire');
    await page.locator(binding.idle).waitFor();
    checks.push('canonical-final-and-completed-turn-applied-before-selection-check');
    assert(await originalManual.evaluate(node => node.isConnected && node.readOnly &&
      document.activeElement === node && node.selectionStart === 0 &&
      node.selectionEnd === node.value.length), 'stream-update-retains-original-manual-selection');
    checks.push('stream-update-retains-original-manual-selection');
    assert.equal(await manual.inputValue(), fixture.payload);
    assert.equal(sent.filter(frame => frame.type === 'prompt').length, 1);
    assert.equal(await page.evaluate(() => window.streamFixtureClipboard.reads), 0);
    checks.push('stream-copy-does-not-read-or-resubmit');
    assert.deepEqual(errors, []);
    assert.deepEqual(blocked, []);
    checks.push('stream-page-has-no-errors-or-off-origin-fetches');
    await original.dispose();
    await originalManual.dispose();
    await context.close();
  } catch (error) {
    failure = error.message;
  } finally {
    await browser?.close();
    process.removeListener('SIGTERM', stop);
  }
  if (interrupted) failure ??= 'Browser component interrupted';
  console.log(JSON.stringify({
    contract: 'chapter7-stream-browser-component-v1',
    passed: failure === null, checks, failure, errors, blocked,
    promptFrames: sent.filter(frame => frame.type === 'prompt').length
  }));
  return interrupted ? 2 : failure === null ? 0 : 1;
}

if (require.main === module) {
  main(process.argv.slice(2)).then(code => { process.exitCode = code; }).catch(error => {
    console.error('Browser component setup or cleanup failed: ' + error.message);
    process.exitCode = 2;
  });
}
module.exports = {main};
