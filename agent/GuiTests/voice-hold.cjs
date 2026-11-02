// Real Chrome module execution with gated synthetic capture. No microphone use.
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const root = path.resolve(__dirname, '../gui/wwwroot');
const server = http.createServer((req, res) => {
  const name = new URL(req.url, 'http://localhost').pathname.slice(1);
  if (!name) { res.end('<title>Owned voice hold test</title>'); return; }
  if (!['voice-hold.js', 'voice-recorder.js', 'voice-pcm.js', 'voice-worklet.js'].includes(name)) {
    res.writeHead(404); res.end(); return;
  }
  res.setHeader('Content-Type', 'text/javascript');
  res.end(fs.readFileSync(path.join(root, name)));
});
(async () => {
  let browser;
  try {
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    browser = await chromium.launch({ headless: true, channel: 'chrome',
      args: ['--autoplay-policy=no-user-gesture-required', '--mute-audio'] });
    const page = await browser.newPage();
    await page.goto(`http://127.0.0.1:${server.address().port}/`);
    const errors = [];
    page.on('pageerror', e => errors.push(e.message));
    const results = await page.evaluate(async () => {
      const { VoiceHold } = await import('./voice-hold.js');
      const check = (ok, message) => { if (!ok) throw new Error(message); };
      const gate = () => {
        let resolve, reject;
        const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
        return { promise, resolve, reject };
      };
      const bounded = async promise => {
        let timer;
        try {
          return await Promise.race([promise, new Promise((_, reject) => {
            timer = setTimeout(() => reject(new Error('Hold operation failed to join within 2 seconds')), 2000);
          })]);
        } finally { clearTimeout(timer); }
      };
      const results = [];
      function fixture() {
        const permission = gate(), preparation = gate(), finish = gate();
        const tracks = [{ stops: 0, stop() { this.stops++; } }];
        const stream = { getTracks: () => tracks };
        const recorder = new EventTarget();
        Object.assign(recorder, {
          starts: 0, finishes: 0, cancels: 0,
          start() { this.starts++; return preparation.promise; },
          finish() { this.finishes++; return finish.promise; },
          cancel() { this.cancels++; preparation.resolve(); finish.resolve(new Uint8Array([1])); return Promise.resolve(); }
        });
        let requests = 0, constructions = 0;
        const hold = new VoiceHold({
          requestStream: () => { requests++; return permission.promise; },
          createRecorder: () => { constructions++; return recorder; }
        });
        return { hold, permission, preparation, finish, tracks, stream, recorder,
          counts: () => ({ requests, constructions }),
          ready: async () => { const start = hold.start(); permission.resolve(stream); preparation.resolve(); check(await bounded(start), 'Not ready'); }
        };
      }
      {
        const f = fixture(); const start = f.hold.start(); const release = f.hold.release();
        f.permission.resolve(f.stream);
        check(await bounded(start) === false, 'Late permission started recording');
        check(await bounded(release) === null, 'Early release returned audio');
        check(f.tracks[0].stops === 1 && f.counts().constructions === 0, 'Late stream not stopped');
        results.push('release during permission joins and stops late stream');
      }
      {
        const f = fixture(); const start = f.hold.start(); const cancel = f.hold.cancel();
        check(cancel === f.hold.cancel(), 'Repeated cancel changed ownership');
        f.permission.resolve(f.stream); await bounded(cancel); await bounded(start);
        check(f.tracks[0].stops === 1 && f.counts().constructions === 0, 'Cancelled permission leaked');
        results.push('cancel during permission is idempotent');
      }
      {
        const f = fixture(); const start = f.hold.start(); f.permission.resolve(f.stream);
        await Promise.resolve(); const release = f.hold.release();
        await bounded(release); await bounded(start);
        check(f.recorder.cancels > 0 && f.recorder.finishes === 0, 'Early release finished capture');
        results.push('release during recorder preparation cancels');
      }
      {
        const f = fixture(); await f.ready(); const release = f.hold.release();
        check(release === f.hold.release(), 'Duplicate release changed result promise');
        const wave = new Uint8Array([2, 3]); f.finish.resolve(wave);
        check(await bounded(release) === wave && f.recorder.finishes === 1, 'Duplicate or lost finish');
        results.push('normal release returns owned audio once');
      }
      {
        const f = fixture(); await f.ready(); const release = f.hold.release();
        await bounded(f.hold.cancel()); check(await bounded(release) === null, 'Cancelled finish returned late audio');
        results.push('cancel during finish suppresses late audio and joins');
      }
      {
        const f = fixture(); const start = f.hold.start(); f.permission.reject(new Error('Permission denied'));
        let failed = false; try { await bounded(start); } catch { failed = true; }
        check(failed && f.hold.state === 'failed', 'Permission refusal hidden');
        results.push('permission refusal is explicit');
      }
      {
        const f = fixture(); await f.ready(); f.recorder.cancel = () => Promise.reject(new Error('Cleanup failed'));
        let failed = false; try { await bounded(f.hold.cancel()); } catch (e) { failed = e instanceof AggregateError; }
        check(failed, 'Cleanup failure swallowed');
        results.push('cleanup failure remains observable');
      }
      {
        const f = fixture(); await f.ready(); f.recorder.dispatchEvent(new Event('cancelled'));
        check(await bounded(f.hold.release()) === null, 'Device cancellation returned audio');
        results.push('device cancellation prevents release');
      }
      {
        const { PcmVoiceRecorder } = await import('./voice-recorder.js');
        const generator = new AudioContext({ sampleRate: 16000 });
        const destination = generator.createMediaStreamDestination();
        const oscillator = generator.createOscillator();
        oscillator.connect(destination); oscillator.start();
        let recorder;
        const hold = new VoiceHold({
          requestStream: async () => destination.stream,
          createRecorder: stream => { recorder = new PcmVoiceRecorder(stream); return recorder; }
        });
        try {
          await generator.resume();
          check(await bounded(hold.start()), 'Native recorder did not start');
          await bounded((async () => {
            while (recorder.sampleCount < 512) await new Promise(resolve => setTimeout(resolve, 10));
          })());
          const wave = await bounded(hold.release());
          check(wave instanceof Uint8Array && wave.byteLength > 44, 'Native recorder returned no WAV');
          check(new TextDecoder().decode(wave.slice(0, 4)) === 'RIFF', 'Native output is not WAV');
          check(recorder.state === 'closed' && destination.stream.getTracks().every(t => t.readyState === 'ended'),
            'Native recorder resources not closed before release');
          results.push('real PCM recorder returns WAV and closes devices through hold');
        } finally {
          await hold.cancel(); oscillator.stop(); await generator.close();
        }
      }
      return results;
    });
    assert.deepEqual(errors, []);
    for (const result of results) console.log('PASS', result);
    console.log(`${results.length} voice hold groups passed; no microphone requested.`);
  } finally {
    if (browser) await browser.close();
    await new Promise(resolve => server.close(resolve));
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
