'use strict';
// Real embedded page and pinned parser; only clipboard permission/transport is
// synthetic. No operating-system clipboard reads or writes occur in this test.
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
        // Opt-in sensitivity seam: serve an explicit local candidate script in
        // place of this one asset. Default tests use only the embedded bytes.
        if (process.env.GUI_MARKDOWN_SOURCE) {
            const file = require('node:path').resolve(process.env.GUI_MARKDOWN_SOURCE);
            const body = require('node:fs').readFileSync(file, 'utf8');
            await page.route('**/markdown.js', route => route.fulfill({
                status: 200, contentType: 'text/javascript', body
            }));
        }
        const errors = [];
        page.on('pageerror', error => errors.push(error.message));
        await page.goto(base);
        await page.waitForFunction(() => settings !== null && !selectionPending);
        await page.evaluate(() => {
            window.clipboardWrites = [];
            window.clipboardMode = 'success';
            Object.defineProperty(navigator, 'clipboard', { configurable: true, value: {
                writeText: async text => {
                    clipboardWrites.push(text);
                    if (clipboardMode === 'denied') throw new DOMException('Denied', 'NotAllowedError');
                    if (clipboardMode === 'delayed') await new Promise(resolve => { window.finishCopy = resolve; });
                }
            }});
        });
        const closureCases = [
            ['```\nx\n```', true], ['```\nx', false], ['```\nx\n', false],
            ['```\n', false], ['```\n```', true], ['````\n```\n````', true],
            ['````\n```', false], ['```\nx\n    ```', false],
            ['> ```\n> x\n> ```', true], ['> ```\n> x\n\noutside', false],
            ['- ```\n  x\n  ```', true], ['- ```\n  x\n- next', false],
            ['~~~\r\nx\r\n~~~\r\n', true], ['```\nx\n``` trailing', false],
            ['```\nx\n~~~', false], ['  ```\n  x\n  ```', true],
            ['```\n\n\n```', true], ['```\n\n', false]
        ];
        for (const [source, expected] of closureCases) {
            assert.equal(await page.evaluate(source => closedCodeFence(
                messageMarkdown.parse(source, {}).find(token => token.type === 'fence')), source), expected, source);
        }
        const renderText = async (id, text, role = 'assistant', type = 'text') => {
            await page.evaluate(({ id, text, role, type }) => handle({
                type: 'part_final', part_id: id, part: { type, text, role }
            }), { id, text, role, type });
            return page.locator(`[data-part-id="${id}"]`);
        };
        const samples = [
            ['```js\n  const x = "界 😀";\n\n\treturn x;  \n```', '  const x = "界 😀";\n\n\treturn x;  \n'],
            ['> ```\r\n>   quoted\r\n> \r\n> ```\r\n', '  quoted\n\n'],
            ['- item\n\n  ~~~\n  nested\n  ~~~', 'nested\n'],
            ['````\n```\n<script>window.snippetExecuted = true</script>\n````', '```\n<script>window.snippetExecuted = true</script>\n'],
            ['```\n```', '']
        ];
        for (const [index, [markdown, payload]] of samples.entries()) {
            const card = await renderText(`copy-${index}`, markdown);
            await card.locator('[data-action="copy-code"]').click();
            assert.equal(await page.evaluate(() => clipboardWrites.at(-1)), payload);
            assert.equal(await card.locator('[role="status"]').textContent(), 'Copied');
            assert.equal(await page.evaluate(id => cards.get(id).text, `copy-${index}`), markdown);
        }
        assert.equal(await page.evaluate(() => window.snippetExecuted), undefined);
        console.log('PASS fence closure, exact logical bytes, LF normalization, nested/empty/literal snippets');

        for (const [id, text, role, type] of [
            ['unfinished', '```\nx', 'assistant', 'text'],
            ['indented', '    x\n', 'assistant', 'text'],
            ['inline', '`x`', 'assistant', 'text'],
            ['human-code', '```\nx\n```', 'human', 'text'],
            ['thinking-code', '```\nx\n```', 'assistant', 'thinking']
        ]) {
            const card = await renderText(id, text, role, type);
            assert.equal(await card.locator('[data-action="copy-code"]').count(), 0, id);
        }
        await page.evaluate(() => handle({ type: 'part_delta', part_id: 'copy-stream', kind: 'text', chunk: '```\nx' }));
        assert.equal(await page.locator('[data-part-id="copy-stream"] [data-action="copy-code"]').count(), 0);
        await page.evaluate(() => handle({ type: 'part_delta', part_id: 'copy-stream', kind: 'text', chunk: '\n```' }));
        const stream = page.locator('[data-part-id="copy-stream"]');
        await stream.locator('[data-action="copy-code"]').waitFor();
        await stream.locator('[data-action="copy-code"]').focus();
        await page.evaluate(() => {
            window.originalCopyButton = document.activeElement;
            handle({ type: 'part_delta', part_id: 'copy-stream', kind: 'text', chunk: '\n\nLater prose' });
        });
        await page.waitForFunction(() => pendingCardRenders.size === 0);
        assert(await page.evaluate(() => document.activeElement === originalCopyButton && originalCopyButton.isConnected));
        console.log('PASS no premature/nonassistant copy; same control and focus survive streaming');

        const payload = '  long code 界\n'.repeat(200);
        assert(payload.length > 1800, 'fixture must exceed character and line preview limits');
        const long = await renderText('copy-long', '```\n' + payload + '```');
        assert.equal(await long.locator('details').getAttribute('open'), null);
        await long.locator('[data-action="copy-code"]').click();
        assert.equal(await page.evaluate(() => clipboardWrites.at(-1)), payload, 'full-payload copy must not use preview');
        await long.locator('summary').click();
        await page.evaluate(() => render(cards.get('copy-long')));
        assert.equal(await long.locator('details').getAttribute('open'), '');
        assert.equal(await long.locator('[role="status"]').textContent(), 'Copied');
        console.log('PASS collapsed preview copies full payload; expansion and feedback survive rerender');

        await page.evaluate(() => { clipboardMode = 'denied'; });
        await stream.locator('[data-action="copy-code"]').click();
        const manual = stream.locator('textarea');
        assert.equal(await manual.inputValue(), 'x\n');
        assert(await manual.evaluate(node => node.readOnly && document.activeElement === node && node.selectionStart === 0 && node.selectionEnd === node.value.length));
        await page.evaluate(() => handle({ type: 'part_delta', part_id: 'copy-stream', kind: 'text', chunk: '\nMore prose' }));
        await page.waitForFunction(() => pendingCardRenders.size === 0);
        assert(await manual.evaluate(node => document.activeElement === node && node.selectionEnd === node.value.length));
        assert.match(await stream.locator('[role="status"]').textContent(), /Clipboard unavailable/);
        // Missing API takes the same explicit manual path, without an attempted write.
        await page.evaluate(() => { Object.defineProperty(navigator, 'clipboard', { configurable: true, value: undefined }); });
        await long.locator('[data-action="copy-code"]').click();
        assert.equal(await long.locator('textarea').inputValue(), payload);
        assert(await long.locator('textarea').evaluate(node => node.previousElementSibling.classList.contains('snippet-controls')));
        console.log('PASS denied and missing clipboard expose exact read-only fallback, stable through streaming');

        await page.evaluate(() => {
            clipboardMode = 'delayed';
            Object.defineProperty(navigator, 'clipboard', { configurable: true, value: {
                writeText: text => { clipboardWrites.push(text); return new Promise(resolve => { window.finishCopy = resolve; }); }
            }});
            getSelection().removeAllRanges();
        });
        const late = await renderText('copy-late', '```\noriginal\n```');
        await late.locator('[data-action="copy-code"]').click();
        await page.evaluate(() => {
            cards.get('copy-late').node.remove();
            cards.delete('copy-late');
            finishCopy();
        });
        const replacement = await renderText('copy-late', '```\nreplacement\n```');
        assert.equal(await page.evaluate(() => clipboardWrites.at(-1)), 'original\n');
        assert.equal(await replacement.locator('[role="status"]').textContent(), '');
        assert(await page.evaluate(() => transcript.scrollWidth <= transcript.clientWidth + 1));
        assert.deepEqual(errors, []);
        console.log('PASS late clipboard completion cannot update replacement card; mobile containment and no page errors');
    } finally {
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
