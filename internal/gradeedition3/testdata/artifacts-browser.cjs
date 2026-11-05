'use strict';
// Candidate's public page and HTTP routes. No app globals or injected wire frames.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE);
const assert = require('node:assert/strict');
const fs = require('node:fs');
const crypto = require('node:crypto');
const base = process.env.ARTIFACT_GUI_URL;
const origin = new URL(base).origin;
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
(async () => {
  const browser = await chromium.launch({headless: true, channel: 'chrome'});
  process.once('SIGINT', () => { void browser.close().finally(() => process.exit(130)); });
  try {
    const page = await browser.newPage({viewport: {width: 390, height: 844}});
    page.setDefaultTimeout(20000);
    const errors = [], transfers = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('request', request => { if (request.url().includes('/work-products/download?')) transfers.push(request); });
    await page.goto(base);
    await page.locator('.pane-tab[data-pane="agents"]').click();
    await page.locator('#agent-tree [data-agent="candidate"]').click();
    await page.waitForFunction(() => document.querySelector('#state').textContent === 'idle'
      && document.querySelector('#prompt').placeholder.includes('candidate'));
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    let status = await (await page.request.get(origin + '/grade/status')).json();
    await page.locator('#prompt').fill('prepare ' + status.nonce.slice(0, 16));
    await page.locator('#send').click();
    await page.getByText(status.final, {exact: true}).waitFor();
    status = await (await page.request.get(origin + '/grade/status')).json();
    assert.equal(status.failure, null, 'provider oracle must succeed');
    assert.equal(status.requests, 3, 'real write/offer/result sequence must finish once');
    assert.equal(status.returned, 2, 'two tool results must be recorded');
    assert.equal(status.offers, 1, 'exactly one successful offer must be retained');
    assert.equal(transfers.length, 0, 'rendering an offer must not transfer it');
    await page.locator('.pane-tab[data-pane="tools"]').click();
    const button = page.getByRole('button', {name: 'Download ' + status.filename, exact: true});
    await button.waitFor();
    // Change grader-owned file after offering; click must read the new bytes.
    status = await (await page.request.post(origin + '/grade/change-file')).json();
    const awaited = page.waitForEvent('download');
    await button.click();
    const download = await awaited;
    assert.equal(download.suggestedFilename(), status.filename);
    const bytes = fs.readFileSync(await download.path());
    assert.equal(bytes.length, status.bytes);
    assert.equal(crypto.createHash('sha256').update(bytes).digest('hex'), status.hash, 'exact click-time binary download');
    assert.equal(transfers.length, 1);
    const route = origin + '/assistant/work-products/download?agent=candidate&offer=' + status.offer.id;
    assert.equal((await page.request.get(route)).status(), 403, 'custom header required');
    const headers = {'X-Ensemble-Work-Product': '1'};
    assert.equal((await page.request.get(route, {headers:{...headers, Origin:'https://foreign.invalid'}})).status(), 403);
    assert.equal((await page.request.get(route + '&path=anything', {headers})).status(), 400);
    assert.equal((await page.request.get(route.replace('agent=candidate', 'agent=viewer'), {headers})).status(), 404);
    await page.request.post(origin + '/grade/replace-owner');
    assert.equal((await page.request.get(route, {headers})).status(), 404, 'replacement Agent must not revive offer');
    assert.deepEqual(errors, []);
    console.log(JSON.stringify({contract:'artifact-browser-work-product-v1', passed:true,
      checks:['real-model-tool-sequence','click-only','exact-current-bytes','filename','origin-header','no-path-query','foreign-owner','replacement-expiry'],
      ungraded:'Full artifact rubric and chapter points are not implied.'}));
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
