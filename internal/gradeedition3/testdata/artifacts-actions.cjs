'use strict';
// Real public page, model turns and custom-tool Jobs; no application globals.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE);
const assert = require('node:assert/strict');
const base = process.env.ARTIFACT_GUI_URL;
const origin = new URL(base).origin;
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
(async () => {
  const browser = await chromium.launch({headless: true, channel: 'chrome'});
  try {
    const page = await browser.newPage({viewport: {width: 390, height: 844}, locale: 'en-US', timezoneId: 'Pacific/Honolulu'});
    page.setDefaultTimeout(20000);
    const errors = [], transfers = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('request', request => { if (request.url().includes('/work-products/download')) transfers.push(request.url()); });
    async function statusUntil(predicate) {
      for (let attempt = 0; attempt < 200; attempt++) {
        const state = await (await page.request.get(origin + '/grade/status')).json();
        assert.equal(state.failure, null, 'Actions provider oracle');
        if (predicate(state)) return state;
        await new Promise(resolve => setTimeout(resolve, 100));
      }
      throw new Error('Actions state did not reach expected boundary');
    }
    async function checkRecordedTiming(card, fact) {
      assert(fact && typeof fact.recorded === 'string', 'canonical result has recorded time');
      const stamp = card.locator(':scope > time.recorded-time');
      assert.equal(await stamp.getAttribute('datetime'), fact.recorded,
        'display time equals canonical recorded fact, not browser arrival');
      const local = await page.evaluate(time => new Date(time).toLocaleString(), fact.recorded);
      assert.equal(await stamp.innerText(), 'Recorded ' + local, 'recorded time uses local timezone');
      assert((await stamp.getAttribute('title')).includes(fact.recorded), 'exact timestamp remains inspectable');
      const duration = card.locator(':scope > .observation-duration');
      if (fact.milliseconds === null) {
        assert.equal(await duration.count(), 0, 'unknown observation duration is not invented');
      } else {
        assert(fact.milliseconds >= 0, 'canonical observation duration is nonnegative');
        assert.equal(await duration.innerText(), 'Observation: ' +
          (fact.milliseconds / 1000).toLocaleString('en-US', {maximumFractionDigits: 3}) + ' s',
          'display duration equals recorded observation');
        assert.equal(await duration.getAttribute('title'), 'Dispatch to returned observation: ' +
          fact.milliseconds + ' ms; not background Job runtime', 'exact observation duration remains inspectable');
      }
    }
    await page.goto(base);
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    const initial = await statusUntil(() => true);
    await page.locator('#prompt').fill('exercise Actions ' + initial.nonce);
    await page.locator('#send').click();
    await statusUntil(state => state.entered && state.running === 1);
    await page.locator('.pane-tab[data-pane="tools"]').click();
    const running = page.locator('[data-kind="tool_result"]');
    await running.waitFor();
    assert((await running.locator('.job-state').innerText()).includes('state when observed: running'),
      'real-running-observation must display running, not finished');
    assert.equal(await running.locator('.tool-state:not(.job-state)').innerText(), 'Result returned', 'returned observation is not completed work');
    const resultID = await running.getAttribute('data-part-id');
    await running.getByRole('button', {name: 'View call', exact: true}).click();
    const call = page.locator('[data-kind="tool_call"]:focus');
    assert.equal(await call.count(), 1, 'View call focuses a real call card');
    const callID = await call.getAttribute('data-part-id');
    assert((await call.innerText()).includes('held_action'), 'paired call is the actual held tool');
    await call.getByRole('button', {name: 'View latest result', exact: true}).click();
    assert.equal(await page.locator('[data-kind="tool_result"]:focus').getAttribute('data-part-id'), resultID,
      'call navigation returns to its actual observation');
    assert.equal(await page.locator('.work-product-links').count(), 0, 'running result has no offer control');
    const observed = (await statusUntil(state => state.running === 1)).observations[0];
    assert(observed.milliseconds > 0, 'admitted real observation has measured duration');
    await checkRecordedTiming(running, observed);
    const details = page.locator(`[data-part-id="${resultID}"]`).locator(':scope > details');
    const summary = details.locator(':scope > summary');
    await summary.click();
    await summary.focus();
    const readerNode = await summary.elementHandle();
    const receivedText = await details.locator('pre').textContent();
    assert(await details.evaluate(node => node.open), 'reader opens actual received payload');
    await page.request.post(origin + '/grade/release-action');
    const completed = await statusUntil(state => state.requests === 3 && state.done === 1);
    assert.equal(completed.executions, 1);
    const oldResult = page.locator(`[data-part-id="${resultID}"]`);
    assert((await oldResult.locator('.job-state').innerText()).includes('state when observed: running'),
      'later completion must not rewrite the earlier observation');
    const terminal = page.locator('[data-kind="tool_result"]').filter({has: page.locator('.job-state', {hasText: 'state when observed: finished'})});
    await terminal.waitFor();
    assert(await readerNode.evaluate(node => node.isConnected && document.activeElement === node),
      'later-observation preserves the actual focused reader control');
    assert(await details.evaluate(node => node.open), 'later-observation preserves expanded payload');
    assert.equal(await details.locator('pre').textContent(), receivedText,
      'later-observation preserves complete original payload');
    await readerNode.dispose();
    assert.deepEqual(completed.observations[0], observed, 'later completion does not rewrite recorded timing');
    await checkRecordedTiming(oldResult, observed);
    await checkRecordedTiming(terminal, completed.observations[1]);
    assert.equal(await terminal.locator('.tool-state:not(.job-state)').innerText(), 'Result returned');
    assert((await terminal.textContent()).includes(completed.output), 'full pathname result is retained');
    assert.equal(await page.locator('.work-product-links').count(), 0, 'pathname prose grants no work-product control');
    assert.equal(transfers.length, 0, 'rendering pathname output transfers nothing');
    await page.reload();
    await page.locator('.pane-tab[data-pane="tools"]').click();
    await page.locator(`[data-part-id="${callID}"]`).waitFor();
    assert((await page.locator(`[data-part-id="${resultID}"] .job-state`).innerText()).includes('state when observed: running'),
      'replay preserves running snapshot');
    const replay = await statusUntil(() => true);
    assert.equal(replay.executions, 1, 'replay cannot reexecute the callback');
    assert.equal(replay.requests, 3, 'replay cannot issue model requests');
    assert.deepEqual(replay.observations, completed.observations, 'replay keeps canonical timing');
    await checkRecordedTiming(page.locator(`[data-part-id="${resultID}"]`), observed);
    await checkRecordedTiming(page.locator('[data-kind="tool_result"]').filter({has:
      page.locator('.job-state', {hasText: 'state when observed: finished'})}), completed.observations[1]);
    assert.deepEqual(errors, []);
    console.log(JSON.stringify({contract:'artifact-actions-v1', passed:true, checks:[
      'real-running-observation', 'call-result-navigation', 'historical-running-preserved',
      'terminal-pathname-not-grant', 'replay-no-execution', 'focused-expanded-reader-preserved', 'canonical-recorded-time-and-duration', 'timing-survives-replay'
    ]}));
  } finally { await browser.close(); }
})().catch(error => { console.error(error.stack); process.exitCode = 1; });
