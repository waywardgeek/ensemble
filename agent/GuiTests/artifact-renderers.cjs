'use strict';
// Actual shipped page and production handler; synthetic facts, no inference.
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL;
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
(async () => {
    const browser = await chromium.launch({
        executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true
    });
    try {
        for (const width of [1280, 390]) {
            const page = await browser.newPage({ viewport: { width, height: 844 } });
            const errors = [];
            page.on('pageerror', error => errors.push(error.message));
            await page.goto(base);
            await page.waitForFunction(() => settings !== null && !selectionPending);
            if (width === 390) await page.locator('.pane-tab[data-pane="tools"]').click();
            const text = '<img src="https://invalid.example/leak" onerror="window.leaked=true">\nfolder/\nname\nwith-newline\n';
            await page.evaluate(text => {
                settings = { ...settings, tts_enabled: false };
                handle({ type: 'tool_dispatched', part_id: 'directory-call', name: 'list_directory', input: { path: '.' } });
                handle({ type: 'tool_finished', part_id: 'directory-result', tool_part_id: 'directory-call', result: text });
            }, text);
            const result = page.locator('[data-part-id="directory-result"]');
            assert.equal(await result.locator('.directory-preview pre').textContent(), text);
            assert.equal(await result.locator('details > pre').textContent(), text);
            assert.equal(await result.locator('img, script').count(), 0);
            assert(await page.evaluate(() => {
                const card = cards.get('directory-result');
                const body = card.node.querySelector('.directory-preview');
                const summary = card.node.querySelector('summary');
                summary.focus();
                if (document.activeElement !== summary) throw new Error("Test precondition: summary must receive focus");
                render(card);
                return body === card.node.querySelector('.directory-preview') && document.activeElement === summary;
            }), 'metadata rendering replaced body or focus');
            assert(await page.evaluate(() => {
                const card = cards.get('directory-result');
                window.savedDirectoryRenderer = window.EnsembleArtifactRenderers.list_directory;
                delete window.EnsembleArtifactRenderers.list_directory;
                render(card);
                return !card.node.querySelector('.directory-preview') && card.node.querySelector('.preview').textContent === card.text;
            }), 'missing renderer did not fall back');
            assert(await page.evaluate(() => {
                const card = cards.get('directory-result');
                window.EnsembleArtifactRenderers.list_directory = () => { throw new Error('test failure'); };
                render(card);
                return card.node.querySelector('.preview').textContent === card.text;
            }), 'throwing renderer hid output');
            for (const invalid of ['null', 'connected']) {
                assert(await page.evaluate(invalid => {
                    const card = cards.get('directory-result');
                    const heading = card.node.querySelector('.card-heading');
                    window.EnsembleArtifactRenderers.list_directory = () => invalid === 'null' ? null : heading;
                    render(card);
                    return heading.parentElement === card.node && card.node.querySelector('.preview').textContent === card.text;
                }, invalid), 'invalid renderer result moved controls or hid output');
            }
            assert(await page.evaluate(() => {
                const card = cards.get('directory-result');
                window.EnsembleArtifactRenderers.list_directory = window.savedDirectoryRenderer;
                card.status = 'error';
                render(card);
                return !card.node.querySelector('.directory-preview');
            }), 'error result used successful directory renderer');
            await page.evaluate(() => {
                handle({ type: 'tool_finished', part_id: 'unpaired-result', result: 'unpaired' });
                handle({ type: 'tool_dispatched', part_id: 'large-call', name: 'list_directory', input: { path: '.' } });
                handle({ type: 'tool_finished', part_id: 'large-result', tool_part_id: 'large-call', result: 'entry\n'.repeat(1000) });
            });
            assert.equal(await page.locator('[data-part-id="unpaired-result"] .directory-preview').count(), 0);
            assert.equal(await page.locator('[data-part-id="directory-call"] .directory-preview').count(), 0);
            const large = page.locator('[data-part-id="large-result"]');
            assert((await large.locator('.preview').textContent()).includes('(bounded preview)'));
            assert.equal((await large.locator('details > pre').textContent()).length, 6000);
            assert.deepEqual(errors, []);
            await page.close();
            console.log(`PASS directory renderer ${width}: literal data, raw detail, stable nodes/focus, missing/failing/error/unpaired fallback, bounded preview`);
        }
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
