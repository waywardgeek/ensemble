// Real shipped HTML/CSS, synthetic status text; no app socket or inference.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const assert = require('node:assert/strict');
const root = path.resolve(__dirname, '../gui/wwwroot');

(async () => {
    const server = http.createServer((req, res) => {
        const name = new URL(req.url, 'http://localhost').pathname;
        const file = path.resolve(root, '.' + (name === '/' ? '/index.html' : name));
        if (!file.startsWith(root + path.sep)) { res.writeHead(403).end(); return; }
        try {
            const data = fs.readFileSync(file);
            res.setHeader('Content-Type', file.endsWith('.css') ? 'text/css' :
                file.endsWith('.html') ? 'text/html' : 'application/octet-stream');
            res.end(data);
        } catch { res.writeHead(404).end(); }
    });
    let browser;
    try {
        await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
        browser = await chromium.launch({ headless: true, executablePath:
            process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' });
        const page = await browser.newPage();
        await page.route('**/*.js', route => route.abort());
        let checks = 0;
        for (const width of [390, 640, 1024, 2048]) {
            for (const font of [18, 28]) {
                await page.setViewportSize({ width, height: 900 });
                await page.goto(`http://127.0.0.1:${server.address().port}/`);
                for (const label of ['paused · 12% context', 'toolspending · 100% context',
                    'inflight · 125% context', 'disconnected · context unknown']) {
                    await page.evaluate(({ font, label }) => {
                        document.documentElement.style.fontSize = font + 'px';
                        document.querySelector('#connection').textContent = 'Connected';
                        document.querySelector('#state').textContent = label;
                        document.querySelector('#model-choice').replaceChildren(new Option('GPT-6 Astra', 'astra'));
                        document.querySelector('#header-goal').textContent = 'Don’t Be Evil';
                    }, { font, label });
                    const geometry = await page.evaluate(() => {
                        const rect = selector => {
                            const r = document.querySelector(selector).getBoundingClientRect();
                            return { left: r.left, right: r.right };
                        };
                        const labels = ['#connection', '#state'].map(selector => {
                            const e = document.querySelector(selector);
                            const range = document.createRange(); range.selectNodeContents(e);
                            const text = range.getBoundingClientRect();
                            return { selector, box: rect(selector), text: { left: text.left, right: text.right },
                                client: e.clientWidth, scroll: e.scrollWidth };
                        });
                        const strip = document.querySelector('.header-scroll');
                        strip.scrollLeft = strip.scrollWidth;
                        const actions = rect('.status-actions'), viewport = rect('.header-scroll');
                        const reachable = actions.right <= viewport.right + 1 && actions.left >= viewport.left - 1;
                        strip.scrollLeft = 0;
                        return { labels, start: rect('.header-start'), model: rect('#model-controls'),
                            end: rect('.header-end'), reachable, viewport,
                            width: innerWidth };
                    });
                    const context = `${width}px / ${font}px / ${label}`;
                    for (const item of geometry.labels) {
                        assert(item.scroll <= item.client + 1, `${context}: ${item.selector} clips`);
                        assert(item.text.left >= item.box.left - 1 && item.text.right <= item.box.right + 1,
                            `${context}: text extends outside ${item.selector}`);
                    }
                    assert(geometry.start.right <= geometry.model.left + 1, `${context}: status/model overlap`);
                    assert(geometry.model.right <= geometry.end.left + 1, `${context}: model/actions overlap`);
                    assert(geometry.reachable, `${context}: actions unreachable by local scroll`);
                    assert(geometry.viewport.left >= 0 && geometry.viewport.right <= geometry.width,
                        `${context}: header scroller exceeds viewport`);
                    checks++;
                }
            }
        }
        console.log(`PASS ${checks} header layouts: complete status labels, separate groups, reachable actions, viewport-bounded header`);
    } finally {
        try { if (browser) await browser.close(); }
        finally { await new Promise(resolve => server.close(resolve)); }
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
