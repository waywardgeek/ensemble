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
        await page.evaluate(() => handle({ type: 'part_final', part_id: 'state-call',
            part: { type: 'tool_call', name: 'run_command', args: { command: 'synthetic' } } }));
        const call = page.locator('[data-part-id="state-call"]');
        assert.equal(await call.locator('.tool-state').textContent(), 'Requested');
        await page.evaluate(() => handle({ type: 'tool_dispatched', part_id: 'state-call',
            name: 'run_command', input: { command: 'synthetic' } }));
        assert.equal(await call.locator('.tool-state').textContent(), 'Dispatched · awaiting result');
        assert.equal(await call.getAttribute('data-status'), 'started');
        const cases = [
            { id: 'running', job: { handle: 81, status: 'running' }, label: 'running' },
            { id: 'done', job: { handle: 81, status: 'done', exit_code: 0 }, label: 'finished · exit code 0' },
            { id: 'nonzero', job: { handle: 82, status: 'done', exit_code: 7 }, label: 'finished · exit code 7', error: true },
            { id: 'done-unknown-exit', job: { handle: 83, status: 'done' }, label: 'finished' },
            { id: 'killed', job: { handle: 84, status: 'killed' }, label: 'killed', error: true },
            { id: 'unknown', job: { handle: 85, status: 'toString' }, label: 'unavailable' },
            { id: 'invalid-handle', job: { handle: '81', status: 'done' }, label: null },
            { id: 'no-metadata', job: null, label: null }
        ];
        for (const item of cases) {
            await page.evaluate(item => handle({ type: 'tool_finished', part_id: 'state-' + item.id,
                tool_part_id: 'state-call', job: item.job, is_error: !!item.error,
                result: 'Untrusted prose says: job 999, status: done, exit_code: 0' }), item);
            const result = page.locator(`[data-part-id="state-${item.id}"]`);
            assert.equal(await result.locator('.tool-state:not(.job-state)').textContent(), item.error ? 'Error reported' : 'Result returned', 'observation is not execution completion');
            assert.equal(await result.getAttribute('data-status'), item.error ? 'error' : 'finished');
            if (item.label === null) assert.equal(await result.locator('.job-state').count(), 0);
            else assert.equal(await result.locator('.job-state').textContent(), `Job ${item.job.handle} · state when observed: ${item.label}`, 'recorded Job state must be preserved');
        }
        // The later observation of Job81 does not rewrite its earlier snapshot.
        const running = page.locator('[data-part-id="state-running"]');
        assert.equal(await running.locator('.job-state').textContent(), 'Job 81 · state when observed: running');
        await page.evaluate(() => render(cards.get('state-running')));
        assert.equal(await running.locator('.job-state').textContent(), 'Job 81 · state when observed: running');
        assert.equal(await call.locator('.tool-state').textContent(), 'Result returned');
        assert.match(await call.locator('.card-heading strong').getAttribute('title'), /Result returned/);
        assert(await page.evaluate(() => toolTranscript.scrollWidth <= toolTranscript.clientWidth + 1));
        assert.deepEqual(errors, []);
        console.log('PASS requested/dispatched/result/error labels preserve machine states');
        console.log('PASS typed recorded running/done/killed/exit/unknown states; prose cannot invent state');
        console.log('PASS later observation and rerender preserve earlier recorded state; narrow containment');
    } finally {
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
