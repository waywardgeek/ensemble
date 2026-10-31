// Start --crossover-host in an isolated directory first. No system clipboard access.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || '../../edition-2/agent/gui/node_modules/playwright');
const assert = require('node:assert/strict');
const base = process.env.GUI_URL || 'http://127.0.0.1:18202/assistant/';
const target = new URL(base);
assert.equal(target.hostname, '127.0.0.1', 'Use a disposable loopback fixture');
assert.ok(!['8088', '8089'].includes(target.port), 'Never target working hosts');
(async () => {
  const browser = await chromium.launch({ headless: true, channel: "chrome" });
  const page = await browser.newPage();
  page.setDefaultTimeout(15000);
  const errors = [];
  page.on('pageerror', error => { errors.push(error.message); console.error('PAGE ERROR', error.stack); });
  try {
    if (process.env.IMAGE_MUTANT) {
      await page.route('**/images.js', async route => {
        const response = await route.fetch();
        let body = await response.text();
        const mutation = process.env.IMAGE_MUTANT === 'duplicate'
          ? ['if (!this.hasImages || this.blocked || this.agent !== agent) return false;',
             'if (!this.hasImages || this.agent !== agent) return false;']
          : ['selectionChanged(agent) {', 'selectionChanged(agent) { return;'];
        assert.ok(body.includes(mutation[0]), 'mutant anchor must match');
        body = body.replace(mutation[0], mutation[1]);
        await route.fulfill({ response, body });
      });
    }
    await page.addInitScript(() => {
      window.imageSends = []; window.imageAcks = []; window.revoked = []; window.created = [];
      const native = WebSocket;
      window.WebSocket = class extends native {
        constructor(...args) {
          super(...args);
          this.addEventListener('message', e => {
            const frame = JSON.parse(e.data);
            if (frame.type === 'image_accepted') imageAcks.push(frame);
          });
        }
        send(data) {
          const frame = JSON.parse(data);
          if (frame.type === 'image_prompt') {
            imageSends.push(frame);
            if (window.dropImages) return;
          }
          super.send(data);
        }
      };
      const create = URL.createObjectURL.bind(URL), revoke = URL.revokeObjectURL.bind(URL);
      URL.createObjectURL = value => { const url = create(value); created.push(url); return url; };
      URL.revokeObjectURL = url => { revoked.push(url); revoke(url); };
      window.pasteImages = async (count = 1, caption = '') => {
        const canvas = document.createElement('canvas'); canvas.width = 12; canvas.height = 8;
        canvas.getContext('2d').fillStyle = '#d6317e'; canvas.getContext('2d').fillRect(0, 0, 12, 8);
        const blob = await new Promise(resolve => canvas.toBlob(resolve, 'image/png'));
        const transfer = new DataTransfer();
        for (let i = 0; i < count; i++) transfer.items.add(new File([blob], `synthetic-${i}.png`, {type:'image/png'}));
        if (caption) transfer.setData('text/plain', caption);
        document.querySelector('#prompt').dispatchEvent(new ClipboardEvent('paste', {clipboardData:transfer, bubbles:true, cancelable:true}));
      };
    });
    await page.goto(base);
    await page.waitForFunction(() => !document.querySelector('#send').disabled);
    await page.evaluate(() => send('set_model', 'claude-3-haiku-20240307'));
    await page.waitForFunction(() => document.querySelector('#model-choice').value === 'claude-3-haiku-20240307' && !document.querySelector('#model-choice').disabled);
    await page.locator('#prompt').fill('kept ');
    await page.evaluate(() => pasteImages(2, 'caption'));
    await page.waitForFunction(() => imageComposer.entries.length === 2 && !imageComposer.blocked);
    assert.equal(await page.locator('#image-previews img').count(), 2);
    assert.equal(await page.locator('#prompt').inputValue(), 'kept caption');
    await page.locator('#image-previews button').first().click();
    assert.equal(await page.locator('#image-previews img').count(), 1);
    await page.evaluate(() => { submit(); submit(); });
    await page.waitForFunction(() => imageAcks.length >= 1 && !imageComposer.hasImages);
    const admitted = await page.evaluate(() => ({ sends:imageSends.length, frame:imageSends[0], ack:imageAcks[0], text:document.querySelector('#prompt').value }));
    assert.equal(admitted.sends, 1); assert.equal(admitted.frame.text, 'kept caption');
    assert.equal(admitted.frame.images.length, 1); assert.ok(admitted.frame.images[0].data.length > 40);
    assert.equal(admitted.ack.images[0].width, 12); assert.equal(admitted.ack.images[0].height, 8);
    assert.equal(admitted.text, '');
    console.log('PASS real server: preview/removal, caption, exact-once send, pixel preparation acknowledgment');

    // Model-independent lifecycle checks use the real component in Chromium.
    await page.evaluate(async () => {
      await pasteImages();
    });
    await page.waitForFunction(() => !imageComposer.blocked && imageComposer.hasImages);
    await page.evaluate(() => {
      window.controlledSendCount = 0;
      const connection = { send() { controlledSendCount++; } };
      window.controlledConnection = connection;
      imageComposer.send(connection, agentID);
      imageComposer.send(connection, agentID);
      window.ignoredAcks = [
        imageComposer.acknowledge({type:'error', image_sequence:99}, connection),
        imageComposer.acknowledge({type:'error', image_sequence:1}, {}),
        imageComposer.acknowledge({type:'image_accepted', sequence:1, agent_id:'wrong-agent'}, connection)
      ];
      window.pendingAfterWrongAcks = !!imageComposer.pending;
      imageComposer.acknowledge({type:'error', image_sequence:1, message:'Synthetic refusal'}, connection);
    });
    assert.equal(await page.evaluate(() => controlledSendCount), 1);
    assert.deepEqual(await page.evaluate(() => ignoredAcks), [false, false, false]);
    assert.equal(await page.evaluate(() => pendingAfterWrongAcks), true);
    assert.equal(await page.locator('#image-previews img').count(), 1);
    assert.equal(await page.locator('#prompt').evaluate(e => e.readOnly), false);
    await page.evaluate(() => imageComposer.clear());
    await page.evaluate(async () => {
      window.originalReader = FileReader; window.heldReaders = []; window.aborts = 0;
      window.FileReader = class {
        static LOADING = 1;
        constructor() { this.readyState = 0; heldReaders.push(this); }
        readAsDataURL() { this.readyState = 1; }
        abort() { this.readyState = 2; aborts++; }
      };
      await pasteImages();
      imageComposer.selectionChanged('different-agent');
      heldReaders[0].result = 'data:image/png;base64,AAAA'; heldReaders[0].onload();
      window.FileReader = originalReader;
    });
    assert.equal(await page.evaluate(() => imageComposer.hasImages), false);
    assert.equal(await page.evaluate(() => aborts), 1);
    console.log('PASS controlled lifecycle: refusal retains draft; selection aborts reader and ignores late callback');

    await page.evaluate(() => pasteImages(5));
    assert.equal(await page.evaluate(() => imageComposer.hasImages), false);
    await page.evaluate(() => pasteImages());
    await page.waitForFunction(() => !imageComposer.blocked && imageComposer.hasImages);
    await page.setViewportSize({width:393, height:852});
    assert.equal(await page.locator('#image-previews button').isVisible(), true);
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1));
    await page.evaluate(() => { window.dropImages = true; submit(); socket.close(); });
    await page.waitForFunction(() => imageComposer.pending?.uncertain && !selectionPending && socket.readyState === WebSocket.OPEN);
    const count = await page.evaluate(() => imageSends.length);
    await page.evaluate(() => submit());
    assert.equal(await page.evaluate(() => imageSends.length), count);
    assert.equal(await page.locator('#send').isDisabled(), true);
    assert.equal(await page.locator('#discard-image-draft').isVisible(), true);
    await page.locator('#discard-image-draft').click();
    assert.equal(await page.evaluate(() => imageComposer.hasImages), false);
    assert.equal(await page.evaluate(() => created.every(url => revoked.includes(url))), true);
    console.log('PASS phone geometry, count bound, real reconnect without resend, explicit discard, URL cleanup');
    assert.deepEqual(errors, []);
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
