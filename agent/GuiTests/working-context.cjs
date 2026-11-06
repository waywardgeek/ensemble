'use strict';
// Actual embedded page, real Agent and owned provider. DOM only; no injected frames.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const settingsPath = process.env.GUI_CONTEXT_SETTINGS;
assert(settingsPath, 'Supply the owned fixture settings path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
(async () => {
    const browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
    try {
        const page = await browser.newPage({ viewport: { width: 1280, height: 900 } });
        const errors = [];
        page.on('pageerror', error => { errors.push(error.message); console.error('PAGE ERROR:', error.message); });
        await page.goto('http://127.0.0.1:18202/assistant/');
        await page.locator('#model-choice').selectOption('claude-3-haiku-20240307');
        await page.waitForFunction(() => document.querySelector('#state').textContent.includes('context unknown'));
        await page.locator('#settings-open').click();
        await page.locator('#working-context-tokens').fill('100');
        await page.locator('#context-capacity-tokens').fill('200000');
        await page.locator('#save-working-context').click();
        await page.waitForFunction(() => document.querySelector('#notice').textContent === '');
        await page.locator('#settings-close').click();
        await page.locator('#prompt').fill('Measure the current working context');
        await page.locator('#send').click();
        await page.waitForFunction(() => document.querySelector('#state').textContent.includes('31% context'));
        assert.match(await page.locator('#state').getAttribute('title'), /31 input tokens \/ 100 working limit/);
        // Real save failure, controlled only at our owned fixture filesystem.
        const backup = settingsPath + '.test-backup';
        fs.renameSync(settingsPath, backup);
        fs.mkdirSync(settingsPath);
        try {
            await page.locator('#settings-open').click();
            await page.locator('#working-context-tokens').fill('200');
            await page.locator('#save-working-context').click();
            await page.waitForFunction(() => document.querySelector('#notice').textContent.includes('not acknowledged'));
            assert.match(await page.locator('#state').textContent(), /31% context/,
                'Failed-save form rollback must not replace the latest measurement');
            assert.equal(await page.locator('#working-context-tokens').inputValue(), '100');
            await page.locator('#settings-close').click();
        } finally {
            fs.rmdirSync(settingsPath);
            fs.renameSync(backup, settingsPath);
        }
        await page.locator('#prompt').fill('Unsent draft preserves paused state');
        await page.waitForFunction(() => document.querySelector('#state').textContent === 'paused · 31% context');
        await page.locator('#settings-open').click();
        await page.locator('#working-context-tokens').fill('200');
        await page.locator('#save-working-context').click();
        await page.waitForFunction(() => document.querySelector('#state').textContent === 'paused · 15% context');
        await page.locator('#settings-close').click();
        assert.equal(await page.locator('#prompt').inputValue(), 'Unsent draft preserves paused state');
        await page.reload();
        await page.waitForFunction(() => document.querySelector('#state').textContent.includes('15% context'));
        await page.locator('#settings-open').click();
        assert.equal(await page.locator('#working-context-tokens').inputValue(), '200');
        assert.equal(await page.locator('#context-capacity-tokens').inputValue(), '200000');
        await page.locator('#working-context-tokens').fill('31');
        await page.locator('#save-working-context').click();
        await page.waitForFunction(() => document.querySelector('#state').textContent.includes('context unknown'));
        await page.locator('#settings-close').click();
        // Model switching must not let a stale settings rollback relabel values.
        await page.locator('#model-choice').selectOption('claude-opus-4-6');
        await page.waitForFunction(() => document.querySelector('#working-context-model').textContent.includes('claude-opus-4-6'));
        await page.locator('#settings-open').click();
        const switchedLimit = await page.locator('#working-context-tokens').inputValue();
        assert.notEqual(switchedLimit, '31');
        fs.renameSync(settingsPath, backup);
        fs.mkdirSync(settingsPath);
        try {
            await page.locator('#working-context-tokens').fill('200');
            await page.locator('#save-working-context').click();
            await page.waitForFunction(() => document.querySelector('#notice').textContent.includes('not acknowledged'));
            assert.match(await page.locator('#working-context-model').textContent(), /claude-opus-4-6/);
            assert.equal(await page.locator('#working-context-tokens').inputValue(), switchedLimit,
                'Failed save after model switch must restore that model, not stale prior-model settings');
        } finally {
            fs.rmdirSync(settingsPath);
            fs.renameSync(backup, settingsPath);
        }
        assert.equal(errors.length, 0, errors.join('\n'));
        console.log('PASS working-context page: exact units, last-input ratio, paused placement, live limit change, draft preservation, reload and automatic-checkpoint invalidation');
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
