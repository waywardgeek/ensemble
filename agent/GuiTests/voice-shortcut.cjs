const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const http = require('http');
const fs = require('fs');
const path = require('path');
const assert = require('assert/strict');
const root = path.resolve(__dirname, '../gui/wwwroot');
const server = http.createServer((req, res) => {
  const name = req.url.slice(1);
  if (!name) { res.setHeader('Content-Type', 'text/html'); res.end('<button id="mic">Hold</button><button id="cancel" hidden>Cancel</button><div id="status"></div><input id="other"><dialog id="dialog">Settings</dialog>'); return; }
  if (!/^voice-[a-z]+\.js$/.test(name)) { res.writeHead(404); res.end(); return; }
  res.setHeader('Content-Type', 'text/javascript'); res.end(fs.readFileSync(path.join(root, name)));
});
(async () => {
  let browser;
  try {
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    browser = await chromium.launch({ headless: true, channel: 'chrome' });
    for (const mode of ['f8', 'right-alt', 'off', 'apple', 'altgr', 'ctrl', 'shift', 'meta', 'f8-alt', 'untrusted', 'dialog', 'chord', 'setting', 'blur', 'composition', 'focus', 'draft', 'funding']) {
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
      await page.evaluate(mode => {
        Object.defineProperty(navigator, 'platform', { value: mode === 'apple' ? 'MacIntel' : 'Win32', configurable: true });
        if (mode === 'altgr') {
          const original = KeyboardEvent.prototype.getModifierState;
          KeyboardEvent.prototype.getModifierState = function(key) { return key === 'AltGraph' || original.call(this, key); };
        }
        control.setShortcut(mode === 'off' ? 'off' : ['right-alt', 'apple', 'altgr'].includes(mode) ? 'AltRight' : 'F8');
        if (mode === 'dialog') document.querySelector('#dialog').showModal();
        if (mode === 'draft') control.bind(connection, '', true, true);
      }, mode);
      const key = ['right-alt', 'apple', 'altgr'].includes(mode) ? 'AltRight' : 'F8';
      const modifier = { ctrl: 'Control', shift: 'Shift', meta: 'Meta', 'f8-alt': 'Alt' }[mode];
      if (modifier) await page.keyboard.down(modifier);
      if (mode === 'untrusted') await page.evaluate(() => window.dispatchEvent(new KeyboardEvent('keydown', { code: 'F8' })));
      else { await page.keyboard.down(key); await page.keyboard.down(key); }
      const running = ['f8', 'right-alt', 'chord', 'setting', 'blur', 'composition', 'focus'].includes(mode);
      if (running) await page.waitForFunction(() => starts === 1);
      if (mode === 'chord') { await page.keyboard.down('a'); await page.keyboard.up('a'); }
      if (mode === 'setting') await page.evaluate(() => control.setShortcut('off'));
      if (mode === 'blur') await page.evaluate(() => window.dispatchEvent(new Event('blur')));
      if (mode === 'composition') await page.evaluate(() => document.dispatchEvent(new CompositionEvent('compositionstart')));
      if (mode === 'focus') await page.locator('#other').focus();
      await page.keyboard.up(key);
      if (modifier) await page.keyboard.up(modifier);
      await page.waitForFunction(() => !control.blocked);
      const result = await page.evaluate(() => ({ starts, cleanups, types: sent.map(frame => frame.type) }));
      assert.equal(result.starts, running ? 1 : 0, mode + ' capture admission');
      assert.equal(result.cleanups, running ? 1 : 0, mode + ' cleanup');
      assert.equal(result.types.filter(t => t === 'voice_commit').length, ['f8', 'right-alt'].includes(mode) ? 1 : 0, mode + ' commit');
      assert.deepEqual(errors, []);
      console.log(`PASS voice shortcut ${mode}`); await page.close();
    }
  } finally { if (browser) await browser.close(); await new Promise(resolve => server.close(resolve)); }
})().catch(error => { console.error(error); process.exitCode = 1; });
