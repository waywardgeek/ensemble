// Only the isolated CrossoverFixture: no model prompt or live user browser.
'use strict';
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089'].includes(new URL(base).port));
(async () => {
    const browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
    const errors = [], sent = [];
    try {
        const context = await browser.newContext({ colorScheme: 'light' });
        const page = await context.newPage();
        page.on('pageerror', error => errors.push(error.message));
        page.on('websocket', ws => ws.on('framesent', frame => { try { sent.push(JSON.parse(String(frame.payload))); } catch {} }));
        async function ready(p) {
            await p.waitForFunction(() => settings !== null && !selectionPending);
        }
        await page.goto(base); await ready(page);
        const legacy = await page.evaluate(() => settings.theme);
        assert.equal(await page.evaluate(() => localStorage.getItem('ensemble.theme.v1')), legacy);
        await page.click('#theme-toggle');
        const selected = await page.locator('html').getAttribute('data-bs-theme');
        await page.evaluate(() => { window.previousThemeSettings = settings; send('get_settings'); });
        await page.waitForFunction(() => settings !== window.previousThemeSettings);
        assert.equal(await page.locator('html').getAttribute('data-bs-theme'), selected);
        assert.equal(await page.evaluate(() => settings.theme), legacy);
        await page.reload(); await ready(page);
        assert.equal(await page.locator('html').getAttribute('data-bs-theme'), selected);
        await page.click('#settings-open');
        await page.click('#settings-display-tab');
        await page.selectOption('#theme', 'system');
        await page.emulateMedia({ colorScheme: 'dark' });
        await page.waitForFunction(() => document.documentElement.dataset.bsTheme === 'dark');
        await page.emulateMedia({ colorScheme: 'light' });
        await page.waitForFunction(() => document.documentElement.dataset.bsTheme === 'light');
        await page.click('#settings-close');
        const second = await context.newPage(); await second.goto(base); await ready(second);
        await page.click('#theme-toggle');
        await second.waitForFunction(() => document.documentElement.dataset.theme === 'dark');
        // An unrelated storage area must not override the saved local preference.
        await page.evaluate(() => window.dispatchEvent(new StorageEvent('storage', {
            key: 'ensemble.theme.v1', newValue: 'light', storageArea: sessionStorage
        })));
        assert.equal(await page.locator('html').getAttribute('data-bs-theme'), 'dark');
        const other = await browser.newContext({ colorScheme: 'light' });
        const third = await other.newPage(); await third.goto(base); await ready(third);
        assert.equal(await third.evaluate(() => localStorage.getItem('ensemble.theme.v1')), legacy);
        await page.click('#theme-toggle');
        assert.equal(await third.evaluate(() => localStorage.getItem('ensemble.theme.v1')), legacy);
        // Offline changes are local; no server acknowledgement is necessary.
        await context.setOffline(true);
        await page.evaluate(() => socket.close());
        await page.waitForFunction(() => document.querySelector('#connection').textContent.startsWith('Disconnected'));
        assert.equal(await page.locator('#theme-toggle').isEnabled(), true);
        await page.click('#theme-toggle');
        assert.equal(await page.locator('html').getAttribute('data-bs-theme'), 'dark');
        await context.setOffline(false);
        await page.waitForFunction(() => document.querySelector('#connection').textContent === 'Connected');
        await ready(page);
        assert.equal(await page.locator('html').getAttribute('data-bs-theme'), 'dark');
        const themeWrites = sent.filter(frame => frame.type === 'update_settings' && JSON.stringify(frame).includes('"theme"'));
        assert.deepEqual(themeWrites, []);
        assert.deepEqual(errors, []);
        console.log('PASS full page: migration, header/dialog controls, settings refresh, reload, OS, tabs, profiles, storage filtering, disconnect/reconnect and no backend theme writes');
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
