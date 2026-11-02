'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const http = require('node:http');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const script = fs.readFileSync(path.join(__dirname, '../gui/wwwroot/theme.js'));
const html = '<button id="theme-toggle"><svg><use></use></svg></button><select id="theme"><option>system</option><option>dark</option><option>light</option></select><script src="/theme.js"></script>';
const server = http.createServer((req, res) => {
    res.setHeader('Content-Type', req.url === '/theme.js' ? 'text/javascript' : 'text/html');
    res.end(req.url === '/theme.js' ? script : html);
});
(async () => {
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    let browser;
    try {
        browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
        const context = await browser.newContext({ colorScheme: 'light' });
        const url = `http://127.0.0.1:${server.address().port}`;
        const first = await context.newPage(); await first.goto(url);
        await first.evaluate(() => browserTheme.seed('dark'));
        assert.equal(await first.locator('html').getAttribute('data-bs-theme'), 'dark');
        await first.selectOption('#theme', 'light');
        await first.evaluate(() => browserTheme.seed('dark'));
        assert.equal(await first.locator('html').getAttribute('data-bs-theme'), 'light');
        await first.reload();
        assert.equal(await first.locator('html').getAttribute('data-bs-theme'), 'light');
        const second = await context.newPage(); await second.goto(url);
        await first.click('#theme-toggle');
        await second.waitForFunction(() => document.documentElement.dataset.bsTheme === 'dark');
        const isolated = await browser.newContext({ colorScheme: 'light' });
        const third = await isolated.newPage(); await third.goto(url);
        assert.equal(await third.locator('html').getAttribute('data-bs-theme'), 'light');
        await first.selectOption('#theme', 'system');
        await first.emulateMedia({ colorScheme: 'dark' });
        await first.waitForFunction(() => document.documentElement.dataset.bsTheme === 'dark');
        await first.emulateMedia({ colorScheme: 'light' });
        await first.waitForFunction(() => document.documentElement.dataset.bsTheme === 'light');
        await first.evaluate(() => localStorage.clear());
        await second.waitForFunction(() => document.documentElement.dataset.theme === 'system');
        const blocked = await browser.newContext();
        await blocked.addInitScript(() => Object.defineProperty(window, 'localStorage', { get() { throw new DOMException('blocked', 'SecurityError'); } }));
        const fourth = await blocked.newPage(); await fourth.goto(url);
        await fourth.evaluate(() => browserTheme.seed('dark'));
        await fourth.selectOption('#theme', 'light');
        await fourth.evaluate(() => browserTheme.seed('dark'));
        assert.equal(await fourth.locator('html').getAttribute('data-bs-theme'), 'light');
        assert.equal(await fourth.locator('#theme-toggle').isEnabled(), true);
        console.log('PASS legacy seed, local precedence, reload, cross-tab sync, profile isolation, OS changes, clear, blocked storage');
    } finally {
        if (browser) await browser.close();
        await new Promise(resolve => server.close(resolve));
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
