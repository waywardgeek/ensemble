// Run with a disposable --artifact-renderer-fixture directory. No inference.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const directory = process.argv[2];
(async () => {
    const origin = fs.readFileSync(path.join(directory, 'url'), 'utf8').trim();
    const url = new URL(origin);
    assert.equal(url.hostname, '127.0.0.1');
    assert(!['8088', '8089', '8090'].includes(url.port));
    const browser = await chromium.launch({channel: 'chrome', headless: true});
    try {
        const page = await browser.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto(origin);
        await page.waitForFunction(() => settings !== null && !selectionPending);
        const text = '[Auto-recalled memories]\n> Exact supplied text <img src=x>\n> Second line.';
        await page.evaluate(text => {
            socket.onclose = null; socket.close(); socket = null;
            window.recallFrame = {type: 'part_final', agent_id: agentID,
                part_id: `${agentID}:e700:p0`, seq: 700,
                part: {type: 'recall', role: 'human', text, hashes: ['not display text']}};
            handle(recallFrame);
        }, text);
        const card = page.locator('#transcript article[data-kind="recall"]');
        await card.waitFor({state: 'visible'});
        assert.equal(await card.count(), 1);
        await card.locator('summary').click();
        assert.equal(await card.locator('pre').textContent(), text);
        assert.equal(await card.locator('img').count(), 0, 'Recall remains literal data');
        await page.evaluate(() => handle({type: 'replay', agent_id: agentID,
            data: {events: [recallFrame], partials: [], agents: []}}));
        assert.equal(await card.count(), 1);
        assert.equal(await card.locator('pre').textContent(), text, 'Replay preserves content');
        await page.evaluate(() => handle({type: 'display_window', agent_id: agentID, part_ids: []}));
        assert.equal(await card.count(), 0, 'Recall obeys display window');
        assert.deepEqual(errors, []);
        console.log('PASS recall browser: live exact payload, literal markup, replay, bounded eviction');
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; })
    .finally(() => fs.writeFileSync(path.join(directory, 'stop'), '\n'));
