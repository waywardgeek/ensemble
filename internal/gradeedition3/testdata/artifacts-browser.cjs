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
    const errors = [], transfers = [], downloads = [];
    page.on('download', download => downloads.push(download));
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
    const finalAnswer = page.getByText(status.final, {exact: true});
    for (let attempt = 0; attempt < 200; attempt++) {
      status = await (await page.request.get(origin + '/grade/status')).json();
      assert.equal(status.failure, null, 'provider oracle must succeed');
      if (await finalAnswer.isVisible()) break;
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    await finalAnswer.waitFor();
    status = await (await page.request.get(origin + '/grade/status')).json();
    assert.equal(status.failure, null, 'provider oracle must succeed');
    assert.equal(status.requests, 4, 'real write/offer/result sequence must finish once');
    assert.equal(status.returned, 3, 'three tool results must be recorded');
    assert.equal(status.offers, 1, 'exactly one successful offer must be retained');
    assert.equal(transfers.length, 0, 'rendering an offer must not transfer it');
    await page.locator('.pane-tab[data-pane="tools"]').click();
    const button = page.getByRole('button', {name: 'Download ' + status.filename, exact: true});
    await button.waitFor();
    assert.equal(status.failedOffers, 1, 'one failed offer must be recorded without metadata');
    const failedCard = page.locator('[data-kind="tool_result"][data-status="error"]');
    assert.equal(await failedCard.count(), 1, 'one real failed tool result is visible');
    assert.equal(await failedCard.locator('.work-product-links').count(), 0,
      'failed offer must not have a download control');
    assert.equal(await page.locator('.work-product-links').count(), 1, 'only successful offer creates control');
    const failedResultID = await failedCard.getAttribute('data-part-id');
    let failedCallID, failureText;
    async function checkFailedPair() {
      const result = page.locator(`[data-part-id="${failedResultID}"]`);
      assert.equal(await result.locator(':scope > .tool-state:not(.job-state)').innerText(), 'Error reported',
        'real failed offer is explicitly reported as an error');
      const details = result.locator(':scope > details');
      if (!(await details.evaluate(node => node.open))) await details.locator('summary').click();
      const raw = await details.locator('pre').innerText();
      assert(raw.trim().length > 0, 'failed result exposes nonempty complete received payload');
      if (failureText === undefined) failureText = raw;
      else assert.equal(raw, failureText, 'replay preserves complete failure payload');
      const link = result.getByRole('button', {name: 'View call', exact: true});
      assert.equal(await link.count(), 1, 'failed result retains its actual call relationship');
      await link.click();
      const id = await page.locator('[data-kind="tool_call"]:focus').getAttribute('data-part-id');
      assert(id, 'failed-result navigation focuses an identified call');
      const call = page.locator(`[data-part-id="${id}"]`);
      if (failedCallID === undefined) failedCallID = id;
      else assert.equal(id, failedCallID, 'replay failure navigation preserves exact call identity');
      assert.equal(await call.locator(':scope > .tool-state:not(.job-state)').innerText(), 'Error reported',
        'failed call is explicitly reported as an error');
      const callDetails = call.locator(':scope > details');
      if (!(await callDetails.evaluate(node => node.open))) await callDetails.locator('summary').click();
      const input = await callDetails.locator('pre').innerText();
      assert(input.includes('offer_work_product') && input.includes(status.filename + '.missing'),
        'failed result pairs with actual missing-file request, not successful offer');
      await call.getByRole('button', {name: 'View latest result', exact: true}).click();
      assert.equal(await page.locator('[data-kind="tool_result"]:focus').getAttribute('data-part-id'), failedResultID,
        'failed call navigation returns to its exact error result');
    }
    await checkFailedPair();
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
    const beforeReplay = transfers.length;
    await page.reload();
    await page.locator('.pane-tab[data-pane="agents"]').click();
    await page.locator('#agent-tree [data-agent="candidate"]').click();
    await page.locator('.pane-tab[data-pane="tools"]').click();
    await button.waitFor();
    status = await (await page.request.get(origin + '/grade/status')).json();
    assert.equal(status.requests, 4, 'replay must not rerun model or tools');
    await checkFailedPair();
    assert.equal(transfers.length, beforeReplay, 'replay must not transfer a retained offer');
    status = await (await page.request.post(origin + '/grade/boundary-file')).json();
    assert.equal(status.bytes, 64 * 1024 * 1024, 'grader-owned size boundary');
    const boundaryAwaited = page.waitForEvent('download').catch(error => {
      throw new Error('exact-64mib-browser-hash: allowed boundary download did not complete', {cause:error});
    });
    await button.click();
    const boundaryDownload = await boundaryAwaited;
    const boundaryBytes = fs.readFileSync(await boundaryDownload.path());
    assert.equal(boundaryBytes.length, 64 * 1024 * 1024, 'browser accepts exact 64 MiB');
    assert.equal(crypto.createHash('sha256').update(boundaryBytes).digest('hex'), status.hash,
      'exact 64 MiB browser download hash');
    await page.request.post(origin + '/grade/oversize-file');
    const oversized = await page.request.get(route, {headers});
    assert.equal(oversized.status(), 404, 'HTTP refuses click-time growth above 64 MiB');
    assert.equal(oversized.headers()['content-disposition'], undefined,
      'refusal must not masquerade as an attachment');
    const beforeRefusal = downloads.length;
    await button.click();
    await page.locator('.work-product-links [role="status"]').filter({hasText: 'Download failed:'}).waitFor();
    await button.waitFor();
    assert.equal(downloads.length, beforeRefusal, 'oversized refusal must not save a file');
    await page.request.post(origin + '/grade/change-file');
    async function waitTransfer(field) {
      for (let attempt = 0; attempt < 200; attempt++) {
        const state = await (await page.request.get(origin + '/grade/transfer-state')).json();
        if (state?.[field]) return state;
        await new Promise(resolve => setTimeout(resolve, 50));
      }
      assert.fail('HTTP transfer did not reach ' + field);
    }
    for (const action of ['cancel', 'selection-replay']) {
      assert.equal((await page.request.post(origin + '/grade/hold-next-transfer')).status(), 200);
      const beforeCancel = downloads.length;
      await button.click();
      await waitTransfer('entered');
      if (action === 'cancel') {
        await page.getByRole('button', {name:'Cancel download', exact:true}).click();
        await page.getByText('Download cancelled.', {exact:true}).waitFor();
      } else {
        await page.locator('.pane-tab[data-pane="agents"]').click();
        // The viewer can be unnamed; use its actual public tree address.
        const address = await page.locator('#agent-tree [data-agent]').evaluateAll(nodes =>
          nodes.map(node => node.dataset.agent).find(value => value !== 'candidate'));
        assert.notEqual(address, undefined, 'viewer tree entry exists');
        await page.locator('#agent-tree [data-agent=' + JSON.stringify(address) + ']').click();
      }
      const cancelled = await waitTransfer('cancelled');
      assert.equal(cancelled.done, false, action + ': request must wait for copy cleanup');
      await page.request.post(origin + '/grade/release-transfer');
      await waitTransfer('done');
      assert.equal(downloads.length, beforeCancel, action + ': partial bytes must not save');
      if (action === 'selection-replay') {
        await page.locator('#agent-tree [data-agent="candidate"]').click();
        await page.locator('.pane-tab[data-pane="tools"]').click();
        await button.waitFor();
      }
    }
    await page.request.post(origin + '/grade/replace-owner');
    assert.equal((await page.request.get(route, {headers})).status(), 404, 'replacement Agent must not revive offer');
    assert.deepEqual(errors, []);
    console.log(JSON.stringify({contract:'artifact-browser-work-product-v1', passed:true,
      checks:['real-model-tool-sequence','click-only','exact-current-bytes','filename','origin-header','no-path-query','foreign-owner','exact-64mib-browser-hash','oversize-http-refused','oversize-browser-refused','replacement-expiry','replay-no-effects','cancel-joins-http','selection-replay-cancels-http','failed-offer-no-download','failed-offer-label-payload-and-pair-replay'],
      ungraded:'Full artifact rubric and chapter points are not implied.'}));
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
