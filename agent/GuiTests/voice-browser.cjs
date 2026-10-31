// Owns a loopback server and isolated Chrome. Never requests a microphone.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || '../../edition-2/agent/gui/node_modules/playwright');
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const assets = path.resolve(__dirname, '../gui/wwwroot');
let mode = 'normal', held = [];
function serveModule(response, name) {
  let body = fs.readFileSync(path.join(assets, name), 'utf8');
  if (process.env.VOICE_MUTANT === 'rate' && name === 'voice-pcm.js') {
    const anchor = 'view.setUint32(24, VOICE_PCM.sampleRate, true)';
    assert.ok(body.includes(anchor)); body = body.replace(anchor, 'view.setUint32(24, 8000, true)');
  }
  if (process.env.VOICE_MUTANT === 'feedback' && name === 'voice-worklet.js') {
    const anchor = 'process(inputs) {';
    assert.ok(body.includes(anchor));
    body = body.replace(anchor, 'process(inputs, outputs) { if (inputs[0]?.[0]) outputs[0][0].set(inputs[0][0]);');
  }
  response.writeHead(200, { 'Content-Type': 'text/javascript', 'Cache-Control': 'no-store' });
  response.end(body);
}
const server = http.createServer((request, response) => {
  const pathname = new URL(request.url, 'http://127.0.0.1').pathname;
  if (!pathname.startsWith('/assistant/')) { response.writeHead(404); response.end(); return; }
  const name = pathname.slice('/assistant/'.length);
  if (name === '') { response.writeHead(200, {'Content-Type':'text/html'}); response.end('<title>Owned voice fixture</title>'); }
  else if (['voice-pcm.js', 'voice-worklet.js', 'voice-recorder.js'].includes(name)) {
    if (name === 'voice-worklet.js' && mode === 'fail') { response.writeHead(404); response.end(); }
    else if (name === 'voice-worklet.js' && mode === 'hold') held.push(response);
    else serveModule(response, name);
  } else { response.writeHead(404); response.end(); }
});
(async () => {
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  const port = server.address().port;
  assert.ok(![8088, 8089].includes(port));
  let browser;
  try {
    browser = await chromium.launch({ headless: true, channel: 'chrome',
      args: ['--autoplay-policy=no-user-gesture-required', '--mute-audio'] });
    const page = await browser.newPage(); page.setDefaultTimeout(10000);
    const errors = []; page.on('pageerror', error => errors.push(error.message));
    await page.goto(`http://127.0.0.1:${port}/assistant/`);
    await page.evaluate(async () => {
      window.microphoneCalls = 0;
      navigator.mediaDevices.getUserMedia = async () => { ++microphoneCalls; throw new Error('Microphone access forbidden in fixture'); };
      window.contexts = []; window.nodes = [];
      const Context = AudioContext, Node = AudioWorkletNode;
      window.AudioContext = class extends Context { constructor(...args) { super(...args); contexts.push(this); } };
      window.AudioWorkletNode = class extends Node { constructor(...args) { super(...args); nodes.push(this); } };
      window.voice = await import('./voice-recorder.js');
      window.pcm = await import('./voice-pcm.js');
      window.make = async options => {
        const generator = new AudioContext({sampleRate:16000});
        const destination = generator.createMediaStreamDestination();
        const oscillator = generator.createOscillator(), gain = generator.createGain();
        oscillator.frequency.value = 440; gain.gain.value = 0.25;
        oscillator.connect(gain).connect(destination); oscillator.start(); await generator.resume();
        const recorder = new voice.PcmVoiceRecorder(destination.stream, options);
        window.current = { generator, oscillator, recorder, stream: destination.stream, cancelled: [] };
        recorder.addEventListener('cancelled', event => current.cancelled.push(event.detail));
      };
      window.clean = async () => {
        if (window.current) {
          await current.recorder.cancel(); current.oscillator.stop(); await current.generator.close();
        }
      };
    });
    await page.evaluate(async () => { await make(); await current.recorder.start(); });
    await page.waitForFunction(() => current.recorder.sampleCount >= 4800);
    const result = await page.evaluate(async () => {
      const monitor = contexts.at(-1).createAnalyser(); monitor.fftSize = 128;
      nodes.at(-1).connect(monitor);
      // Pull a fresh analysis window, not the analyser's initially empty buffer.
      await new Promise(resolve => setTimeout(resolve, 100));
      const output = new Float32Array(128); monitor.getFloatTimeDomainData(output); monitor.disconnect();
      const one = current.recorder.finish(), two = current.recorder.finish();
      const samePromise = one === two; const bytes = await one;
      return {bytes:Array.from(bytes), samePromise, silentOutput:output.every(x => x === 0),
        state:current.recorder.state, ended:current.stream.getTracks().every(t => t.readyState === 'ended'),
        context:contexts.at(-1).state};
    });
    assert.equal(result.samePromise, true); assert.equal(result.silentOutput, true, 'Recorder leaked samples to audio output');
    assert.equal(result.state, 'closed'); assert.equal(result.ended, true); assert.equal(result.context, 'closed');
    const wave = Buffer.from(result.bytes);
    assert.equal(wave.subarray(0,4).toString(), 'RIFF'); assert.equal(wave.readUInt32LE(4), wave.length-8);
    assert.equal(wave.subarray(8,16).toString(), 'WAVEfmt '); assert.equal(wave.readUInt32LE(16),16);
    assert.equal(wave.readUInt16LE(20),1); assert.equal(wave.readUInt16LE(22),1);
    assert.equal(wave.readUInt32LE(24),16000); assert.equal(wave.readUInt32LE(28),32000);
    assert.equal(wave.readUInt16LE(32),2); assert.equal(wave.readUInt16LE(34),16);
    assert.equal(wave.subarray(36,40).toString(),'data'); assert.equal(wave.readUInt32LE(40),wave.length-44);
    assert.ok(wave.length > 44+4800*2 && wave.length <= 1920044 && (wave.length-44)%2 === 0);
    const samples = []; for(let i=44;i<wave.length;i+=2) samples.push(wave.readInt16LE(i));
    assert.ok(samples.some(x=>x>1000) && samples.some(x=>x< -1000), 'No captured waveform');
    assert.ok(samples.every(x=>Math.abs(x)<=8192), 'Unexpected generated amplitude');
    if (process.env.VOICE_OUTPUT) fs.writeFileSync(process.env.VOICE_OUTPUT, wave);
    await page.evaluate(() => clean()); console.log('PASS real worklet samples, canonical WAV, silent output, finish-once and joined cleanup');

    await page.evaluate(async () => { await make(); await current.recorder.start(); });
    const race = await page.evaluate(async () => {
      const pending = current.recorder.finish(); const cancelled = current.recorder.cancel('fixture cancel');
      const settled = await Promise.allSettled([pending, cancelled]);
      const late = await Promise.allSettled([current.recorder.finish(), current.recorder.start()]);
      return {states:settled.map(x=>x.status), late:late.map(x=>x.status), ended:current.stream.getTracks()[0].readyState};
    });
    assert.deepEqual(race.states,['rejected','fulfilled']); assert.deepEqual(race.late,['rejected','rejected']); assert.equal(race.ended,'ended');
    await page.evaluate(() => clean()); console.log('PASS cancel racing finish and late restart/refinish refusal');

    await page.evaluate(async () => { await make({maximumSamples:3200}); await current.recorder.start(); });
    await page.waitForFunction(() => current.recorder.state === 'closed');
    const limit = await page.evaluate(async () => ({events:current.cancelled, ended:current.stream.getTracks()[0].readyState,
      result:(await Promise.allSettled([current.recorder.finish()]))[0].status}));
    assert.deepEqual(limit.events,['Voice recording reached its duration limit.']); assert.equal(limit.ended,'ended'); assert.equal(limit.result,'rejected');
    await page.evaluate(() => clean()); console.log('PASS sample ceiling cancels without partial audio');

    await page.evaluate(async () => {
      await make();
      for (const maximumSamples of [0,-1,960001,1.5,NaN,Infinity]) {
        let refused = false;
        try { new voice.PcmVoiceRecorder(current.stream, {maximumSamples}); } catch { refused = true; }
        if (!refused || current.stream.getTracks()[0].readyState !== 'live') throw new Error('Invalid limit changed ownership or was admitted');
      }
      for (const size of [0,1,1920002]) {
        let refused = false;
        try { pcm.voiceWave(new ArrayBuffer(size)); } catch { refused = true; }
        if (!refused) throw new Error('Invalid PCM size admitted');
      }
      const source = new Uint8Array([1,2,3,4]);
      const wave = pcm.voiceWave(source.buffer); source.fill(0);
      if (wave[44] !== 1 || wave[47] !== 4) throw new Error('WAV shares mutable caller samples');
      await clean();
    });
    console.log('PASS invalid limits preserve caller ownership; WAV byte bounds and copy isolation');

    await page.evaluate(async () => { await make(); await current.recorder.start(); });
    const ended = await page.evaluate(async () => {
      current.stream.getTracks()[0].stop();
      const result = await Promise.allSettled([current.recorder.finish()]); await current.recorder.cancel();
      return {status:result[0].status,state:current.recorder.state};
    });
    assert.deepEqual(ended,{status:'rejected',state:'closed'}); await page.evaluate(() => clean());
    console.log('PASS ended source refuses partial audio even without an ended event');

    await page.evaluate(async () => { await make(); await current.recorder.start(); await contexts.at(-1).suspend(); });
    await page.waitForFunction(() => current.recorder.state === 'closed');
    assert.equal(await page.evaluate(async () => (await Promise.allSettled([current.recorder.finish()]))[0].status),'rejected');
    await page.evaluate(() => clean()); console.log('PASS audio-context suspension cancels rather than finishing');

    const membership = await page.evaluate(async () => {
      await make();
      const original = current.stream.getTracks()[0];
      const unrelated = original.clone();
      try {
        current.stream.removeTrack(original); current.stream.addTrack(unrelated);
        await current.recorder.start(); await current.recorder.finish();
        return {original:original.readyState, unrelated:unrelated.readyState};
      } finally { unrelated.stop(); await clean(); }
    });
    assert.deepEqual(membership,{original:'ended',unrelated:'live'});
    console.log('PASS stream membership edits cannot evade cleanup or transfer unrelated tracks');

    await page.evaluate(async () => { await make(); await current.recorder.start(); nodes.at(-1).port.postMessage({type:'unexpected'}); });
    await page.waitForFunction(() => current.recorder.state === 'closed');
    assert.deepEqual(await page.evaluate(() => current.cancelled),['Voice audio processing failed.']);
    await page.evaluate(() => clean()); console.log('PASS unknown worklet control cancels explicitly');

    mode = 'hold';
    await page.evaluate(async () => { await make(); window.starting = current.recorder.start(); starting.catch(()=>{}); });
    for(let i=0; held.length===0 && i<100; ++i) await new Promise(resolve=>setTimeout(resolve,50));
    assert.equal(held.length,1,'Worklet preparation did not reach gate');
    await page.evaluate(() => { window.cancelling = current.recorder.cancel('pending module'); });
    assert.equal(await page.evaluate(()=>current.stream.getTracks()[0].readyState),'ended');
    mode = 'normal'; held.splice(0).forEach(response=>serveModule(response,'voice-worklet.js'));
    const prepare = await page.evaluate(async () => (await Promise.allSettled([starting,cancelling])).map(x=>x.status));
    assert.deepEqual(prepare,['rejected','fulfilled']);
    await page.evaluate(() => clean()); console.log('PASS cancellation during gated module preparation; late completion inert');

    mode = 'hold';
    await page.evaluate(async () => { await make(); window.starting = current.recorder.start(); starting.catch(()=>{}); });
    for(let i=0; held.length===0 && i<100; ++i) await new Promise(resolve=>setTimeout(resolve,50));
    assert.equal(held.length,1);
    await page.evaluate(() => current.stream.getTracks()[0].stop());
    mode = 'normal'; held.splice(0).forEach(response=>serveModule(response,'voice-worklet.js'));
    const lost = await page.evaluate(async () => ({status:(await Promise.allSettled([starting]))[0].status,state:current.recorder.state}));
    assert.deepEqual(lost,{status:'rejected',state:'closed'});
    await page.evaluate(() => clean()); console.log('PASS source ending during preparation cannot start a late recorder');

    mode = 'fail';
    const failure = await page.evaluate(async () => {
      await make(); const result = await Promise.allSettled([current.recorder.start()]);
      return {status:result[0].status, state:current.recorder.state, track:current.stream.getTracks()[0].readyState, context:contexts.at(-1).state};
    });
    assert.deepEqual(failure,{status:'rejected',state:'closed',track:'ended',context:'closed'});
    mode = 'normal'; await page.evaluate(() => clean()); console.log('PASS failed worklet load releases owned devices');
    const closed = await page.evaluate(() => ({microphoneCalls, closed:contexts.every(c=>c.state==='closed')}));
    assert.deepEqual(closed,{microphoneCalls:0,closed:true}); assert.deepEqual(errors,[]);
    console.log('PASS no microphone access, unhandled errors or retained AudioContexts');
  } finally {
    held.splice(0).forEach(response=>{response.writeHead(503);response.end();});
    if (browser) await browser.close();
    await new Promise(resolve=>server.close(resolve));
  }
})().catch(error=>{console.error(error);process.exitCode=1;});
