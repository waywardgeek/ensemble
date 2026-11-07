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
            await page.waitForFunction(() => settings !== null && !selectionPending && typeof window.EnsembleArtifactRenderers?.edit_file === 'function');
            if (width === 390) await page.locator('.pane-tab[data-pane="tools"]').click();
            await page.evaluate(() => {
                settings = { ...settings, tts_enabled: false };
                const segment = (text, extra = 0) => ({ text, total_bytes: new TextEncoder().encode(text).length + extra });
                const add = (id, tool, evidence, error = false) => {
                    handle({ type: 'tool_dispatched', part_id: id + '-call', name: tool, input: {} });
                    handle({ type: 'tool_finished', part_id: id, tool_part_id: id + '-call', result: 'Recorded raw output', file_evidence: evidence, is_error: error });
                };
                const evidence = { operation: 'read', path: '/sample.cs', start_line: 27, start_column: 1,
                    content: segment('const x = "<img src=x>";\nreturn 42;\n', 100), removed: null };
                add('file-read', 'read_file', evidence);
                add('file-append', 'write_file', { ...evidence, operation: 'append', start_line: 3, start_column: 5, content: segment('tail') });
                add('file-edit', 'edit_file', { ...evidence, operation: 'edit', content: segment('new\nvalue'), removed: segment('old') });
                add('file-empty', 'write_file', { ...evidence, operation: 'write', content: segment('') });
                add('file-plain', 'read_file', { ...evidence, path: '/sample.unknown' });
                add('file-legacy', 'read_file', null);
                add('file-error', 'write_file', evidence, true);
            });
            const read = page.locator('[data-part-id="file-read"]');
            assert.deepEqual(await read.locator('.file-line-number').allTextContents(), ['27', '28']);
            assert.equal(await read.locator('.file-source code').allTextContents().then(x => x.join('\n')), 'const x = "<img src=x>";\nreturn 42;');
            assert(await read.locator('.file-truncation').isVisible());
            assert(await read.locator('.code-keyword').count() > 0);
            assert.equal(await read.locator('img, script').count(), 0);
            assert.equal(await read.locator('details > pre').textContent(), 'Recorded raw output');
            assert.match(await page.locator('[data-part-id="file-append"] .file-preview').textContent(), /line 3, column 5/);
            const edit = await page.locator('[data-part-id="file-edit"] .file-preview').textContent();
            for (const label of ['not a whole-file diff', 'Removed', 'Inserted']) assert(edit.includes(label));
            assert.match(await page.locator('[data-part-id="file-empty"] .file-preview').textContent(), /No content/);
            assert.equal(await page.locator('[data-part-id="file-plain"] .code-keyword').count(), 0);
            for (const id of ['file-legacy', 'file-error']) assert.equal(await page.locator(`[data-part-id="${id}"] .preview`).textContent(), 'Recorded raw output');
            assert(await page.evaluate(() => {
                const card = cards.get('file-read'); const preview = card.node.querySelector('.file-preview');
                const summary = card.node.querySelector('summary'); summary.focus();
                if (document.activeElement !== summary) throw Error('Focus precondition failed');
                render(card);
                return card.node.querySelector('.file-preview') === preview && document.activeElement === summary;
            }));
            assert.deepEqual(errors, []);
            await page.close();
            console.log(`PASS file renderers ${width}: offsets, literal content, suffix highlighting, truncation, replacement labels, fallback, raw output and focus`);
        }
    } finally { await browser.close(); }
})().catch(e => { console.error(e); process.exitCode = 1; });
