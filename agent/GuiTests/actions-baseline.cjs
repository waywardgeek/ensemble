'use strict';
// Baseline on the real embedded page with synthetic frames at its public wire
// handler. This checks presentation, not tool execution, persistence or a grade.
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
const origin = new URL(base);
assert.equal(origin.hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(origin.port));

(async () => {
    const browser = await chromium.launch({
        executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
        headless: true
    });
    try {
        const page = await browser.newPage({ viewport: { width: 390, height: 844 } });
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto(base);
        await page.waitForFunction(() => settings !== null && !selectionPending);
        await page.locator('.pane-tab[data-pane="tools"]').click();
        const cases = [
            { id: 'command', name: 'run_command', input: { command: 'printf complete' }, result: 'complete\nexit_code: 0' },
            { id: 'edit', name: 'edit_file', input: { path: 'sample.txt', old_text: 'before', new_text: 'after' }, result: 'File edited.' },
            { id: 'search', name: 'search_files', input: { path: '.', pattern: 'needle' }, result: 'sample.txt:3:needle' },
            { id: 'failure', name: 'read_file', input: { path: 'missing.txt' }, result: 'File not found.', is_error: true },
            { id: 'running', name: 'run_command', input: { command: 'long-running-test' }, result: 'job 987654, status: running\npartial output' },
            { id: 'report', name: 'write_file', input: { path: 'report.pdf', content: '(synthetic report placeholder)' }, result: 'Report written: report.pdf' }
        ];
        await page.evaluate(cases => {
            for (const item of cases) {
                const call = `actions-baseline-${item.id}-call`;
                handle({ type: 'tool_dispatched', part_id: call, name: item.name, input: item.input });
                handle({ type: 'tool_finished', part_id: `actions-baseline-${item.id}-result`,
                    tool_part_id: call, result: item.result, is_error: !!item.is_error });
            }
        }, cases);
        for (const item of cases) {
            const call = page.locator(`[data-part-id="actions-baseline-${item.id}-call"]`);
            const result = page.locator(`[data-part-id="actions-baseline-${item.id}-result"]`);
            assert.equal(await call.count(), 1);
            assert.equal(await result.count(), 1);
            assert.equal(await result.locator('.card-heading strong').textContent(), item.name);
            assert.equal(await result.locator('details > pre.tool-output, details > pre.search-results').textContent(), item.result);
            assert.equal(await result.getAttribute('data-status'), item.is_error ? 'error' : 'finished');
        }
        assert.match(await page.locator('[data-part-id="actions-baseline-edit-result"] .edit-change').textContent(), /- before\n\+ after/);
        assert.equal(await page.locator('[data-part-id="actions-baseline-report-result"] a[download]').count(), 0);
        console.log('PASS six synthetic Actions pairs, raw results, edit detail and no invented report download');
        console.log('BASELINE LIMIT: observation is labeled finished even when its text reports a running Job');

        const search = page.locator('[data-part-id="actions-baseline-search-result"]');
        await search.locator('summary').click();
        await page.evaluate(() => render(cards.get('actions-baseline-search-result')));
        assert.equal(await search.locator('details').getAttribute('open'), '');
        assert(await page.evaluate(() => toolTranscript.scrollWidth <= toolTranscript.clientWidth + 1));
        assert.deepEqual(errors, []);
        console.log('PASS expanded Actions detail survives render and narrow pane contains its content');
    } finally {
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
