// External candidate browser driver. No application-private globals or fixture host.
// Provider-side pixels and recipient ownership must be checked by the owning grader.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE);
const base = new URL(process.env.GUI_URL);
assert.equal(base.hostname, '127.0.0.1', 'Only an isolated loopback candidate is allowed');
assert.ok(base.port && !['8088', '8089', '8090'].includes(base.port), 'Never target working services');
const caption = process.env.ATTACHMENT_CAPTION;
assert.ok(caption && process.env.ATTACHMENT_PNG && process.env.ATTACHMENT_RESULT);

(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' });
  try {
    const page = await browser.newPage();
    page.setDefaultTimeout(15000);
    const frames = [];
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('websocket', socket => {
      for (const [event, direction] of [['framesent', 'sent'], ['framereceived', 'received']]) {
        socket.on(event, ({ payload }) => {
          try {
            const frame = JSON.parse(String(payload));
            // Retain protocol evidence, never copy image payloads into the report.
            frames.push({ direction, type: frame.type, sequence: frame.sequence,
              agent_id: frame.agent_id, text: frame.text, message: frame.message, image_sequence: frame.image_sequence, images: frame.images?.map(image =>
                ({ width: image.width, height: image.height })) });
          } catch { /* Non-JSON traffic is not admission evidence. */ }
        });
      }
    });
    if (process.env.ATTACHMENT_MUTATION) {
      const mutations = {
        'lost-caption': ['text: this.input.value, images:', "text: '', images:"],
        'preview-only': ['connection.send(JSON.stringify', 'void (JSON.stringify'],
        'misdirected': ['agent_id: agent, sequence,', 'agent_id: "AttachmentProbe", sequence,'],
        'duplicate-send': ['connection.send(JSON.stringify', 'for (let duplicate = 0; duplicate < 2; duplicate++) connection.send(JSON.stringify']
      };
      const mutation = mutations[process.env.ATTACHMENT_MUTATION];
      assert.ok(mutation, 'Unknown mutation');
      await page.route('**/images.js', async route => {
        const response = await route.fetch();
        const source = await response.text();
        assert.ok(source.includes(mutation[0]), 'Mutation anchor missing');
        await route.fulfill({ response, body: source.replace(mutation[0], mutation[1]) });
      });
    }
    if (['refusal', 'uncertain', 'duplicate-wire'].includes(process.env.ATTACHMENT_FAULT)) {
      await page.addInitScript(fault => {
        const Native = WebSocket;
        let damaged = false;
        window.attachmentFaultEvidence = { opens: 0, attempts: 0 };
        window.WebSocket = class extends Native {
          constructor(...args) {
            super(...args);
            this.addEventListener('open', () => window.attachmentFaultEvidence.opens++);
          }
          send(data) {
            const frame = JSON.parse(data);
            if (frame.type === 'image_prompt') window.attachmentFaultEvidence.attempts++;
            if (!damaged && frame.type === 'image_prompt') {
              damaged = true;
              if (fault === 'uncertain') { this.close(); return; }
              if (fault === 'duplicate-wire') { super.send(data); super.send(data); return; }
              frame.images[0].data = 'invalid-base64!';
              data = JSON.stringify(frame);
            }
            super.send(data);
          }
        };
      }, process.env.ATTACHMENT_FAULT);
    }
    if (process.env.ATTACHMENT_FAULT === 'acknowledgements') {
      await page.addInitScript(() => {
        const Native = WebSocket;
        window.attachmentAckEvidence = { held: false, release: null };
        window.WebSocket = class extends Native {
          constructor(...args) {
            super(...args);
            this.addEventListener('message', event => {
              if (!event.isTrusted) return; // Allow our explicitly injected events through.
              const frame = JSON.parse(event.data);
              if (frame.type !== 'image_accepted') return;
              event.stopImmediatePropagation();
              const deliver = value => this.dispatchEvent(new MessageEvent('message', {
                data: JSON.stringify(value)
              }));
              window.attachmentAckEvidence.release = () => deliver(frame);
              deliver({ ...frame, sequence: frame.sequence + 100 });
              deliver({ ...frame, agent_id: 'not-the-recording-agent' });
              window.attachmentAckEvidence.held = true;
            });
          }
        };
      });
    }
    await page.goto(base.href);
    await page.locator('#send').waitFor({ state: 'visible' });
    await page.waitForFunction(() => !document.querySelector('#send').disabled);
    if (process.env.ATTACHMENT_RECIPIENT === '1') {
      await page.locator('#prompt').fill(caption + '-original-history');
      await page.locator('#send').click();
      await page.getByText('ATTACHMENT_PROBE_COMPLETE', { exact: true }).waitFor({ state: 'visible' });
      await page.locator('#peer-open').click();
      await page.locator('#peer-name').fill('AttachmentRecipient');
      await page.locator('#peer-model').fill('claude-3-haiku-20240307');
      await page.locator('#peer-create').click();
      const recipient = page.locator('#agent-tree [data-agent]').filter({ hasText: 'AttachmentRecipient' });
      await recipient.click();
      await page.waitForFunction(() => !document.querySelector('#send').disabled);
      await page.locator('#prompt').fill(caption + '-recipient-history');
      await page.locator('#send').click();
      await page.getByText('ATTACHMENT_PROBE_COMPLETE', { exact: true }).waitFor({ state: 'visible' });
    }
    if (process.env.ATTACHMENT_FAULT === 'late-reader') {
      await page.evaluate(() => {
        const Native = FileReader;
        let held = false;
        window.attachmentReaderEvidence = { aborts: 0, reader: null };
        window.FileReader = class {
          static LOADING = 1;
          constructor() {
            if (held) return new Native();
            held = true; this.readyState = 0;
            window.attachmentReaderEvidence.reader = this;
          }
          readAsDataURL() { this.readyState = 1; }
          abort() { this.readyState = 2; window.attachmentReaderEvidence.aborts++; }
        };
      });
      const chooser = page.waitForEvent('filechooser');
      await page.locator('#attach-images').click();
      await (await chooser).setFiles(process.env.ATTACHMENT_PNG);
      await page.waitForFunction(() => window.attachmentReaderEvidence.reader?.readyState === 1);
      await page.locator('#agent-tree [data-agent]').filter({ hasText: 'AttachmentProbe' }).click();
      await page.waitForFunction(() => window.attachmentReaderEvidence.aborts === 1);
      await page.locator('#agent-tree [data-agent]').filter({ hasText: 'AttachmentRecipient' }).click();
      await page.waitForFunction(() => !document.querySelector('#send').disabled);
      await page.evaluate(encoded => {
        const reader = window.attachmentReaderEvidence.reader;
        reader.result = 'data:image/png;base64,' + encoded;
        reader.onload();
      }, fs.readFileSync(process.env.ATTACHMENT_PNG).toString('base64'));
      assert.equal(await page.locator('#image-previews img').count(), 0, 'Stale reader restored an image after switching');
    }
    await page.locator('#prompt').fill(caption);
    if (process.env.ATTACHMENT_INPUT === 'paste') {
      const encoded = fs.readFileSync(process.env.ATTACHMENT_PNG).toString('base64');
      await page.locator('#prompt').evaluate((input, encoded) => {
        const bytes = Uint8Array.from(atob(encoded), value => value.charCodeAt(0));
        const transfer = new DataTransfer();
        for (let index = 0; index < 2; index++) {
          transfer.items.add(new File([bytes], `synthetic-${index}.png`, { type: 'image/png' }));
        }
        input.dispatchEvent(new ClipboardEvent('paste', {
          clipboardData: transfer, bubbles: true, cancelable: true
        }));
      }, encoded);
      await page.waitForFunction(() => document.querySelectorAll('#image-previews img').length === 2);
      await page.locator('#image-previews button').first().click();
    } else {
      const chooser = page.waitForEvent('filechooser');
      await page.locator('#attach-images').click();
      await (await chooser).setFiles(process.env.ATTACHMENT_PNG);
    }
    await page.locator('#image-previews img').waitFor({ state: 'visible' });
    assert.equal(await page.locator('#image-previews img').count(), 1, 'Preview removal failed');
    assert.equal(await page.locator('#prompt').inputValue(), caption);
    await page.waitForFunction(() => !document.querySelector('#send').disabled);
    const expectedAgent = await page.locator('#agent-tree [aria-selected="true"] > .tree-row > [data-agent]').getAttribute('data-agent');
    // Public DOM submission twice within one task exercises the once-only gate.
    await page.locator('#prompt').evaluate(input => {
      input.form.requestSubmit();
      input.form.requestSubmit();
    });
    if (process.env.ATTACHMENT_FAULT !== 'uncertain') {
      const deadline = Date.now() + 5000;
      while (!frames.some(frame => frame.direction === 'sent' && frame.type === 'image_prompt')) {
        assert.ok(Date.now() < deadline, 'Preview-only: no image envelope was sent');
        await page.waitForTimeout(20);
      }
      const first = frames.find(frame => frame.direction === 'sent' && frame.type === 'image_prompt');
      assert.equal(first.agent_id, expectedAgent, 'Image envelope targets the wrong Agent');
      assert.equal(first.text, caption, 'Caption changed in transport');
    }
    if (process.env.ATTACHMENT_FAULT === 'acknowledgements') {
      await page.waitForFunction(() => window.attachmentAckEvidence.held);
      assert.equal(await page.locator('#image-previews img').count(), 1, 'Wrong acknowledgement cleared image');
      assert.equal(await page.locator('#prompt').inputValue(), caption, 'Wrong acknowledgement cleared caption');
      assert.equal(await page.locator('#prompt').evaluate(input => input.readOnly), true,
        'Wrong acknowledgement released pending draft');
      await page.evaluate(() => {
        window.attachmentAckEvidence.release();
        window.attachmentAckEvidence.release();
      });
      await page.waitForFunction(() => document.querySelectorAll('#image-previews img').length === 0);
      await page.locator('#prompt').fill('fresh draft must survive an old acknowledgement');
      const chooser = page.waitForEvent('filechooser');
      await page.locator('#attach-images').click();
      await (await chooser).setFiles(process.env.ATTACHMENT_PNG);
      await page.locator('#image-previews img').waitFor({ state: 'visible' });
      await page.evaluate(() => window.attachmentAckEvidence.release());
      assert.equal(await page.locator('#image-previews img').count(), 1, 'Old acknowledgement cleared fresh image');
      assert.equal(await page.locator('#prompt').inputValue(), 'fresh draft must survive an old acknowledgement');
      await page.locator('#image-previews button').click();
      await page.locator('#prompt').fill('');
    }
    if (process.env.ATTACHMENT_FAULT === 'uncertain') {
      await page.locator('#discard-image-draft').waitFor({ state: 'visible' });
      await page.waitForFunction(() => window.attachmentFaultEvidence.opens >= 2);
      assert.equal(await page.locator('#send').isDisabled(), true, 'Uncertain draft became sendable');
      await page.locator('#prompt').evaluate(input => {
        input.form.requestSubmit(); input.form.requestSubmit();
      });
      assert.equal(await page.evaluate(() => window.attachmentFaultEvidence.attempts), 1,
        'Uncertain send retried after reconnect');
      assert.equal(await page.locator('#prompt').inputValue(), caption, 'Uncertain caption lost');
      assert.equal(await page.locator('#image-previews img').count(), 1, 'Uncertain image lost');
      await page.locator('#discard-image-draft').click();
      assert.equal(await page.locator('#image-previews img').count(), 0);
      await page.locator('#prompt').fill(caption);
      const chooser = page.waitForEvent('filechooser');
      await page.locator('#attach-images').click();
      await (await chooser).setFiles(process.env.ATTACHMENT_PNG);
      await page.waitForFunction(() => !document.querySelector('#send').disabled);
      await page.locator('#send').click();
    }
    if (process.env.ATTACHMENT_FAULT === 'refusal') {
      const deadline = Date.now() + 15000;
      while (!frames.some(frame => frame.direction === 'received' && frame.type === 'error' && frame.image_sequence === 1)) {
        assert.ok(Date.now() < deadline, 'Missing correlated image refusal');
        await page.waitForTimeout(20);
      }
      await page.waitForFunction(() => !document.querySelector('#prompt').readOnly);
      assert.equal(await page.locator('#prompt').inputValue(), caption, 'Refusal lost caption');
      assert.equal(await page.locator('#prompt').evaluate(input => input.readOnly), false, 'Refusal locked draft');
      assert.equal(await page.locator('#image-previews img').count(), 1, 'Refusal lost image');
      await page.locator('#send').click();
    }
    await page.waitForFunction(() => document.querySelectorAll('#image-previews img').length === 0);
    const sends = frames.filter(frame => frame.direction === 'sent' && frame.type === 'image_prompt');
    assert.equal(sends.length, ['refusal', 'duplicate-wire'].includes(process.env.ATTACHMENT_FAULT) ? 2 : 1, 'Duplicate or missing image admission request');
    const admitted = sends.at(-1);
    assert.equal(admitted.text, caption, 'Caption changed in transport');
    assert.equal(admitted.images.length, 1, 'Image missing in transport');
    const accepted = frames.filter(frame => frame.direction === 'received' &&
      frame.type === 'image_accepted' && frame.sequence === admitted.sequence &&
      frame.agent_id === admitted.agent_id);
    assert.equal(accepted.length, 1, 'Missing or duplicate addressed acknowledgement');
    if (process.env.ATTACHMENT_FAULT === 'duplicate-wire') {
      const deadline = Date.now() + 15000;
      while (!frames.some(frame => frame.direction === 'received' && frame.type === 'error' && frame.image_sequence === admitted.sequence)) {
        assert.ok(Date.now() < deadline, 'Replayed image envelope was not refused');
        await page.waitForTimeout(20);
      }
    }
    assert.equal(accepted[0].images[0].width, 2048);
    assert.equal(accepted[0].images[0].height, 32);
    assert.equal(await page.locator('#prompt').inputValue(), '', 'Accepted draft not cleared');
    // Admission is not completion: require the newly produced answer.
    const replies = process.env.ATTACHMENT_RECIPIENT === '1' ? 2 : 1;
    await page.waitForFunction(count => [...document.querySelectorAll('#transcript p')]
      .filter(node => node.textContent === 'ATTACHMENT_PROBE_COMPLETE').length === count, replies);
    assert.deepEqual(errors, [], 'Browser exceptions');
    fs.writeFileSync(process.env.ATTACHMENT_RESULT, JSON.stringify({
      scope: 'attachment-picker-admission-only', caption, request: admitted, acknowledgement: accepted[0]
    }, null, 2));
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error.stack); process.exitCode = 1; });
