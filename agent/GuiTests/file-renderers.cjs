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
                add('file-append', 'write_file', { ...evidence, operation: 'append', start_line: 3, start_column: 5, content: segment('tail\")\nreturn 42;') });
                add('file-edit', 'edit_file', { ...evidence, operation: 'edit', content: segment('new\nvalue'), removed: segment('old') });
                const middle = Array.from({length: 20}, (_, i) => `context ${i}\n`).join('');
                add('file-separated', 'edit_file', { ...evidence, operation: 'edit',
                    removed: segment('old one\n' + middle + 'old two\n'), content: segment('new one\n' + middle + 'new two\n') });
                add('file-truncated-edit', 'edit_file', { ...evidence, operation: 'edit',
                    removed: segment('old', 10), content: segment('new') });
                add('file-budget', 'edit_file', { ...evidence, operation: 'edit',
                    removed: segment('old\n'.repeat(300)), content: segment('new\n'.repeat(300)) });
                add('file-fragment', 'edit_file', { ...evidence, operation: 'edit', start_column: 17,
                    removed: segment('<img src=x>'), content: segment('<script>evil()</script>') });
                add('file-newline', 'edit_file', { ...evidence, operation: 'edit', removed: segment('same'), content: segment('same\n') });
                add('file-delete', 'edit_file', { ...evidence, operation: 'edit', removed: segment('gone\n'), content: segment('') });
                add('file-insert', 'edit_file', { ...evidence, operation: 'edit', removed: segment(''), content: segment('new\r\n') });
                add('file-empty', 'write_file', { ...evidence, operation: 'write', content: segment('') });
                add('file-plain', 'read_file', { ...evidence, path: '/sample.unknown' });
                add('file-long', 'read_file', { ...evidence, path: '/long/path/to/cmd/main.go',
                    content: segment(('import \"fmt\" // ' + 'long source '.repeat(40) + '\n').repeat(80)) });
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
            const append = page.locator('[data-part-id="file-append"] .file-preview');
            assert.equal(await append.locator('.file-name').textContent(), 'sample.cs · appended at 3:5');
            assert.match(await append.locator('.file-name').getAttribute('title'), /line 3, column 5 \(UTF-16\)/);
            assert.equal(await append.locator('.file-name').getAttribute('aria-label'), await append.locator('.file-name').getAttribute('title'));
            assert.equal(await append.locator(':scope > p').count(), 1, 'Append position must not consume a separate paragraph');
            assert.equal(await append.locator('[class^="code-"]').count(), 0, 'Partial fragment must not guess lexical context');
            assert.deepEqual(await append.locator('.file-source code').allTextContents(), ['tail")', 'return 42;']);
            const edit = await page.locator('[data-part-id="file-edit"] .file-preview').textContent();
            for (const label of ['−', '+', 'No final newline']) assert(edit.includes(label));
            assert.match(await page.locator('[data-part-id="file-edit"] .file-name').getAttribute('title'), /not a whole-file diff/);
            assert.equal(await page.locator('[data-part-id="file-edit"] .file-name').textContent(), 'sample.cs');
            assert.match(await page.locator('[data-part-id="file-edit"] .file-name').getAttribute('aria-label'), /old\/new line numbers/);
            const separated = page.locator('[data-part-id="file-separated"]');
            assert.match(await separated.locator('.file-diff-gap').textContent(), /14 unchanged lines omitted/);
            const changed = await separated.locator('.file-source-line').evaluateAll(nodes => nodes.filter(n => n.querySelector('code') && /^(old|new) /.test(n.querySelector('code').textContent)).map(n => [...n.children].map(c => c.textContent)));
            assert.deepEqual(changed, [['27', '', '−', 'old one'], ['', '27', '+', 'new one'], ['48', '', '−', 'old two'], ['', '48', '+', 'new two']]);
            assert.match(await page.locator('[data-part-id="file-truncated-edit"] .file-preview').textContent(), /Incomplete evidence.*Removed.*Inserted/s);
            assert.match(await page.locator('[data-part-id="file-budget"] .file-preview').textContent(), /Comparison budget reached/);
            const fragment = page.locator('[data-part-id="file-fragment"]');
            assert.equal(await fragment.locator('img, script, [class^="code-"]').count(), 0);
            assert.match(await fragment.locator('.file-name').textContent(), /partial line at 27:17/);
            assert.deepEqual(await fragment.locator('code').allTextContents(), ['<img src=x>', '<script>evil()</script>']);
            const newline = page.locator('[data-part-id="file-newline"]');
            assert.deepEqual(await newline.locator('code').allTextContents(), ['same', 'same']);
            assert.match(await newline.locator('.file-diff-boundary').textContent(), /removed segment/);
            assert.deepEqual(await page.locator('[data-part-id="file-delete"] code').allTextContents(), ['gone']);
            assert.deepEqual(await page.locator('[data-part-id="file-insert"] code').allTextContents(), ['new\r']);
            const pkg = await page.evaluate(async () => {
                const response = await fetch('./renderers/shipped?renderer=edit_file');
                if (!response.ok) throw Error('Shipped edit package unavailable: ' + response.status);
                return response.json();
            });
            for (const file of ['shared/file-diff.js', 'shared/jsdiff/diff/line.js', 'shared/jsdiff/diff/base.js', 'shared/jsdiff/util/params.js', 'shared/jsdiff/license.js']) assert.equal(typeof pkg.files[file], 'string', file);

            assert.match(await page.locator('[data-part-id="file-empty"] .file-preview').textContent(), /No content/);
            assert.equal(await page.locator('[data-part-id="file-plain"] .code-keyword').count(), 0);
            for (const id of ['file-legacy', 'file-error']) assert.equal(await page.locator(`[data-part-id="${id}"] .preview`).textContent(), 'Recorded raw output');
            assert(await page.evaluate(() => {
                const card = cards.get('file-read'); const preview = card.node.querySelector('.file-preview');
                const summary = card.node.querySelector('.preview'); summary.focus();
                if (document.activeElement !== summary) throw Error('Focus precondition failed');
                render(card);
                return card.node.querySelector('.file-preview') === preview && document.activeElement === summary;
            }));
            const long = page.locator('[data-part-id="file-long"]');
            assert.equal(await long.locator('.file-name').textContent(), 'main.go');
            assert.equal(await long.locator('.file-name').getAttribute('title'), '/long/path/to/cmd/main.go');
            assert(!((await long.locator('.file-preview').textContent()).includes('Captured content')));
            for (const size of ['18px', '28px']) {
                await page.evaluate(size => { document.documentElement.style.fontSize = size; }, size);
                const toolsTab = page.locator('.pane-tab[data-pane="tools"]');
                if (await toolsTab.isVisible()) await toolsTab.click();
                await long.locator('.preview').waitFor({ state: 'visible' });
                const geometry = await long.evaluate(card => {
                    const preview = card.querySelector('.preview'), source = card.querySelector('.file-source');
                    const style = getComputedStyle(source);
                    preview.scrollTop = 60; preview.scrollLeft = 60;
                    return { vertical: preview.scrollTop > 0, horizontal: preview.scrollLeft > 0,
                        innerOverflow: style.overflowY, maximum: style.maxHeight,
                        width: preview.clientWidth, scrollWidth: preview.scrollWidth, height: preview.clientHeight, scrollHeight: preview.scrollHeight,
                        overflow: getComputedStyle(preview).overflow,
                        noWrap: getComputedStyle(source.querySelector('code')).whiteSpace,
                        sourceOffset: source.getBoundingClientRect().top - card.querySelector('.file-name').getBoundingClientRect().bottom };
                });
                assert(geometry.vertical && geometry.horizontal, 'One outer preview must scroll in both directions: ' + JSON.stringify(geometry));
                assert.equal(geometry.innerOverflow, 'visible');
                assert.equal(geometry.maximum, 'none');
                assert.equal(geometry.noWrap, 'pre');
                assert(geometry.sourceOffset < 20, 'Redundant header space above source');
            }
            assert.deepEqual(errors, []);
            await page.close();
            console.log(`PASS file renderers ${width}: offsets, literal content, suffix highlighting, truncation, replacement labels, fallback, raw output and focus`);
        }
    } finally { await browser.close(); }
})().catch(e => { console.error(e); process.exitCode = 1; });
