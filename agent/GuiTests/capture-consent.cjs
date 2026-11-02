const assert = require('node:assert/strict');
const fs = require('node:fs');
const http = require('node:http');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const assets = path.resolve(__dirname, '../gui/wwwroot');
const markup = `<dialog id="consent"><h2>Share this page with <span data-capture-requester></span>?</h2><p data-capture-status></p><button id="share" data-capture-share>Share this page</button><button id="decline" data-capture-cancel>Cancel</button></dialog>`;
(async () => {
    const server = http.createServer((req, res) => {
        if (req.url === '/capture-consent.js' || req.url === '/mcp.js') {
            res.setHeader('Content-Type', 'application/javascript');
            res.end(fs.readFileSync(path.join(assets, req.url.slice(1))));
        } else {
            res.setHeader('Content-Type', 'text/html; charset=utf-8');
            res.end('<!doctype html><html><body>' + markup + '<script src="/mcp.js"></script></body></html>');
        }
    });
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    let browser;
    try {
        browser = await chromium.launch({ channel: 'chrome', headless: true });
        const page = await browser.newPage(); page.setDefaultTimeout(8000);
        await page.goto(`http://127.0.0.1:${server.address().port}/`);
        await page.evaluate(async () => {
            const { CaptureConsent } = await import('./capture-consent.js');
            const canvas = document.createElement('canvas'); canvas.width = 32; canvas.height = 16;
            const context = canvas.getContext('2d'); context.fillStyle = '#f00'; context.fillRect(0, 0, 16, 16);
            context.fillStyle = '#00f'; context.fillRect(16, 0, 16, 16);
            const blob = await new Promise(resolve => canvas.toBlob(resolve, 'image/png'));
            window.png = new File([blob], 'synthetic.png', { type: 'image/png' });
            window.capture = { supported: true, busy: false, calls: 0, stops: 0, held: null, fail: false,
                async capture() {
                    this.calls++; this.busy = true;
                    if (this.held) await this.held;
                    this.busy = false;
                    if (this.fail) throw new Error('synthetic capture failure');
                    return { file: png };
                },
                async cancel() { this.stops++; if (this.held) await this.held; this.busy = false; }
            };
            window.peer = { readyState: 1 };
            window.consent = new CaptureConsent(capture, document.querySelector('#consent'));
            window.begin = id => { window.result = null; window.reply = consent.request(peer, 'fixture', id).then(value => { result = value; return value; }); };
            window.hold = () => { capture.held = new Promise(resolve => { window.release = resolve; }); };
        });
        await page.evaluate(() => { begin(1); browserCall('gui_click', { selector: '#share' }); });
        assert.equal(await page.evaluate(() => capture.calls), 0);
        assert.equal(await page.locator('#consent').evaluate(e => e.open), true);
        await page.locator('#share').click();
        const first = await page.evaluate(() => reply);
        assert.equal(first.content[0].type, 'image'); assert.equal(first.content[0].mimeType, 'image/png');
        assert.ok(first.content[0].data.length > 0);
        const pixels = await page.evaluate(async value => {
            const blob = await (await fetch('data:image/png;base64,' + value)).blob();
            const bitmap = await createImageBitmap(blob); const canvas = document.createElement('canvas');
            canvas.width = bitmap.width; canvas.height = bitmap.height; const context = canvas.getContext('2d');
            context.drawImage(bitmap, 0, 0); const result = [bitmap.width, bitmap.height, ...context.getImageData(8,8,1,1).data, ...context.getImageData(24,8,1,1).data];
            bitmap.close(); return result;
        }, first.content[0].data);
        assert.deepEqual(pixels, [32,16,255,0,0,255,0,0,255,255]);
        assert.equal(await page.evaluate(() => capture.calls), 1);
        console.log('PASS trusted human click required; gui_click cannot approve; actual PNG pixels returned');
        assert.deepEqual(await page.evaluate(async()=>{
            const missing=await consent.request(peer,undefined,17);
            const pending=consent.request(peer,'',18,'Unnamed root');
            const visible=document.querySelector('#consent').open;
            await consent.cancelMatching(peer,'not-root',18);
            const preserved=document.querySelector('#consent').open;
            await consent.cancelMatching(peer,'',18);
            return [missing.isError,visible,preserved,(await pending).isError];
        }),[true,true,true,true]);
        console.log('PASS empty canonical root address stays bound; missing address refuses');


        await page.evaluate(() => begin(2)); await page.locator('#decline').click();
        assert.equal((await page.evaluate(() => reply)).isError, true);
        assert.equal(await page.evaluate(() => capture.calls), 1);
        await page.evaluate(async () => {
            begin(3);
            await consent.cancelMatching({readyState:1}, 'fixture', 3);
            await consent.cancelMatching(peer, 'other', 3);
            await consent.cancelMatching(peer, 'fixture', 4);
        });
        assert.equal(await page.evaluate(() => consent.busy), true);
        await page.evaluate(() => consent.cancelMatching(peer, 'fixture', 3));
        assert.equal((await page.evaluate(() => reply)).isError, true);
        console.log('PASS decline and cancellation bound to exact socket, Agent and RPC');

        await page.evaluate(() => { hold(); begin(4); }); await page.locator('#share').click();
        await page.evaluate(() => { window.cleaned = false; consent.cancel().then(() => { cleaned = true; }); });
        assert.equal(await page.evaluate(() => cleaned), false);
        assert.equal(await page.evaluate(() => consent.busy), true);
        await page.evaluate(() => { release(); capture.held = null; });
        await page.waitForFunction(() => cleaned);
        assert.equal((await page.evaluate(() => reply)).isError, true);
        assert.equal(await page.evaluate(() => consent.busy), false);
        console.log('PASS cancellation joins uncooperative capture and suppresses late pixels');

        await page.evaluate(() => { hold(); begin(5); }); await page.locator('#share').click();
        await page.evaluate(() => { peer.readyState = 3; release(); capture.held = null; });
        assert.equal((await page.evaluate(() => reply)).isError, true);
        await page.evaluate(() => { peer.readyState = 1; });
        console.log('PASS closed connection cannot receive a late image');

        const refusals = await page.evaluate(async () => {
            capture.supported = false; const unsupported = await consent.request(peer,'fixture',6);
            capture.supported = true; capture.busy = true; const manual = await consent.request(peer,'fixture',7);
            const before = capture.stops; await consent.cancel(); const after = capture.stops; capture.busy = false;
            begin(8); const duplicate = await consent.request(peer,'fixture',9);
            await consent.cancel(); await reply;
            begin(10); capture.busy = true; window.manualStops = capture.stops;
            return { unsupported, manual, before, after, duplicate };
        });
        for (const key of ['unsupported','manual','duplicate']) assert.equal(refusals[key].isError, true);
        assert.equal(refusals.before, refusals.after);
        await page.locator('#share').click(); assert.equal((await page.evaluate(() => reply)).isError, true);
        assert.equal(await page.evaluate(() => capture.stops === manualStops && capture.busy), true);
        await page.evaluate(() => { capture.busy = false; capture.fail = true; begin(11); });
        await page.locator('#share').click(); assert.equal((await page.evaluate(() => reply)).isError, true);
        assert.equal(await page.evaluate(() => consent.busy), false);
        console.log('PASS unsupported/concurrent/manual capture and native failure refuse without stealing ownership');
        await page.evaluate(() => {
            capture.fail = false; peer.readyState = 1; peer.relayAgentID = 'fixture'; peer.mcpInitialized = true;
            window.sent = []; peer.send = text => sent.push(JSON.parse(text));
            window.requestBrowserCapture = (connection, agent, id) => consent.request(connection, agent, id);
            window.cancelBrowserCapture = (connection, agent, id) => consent.cancelMatching(connection, agent, id);
            window.closeBrowserCapture = connection => consent.cancelConnection(connection);
            window.rpc = (method, id, params) => handleMcp(peer, {agent_id:'fixture',data:{jsonrpc:'2.0',method,...(id === undefined ? {} : {id}),params}});
            rpc('tools/call',20,{name:'gui_capture',arguments:{}});
            rpc('tools/call',21,{name:'gui_snapshot',arguments:{}});
            rpc('tools/call',22,{name:'gui_click',arguments:{selector:'#share'}});
        });
        assert.deepEqual(await page.evaluate(() => sent.map(item => item.data.id)), [21,22]);
        assert.equal(await page.evaluate(() => sent.every(item => item.data.result && !item.data.result.isError)), true);
        assert.equal(await page.evaluate(() => consent.busy), true);
        await page.evaluate(() => rpc('$/cancelRequest',undefined,{id:20}));
        await page.waitForFunction(() => sent.some(item => item.data.id === 20));
        assert.equal(await page.evaluate(() => sent.find(item=>item.data.id===20).data.result.isError), true);
        await page.evaluate(() => { rpc('tools/call',23,{name:'gui_capture',arguments:{}}); rpc('notifications/ensemble/browser_closed'); });
        await page.waitForFunction(() => sent.some(item=>item.data.id===23));
        assert.equal(await page.evaluate(() => !peer.mcpInitialized && sent.find(item=>item.data.id===23).data.result.isError), true);
        console.log('PASS asynchronous MCP consent preserves ordinary RPCs and handles cancellation/session close');

        await page.evaluate(() => {
            peer.mcpInitialized = true; hold(); rpc('tools/call',24,{name:'gui_capture',arguments:{}});
        });
        await page.locator('#share').click();
        await page.evaluate(() => {
            peer.readyState = 3;
            window.replacement = {readyState:1, sent:[], send(value) {this.sent.push(value);}};
            window.socket = replacement;
            release(); capture.held = null;
        });
        await page.waitForFunction(() => !consent.busy);
        assert.equal(await page.evaluate(() => sent.some(item=>item.data.id===24)), false);
        assert.equal(await page.evaluate(() => replacement.sent.length), 0);
        console.log('PASS delayed MCP result is not replayed onto a replacement connection');

    } finally {
        if (browser) await browser.close();
        await new Promise(resolve => server.close(resolve));
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
