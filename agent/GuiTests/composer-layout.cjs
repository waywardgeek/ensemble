'use strict';
// Geometry only: actual markup/styles, without app startup or any media permission.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const root = path.resolve(__dirname, '../gui/wwwroot');
const server = http.createServer((req, res) => {
    const name = new URL(req.url, 'http://localhost').pathname;
    const file = path.resolve(root, '.' + (name === '/' ? '/index.html' : name));
    if (!file.startsWith(root + path.sep) || !fs.existsSync(file) || !fs.statSync(file).isFile()) {
        res.writeHead(404); res.end(); return;
    }
    const types = { '.html': 'text/html', '.css': 'text/css', '.svg': 'image/svg+xml' };
    res.setHeader('Content-Type', types[path.extname(file)] || 'application/octet-stream');
    const bytes = fs.readFileSync(file);
    res.end(file.endsWith('index.html') ? bytes.toString().replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '') : bytes);
});
(async () => {
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    let browser;
    try {
        browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
        const page = await browser.newPage();
        let groups = 0;
        for (const width of [1440, 390, 320]) for (const font of [16, 28]) for (const extra of [false, true]) {
            await page.setViewportSize({ width, height: 900 });
            await page.goto(`http://127.0.0.1:${server.address().port}/`);
            const metrics = await page.evaluate(({ font, extra }) => {
                document.documentElement.style.fontSize = `${font}px`;
                if (extra) {
                    const mic = document.createElement('button');
                    mic.type = 'button'; mic.id = 'fixture-microphone'; mic.textContent = 'Mic';
                    mic.style.cssText = 'width:44px;min-width:44px;height:44px;padding:0';
                    document.querySelector('#send').before(mic);
                }
                const rect = selector => {
                    const r = document.querySelector(selector).getBoundingClientRect();
                    return { left: r.left, right: r.right, top: r.top, bottom: r.bottom, width: r.width, height: r.height };
                };
                return { input: rect('#prompt'), actions: rect('#compose-actions'), attach: rect('#attach-images'),
                    send: rect('#send'), mic: extra ? rect('#fixture-microphone') : null };
            }, { font, extra });
            const label = `${width}px / ${font}px text / third control ${extra}`;
            assert.ok(metrics.actions.top >= metrics.input.bottom, `${label}: controls must be below input`);
            assert.ok(Math.abs(metrics.input.width - metrics.actions.width) < 2, `${label}: full-width input`);
            for (const button of [metrics.attach, metrics.send, metrics.mic].filter(Boolean)) {
                assert.ok(button.width >= 44 && button.height >= 44, `${label}: touch target`);
                assert.ok(button.left >= metrics.actions.left - 1 && button.right <= metrics.actions.right + 1,
                    `${label}: button overflow`);
            }
            assert.ok(metrics.attach.right <= metrics.send.left || metrics.attach.bottom <= metrics.send.top,
                `${label}: overlapping actions`);
            if (metrics.mic) {
                assert.ok(metrics.attach.right <= metrics.mic.left || metrics.attach.bottom <= metrics.mic.top,
                    `${label}: microphone overlaps attachment`);
                assert.ok(metrics.mic.right <= metrics.send.left || metrics.mic.bottom <= metrics.send.top,
                    `${label}: microphone overlaps Send`);
            }
            console.log(`PASS ${label}`); groups++;
        }
        console.log(`${groups} composer geometry groups passed`);
    } finally {
        if (browser) await browser.close();
        await new Promise(resolve => server.close(resolve));
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
