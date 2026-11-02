// Isolated native Chrome module tests. No microphone, provider or live GUI.
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const root = path.resolve(__dirname, '../gui/wwwroot');
const server = http.createServer((req, res) => {
  const name = new URL(req.url, 'http://localhost').pathname.slice(1);
  if (!name) { res.end('<title>Owned voice session test</title>'); return; }
  if (!['voice-session.js', 'voice-hold.js', 'voice-recorder.js', 'voice-pcm.js'].includes(name)) {
    res.writeHead(404); res.end(); return;
  }
  res.setHeader('Content-Type', 'text/javascript');
  res.end(fs.readFileSync(path.join(root, name)));
});
(async () => {
  let browser;
  try {
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    browser = await chromium.launch({ headless: true, channel: 'chrome' });
    const page = await browser.newPage();
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto(`http://127.0.0.1:${server.address().port}/`);
    const results = await page.evaluate(async () => {
      const { VoiceSession } = await import('./voice-session.js');
      const { VoiceHold } = await import('./voice-hold.js');
      const check = (ok, message) => { if (!ok) throw new Error(message); };
      const gate = () => {
        let resolve;
        const promise = new Promise(yes => { resolve = yes; });
        return { promise, resolve };
      };
      const bounded = async promise => {
        let timer;
        try { return await Promise.race([promise, new Promise((_, reject) => {
          timer = setTimeout(() => reject(new Error('Session operation did not join')), 2000);
        })]); } finally { clearTimeout(timer); }
      };
      function fixture(hold) {
        const f = { current: true, sent: [], fail: null, stopped: 0, finishes: 0 };
        f.hold = hold ?? new class extends EventTarget {
          async start() { return true; }
          async release() { f.finishes++; return new Uint8Array(46).fill(7); }
          async cancel() { f.stopped++; }
        }();
        f.socket = { readyState: 1, send(text) {
          const message = JSON.parse(text);
          f.sent.push(message);
          if (f.fail === message.type) throw new Error('Simulated send failure');
        } };
        f.session = new VoiceSession({ socket: f.socket, agent: '', sequence: 1,
          isCurrent: () => f.current, hold: f.hold });
        f.reply = (type, fields = {}, socket = f.socket) => f.session.receive(socket,
          { type, agent_id: '', sequence: 1, ...fields });
        f.count = type => f.sent.filter(message => message.type === type).length;
        f.audio = async () => {
          await f.session.start(); f.reply('voice_begun'); await f.session.release();
          check(f.count('voice_audio') === 1, 'Audio not sent once');
        };
        return f;
      }
      const results = [];
      {
        const f = fixture(); await f.audio();
        f.reply('voice_begun'); f.reply('voice_ready'); f.reply('voice_ready');
        await f.session.release();
        check(f.count('voice_commit') === 1 && f.finishes === 1, 'Duplicate release/ready sent twice');
        check(f.sent.every(m => m.agent_id === '' && m.sequence === 1), 'Original root binding lost');
        const audio = f.sent.find(m => m.type === 'voice_audio').audio;
        check(audio.mime === 'audio/wav' && atob(audio.data).length === 46, 'Actual audio bytes missing');
        f.reply('voice_accepted'); await f.session.cancel();
        check(f.session.state === 'accepted' && f.stopped === 1, 'Accepted capture did not join');
        results.push('original root, audio bytes, automatic once-only commit and acceptance');
      }
      {
        const f = fixture(); await f.session.start(); await f.session.release();
        check(f.count('voice_audio') === 0, 'Audio sent before begin acknowledgement');
        f.reply('voice_ready'); check(f.count('voice_commit') === 0, 'Premature ready committed');
        f.reply('voice_begun'); f.reply('voice_begun');
        check(f.count('voice_audio') === 1, 'Late begin failed or duplicated audio');
        await f.session.cancel(); results.push('release before begin reply');
      }
      {
        const f = fixture(); await f.audio();
        f.reply('voice_ready', {}, {}); f.reply('voice_ready', { agent_id: 'other' });
        f.reply('voice_ready', { sequence: 2 });
        f.reply('error', { voice_sequence: 2 });
        check(f.session.state === 'transcribing' && !f.count('voice_commit'), 'Misdirected reply changed state');
        f.reply('error', { voice_sequence: 1 }); await f.session.cancel();
        check(f.session.state === 'failed', 'Correlated refusal ignored');
        results.push('socket, Agent, sequence and error correlation');
      }
      {
        const f = fixture(); await f.audio(); f.current = false;
        f.reply('voice_ready'); f.current = true; f.reply('voice_ready');
        await f.session.cancel();
        check(!f.count('voice_commit') && f.session.state === 'cancelled', 'Switch-away/back sent stale speech');
        results.push('stale selection never commits');
      }
      {
        const f = fixture(); const finishing = gate();
        f.hold.release = () => finishing.promise;
        await f.session.start(); f.reply('voice_begun'); const release = f.session.release();
        await f.session.cancel(); finishing.resolve(new Uint8Array(46)); await bounded(release);
        f.reply('voice_ready');
        check(!f.count('voice_audio') && !f.count('voice_commit'), 'Late finish sent cancelled capture');
        results.push('cancel during finish invalidates late audio');
      }
      {
        const f = fixture(); await f.audio(); f.socket.readyState = 3;
        await f.session.cancel(); f.reply('voice_ready');
        check(f.session.state === 'cancelled' && !f.count('voice_commit'), 'Disconnect before commit sent');
        results.push('disconnect before commit');
      }
      {
        for (const failure of ['disconnect', 'throw', 'error', 'cancel']) {
          const f = fixture(); await f.audio();
          if (failure === 'throw') f.fail = 'voice_commit';
          f.reply('voice_ready');
          if (failure === 'disconnect') f.socket.readyState = 3;
          if (failure === 'error') f.reply('error', { voice_sequence: 1 });
          await f.session.cancel(); f.reply('voice_ready'); f.reply('voice_accepted');
          check(f.session.state === 'uncertain' && f.count('voice_commit') === 1,
            `Commit outcome was retried or falsely cancelled: ${failure}`);
        }
        results.push('uncertain commit never retries or claims retraction');
      }
      {
        for (const failure of ['permission', 'finish', 'limit', 'device']) {
          const f = fixture();
          if (failure === 'permission') f.hold.start = async () => { throw new Error('denied'); };
          if (failure === 'finish') f.hold.release = async () => { throw new Error('failed'); };
          if (failure === 'limit') f.hold.release = async () => new Uint8Array(1920045);
          await f.session.start().catch(() => {}); f.reply('voice_begun');
          if (failure === 'device') f.hold.dispatchEvent(new Event('cancelled'));
          await f.session.release().catch(() => {}); await f.session.cancel();
          check(f.session.terminal && !f.count('voice_audio') && !f.count('voice_commit'), `${failure} sent partial speech`);
        }
        results.push('permission, finish, limit and device failures send nothing');
      }
      {
        const permission = gate(); let tracksStopped = 0, created = 0;
        const hold = new VoiceHold({ requestStream: () => permission.promise,
          createRecorder: () => { created++; throw new Error('Must not start after release'); } });
        const f = fixture(hold); const starting = f.session.start();
        const release = f.session.release();
        permission.resolve({ getTracks: () => [{ stop: () => tracksStopped++ }] });
        await bounded(Promise.all([starting, release, f.session.cancel()]));
        check(tracksStopped === 1 && created === 0 && !f.count('voice_audio'), 'Late permission acquired capture');
        results.push('real hold early release joins late permission and stops track');
      }
      {
        const f = fixture(); const cleanup = gate();
        f.hold.cancel = () => cleanup.promise;
        await f.audio(); let joined = false;
        const cancelling = f.session.cancel().then(() => { joined = true; });
        await Promise.resolve(); check(!joined && f.session.state === 'cancelled', 'Cancel did not invalidate before joining');
        f.reply('voice_ready'); check(!f.count('voice_commit'), 'Cleanup wait allowed stale commit');
        cleanup.resolve(); await bounded(cancelling);
        results.push('cleanup joins without delaying invalidation');
      }
      {
        const f = fixture();
        f.session.addEventListener('change', () => {
          if (f.session.state === 'committing') f.session.cancel();
        });
        await f.audio(); f.reply('voice_ready'); await f.session.cancel();
        check(!f.count('voice_commit'), 'Invalidation during notification allowed a late commit');
        results.push('state notification invalidation precedes effects');
      }
      {
        const f = fixture();
        f.hold.cancel = async () => { throw new Error('cleanup failed'); };
        await f.audio(); let refused = false;
        try { await f.session.cancel(); } catch { refused = true; }
        f.reply('voice_ready');
        check(refused && f.session.state === 'cancelled' && !f.count('voice_commit'),
          'Cleanup error was hidden or restored send authority');
        results.push('cleanup failure remains observable after invalidation');
      }
      return results;
    });
    assert.deepEqual(errors, []);
    for (const result of results) console.log('PASS', result);
    console.log(`${results.length} voice session groups passed; no microphone/provider use.`);
  } finally {
    if (browser) await browser.close();
    await new Promise(resolve => server.close(resolve));
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
