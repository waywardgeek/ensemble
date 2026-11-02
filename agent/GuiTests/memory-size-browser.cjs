'use strict';
const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const bands = ['conversation', 'soul', 'memory', 'archive64', 'archive8', 'session'];
(async () => {
    const browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
    try {
        const page = await browser.newPage();
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto('http://127.0.0.1:18202/assistant/');
        await page.waitForFunction(() => !selectionPending && agentID === 'captain' && !!settings.memory);
        assert.equal(await page.locator('#memory-settings-form input[type="number"]').count(), 6);
        assert.equal(await page.locator('#memory-settings-form [id*="watermark"]').count(), 0);
        await page.evaluate(bands => {
            for (const [index, band] of bands.entries()) {
                document.querySelector(`#memory-${band}-budget`).value = String(6000 + index);
                document.querySelector(`#memory-${band}-enabled`).checked = false;
            }
            document.querySelector('#memory-settings-form').requestSubmit();
        }, bands);
        await page.waitForFunction(() => settings.memory.session.budget === 6005 && settings.memory.session.disabled);
        const policy = await page.evaluate(() => settings.memory);
        for (const [index, band] of bands.entries()) {
            assert.deepEqual(Object.keys(policy[band]).sort(), ['budget', 'disabled']);
            assert.deepEqual(policy[band], {budget: 6000 + index, disabled: true});
        }
        await page.evaluate(() => {
            document.querySelector('#memory-session-budget').value = '9999';
            document.querySelector('#memory-settings-reload').click();
        });
        await page.waitForFunction(() => document.querySelector('#memory-session-budget').value === '6005');
        await page.reload();
        await page.waitForFunction(() => !selectionPending && settings.memory?.session?.budget === 6005);
        for (const [index, band] of bands.entries()) {
            assert.equal(await page.locator(`#memory-${band}-budget`).inputValue(), String(6000 + index));
            assert.equal(await page.locator(`#memory-${band}-enabled`).isChecked(), false);
        }
        assert.deepEqual(errors, []);
        console.log('PASS: six size fields, no watermark controls, acknowledged save, explicit reload, browser refresh');
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
