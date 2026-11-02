const fs = require('fs');
const http = require('http');
const path = require('path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const asset = path.resolve(__dirname, '../gui/wwwroot/page-capture.js');
function source() {
  let text = fs.readFileSync(asset, 'utf8');
  const mutant = process.env.PAGE_CAPTURE_MUTANT;
  let before, after;
  if (mutant === 'identity') {
    before = '|| identity?.handle !== state.handle || identity?.origin !== location.origin';
    after = '|| false';
  } else if (mutant === 'late-result') {
    before = 'const blob = await new Promise(resolve => canvas.toBlob(resolve, "image/png"));\n            this.#check(state, track);';
    after = 'const blob = await new Promise(resolve => canvas.toBlob(resolve, "image/png"));';
  } else if (mutant) throw new Error('Unknown fixture mutant');
  if (before) {
    if (text.split(before).length !== 2) throw new Error('Mutation anchor not unique');
    text = text.replace(before, after);
  }
  return text;
}
const servedSource = source();
const server = http.createServer((req, res) => {
  if (req.url === '/') { res.setHeader('Content-Type', 'text/html'); res.end('<!doctype html><title>Owned capture fixture</title>'); }
  else if (req.url === '/page-capture.js') { res.setHeader('Content-Type', 'text/javascript'); res.end(servedSource); }
  else { res.statusCode = 404; res.end(); }
});
(async () => {
  let browser;
  try {
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    browser = await chromium.launch({headless: true, channel: 'chrome', args: ['--autoplay-policy=no-user-gesture-required', '--mute-audio']});
    const page = await browser.newPage();
    page.setDefaultTimeout(10000);
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto(`http://127.0.0.1:${server.address().port}/`);
    const results = await page.evaluate(async () => {
      const { PageCapture } = await import('./page-capture.js');
      const results = [];
      const check = (value, reason) => { if (!value) throw new Error(reason); };
      const gate = () => { let resolve; const promise = new Promise(r => { resolve = r; }); return {promise, resolve}; };
      const bounded = async promise => {
        let timer;
        try { return await Promise.race([promise, new Promise((_, reject) => { timer = setTimeout(() => reject(new Error('Fixture timed out')), 8000); })]); }
        finally { clearTimeout(timer); }
      };
      const rejects = async promise => { let failed = false; try { await bounded(promise); } catch (error) { if (error.message === 'Fixture timed out') throw error; failed = true; } check(failed, 'Expected explicit refusal'); };
      function fixture(policy = {}) {
        let config, calls = 0;
        const canvas = document.createElement('canvas');
        canvas.width = 64; canvas.height = 32;
        const context = canvas.getContext('2d');
        context.fillStyle = '#ff0000'; context.fillRect(0, 0, 32, 32);
        context.fillStyle = '#0000ff'; context.fillRect(32, 0, 32, 32);
        const stream = canvas.captureStream();
        const track = stream.getVideoTracks()[0];
        let identity;
        Object.defineProperty(track, 'getSettings', {value: () => ({displaySurface: policy.surface || 'browser'})});
        Object.defineProperty(track, 'getCaptureHandle', {value: () => { policy.checked?.(); return identity || ({handle: policy.handle || config.handle, origin: policy.origin || location.origin}); }});
        const devices = {
          setCaptureHandleConfig: value => { config = value; },
          getDisplayMedia: options => {
            calls++;
            check(options.audio === false && options.selfBrowserSurface === 'include' && options.surfaceSwitching === 'exclude', 'Unsafe capture hints');
            check(config.exposeOrigin && config.permittedOrigins.length === 1 && config.permittedOrigins[0] === location.origin, 'Overbroad handle exposure');
            return policy.request ? policy.request(stream) : Promise.resolve(stream);
          }
        };
        return {capture: new PageCapture(devices), stream, track, canvas, calls: () => calls, changeIdentity: value => { identity = value; track.dispatchEvent(new Event('capturehandlechange')); }, cleanup: () => { for (const t of stream.getTracks()) t.stop(); track.stop(); canvas.width = 0; }};
      }
      let f = fixture();
      try {
        const image = await bounded(f.capture.capture());
        check(image.width === 64 && image.height === 32 && image.file.type === 'image/png', 'Missing raster result');
        const bitmap = await createImageBitmap(image.file);
        const canvas = document.createElement('canvas'); canvas.width = 64; canvas.height = 32;
        const context = canvas.getContext('2d'); context.drawImage(bitmap, 0, 0); bitmap.close();
        const left = context.getImageData(16, 16, 1, 1).data, right = context.getImageData(48, 16, 1, 1).data;
        check(left[0] > 240 && left[2] < 15 && right[2] > 240 && right[0] < 15, 'PNG did not carry actual red/blue pixels');
        check(f.track.readyState === 'ended' && !f.capture.busy, 'Successful capture leaked resources');
        canvas.width = 0;
        results.push('PASS real video frame -> actual red/blue PNG pixels and joined cleanup');
      } finally { f.cleanup(); }
      for (const policy of [{handle:'another-tab'}, {origin:'https://unrelated.invalid'}, {surface:'monitor'}, {surface:'window'}]) {
        f = fixture(policy);
        try { await rejects(f.capture.capture()); check(f.track.readyState === 'ended' && !f.capture.busy, 'Rejected capture leaked track'); }
        finally { f.cleanup(); }
      }
      results.push('PASS wrong tab, origin, monitor and window refuse and stop tracks');
      let invoked = false;
      const unsupported = new PageCapture({getDisplayMedia: () => { invoked = true; }});
      await rejects(unsupported.capture()); check(!invoked, 'Unsupported capture prompted user');
      f = fixture({request: () => Promise.reject(new DOMException('Denied', 'NotAllowedError'))});
      try { await rejects(f.capture.capture()); check(!f.capture.busy, 'Denied picker wedged owner'); }
      finally { f.cleanup(); }
      results.push('PASS unsupported and denied capture fail explicitly without stuck owner');
      const pending = gate();
      f = fixture({request: () => pending.promise});
      try {
        const capture = f.capture.capture(); const refused = rejects(capture);
        await rejects(f.capture.capture()); check(f.calls() === 1, 'Duplicate call opened another picker');
        const closing = f.capture.cancel(); check(f.capture.busy, 'Cancellation abandoned pending picker');
        pending.resolve(f.stream); await bounded(closing); await refused;
        check(f.track.readyState === 'ended' && !f.capture.busy, 'Late picker result leaked or survived');
      } finally { pending.resolve(f.stream); f.cleanup(); }
      results.push('PASS single picker, cancellation joins late permission and suppresses result');
      const checking = gate();
      f = fixture({checked: () => checking.resolve()});
      try {
        const capture = f.capture.capture();
        const refused = rejects(capture);
        await bounded(checking.promise);
        await bounded(f.capture.cancel());
        await refused;
        check(f.track.readyState === 'ended' && !f.capture.busy, 'Frame-wait cancellation did not join video cleanup');
      } finally { await f.capture.cancel(); f.cleanup(); }
      results.push('PASS cancellation before first frame joins pending playback and stops tracks');
      const originalBlob = HTMLCanvasElement.prototype.toBlob;
      for (const action of ['cancel', 'identity', 'ended', 'membership']) {
        const encoding = gate(); let finish;
        HTMLCanvasElement.prototype.toBlob = function(callback, type) {
          const canvas = this;
          finish = () => originalBlob.call(canvas, callback, type);
          encoding.resolve();
        };
        f = fixture();
        let foreign;
        try {
          const capture = f.capture.capture();
          // Install rejection observation before intentionally invalidating.
          const observed = capture.then(value => ({value}), error => ({error}));
          await bounded(encoding.promise);
          let closing;
          if (action === 'cancel') { closing = f.capture.cancel(); check(f.track.readyState === 'ended' && f.capture.busy, 'Cancel did not stop tracks while joining encoding'); }
          if (action === 'identity') f.changeIdentity({handle:'other', origin:location.origin});
          if (action === 'ended') { f.track.stop(); f.track.dispatchEvent(new Event('ended')); }
          if (action === 'membership') {
            foreign = fixture(); f.stream.removeTrack(f.track); f.stream.addTrack(foreign.track);
          }
          const complete = finish; finish = null; complete(); const result = await bounded(observed); if (closing) await bounded(closing);
          if (action === 'membership') check(result.value && foreign.track.readyState === 'live', 'Mutated stream redirected cleanup');
          else check(result.error && !result.value, 'Invalid capture published late PNG');
          check(f.track.readyState === 'ended' && !f.capture.busy, 'Encoding cleanup incomplete');
          results.push('PASS encoding ' + action + ': original-track ownership, cleanup and late-result policy');
        } finally { HTMLCanvasElement.prototype.toBlob = originalBlob; finish?.(); await f.capture.cancel(); f.cleanup(); foreign?.cleanup(); }
      }
      const widthDescriptor = Object.getOwnPropertyDescriptor(HTMLVideoElement.prototype, 'videoWidth');
      const heightDescriptor = Object.getOwnPropertyDescriptor(HTMLVideoElement.prototype, 'videoHeight');
      f = fixture();
      try {
        Object.defineProperty(HTMLVideoElement.prototype, 'videoWidth', {configurable:true, get:()=>5000});
        Object.defineProperty(HTMLVideoElement.prototype, 'videoHeight', {configurable:true, get:()=>5000});
        await rejects(f.capture.capture());
        check(f.track.readyState === 'ended' && !f.capture.busy, 'Pixel refusal leaked resources');
      } finally {
        Object.defineProperty(HTMLVideoElement.prototype, 'videoWidth', widthDescriptor);
        Object.defineProperty(HTMLVideoElement.prototype, 'videoHeight', heightDescriptor);
        f.cleanup();
      }
      for (const oversized of [false,true]) {
        f = fixture();
        try {
          HTMLCanvasElement.prototype.toBlob = function(callback) {
            callback(oversized ? new Blob([new Uint8Array(32*1024*1024+1)],{type:'image/png'}) : null);
          };
          await rejects(f.capture.capture());
          check(f.track.readyState === 'ended' && !f.capture.busy, 'PNG refusal leaked resources');
        } finally { HTMLCanvasElement.prototype.toBlob=originalBlob; f.cleanup(); }
      }
      results.push('PASS over-limit pixel metadata, failed encoding and oversized PNG refuse with cleanup');
      return results;
    });
    if (errors.length) throw new Error('Unexpected page errors: ' + errors.join('; '));
    for (const result of results) console.log(result);
  } finally { await browser?.close(); await new Promise(resolve => server.close(resolve)); }
})().catch(error => { console.error(error); process.exitCode = 1; });
