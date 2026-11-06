'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const {readBindings} = require('./bindings.cjs');

async function main(argv) {
  if (argv.length < 3 || argv.length > 4) {
    throw new Error('Expected owned GUI URL, fixture JSON, Playwright module, optional bindings JSON');
  }
  const [url, fixturePath, modulePath, bindingPath] = argv;
  const target = new URL(url);
  if (target.protocol !== 'http:' || !['127.0.0.1', 'localhost', '[::1]'].includes(target.hostname)) {
    throw new Error('The browser component requires an owned loopback HTTP GUI');
  }
  const binding = readBindings(bindingPath);
  const fixture = JSON.parse(fs.readFileSync(fixturePath, 'utf8'));
  for (const key of ['prompt', 'marker', 'human', 'markdown']) {
    if (typeof fixture[key] !== 'string' || !fixture[key]) throw new Error('Invalid fixture: ' + key);
  }
  if (!Array.isArray(fixture.payloads) || fixture.payloads.length !== 4 ||
      fixture.payloads.some(value => typeof value !== 'string')) {
    throw new Error('The buffered fixture requires four independently specified payloads');
  }
  const {chromium} = require(modulePath);
  const checks = [], sent = [], errors = [], blocked = [];
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
      const clipboard = window.copyFixtureClipboard = {writes: [], denied: false, reads: 0};
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
      socket.on('framesent', frame => {
        try { sent.push(JSON.parse(String(frame.payload))); }
        catch (error) { errors.push(error.message); }
      });
    });
    await page.goto(url);
    await page.locator(binding.connected).waitFor();
    const composer = page.locator(binding.composer);
    await composer.fill(fixture.prompt);
    await page.locator(binding.send).click();
    const answer = page.locator(binding.cards).filter({hasText: fixture.marker});
    await answer.getByText(fixture.marker, {exact: true}).waitFor();
    checks.push('actual-provider-markdown-visible');

    assert.equal(await answer.locator(binding.copy).count(), fixture.payloads.length,
      'closed-assistant-copy-controls');
    checks.push('only-closed-assistant-fences-copyable');
    const snippets = answer.locator(binding.snippet).filter({has: page.locator(binding.copy)});
    const human = page.locator(binding.cards).filter({hasText: fixture.human});
    assert.equal(await human.count(), 1);
    assert.equal(await human.locator(binding.copy).count(), 0);
    checks.push('real-human-fence-has-no-copy-control');
    assert.deepEqual(await page.evaluate(() => window.copyFixtureClipboard.writes), []);
    checks.push('clipboard-write-requires-click');
    assert(fixture.payloads[2].length > 1800);
    assert.equal(await snippets.nth(2).locator(binding.expanded).count(), 0);
    checks.push('long-copy-starts-collapsed');

    const draft = 'Unsent draft 界 ' + fixture.marker;
    await composer.fill(draft);
    for (let index = 0; index < fixture.payloads.length; index++) {
      const snippet = snippets.nth(index);
      await snippet.locator(binding.copy).click();
      assert.equal(await page.evaluate(() => window.copyFixtureClipboard.writes.at(-1)),
        fixture.payloads[index], 'logical-payload-' + index);
      await snippet.locator(binding.feedback).filter({hasText: binding.copied}).waitFor();
    }
    checks.push('exact-logical-payloads-including-empty-and-full-long-code');
    assert.equal(await answer.locator('table').count(), 1);
    assert.equal(await answer.locator('img,script,a[href^="javascript:"]').count(), 0);
    assert.equal(await page.evaluate(() => window.copyFixtureExecuted), undefined);
    assert.deepEqual(blocked, []);
    checks.push('table-and-literal-safe-rendering-without-image-fetch');

    await page.evaluate(() => { window.copyFixtureClipboard.denied = true; });
    await snippets.nth(2).locator(binding.copy).click();
    const manual = snippets.nth(2).locator(binding.manual);
    await manual.waitFor();
    assert.equal(await manual.inputValue(), fixture.payloads[2]);
    assert(await manual.evaluate(node => node.readOnly && document.activeElement === node &&
      node.selectionStart === 0 && node.selectionEnd === node.value.length));
    checks.push('denied-write-shows-exact-readonly-selected-fallback');
    await page.evaluate(() => Object.defineProperty(navigator, 'clipboard', {
      configurable: true, value: undefined
    }));
    await snippets.nth(0).locator(binding.copy).click();
    const absent = snippets.nth(0).locator(binding.manual);
    await absent.waitFor();
    assert.equal(await absent.inputValue(), fixture.payloads[0]);
    checks.push('missing-api-shows-exact-fallback');
    assert.equal(await composer.inputValue(), draft);
    assert.equal(sent.filter(frame => frame.type === 'prompt').length, 1);
    assert.equal(await page.evaluate(() => window.copyFixtureClipboard.reads), 0);
    checks.push('copy-preserves-draft-and-does-not-read-or-submit');
    assert.deepEqual(errors, []);
    assert.deepEqual(blocked, []);
    checks.push('no-page-errors-or-off-origin-fetches');
    await context.close();
  } catch (error) {
    failure = error.message;
  } finally {
    await browser?.close();
    process.removeListener('SIGTERM', stop);
  }
  if (interrupted) failure ??= 'Browser component interrupted';
  console.log(JSON.stringify({
    contract: 'chapter7-copy-browser-component-v1',
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
