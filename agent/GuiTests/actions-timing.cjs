'use strict';
// Actual page handler and rendering; synthetic wire facts isolate presentation.
// These snapshots are recorded observations, not a live Job monitor.
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
(async () => {
    const browser = await chromium.launch({
        executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true
    });
    try {
        const page = await browser.newPage({ viewport: { width: 390, height: 844 } });
        // Explicit local candidate override for sensitivity checks only.
        if (process.env.GUI_ACTIONS_SOURCE) {
            const body = require('node:fs').readFileSync(process.env.GUI_ACTIONS_SOURCE, 'utf8');
            await page.route('**/app.js', route => route.fulfill({ status: 200, contentType: 'text/javascript', body }));
        }
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto(base);
        await page.waitForFunction(() => settings !== null && !selectionPending);
        await page.locator('.pane-tab[data-pane="tools"]').click();
        const recorded = '2026-11-05T09:10:11.0000000-08:00';
        await page.evaluate(recorded => {
            handle({ type: 'part_delta', part_id: 'timing-text', kind: 'text', chunk: 'draft' });
            handle({ type: 'part_final', part_id: 'timing-call', recorded_at: recorded,
                part: { type: 'tool_call', name: 'run_command', args: {} } });
            handle({ type: 'tool_dispatched', part_id: 'timing-call', recorded_at: '2026-11-05T17:10:12Z', name: 'run_command', input: {} });
            handle({ type: 'tool_finished', part_id: 'timing-result', tool_part_id: 'timing-call', recorded_at: recorded,
                observation_ms: 123.4567, job: { handle: 91, status: 'running' }, result: 'still running' });
            handle({ type: 'tool_finished', part_id: 'timing-legacy', result: 'legacy' });
            handle({ type: 'tool_finished', part_id: 'timing-invalid', recorded_at: 'invalid', observation_ms: -1, result: 'invalid metadata' });
        }, recorded);
        assert.equal(await page.locator('[data-part-id="timing-text"] .recorded-time').count(), 0, 'draft must not acquire browser time');
        const stamp = page.locator('[data-part-id="timing-result"] .recorded-time');
        assert.equal(await stamp.getAttribute('datetime'), recorded, 'original timestamp must survive display');
        assert.equal(await stamp.textContent(), await page.evaluate(recorded => 'Recorded ' + new Date(recorded).toLocaleString(), recorded));
        assert((await stamp.getAttribute('title')).includes(recorded));
        assert.equal(await page.locator('[data-part-id="timing-call"] .recorded-time').getAttribute('datetime'), recorded, 'dispatch must not overwrite original call time');
        const duration = page.locator('[data-part-id="timing-result"] .observation-duration');
        assert.equal(await duration.textContent(), 'Observation: 0.123 s');
        assert((await duration.getAttribute('title')).includes('not background Job runtime'));
        assert.equal(await page.locator('[data-part-id="timing-legacy"] .recorded-time, [data-part-id="timing-legacy"] .observation-duration').count(), 0);
        assert.equal(await page.locator('[data-part-id="timing-invalid"] .recorded-time, [data-part-id="timing-invalid"] .observation-duration').count(), 0);
        await page.evaluate(() => render(cards.get('timing-result')));
        assert.equal(await duration.textContent(), 'Observation: 0.123 s', 'rerender must not advance recorded duration');
        assert.equal(await stamp.getAttribute('datetime'), recorded);
        assert.deepEqual(errors, []);
        console.log('PASS local recorded timestamp with exact value; original call time retained; drafts and legacy unknown');
        console.log('PASS fixed observation duration distinct from running Job; invalid metadata refused');
    } finally {
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
