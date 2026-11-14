// Run: node agent/GuiTests/layout-browser.cjs
// Real layout assets in an isolated browser origin; no Agent or provider requests.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const root = path.join(__dirname, '../gui/wwwroot');
// Use the shipped page structure but activate only layout; no socket or peer requests.
const html = fs.readFileSync(path.join(root, 'index.html'), 'utf8')
    .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '')
    .replace('</body>', '<script src="/layout.js"></script></body>');
(async () => {
    const server = http.createServer((req, res) => {
        if (req.url === '/') { res.setHeader('Content-Type', 'text/html'); res.end(html); }
        else if (['/style.css', '/layout.js'].includes(req.url)) {
            res.setHeader('Content-Type', req.url.endsWith('.js') ? 'text/javascript' : 'text/css');
            res.end(fs.readFileSync(path.join(root, req.url.slice(1))));
        } else { res.writeHead(404); res.end(); }
    });
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    let browser;
    try {
        browser = await chromium.launch({ channel: 'chrome', headless: true });
        const page = await browser.newPage({ viewport: { width: 1600, height: 900 } });
        const errors = []; page.on('pageerror', e => errors.push(e.message));
        const url = `http://127.0.0.1:${server.address().port}`;
        await page.goto(url);
        const create = page.getByRole('button', { name: 'Create Agent', exact: true });
        assert.equal(await create.count(), 1);
        assert.equal(await create.textContent(), '+');
        const heading = await page.locator('#agents-tab').boundingBox();
        const button = await create.boundingBox();
        assert(button.height >= 44 && Math.abs(button.y + button.height / 2 - heading.y - heading.height / 2) < 2);
        await create.evaluate(el => el.addEventListener('click', () => window.createClicked = true));
        await create.focus(); await page.keyboard.press('Enter');
        assert(await page.evaluate(() => window.createClicked));
        const widths = () => page.evaluate(() => ['left', 'right'].map(panePercent));
        const stored = () => page.evaluate(() => localStorage.getItem(paneStorageKey));
        const close = (actual, expected) => actual.forEach((n, i) => assert(Math.abs(n - expected[i]) < .2, `${actual} != ${expected}`));
        close(await widths(), [18, 25]);
        const handle = await page.locator('#left-splitter').boundingBox();
        await page.mouse.move(handle.x + handle.width / 2, handle.y + 20);
        await page.mouse.down(); await page.mouse.move(450, handle.y + 20); await page.mouse.up();
        await page.locator('#right-splitter').press('ArrowLeft');
        const chosen = await widths(), preference = await stored();
        assert(chosen[0] > 27 && chosen[1] > 26);
        await page.reload(); close(await widths(), chosen);
        await page.setViewportSize({ width: 390, height: 844 });
        await page.waitForFunction(() => document.body.classList.contains('compact'));
        await page.reload();
        assert.equal(await stored(), preference, 'Phone reload must not replace desktop preference');
        assert(await page.locator('#chat-pane').isVisible());
        assert(!(await page.locator('#left-splitter').isVisible()));
        await page.setViewportSize({ width: 1600, height: 900 });
        await page.waitForFunction(() => !document.body.classList.contains('compact'));
        close(await widths(), chosen);
        await page.locator('#right-splitter').press('End');
        const widePreference = await stored();
        await page.setViewportSize({ width: 1150, height: 900 });
        await page.waitForFunction(() => ['left-pane','chat-pane','tools-pane'].every(id => document.getElementById(id).getBoundingClientRect().width >= 159));
        assert.equal(await stored(), widePreference, 'Automatic clamp must not save');
        await page.setViewportSize({ width: 1600, height: 900 });
        for (const value of ['{', 'null', '{"left":90,"right":90}', '{"left":"20","right":25}', '{"left":-1,"right":25}']) {
            await page.evaluate(value => localStorage.setItem(paneStorageKey, value), value);
            await page.reload(); close(await widths(), [18, 25]);
        }
        await page.addInitScript(() => {
            Storage.prototype.getItem = () => { throw new Error('storage blocked'); };
            Storage.prototype.setItem = () => { throw new Error('storage blocked'); };
        });
        await page.reload(); close(await widths(), [18, 25]);
        await page.locator('#left-splitter').press('ArrowRight');
        assert((await widths())[0] > 19);
        assert.deepEqual(errors, []);
        console.log('PASS pane layout: drag/keyboard persistence, reload, compact mode, clamping, malformed/blocked storage');
    } finally {
        if (browser) await browser.close();
        await new Promise(resolve => server.close(resolve));
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
