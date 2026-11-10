'use strict';
// Real shipped page; synthetic interleaved frames isolate browser ownership.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const directory = process.argv[2];
const origin = fs.readFileSync(path.join(directory, 'url'), 'utf8').trim();
(async () => {
    const browser = await chromium.launch({ channel: 'chrome', headless: true });
    try {
        const page = await browser.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.route('**/jobs/window?*', route => route.fulfill({ status: 404, body: '' }));
        await page.goto(origin);
        await page.waitForFunction(() => settings !== null && !selectionPending && jobSessions !== null);
        await page.evaluate(() => {
            socket.onclose = null;
            socket.close();
            socket = null;
            window.replayAgent = (agent, events = []) => handle({ type: 'replay', agent_id: agent,
                data: { events, partials: [], agents: [] } });
            window.result = (agent, id) => {
                handle({ type: 'tool_dispatched', agent_id: agent, part_id: id + '-call', name: 'run_command', input: {} });
                handle({ type: 'tool_finished', agent_id: agent, part_id: id, tool_part_id: id + '-call',
                    result: 'output for ' + agent, job: { handle: 77, status: 'done',
                        origin_call_id: 'same-call-id', origin_tool: 'run_command' } });
            };
            replayAgent('parent');
            result('parent', 'parent-result');
            window.oldCard = cards.get('parent-result');
            replayAgent('parent/child');
            result('parent/child', 'child-result');
        });
        await page.waitForTimeout(150);
        assert.equal(await page.locator('[data-part-id="parent-result"]').count(), 0,
            'Job reset resurrected the parent result in the child pane');
        assert.equal(await page.locator('.job-session').count(), 1);
        await page.evaluate(() => {
            queueCardRender(oldCard);
            render(oldCard);
            result('parent', 'late-parent');
        });
        await page.waitForTimeout(100);
        assert.equal(await page.locator('[data-part-id^="parent"], [data-part-id^="late-parent"]').count(), 0);
        // Same-agent replay also invalidates old objects, even with reused IDs.
        await page.evaluate(() => {
            window.oldChild = cards.get('child-result');
            replayAgent('parent/child');
            result('parent/child', 'child-result');
            oldChild.text = 'STALE';
            queueCardRender(oldChild);
        });
        await page.waitForTimeout(100);
        assert.equal(await page.locator('[data-part-id="child-result"]').count(), 1);
        assert.equal(await page.locator('#tool-transcript').textContent().then(text => text.includes('STALE')), false);
        await page.evaluate(() => {
            const retained = [];
            for (let i = 0; i < 350; i++) {
                const id = 'new-' + i;
                handle({ type: 'part_final', agent_id: 'parent/child', part_id: id,
                    part: { type: 'text', role: 'human', text: id } });
                retained.push(id);
                if (retained.length > 100) retained.shift();
                handle({ type: 'display_window', agent_id: 'parent/child', part_ids: retained });
            }
            // A different Agent's window cannot evict this Agent's cards.
            handle({ type: 'display_window', agent_id: 'parent', part_ids: [] });
        });
        await page.waitForTimeout(100);
        assert.equal(await page.evaluate(() => cards.size), 100);
        assert.equal(await page.locator('article[data-part-id]').count(), 100);
        assert.equal(await page.locator('.job-session').count(), 0, 'Evicted job session retained polling/DOM');
        assert.equal(await page.locator('[data-part-id="new-249"]').count(), 0);
        assert.equal(await page.locator('[data-part-id="new-349"]').count(), 1);
        assert.deepEqual(errors, []);
        console.log('PASS browser: parent/child isolation, same-agent replay, stale renders, live eviction and job cleanup');
    } finally {
        await browser.close();
        fs.writeFileSync(path.join(directory, 'stop'), 'stop');
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
