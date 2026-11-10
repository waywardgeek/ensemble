'use strict';
// Real page, recorded MCP results with no assistant call card, owned fake HTTP.
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
        let status = 'running', content = 'task admitted\n', unavailable = false;
        await page.route('**/jobs/window?*', route => {
            if (unavailable) return route.fulfill({ status: 404, body: '' });
            const bytes = Buffer.from(content);
            return route.fulfill({ json: { handle: 32, origin_call_id: 'spawn', origin_tool: 'spawn_sub_agent',
                status, start: 0, end: bytes.length, rolled_out: false, encoding: 'base64', data: bytes.toString('base64') } });
        });
        await page.goto(origin);
        await page.waitForFunction(() => settings !== null && !selectionPending && jobSessions !== null);
        await page.evaluate(() => {
            socket.onclose = null;
            socket.close();
            socket = null;
            handle({ type: 'replay', agent_id: 'parent', data: { events: [], partials: [], agents: [] } });
            window.poll = (id, name = 'wait_for_job', status = 'running', error = false) => handle({
                type: 'tool_finished', agent_id: 'parent', part_id: id, name, tool_part_id: null,
                result: 'recorded ' + id, is_error: error,
                job: { handle: 32, origin_call_id: 'spawn', origin_tool: 'spawn_sub_agent', status }
            });
            poll('admitted', 'spawn_sub_agent');
        });
        await page.waitForFunction(() => document.querySelector('.job-observations [data-part-id="admitted"]'));
        await page.evaluate(() => {
            window.originalOutput = document.querySelector('.job-session');
            const marker = document.createElement('div');
            marker.id = 'after-output';
            toolTranscript.append(marker);
            poll('wait-1');
        });
        assert(await page.evaluate(() => toolTranscript.lastElementChild === originalOutput));
        assert.equal(await page.locator('#tool-transcript > article[data-kind="tool_result"]').count(), 0);
        await page.evaluate(() => {
            for (let i = 2; i < 150; i++) poll('wait-' + i);
            handle({ type: 'display_window', agent_id: 'parent', part_ids: ['wait-148', 'wait-149'] });
        });
        assert.equal(await page.locator('.job-session').count(), 1);
        assert(await page.evaluate(() => document.querySelector('.job-session') === originalOutput));
        assert.equal(await page.locator('.job-observations article').count(), 2);
        assert.equal(await page.locator('#tool-transcript > article[data-kind="tool_result"]').count(), 0);
        status = 'done';
        content += 'task completed\n';
        await page.evaluate(() => poll('complete', 'wait_for_job', 'done'));
        await page.waitForFunction(() => document.querySelector('.terminal-output').textContent.includes('task completed'));
        assert(await page.evaluate(() => document.querySelector('.job-session') === originalOutput));
        assert.equal(await page.locator('.job-observations').getAttribute('open'), null);
        await page.locator('.job-observations > summary').click();
        assert(await page.locator('[data-part-id="complete"]').isVisible());
        await page.evaluate(() => poll('error', 'wait_for_job', 'done', true));
        assert(await page.locator('#tool-transcript > [data-part-id="error"]').isVisible());
        // If readback fails, restore literal recorded results, never lose output.
        unavailable = true;
        await page.evaluate(() => poll('unavailable', 'wait_for_job', 'done'));
        await page.waitForFunction(() => document.querySelector('#tool-transcript > [data-part-id="unavailable"]'));
        await page.evaluate(() => {
            handle({ type: 'replay', agent_id: 'parent/child', data: { events: [], partials: [], agents: [] } });
            poll('late-parent');
        });
        await page.waitForTimeout(100);
        assert.equal(await page.locator('.job-session, #tool-transcript article').count(), 0);
        assert.deepEqual(errors, []);
        console.log('PASS workflow Jobs: no call card required, same output node moves, observations grouped, eviction, completion, refusal and Agent isolation');
    } finally {
        await browser.close();
        fs.writeFileSync(path.join(directory, 'stop'), 'stop');
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
