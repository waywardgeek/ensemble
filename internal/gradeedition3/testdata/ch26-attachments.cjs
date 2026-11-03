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
              agent_id: frame.agent_id, text: frame.text, images: frame.images?.map(image =>
                ({ width: image.width, height: image.height })) });
          } catch { /* Non-JSON traffic is not admission evidence. */ }
        });
      }
    });
    if (process.env.ATTACHMENT_MUTATION) {
      assert.equal(process.env.ATTACHMENT_MUTATION, 'lost-caption');
      await page.route('**/images.js', async route => {
        const response = await route.fetch();
        const source = await response.text();
        const anchor = 'text: this.input.value, images:';
        assert.ok(source.includes(anchor), 'Mutation anchor missing');
        await route.fulfill({ response, body: source.replace(anchor, "text: '', images:") });
      });
    }
    await page.goto(base.href);
    await page.locator('#send').waitFor({ state: 'visible' });
    await page.waitForFunction(() => !document.querySelector('#send').disabled);
    await page.locator('#prompt').fill(caption);
    const chooser = page.waitForEvent('filechooser');
    await page.locator('#attach-images').click();
    await (await chooser).setFiles(process.env.ATTACHMENT_PNG);
    await page.locator('#image-previews img').waitFor({ state: 'visible' });
    assert.equal(await page.locator('#prompt').inputValue(), caption);
    await page.waitForFunction(() => !document.querySelector('#send').disabled);
    // Public DOM submission twice within one task exercises the once-only gate.
    await page.locator('#prompt').evaluate(input => {
      input.form.requestSubmit();
      input.form.requestSubmit();
    });
    await page.waitForFunction(() => document.querySelectorAll('#image-previews img').length === 0);
    const sends = frames.filter(frame => frame.direction === 'sent' && frame.type === 'image_prompt');
    assert.equal(sends.length, 1, 'Duplicate or missing image admission request');
    assert.equal(sends[0].text, caption, 'Caption changed in transport');
    assert.equal(sends[0].images.length, 1, 'Image missing in transport');
    const accepted = frames.filter(frame => frame.direction === 'received' &&
      frame.type === 'image_accepted' && frame.sequence === sends[0].sequence &&
      frame.agent_id === sends[0].agent_id);
    assert.equal(accepted.length, 1, 'Missing or duplicate addressed acknowledgement');
    assert.equal(accepted[0].images[0].width, 2048);
    assert.equal(accepted[0].images[0].height, 32);
    assert.equal(await page.locator('#prompt').inputValue(), '', 'Accepted draft not cleared');
    // Admission is not completion: observe the answer through the public UI.
    await page.getByText('ATTACHMENT_PROBE_COMPLETE', { exact: true }).waitFor({ state: 'visible' });
    assert.deepEqual(errors, [], 'Browser exceptions');
    fs.writeFileSync(process.env.ATTACHMENT_RESULT, JSON.stringify({
      scope: 'attachment-picker-admission-only', caption, request: sends[0], acknowledgement: accepted[0]
    }, null, 2));
  } finally {
    await browser.close();
  }
})().catch(error => { console.error(error.stack); process.exitCode = 1; });
