'use strict';
// Real controls and imports against an isolated, no-inference host.
const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const os = require('node:os');
const path = require('node:path');
const { spawn } = require('node:child_process');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
(async () => {
    const directory = await fs.mkdtemp(path.join(os.tmpdir(), 'renderer-trial-'));
    const child = spawn(process.argv[2], [process.argv[3], '--artifact-renderer-fixture', directory]);
    let log = '';
    child.stdout.on('data', data => { log += data; });
    child.stderr.on('data', data => { log += data; });
    const exited = new Promise(resolve => child.on('exit', resolve));
    let browser;
    try {
        let base;
        for (let i = 0; i < 300; i++) {
            try { base = await fs.readFile(path.join(directory, 'url'), 'utf8'); break; } catch {}
            if (child.exitCode !== null) throw new Error(log);
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        assert(base, 'fixture readiness');
        browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
        for (const width of [1280, 390]) {
            const page = await browser.newPage({ viewport: { width, height: 844 } });
            const errors = [];
            page.on('pageerror', error => errors.push(error.message));
            const state = page.locator('#renderer-status');
            async function ready(text) {
                try { await page.waitForFunction(text => document.querySelector('#renderer-status').textContent.includes(text), text); }
                catch (error) { throw new Error(`Expected ${text}; tool=${await page.locator('#renderer-tool').inputValue()}; status=${await state.textContent()}`, { cause: error }); }
            }
            async function panel() {
                await page.locator('#settings-open').click();
                await page.locator('#settings-display-tab').click();
            }
            async function candidate(marker, broken = false) {
                const pkg = { entry: 'main.js', files: {
                    'main.js': broken ? "throw new Error('deliberate candidate failure');" :
                        "import { label } from './helper.js'; window.trialMarker = label; export default (document,text) => { const node=document.createElement('pre'); node.textContent=label+': '+text; return node; };",
                    'helper.js': 'export const label = ' + JSON.stringify(marker) + ';'
                } };
                await page.locator('#renderer-package').setInputFiles({ name: 'candidate.json', mimeType: 'application/json', buffer: Buffer.from(JSON.stringify(pkg)) });
                await ready('Staging has not executed');
                assert((await page.locator('#renderer-source').textContent()).includes('helper.js'));
            }
            await page.goto(base);
            await ready('Using the shipped');
            await page.waitForFunction(() => settings !== null && !selectionPending);
            await page.evaluate(() => {
                handle({ type: 'tool_dispatched', part_id: 'trial-call', name: 'list_directory', input: { path: '.' } });
                handle({ type: 'tool_finished', part_id: 'trial-result', tool_part_id: 'trial-call', result: 'recorded-file.txt' });
            });
            await panel();
            await candidate('trial-' + width);
            assert.equal(await page.evaluate(() => window.trialMarker), undefined, 'stage executed module');
            await page.locator('#renderer-try').click();
            await ready('Trial active');
            assert.equal(await page.evaluate(() => window.trialMarker), 'trial-' + width);
            await page.waitForFunction(marker => cards.get('trial-result').node.querySelector('.preview').textContent === marker + ': recorded-file.txt', 'trial-' + width);
            assert.equal(await page.evaluate(() => cards.get('trial-result').node.querySelector('details > pre').textContent), 'recorded-file.txt');
            await page.reload();
            await ready('Using the shipped');
            assert.equal(await page.evaluate(() => window.trialMarker), undefined, 'reload resumed trial');
            await panel();
            await candidate('kept-' + width);
            await page.locator('#renderer-try').click();
            await ready('Trial active');
            await page.locator('#renderer-keep').click();
            await ready('Accepted ');
            await page.reload();
            await ready('Using accepted');
            assert.equal(await page.evaluate(() => window.trialMarker), 'kept-' + width);
            await panel();
            await candidate('broken', true);
            await page.locator('#renderer-try').click();
            await ready('Candidate could not load');
            assert(await page.locator('#renderer-keep').isDisabled(), 'failed import can be kept');
            await page.reload();
            await ready('Using accepted');
            assert.equal(await page.evaluate(() => window.trialMarker), 'kept-' + width);
            await page.goto(base + '?renderers=shipped');
            await ready('Recovery mode');
            assert.equal(await page.evaluate(() => window.trialMarker), undefined, 'recovery mode imported custom code');
            await panel();
            await page.locator('#renderer-rollback').click();
            await ready('Recovery mode');
            await page.goto(base);
            await ready('Using the shipped');
            assert.equal(await page.evaluate(() => window.trialMarker), undefined);
            for (const tool of ['read_file', 'search_files', 'write_file', 'edit_file', 'tool_default', 'prose', 'command']) {
                await panel();
                await page.locator('#renderer-tool').selectOption(tool);
                await ready('Using the shipped ' + tool);
                const pkg = await page.evaluate(async tool => {
                    const response = await fetch('./renderers/shipped?renderer=' + tool);
                    if (!response.ok) throw new Error('shipped package');
                    return response.json();
                }, tool);
                pkg.files[pkg.entry] += '\nwindow.packageMarker = ' + JSON.stringify(tool) + ';\n';
                await page.locator('#renderer-package').setInputFiles({ name: tool + '.json', mimeType: 'application/json', buffer: Buffer.from(JSON.stringify(pkg)) });
                await ready('Staging has not executed');
                assert.equal(await page.evaluate(() => window.packageMarker), undefined, 'staging executed tool package');
                await page.locator('#renderer-try').click();
                await ready('Trial active');
                assert(await page.locator('#renderer-tool').isDisabled(), 'tool changed under active trial');
                assert.equal(await page.evaluate(() => window.packageMarker), tool);
                if (tool === 'command') {
                    assert(await page.evaluate(() => {
                        const node = document.createElement('pre');
                        window.EnsembleArtifactRenderers.command(node, new TextEncoder().encode('\x1b[31mSAFE <img>\x1b[0m'), false);
                        return node.textContent === 'SAFE <img>' && !!node.querySelector('.terminal-fg-1') && !node.querySelector('img');
                    }), 'command package failed its decoder import');
                }
                if (tool === 'prose') {
                    assert.equal(await page.evaluate(() => {
                        const node = document.createElement('div');
                        window.EnsembleArtifactRenderers.prose(node, '**Recorded**\n\n```js\nconst answer = 42;\n```', true);
                        return node.querySelector('strong')?.textContent === 'Recorded'
                            && !!node.querySelector('.code-keyword')
                            && !!node.querySelector('[data-action="copy-code"]')
                            && window.markdownit === undefined;
                    }), true, 'prose package depended on page-global parser or lost controls');
                }
                if (['read_file', 'write_file', 'edit_file'].includes(tool)) {
                    assert.equal(await page.evaluate(tool => {
                        const operation = { read_file: 'read', write_file: 'write', edit_file: 'edit' }[tool];
                        const segment = { text: 'const answer = 42;', total_bytes: 18 };
                        const node = window.EnsembleArtifactRenderers[tool](document, '', null,
                            { operation, path: 'example.js', start_line: 1, start_column: 1,
                                content: segment, removed: segment, inserted: segment });
                        return !!node.querySelector('.code-keyword');
                    }, tool), true, 'packaged lexical helper did not render');
                }
                await page.locator('#renderer-keep').click();
                await ready('Accepted ');
                await page.reload();
                await ready('Using the shipped');
                assert.equal(await page.evaluate(() => window.packageMarker), tool, 'accepted tool not restored at startup');
                await page.goto(base + '?renderers=shipped');
                await ready('Recovery mode');
                assert.equal(await page.evaluate(() => window.packageMarker), undefined, 'safe mode imported accepted tool');
                await panel();
                await page.locator('#renderer-tool').selectOption(tool);
                await page.waitForFunction(() => !document.querySelector('#renderer-rollback').disabled);
                await page.locator('#renderer-rollback').click();
                await ready('Recovery mode');
                await page.goto(base);
                await ready('Using the shipped');
                assert.equal(await page.evaluate(() => window.packageMarker), undefined);
            }
            for (const name of ['remote_weather', 'command']) {
                await panel();
                const key = await page.locator('#renderer-tool option').evaluateAll((options, name) =>
                    options.find(option => option.textContent === 'Tool: ' + name)?.value, name);
                assert(key, 'registered MCP tool absent from choices');
                await page.locator('#renderer-tool').selectOption(key);
                await ready('Using the default preview');
                assert(await page.locator('#renderer-download').isHidden(), 'invented shipped MCP source');
                await candidate('remote-' + name + '-' + width);
                assert.equal(await page.evaluate(() => window.trialMarker), undefined);
                await page.locator('#renderer-try').click(); await ready('Trial active');
                assert.equal(await page.evaluate(name => {
                    const card = cardFor('remote-result');
                    Object.assign(card, { kind: 'tool_result', role: 'tool', status: 'completed', toolName: name, text: 'recorded remote data' });
                    render(card);
                    return card.node.querySelector('.preview').textContent;
                }, name), 'remote-' + name + '-' + width + ': recorded remote data');
                assert.equal(await page.evaluate(() => typeof window.EnsembleArtifactRenderers.command), 'function');
                await page.locator('#renderer-keep').click(); await ready('Accepted ');
                await page.reload(); await ready('Using the shipped');
                assert.equal(await page.evaluate(() => window.trialMarker), 'remote-' + name + '-' + width);
                await page.goto(base + '?renderers=shipped'); await ready('Recovery mode');
                assert.equal(await page.evaluate(() => window.trialMarker), undefined);
                await panel(); await page.locator('#renderer-tool').selectOption(key); await ready('Recovery mode');
                await page.locator('#renderer-rollback').click(); await ready('Recovery mode');
                await page.goto(base); await ready('Using the shipped');
            }
            console.log(`PASS MCP renderer ${width}: actual registered names, inert stage, name collision, try/keep/reload/recovery/rollback`);
            console.log(`PASS shared approval ${width}: seven additional output selectors, inert packages, helper imports, independent keep/reload, safe mode, rollback`);
            assert.deepEqual(await page.evaluate(() => {
                const registry = window.EnsembleArtifactRenderers;
                const original = registry.tool_default;
                const card = cardFor('default-preview-check');
                Object.assign(card, { kind: 'tool_result', role: 'tool', status: 'completed',
                    toolName: 'unfamiliar_tool', text: '<hostile> recorded output' });
                render(card);
                const preview = () => card.node.querySelector('.preview');
                const literal = preview().textContent;
                registry.tool_default = (document, text) => {
                    const node = document.createElement('div'); node.textContent = 'Custom: ' + text; return node;
                };
                render(card); const custom = preview().textContent;
                window.EnsembleToolRenderers.unfamiliar_tool = () => { throw new Error('specialized refusal'); };
                render(card); const specializedFailure = preview().textContent;
                registry.tool_default = () => { throw new Error('default refusal'); };
                render(card); const defaultFailure = preview().textContent;
                const heading = card.node.querySelector('.card-heading');
                registry.tool_default = () => heading;
                render(card); const connectedRejected = heading.parentElement === card.node && preview().textContent === literal;
                delete window.EnsembleToolRenderers.unfamiliar_tool; registry.tool_default = original; render(card);
                return { literal, custom, specializedFailure, defaultFailure, connectedRejected };
            }), { literal: '<hostile> recorded output', custom: 'Custom: <hostile> recorded output',
                specializedFailure: 'Custom: <hostile> recorded output',
                defaultFailure: '<hostile> recorded output', connectedRejected: true });
            console.log('PASS default preview: unknown tools, custom refresh, throwing specialized/default, connected-node refusal');
            assert.deepEqual(errors, []);
            console.log(`PASS renderer trial ${width}: inert stage, helpers, try, reload discard, keep, failed import, recovery, rollback`);
            await page.close();
        }
    } finally {
        if (browser) await browser.close();
        await fs.writeFile(path.join(directory, 'stop'), '');
        let timer;
        const code = await Promise.race([exited, new Promise(resolve => { timer = setTimeout(() => resolve('timeout'), 20000); })]);
        clearTimeout(timer);
        if (code === 'timeout') { child.kill('SIGKILL'); await exited; throw new Error('Fixture required forced cleanup'); }
        assert.equal(code, 0, log);
        await fs.rm(directory, { recursive: true, force: true });
        console.log('PASS fixture host joined');
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
