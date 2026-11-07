'use strict';
// Real page, synthetic recorded facts; filesystem evidence is tested separately.
const assert = require('node:assert/strict');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL;
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
(async () => {
    const browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
    try {
        for (const width of [1280, 390]) {
            const page = await browser.newPage({ viewport: { width, height: 844 } });
            const errors = [];
            page.on('pageerror', e => errors.push(e.message));
            await page.goto(base);
            await page.waitForFunction(() => settings !== null && !selectionPending && typeof window.EnsembleArtifactRenderers?.search_files === 'function');
            if (width === 390) await page.locator('.pane-tab[data-pane="tools"]').click();
            await page.evaluate(() => {
                settings = { ...settings, tts_enabled: false };
                const add = (id, evidence, error = false) => {
                    handle({ type: 'tool_dispatched', part_id: id + '-call', name: 'search_files', input: {} });
                    handle({ type: 'tool_finished', part_id: id, tool_part_id: id + '-call', result: 'Unchanged raw grep text', search_evidence: evidence, is_error: error });
                };
                const evidence = { total_lines: 7, lines: [
                    { path: '/odd:<img>\nfile', number: 8, text: 'before', matches: [] },
                    { path: '/odd:<img>\nfile', number: 9, text: '😀 hit hit <img src=x>', matches: [{ start: 3, length: 3 }, { start: 7, length: 3 }] },
                    { path: '/odd:<img>\nfile', number: 14, text: 'tail', matches: [{ start: 0, length: 0 }] },
                    { path: '/second', number: 2, text: 'hit', matches: [{ start: 0, length: 3 }] }
                ] };
                add('search-result', evidence); add('search-legacy', null); add('search-error', evidence, true);
                add('search-empty', { lines: [], total_lines: 0 });
            });
            const result = page.locator('[data-part-id="search-result"]');
            assert.deepEqual(await result.locator('.search-path').allTextContents(), ['/odd:<img>\nfile', '/second']);
            assert.deepEqual(await result.locator('.file-line-number').allTextContents(), ['8', '9', '14', '2']);
            assert.deepEqual(await result.locator('mark').allTextContents(), ['hit', 'hit', '[zero-width match]', 'hit']);
            assert.equal(await result.locator('.search-gap').count(), 1);
            assert.match(await result.locator('.search-preview').textContent(), /4 of 7.*3 omitted/);
            assert.deepEqual(await result.locator('.search-line-kind').allTextContents(), ['context: ', 'match: ', 'match: ', 'match: ']);
            assert.equal(await result.locator('img, script').count(), 0);
            assert.equal(await result.locator('details > pre').textContent(), 'Unchanged raw grep text');
            for (const id of ['search-legacy', 'search-error']) assert.equal(await page.locator(`[data-part-id="${id}"] .preview`).textContent(), 'Unchanged raw grep text');
            assert.match(await page.locator('[data-part-id="search-empty"] .search-preview').textContent(), /0 returned/);
            assert(await page.evaluate(() => {
                const card = cards.get('search-result'); const preview = card.node.querySelector('.search-preview');
                const summary = card.node.querySelector('summary'); summary.focus();
                if (document.activeElement !== summary) throw Error('Focus precondition failed');
                render(card); return card.node.querySelector('.search-preview') === preview && document.activeElement === summary;
            }));
            assert.deepEqual(errors, []);
            await page.close();
            console.log(`PASS search renderer ${width}: groups, spans, line numbers, gaps, zero width, truncation, safe text, fallback, raw output and focus`);
        }
    } finally { await browser.close(); }
})().catch(e => { console.error(e); process.exitCode = 1; });
