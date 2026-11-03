// Real page, PCM recorder, WebSocket, lease, transcription client and Agent.
// The separately compiled grader consumer owns the fake HTTP oracles.
// This script uses browser APIs and public UI/wire observations, not app internals.
const {chromium} = require(process.env.PLAYWRIGHT_MODULE);
const assert = require('node:assert/strict');
const base = process.env.VOICE_GUI_URL;
const target = new URL(base);
assert.equal(target.hostname, '127.0.0.1', 'Only disposable loopback hosts');
assert.ok(target.port && !['8088', '8089', '8090'].includes(target.port), 'Never target live services');
(async () => {
  const browser = await chromium.launch({headless:true, channel:'chrome'});
  process.once('SIGINT', () => { void browser.close().finally(() => process.exit(130)); });
  const page = await browser.newPage(); page.setDefaultTimeout(15000);
  const errors = []; page.on('pageerror', e => errors.push(e.message));
  try {
    if (process.env.VOICE_GRADE_MUTATION) {
      const mutations = {
        'extra-send': ["this.#send('voice_commit');", "/* Incorrect: waits for a second human Send. */"],
        'duplicate-commit': ["this.#send('voice_commit');", "this.#send('voice_commit'); this.#send('voice_commit');"],
        'wrong-agent': ['agent_id: this.#agent,', 'agent_id: "VoiceRecipient",'],
        'leaked-tracks': ['attempt(() => track.stop());', '/* Incorrect: owned track left live. */']
      };
      const mutation = mutations[process.env.VOICE_GRADE_MUTATION];
      assert.ok(mutation, 'Unknown voice mutation');
      await page.route(process.env.VOICE_GRADE_MUTATION === 'leaked-tracks' ? '**/voice-recorder.js' : '**/voice-session.js', async route => {
        const response = await route.fetch();
        const source = await response.text();
        assert.ok(source.includes(mutation[0]), 'Voice mutation anchor missing');
        await route.fulfill({ response, body: source.replace(mutation[0], mutation[1]) });
      });
    }
    await page.addInitScript(mode => {
      window.voiceFrames = []; window.voiceReplies = []; window.testAudio = [];
      navigator.mediaDevices.getUserMedia = async () => {
        const context = new AudioContext(); const oscillator = context.createOscillator();
        const destination = context.createMediaStreamDestination();
        oscillator.connect(destination); oscillator.start(); await context.resume();
        testAudio.push({context, oscillator, stream:destination.stream});
        if (mode === 'permission-delay') return await new Promise(resolve => { window.voicePermissionResolve = () => resolve(destination.stream); });
        return destination.stream;
      };
      const NativeSocket = WebSocket;
      window.WebSocket = class extends NativeSocket {
        constructor(...args) {
          super(...args); window.voiceSocket = this;
          this.addEventListener('message', event => {
            const frame = JSON.parse(event.data);
            if (frame.type?.startsWith('voice_') || frame.type === 'error') voiceReplies.push(frame);
          });
        }
        send(text) {
          const frame = JSON.parse(text);
          if (frame.type?.startsWith('voice_')) voiceFrames.push({type:frame.type, agent_id:frame.agent_id, sequence:frame.sequence});
          return super.send(text);
        }
      };
    }, process.env.VOICE_GRADE_MODE);
    await page.goto(base);
    if (process.env.VOICE_GRADE_MODE === 'funding-off') {
      await page.waitForFunction(() => !document.querySelector('#send').disabled);
      assert.equal(await page.locator('#voice-hold').isDisabled(), true, 'Model funding enabled microphone');
      const recipient = await page.locator('#agent-tree [aria-selected="true"] > .tree-row > [data-agent]').getAttribute('data-agent');
      await page.evaluate(agent => voiceSocket.send(JSON.stringify({type:'voice_begin', agent_id:agent, sequence:1})), recipient);
      await page.waitForFunction(() => voiceReplies.some(frame => frame.type === 'error' && frame.voice_sequence === 1));
      assert.equal(await page.evaluate(() => testAudio.length), 0, 'Disabled voice requested audio');
      const result = await page.evaluate(async () => (await fetch('/fixture/status')).json());
      assert.equal(result.audioCalls, 0); assert.equal(result.modelRequests, 0); assert.equal(result.admitted, 0);
      assert.equal(result.failure, null); assert.deepEqual(errors, []);
      console.log('PASS funding off: disabled microphone, forged lease refused, no model-key fallback or admission');
      return;
    }
    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    if (process.env.VOICE_GRADE_MODE === 'transcription-error') {
      const record = async () => {
        const box = await page.locator('#voice-hold').boundingBox();
        await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
        await page.mouse.down();
        await page.waitForFunction(() => document.querySelector('#voice-status').textContent.startsWith('Recording.'));
        await page.waitForTimeout(250); await page.mouse.up();
      };
      await record();
      await page.waitForFunction(() => voiceReplies.some(frame => frame.type === 'voice_failed'));
      let result = await page.evaluate(async () => (await fetch('/fixture/status')).json());
      assert.equal(result.audioCalls, 1, 'Transcription failure retried automatically');
      assert.equal(result.modelRequests, 0); assert.equal(result.admitted, 0);
      assert.equal(await page.evaluate(() => voiceFrames.filter(frame => frame.type === 'voice_commit').length), 0);
      await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
      await record();
      await page.waitForFunction(() => document.querySelector('#voice-status').textContent === 'Voice message sent.');
      await page.waitForFunction(async () => (await (await fetch('/fixture/status')).json()).modelRequests === 1);
      result = await page.evaluate(async () => (await fetch('/fixture/status')).json());
      assert.equal(result.audioCalls, 2); assert.equal(result.admitted, 1); assert.equal(result.failure, null);
      assert.ok(await page.evaluate(() => testAudio.every(audio => audio.stream.getTracks().every(track => track.readyState === 'ended'))));
      assert.deepEqual(errors, []);
      console.log('PASS transcription failure: no retry/admission, new deliberate recording recovers');
      return;
    }
    if (['permission-delay', 'disconnect-recording', 'recipient-recording', 'recipient-transcription', 'disconnect-transcription'].includes(process.env.VOICE_GRADE_MODE)) {
      const box = await page.locator('#voice-hold').boundingBox();
      await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await page.mouse.down();
      if (process.env.VOICE_GRADE_MODE === 'permission-delay') {
        await page.waitForFunction(() => typeof window.voicePermissionResolve === 'function');
        await page.mouse.up();
        await page.evaluate(() => voicePermissionResolve());
      } else {
        await page.waitForFunction(() => document.querySelector('#voice-status').textContent.startsWith('Recording.'));
        const heldTranscription = process.env.VOICE_GRADE_MODE.endsWith('-transcription');
        if (heldTranscription) {
          await page.waitForTimeout(250);
          await page.mouse.up();
          await page.waitForFunction(async () => (await (await fetch('/fixture/status')).json()).audioCalls === 1);
        }
        if (process.env.VOICE_GRADE_MODE.startsWith('recipient-')) {
          // DOM click uses the real selection control without releasing the held pointer first.
          await page.locator('#agent-tree [data-agent]').filter({ hasText: 'VoiceRecipient' }).evaluate(button => button.click());
          await page.waitForFunction(() => document.querySelector('#agent-tree [aria-selected="true"] .tree-name')?.textContent === 'VoiceRecipient');
        } else {
          await page.evaluate(() => voiceSocket.close());
        }
        if (!heldTranscription) await page.mouse.up();
        if (heldTranscription) {
          await page.evaluate(() => fetch('/fixture/release', { method:'POST' }));
          await page.waitForFunction(async () => (await (await fetch('/fixture/status')).json()).audioCompleted === 1);
        }
      }
      await page.waitForFunction(() => testAudio.length > 0 && testAudio.every(audio => audio.stream.getTracks().every(track => track.readyState === 'ended')))
        .catch(error => { throw new Error('Cancelled voice tracks did not end', { cause: error }); });
      const expectedAudio = process.env.VOICE_GRADE_MODE.endsWith('-transcription') ? 1 : 0;
      assert.equal(await page.evaluate(() => voiceFrames.filter(frame => frame.type === 'voice_audio').length), expectedAudio);
      assert.equal(await page.evaluate(() => voiceFrames.filter(frame => frame.type === 'voice_commit').length), 0,
        'Cancelled capture committed a transcript');
      const result = await page.evaluate(async () => (await fetch('/fixture/status')).json());
      assert.equal(result.audioCalls, expectedAudio); assert.equal(result.modelRequests, 0); assert.equal(result.admitted, 0);
      assert.equal(result.otherAdmitted, 0, 'Transcript reached a different Agent');
      assert.equal(result.failure, null); assert.deepEqual(errors, []);
      console.log('PASS ' + process.env.VOICE_GRADE_MODE + ': tracks ended, no commit or admission');
      return;
    }
    await page.evaluate(() => {
      document.querySelector('#narrate').checked = false;
      document.querySelector('#narrate').dispatchEvent(new Event('change', {bubbles:true}));
    });
    const status = () => page.evaluate(async () => (await fetch('/fixture/status')).json());
    const hold = async () => {
      const box = await page.locator('#voice-hold').boundingBox();
      await page.mouse.move(box.x+box.width/2,box.y+box.height/2); await page.mouse.down();
      await page.waitForFunction(() => document.querySelector('#voice-status').textContent.startsWith('Recording.'));
      await page.waitForTimeout(250);
    };
    await page.evaluate(() => voiceSocket.addEventListener('message', event => {
      const frame = JSON.parse(event.data);
      if (frame.type === 'settings_updated') globalThis.shortcutSaved = frame.settings.voice_shortcut;
    }));
    await page.locator('#settings-open').click();
    await page.locator('#settings-display-tab').click();
    await page.locator('#voice-shortcut').selectOption('F8');
    await page.waitForFunction(() => globalThis.shortcutSaved === 'F8');
    await page.keyboard.down('F8'); await page.keyboard.up('F8');
    assert.equal(await page.evaluate(() => voiceFrames.length), 0, 'Settings dialog allowed capture');
    await page.reload();
    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    await page.locator('#settings-open').click();
    await page.locator('#settings-display-tab').click();
    assert.equal(await page.locator('#voice-shortcut').inputValue(), 'F8');
    await page.locator('#settings-close').click();
    console.log('PASS real shortcut settings save/reload and open-dialog refusal');
    const expectedRecipient = await page.locator('#agent-tree [aria-selected="true"] > .tree-row > [data-agent]').getAttribute('data-agent');
    await page.keyboard.down('F8'); await page.keyboard.down('F8');
    await page.waitForFunction(() => voiceFrames.some(frame => frame.type === 'voice_begin'));
    assert.equal(await page.evaluate(() => voiceFrames.find(frame => frame.type === 'voice_begin').agent_id),
      expectedRecipient, 'Voice session targeted wrong Agent');
    await page.waitForFunction(() => document.querySelector('#voice-status').textContent.startsWith('Recording.'));
    await page.waitForTimeout(250); await page.keyboard.up('F8');
    await page.waitForFunction(() => document.querySelector('#voice-status').textContent === 'Voice message sent.')
      .catch(error => { throw new Error('Release did not automatically send', { cause: error }); });
    await page.waitForFunction(async () => {
      const s = await (await fetch('/fixture/status')).json(); return s.modelRequests === 1 && s.admitted === 1;
    });
    assert.equal((await status()).failure,null);
    await page.waitForFunction(() => document.querySelector('#transcript').textContent.includes('VOICE-FIXTURE model received transcript'));
    let s = await status(); assert.equal(s.failure,null); assert.equal(s.audioCalls,1); assert.ok(s.waveBytes > 44);
    assert.deepEqual(await page.evaluate(() => voiceFrames.map(f=>f.type)), ['voice_begin','voice_audio','voice_commit'], 'Voice commit must be automatic and once-only');
    assert.equal(await page.evaluate(() => voiceReplies.filter(f=>f.type==='voice_accepted').length),1);
    console.log('PASS real browser PCM -> separately funded fake transcription -> once-only actor admission, model request and History');

    const beforeErrors = await page.evaluate(() => voiceReplies.filter(f=>f.type==='error').length);
    await page.evaluate(() => voiceSocket.send(JSON.stringify(voiceFrames.find(f=>f.type==='voice_commit'))));
    await page.waitForFunction(n => voiceReplies.filter(f=>f.type==='error').length > n, beforeErrors);
    s = await status(); assert.equal(s.audioCalls,1); assert.equal(s.modelRequests,1); assert.equal(s.admitted,1);
    console.log('PASS real server refuses duplicate commit without another transcription or admission');

    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    await hold(); await page.evaluate(() => window.dispatchEvent(new Event('blur'))); await page.mouse.up();
    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    s = await status(); assert.equal(s.audioCalls,1);
    assert.equal(await page.evaluate(() => voiceFrames.filter(f=>f.type==='voice_audio').length),1);
    console.log('PASS blur cancels real recording before upload or provider work');

    await hold(); await page.mouse.up();
    await page.waitForFunction(async () => (await (await fetch('/fixture/status')).json()).audioCalls === 2);
    await page.locator('#voice-cancel').click();
    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    await page.waitForFunction(() => voiceReplies.some(f=>f.type==='voice_cancelled' && f.sequence===3));
    await page.evaluate(() => fetch('/fixture/release', {method:'POST'}));
    await page.waitForFunction(async () => (await (await fetch('/fixture/status')).json()).audioCompleted === 2);
    // Also allow the notification continuation to process the deliberately late response.
    await page.waitForTimeout(400);
    s = await status(); assert.equal(s.failure,null); assert.equal(s.modelRequests,1); assert.equal(s.admitted,1);
    assert.equal(await page.evaluate(() => voiceReplies.filter(f=>f.type==='voice_accepted').length),1);
    assert.equal(await page.evaluate(() => voiceFrames.filter(f=>f.type==='voice_commit').length),2); // includes deliberate duplicate
    assert.ok(await page.evaluate(() => testAudio.every(a=>a.stream.getTracks().every(t=>t.readyState==='ended'))));
    assert.deepEqual(errors,[]);
    console.log('PASS cancellation while transcription is held suppresses late success; tracks ended and no misadmission');
    console.log(JSON.stringify(s));
  } finally {
    await page.evaluate(async () => { for (const a of window.testAudio || []) { a.oscillator.stop(); a.stream.getTracks().forEach(t=>t.stop()); await a.context.close(); } }).catch(()=>{});
    await browser.close();
  }
})().catch(error => {console.error(error);process.exitCode=1;});
