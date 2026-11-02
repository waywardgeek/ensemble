// Real page + PCM recorder; synthetic audio and explicitly simulated voice replies.
// No microphone permission, paid transcription, or actor voice admission is claimed.
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const assert = require('node:assert/strict');
const base = process.env.GUI_URL || 'http://127.0.0.1:18202/assistant/';
assert.equal(new URL(base).hostname, '127.0.0.1');
assert.ok(!['8088', '8089'].includes(new URL(base).port));
(async () => {
  const browser = await chromium.launch({headless:true, channel:'chrome'});
  const page = await browser.newPage(); page.setDefaultTimeout(15000);
  const errors = []; page.on('pageerror', e => errors.push(e.message));
  try {
    await page.addInitScript(() => {
      window.voiceFrames = []; window.testAudio = []; window.ackVoice = true;
      navigator.mediaDevices.getUserMedia = async () => {
        const context = new AudioContext(); const oscillator = context.createOscillator();
        const destination = context.createMediaStreamDestination();
        oscillator.connect(destination); oscillator.start(); await context.resume();
        testAudio.push({context, oscillator, stream:destination.stream});
        return destination.stream;
      };
      const Native = WebSocket;
      window.WebSocket = class extends Native {
        constructor(...args) {
          super(...args);
          this.addEventListener('message', event => {
            const frame = JSON.parse(event.data);
            if (frame.type === 'connection') {
              frame.transcription = {enabled:true, payer:'openai-api-key'};
              Object.defineProperty(event, 'data', {value:JSON.stringify(frame)});
            }
          });
        }
        send(text) {
          const frame = JSON.parse(text);
          if (!frame.type?.startsWith('voice_')) return super.send(text);
          voiceFrames.push(frame);
          const type = {voice_begin:'voice_begun', voice_audio:'voice_ready',
            voice_commit: ackVoice ? 'voice_accepted' : null, voice_cancel:'voice_cancelled'}[frame.type];
          if (type) queueMicrotask(() => this.dispatchEvent(new MessageEvent('message', {
            data: JSON.stringify({type, agent_id:frame.agent_id, sequence:frame.sequence})
          })));
        }
      };
    });
    await page.goto(base);
    await page.waitForFunction(() => window.voiceFrames && !document.querySelector('#voice-hold').disabled);
    await page.evaluate(() => {
      document.querySelector('#narrate').checked = false;
      document.querySelector('#narrate').dispatchEvent(new Event('change', {bubbles:true}));
    });
    const hold = async () => {
      const box = await page.locator('#voice-hold').boundingBox();
      await page.mouse.move(box.x + box.width/2, box.y + box.height/2); await page.mouse.down();
      await page.waitForFunction(() => document.querySelector('#voice-status').textContent.startsWith('Recording.'));
      await page.waitForTimeout(250);
    };
    await hold(); await page.mouse.up();
    await page.waitForFunction(() => document.querySelector('#voice-status').textContent === 'Voice message sent.' && !document.querySelector('#voice-hold').disabled);
    const first = await page.evaluate(() => ({frames:voiceFrames, ended:testAudio.every(a => a.stream.getTracks().every(t=>t.readyState==='ended'))}));
    assert.deepEqual(first.frames.map(f=>f.type), ['voice_begin','voice_audio','voice_commit']);
    assert.ok(first.ended);
    const wav = Buffer.from(first.frames[1].audio.data,'base64');
    assert.equal(wav.toString('ascii',0,4),'RIFF'); assert.equal(wav.readUInt32LE(24),16000);
    assert.ok(wav.length > 44); assert.equal(wav.readUInt32LE(40),wav.length-44);
    assert.equal(first.frames[0].agent_id,first.frames[2].agent_id);
    console.log('PASS full page: native PCM and automatic single commit to original Agent');

    await page.locator('#prompt').fill('Keep this typed draft');
    assert.ok(await page.locator('#voice-hold').isDisabled());
    assert.equal(await page.locator('#prompt').inputValue(),'Keep this typed draft');
    await page.locator('#prompt').fill('');
    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    console.log('PASS full page: existing text draft blocks voice without mutation');

    await hold(); await page.evaluate(() => window.dispatchEvent(new Event('blur'))); await page.mouse.up();
    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    assert.equal(await page.evaluate(() => voiceFrames.filter(f=>f.type==='voice_commit').length),1);
    assert.equal(await page.evaluate(() => voiceFrames.filter(f=>f.type==='voice_audio').length),1);
    console.log('PASS full page: blur cancels recording without partial audio or commit');

    await page.evaluate(() => { ackVoice = false; });
    await hold(); await page.mouse.up();
    await page.waitForFunction(() => voiceFrames.filter(f=>f.type==='voice_commit').length===2);
    await page.locator('#voice-cancel').click();
    await page.waitForFunction(() => document.querySelector('#voice-cancel').textContent==='Dismiss uncertainty' && !document.querySelector('#voice-cancel').disabled);
    assert.ok(await page.locator('#send').isDisabled());
    assert.ok(await page.locator('#voice-hold').isDisabled());
    await page.locator('#voice-cancel').click();
    await page.waitForFunction(() => !document.querySelector('#voice-hold').disabled);
    assert.equal(await page.evaluate(() => voiceFrames.filter(f=>f.type==='voice_commit').length),2);
    assert.ok(await page.evaluate(() => testAudio.every(a=>a.stream.getTracks().every(t=>t.readyState==='ended'))));
    console.log('PASS full page: uncertain commit requires explicit dismissal and never retries');
    assert.deepEqual(errors,[]);
    const unfunded = await browser.newPage();
    try {
      await unfunded.goto(base);
      await unfunded.waitForFunction(() => document.querySelector('#voice-status').textContent.startsWith('Voice is off:'));
      assert.ok(await unfunded.locator('#voice-hold').isDisabled());
      console.log('PASS full page: unmodified server funding-off advertisement disables microphone');
    } finally { await unfunded.close(); }
  } finally {
    await page.evaluate(async () => {
      for (const a of window.testAudio || []) { a.oscillator.stop(); a.stream.getTracks().forEach(t=>t.stop()); await a.context.close(); }
    }).catch(()=>{});
    await browser.close();
  }
})().catch(e=>{console.error(e); process.exitCode=1;});
