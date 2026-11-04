'use strict';
// Actual page handler and rendering; synthetic wire facts isolate presentation.
// These snapshots are recorded observations, not a live Job monitor.
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
(async () => {
    const browser = await chromium.launch({
        executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true
    });
    try {
        const page = await browser.newPage({ viewport: { width: 390, height: 844 } });
        // Explicit local candidate override for sensitivity checks only.
        if (process.env.GUI_ACTIONS_SOURCE) {
            const body = require('node:fs').readFileSync(process.env.GUI_ACTIONS_SOURCE, 'utf8');
            await page.route('**/app.js', route => route.fulfill({ status: 200, contentType: 'text/javascript', body }));
        }
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        const productRequests = [];
        page.on('request', request => {
            if (request.url().includes('/work-products/download?')) productRequests.push(request);
        });
        await page.goto(base);
        await page.waitForFunction(() => settings !== null && !selectionPending);
        await page.locator('.pane-tab[data-pane="tools"]').click();
        const fs = require('node:fs');
        const fixture = JSON.parse(fs.readFileSync('.crossover/runtime/fixture.json', 'utf8'));
        await page.waitForFunction(id => [...cards.values()].some(card => card.workProduct?.id === id), fixture.workProduct.ID);
        const selector = await page.evaluate(id => {
            const card = [...cards.values()].find(card => card.workProduct?.id === id);
            return '[data-part-id="' + card.id + '"]';
        }, fixture.workProduct.ID);
        const card = page.locator(selector);
        const button = card.locator('.work-product-links button');
        assert.equal(productRequests.length, 0, 'rendering must not automatically transfer a file');
        const downloadEvent = page.waitForEvent('download');
        await button.click();
        const download = await downloadEvent;
        assert.equal(download.suggestedFilename(), fixture.workProduct.Name);
        const bytes = fs.readFileSync(await download.path());
        assert.equal(require('node:crypto').createHash('sha256').update(bytes).digest('hex'), fixture.workProductHash, 'download must contain full click-time bytes');
        assert.equal(bytes.length, 32769);
        await card.locator('.work-product-links [role="status"]').filter({ hasText: 'Downloaded 32,769 bytes.' }).waitFor();
        assert.equal(productRequests.length, 1);
        assert.equal(productRequests[0].headers()['x-ensemble-work-product'], '1');
        assert.equal(new URL(productRequests[0].url()).searchParams.get('offer'), fixture.workProduct.ID);
        await page.evaluate(id => {
            const card = [...cards.values()].find(card => card.workProduct?.id === id);
            window.productButton = card.node.querySelector('.work-product-links button');
            productButton.focus(); render(card);
        }, fixture.workProduct.ID);
        assert(await button.evaluate(node => node === window.productButton && node === document.activeElement), 'rerender must preserve offer control and focus');
        // Stub only a blocked transport for a deterministic cancellation boundary.
        await page.evaluate(() => {
            window.originalProductFetch = window.fetch;
            window.fetch = (url, options) => String(url).includes('/work-products/download?')
                ? new Promise((resolve, reject) => options.signal.addEventListener('abort', () => {
                    window.productAbortSeen = true;
                    reject(new DOMException('cancelled', 'AbortError'));
                }, { once: true }))
                : originalProductFetch(url, options);
        });
        await button.click();
        await button.filter({ hasText: 'Cancel download' }).waitFor();
        await button.click();
        await page.waitForFunction(() => window.productAbortSeen === true);
        await card.locator('.work-product-links [role="status"]').filter({ hasText: 'Download cancelled.' }).waitFor();
        // Reject an oversized declared response before reading or creating a Blob.
        await page.evaluate(() => {
            window.fetch = async () => new Response('unused', { headers: { 'Content-Length': String(64 * 1024 * 1024 + 1) } });
        });
        await button.click();
        await card.locator('.work-product-links [role="status"]').filter({ hasText: 'exceeds the 64 MiB download limit' }).waitFor();
        await page.evaluate(() => {
            window.fetch = async () => {
                const chunk = new Uint8Array(1024 * 1024);
                let count = 0;
                return new Response(new ReadableStream({ pull(controller) {
                    if (++count <= 64) controller.enqueue(chunk);
                    else controller.close();
                } }));
            };
        });
        const boundaryEvent = page.waitForEvent('download');
        await button.click();
        const boundaryDownload = await boundaryEvent;
        assert.equal(fs.statSync(await boundaryDownload.path()).size, 64 * 1024 * 1024, 'exact browser 64 MiB boundary must succeed');
        // Also enforce the cap when Content-Length is absent. Reusing one
        // synthetic chunk avoids allocating a large fixture outside the reader.
        await page.evaluate(() => {
            window.fetch = async () => {
                const chunk = new Uint8Array(1024 * 1024);
                let count = 0;
                return new Response(new ReadableStream({ pull(controller) {
                    if (++count <= 65) controller.enqueue(chunk);
                    else controller.close();
                } }));
            };
        });
        await button.click();
        await card.locator('.work-product-links [role="status"]').filter({ hasText: 'exceeds the 64 MiB download limit' }).waitFor();
        // A late response cannot download from a detached/replaced card.
        let staleDownloads = 0;
        const onLateDownload = () => staleDownloads++;
        page.on('download', onLateDownload);
        await page.evaluate(() => {
            window.lateProductAnchors = 0;
            const originalClick = HTMLAnchorElement.prototype.click;
            HTMLAnchorElement.prototype.click = function() {
                if (this.download) window.lateProductAnchors++;
                return originalClick.call(this);
            };
            window.fetch = () => new Promise(resolve => { window.finishStaleProduct = resolve; });
        });
        await button.click();
        await page.waitForFunction(() => typeof window.finishStaleProduct === 'function');
        await page.evaluate(id => {
            const old = [...cards.values()].find(card => card.workProduct?.id === id);
            cards.delete(old.id); old.node.remove();
            finishStaleProduct(new Response('stale bytes'));
        }, fixture.workProduct.ID);
        await page.waitForFunction(() => !window.productButton.textContent.includes('Cancel'));
        assert.equal(staleDownloads, 0, 'detached offer must not start a late download');
        assert.equal(await page.evaluate(() => window.lateProductAnchors), 0, 'detached offer must not attempt a late download');
        page.off('download', onLateDownload);
        await page.evaluate(() => { window.fetch = window.originalProductFetch; });
        // Same typed description without live authority remains unavailable.
        page.removeAllListeners('request');
        await page.evaluate(() => handle({ type: 'tool_finished', part_id: 'expired-product', result: 'old offer',
            work_product: { id: 'f'.repeat(32), name: 'expired.bin', offered_bytes: 1 } }));
        const expired = page.locator('[data-part-id="expired-product"]');
        await expired.locator('.work-product-links button').click();
        await expired.locator('[role="status"]').filter({ hasText: 'Offer expired, unavailable' }).waitFor();
        await page.evaluate(() => {
            handle({ type: 'tool_finished', part_id: 'path-only-product', result: 'Download /tmp/report.bin' });
            handle({ type: 'tool_finished', part_id: 'error-product', is_error: true, result: 'failed',
                work_product: { id: 'a'.repeat(32), name: 'bad.bin', offered_bytes: 1 } });
        });
        assert.equal(await page.locator('[data-part-id="path-only-product"] .work-product-links').count(), 0);
        assert.equal(await page.locator('[data-part-id="error-product"] .work-product-links').count(), 0);
        // Exercise the actual replay handler, not just manually detached nodes.
        await page.evaluate(() => {
            handle({ type: 'tool_finished', part_id: 'replay-product', result: 'offer',
                work_product: { id: 'b'.repeat(32), name: 'replay.bin', offered_bytes: 1 } });
            window.replayProduct = cards.get('replay-product');
            window.fetch = (url, options) => new Promise((resolve, reject) => {
                options.signal.addEventListener('abort', () => {
                    window.replayProductAborted = true;
                    reject(new DOMException('replay cancelled', 'AbortError'));
                }, { once: true });
            });
        });
        await page.locator('[data-part-id="replay-product"] .work-product-links button').click();
        await page.evaluate(() => handle({ type: 'replay', agent_id: agentID,
            data: { events: [], partials: [], agents: [] } }));
        assert(await page.evaluate(() => window.replayProductAborted === true), 'replay must cancel the discarded transfer');
        await page.waitForFunction(() => !window.replayProduct.workProductController);
        assert.equal(await page.locator('[data-part-id="replay-product"]').count(), 0);
        assert.equal(await page.evaluate(() => window.lateProductAnchors), 0, 'replay must not start a discarded download');
        await page.evaluate(() => { window.fetch = window.originalProductFetch; });
        console.log('PASS actual replay aborts discarded offer request and clears its controller');
        assert.deepEqual(errors, []);
        console.log('PASS explicit click downloads exact real-route click-time bytes and original filename; no automatic transfer');
        console.log('PASS stable offer controls, cancel transport, expired offer refusal, no prose/error offers');
    } finally {
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
