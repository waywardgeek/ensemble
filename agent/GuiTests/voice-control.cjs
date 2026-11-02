const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const http = require('http');
const fs = require('fs');
const path = require('path');
const assert = require('assert/strict');
const root = path.resolve(__dirname, '../gui/wwwroot');
const server = http.createServer((req, res) => {
  const name = req.url.slice(1);
  if (!name) { res.setHeader('Content-Type', 'text/html'); res.end('<button id="mic">Hold</button><button id="cancel" hidden>Cancel</button><div id="status"></div>'); return; }
  if (!/^voice-[a-z]+\.js$/.test(name)) { res.writeHead(404); res.end(); return; }
  res.setHeader('Content-Type', 'text/javascript'); res.end(fs.readFileSync(path.join(root, name)));
});
(async () => {
  let browser;
  try {
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    browser = await chromium.launch({ headless: true, channel: 'chrome' });
    for (const mode of ['funding', 'untrusted', 'pointer', 'keyboard', 'draft', 'selection', 'blur', 'uncertain']) {
      const page = await browser.newPage(); const errors = [];
      page.on('pageerror', error => errors.push(error.message));
      await page.goto(`http://127.0.0.1:${server.address().port}/`);
      await page.evaluate(async mode => {
        const { VoiceControl } = await import('./voice-control.js');
        const { VoiceSession } = await import('./voice-session.js');
        window.sent = []; window.starts = 0; window.cleanups = 0; window.uncertain = mode === 'uncertain';
        window.connection = { readyState: 1, voiceFunding: mode !== 'funding', send(text) {
          const frame = JSON.parse(text); sent.push(frame);
          const types = { voice_begin: 'voice_begun', voice_audio: 'voice_ready', voice_commit: 'voice_accepted' };
          if (types[frame.type] && !(uncertain && frame.type === 'voice_commit'))
            queueMicrotask(() => control.receive(connection, { type: types[frame.type], agent_id: frame.agent_id, sequence: frame.sequence }));
        }};
        window.control = new VoiceControl(document.querySelector('#mic'), document.querySelector('#cancel'), document.querySelector('#status'), options => {
          const hold = new class extends EventTarget {
            async start() { starts++; return true; }
            async release() { return new Uint8Array(46); }
            async cancel() { cleanups++; }
          }();
          return new VoiceSession({ ...options, hold });
        });
        control.bind(connection, '', true, false); // Canonical unnamed root is valid.
      }, mode);
      const mic = page.locator('#mic');
      if (mode === 'funding') {
        assert.equal(await mic.isDisabled(), true);
      } else if (mode === 'untrusted') {
        await mic.dispatchEvent('pointerdown', { button: 0, pointerId: 1 });
        await mic.dispatchEvent('keydown', { code: 'Space' });
      } else if (mode === 'keyboard') {
        await mic.focus(); await page.keyboard.down('Space'); await page.keyboard.down('Space');
        await page.waitForFunction(() => starts === 1); await page.keyboard.up('Space');
        await page.waitForFunction(() => !control.blocked && document.querySelector('#status').textContent === 'Voice message sent.');
      } else {
        const box = await mic.boundingBox(); await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
        await page.mouse.down(); await page.waitForFunction(() => starts === 1);
        if (mode === 'draft') await page.evaluate(() => control.bind(connection, '', true, true));
        if (mode === 'selection') await page.evaluate(() => control.bind(connection, 'other', true, false));
        if (mode === 'blur') await page.evaluate(() => window.dispatchEvent(new Event('blur')));
        await page.mouse.up();
        if (mode === 'pointer') await page.waitForFunction(() => !control.blocked && document.querySelector('#status').textContent === 'Voice message sent.');
        else if (mode === 'uncertain') {
          await page.waitForFunction(() => sent.some(frame => frame.type === 'voice_commit'));
          await page.evaluate(() => { connection.readyState = 3; control.bind(connection, '', false, false); });
          await page.waitForFunction(() => document.querySelector('#cancel').textContent === 'Dismiss uncertainty' && !document.querySelector('#cancel').disabled);
          assert.equal(await page.evaluate(() => control.blocked), true);
          await page.locator('#cancel').click();
          assert.equal(await page.evaluate(() => control.blocked), false);
        } else await page.waitForFunction(() => !control.blocked);
      }
      const result = await page.evaluate(() => ({ starts, cleanups, types: sent.map(frame => frame.type) }));
      if (['funding', 'untrusted'].includes(mode)) assert.equal(result.starts, 0);
      else {
        assert.equal(result.starts, 1); assert.equal(result.cleanups, 1);
        assert.equal(result.types.filter(t => t === 'voice_commit').length, ['pointer', 'keyboard', 'uncertain'].includes(mode) ? 1 : 0);
      }
      assert.deepEqual(errors, []);
      console.log(`PASS voice control ${mode}`); await page.close();
    }
  } finally { if (browser) await browser.close(); await new Promise(resolve => server.close(resolve)); }
})().catch(error => { console.error(error); process.exitCode = 1; });
