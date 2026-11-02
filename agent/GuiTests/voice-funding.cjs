// Real GUI host funding checks. Synthetic keys only; no inference or audio requests.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const http = require('node:http');
const { spawn } = require('node:child_process');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const sdk = '/Users/bill/.local/share/ensemble-dotnet10';
const binary = process.env.GUI_TEST_BINARY || '/private/tmp/cadence-v3-promotion/artifacts/bin/Gui/debug/Gui.dll';
const delay = ms => new Promise(resolve => setTimeout(resolve, ms));
async function bounded(promise, name) {
  let timer;
  try { return await Promise.race([promise, new Promise((_, reject) => {
    timer = setTimeout(() => reject(new Error(name + ' timed out')), 15000);
  })]); } finally { clearTimeout(timer); }
}
(async () => {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'ensemble-voice-funding-'));
  fs.chmodSync(root, 0o700);
  const skills = path.join(root, 'skills');
  fs.mkdirSync(path.join(skills, 'fixture'), { recursive: true });
  fs.writeFileSync(path.join(skills, 'fixture', 'SKILL.md'),
    '---\nname: fixture\ntype: primary\ndescription: Local funding test\ntools: \n---\nNo inference is requested.\n');
  let providerCalls = 0;
  const provider = http.createServer((request, response) => { providerCalls++; response.writeHead(500); response.end(); });
  await new Promise(resolve => provider.listen(0, '127.0.0.1', resolve));
  const providerUrl = `http://127.0.0.1:${provider.address().port}`;
  const browser = await chromium.launch({ headless: true, channel: 'chrome' });
  try {
    const keyFiles = new Map();
    function keyFile(name, text) {
      const file = path.join(root, name + '.json');
      fs.writeFileSync(file, text, { mode: 0o600 }); keyFiles.set(file, text);
      return file;
    }
    const valid = keyFile('valid', JSON.stringify({ openai_api_key: 'synthetic-file-key', other_setting: true }));
    const duplicate = keyFile('duplicate', '{"openai_api_key":"synthetic-file-key","openai_api_key":"other"}');
    const missing = keyFile('missing', '{"anthropic_api_key":"synthetic-other-provider"}');
    const malformed = keyFile('malformed', '{');
    const wrongType = keyFile('wrong-type', '{"openai_api_key":42}');
    const oversized = keyFile('oversized', ' '.repeat(1024 * 1024 + 1));
    const cases = [
      { name: 'default-off-with-ambient-speech-key', env: { EN_TRANSCRIPTION_API_KEY: 'synthetic-speech-key' }, enabled: false },
      { name: 'explicit-enable', args: ['--transcription', 'openai-api-key'], env: { EN_TRANSCRIPTION_API_KEY: 'synthetic-speech-key' }, enabled: true },
      { name: 'explicit-off-wins', args: ['--transcription', 'off'], env: { EN_TRANSCRIPTION: 'openai-api-key', EN_TRANSCRIPTION_API_KEY: 'synthetic-speech-key' }, enabled: false },
      { name: 'missing-speech-key-no-model-key-fallback', args: ['--transcription', 'openai-api-key'], refusal: true },
      { name: 'invalid-payer', args: ['--transcription', 'subscription'], refusal: true },
      { name: 'explicit-key-file', args: ['--transcription', 'openai-api-key', '--transcription-key-file', valid], env: { EN_TRANSCRIPTION_API_KEY: 'invalid\nfallback' }, enabled: true },
      { name: 'environment-key-file', env: { EN_TRANSCRIPTION: 'openai-api-key', EN_TRANSCRIPTION_KEY_FILE: valid }, enabled: true },
      { name: 'disabled-file-not-read', args: ['--transcription', 'off', '--transcription-key-file', path.join(root, 'absent.json')], enabled: false },
      ...[duplicate, missing, malformed, wrongType, oversized].map((file, index) => ({
        name: 'invalid-file-' + index, args: ['--transcription', 'openai-api-key', '--transcription-key-file', file],
        env: { EN_TRANSCRIPTION_API_KEY: 'synthetic-speech-key' }, refusal: true
      }))
    ];
    for (const test of cases) {
      const home = path.join(root, test.name); fs.mkdirSync(home);
      const reservation = http.createServer();
      await new Promise(resolve => reservation.listen(0, '127.0.0.1', resolve));
      const port = reservation.address().port;
      await new Promise(resolve => reservation.close(resolve));
      let output = '';
      const child = spawn(path.join(sdk, 'dotnet'), [binary, '--port', String(port), '--vendor', 'openai',
        '--model', 'gpt-6-astra', '--base-url', providerUrl, '--surface', 'responses', '--no-save',
        '--data-dir', path.join(home, 'data'), '--agent-name', 'Fixture', '--skills-dir', skills,
        '--primary-skill', 'fixture', '--workspace', home, ...(test.args || [])], {
        cwd: home, env: { HOME: home, TMPDIR: home, DOTNET_ROOT: sdk, DOTNET_NOLOGO: '1',
          PATH: `${sdk}:/usr/bin:/bin`, LLM_API_KEY: 'synthetic-model-key', ...test.env }
      });
      child.stdout.on('data', data => output += data); child.stderr.on('data', data => output += data);
      const exited = new Promise((resolve, reject) => { child.once('exit', (code, signal) => resolve({ code, signal })); child.once('error', reject); });
      let page;
      try {
        if (test.refusal) {
          const result = await bounded(exited, test.name);
          assert.equal(result.code, 1, test.name + ': expected explicit refusal');
        } else {
          const url = `http://127.0.0.1:${port}`;
          await bounded((async () => {
            while (true) {
              if (child.exitCode !== null) throw new Error(test.name + ': host exited: ' + output);
              try { if ((await fetch(url)).ok) break; } catch { }
              await delay(30);
            }
          })(), test.name + ' readiness');
          page = await browser.newPage();
          await page.addInitScript(() => {
            window.connectionFrame = null;
            const Native = WebSocket;
            window.WebSocket = class extends Native {
              constructor(...args) { super(...args); this.addEventListener('message', event => {
                const frame = JSON.parse(event.data);
                if (frame.type === 'connection') window.connectionFrame = frame;
              }); }
            };
          });
          await page.goto(url);
          await page.waitForFunction(() => window.connectionFrame !== null, null, { timeout: 10000 });
          const frame = await page.evaluate(() => window.connectionFrame);
          assert.deepEqual(frame.transcription, { enabled: test.enabled,
            payer: test.enabled ? 'openai-api-key' : 'off', model: test.enabled ? 'gpt-transcribe' : null });
          assert.ok(!JSON.stringify(frame).includes('synthetic-'), 'Connection frame leaked key');
          await page.close(); page = null;
          child.kill('SIGTERM');
          const result = await bounded(exited, test.name + ' full shutdown');
          assert.equal(result.code, 0, test.name + ': nonzero shutdown');
        }
        assert.ok(!output.includes('synthetic-speech-key') && !output.includes('synthetic-model-key') && !output.includes('synthetic-file-key'), 'Console leaked key');
        console.log('PASS ' + test.name);
      } finally {
        if (page) await page.close();
        if (child.exitCode === null && child.signalCode === null) {
          child.kill('SIGTERM');
          try { await bounded(exited, 'cleanup'); }
          catch (error) { child.kill('SIGKILL'); await exited; throw error; }
        }
      }
    }
    for (const [file, text] of keyFiles) assert.equal(fs.readFileSync(file, 'utf8'), text, 'Key file modified');
    assert.equal(providerCalls, 0, 'Startup contacted model provider');
    console.log('PASS zero provider requests; all host processes joined');
  } finally {
    await browser.close(); await new Promise(resolve => provider.close(resolve));
    console.log('Evidence home: ' + root);
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
