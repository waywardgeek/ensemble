// Usage: node agent/GuiTests/error-cards-browser.cjs <artifact-renderer-fixture-directory>
// Disposable loopback host only; real handler and DOM, no inference.
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
    const browser = await chromium.launch({ channel: 'chrome', headless: true });
    try {
        const page = await browser.newPage();
        const exceptions = [];
        page.on('pageerror', error => exceptions.push(error.message));
        await page.goto(origin);
        await page.waitForFunction(() => settings !== null && !selectionPending);
        await page.evaluate(() => {
            socket.onclose = null; socket.close(); socket = null;
            window.errorSpeech = [];
            speakText = (id, text) => errorSpeech.push(text);
            browserSpeech.select({ enabled: true });
            window.failure = { type: 'lifecycle', agent_id: agentID, seq: 700,
                part_id: `${agentID}:e700:error`, kind: 'error_occurred',
                message: 'Provider rejected the request.\nLiteral <img src=x onerror=bad()>',
                recorded_at: '2026-11-14T00:00:00Z' };
            handle(failure); handle(failure);
            handle({ type: 'lifecycle', agent_id: agentID, kind: 'turn_interrupted', message: 'Later notice' });
            handle({ type: 'display_window', agent_id: agentID, part_ids: [failure.part_id] });
        });
        const card = page.locator('#transcript article[data-kind="error"]');
        await card.waitFor({ state: 'visible' });
        assert.equal(await card.count(), 1);
        assert.equal(await card.locator('.card-heading strong').textContent(), 'Error');
        assert.equal(await card.locator('.message').textContent(),
            'Provider rejected the request.\nLiteral <img src=x onerror=bad()>');
        assert.equal(await card.locator('img, details').count(), 0);
        assert.equal(await page.evaluate(() => errorSpeech.length), 1);
        await page.evaluate(() => handle({ type: 'replay', agent_id: agentID,
            data: { events: [failure], partials: [], agents: [] } }));
        assert.equal(await card.count(), 1, 'Replay restores error card');
        assert.equal(await page.evaluate(() => errorSpeech.length), 1, 'Replay is silent');
        await page.evaluate(() => handle({ type: 'error', message: 'Control refusal' }));
        assert.equal(await card.count(), 2);
        await page.evaluate(() => handle({ type: 'display_window', agent_id: agentID, part_ids: [] }));
        assert.equal(await card.count(), 0);
        assert.deepEqual(exceptions, []);
        console.log('PASS browser errors: literal visible body, deduplication, notice survival, silent replay, controls and eviction');
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; })
    .finally(() => fs.writeFileSync(path.join(directory, 'stop'), 'stop'));
