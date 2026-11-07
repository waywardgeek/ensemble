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
        await page.evaluate(() => {
            handle({ type: 'tool_dispatched', part_id: 'pair-call', name: 'run_command', input: { command: 'synthetic' } });
            const call = cards.get('pair-call');
            window.savedSummary = call.node.querySelector('summary');
            call.node.querySelector('details').open = true;
            window.savedSummary.focus();
            handle({ type: 'tool_finished', part_id: 'pair-result', tool_part_id: 'pair-call', result: 'first result',
                job: { handle: 81, status: 'running' } });
        });
        assert(await page.evaluate(() => document.activeElement === window.savedSummary), 'arriving result must retain focused summary');
        assert(await page.locator('[data-part-id="pair-call"] details').evaluate(node => node.open));
        const callLink = page.locator('[data-part-id="pair-call"] [data-action="view-pair"]');
        await callLink.click();
        assert.equal(await page.evaluate(() => document.activeElement.dataset.partId), 'pair-result');
        await page.locator('[data-part-id="pair-result"] [data-action="view-pair"]').click();
        assert.equal(await page.evaluate(() => document.activeElement.dataset.partId), 'pair-call');
        await page.evaluate(() => {
            const result = cards.get('pair-result');
            window.savedDownload = result.node.querySelector('.card-heading [aria-label="Download complete output"]');
            window.savedDownload.disabled = true;
            result.node.querySelector('.artifact-links [role="status"]').textContent = 'test pending';
            render(result);
        });
        assert(await page.evaluate(() => cards.get('pair-result').node.querySelector('.card-heading [aria-label="Download complete output"]') === window.savedDownload
            && window.savedDownload.disabled), 'rerender must retain pending download control');
        assert.equal(await page.locator('[data-part-id="pair-result"] .artifact-links [role="status"]').textContent(), 'test pending');
        await callLink.focus();
        await page.evaluate(() => handle({ type: 'tool_finished', part_id: 'pair-later', tool_part_id: 'pair-call', result: 'later result' }));
        assert(await callLink.evaluate(node => document.activeElement === node), 'latest-result update retains focused pair control');
        await callLink.click();
        assert.equal(await page.evaluate(() => document.activeElement.dataset.partId), 'pair-later', 'call navigates to latest result');
        await page.locator('[data-part-id="pair-result"] [data-action="view-pair"]').click();
        assert.equal(await page.evaluate(() => document.activeElement.dataset.partId), 'pair-call', 'earlier result retains call');
        await page.evaluate(() => {
            handle({ type: 'tool_finished', part_id: 'pair-orphan', tool_part_id: 'absent-call', result: 'unpaired' });
            handle({ type: 'tool_finished', part_id: 'pair-invalid', tool_part_id: 'pair-result', result: 'invalid counterpart' });
        });
        assert.equal(await page.locator('[data-part-id="pair-orphan"] [data-action="view-pair"]').count(), 0);
        assert.equal(await page.locator('[data-part-id="pair-invalid"] [data-action="view-pair"]').count(), 0);
        assert.equal(await page.locator('[data-part-id="pair-invalid"] details > .tool-output').textContent(), 'invalid counterpart');
        await page.evaluate(() => {
            const result = cards.get('pair-result');
            result.node.querySelector('details').open = true;
            const text = result.node.querySelector('.preview .tool-output').firstChild;
            const range = document.createRange();
            range.selectNodeContents(text);
            getSelection().removeAllRanges(); getSelection().addRange(range);
            result.text = 'changed while selected';
            render(result);
        });
        assert.equal(await page.evaluate(() => getSelection().toString()), 'first result');
        assert.equal(await page.locator('[data-part-id="pair-result"] .preview .tool-output').textContent(), 'first result');
        await page.evaluate(() => getSelection().removeAllRanges());
        await page.waitForFunction(() => cards.get('pair-result').node.querySelector('.preview .tool-output').textContent === 'changed while selected');
        assert(await page.evaluate(() => {
            const call = cards.get('pair-call');
            call.node.querySelector('summary').focus();
            toolTranscript.scrollTop = toolTranscript.scrollHeight;
            const before = toolTranscript.scrollTop;
            handle({ type: 'tool_finished', part_id: 'pair-scroll', tool_part_id: 'pair-call', result: 'new output\n'.repeat(80) });
            return Math.abs(toolTranscript.scrollTop - before) <= 1;
        }), 'focused reader must not be pulled to new output');
        assert(await callLink.evaluate(node => node.getBoundingClientRect().height >= 44));
        // A detached control must not navigate to new objects with recycled IDs.
        assert(await page.evaluate(() => {
            const oldCall = cards.get('pair-call');
            const oldButton = oldCall.node.querySelector('[data-action="view-pair"]');
            const oldResult = cards.get('pair-scroll');
            const oldResultButton = oldResult.node.querySelector('[data-action="view-pair"]');
            cards.delete(oldCall.id); cards.delete(oldResult.id);
            oldCall.node.remove(); oldResult.node.remove();
            handle({ type: 'tool_dispatched', part_id: 'pair-call', name: 'search_files', input: {} });
            handle({ type: 'tool_finished', part_id: 'pair-later', tool_part_id: 'pair-call', result: 'new transcript objects' });
            const active = document.activeElement;
            oldButton.click();
            oldResultButton.click();
            return document.activeElement === active;
        }), 'stale pair control must not target reused IDs');
        assert.deepEqual(errors, []);
        console.log('PASS retained focus, expansion, download control and feedback; selection deferral and reader scroll');
        console.log('PASS explicit call/latest-result navigation, orphan refusal and stale-object fencing');
    } finally {
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
